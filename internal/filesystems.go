package internal

import (
	"github.com/spf13/afero"
)

var AppFs = afero.NewOsFs()

func WriteToFileSystem(fs afero.Fs, content, filename string) error {
	err := afero.WriteFile(fs, filename, []byte(content), 0o644)
	if err != nil {
		return err
	}

	return nil
}
