
# Nexema - binary interchange made simple

![Nexema logo](https://raw.githubusercontent.com/nexema/resources/main/nexemalogo-160.png)


## What is this?
Nexema born as a side project when I was working in a very large project and  writing .proto files for the HTTP services got me a lot of heaches. I'm sure a lot of optimizations can be done, either on the **tool** or in any plugin. 

Nexema is an alternative for Protobuf or MessagePack that aims to be developer-friendly. You write your types and services one time, you run a command and then you have source code files for every language you want.

## Another one?
The first thing that maybe came to your mind was: *really? Another data interchange format?*  
Nexema is not another data interchange format, it aims to solve common problems that appears when working with Protobuf or MessagePack for example.   

* **MessagePack misses an schema**, if you want to use it for sharing data in a HTTP server, probably you will end up writing classes for the backend and then for the client. 
* **Protobuf misses some OOP features**, if you want inheritance, you will probability write another message and then append it to "extended" version. If you want nullability, you have `optional`, but this does not generate true nullable types is some languages. 


Nexema gives you more flexibility when generating code. You have enums, structs, unions, inheritance, documentation. You can select what kind of serialization process does the generated code use ([read more here](#serialization)).

## Features
* **Inheritance**, write base types and then extend them
* **Unions**, like classes but one field can be set at time
* **Enums**
* **You decide what to generate**, if you want only types, you got it. If you want types and serialization, you got it. 
* **Faster JSON serialization**, made by generating json serialization/deserialization code. No more reflection.
* **Auto generated HTTP/GRPC services**
* **Realtime serialization/deserialization**, if you are working with large types, instead of serializing/deserializing when creating the object, you can keep the buffer in the type and read from it when necessary.
* **Dependencies**, reuse your types in multiple projects.

## Getting Started
To get started, first [download](https://github.com/nexema/tool/releases) the **Nexema Tool**, which is a command-line app used to parse and compile **.nex** files to different programming. Add it to your PATH and then:

1- Open up a terminal and write `nexema init`. This will guide you to create a new Nexema project.  
2- Open **nexema.yaml** and add the generators you need:  
```yaml
name: my_amazing_project
author: ImTheAuthor
version: 1.0.0

generators:
  go:
    out: ./out/go
  javascript:
    out: ./out/js
  csharp:
    out: ./out/csharp
  dart:
    out: ./out/dart
```
> List of supported languages [here](#supported-languages)

3- Create a new file called `user.nex` with the following content:
```go
type User struct {
	first_name: string
	last_name: string
	email: string
	active: bool
	tags: list(string)
	preferences: map(string, string?)
	account_type: AccountType
}

type AccountType enum {
	unknown
	admin
	customer
	seller
}
```
4- Then run `nexema generate ./ ./output` and enjoy your source code 😋

For a more in-deep guide, read [schema](#schema.md) .

## Serialization
If you want to know more about how Nexema serializes/deserializes binary data, you can read more about it [here](#definition.md).

**Serialization modes**
Nexema can generate source code files that uses one of the two supported serialization modes.  For a better explanation, **JavaScript** output will be used as an example.

* **Realtime mode** won't fill up fields when calling the constructor of the `User` class, instead, it will generate an extra fields in the class called `buffer` and be of type **Buffer** on Node or **Uint8Array** on JavaScript for the browser. If you write the following code:
```javascript
const user = User.use(buffer) // Buffer not deserialized
console.log(user.firstName) // Searches in the buffer for field
```
Only when you make the first use of a field, in this case, in a getter, it will check if the field was cached (was called before) and return that value, otherwise, will only deserialize that field. When calling the setter, it will write to the underlying buffer, modifying it as needed.
* **Ahead-of-time mode** will completely deserialize the received buffer and fill all the fields. When calling the setter of a field, it will only change the value in memory, but won't create an underlying buffer. Only when you call `.encode()` it will actually serialize the class and give you a **Buffer** or an **Uint8Array**.

You can guess the benefits of each other. **Realtime mode** will give you faster performance when reading (maybe when writing it will suffer the performance of copies, because if you modify a string field that was "hello" to "hello world", the underlying buffer has to be reallocated with more memory.). The only performance hit of **Ahead-of-time mode** is deserializing the buffer (which is actually quite fast on most languages) if it's considerably large.


## Supported Languages
* [Golang](#go)
* [C#](#csharp)
* [Dart](#dart)
* [JavaScript/TypeScript](#javascript)

## Contributing
If you want to contribute to Nexema either writing/fixing/modifying a plugin, the **tool**, suggest new features or report issues I would be very grateful. Any pull request is well received too. 

## License
Nexema is under the GNU General Public License v3.0.

## Special Thanks
Dedicated to the person I've most loved, D. Who motivated me a lot during my young programming life.