package parser

import (
	"container/heap"
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
	"github.com/andygello555/gotils/strings"
	"reflect"
	str "strings"
)

const DefaultObjectKey = ""

type Path []interface{}

func abs(x int) int {
	if x < 0 {
		return (-x) - 1
	}
	return x
}

func mod(x, n int) int {
	return (x % n + n) % n
}

func firstKey(obj map[string]interface{}, idx int) (key string, ok bool) {
	// Get all script keys at the current level
	keyQueue := make(strings.StringHeap, 0)
	heap.Init(&keyQueue)
	for k := range obj {
		heap.Push(&keyQueue, k)
	}

	// Keep popping until we get to the needed index
	ok = false
	for i := 0; i < keyQueue.Len(); i++ {
		key = heap.Pop(&keyQueue).(string)
		if i == idx {
			ok = true
			break
		}
	}
	return key, ok
}

func set(current interface{}, to interface{}, path *Path) interface{} {
	if len(*path) > 0 {
		p := (*path)[0]
		property := false
		switch p.(type) {
		case string:
			property = true
		default:
			break
		}

		if current == nil {
			// We don't pop the next element from the path because we need to use this frame to construct a new value 
			// and interface then we can use the next called frame to set the property of index.
			if property {
				// If the path is a property we construct a new object
				current = set(make(map[string]interface{}), to, path)
			} else {
				// Otherwise, we create a new interface array
				current = set(make([]interface{}, 0), to, path)
			}
		} else {
			// We only remove the current path element if we are currently on a non-nil value.
			*path = (*path)[1:]
			switch current.(type) {
			case map[string]interface{}:
				obj := current.(map[string]interface{})
				var key string

				if !property {
					// If the current path is a number index then we will sort the keys of the current value 
					// lexicographically to find correct key
					var ok bool
					if key, ok = firstKey(obj, p.(int)); !ok {
						// If we cannot find the needed key we panic
						panic(errors.JSONPathError.Errorf("object", fmt.Sprintf("index %d", p.(int))))
					}
				} else {
					// Otherwise, we assert that the path is a string key
					key = p.(string)
				}

				var val interface{} = nil
				if _, ok := obj[key]; ok {
					val = obj[key]
				}
				current.(map[string]interface{})[key] = set(val, to, path)
			case []interface{}:
				if property {
					panic(errors.JSONPathError.Errorf("array", "property"))
				}

				arr := current.([]interface{})
				idx := p.(int)
				if abs(idx) < len(arr) {
					idx = mod(idx, len(arr))
					current.([]interface{})[idx] = set(arr[idx], to, path)
				} else {
					if idx < 0 {
						panic(errors.JSONPathError.Errorf("array", fmt.Sprintf("negative index that is out of array bounds (%d)", idx)))
					}

					// We insert nils up until we get to the index to set at, at which point we recurse.
					for i := len(arr); i <= idx; i++ {
						var val interface{} = nil
						if i == idx {
							val = set(nil, to, path)
						}
						arr = append(arr, val)
					}
					current = arr
				}
			default:
				if property {
					// If accessing a property then we will wrap the value in an object, assigning it to the key 
					// DefaultObjectKey
					obj := make(map[string]interface{})
					obj[DefaultObjectKey] = current
					obj[p.(string)] = set(nil, to, path)
					current = obj
				} else {
					// If accessing an index then we will create an array with p.(int) + 2 spaces. Recursing down the 
					// p.(int) space and inserting the existing value in the p.(int) + 1 space.
					idx := p.(int)
					if idx >= 0 {
						arr := make([]interface{}, idx+2)
						arr[idx+1] = current
						for i := 0; i < len(arr)-1; i++ {
							var val interface{} = nil
							if i == idx {
								val = set(nil, to, path)
							}
							arr[i] = val
						}
						current = arr
					} else {
						panic(errors.JSONPathError.Errorf("non-object/array type", fmt.Sprintf("a negative index (%d)", idx)))
					}
				}
			}
		}
	} else {
		current = to
	}
	return current
}

