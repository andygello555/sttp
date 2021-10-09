# Individual Project 2021 - Jakab Zeller

## Abstract

The idea for creating a simple scripting language for the development and testing of Web APIs came from my use, reliance and creation of such APIs in my work life as well as in my spare time. I’ve often found the available tools for such development (such as Postman or Insomnia) quite limited when it comes to control-flow before or after requesting a resource from an API. Thus, the idea of a scripting language for this very purpose came about.
The language will include variable declaration/definition, control-flow (if and for statements), function definitions, short builtin functions for every HTTP method, and JSON manipulation via json-path. The language will also be dynamically typed with values being stored as JSON parsable strings. My hopes are that this will make working with JSON Web APIs (the standard when it comes to popular Web APIs) easier and more intuitive. This means that the only supported “types” will be strings, integers, floats, JSON objects and lists.

## Test Programs

All programs are written using Go so the Go executable will need to be installed for these programs to be run/compiled. There are instructions on how to run each program in their respective sections. These instructions require you to be in the test program's respective directory.<br/>

Note that the execution instructions use the `run` subcommand of the Go executable. This will compile and run the `main` package of the current directory.

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

#### Executing a file/input

`go run . [<FILE> | <INPUT>]`

#### Interactive mode

`go run .`

### Thompsons

*Located in: `test_programs/thompsons/`*<br/>

A parser for simple regular expressions which outputs minimised DFAs. The parser first produces a NFA using Thompson's construction, then converts this to a DFA using the Subset construction, and finally minimises this DFA using Dead State minimisation.<br/>

Behind the scenes this uses [participle](https://github.com/alecthomas/participle) to parse the regular expression and the [go-graphviz](https://github.com/goccy/go-graphviz) library to render the graphs to PNG files.

#### How to use

`go run . <REGEX>`

This will parse the given input regex and produce the following files (in the current directory):

- `thompsons.png`: the graph after parsing the regular expression to an NFA using Thompson's construction
- `subset.png`: the graph after carrying out the Subset construction on the NFA
- `deadstate.png`: the graph after minimising the DFA using Dead State minimisation
