package testhelpers

import (
	"os"
	"path/filepath"

	"github.com/ztrue/tracerr"
)

const lookup = "go.mod"

func GetCurrentGoModulePath() (string, error) {
	var err error
	found := false
	wd, _ := os.Getwd()

	for wd != "/" {
		found, err = hasTarget(wd, lookup)
		if err != nil {
			return "", err
		}
		if found {
			return wd, nil
		}

		wd = filepath.Dir(wd)
	}

	if !found {
		return "", tracerr.Errorf("can't find the root directory containing %q", lookup)
	}

	return wd, nil
}

func hasTarget(source, target string) (bool, error) {
	files, err := os.ReadDir(source)
	if err != nil {
		tracerr.Wrap(err)
	}
	for _, file := range files {
		if file.Name() == target {
			return true, nil
		}
	}
	return false, nil
}
