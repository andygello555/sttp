# Handout for presentation

## Slide 4 - Intro to `sttp`

Example code shown on slide with comments. Uses Javascript syntax highlighting for readability.

```js
// JSON object literal
set object = {
    "repeats": 30,
    "null": null,
    "false": false,
    "true": true,
    "pi": 3.142,
    "object": {
        "hello": "world!"
        "array": [
            1,
            1 + 1,
            1 + 1 + 1
        ]
    },
    // Function calls for each HTTP method
    "methodcall": GET("https://example.com"),
    "results": []
};

// Defining object methods
// This is done via JSON path
function object.execute(name)
    // Object properties are accessed with "self"
    // Following is a traditional C-style loop. Also
    // support for iterator style loops
    for set i = 0; i < self.repeats; set i = i + 1 do
        if i % 3 == 0 then
            set self.results = self.results + GET("https://google.com");
        elif i % 5 == 0 then
            set self.results = self.results + GET("https://twitter.com");
        elif i % 7 == 0 then
            set self.results = self.results + GET("https://reddit.com");
        else
            set self.results = self.results + GET("https://facebook.com");
        end;
    end;
    return "hello" + name;
end;

try this
    // Batch statement executes all HTTP as 
    // separate goroutines
    batch this
        for url in [
            "https://a.com"
            "https://b.com"
            "https://c.com"
        ] do
            set result = GET(url);
            print(result);
        end;
        set oops_forgot_this = GET("https://d.com");
        print(oops_forgot_this);
    end;
    throw ("error": "time", "ohno": true};
catch as err then
    // Catch statements force the user to define
    // a variable for the exception to be stored in
    print(err);
end;

object.execute("sttp");
```

## Slide 6 - Four function calculator

Below is a code snippet from the four function calculator which explains how the participle parser generator works.

```go
// Statement can either be an expression, variable assignment or variable clear.
// The EBNF for this non-terminal:
//  ( Clear | Assignment | Expression ) ( EOL | ";" | EOF )
type Statement struct {
    // Pointer to Clear statement node. Can be nil.
    Clear      *Clear      `(   @@`  // "@@" denotes the capture of the type 
                                     // of the field. In this case the Clear
                                     // non-terminal.

    // Pointer to Assignment statement node. Can be nil.
    Assignment *Assignment `  | @@`

    // Pointer to Expression statement node. Can be nil.
    // EOL and EOF are passed into the parser from lexer.
    // Each statement can be finished by either an EOL, semicolon or EOF.
    Expression *Expression `  | @@ ) (EOL | ";" | EOF)`
}

// Each node implements the same Eval function signature.
// This means each node implements an interface which defines
// such a signature.
// The Memory type is a map containing the current values of each defined
// variable and is passed to each Eval function.
func (s *Statement) Eval(ctx Memory) (float64, *Memory) {
    // Nil switch as we have three alternates.
    switch {
    case s.Clear != nil:
        s.Clear.Eval(ctx)
        return 0, &ctx
    case s.Assignment != nil:
        s.Assignment.Eval(ctx)
        return 0, &ctx
    }
    return s.Expression.Eval(ctx), &ctx
}
```

## Slide 7 - Regex parser

**Some example inputs.**

### `aa*`

*Example that is on slides.*

![NFA, DFA, and minimised DFA for `aa*`](assets/regex-parser-aa*.png)

### `(a|b)a(b|e)*`

*Note that "e" stands for epsilon.*

#### NFA

![NFA for `(a|b)a(b|e)*`](assets/regex-parser-example-2-thompsons.png)

#### DFA

![DFA for `(a|b)a(b|e)*`](assets/regex-parser-example-2-subset.png)

#### Minimised DFA

![Minimised DFA for `(a|b)a(b|e)*`](assets/regex-parser-example-2-deadstate.png)

### `(a|b)*a`

#### NFA

![NFA for `(a|b)*a`](assets/regex-parser-example-3-thompsons.png)

#### DFA

![DFA for `(a|b)*a`](assets/regex-parser-example-3-subset.png)

#### Minimised DFA

![Minimised DFA for `(a|b)*a`](assets/regex-parser-example-3-deadstate.png)

## Slide 8 - `sttp` grammar

The entire grammar of `sttp`, in EBNF, is shown below:

![sttp grammar in EBNF](assets/grammar-ebnf.png)

## Slide 9 - Echo-chamber Web API

Examples of requests and responses.

```json
GET 127.0.0.1:3000?hello=world

{
    "code": null,
    "headers": {
        "accept": "*/*",
        "accept-encoding": "gzip, deflate",
        "connection": "keep-alive",
        "host": "127.0.0.1:3000",
        "user-agent": "HTTPie/2.6.0",
    },
    "method": "GET",
    "query_params": {
        "hello": "world"
    },
    "url": "http://127.0.0.1:3000/?hello=world",
    "version": "1.1"
}
```

```json
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