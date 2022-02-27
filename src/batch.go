package main

import (
	"container/heap"
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/parser"
	"strings"
	"sync"
)

// MaxWorkers are the max number of worker goroutines we can create in our pool.
const MaxWorkers = 20

// BatchItem is a unit of work that is distributed amongst each methodWorker.
type BatchItem struct {
	Method *parser.MethodCall
	Args   []*data.Value
	Id     int
}

// BatchResult contains the result for one BatchItem.
type BatchResult struct {
	Id     int
	Method *parser.MethodCall
	Err    error
	Value  *data.Value
}

// GetErr will return the Err for this BatchResult.
func (br *BatchResult) GetErr() error {
	return br.Err
}

// GetValue will return the Value for this BatchResult.
func (br *BatchResult) GetValue() *data.Value {
	return br.Value
}

// GetMethodCall will return the parser.MethodCall for this BatchResult.
func (br *BatchResult) GetMethodCall() *parser.MethodCall {
	return br.Method
}

// BatchResults implements heap.Interface, so that BatchResults can be quickly added back in the order in which they 
// arrived.
type BatchResults []*BatchResult

func (b BatchResults) Len() int { return len(b) }

func (b BatchResults) Less(i, j int) bool { return b[i].Id < b[j].Id }

func (b BatchResults) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b *BatchResults) Push(x interface{}) { *b = append(*b, x.(*BatchResult)) }

func (b *BatchResults) Pop() interface{} {
	old := *b
	n := len(old)
	x := old[n-1]
	*b = old[0:n-1]
	return x
}

func (b *BatchResults) String() string {
	var builder strings.Builder
	for _, result := range *b {
		builder.WriteString(fmt.Sprintf("%d: %s\n", result.Id, result.Value.String()))
	}
	return builder.String()
}

// BatchSuite represents a currently running parser.Batch statement. It contains the properties required to manage the
// worker pool that will be started to execute the MethodCalls enqueued in it.
type BatchSuite struct {
	// BatchStatement is the AST node that this BatchSuite is related to.
	BatchStatement *parser.Batch
	// Results is a heap.Interface that has the results of every job that was enqueued into the BatchSuite, in the order
	// that it was enqueued.
	Results        BatchResults
	// CurrentId is a counter for the ID that is given to each enqueued job.
	CurrentId      int
	// jobChan is a buffered channel that holds the jobs to execute within the worker goroutines.
	jobChan        chan *BatchItem
	// resultChan is a buffered channel that the workers enqueue their results into.
	resultChan     chan *BatchResult
	// consumerDone is an unbuffered channel used to block the interpreter thread until the consumer has added all the 
	// results to Results.
	consumerDone   chan struct{}
	// workerGroup is a sync.WaitGroup that is used to wait until all workers have executed the work that they have been
	// given.
	workerGroup    sync.WaitGroup
	// close is used to execute the Stop only once, so that no panics occur if the channels (mentioned above) are 
	// already closed.
	close          sync.Once
}

// Batch creates a new BatchSuite. It creates buffered job and result channels that have a capacity of MaxWorkers.
func Batch(statement *parser.Batch) *BatchSuite {
	return &BatchSuite{
		BatchStatement: statement,
		Results:        make(BatchResults, 0),
		CurrentId:      0,
		jobChan:        make(chan *BatchItem, MaxWorkers),
		resultChan:     make(chan *BatchResult, MaxWorkers),
		consumerDone:   make(chan struct{}),
	}
}

// methodWorker is the worker routine used within the BatchSuite.Execute function. It reads from a channel of jobs and 
// writes to a channel of results. When finished, the worker decrements a sync.WaitGroup.
func methodWorker(wg *sync.WaitGroup, jobs <-chan *BatchItem, results chan<- *BatchResult) {
	defer wg.Done()
	for j := range jobs {
		// Call eval.Method.Call for the parser.MethodCall's eval.Method and queue the result and err up in a 
		// BatchResult
		err, result := j.Method.Method.Call(j.Args...)
		results <- &BatchResult{
			Id:     j.Id,
			Method: j.Method,
			Err:    err,
			Value:  result,
		}
	}
}

// AddWork will enqueue the given parser.MethodCall, and its args, as a BatchItem to be executed by the workers.
func (b *BatchSuite) AddWork(method *parser.MethodCall, args... *data.Value) {
	b.jobChan <- &BatchItem{
		Method: method,
		Args:   args,
		Id:     b.CurrentId,
	}
	b.CurrentId ++
}

// GetStatement will return a pointer to a parser.Batch statement so that it can be compared and or set.
func (b *BatchSuite) GetStatement() *parser.Batch {
	return b.BatchStatement
}

// Start will spin-up the worker goroutines that will be fed the work accumulated over the course of a batch statement.
// Can be given the number of workers to spin up, if this is a negative integer, or the number of workers exceeds 
// MaxWorkers, then MaxWorkers will be used instead. A consumer goroutine will pull results from the result channel and
// push them to the Results heap.
func (b *BatchSuite) Start(workers int) {
	if workers < 0 || workers > MaxWorkers {
		workers = MaxWorkers
	}

	// We spin up the workers
	for w := 0; w < workers; w++ {
		b.workerGroup.Add(1)
		go methodWorker(&b.workerGroup, b.jobChan, b.resultChan)
	}

	// Start a consumer goroutine that will consume results and append them to the heap. We only start one consumer 
	// because it does not make sense to try and manage a mutex between several.
	go func() {
		for result := range b.resultChan {
			heap.Push(&b.Results, result)
		}
		b.consumerDone <- struct{}{}
	}()
}

// Stop will close the Batch channel, indicating to the workers that there is no more work to execute. We will also wait
// for the result consumer to finish.
func (b *BatchSuite) Stop() heap.Interface {
	b.close.Do(func() {
		close(b.jobChan)
		b.workerGroup.Wait()
		close(b.resultChan)
		<-b.consumerDone
	})
	return &b.Results
}
