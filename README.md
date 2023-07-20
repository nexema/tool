# Nexema - binary interchange made simple

![Nexema logo](https://raw.githubusercontent.com/nexema/resources/main/nexemalogo-160.png)

## What is this?

Nexema born as a side project when I was working in a very large project and writing .proto files for the HTTP services got me a lot of heaches. I'm sure a lot of optimizations can be done, either on the **tool** or in any plugin.

Nexema is an alternative for Protobuf or MessagePack that aims to be developer-friendly. You write your types and services one time, you run a command and then you have source code files for every language you want.

## Another one?

The first thing that maybe came to your mind was: _really? Another data interchange format?_  
Nexema is not another data interchange format, it aims to solve common problems that appears when working with Protobuf or MessagePack for example.

- **MessagePack misses an schema**, if you want to use it for sharing data in a HTTP server, probably you will end up writing classes for the backend and then for the client.
- **Protobuf misses some OOP features**, if you want inheritance, you will probability write another message and then append it to "extended" version. If you want nullability, you have `optional`, but this does not generate true nullable types is some languages.

## Features

- **Inheritance**, write base types and then extend them
- **Unions**, like classes but one field can be set at time
- **Enums**
- **Faster JSON serialization**, made by generating json serialization/deserialization code. No more reflection.

## More incoming

- **You decide what to generate**, if you want only types, you got it. If you want types and serialization, you got it.
- **Auto generated HTTP/GRPC services**
- **Realtime serialization/deserialization**, if you are working with large types, instead of serializing/deserializing when creating the object, you can keep the buffer in the type and read from it when necessary.
- **Dependencies**, reuse your types in multiple projects.

## Documentation

Check out the complete documentation at [nexema.dev](https://nexema.dev)
