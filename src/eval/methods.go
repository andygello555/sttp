package eval

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"github.com/go-resty/resty/v2"
	"net/http"
)

// Method represents a valid HTTP method supported by sttp.
type Method int

const (
	GET Method = iota
	HEAD
	POST
	PUT
	DELETE
	OPTIONS
	PATCH
)

var methodMap = map[string]Method{
	"GET":     GET,
	"HEAD":    HEAD,
	"POST":    POST,
	"PUT":     PUT,
	"DELETE":  DELETE,
	"OPTIONS": OPTIONS,
	"PATCH":   PATCH,
}

var methodNameMap = map[Method]string{
	GET:     "GET",
	HEAD:    "HEAD",
	POST:    "POST",
	PUT:     "PUT",
	DELETE:  "DELETE",
	OPTIONS: "OPTIONS",
	PATCH:   "PATCH",
}

type MethodParamType int

const (
	Url MethodParamType = iota
	Body
	Headers
	Cookies
)

var methodParamTypeName = map[MethodParamType]string{
	Url: "url",
	Body: "body",
	Headers: "headers",
	Cookies: "cookies",
}

func (mpt MethodParamType) String() string {
	return methodParamTypeName[mpt]
}

// methodParams is a lookup of parameters which are required for all the supported Methods. true indicates the argument
// is required, false indicates that the argument is not required.
var methodParams = map[Method]map[MethodParamType]bool {
	GET:     {Url: true, Headers: false, Cookies: false},
	HEAD:    {Url: true, Headers: false, Cookies: false},
	POST:    {Url: true, Headers: false, Cookies: false, Body: false},
	PUT:     {Url: true, Headers: false, Cookies: false, Body: false},
	DELETE:  {Url: true, Headers: false, Cookies: false, Body: false},
	OPTIONS: {Url: true, Headers: false, Cookies: false},
	PATCH:   {Url: true, Headers: false, Cookies: false, Body: false},
}

// ApplyArg will call the relevant setter on the given resty.Request pointer. Will return an error if a Cast went awry.
func (mpt MethodParamType) ApplyArg(arg *data.Value, request *resty.Request) error {
	var stringMap map[string]string
	switch mpt {
	case Url:
		arg = &data.Value{
			Value:    arg.String(),
			Type:     data.String,
			Global:   arg.Global,
			ReadOnly: arg.ReadOnly,
		}
	case Body:
		request.SetBody(arg.Value)
	case Cookies:
		fallthrough
	case Headers:
		var err error
		if arg.Type != data.Object {
			if err, arg = Cast(arg, data.Object); err != nil {
				return err
			}
		}

		// Then we will construct a map[string]string for the headers
		stringMap = make(map[string]string)
		for k, v := range arg.Value.(map[string]interface{}) {
			var vString string
			switch v.(type) {
			case string:
				vString = v.(string)
			default:
				var t data.Type
				var newVal *data.Value
				if err = t.Get(v); err != nil {
					return err
				}
				if err, newVal = Cast(&data.Value{
					Value:    v,
					Type:     t,
					Global:   arg.Global,
					ReadOnly: arg.ReadOnly,
				}, data.String); err != nil {
					return err
				}
				vString = newVal.Value.(string)
			}
			stringMap[k] = vString
		}

		if mpt == Headers {
			request.SetHeaders(stringMap)
		} else {
			for n, v := range stringMap {
				request.SetCookie(&http.Cookie{
					Name:       n,
					Value:      v,
				})
			}
		}
	}
	return nil
}

// GetParamType will return the MethodParamType for the given i-th argument.
func (m *Method) GetParamType(arg int) MethodParamType {
	var mpt, i int
	for mpt, i = 0, 0; mpt < len(methodParams[*m]); mpt++ {
		if _, ok := methodParams[*m][MethodParamType(mpt)]; ok {
			if arg == i {
				break
			}
			// We only increment i if the current MethodParamType is used within the Method's Call function
			i++
		}
	}
	return MethodParamType(mpt)
}

// Call will call the HTTP method.
func (m *Method) Call(args ...*data.Value) (err error, value *data.Value) {
	if len(args) > 0 {
		request := resty.New().R()
		for i, arg := range args {
			mpt := m.GetParamType(i)
			if arg.Type != data.Null {
				if err = mpt.ApplyArg(arg, request); err != nil {
					return err, nil
				}
			} else {
				// Otherwise, if the value is null and the parameter is not optional we through an error
				if methodParams[*m][mpt] {
					return errors.MethodParamNotOptional.Errorf(mpt.String()), nil
				}
			}
		}

		var resp *resty.Response
		resp, err = request.Execute(m.String(), args[0].Value.(string))

		var err2 error
		var body *data.Value
		if err2, body = data.ConstructSymbol(string(resp.Body()), false); err2 != nil {
			return err2, nil
		}

		value = &data.Value{
			Value:    map[string]interface{}{
				"content": body.Value,
				"cookies": func() (cookies []interface{}) {
					cookies = make([]interface{}, len(resp.Cookies()))
					for i, cookie := range resp.Cookies() {
						cookies[i] = map[string]interface{}{
							"name": cookie.Name,
							"value": cookie.Value,
							"max_age": float64(cookie.MaxAge),
							"secure": cookie.Secure,
							"http_only": cookie.HttpOnly,
							"same_site": float64(cookie.SameSite),
							"raw": cookie.Raw,
						}
					}
					return cookies
				}(),
				"headers": func() (headers map[string]interface{}) {
					headers = make(map[string]interface{})
					for k, v := range resp.Header() {
						headers[k] = v
					}
					return headers
				}(),
				"received": resp.ReceivedAt().String(),
				"size": float64(resp.Size()),
				"status": resp.Status(),
				"code": float64(resp.StatusCode()),
				"time": resp.Time().String(),
			},
			Type:     data.Object,
			Global:   false,
			ReadOnly: false,
		}
	}
	return err, value
}

// Capture method for participle lexer.
func (m *Method) Capture(s []string) error {
	var ok bool
	*m, ok = methodMap[s[0]]
	if !ok {
		panic(fmt.Sprintf("Unsupported HTTP method: %s", s[0]))
	}
	return nil
}

// String returns the name of the method from the methodNameMap.
func (m *Method) String() string {
	return methodNameMap[*m]
}
