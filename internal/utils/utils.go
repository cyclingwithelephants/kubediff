package utils

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
)

// IsDirectory checks if the given filepath points to a directory.
func IsDirectory(filePath string) (bool, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		// Return false and the error if there's any issue accessing the file/directory
		return false, err
	}

	// Check if the file is a directory
	return fileInfo.IsDir(), nil
}

// ListDirectories returns a list of all directories in the given path.
func ListDirectories(path string) ([]string, error) {
	var directories []string

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		// Check if the entry is a directory (excluding the root path itself)
		if file.IsDir() && file.Name() != "." && file.Name() != ".." {
			directories = append(directories, filepath.Join(path, file.Name()))
		}
	}

	return directories, nil
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

func MustGetEnvAsInt(envVarName string) int {
	result := MustGetEnv(envVarName)
	resultAsInteger, err := strconv.Atoi(result)
	if err != nil {
		log.Fatalf("%s must be an integer: %s", envVarName, err)
	}
	return resultAsInteger
}

func WriteStringToFile(path string, data string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return errors.Wrap(err, "could not open file '"+path+"'")
	}
	defer file.Close()

	// Write the data to the file
	_, err = file.WriteString(data)
	return err
}

func WriteToFile(strings []string, writePath string) error {
	// open a file for writing
	file, err := os.Create(writePath)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer file.Close()

	// use buffered writer to write to file
	writer := bufio.NewWriter(file)

	for _, line := range strings {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			log.Fatalf("failed writing to file: %s", err)
		}
	}

	// use Flush to ensure all buffered operations have been applied to the underlying writer
	writer.Flush()
	return nil
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
