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

const (
	DefaultObjectKey = ""
	CurrentNodeVariableName = "curr"
)

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

func set(vm VM, current interface{}, to interface{}, path *Path) interface{} {
	if len(*path) > 0 {
		p := (*path)[0]
		property := false
		filter := false
		switch p.(type) {
		case string:
			property = true
		case *Block:
			filter = true
		}

		if current == nil {
			// We don't pop the next element from the path because we need to use this frame to construct a new value 
			// and interface then we can use the next called frame to set the property of index.
			if property {
				// If the path is a property we construct a new object
				current = set(vm, make(map[string]interface{}), to, path)
			} else {
				// Otherwise, we create a new interface array. This will also be the case for a filter.
				current = set(vm, make([]interface{}, 0), to, path)
			}
		} else {
			// We only remove the current path element if we are currently on a non-nil value.
			*path = (*path)[1:]
			var heap *data.Heap
			var existingSelf *data.Value
			var err error

			filterSetup := func(block *Block, toIterate *data.Value) (it *data.Iterator) {
				heap = vm.GetCallStack().Current().GetHeap()
				existingSelf = heap.Get(CurrentNodeVariableName)

				if err, it = data.Iterate(toIterate); err != nil {
					panic(err)
				}
				return it
			}

			filterNodeEval := func(node *data.Element) bool {
				if err = heap.Assign(CurrentNodeVariableName, map[string]interface{} {
					"key": node.Key.Value,
					"value": node.Val.Value,
				}, false, false); err != nil {
					panic(err)
				}

				var result *data.Value
				// We evaluate the block
				if err, result = p.(*Block).Eval(vm); err != nil {
					switch err.(type) {
					case errors.PurposefulError:
						// If we have a purposeful error then we will check if it is Return. If so we will set err to nil.
						if err.(errors.PurposefulError) == errors.Return {
							err = nil
							break
						}
						panic(err)
					default:
						panic(err)
					}
				}

				// We cast the result into a boolean
				if err, result = eval.Cast(result, data.Boolean); err != nil {
					panic(err)
				}
				return result.Value.(bool)
			}

			filterTeardown := func() {
				if existingSelf == nil {
					// If there was no existing self then we will just delete the variable
					heap.Delete(CurrentNodeVariableName)
				} else {
					// Assign will never throw an error as the existing self value would've been assigned before this.
					_ = heap.Assign(CurrentNodeVariableName, existingSelf.Value, existingSelf.Global, existingSelf.ReadOnly)
				}
			}

			switch current.(type) {
			case map[string]interface{}:
				obj := current.(map[string]interface{})
				if !filter {
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
					current.(map[string]interface{})[key] = set(vm, val, to, path)
				} else {
					it := filterSetup(p.(*Block), &data.Value{
						Value: obj,
						Type:  data.Object,
					})
					for it.Len() > 0 {
						// We get the next node and assign the key and the value to the self variable
						node := it.Next()
						// If the result is truthy then we will recurse down the current subtree.
						if filterNodeEval(node) {
							current.(map[string]interface{})[node.Key.StringLit()] = set(vm, node.Val.Value, to, path)
						}
					}
					filterTeardown()
				}
			case []interface{}:
				arr := current.([]interface{})
				if !filter {
					if property {
						panic(errors.JSONPathError.Errorf("array", "property"))
					}

					idx := p.(int)
					if abs(idx) < len(arr) {
						idx = mod(idx, len(arr))
						current.([]interface{})[idx] = set(vm, arr[idx], to, path)
					} else {
						if idx < 0 {
							panic(errors.JSONPathError.Errorf("array", fmt.Sprintf("negative index that is out of array bounds (%d)", idx)))
						}

						// We insert nils up until we get to the index to set at, at which point we recurse.
						for i := len(arr); i <= idx; i++ {
							var val interface{} = nil
							if i == idx {
								val = set(vm, nil, to, path)
							}
							arr = append(arr, val)
						}
						current = arr
					}
				} else {
					it := filterSetup(p.(*Block), &data.Value{
						Value: arr,
						Type:  data.Array,
					})
					for it.Len() > 0 {
						// We get the next node and assign the key and the value to the self variable
						node := it.Next()
						// If the result is truthy then we will recurse down the current subtree.
						if filterNodeEval(node) {
							current.([]interface{})[node.Key.Int()] = set(vm, node.Val.Value, to, path)
						}
					}
					filterTeardown()
				}
			default:
				if property {
					// If accessing a property then we will wrap the value in an object, assigning it to the key 
					// DefaultObjectKey
					obj := make(map[string]interface{})
					obj[DefaultObjectKey] = current
					obj[p.(string)] = set(vm, nil, to, path)
					current = obj
				} else {
					// We have a special case for string access, but we keep the default logic of creating a sparse map
					// when we are accessing a property (above).
					switch current.(type) {
					case string:
						st := current.(string)
						if !filter {
							idx := p.(int)
							var char string
							// Find the character to recurse down
							if abs(idx) < len(st) {
								idx = mod(idx, len(st))
								char = string(current.(string)[idx])
							} else if idx < 0 {
								// If the absolute value of the idx is greater than the length of the string then we return an error
								panic(errors.JSONPathError.Errorf("string", fmt.Sprintf("negative index that is out of string bounds (%d)", idx)))
							} else {
								char = " "
							}

							// We iterate down the given character
							s := set(vm, char, to, path)
							var t data.Type
							if err = t.Get(s); err != nil {
								panic(err)
							}

							// Convert the returned value to a string
							var value *data.Value
							if err, value = eval.Cast(&data.Value{
								Value: s,
								Type:  t,
							}, data.String); err != nil {
								panic(err)
							}

							if idx >= len(st) {
								// We construct a string builder and write the existing string to that builder
								var b str.Builder
								b.WriteString(st)
								// We insert whitespace up until we get to the index to set at, at which point we recurse.
								for i := len(st); i <= idx; i++ {
									if i == idx {
										b.WriteString(value.StringLit())
									} else {
										b.WriteString(" ")
									}
								}
								current = b.String()
							} else {
								current = current.(string)[:idx] + value.StringLit() + current.(string)[idx+1:]
							}
						} else {
							it := filterSetup(p.(*Block), &data.Value{
								Value: current,
								Type:  data.String,
							})
							replacementStrings := make([]string, 0)
							replacementIndices := make([]int, 0)
							for it.Len() > 0 {
								// We get the next node and assign the key and the value to the self variable
								node := it.Next()
								// If the result is truthy then we will recurse down the current subtree.
								if filterNodeEval(node) {
									var value *data.Value
									if err, value = eval.CastInterface(set(vm, node.Val.Value, to, path), data.String); err != nil {
										panic(err)
									}
									// Add the current node's index and value to the replacementIndices and 
									// replacementStrings arrays respectively.
									replacementStrings = append(replacementStrings, value.StringLit())
									replacementIndices = append(replacementIndices, node.Key.Int())
								}
							}
							// We set the string at the end, so we don't interfere with the loop above
							current = strings.ReplaceCharIndex(current.(string), replacementIndices, replacementStrings...)
							filterTeardown()
						}
					default:
						// If accessing an index then we will create an array with p.(int) + 2 spaces. Recursing down the 
						// p.(int) space and inserting the existing value in the p.(int) + 1 space.
						idx := p.(int)
						if idx >= 0 {
							arr := make([]interface{}, idx+2)
							arr[idx+1] = current
							for i := 0; i < len(arr)-1; i++ {
								var val interface{} = nil
								if i == idx {
									val = set(vm, nil, to, path)
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
		}
	} else {
		current = to
	}
	return current
}

func (p *Path) Set(vm VM, current interface{}, to interface{}) (err error, new interface{}) {
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
	new = set(vm, current, to, &c)
	return err, new
}

// Get iterates over the referred to Path and descends the given value and returns the value pointed to by the path. 
// If the path doesn't lead anywhere it will return nil.
func (p *Path) Get(vm VM, current interface{}) (err error, gotten interface{}) {
	defer func() {
		if p := recover(); p != nil {
			switch p.(type) {
			case struct { errors.ProtoSttpError }:
				err = p.(struct { errors.ProtoSttpError })
			default:
				err = fmt.Errorf("%v", p)
			}
			gotten = nil
		}
	}()

	c := make(Path, len(*p) - 1)
	// Copy everything but the first element (variable name)
	copy(c, (*p)[1:])

	filterIterator := func(block *Block, toIterate *data.Value) []interface{} {
		heap := vm.GetCallStack().Current().GetHeap()
		existingSelf := heap.Get(CurrentNodeVariableName)
		filtered := make([]interface{}, 0)

		var it *data.Iterator
		if err, it = data.Iterate(toIterate); err != nil {
			panic(err)
		}

		for it.Len() > 0 {
			// We get the next node and assign the key and the value to the self variable
			node := it.Next()
			if err = heap.Assign(CurrentNodeVariableName, map[string]interface{} {
				"key": node.Key.Value,
				"value": node.Val.Value,
			}, false, false); err != nil {
				panic(err)
			}

			var result *data.Value
			// We evaluate the block
			if err, result = block.Eval(vm); err != nil {
				switch err.(type) {
				case errors.PurposefulError:
					// If we have a purposeful error then we will check if it is Return. If so we will set err to nil.
					if err.(errors.PurposefulError) == errors.Return {
						err = nil
						break
					}
					panic(err)
				default:
					panic(err)
				}
			}

			// We cast the result into a boolean
			if err, result = eval.Cast(result, data.Boolean); err != nil {
				panic(err)
			}

			// If the result is truthy then we will append the current node's value onto the filtered array.
			if result.Value.(bool) {
				filtered = append(filtered, node.Val.Value)
			}
		}

		if existingSelf == nil {
			// If there was no existing self then we will just delete the variable
			heap.Delete(CurrentNodeVariableName)
		} else {
			// Assign will never throw an error as the existing self value would've been assigned before this.
			_ = heap.Assign(CurrentNodeVariableName, existingSelf.Value, existingSelf.Global, existingSelf.ReadOnly)
		}
		return filtered
	}

	for _, e := range c {
		property := false
		filter := false
		switch e.(type) {
		case string:
			property = true
		case *Block:
			filter = true
		}

		if current == nil {
			break
		} else {
			switch current.(type) {
			case map[string]interface{}:
				var key string
				obj := current.(map[string]interface{})

				if filter {
					// If the current path is a filter, then we will filter the current object and break out the switch
					current = filterIterator(e.(*Block), &data.Value{
						Value: obj,
						Type:  data.Object,
					})
					break
				} else if !property {
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
				} else if filter {
					current = filterIterator(e.(*Block), &data.Value{
						Value: current.([]interface{}),
						Type:  data.Array,
					})
					break
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
			case string:
				// String access is pretty much the same as array access
				if property {
					current = nil
				} else if filter {
					current = filterIterator(e.(*Block), &data.Value{
						Value: current.(string),
						Type:  data.String,
					})
					break
				} else {
					str := current.(string)
					idx := e.(int)
					// If the absolute value of the idx is greater than the length of the string then we set current to 
					// nil
					if abs(idx) >= len(str) {
						current = nil
					} else {
						idx = mod(idx, len(str))
						current = str[idx]
					}
				}
			default:
				current = nil
			}
		}
	}
	return nil, current
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
		switch {
		case i.ExpressionIndex != nil:
			var idx *data.Value
			err, idx = i.ExpressionIndex.Eval(vm)
			switch idx.Type {
			case data.Number:
				path = append(path, int(idx.Value.(float64)))
			case data.String:
				path = append(path, idx.StringLit())
			default:
				// Otherwise, we try to cast it to a Number then a String
				var idxCast *data.Value
				err, idxCast = eval.Cast(idx, data.Number)
				if err != nil {
					err, idxCast = eval.Cast(idx, data.String)
					if err != nil {
						return err, nil
					}
					path = append(path, idxCast.StringLit())
					continue
				}
				path = append(path, int(idxCast.Value.(float64)))
			}
		case i.FilterIndex != nil:
			path = append(path, i.FilterIndex)
		}
	}
	return nil, path
}