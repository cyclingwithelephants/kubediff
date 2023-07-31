package file

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/gosimple/hashdir"

	"github.com/martinohmann/go-difflib/difflib"
)

type Diff struct {
	AppPath string
	Diff    string
}

type RealDiffer struct {
	logger *log.Logger
}

func NewRealDiffer(logger *log.Logger) *RealDiffer {
	return &RealDiffer{
		logger: logger,
	}
}

func (D *RealDiffer) Diff(a, b string) (string, error) {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(a),
		B:        difflib.SplitLines(b),
		FromFile: "",
		ToFile:   "",
		Context:  3,
		Color:    true,
	}
	return difflib.GetUnifiedDiffString(diff)
}

func (D *RealDiffer) HasDiff(dir1, dir2 string) (bool, string, error) {
	// Check if both directories exist
	// If either doesn't exist, it counts as a diff (e.g. adding or removing an app)
	_, err1 := os.Stat(dir1)
	_, err2 := os.Stat(dir2)
	if os.IsNotExist(err1) {
		return true, fmt.Sprintf("the directory doesn't exist: %s", dir1), nil
	} else if os.IsNotExist(err2) {
		return true, fmt.Sprintf("the directory doesn't exist: %s", dir2), nil
	} else if err1 != nil {
		return false, "", fmt.Errorf("error accessing directory %s: %w", dir1, err1)
	} else if err2 != nil {
		return false, "", fmt.Errorf("error accessing directory %s: %w", dir1, err2)
	}

	prevDir, err := os.Getwd()
	if err != nil {
		return false, "", fmt.Errorf("error getting current directory: %w", err)
	}

	hashes := []string{}
	for _, dir := range []string{dir1, dir2} {
		err = os.Chdir(path.Join(dir, ".."))
		if err != nil {
			return false, "", fmt.Errorf("error changing directory to %s: %w", dir1, err)
		}
		hash, err := hashdir.Make(filepath.Base(dir1), "md5")
		if err != nil {
			return false, "", fmt.Errorf("error hashing directory %s: %w", dir1, err)
		}
		hashes = append(hashes, hash)
		err = os.Chdir(prevDir)
		if err != nil {
			return false, "", fmt.Errorf("error changing directory to %s: %w", prevDir, err)
		}
	}

	if hashes[0] != hashes[1] {
		return true, "the directories have different hashes", nil
	}

	return false, "", nil
}
