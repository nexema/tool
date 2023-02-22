package cmd

import "tomasweigenast.com/nexema/tool/builder"

func build(path string) (*builder.Builder, error) {
	builder := builder.NewBuilder(path)
	err := builder.Build()
	if err != nil {
		return nil, err
	}

	return builder, nil
}
