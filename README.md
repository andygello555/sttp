<!-- omit in toc -->
# `sttp`

- [Abstract](#abstract)
- [Source code](#source-code)
  - [Libraries/Code used](#librariescode-used)
  - [How to use](#how-to-use)
    - [Prerequisites](#prerequisites)
    - [Running from within the repository](#running-from-within-the-repository)
    - [Building an executable and running it](#building-an-executable-and-running-it)
    - [Examples](#examples)
  - [Running tests](#running-tests)
    - [Debugging info](#debugging-info)
    - [Benchmarks](#benchmarks)
- [Echo-chamber Web API](#echo-chamber-web-api)
    - [Prerequisites](#prerequisites-1)
    - [How to use](#how-to-use-1)

## Abstract

The idea for creating a simple scripting language for the development and testing of Web APIs came from my use, reliance and creation of such APIs in my work life as well as in my spare time. I’ve often found the available tools for such development (such as Postman or Insomnia) quite limited when it comes to control-flow before or after requesting a resource from an API. Thus, the idea of a scripting language for this very purpose came about.
The language will include variable declaration/definition, control-flow (if and for statements), function definitions, short builtin functions for every HTTP method, and JSON manipulation via json-path. The language will also be dynamically typed with values being stored as JSON parsable strings. My hopes are that this will make working with JSON Web APIs (the standard when it comes to popular Web APIs) easier and more intuitive. This means that the only supported “types” will be strings, integers, floats, JSON objects and lists.

## Source code

### Libraries/Code used

- [participle/v2](https://github.com/alecthomas/participle) (licensed under MIT): recursive descent parser generator for Go by Alec Thomas.
- [resty/v2](https://github.com/go-resty/resty) (licensed under MIT): a HTTP/REST client for Go.
- [atomicgo/cursor](https://github.com/atomicgo/cursor) (licensed under MIT): cross-platform methods to move the terminal cursor in different directions. For `sttp` REPL mode.
- [pkg/term](https://github.com/pkg/term) (licensed under BSD 2): used for reading from terminal in raw mode. Again for `sttp` REPL.
- [`getChar` function from climenu](https://github.com/paulrademacher/climenu/blob/master/getchar.go) (licensed under [The Unlicense](https://unlicense.org/)): a `getChar` function (similar to C implementation), written for Go by Paul Rademacher for their library: [climenu](https://github.com/paulrademacher/climenu). This function is defined in `src/repl.go` for use in `sttp`'s REPL mode.
- [gotils v1.2.7](https://pkg.go.dev/github.com/andygello555/gotils) (licensed under GPL-3.0): a set of utility functions written for Go by me.
- [go v1.17](https://go.dev/) (licensed under a [BSD-style](https://go.dev/LICENSE) license): programming language.

### How to use

You can either run the `sttp` interpreter from within the repository, or by creating an executable that can be run later. You can provide the interpreter with the following inputs:

- Interactively using the new REPL mode. This is entered by not supplying any arguments.
- An `.sttp` file containing `sttp` source code.
- The root of a directory containing `.sttp` files to run as a TestSuite.
- Raw `sttp` code to execute from the terminal. E.g. `./sttp '$print("Hello World!");'`

#### Prerequisites

You must have `go 1.17` (or later) installed.

#### Running from within the repository

`go run . [ FILE | DIRECTORY | INPUT ]`

#### Building an executable and running it

```console
go build -o sttp
./sttp [ FILE | DIRECTORY | INPUT ]
```

#### Examples

Examples for `sttp` can be found within the `_examples/` directory. Each example is stored within its own directory holding the following files:

- `*.sttp` file (required): containing the source code for the example.
- `*.stdout` file (optional): containing the expected stdout output for the example.
- `*.stderr` file (optional): containing the expected stderr output for the example.
- `*.tests` file (optional): containing the expected test output.
- `*.err` file (optional): containing the expected errors that bubble up to the bottommost stack frame during execution.

Some of these examples utilise the echo-chamber web API. The web server for which is located within: `_examples/echo_chamber/` directory. The usage of which can be found [here](#echo-chamber-web-api).<br/>

### Running tests

The Go tests for `sttp` can be run using the following command when inside the `src/` directory:

```console
go test ./...
```

You might need to increase the limit for the number of open sockets on your system (via `ulimit` or Windows alternative). This is because Go is known to run some test cases in parallel in order to speed up test execution. In conjunction to this, there are also a lot of tests that manage interactions between the [echo chamber web API](#echo-chamber-web-api), and batched HTTP requests made via `sttp`.<br/>

#### Debugging info

Debugging info can be enabled by appending the `-v` flag to the command mentioned [above](#running-tests). **This will lead to a lot of output when running tests**, so it might be beneficial to only run specific `*_test.go` files/test cases by using the `-run PATTERN` option.

#### Benchmarks

Benchmarks used for testing performance of the `batch` statement can be run using the following command:

```console
go test -run=XXX -bench="Benchmark(No)?Batch" -benchtime=5x -cpu=8 -count=3
```

As with [tests](#running-tests), this can chew through **a lot** of sockets on your system, so it is advised you tinker around with the `-benchtime` and `-count` flags.

## Echo-chamber Web API

*Located in: `_examples/echo_chamber/`*<br/>

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
