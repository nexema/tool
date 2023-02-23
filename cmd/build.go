package cmd

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/builder"
)

func buildCmd(path, snapshotOut string) error {
	builder := builder.NewBuilder(path)
	err := builder.Discover()
	if err != nil {
		return err
	}

	err = builder.Build()
	if err != nil {
		return err
	}

	if len(snapshotOut) > 0 {
		fmt.Println("saving snapshot...")
		filepath, err := builder.SaveSnapshot(snapshotOut)
		if err != nil {
			return err
		}

		fmt.Printf("snapshot file saved at %s\n", filepath)

		return nil
	}

	return nil
}
