# MessagePack Schema Specification

## Table of contents
1. [Primitive Data Types](#primitive-data-types)
2. [Types](#types)
	1. [Struct](#struct-type)
	    1. [Field declaration](#field-declaration) 
	    2. [Default values](#default-values)
	    3. [Metadata](#metadata)
	    4. [Nullable](#nullable-fields)
	    5. [Example](#example)
	2. [Enum](#enum-type)
	3. [Union](#union-type)
3. [Writing schema files](#writing-schema-files)

## Primitive Data Types <a name="primitive-data-types"></a>

The MessagePack Schema suports the same MessagePack's data types. They are listed down:
| Data Type | C# Type | Dart Type |
|--- |--- |--- |
| ```boolean``` | ```bool``` | ```bool``` |
| ```string``` | ```string``` | ```String``` |
| ```uint8``` | ```byte``` | ```int``` |
| ```uint16``` | ```ushort``` | ```int``` |
| ```uint32``` | ```uint``` | ```int``` |
| ```uint64``` | ```uint``` | ```int``` |
| ```int8``` | ```sbyte``` | ```int``` |
| ```int16``` | ```short``` | ```int``` |
| ```int32``` | ```int``` | ```int``` |
| ```int64``` | ```long``` | ```int``` |
| ```float32``` | ```float``` | ```double``` |
| ```float64``` | ```double``` | ```double``` |
| ```binary``` | ```byte[]``` | ```Uint8List``` |
| ```list(T)``` | ```List<T>``` | ```List<T> ``` |
| ```map(TKey, TValue)``` | ```Dictionary<TKey, TValue>``` | ```Map<TKey, TValue>``` |

> It is known that Map and List are not data types in many programming languages, but for MessagePack Schema they are ;)


## Types
When using **MessagePack Schema**, [extensions](https://github.com/msgpack/msgpack/blob/master/spec.md#extension-types) are removed in favor of `types`. You can define  3 different `types`.

### Struct type
Struct type defines a simple object with properties. You can think structs as a `Class` in C# or as an `Object` in JavaScript.

**Basic `struct` syntax**:
```
type [Type Name] {
	[field name]:[data type] [field index] = (default value) @(metadata)
	[...other fields]
}
```
Let's take a look at how `structs` are built.
First, we use the keyword `type` to tell the schema we are declarating a `struct`, following its name. [Naming conventions are listed down.](#)
Then, all the fields need to be wrapped between curly brackets, like any other programming language.

**Field declaration:**
Fields must start with their name, following a colon, and the data type they will store. For example:
```
my_amazing_field:string
```
After that, following an whitespace, the index of the field must be indicated. **First field must have the 0-index and then be consecutive.**
```
my_amazing_field:string 0
another_amazing_field:binary 1
```
Here, we are declarating two fields, the first one, with the 0-index, its a `string` field named `my_amazing_field`. The second one, named `another_amazing_field` stores `binary` data, with the index 1.

**Default values**
If you want a field to have a default value, you can specify it after the field index, using the *equals* sign (**=**).
E.g.:
```
my_amazing_field:string 0 = "here! a default value!"
another_field:float32 1 = -242.32
my_amazing_enum:MyEnum 2 = MyEnum.second
```
* `string` fields must declare their default value between quotation marks, as a normal string in many programming languages.
* `list(T)` fields declare their default values between brackets, each value comma separated.
* `map(TKey, TValue)` fields' default values are declared between brackets, like lists, each entry, denoted using the following syntax: `key:value`, between parenthesis and comma separated.


> Keep in mind that **binary**, **struct** and **union** fields cannot declare default values. If **list(T)**'s or **map(TKey, TValue)**'s contains as generic argument one of the just mentioned data types, they cannot declare default values as well.

`list(T)` and `map(TKey, TValue)` default values example:
```
my_amazing_field:list(string) 0 = ["one", "two", "three"]
another_field:map(string,int) 1 = [("one":1), ("two":2), ("three:3")]
```

> **NOTE:** If you don't define a default value for fields that are not nullable, the default value of each programming language will be used. For example, for strings, it will be an **empty string**, for booleans, **false**, for ints, **0** and so on.

**Metadata**  
Metadata are `map(string, [string|bool|int])` which can be used to annotate fields for later use. You can specify metadata to any field, using **@**, followed by a  `map(string, [string|bool|int])` value signature.
For example:
```
a_field:string 0 @[("obsolete":true)]
```

They can be used with default values too:
```
a_field:string 0 = "hello world!" @[("obsolete":true)]
```

**Nullable fields**  
You are allowed to define certain fields are nullable, except `list(T)` and `map(TKey, TValue)`.
You can declare a field as nullable simply adding a question mark (**?**)

**Example**
```
type Address {
    street_name: string 0
    street_number: uint32? 1
    name: string 2 = "Default"
    coordinates: Coordinates 3 
}

type Coordinates {
    latitude: float32 0
    longitude: float32 1
}
```

### Enum type
Enum type defines an object whose fields are constants values. They cannot have default values, metadata or be nullable.

**Basic `enum` syntax**
```
type [Enum Name] enum {
	[value name] [field index]
	[...other fields]
}
```

Note that it has the same declaration as `struct`, but it has the `enum` keyword after the name.

**Enum example**
```
type MyEnum enum {
	unknown 0
	value 1
	another 2
	third 3
}
```
> The default value of the enum if the field with the 0-index. In the above example, `unknown` is the default enum value.

### Union type
Union type is the same as `struct`, with the difference only one field can be set at time and they use the same memory allocation while serializing.

**Basic `union` syntax**
```
type [Union Name] union {
	[field name]:[data type] [field index] = (default value) @(metadata)
	[...other fields]
}
```
>It uses the same signature as `struct`, the only difference is you need to add the `union` keyword after its name.

## Writing schema files
Schema files are organized in folders. A folder is represented as a `package` with n-`.mpack` files.
All `.mpack` files inside a folder must contain the `package` keyword equals to the folder's name. 

The root directory must contain a single file called `mpack.yaml` which will define some details for the project, it has the following structure:
```yaml
name: example.com/myProject
author: ImTheAuthor
version: 1.0.0

dependencies:
    - git:github.com/dep/myDependency
    - local:../dep/myDependency

```

### Importing schema files
You can import schema files from another package using the `import` keyword.
For example, you create a package called `identity` and then another called `common` which declares a struct called `Address`.
```
version:1
package: identity
import "common" // Here we are importing the common package

type User {
	address:common.Address 0
}
```

If you want to create a subpackage, you don't need to do extra work. Just include the base package in the import declaration, like:
```
import
    - "foo/bar"
    
type Baz {
    myField:foo.bar.Vim 0
}
```
