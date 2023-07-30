package file

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cyclingwithelephants/kubediff/internal/utils"
)

type AppFinder struct {
	foundPaths utils.Set // a set of paths to all applications found, with the envsDir as the root
	prDir      string    // the directory where the PR branch is checked out
	targetDir  string    // the directory where the target branch is checked out
	envsDir    string    // the directory containing all environment definitions
	globLevels int       // the number of levels to recurse with globbing into the envsDir
	logger     *log.Logger
}

func NewAppFinder(prDir, targetDir, envsDir string, globLevels int, logger *log.Logger) *AppFinder {
	return &AppFinder{
		foundPaths: utils.NewSet(),
		prDir:      prDir,
		targetDir:  targetDir,
		envsDir:    envsDir,
		globLevels: globLevels,
		logger:     logger,
	}
}

// findAppRemovingPrefixPath finds all paths in a directory, and removes the prefixPath from each path
// This is used because we pull the same repo in two different directories, and we want to compare
// the paths relative to the root of the repo, not the root of the directory where the repo is pulled
func (F *AppFinder) findAppRemovingPrefixPath(prefixPath string, dirPath string) (utils.Set, error) {
	paths := utils.NewSet()
	globPath := path.Join(prefixPath, dirPath)
	F.logger.Println("globbing:", globPath)
	apps, err := filepath.Glob(globPath)
	if err != nil {
		return nil, err
	}
	F.logger.Println("found", len(apps), "apps")

	for _, app := range apps {
		// validate that application is directory
		fileInfo, err := os.Stat(app)
		if err != nil {
			return nil, err
		}
		if !fileInfo.IsDir() {
			// TODO: Make this check if the containing directory is an app
			F.logger.Println("skipping non-directory: ", app)
			continue
		}

		appWithoutPrefix := strings.TrimPrefix(app, prefixPath+"/")
		F.logger.Println("found app in directory", prefixPath, ":", appWithoutPrefix)
		paths.Add(appWithoutPrefix)
	}
	return paths, nil
}

func (F *AppFinder) GetAllAppPaths() (utils.Set, error) {
	globs := ""
	for i := 0; i < F.globLevels; i++ {
		globs += "/*"
	}

	// Find app paths in both the PR and target branches
	prefix := path.Join(F.prDir, F.envsDir)
	prPaths, err := F.findAppRemovingPrefixPath(prefix, globs)
	if err != nil {
		return nil, err
	}
	prefix = path.Join(F.targetDir, F.envsDir)
	targetPaths, err := F.findAppRemovingPrefixPath(prefix, globs)
	if err != nil {
		return nil, err
	}

	return prPaths.Union(targetPaths), nil
}
