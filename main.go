package main

import "gopkg.in/yaml.v3"

func main() {
	b, _ := yaml.Marshal(map[string]interface{}{
		"dart": map[string]interface{}{
			"out": "./dist/dart",
			"options": []interface{}{
				"writeReflection",
				"omitReflection",
			},
		},
		"csharp": map[string]interface{}{
			"out": "./dist/csharp",
			"options": []interface{}{
				"omitReflection",
			},
		},
	})

	println(string(b))
}
