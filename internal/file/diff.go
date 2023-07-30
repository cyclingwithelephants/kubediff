package file

import (
	"fmt"
	"github.com/gosimple/hashdir"
	"log"
	"os"

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

func (D *RealDiffer) HasDiff(dir1, dir2 string) (bool, error) {
	// Check if both directories exist
	// If either doesn't exist, it counts as a diff (e.g. adding or removing an app)
	_, err1 := os.Stat(dir1)
	_, err2 := os.Stat(dir2)
	if os.IsNotExist(err1) {
		D.logger.Println("Found diff because the directory doesn't exist:", dir1)
		return true, nil
	} else if os.IsNotExist(err2) {
		D.logger.Println("Found diff because the directory doesn't exist:", dir2)
		return true, nil
	} else if err1 != nil {
		return false, fmt.Errorf("error accessing directory %s: %w", dir1, err1)
	} else if err2 != nil {
		return false, fmt.Errorf("error accessing directory %s: %w", dir1, err2)
	}

	dir1Hash, err := hashdir.Make(dir1, "md5")
	if err != nil {
		return false, fmt.Errorf("error hashing directory %s: %w", dir1, err)
	}

	dir2Hash, err := hashdir.Make(dir2, "md5")
	if err != nil {
		return false, fmt.Errorf("error hashing directory %s: %w", dir2, err)
	}

	if dir1Hash != dir2Hash {
		D.logger.Println("Found diff because the directories have different hashes")
		return true, nil
	}

	return false, nil
}
