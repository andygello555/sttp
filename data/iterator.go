package data

import (
	"container/heap"
	"fmt"
	"github.com/andygello555/errors"
	"strings"
)

type Element struct {
	// Key of the current element in the list of elements. This can either be a Number or a String.
	Key *Value
	// Val is the value of current element in the list of elements. This can be a Value of any Type.
	Val *Value
}

type Iterator []*Element

func (it Iterator) Len() int { return len(it) }

func (it Iterator) Less(i, j int) bool {
	switch it[i].Key.Type {
	case Number:
		return it[i].Key.Value.(float64) < it[j].Key.Value.(float64)
	case String:
		return strings.Compare(it[i].Key.StringLit(), it[j].Key.Value.(string)) <= 0
	default:
		panic(fmt.Errorf("cannot have iterator with keys of type: %s", it[i].Key.Type.String()))
	}
}

func (it Iterator) Swap(i, j int) { it[i], it[j] = it[j], it[i] }

func (it *Iterator) Push(x interface{}) { *it = append(*it, x.(*Element)) }

func (it *Iterator) Pop() interface{} {
	old := *it
	n := len(old)
	elem := old[n-1]
	old[n-1] = nil // avoid memory leak
	*it = old[0 : n-1]
	return elem
}

func (it *Iterator) Next() *Element {
	if it.Len() > 0 {
		if (*it)[0].Key.Type == String {
			return heap.Pop(it).(*Element)
		} else {
			old := *it
			elem := old[0]
			old[0] = nil
			*it = old[1:]
			return elem
		}
	}
	return nil
}

// Iterate will construct an iterator from the given Value. This Value must be of Type: String, Object, or Array.
func Iterate(result *Value) (err error, it *Iterator) {
	defer func() {
		if p := recover(); p != nil {
			switch p.(type) {
			case struct{ errors.ProtoSttpError }:
				err = p.(struct{ errors.ProtoSttpError })
			default:
				err = fmt.Errorf("%v", p)
			}
		}
	}()

	t := func(x interface{}) Type {
		var t Type
		if err = t.Get(x); err != nil {
			panic(err)
		}
		return t
	}

	var iterator Iterator
	switch result.Type {
	case Object:
		obj := result.Value.(map[string]interface{})
		iterator = make(Iterator, 0)

		for k, v := range obj {
			heap.Push(&iterator, &Element{
				Key: &Value{
					Value:    k,
					Type:     String,
					Global:   false,
					ReadOnly: true,
				},
				Val: &Value{
					Value:    v,
					Type:     t(v),
					Global:   false,
					ReadOnly: true,
				},
			})
		}
	case Array:
		arr := result.Value.([]interface{})
		iterator = make(Iterator, len(arr))

		for i, v := range arr {
			iterator[i] = &Element{
				Key: &Value{
					Value:    float64(i),
					Type:     Number,
					Global:   false,
					ReadOnly: true,
				},
				Val: &Value{
					Value:    v,
					Type:     t(v),
					Global:   false,
					ReadOnly: true,
				},
			}
		}
	case String:
		str := result.StringLit()
		iterator = make(Iterator, len(str))

		for i, v := range str {
			iterator[i] = &Element{
				Key: &Value{
					Value:    float64(i),
					Type:     Number,
					Global:   false,
					ReadOnly: true,
				},
				Val: &Value{
					Value:    string(v),
					Type:     String,
					Global:   false,
					ReadOnly: true,
				},
			}
		}
	}
	return nil, &iterator
}
