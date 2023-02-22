package cmd

import "fmt"

func buildCmd(path, snapshotOut string) error {
	builder, err := build(path)
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
