package parser

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
	"reflect"
	"strconv"
	"strings"
)

func init() {
	builtins = map[string]BuiltinFunction{
		"print": func(vm VM, uncomputedArgs ...*Expression) (err error, value *data.Value) {
			var args []*data.Value
			if err, args = computeArgs(vm, uncomputedArgs...); err != nil {
				return err, nil
			}

			var b strings.Builder
			for i, arg := range args {
				if arg.Type == data.String {
					b.WriteString(arg.StringLit())
				} else {
					b.WriteString(arg.String())
				}
				if i != len(args) - 1 {
					b.WriteString(" ")
				}
			}
			_, err = fmt.Fprintf(vm.GetStdout(), "%s\n", b.String())

			return err, &data.Value{
				Value:    nil,
				Type:     data.Null,
				Global:   false,
				ReadOnly: false,
			}
		},
		"free": func(vm VM, uncomputedArgs ...*Expression) (err error, value *data.Value) {
			// For each uncomputed arg, we will find if there is a JSONPathFactor terminal contained in the left-hand 
			// side of the expression.
			for _, uncomputedArg := range uncomputedArgs {
				// We look down the left-hand side of the expression tree and see if the terminal factor is a 
				// JSONPathFactor. We will only consider this JSONPathFactor 
				var t term = uncomputedArg
				var e evalNode
				for {
					// We first get the evalNode from the left() method of the current argument.
					e = t.left()
					if t == nil {
						// We return an error if we cannot continue down the left path before finding a JSONPathFactor terminal.
						return errors.InvalidOperation.Errorf(vm, "builtin:delete", fmt.Sprintf("non-JSONPath value: \"%s\"", uncomputedArg.String(0)), "delete"), nil
					} else {
						// We do a type switch for the evalNode to find out if the underlying type is a JSONPathFactor.
						// If so we can stop iteration. Otherwise, we cast the evalNode to a term interface and assign
						// the t var.
						stop := false
						switch e.(type) {
						case *JSONPathFactor:
							stop = true
						case *Null, *Boolean, *JSON, *FunctionCall, *MethodCall, *Expression, *struct { protoEvalNode }:
							// We found an expression terminal/factor before finding a JSONPathFactor.
							return errors.InvalidOperation.Errorf(vm, "builtin:delete", fmt.Sprintf("non-JSONPath value: \"%s\"", e.(ASTNode).String(0)), "delete"), nil
						default:
							break
						}
						if stop {
							break
						}
						t = e.(term)
					}
				}

				// e will be a *JSONPathFactor value
				jsonPathFactor := e.(*JSONPathFactor)
				// We will only free the variable if the JSONPathFactor's root property is pointing to a variable...
				if jsonPathFactor.RootProperty != nil {
					// We convert the JSONPathFactor to a Path
					var path Path
					if err, path = jsonPathFactor.Convert(vm); err != nil {
						return err, nil
					}
					// We get the name of the variable as well as the value of the variable.
					variableName := path[0].(string)
					heap := vm.GetCallStack().Current().GetHeap()

					if len(path) > 1 {
						variableVal := heap.Get(variableName)
						if variableVal == nil {
							variableVal = &data.Value{
								Value:  nil,
								Type:   data.Null,
							}
						}

						// We set the value at the path to nil if we have more than element in the path.
						if err, variableVal.Value = path.Set(vm, variableVal.Value, nil); err != nil {
							return err, nil
						}
						if err = heap.Assign(variableName, variableVal.Value, variableVal.Global, variableVal.ReadOnly); err != nil {
							return err, nil
						}
					} else {
						// If we only have one element then we will delete the Value from the heap.
						heap.Delete(variableName)
					}
				}
			}
			return nil, &data.Value{
				Value:    nil,
				Type:     data.Null,
				Global:   false,
				ReadOnly: false,
			}
		},
		"find": func(vm VM, uncomputedArgs ...*Expression) (err error, value *data.Value) {
			return findBuiltin(vm, false, false, uncomputedArgs...)
		},
		"find_all": func(vm VM, uncomputedArgs ...*Expression) (err error, value *data.Value) {
			return findBuiltin(vm, true, true, uncomputedArgs...)
		},
		"find_all_parents": func(vm VM, uncomputedArgs ...*Expression) (err error, value *data.Value) {
			return findBuiltin(vm, true, false, uncomputedArgs...)
		},
	}
}

