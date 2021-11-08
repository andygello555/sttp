package eval

// VM acts as an interface for the overarching state of the VM used for evaluation of programs.
type VM interface {
	Eval(filename, s string) (result *Symbol, err error)
	GetHeap() *Heap
	GetScope() *int
	GetParentStatement() interface{}
}
