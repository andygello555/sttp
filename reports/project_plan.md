# An Interpreted Scripting Language for the Development and Testing of Web APIs

## Project plan

## Abstract

The idea for creating a simple scripting language for the development and testing of Web APIs came from my use, reliance and creation of such APIs in my work life as well as in my spare time. I’ve often found the available tools for such development (such as [Postman](https://www.postman.com/) or [Insomnia](https://insomnia.rest/)) quite limited when it comes to control-flow before or after requesting a resource from an API. Thus, the idea of a scripting language for this very purpose came about.<br/>
The language will include variable declaration/definition, control-flow (if and for statements), function definitions, short builtin functions for every HTTP method, and JSON manipulation via [json-path](https://www.baeldung.com/guide-to-jayway-jsonpath). The language will also be dynamically typed with values being stored as JSON parsable strings. My hopes are that this will make working with JSON Web APIs (the standard when it comes to popular Web APIs) easier and more intuitive. This means that the only supported "types" will be strings, integers, floats, JSON objects and lists.<br/>
The interpreter itself will be implemented in Go using the [participle](https://github.com/alecthomas/participle) parser by Alec Thomas (which enables you to define your grammar within your token structs using tags) and then navigating and executing the produced AST.

## Deliverables plan

### Reports

#### A report on the use of context free grammars to specify syntax, along with manual procedures for generating and parsing languages

This is to gain a better understanding when it comes to implementing a grammar using a context free grammar, as well as giving a deeper understanding as to what makes parser generators tick.<br/>
_Timeline: **15th October - 20th of October**_

#### A report on the use derivations, and grammar idioms to capture notions of associativity and priority in arithmetic expressions

The use of asymmetries is key in modern parser generators to force derivations to be done in a particular way. This is especially important when defining the grammar and semantics for arithmetic expressions. Basic arithmetic expressions will need to be supported by the parser for the language I am creating. Therefore, this is invaluable information.<br/>
_Timeline: **15th November - 20th November**_

### Test programs

#### The development of an interpreter for a language that models a four function calculator with memory

A good warm up to see if my software stack is up to the task of implementing the entire scripting language. It will also give me an incentive to get up to speed with the participle parser generator as well as getting back up to speed with Go.<br/>
_Timeline: **5th - 15th October**_

#### Creating the parser grammar using participle as well as the parser for simple JSON path expressions

This will form one of the constituent parts of the final product so it will be a useful program to implement as it will reduce the time spent implementing the final product.<br/>
_Timeline: **20th October - 30th of October**_

#### A simple Web API for testing

This will aid in testing the HTTP request and JSON functionality of the scripting language. This will either be written in Go or Node.<br/>
_Timeline: **1st November - 10th November**_

### Term 2

#### The interpreter

I plan to start writing the code for the interpreter by the end of the first term (10th December), and be finished by Mid-March. This will give me some overlap when I'm writing my report and programming. I'm hoping this will provide me with up to date content to add to my report.

#### Final report

I plan to start writing the final report at the beginning of the second term (10th January) up until the deadline date. This will provide me with enough time to submit a complete draft of my report by the 18th February. It will also allow me to get started with programming so that issues I face during programming will be fresh in my mind whilst writing the report.<br/>
It will also make the professional issues section easier as by that point I will have been fully engrossed in any professional issues that may occur in this field.

## Research

I have timetabled in reading materials so that I can gain a better understanding of the subject matter. The following books are the ones that I will read as thoroughly as possible within the given time frame.

### Aho, A., Lam, M., Sethi, R. and Ullman, J., 2014. *Compilers; Principles, Techniques, and Tools*, By Alfred V.

This will aid in writing both of my reports as it offers the theory and techniques behind compiler writing.<br/>
_Timeline: **5th - 20th October**_

### Wirth, N., 1976. *Algorithms + Data Structures = Programs*. p.Chapter 5, Appendix B.

Wirth's book on algorithms and data structures also has a chapter on compilers and an appendix which includes Pascal syntax diagrams. These should help me understand the history of compilers as well as giving me examples of syntax diagrams.<br/>
_Timeline: **5th - 10th October**_

## Risks

### The base set of language features might be too easy or too difficult to implement

*Medium likelihood*<br/>
*Medium importance*<br/>

In which case more features could be added or taken away depending on the time available. Some extra features that could be added:

- Support for more advanced HTTP requests, such as proxied requests or asynchronous requests
- More fleshed out builtin library
- Multi-threading support (making use of Go’s brilliant parallelism features)

### There might not be much to write about in reports in regards to the theory

*Low likelihood*<br/>
*Low importance*<br/>

In which case I would implement a bespoke LR parser for the language. Replacing the parser generated by participle. This would only be done if I had enough time. It would give me an opportunity to talk about how LR parsers are implemented but it could also be a hindrance to the grammar of the language as I could only add syntax which I know that the parser will be able to handle.

### I run into roadblocks, due to my software stack, whilst implementing the interpreter for the four function calculator with memory

*Medium-High likelihood*<br/>
*High importance*<br/>

If this happens I will either switch to a different more capable software stack or ask assistance from peers or maintainers of the libraries that I will be using. Some other options for parser generators could be:

- Bison
- Yacc
- ANTLR

These options would mean switching the target language to either C, Java or Python.

<div style="page-break-after: always"></div>

## Material Covered for Project Plan

1. Johnstone, A., n.d. *Software Language Engineering*. 1st ed. (As well as supplementing slides)<br/>
_**For understanding of key subject matters and terminology**_
1. Chong, S., 2018. *CS153: Compilers Lecture 6: LR Parsing*. [online] Groups.seas.harvard.edu. Available at: <https://groups.seas.harvard.edu/courses/cs153/2018fa/lectures/Lec06-LR-Parsing.pdf> [Accessed 19 September 2021].<br/>
_**Research into LR parsing**_
1. Fiore, M., 2010. *Lecture Notes on Regular Languages and Finite Automata for Part IA of the Computer Science Tripos*. [online] Cl.cam.ac.uk. Available at: <https://www.cl.cam.ac.uk/teaching/1011/RLFA/LectureNotes.pdf> [Accessed 19 September 2021].<br/>
_**Recap of finite automata theory taught in second year**_
1. Kling, F., 2017. *AST explorer*. [online] Astexplorer.net. Available at: <https://astexplorer.net/> [Accessed 19 September 2021].<br/>
_**A Javascript AST generator. Gave me some insight as to what tokens I should be parsing and how an AST tree should look like**_
1. Kun, J., 2019. *A Working Mathematician’s Guide to Parsing*. [online] Math ∩ Programming. Available at: <https://jeremykun.com/2019/04/20/a-working-mathematicians-guide-to-parsing/> [Accessed 19 September 2021].<br/>
_**Gives a high-level rundown of how parsing works and capabilities of parsers**_
1. En.wikipedia.org. 2021. *LR parser - Wikipedia*. [online] Available at: <https://en.wikipedia.org/wiki/LR_parser> [Accessed 19 September 2021].<br/>
_**Research into LR parsers to aid section on report on derivations**_
1. Cs.ecu.edu. 2021. *CSCI 5220: Precedence and associativity*. [online] Available at: <http://www.cs.ecu.edu/karl/5220/spr16/Notes/CFG/precedence.html> [Accessed 26 September 2021].<br/>
_**Research into precedence and associativity. Again to aid section of report on derivations**_
1. Aho, A., Lam, M., Sethi, R. and Ullman, J., 2014. *Compilers; Principles, Techniques and Tools*, By Alfred V..<br/>
_**The first of the reading materials mentioned in the [Research](#research) section**_
1. Wirth, N., 1976. *Algorithms + Data Structures = Programs*. p.Chapter 5, Appendix B.<br/>
_**The second of the reading materials mentioned in the [Research](#research) section**_
