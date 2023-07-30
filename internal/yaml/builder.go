package yaml

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

const generatedYamlFileName = "generated.yaml"

type BuiltYaml struct {
	AppPath          string
	YamlPrBranch     string
	YamlTargetBranch string
}

type Builder struct {
	prDir                 string
	targetDir             string
	envsDir               string
	renderedYamlWriteRoot string
	logger                *log.Logger
}

func NewBuilder(
	prDir string,
	targetDir string,
	envsDir string,
	renderedYamlWriteRoot string,
	logger *log.Logger,
) *Builder {
	return &Builder{
		prDir:                 prDir,
		targetDir:             targetDir,
		envsDir:               envsDir,
		renderedYamlWriteRoot: renderedYamlWriteRoot,
		logger:                logger,
	}
}

func (B *Builder) kustomizeBuild(directory string) (string, error) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return "", fmt.Errorf("directory %s does not exist", directory)
	}
	if _, err := os.Stat(path.Join(directory, "kustomization.yaml")); errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("directory %s does not contain kustomization.yaml", directory)
	}
	cmd := exec.Command("kustomize", "build", "--enable-helm", directory)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func (B *Builder) Build(appPath string) (BuiltYaml, error) {
	renderedYamls, err := B.buildForEach(appPath)
	if err != nil {
		return BuiltYaml{}, err
	}
	if len(renderedYamls) != 2 {
		return BuiltYaml{}, fmt.Errorf("expected 2 rendered yamls, got %d", len(renderedYamls))
	}
	return BuiltYaml{
		AppPath:          appPath,
		YamlPrBranch:     renderedYamls[0],
		YamlTargetBranch: renderedYamls[1],
	}, nil
}

func (B *Builder) buildForEach(appPath string) ([]string, error) {
	branchPaths := []string{
		B.prDir,
		B.targetDir,
	}
	renderedYamls := []string{}
	for _, branchPath := range branchPaths {
		renderedYaml, err := B.build(branchPath, appPath)
		if err != nil {
			return renderedYamls, err
		}
		renderedYamls = append(renderedYamls, renderedYaml)
	}
	return renderedYamls, nil
}

func (B *Builder) build(branchPath, appPath string) (string, error) {
	fullAppPath := filepath.Join(
		branchPath,
		B.envsDir,
		appPath,
	)

	// If there is no directory, the yaml is emptys
	if _, err := os.Stat(fullAppPath); os.IsNotExist(err) {
		return "", nil
	}

	// If there is a kustomization.yaml file one level up, skip this directory
	renderedYaml, err := B.kustomizeBuild(fullAppPath)
	if err != nil {
		return "", err
	}

	return renderedYaml, nil
}

func (B *Builder) ensureDir(dirPath string) error {
	B.logger.Println("creating directory: ", dirPath)
	err := os.MkdirAll(dirPath, 0o755)
	if err != nil {
		return err
	}
	return nil
}
