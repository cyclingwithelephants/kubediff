package utils

import (
	"log"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

// Intentionally backwards due to
// https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func DefaultEnv(envVarName, defaultVal string) string {
	result := os.Getenv(envVarName)
	if result != "" {
		return result
	}
	return defaultVal
}

func MustGetEnv(envVarName string) string {
	result := os.Getenv(envVarName)
	if result != "" {
		return result
	}
	log.Fatalf("%s not set", envVarName)
	return ""
}

func AsInt(val string) int {
	result, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("%s must be an integer: %s", val, err)
	}
	return result
}

func AsBool(val string) bool {
	result, err := strconv.ParseBool(val)
	if err != nil {
		log.Fatalf("%s must be a boolean: %s", val, err)
	}
	return result
}

// Set represents a mathematical set
type Set map[string]struct{}

func NewSet() Set {
	return make(map[string]struct{})
}

func (s1 Set) Union(s2 Set) Set {
	for key := range s2 {
		s1[key] = struct{}{}
	}
	return s1
}

func (s1 *Set) Add(value string) {
	(*s1)[value] = struct{}{}
}
