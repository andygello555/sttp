# Talking notes

## Slide 2 - Summary

- Motivations: why I think there's a gap in the current HTTP/REST client and why I think a specialized scripting language is the perfect fit
- Intro to `sttp`: how I think a scripting language like this would look
- Test programs: the test programs I wrote to build up my confidence with my stack and also the ones I wrote out of interest
- Reports: that I wrote to enhance my knowledge
- And how development of `sttp` is going

## Slide 3 - Motivations

- There are three big HTTP/REST clients
  - Postman: membership fee
  - Insomnia: free and has plugin support
  - Paw: only for macos
- All of them have similar features and design philosophies
- None of them support very complex control flow
  - Make a sequence of requests that can be used in another
  - E.g. Login flow
- No loops or conditionals
- Limited variable or constant definitions
- They can take quite some time to learn
- They frequently leave me wanting more functionality and I usually switch to Python REPL and the requests library. Which made me wonder, why not create a scripting language purposely built for this use case?

## Slide 4 - Intro to `sttp`

- `sttp` is the scripting language I think can get over many of the pitfalls of conventional HTTP/REST clients
- `sttp` uses the types that are available in JSON so responses from web APIs can be directly used/modified.
- There is...
  - JSON path accessing and setting of JSON values
  - HTTP method calls builtin
  - HTTP request batching via the `batch` statement
  - Of course you have typical programming language constructs
    - Looping and conditionals
  - Way to construct test suites from directory structure
    - Each directory can represent a different resource. Such as a user.
    - Then within each directory you would have a script for all the respective actions for that resource with tests inside each
  - Non-whitespace dependant. Useful for when you just want to make a quick response from the command-line
- Interpreted, and implemented within Go using the participle parser generator

## Slide 5

Going to talk about the test programs that I wrote.

## Slide 6 - Four function calculator

- Done mainly to get back up to speed with Go and learn the participle parser generator
- The participle parser generator works by defining you AST nodes as `struct` types then using Go's field annotation support to define the grammar using EBNF
- **See handout**
- It's a simple four function calculator which supports
  - Addition, subtraction, multiplication, and division
  - Variable assignment and deletion
  - Has an interactive REPL mode

## Slide 7 - Regex parser

- Constructs a minimal DFA from a given regular expression by carrying out the Thompson's construction, then carrying out the subset construction then finally carrying out Dead State minimisation
- Although not in my original plan, this program was added after I gained some interest from CS3470
- Graphviz is used to render visualisations of each phase to a `.png` files
- **See handout for some more examples**

## Slide 8 - sttp grammar

- Defining the formal grammar of the sttp programming language
- **See handout for EBNF grammar**
- Then I defined the AST nodes for the grammar and just rewrote the EBNF within the relevant field annotations
- Precedence is encoded within the grammar. This decision was made after writing a report about associativity and precedence in grammars that I will discuss later

## Slide 9 - Echo-chamber Web API

- Simple nodejs web server which echoes back information about the request that was made
- I thought this would be useful whilst testing `sttp`

## Slide 10

Now I move onto reports that I have written.

## Slide 11 - Context-free grammars and manual procedures for parsing languages

- Begun fairly early on in the year
- Most of the knowledge used was learnt before it was formally taught in CS3470 by reading ahead and using other sources
- I cover...
  - Some background information on formal languages such as the Chomsky hierarchy and why context-free languages are used within programming
  - Manual parsing techniques
  - There are a lot of example grammars and derivations which explain some properties of both grammars and languages
  - I also added a note on how the participle parser generator fits into the world of grammars and languages

## Slide 12 - The use derivation rules, and grammar idioms to capture notions of associativity and precedence in arithmetic expressions

- Covers associativity and precedence in formal mathematics and how computer language designers encode this into the grammar or their interpreters or compilers
- Also covers how to get around pitfalls which occur in recursive descent parser generators
  - Such as participle
  - Left recursion and how to make operators left associative
- The research for this report influenced me to encode the precedence of the `sttp` arithmetic operations within the grammar itself

## Slide 13

Now I'll talk a bit about the development that has been done so far on `sttp`.

## Slide 14 - Development timeline

- Mid-October was when I started work on `sttp`
  - Defined the grammar
  - Defined AST nodes
  - Also defined a way of dumping the AST back into `sttp` source code as a way of debugging the grammar
- End of October
  - Focussed on formally defining how the `batch` statement would work
  - I also added some missing features such as try-catch and throw
- November
  - Implemented most of the important data structures such as
    - Variable memory map for variable lookup
    - VM structure which encompasses the state of the interpreter
    - Call stack structure and stack frame structures
  - Also implemented the operator action lookup table as well as a few operator actions
  - As well as the cast function lookup table which defines the available cast actions from one type to another
    - Used in operator actions
