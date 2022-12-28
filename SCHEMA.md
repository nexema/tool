# MessagePack Schema Specification

## Table of contents
1. [Primitive Data Types](#primitive-data-types)
2. [Types](#types)
    1. [Indexes](#indexes) 
    2. [Default values](#default-values)
    3. [Metadata](#metadata)
    4. [Nullable](#nullable-fields)
    5. [Enum](#enum-type)
3. [Writing schema files](#writing-schema-files)
	1. [Importing schema files](#importing-schema-files)
4. [Naming conventions](#naming-conventions)

## Primitive Data Types <a name="primitive-data-types"></a>

The MessagePack Schema suports the same MessagePack's data types. They are listed down:
| Primitive | C# Type | Dart Type |
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

> It is known that Map and List are not primitives in many programming languages, but for MessagePack Schema they are ;)


## Types
When using **MessagePack Schema**, [extensions](https://github.com/msgpack/msgpack/blob/master/spec.md#extension-types) are removed in favor of `types`. 

**Type syntax**:
```
@(metadata)
type [type name] [struct|union|enum] {
	(index) [field name](?):[primitive] = (default value) @(metadata)
	[...other fields]
}
```
You can define  3 different `types`, which are `struct`, `union` and `enum`.

- `struct` contains a set of fields with values. You can think them as classes in C# or objects in JavaScript.
- `union` is declared like a `struct` but the only difference is that only one field can be set at a time, therefore, they use the same memory when created.
- `enum` is a set of key-int value pairs

### **Indexes**
Indexes are `int32` numbers that are optional in `struct` and `union` but required for `enum`. They just indicates the order of serialization/deserialization. If no specified, they are implicit defined starting from 0.

### **Default values**
If you want a field to have a default value, you can specify it after the field index, using the *equals* sign (**=**).
E.g.:
```
my_amazing_field: string = "here! a default value!"
another_field: float32 = -242.32
my_amazing_enum: MyEnum = MyEnum.second
my_amazing_list: list(string) = ["one", "two", "three"]
my_amazing_map: map(string,int) = [("one":1), ("two":2), ("three:3")]
```
* `string` fields must declare their default value between quotation marks, as a normal string in many programming languages.
* `list(T)` fields declare their default values between brackets, each value comma separated.
* `map(TKey, TValue)` fields default values are declared between brackets, like lists, each entry, denoted using the following syntax: `key:value`, between parenthesis and comma separated.


> Keep in mind that **binary**, **struct** and **union** fields cannot declare default values. If **list(T)**'s or **map(TKey, TValue)**'s contains as generic argument one of the just mentioned data types, they cannot declare default values as well.


> **NOTE:** If you don't define a default value for fields that are not nullable, the default value of each programming language will be used. For example, for strings, it will be an **empty string**, for booleans, **false**, for ints, **0** and so on.

### **Metadata**  
Metadata are `map(string, [string|bool|float64])` which can be used to annotate fields/types for later use. You can specify metadata to any field, using **@**, followed by a  `map(string, [string|bool|int])` value signature.
For example:
```
a_field: string @[("obsolete":true)]
```


###  **Nullable fields**  
They are simply fields that accept `null` values.
You can declare a field as nullable simply adding a question mark (**?**)

**Example**
```
type Address {
    street_name: string
    building_number?: uint32
    name: string = "Default"
    coordinates?: Coordinates
    tags?: list(string) 
}

type Coordinates {
    latitude: float32
    longitude: float32
}
```

---

### Enum type
Enum type defines an object whose fields are constants values. They cannot have default values, metadata or be nullable.

**Basic `enum` syntax**
```
type [enum name] enum {
	[index] [name]
	[...other fields]
}
```

Note that it has the same declaration as `struct`, but it has the `enum` keyword after the name.

**Enum example**
```
type MyEnum enum {
	0 unknown
	1 value
	2 another
	3 third
}
```
> The default value of the enum is the field with the 0-index. In the above example, `unknown` is the default enum value.


## Writing schema files
Schema files can be organized in folders, and, when compiled, the output will replicate the folder structure.
To define a schema file, create a file with any name but with the extension `.mpack`

The root directory must contain a single file called `mpack.yaml` which will define some details for the project, it has the following structure:
```yaml
name: my_amazing_project
author: ImTheAuthor
version: 1.0.0

dependencies:
    - git:github.com/dep/myDependency
    - local:../dep/myDependency

skip:
	- skip/this/folder
	- skip/this/file.mpack
	- skip/all/**

generators:
	dart:
        out: ./dist/dart
        options:
            - writeReflection
	csharp:
        out: ./dist/csharp
        options:
            - omitReflection

```

### Importing schema files
You can import schema files  using the `import` keyword.
For example, you created a file called `identity.mpack` and then another called `common.mpack` which declares a struct called `Address`.
```
import "common" // Here we are importing the common.mpack" file

type User {
	address: common.Address
}
```

If you want to create a subpackage, you don't need to do extra work. Just include the base package in the import declaration, like:
```
import "foo/bar"
    
type Baz {
    myField: bar.Vim
}
```
If you have type collision, you can alias your imports:
```
import "foo/bar" AS b1
import "another_bar" AS b2
    
type Baz {
    field: b1.Vim
	other: b2.Vim
}
```

## Naming conventions
In order to messagepack-schema generate correct names for different programming languages and match their own naming conventions:

- **Field names:** snake_case
- **Indexes:** be 0-index
