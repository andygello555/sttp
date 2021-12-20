package main

import (
	"container/heap"
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/parser"
	"strings"
)

// MaxWorkers are the max number of worker goroutines we can create in our pool.
const MaxWorkers = 20

// BatchItem is a unit of work that is distributed amongst each methodWorker.
type BatchItem struct {
	Method eval.Method
	Args   []*data.Value
	Id     int
}

// BatchResult contains the result for one BatchItem.
type BatchResult struct {
	Id    int
	Err   error
	Value *data.Value
}

// GetErr will return the Err for this BatchResult.
func (br *BatchResult) GetErr() error {
	return br.Err
}

// GetValue will return the Value for this BatchResult.
func (br *BatchResult) GetValue() *data.Value {
	return br.Value
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

// BatchSuite represents a currently running parser.Batch statement. It contains a list of 
type BatchSuite struct {
	BatchStatement *parser.Batch
	Batch          []*BatchItem
	CurrentId      int
}

// Batch creates a new BatchSuite. 
func Batch(statement *parser.Batch) *BatchSuite {
	return &BatchSuite{
		BatchStatement: statement,
		Batch:          make([]*BatchItem, 0),
		CurrentId:      0,
	}
}

// methodWorker is the worker routine used within the BatchSuite.Execute function. It reads from a channel of jobs and 
// writes to a channel of results.
func methodWorker(jobs <-chan *BatchItem, results chan<- *BatchResult) {
	for j := range jobs {
		// Call eval.Method.Call for the parser.MethodCall's eval.Method and queue the result and err up in a BatchResult
		err, result := j.Method.Call(j.Args...)
		results <- &BatchResult{
			Id:    j.Id,
			Err:   err,
			Value: result,
		}
	}
}

// AddWork will create and append a BatchItem to the Batch.
func (b *BatchSuite) AddWork(method eval.Method, args... *data.Value) {
	b.Batch = append(b.Batch, &BatchItem{
		Method: method,
		Args:   args,
		Id:     b.CurrentId,
	})
	b.CurrentId ++
}

// GetStatement will return a pointer to a parser.Batch statement so that it can be compared and or set.
func (b *BatchSuite) GetStatement() *parser.Batch {
	return b.BatchStatement
}

// Execute will spin up a number of worker goroutines and then enqueue all the BatchItems in the Batch to a work queue.
// The results will be dequeued from the result queue and added back to a BatchResult queue in the order in which they 
// were added to the Batch. The workers parameter indicates the number of workers to spin up. If a negative number is 
// given then the workers will match the length of the job queue, clamped to MaxWorkers. If the number is not negative 
// then the number of workers will be locked to that number, also clamped to MaxWorkers.
func (b *BatchSuite) Execute(workers int) heap.Interface {
	numJobs := len(b.Batch)
	jobs := make(chan *BatchItem, numJobs)
	results := make(chan *BatchResult, numJobs)

	// We decide how many workers to spin up by looking at the number of jobs we have as well as the workers parameter.
	if workers < 0 {
		workers = numJobs
	}
	if numJobs > MaxWorkers {
		workers = MaxWorkers
	}

	// We spin up the workers
	for w := 1; w <= workers; w++ {
		go methodWorker(jobs, results)
	}

	// Add all the BatchItems as jobs to the jobs queue
	for _, item := range b.Batch {
		jobs <- item
	}
	// Close all the jobs
	close(jobs)

	// Add each result to the BatchResult queue in the order they came in
	orderedResults := make(BatchResults, 0, numJobs)
	for r := 1; r <= numJobs; r++ {
		heap.Push(&orderedResults, <-results)
	}
	return &orderedResults
}
