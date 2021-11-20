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