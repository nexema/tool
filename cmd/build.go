package cmd

import (
	"github.com/sirupsen/logrus"
	"tomasweigenast.com/nexema/tool/builder"
)

func buildCmd(path, snapshotOut string) error {
	builder := builder.NewBuilder(path)
	err := builder.Discover()
	if err != nil {
		return err
	}

	logrus.Infof("Building project...")
	err = builder.Build()
	if err != nil {
		return err
	}

	if !builder.HasOutput() {
		logrus.Infoln("Nothing to build")
		return nil
	}

	if len(snapshotOut) > 0 {
		logrus.Infoln("saving snapshot...")
		filepath, err := builder.SaveSnapshot(snapshotOut)
		if err != nil {
			return err
		}

		logrus.Infof("snapshot file saved at %s\n", filepath)

		return nil
	}

	return nil
}
