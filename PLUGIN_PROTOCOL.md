# Nexema Plugin Protocol

Nexema will compile the project and will generate a json snapshot, which will follow the definition specified in the `definition_schema.json` file. After that, it will iterate over the specified list of plugins to use, and individually it will call the underlying binary.

Nexema will pass to the plugin the following list of arguments:

```
--output-path=[...] // the path to the folder where Nexema will write the output files generated for the plugin
... // other plugin options
```

Also, Nexema will write to the binary stdio the json snapshot and finish it with the newline character \n.

At this time, the plugin is responsible of generating the source code and must write to its stdout the following JSON when the process is finished:

```json
"exitCode": 0,
"error": null,
"files": [{
    "id": "the file id", // as specified in the input
    "name": "the file name", // the full file path, relative to output-path
    "contents": "the file source code"
}]
```

If for some reason the generation process fails, specify an `exitCode` different from 0 and, optionally but recommended populate the `error` field.
