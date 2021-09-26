# Individual Project 2021 - Jakab Zeller

## Abstract

The idea for creating a simple scripting language for the development and testing of Web APIs came from my use, reliance and creation of such APIs in my work life as well as in my spare time. I’ve often found the available tools for such development (such as Postman or Insomnia) quite limited when it comes to control-flow before or after requesting a resource from an API. Thus, the idea of a scripting language for this very purpose came about.
The language will include variable declaration/definition, control-flow (if and for statements), function definitions, short builtin functions for every HTTP method, and JSON manipulation via json-path. The language will also be dynamically typed with values being stored as JSON parsable strings. My hopes are that this will make working with JSON Web APIs (the standard when it comes to popular Web APIs) easier and more intuitive. This means that the only supported “types” will be strings, integers, floats, JSON objects and lists.
