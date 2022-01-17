# Individual Project 2021 - Jakab Zeller

## Abstract

The idea for creating a simple scripting language for the development and testing of Web APIs came from my use, reliance and creation of such APIs in my work life as well as in my spare time. I’ve often found the available tools for such development (such as Postman or Insomnia) quite limited when it comes to control-flow before or after requesting a resource from an API. Thus, the idea of a scripting language for this very purpose came about.
The language will include variable declaration/definition, control-flow (if and for statements), function definitions, short builtin functions for every HTTP method, and JSON manipulation via json-path. The language will also be dynamically typed with values being stored as JSON parsable strings. My hopes are that this will make working with JSON Web APIs (the standard when it comes to popular Web APIs) easier and more intuitive. This means that the only supported “types” will be strings, integers, floats, JSON objects and lists.

## Test Programs

There are instructions on how to run each program in their respective sections. These instructions require you to be in the test program's respective directory.

### Four Function Calculator

*Located in: `test_programs/four_func_calc/`*<br/>

A simple four function calculator which uses the [participle](https://github.com/alecthomas/participle) parser to parse and evaluate simple mathematical expressions. The following is an example script.

```
let a=3*3
let b=a*2;(a+b)/2
clear a
let a=b*2
a+b
```

The output to this example would be a list of numbers representing the output of each statement. To run the interpreter you can use the following options:

#### Prerequisites

You must have `go 1.17` installed.

#### Executing a file/input

`go run . [<FILE> | <INPUT>]`

#### Interactive mode

`go run .`

### Thompsons

*Located in: `test_programs/thompsons/`*<br/>

A parser for simple regular expressions which outputs minimised DFAs. The parser first produces a NFA using Thompson's construction, then converts this to a DFA using the Subset construction, and finally minimises this DFA using Dead State minimisation.<br/>

Behind the scenes this uses [participle](https://github.com/alecthomas/participle) to parse the regular expression and the [go-graphviz](https://github.com/goccy/go-graphviz) library to render the graphs to PNG files.

#### Prerequisites

You must have `go 1.17` installed.

#### How to use

`go run . <REGEX>`

This will parse the given input regex and produce the following files (in the current directory):

- `thompsons.png`: the graph after parsing the regular expression to an NFA using Thompson's construction
- `subset.png`: the graph after carrying out the Subset construction on the NFA
- `deadstate.png`: the graph after minimising the DFA using Dead State minimisation

### Echo-chamber Web API

*Located in: `test_programs/test-web-api/`*<br/>

A simple node.js based web API server which echoes back information about any HTTP request made to it. This was created in order to have a web-API for testing `sttp` with. If the query param `format=html` is provided in the request then the response will be a mirror of the JSON response but will be returned as HTML. The server is forked 6 times creating 6 worker processes to create a rudimentary form of load balancing. This is in the hope that multiple requests can be handled at once.<br/>

The following are examples of some requests and responses:

```
GET 127.0.0.1:3000?hello=world

{
    "code": null,
    "headers": {
        "accept": "*/*",
        "accept-encoding": "gzip, deflate",
        "connection": "keep-alive",
        "host": "127.0.0.1:3000",
        "user-agent": "HTTPie/2.6.0"
    },
    "method": "GET",
    "query_params": {
        "hello": "world"
    },
    "url": "http://127.0.0.1:3000/?hello=world",
    "version": "1.1"
}
```

```
POST 127.0.0.1:3000/helloworld {"hello": "world"}

{
    "body": {
        "hello": "world"
    },
    "code": null,
    "headers": {
        "accept": "application/json, */*;q=0.5",
        "accept-encoding": "gzip, deflate",
        "connection": "keep-alive",
        "content-length": "18",
        "content-type": "application/json",
        "host": "127.0.0.1:3000",
        "user-agent": "HTTPie/2.6.0"
    },
    "method": "POST",
    "query_params": {},
    "url": "http://127.0.0.1:3000/helloworld",
    "version": "1.1"
}
```

```
GET 127.0.0.1:3000/api/hello?format=html

<html lang="en">
    <head>
        <title>GET: http://127.0.0.1:3000/api/hello?format=html</title>
    </head>
    <body>
        <h1>GET: http://127.0.0.1:3000/api/hello?format=html</h1>
        <div>
            <ul>
                <li>method: GET</li>
                <li>url: http://127.0.0.1:3000/api/hello?format=html</li>
                <li>
                    query_params:
                    <ul>
                        <li>format: html</li>
                    </ul>
                </li>
                <li>
                    headers:
                    <ul>
                        <li>host: 127.0.0.1:3000</li>
                        <li>user-agent: HTTPie/2.6.0</li>
                        <li>accept-encoding: gzip, deflate</li>
                        <li>accept: */*</li>
                        <li>connection: keep-alive</li>
                    </ul>
                </li>
                <li>code: null</li>
                <li>version: 1.1</li>
            </ul>
        </div>
    </body>
</html>
```

#### Prerequisites

Have node.js installed. No packages need to be installed.

#### How to use

`node main.js`

This will start the web server on `127.0.0.1:3000`.