func (p *Path) Set(current interface{}, to interface{}) (err error, new interface{}) {
	defer func() {
		if p := recover(); p != nil {
			switch p.(type) {
			case struct { errors.ProtoSttpError }:
				err = p.(struct { errors.ProtoSttpError })
			default:
				err = fmt.Errorf("%v", p)
			}
			new = nil
		}
	}()

	c := make(Path, len(*p) - 1)
	// Copy everything but the first element (variable name)
	copy(c, (*p)[1:])
	new = set(current, to, &c)
	return err, new
}

// Get iterates over the referred to Path and descends the given value and returns the value pointed to by the path. 
// If the path doesn't lead anywhere it will return nil.
func (p *Path) Get(current interface{}) interface{} {
	c := make(Path, len(*p) - 1)
	// Copy everything but the first element (variable name)
	copy(c, (*p)[1:])

	for _, e := range c {
		property := false
		switch e.(type) {
		case string:
			property = true
		}

		if current == nil {
			break
		} else {
			switch current.(type) {
			case map[string]interface{}:
				var key string
				obj := current.(map[string]interface{})

				if !property {
					// If the current path is a number index then we will sort the keys of the current value 
					// lexicographically to find correct key
					var ok bool
					if key, ok = firstKey(obj, e.(int)); !ok {
						current = nil
						continue
					}
				} else {
					key = e.(string)
				}

				current = obj[key]
			case []interface{}:
				if property {
					current = nil
				} else {
					arr := current.([]interface{})
					idx := e.(int)
					// If the absolute value of the idx is greater than the length of the array then we set current to 
					// nil
					if abs(idx) >= len(arr) {
						current = nil
					} else {
						idx = mod(idx, len(arr))
						current = arr[idx]
					}
				}
			default:
				current = nil
			}
		}
	}
	return current
}

// String returns the string representation of the Path. Which is in JSONPath format.
func (p *Path) String() string {
	var b str.Builder
	for i, path := range *p {
		switch path.(type) {
		case string:
			if i != 0 {
				b.WriteString(".")
			}
			b.WriteString(path.(string))
		case int:
			b.WriteString(fmt.Sprintf("[%d]", path.(int)))
		default:
			panic(fmt.Errorf("path element should not be of type %s", reflect.TypeOf(path).String()))
		}
	}
	return b.String()
}

// Pathable defines a structure which can be converted recursively into a path.
type Pathable interface {
	Convert(vm VM) (err error, path Path)
}

// Convert will convert a JSONPath AST node into a Path which can subsequently be used to Set and Get values from a 
// value.
func (j *JSONPath) Convert(vm VM) (err error, path Path) {
	path = make(Path, 0)
	for _, p := range j.Parts {
		var subPath Path
		err, subPath = p.Convert(vm)
		if err != nil {
			return err, nil
		}
		path = append(path, subPath...)
	}
	return nil, path
}

// Convert will convert a Part AST node into a Path.
func (p *Part) Convert(vm VM) (err error, path Path) {
	path = make(Path, 0)
	path = append(path, *p.Property)
	for _, i := range p.Indices {
		var idx *data.Value
		err, idx = i.Eval(vm)
		switch idx.Type {
		case data.Number:
			path = append(path, int(idx.Value.(float64)))
		case data.String:
			path = append(path, idx.Value.(string))
		default:
			// Otherwise, we try to cast it to a Number then a String
			var idxCast *data.Value
			err, idxCast = eval.Cast(idx, data.Number)
			if err != nil {
				err, idxCast = eval.Cast(idx, data.String)
				if err != nil {
					return err, nil
				}
				path = append(path, idxCast.Value.(string))
				continue
			}
			path = append(path, int(idxCast.Value.(float64)))
		}
	}
	return nil, path
}
