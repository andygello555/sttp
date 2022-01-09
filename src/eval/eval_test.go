package eval

import (
	"encoding/json"
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

const (
	EchoChamberCmd    = "node"
	EchoChamberSource = "../_examples/echo_chamber/main.js"
)

func TestCompute(t *testing.T) {
	for testNo, test := range []struct{
		op1 *data.Value
		op2 *data.Value
		operator Operator
		result *data.Value
		err error
	}{
		// Unsupported operation
		{
			op1: &data.Value{
				Value: nil,
				Type:  data.Object,
				Global: false,
			},
			op2: &data.Value{
				Value: nil,
				Type:  data.Null,
				Global: false,
			},
			operator: Mul,
			result: &data.Value{
				Value: nil,
				Type:  0,
				Global: false,
			},
			err: errors.InvalidOperation.Errorf(errors.GetNullVM(), "*", "object", "null"),
		},

		// Array manipulation

		{
			op1: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			op2: &data.Value{
				Value: 4,
				Type:  data.Number,
				Global: false,
			},
			operator: Add,
			result: &data.Value{
				Value: []interface{}{1, 2, 3, 4},
				Type:  data.Array,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			op2: &data.Value{
				Value: map[string]interface{}{"a": 1, "b": 2},
				Type:  data.Object,
				Global: false,
			},
			operator: Add,
			result: &data.Value{
				Value: []interface{}{1, 2, 3, map[string]interface{}{"a": 1, "b": 2}},
				Type:  data.Array,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{4, 5, 6, 7},
				Type:  data.Array,
				Global: false,
			},
			operator: Add,
			result: &data.Value{
				Value: []interface{}{1, 2, 3, 4, 5, 6, 7},
				Type:  data.Array,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: []interface{}{1, 2, 3, 4},
				Type:  data.Array,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{5, 6, 7},
				Type:  data.Array,
				Global: false,
			},
			operator: Add,
			result: &data.Value{
				Value: []interface{}{1, 2, 3, 4, 5, 6, 7},
				Type:  data.Array,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: []interface{}{[]interface{}{1, 2, 3}, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			op2: &data.Value{
				Value: nil,
				Type:  data.Null,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: []interface{}{2, 3},
				Type:  data.Array,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: []interface{}{[]interface{}{1, 2, 3}, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{[]interface{}{1, 2, 3}},
				Type:  data.Array,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: []interface{}{2, 3},
				Type:  data.Array,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{nil, 2},
				Type:  data.Array,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: []interface{}{3},
				Type:  data.Array,
				Global: false,
			},
			err: nil,
		},

		// Object manipulation

		{
			op1: &data.Value{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3},
				Type:  data.Object,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			operator: Div,
			result: &data.Value{
				Value: map[string]interface{}{"0": 1},
				Type:  data.Object,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3},
				Type:  data.Object,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: map[string]interface{}{"3": 3},
				Type:  data.Object,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3},
				Type:  data.Object,
				Global: false,
			},
			op2: &data.Value{
				Value: "{\"hello\":\"world\"}",
				Type:  data.String,
				Global: false,
			},
			operator: Add,
			result: &data.Value{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3, "hello": "world"},
				Type:  data.Object,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3},
				Type:  data.Object,
				Global: false,
			},
			op2: &data.Value{
				Value: 4,
				Type:  data.Number,
				Global: false,
			},
			operator: Add,
			result: &data.Value{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3, "4": nil},
				Type:  data.Object,
				Global: false,
			},
			err: nil,
		},

		// String manipulation

		{
			op1: &data.Value{
				Value: "abc",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: map[string]interface{}{"a": float64(1), "b": float64(2), "c": float64(3)},
				Type:  data.Object,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: "123",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "moomoo cow is here",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{"moo", "is"},
				Type:  data.Array,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: " cow  here",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "123456",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: float64(3),
				Type:  data.Number,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: "123",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "moomoocow",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: "moo",
				Type:  data.String,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: "cow",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "is null nullable?",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: nil,
				Type:  data.Null,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: "is  able?",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "Result is: ",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			operator: Add,
			result: &data.Value{
				Value: "Result is: [1,2,3]",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "Result is: [%%, %%, %%]",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			operator: Mod,
			result: &data.Value{
				Value: "Result is: [1, 2, 3]",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "Result is: [%%]",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: map[string]interface{}{"1": float64(1)},
				Type:  data.Object,
				Global: false,
			},
			operator: Mod,
			result: &data.Value{
				Value: "Result is: [1]",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "Result is: [%%]",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: "nothing",
				Type:  data.String,
				Global: false,
			},
			operator: Mod,
			result: &data.Value{
				Value: "Result is: [nothing]",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: float64(1),
				Type:  data.Number,
			},
			op2: &data.Value{
				Value: "1",
				Type:  data.String,
			},
			operator: Add,
			result: &data.Value{
				Value: 2,
				Type:  data.Number,
			},
			err: nil,
		},
	}{
		var ok bool
		err, result := Compute(test.operator, test.op1, test.op2)
		// Check if the actual result is Equal to the expected result only if there is no error.
		if err == nil {
			err, ok = Equal(result, test.result)
		}

		if testing.Verbose() && result != nil {
			fmt.Printf("%d: %v %s %v = %v\n", testNo + 1, test.op1.String(), test.operator.String(), test.op2.String(), result.String())
		}

		if test.err != nil {
			if err.Error() != test.err.Error() {
				t.Errorf("error \"%s\" for testNo: %d does not match the required error: \"%s\"", err.Error(), testNo + 1, test.err.Error())
			}
		} else if !ok {
			t.Errorf("result \"%v\" for testNo: %d does not match the required result: \"%v\"", result, testNo + 1, test.result)
		}
	}
}

func TestCast(t *testing.T) {
	for testNo, test := range []struct{
		from   *data.Value
		to     data.Type
		result *data.Value
		err    error
	}{
		{
			from: &data.Value{
				Value:    "1",
				Type:     data.String,
			},
			to: data.Number,
			result: &data.Value{
				Value:    float64(1),
				Type:     data.Number,
			},
			err: nil,
		},
		{
			from: &data.Value{
				Value:    "1.23",
				Type:     data.String,
			},
			to: data.Number,
			result: &data.Value{
				Value: 1.23,
				Type:  data.Number,
			},
			err: nil,
		},
		{
			from: &data.Value{
				Value:    "1.23abc",
				Type:     data.String,
			},
			to: data.Number,
			result: &data.Value{
				Value: float64(7),
				Type:  data.Number,
			},
			err: nil,
		},
	}{
		var ok bool
		err, result := Cast(test.from, test.to)
		// Check if the actual result is Equal to the expected result only if there is no error.
		if err == nil {
			err, ok = Equal(result, test.result)
		}

		if testing.Verbose() && result != nil {
			fmt.Printf("%d: %s -(%s)-> %s\n", testNo + 1, test.from.String(), test.to.String(), result.String())
		}

		if test.err != nil {
			if err.Error() != test.err.Error() {
				t.Errorf("error \"%s\" for testNo: %d does not match the required error: \"%s\"", err.Error(), testNo + 1, test.err.Error())
			}
		} else if err != nil {
			t.Errorf("error \"%s\" should not have occurred (testNo: %d)", err.Error(), testNo + 1)
		} else if !ok {
			t.Errorf("result \"%v\" for testNo: %d does not match the required result: \"%v\"", result, testNo + 1, test.result)
		}
	}
}

func TestMethod_Call(t *testing.T) {
	// Start the echo chamber web server
	echoChamber := exec.Command(EchoChamberCmd, EchoChamberSource)
	echoChamber.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := echoChamber.Start(); err != nil {
		panic(fmt.Errorf("could not start echo chamber: \"%s\"", err.Error()))
	}
	time.Sleep(150 * time.Millisecond)

	for testNo, test := range []struct {
		args   []*data.Value
		method Method
		result []byte
		err    error
	}{
		{
			args: []*data.Value{
				{
					Value:    "http://127.0.0.1:3000/hello/world?hello=world",
					Type:     data.String,
					Global:   false,
					ReadOnly: false,
				},
			},
			method: GET,
			result: []byte(`{
	"code": null,
	"headers": {
		"accept-encoding": "gzip",
		"host": "127.0.0.1:3000",
		"user-agent": "go-resty/2.7.0 (https://github.com/go-resty/resty)"
	},
	"method": "GET",
	"query_params": {
		"hello": "world"
	},
	"url": "http://127.0.0.1:3000/hello/world?hello=world",
	"version": "1.1"
}`),
			err: nil,
		},
		{
			args: []*data.Value{
				{
					Value:    "http://127.0.0.1:3000/hello/world?hello=world",
					Type:     data.String,
				},
				{
					Value: map[string]interface{}{
						"header_1": float64(1),
						"header_2": float64(2),
					},
					Type: data.Object,
				},
			},
			method: GET,
			result: []byte(`{
	"code": null,
	"headers": {
		"accept-encoding": "gzip",
		"host": "127.0.0.1:3000",
		"user-agent": "go-resty/2.7.0 (https://github.com/go-resty/resty)",
		"header_1": "1",
		"header_2": "2"
	},
	"method": "GET",
	"query_params": {
		"hello": "world"
	},
	"url": "http://127.0.0.1:3000/hello/world?hello=world",
	"version": "1.1"
}`),
			err: nil,
		},
		{
			args: []*data.Value{
				{
					Value: "http://127.0.0.1:3000/hello/world?hello=world",
					Type: data.String,
				},
				{
					Value: map[string]interface{}{
						"header_1": float64(1),
						"header_2": float64(2),
					},
					Type: data.Object,
				},
				{
					Value: map[string]interface{} {
						"cookie": "nom nom nom",
					},
					Type: data.Object,
				},
			},
			method: GET,
			result: []byte(`{
	"code": null,
	"headers": {
		"accept-encoding": "gzip",
		"host": "127.0.0.1:3000",
		"user-agent": "go-resty/2.7.0 (https://github.com/go-resty/resty)",
		"cookie": "cookie=\"nom nom nom\"",
		"header_1": "1",
		"header_2": "2"
	},
	"method": "GET",
	"query_params": {
		"hello": "world"
	},
	"url": "http://127.0.0.1:3000/hello/world?hello=world",
	"version": "1.1"
}`),
			err: nil,
		},
		{
			args: []*data.Value{
				{
					Value: "http://127.0.0.1:3000/hello/world?hello=world",
					Type: data.String,
				},
				{
					Value: nil,
					Type: data.Null,
				},
				{
					Value: map[string]interface{} {
						"cookie": "nom nom nom",
					},
					Type: data.Object,
				},
			},
			method: GET,
			result: []byte(`{
	"code": null,
	"headers": {
		"accept-encoding": "gzip",
		"host": "127.0.0.1:3000",
		"user-agent": "go-resty/2.7.0 (https://github.com/go-resty/resty)",
		"cookie": "cookie=\"nom nom nom\""
	},
	"method": "GET",
	"query_params": {
		"hello": "world"
	},
	"url": "http://127.0.0.1:3000/hello/world?hello=world",
	"version": "1.1"
}`),
			err: nil,
		},
		{
			args: []*data.Value{
				{
					Value: "http://127.0.0.1:3000/hello/world?hello=world",
					Type: data.String,
				},
				{
					Value: map[string]interface{} {
						"hello": "world",
					},
					Type: data.Object,
				},
			},
			method: POST,
			result: []byte(`{
	"body": {
		"hello": "world"
	},
	"code": null,
	"headers": {
		"accept-encoding": "gzip",
		"content-length": "17",
		"content-type": "application/json",
		"host": "127.0.0.1:3000",
		"user-agent": "go-resty/2.7.0 (https://github.com/go-resty/resty)"
	},
	"method": "POST",
	"query_params": {
		"hello": "world"
	},
	"url": "http://127.0.0.1:3000/hello/world?hello=world",
	"version": "1.1"
}`),
			err: nil,
		},
		{
			args: []*data.Value{
				{
					Value: "http://127.0.0.1:3000?format=html",
					Type: data.String,
				},
			},
			result: []byte(`{
  "attributes": {},
  "data": "",
  "siblings": [
    {
      "attributes": {
        "lang": "en"
      },
      "data": "html",
      "siblings": [
        {
          "attributes": {},
          "data": "head",
          "siblings": [
            {
              "attributes": {},
              "data": "title",
              "siblings": [
                {
                  "attributes": {},
                  "data": "GET: http://127.0.0.1:3000/?format=html",
                  "siblings": [],
                  "type": "text"
                }
              ],
              "type": "element"
            }
          ],
          "type": "element"
        },
        {
          "attributes": {},
          "data": "body",
          "siblings": [
            {
              "attributes": {},
              "data": "h1",
              "siblings": [
                {
                  "attributes": {},
                  "data": "GET: http://127.0.0.1:3000/?format=html",
                  "siblings": [],
                  "type": "text"
                }
              ],
              "type": "element"
            },
            {
              "attributes": {},
              "data": "div",
              "siblings": [
                {
                  "attributes": {},
                  "data": "ul",
                  "siblings": [
                    {
                      "attributes": {},
                      "data": "li",
                      "siblings": [
                        {
                          "attributes": {},
                          "data": "method: GET",
                          "siblings": [],
                          "type": "text"
                        }
                      ],
                      "type": "element"
                    },
                    {
                      "attributes": {},
                      "data": "li",
                      "siblings": [
                        {
                          "attributes": {},
                          "data": "url: http://127.0.0.1:3000/?format=html",
                          "siblings": [],
                          "type": "text"
                        }
                      ],
                      "type": "element"
                    },
                    {
                      "attributes": {},
                      "data": "li",
                      "siblings": [
                        {
                          "attributes": {},
                          "data": "query_params:",
                          "siblings": [],
                          "type": "text"
                        },
                        {
                          "attributes": {},
                          "data": "ul",
                          "siblings": [
                            {
                              "attributes": {},
                              "data": "li",
                              "siblings": [
                                {
                                  "attributes": {},
                                  "data": "format: html",
                                  "siblings": [],
                                  "type": "text"
                                }
                              ],
                              "type": "element"
                            }
                          ],
                          "type": "element"
                        }
                      ],
                      "type": "element"
                    },
                    {
                      "attributes": {},
                      "data": "li",
                      "siblings": [
                        {
                          "attributes": {},
                          "data": "headers:",
                          "siblings": [],
                          "type": "text"
                        },
                        {
                          "attributes": {},
                          "data": "ul",
                          "siblings": [
                            {
                              "attributes": {},
                              "data": "li",
                              "siblings": [
                                {
                                  "attributes": {},
                                  "data": "host: 127.0.0.1:3000",
                                  "siblings": [],
                                  "type": "text"
                                }
                              ],
                              "type": "element"
                            },
                            {
                              "attributes": {},
                              "data": "li",
                              "siblings": [
                                {
                                  "attributes": {},
                                  "data": "user-agent: go-resty/2.7.0 (https://github.com/go-resty/resty)",
                                  "siblings": [],
                                  "type": "text"
                                }
                              ],
                              "type": "element"
                            },
                            {
                              "attributes": {},
                              "data": "li",
                              "siblings": [
                                {
                                  "attributes": {},
                                  "data": "accept-encoding: gzip",
                                  "siblings": [],
                                  "type": "text"
                                }
                              ],
                              "type": "element"
                            }
                          ],
                          "type": "element"
                        }
                      ],
                      "type": "element"
                    },
                    {
                      "attributes": {},
                      "data": "li",
                      "siblings": [
                        {
                          "attributes": {},
                          "data": "code: null",
                          "siblings": [],
                          "type": "text"
                        }
                      ],
                      "type": "element"
                    },
                    {
                      "attributes": {},
                      "data": "li",
                      "siblings": [
                        {
                          "attributes": {},
                          "data": "version: 1.1",
                          "siblings": [],
                          "type": "text"
                        }
                      ],
                      "type": "element"
                    }
                  ],
                  "type": "element"
                }
              ],
              "type": "element"
            }
          ],
          "type": "element"
        }
      ],
      "type": "element"
    }
  ],
  "type": "document"
}`),
			method: GET,
			err: nil,
		},
		{
			args: []*data.Value{
				{
					Value: "http://127.0.0.1:3000?format=html",
					Type:  data.String,
				},
				{
					Value: map[string]interface{} {
						"hello": "world",
					},
					Type: data.Object,
				},
			},
			result: []byte(`{
  "attributes": {},
  "data": "",
  "siblings": [
    {
      "attributes": {
        "lang": "en"
      },
      "data": "html",
      "siblings": [
        {
          "attributes": {},
          "data": "head",
          "siblings": [
            {
              "attributes": {},
              "data": "title",
              "siblings": [
                {
                  "attributes": {},
                  "data": "POST: http://127.0.0.1:3000/?format=html",
                  "siblings": [],
                  "type": "text"
                }
              ],
              "type": "element"
            }
          ],
          "type": "element"
        },
        {
          "attributes": {},
          "data": "body",
          "siblings": [
            {
              "attributes": {},
              "data": "h1",
              "siblings": [
                {
                  "attributes": {},
                  "data": "POST: http://127.0.0.1:3000/?format=html",
                  "siblings": [],
                  "type": "text"
                }
              ],
              "type": "element"
            },
            {
              "attributes": {},
              "data": "div",
              "siblings": [
                {
                  "attributes": {},
                  "data": "ul",
                  "siblings": [
                    {
                      "attributes": {},
                      "data": "li",
                      "siblings": [
                        {
                          "attributes": {},
                          "data": "method: POST",
                          "siblings": [],
                          "type": "text"
                        }
                      ],
                      "type": "element"
                    },
                    {
                      "attributes": {},
                      "data": "li",
                      "siblings": [
                        {
                          "attributes": {},
                          "data": "url: http://127.0.0.1:3000/?format=html",
                          "siblings": [],
                          "type": "text"
                        }
                      ],
                      "type": "element"
                    },
                    {
                      "attributes": {},
                      "data": "li",
                      "siblings": [
                        {
                          "attributes": {},
                          "data": "query_params:",
                          "siblings": [],
                          "type": "text"
                        },
                        {
                          "attributes": {},
                          "data": "ul",
                          "siblings": [
                            {
                              "attributes": {},
                              "data": "li",
                              "siblings": [
                                {
                                  "attributes": {},
                                  "data": "format: html",
                                  "siblings": [],
                                  "type": "text"
                                }
                              ],
                              "type": "element"
                            }
                          ],
                          "type": "element"
                        }
                      ],
                      "type": "element"
                    },
                    {
                      "attributes": {},
                      "data": "li",
                      "siblings": [
                        {
                          "attributes": {},
                          "data": "headers:",
                          "siblings": [],
                          "type": "text"
                        },
                        {
                          "attributes": {},
                          "data": "ul",
                          "siblings": [
                            {
                              "attributes": {},
                              "data": "li",
                              "siblings": [
                                {
                                  "attributes": {},
                                  "data": "host: 127.0.0.1:3000",
                                  "siblings": [],
                                  "type": "text"
                                }
                              ],
                              "type": "element"
                            },
                            {
                              "attributes": {},
                              "data": "li",
                              "siblings": [
                                {
                                  "attributes": {},
                                  "data": "user-agent: go-resty/2.7.0 (https://github.com/go-resty/resty)",
                                  "siblings": [],
                                  "type": "text"
                                }
                              ],
                              "type": "element"
                            },
                            {
                              "attributes": {},
                              "data": "li",
                              "siblings": [
                                {
                                  "attributes": {},
                                  "data": "content-length: 17",
                                  "siblings": [],
                                  "type": "text"
                                }
                              ],
                              "type": "element"
                            },
                            {
                              "attributes": {},
                              "data": "li",
                              "siblings": [
                                {
                                  "attributes": {},
                                  "data": "content-type: application/json",
                                  "siblings": [],
                                  "type": "text"
                                }
                              ],
                              "type": "element"
                            },
                            {
                              "attributes": {},
                              "data": "li",
                              "siblings": [
                                {
                                  "attributes": {},
                                  "data": "accept-encoding: gzip",
                                  "siblings": [],
                                  "type": "text"
                                }
                              ],
                              "type": "element"
                            }
                          ],
                          "type": "element"
                        }
                      ],
                      "type": "element"
                    },
                    {
                      "attributes": {},
                      "data": "li",
                      "siblings": [
                        {
                          "attributes": {},
                          "data": "code: null",
                          "siblings": [],
                          "type": "text"
                        }
                      ],
                      "type": "element"
                    },
                    {
                      "attributes": {},
                      "data": "li",
                      "siblings": [
                        {
                          "attributes": {},
                          "data": "version: 1.1",
                          "siblings": [],
                          "type": "text"
                        }
                      ],
                      "type": "element"
                    },
                    {
                      "attributes": {},
                      "data": "li",
                      "siblings": [
                        {
                          "attributes": {},
                          "data": "body:",
                          "siblings": [],
                          "type": "text"
                        },
                        {
                          "attributes": {},
                          "data": "div",
                          "siblings": [
                            {
                              "attributes": {},
                              "data": "{\"hello\":\"world\"}",
                              "siblings": [],
                              "type": "text"
                            }
                          ],
                          "type": "element"
                        }
                      ],
                      "type": "element"
                    }
                  ],
                  "type": "element"
                }
              ],
              "type": "element"
            }
          ],
          "type": "element"
        }
      ],
      "type": "element"
    }
  ],
  "type": "document"
}`),
			method: POST,
			err:    nil,
		},
	}{
		var ok bool
		var j interface{}
		err, result := test.method.Call(test.args...)
		// Check if the actual result is Equal to the expected result only if there is no error.
		if err == nil {
			if err = json.Unmarshal(test.result, &j); err != nil {
				t.Error(err)
			}
			err, ok = EqualInterface(result.Value.(map[string]interface{})["content"], j)
		}

		if testing.Verbose() && result != nil {
			fmt.Printf("%d: $%s(%v) = %s\n", testNo + 1, test.method.String(), test.args, result.String())
		}

		if test.err != nil {
			if err.Error() != test.err.Error() {
				t.Errorf("error \"%s\" for testNo: %d does not match the required error: \"%s\"", err.Error(), testNo+1, test.err.Error())
			}
		} else if err != nil {
			t.Errorf("error \"%s\" should not have occurred (testNo: %d)", err.Error(), testNo + 1)
		} else if !ok {
			fmt.Println()
			fmt.Println()
			fmt.Println((&data.Value{
				Value:    result.Value.(map[string]interface{})["content"],
				Type:     data.Object,
			}).String())
			fmt.Println("========================= VS ==========================")
			fmt.Println(string(test.result))
			fmt.Println()
			t.Errorf("result \"%v\" for testNo: %d does not match the required result: \"%v\"", result, testNo + 1, string(test.result))
		}
	}

	// Kill the echo chamber
	pgid, err := syscall.Getpgid(echoChamber.Process.Pid)
	if err == nil {
		if err = syscall.Kill(-pgid, 15); err != nil {
			t.Error("failed to kill echo chamber")
		}
	}
	_ = echoChamber.Wait()
}
