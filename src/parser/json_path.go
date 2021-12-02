package parser

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
)

const DefaultObjectKey = " "

type Path []interface{}

func set(current interface{}, to interface{}, path *Path) interface{} {
	if len(*path) > 0 {
		p := (*path)[0]
		*path = (*path)[1:]
		property := false
		switch p.(type) {
		case string:
			property = true
		default:
			break
		}

		if current == nil {
			if property {
				// If the path is a property we construct a new object
				current = set(make(map[string]interface{}), to, path)
			}
			// Otherwise, we create a new interface array
			current = set(make([]interface{}, 0), to, path)
		} else {
			switch current.(type) {
			case map[string]interface{}:
				if !property {
					panic(errors.JSONPathError.Errorf("object", "index"))
				}

				obj := current.(map[string]interface{})
				key := p.(string)
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
				if idx >= 0 && idx < len(arr) {
					current.([]interface{})[idx] = set(arr[idx], to, path)
				} else {
					// We insert nils up until we get to the index to set at, at which point we recurse.
					for i := len(arr); i <= idx; i ++ {
						var val interface{} = nil
						if i == idx {
							val = set(nil, to, path)
						}
						arr = append(arr, val)
					}
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
					arr := make([]interface{}, idx + 2)
					arr[idx + 1] = current
					for i := 0; i < len(arr) - 1; i ++ {
						var val interface{} = nil
						if i == idx {
							val = set(nil, to, path)
						}
						arr[i] = val
					}
					current = arr
				}
			}
		}
	}
	return current
}

func (p *Path) Set(current interface{}, to interface{}) (err error, new interface{}) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()

	c := make(Path, len(*p) - 1)
	// Copy everything but the first element (variable name)
	copy(c, (*p)[1:])
	new = set(current, to, &c)
	return err, new
}

func (p *Path) Get() (err error, get interface{}) {
	return nil, nil
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
	path = append(path, p.Property)
	for _, i := range p.Indices {
		var idx *data.Symbol
		err, idx = i.Eval(vm)
		switch idx.Type {
		case data.Number:
			path = append(path, int(idx.Value.(float64)))
		case data.String:
			path = append(path, idx.Value.(string))
		default:
			// Otherwise, we try to cast it to a Number then a String
			var idxCast *data.Symbol
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