func findBuiltin(vm VM, all bool, deepest bool, uncomputedArgs... *Expression) (err error, value *data.Value) {
	var args []*data.Value
	if err, args = computeArgs(vm, uncomputedArgs...); err != nil {
		return err, nil
	}

	// We insert a value to search if there are no arguments
	if len(args) == 0 {
		args = append(args, &data.Value{
			Value: nil,
			Type:  data.Null,
		})
	}

	defer func() {
		if p := recover(); p != nil {
			err = errors.UpdateError(err, vm)
		}
	}()

	var search interface{}
	results := make([]interface{}, 0)
	for i, arg := range args {
		if i == 0 {
			// The first argument is the value to search
			search = arg.Value
		} else {
			// We cast each of the rest of the arguments to an Objects.
			if err, arg = eval.Cast(arg, data.Object); err != nil {
				return errors.UpdateError(err, vm), nil
			}

			// Then find the search params within the value
			_, argResults := find(search, arg.Map(), false, all, deepest, 0)
			results = append(results, argResults...)
		}
	}

	return nil, &data.Value{
		Value: results,
		Type:  data.Array,
	}
}

// find will find the given search schema within the given interface to search for. The behaviour of this recursive 
// method depends on the current type of the given interface to search and the given search schema, as well as the 
// partialSchema, all, and deepest flags.
//
// - If we are currently searching an Object or an Array, and the search schema is an Object, then we will iterate 
// over the key-value pairs within the searched interface and do three things.
//
//   - Try and find a deeper match by passing down the search schema to the currently iterated value.
//
//   - If an empty string key exists in the search schema then we will find a deeper match by passing down the value of
// the empty string key in the search schema to the currently iterated value. The empty string key effectively acts as 
// a wildcard match for key-value pairs in an Object and elements in an Array, in that it will always match on each.
//
//   - If the currently iterated key exists within the search schema, then we will find a deeper match by passing down 
// the value of the currently iterated key in the search schema to the currently iterated value.
//
//   - If all is not set then we will exit this loop as soon as we have found a value as this will be the deepest match.
//
//   - If all the keys within the search schema have been found then the foundAll return value is set to true and the 
// results will be set to a singleton containing the currently searched Object or Array.
//
// - If we are currently searching any other type, and the search schema is of any type, or we are currently searching 
// an Array, and the search schema is an Array. We will check if the searched value is equal to the search schema. If 
// we are currently searching a String and the current search schema is a String, then we will check if the search 
// schema is a substring of the searched value. Both strings will be converted to lowercase, so this will be a caseless 
// match.
//
// The all flag will find all the matches to the search schema within the searched value. This will include parents, 
// and parents of parents if the deepest flag is not set. If the deepest flag is not set, then the results returned by 
// find will be ordered by depth and left-most first.
//
// The partialSchema flag should be set to false initially. Internally, it indicates whether the search schema passed 
// down in a recursive call was the full schema given to the find call initially.
//
// find will return whether all the keys within the search schema are found, as well as all the nodes which match the 
// given search schema.
func find(search interface{}, searchSchema interface{}, partialSchema bool, all bool, deepest bool, indent int) (foundAll bool, results []interface{}) {
	results = nil
	//t := tabs(indent)
	var subResults []interface{}

	// We cache the types of both the search and the search schema.
	var searchType data.Type
	if err := searchType.Get(search); err != nil {
		panic(err)
	}

	var searchSchemaType data.Type
	if err := searchSchemaType.Get(searchSchema); err != nil {
		panic(err)
	}

	// Checks if subResults is not nil and doesn't have 0 elements
	checkSubResults := func() bool {
		return subResults != nil && len(subResults) > 0
	}

	// Adds the given val to the results return value. Will also create the results array if necessary. If the value is 
	// a []interface{}, then it will be unwrapped first.
	addResults := func(val interface{}) {
		if results == nil {
			results = make([]interface{}, 0)
		}

		//fmt.Println(tabs(indent + 1), indent, "adding", val, "to results")
		switch val.(type) {
		case []interface{}:
			results = append(results, val.([]interface{})...)
		default:
			results = append(results, val)
		}
	}

	iterateSearch := func(searchSchemaMap map[string]interface{}) {
		if err, iterator := data.Iterate(&data.Value{
			Value: search,
			Type:  searchType,
		}); err != nil {
			panic(err)
		} else {
			foundInCurrent := 0
			foundAllNow := foundAll
			for iterator.Len() > 0 {
				node := iterator.Next()
				key, val := node.Key, node.Val
				//fmt.Printf("%s\t%d: checking %s: %v\n", t, indent, key, val)

				// We always iterate down each subtree with the untouched search schema first. This is so we can find 
				// the deepest match. If we found all the keys in the lower branch then we will:
				// If fetching all nodes without deepest match
				if foundAllNow, subResults = find(val.Value, searchSchema, partialSchema, all, deepest, indent + 2); checkSubResults() && foundAllNow {
					foundAll = foundAll || foundAllNow
					if all && !deepest && !partialSchema {
						foundInCurrent ++
					}
					//fmt.Printf("%s\t%d: EXISTS = FALSE \"%v\" exists in \"%v\" subresults \"%v\" %t %t\n", t, indent, searchSchema, val.Value, subResults, foundAllNow, partialSchema)
					if (all && !partialSchema) || !all {
						addResults(subResults)
					}
					if !all { break }
				}

				// If a key of the empty string exists in the search schema then we will search the lower subtrees for 
				// the empty string value of the search schema. This is effectively a "veto" for the current level.
				if _, ok := searchSchemaMap[""]; ok {
					//fmt.Printf("%s\t%d: empty string is inside searchSchema, recursing down \"%v\", with schema \"%v\"\n", t, indent, val.Value, searchSchemaMap[""])
					if foundAllNow, subResults = find(val.Value, searchSchemaMap[""], true, all, deepest, indent+2); checkSubResults() {
						//fmt.Printf("%s\t%d: EMPTY STRING \"%v\" exists in \"%v\" subresults \"%v\" %t\n", t, indent, searchSchemaMap[""], val.Value, subResults, foundAllNow)
						foundAll = foundAll || foundAllNow
						foundInCurrent++
						if foundAllNow && !all {
							addResults(subResults)
						}
						if foundInCurrent == len(searchSchemaMap) && !all {
							break
						}
					}
				}

				var exists bool
				var existsVal interface{}
				switch key.Value.(type) {
				case string:
					existsVal, exists = searchSchemaMap[key.StringLit()]
				case float64:
					existsVal, exists = searchSchemaMap[strconv.Itoa(key.Int())]
				}

				// If the current key is within the search schema, then we will recurse down the value of the key 
				// with the value of the search schema.
				if exists {
					//fmt.Printf("%s\t%d: %s exists within searchSchema\n", t, indent, key)
					if foundAllNow, subResults = find(val.Value, existsVal, true, all, deepest, indent+2); checkSubResults() {
						//fmt.Printf("%s\t%d: EXISTS = TRUE \"%v\" exists \"%v\" subresults \"%v\" %t %t\n", t, indent, existsVal, val.Value, subResults, foundAllNow, partialSchema)
						// If we found all the keys in this subtree then we will add the results to the current frame's
						// results.
						foundAll = foundAll || foundAllNow
						foundInCurrent++
						if foundAllNow && !all {
							addResults(subResults)
						}
						if foundInCurrent == len(searchSchemaMap) && !all {
							break
						}
					}
				}

				//fmt.Println(tabs(indent + 1), indent, "END OF LOOP checkSubResults", checkSubResults(), "results", len(results), "foundInCurrent", foundInCurrent, "searchSchemaMap", len(searchSchemaMap), "partialSchema", partialSchema, "foundAll", foundAll)
			}

			//fmt.Println(tabs(indent + 1), indent, "OUT OF LOOP checkSubResults", checkSubResults(), "results", len(results), "foundInCurrent", foundInCurrent, "searchSchemaMap", len(searchSchemaMap), "partialSchema", partialSchema, "foundAll", foundAll)
			// If we have found all the keys in searchSchemaMap then we will add the current node to the results.
			if !all {
				if foundInCurrent == len(searchSchemaMap) {
					//fmt.Println(tabs(indent+1), indent, "found all in current", foundInCurrent, len(searchSchemaMap))
					results = []interface{}{search}
					foundAll = true
				}
			} else {
				if foundInCurrent == len(searchSchemaMap) {
					addResults(search)
					foundAll = true
				}
			}
		}
	}

	//fmt.Printf("%s%d: trying to find \"%v\" in \"%v\" (pSchema = %t)\n", t, indent, searchSchema, search, partialSchema)

	switch searchType {
	case data.Object:
		if searchSchemaType == data.Object {
			iterateSearch(searchSchema.(map[string]interface{}))
		}
	case data.Array:
		switch searchSchemaType {
		case data.Object:
			iterateSearch(searchSchema.(map[string]interface{}))
		case data.Array:
			if err, equal := eval.EqualInterface(search, searchSchema); err != nil {
				panic(err)
			} else if equal {
				addResults(search)
			}
		}
	case data.String:
		if searchSchemaType == data.String {
			// We do a caseless contains to check if the string we are searching for is contained within the search 
			// string.
			//fmt.Printf("%s\t%d: %s in %s = %v\n", t, indent, strings.ToLower(searchSchema.(string)), strings.ToLower(search.(string)), strings.Contains(strings.ToLower(search.(string)), strings.ToLower(searchSchema.(string))))
			if strings.Contains(strings.ToLower(search.(string)), strings.ToLower(searchSchema.(string))) {
				addResults(search)
			}
		}
	default:
		if err, same := eval.EqualInterface(search, searchSchema); err != nil {
			panic(err)
		} else if same {
			addResults(search)
		}
	}
	//fmt.Println(t, indent, "results =", results)
	return foundAll, results
}

