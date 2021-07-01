package fs

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func DumpStream(path string, input io.ReadCloser) error {
	defer input.Close()
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return errors.Wrapf(err, "ensure directory: %q", path)
	}
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrapf(err, "open %q", path)
	}
	defer out.Close()
	_, err = io.Copy(out, input)
	if err != nil {
		return errors.Wrapf(err, "write %q", path)
	}
	return nil
}