// MarshalJSON is used for marshalling FunctionDefinitions to JSON strings as they appear in the data.Heap. The returned
// byte string is in the format:
//  "function:RAW_JSON_PATH:FUNCTION_DEF_UINTPTR"
func (f *FunctionDefinition) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"function:%s:%d\"", f.JSONPath.String(0), reflect.ValueOf(f).Pointer())), nil
}

// BuiltinFunction denotes the signature of each builtin function in the builtins table.
type BuiltinFunction func(vm VM, uncomputedArgs... *Expression) (err error, value *data.Value)

// String is used for marshalling BuiltinFunctions to strings and JSON strings. The name of the builtin will be found 
// by using reflection to get the pointer to the builtin function. The returned string is in the format:
//  "builtin:NAME:BUILTIN_FUNCTION_UINTPTR"
func (b BuiltinFunction) String() string {
	ptr := reflect.ValueOf(b).Pointer()
	var name string; var v BuiltinFunction
	for name, v = range builtins {
		currPtr := reflect.ValueOf(v).Pointer()
		if ptr == currPtr {
			break
		}
	}
	return fmt.Sprintf("\"builtin:%s:%d\"", name, ptr)
}

func (b BuiltinFunction) MarshalJSON() ([]byte, error) {
	return []byte(b.String()), nil
}

func computeArgs(vm VM, uncomputedArgs... *Expression) (err error, args []*data.Value) {
	// Evaluate arguments and create a list of args
	args = make([]*data.Value, len(uncomputedArgs))
	for i, arg := range uncomputedArgs {
		var computed *data.Value
		if err, computed = arg.Eval(vm); err != nil {
			return err, nil
		}
		args[i] = computed
	}
	return nil, args
}

// builtins contains all builtins in sttp. All builtins take a list of uncomputed arguments. These are uncomputed as 
// there might be special use cases.
var builtins map[string]BuiltinFunction

// CheckBuiltin will check if the function of the given name exists as a builtin.
func CheckBuiltin(name string) bool {
	_, ok := builtins[name]
	return ok
}

// GetBuiltin will return the builtin function encapsulated in a data.Value.
func GetBuiltin(name string) *data.Value {
	if CheckBuiltin(name) {
		return &data.Value{
			Value:    builtins[name],
			Type:     data.Function,
			Global:   true,
			ReadOnly: true,
		}
	}
	return nil
}