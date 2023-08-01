package main

import (
	_ "embed"
	"fmt"
	"log"
	"path/filepath"

	"github.com/cyclingwithelephants/kubediff/internal/file"
	"github.com/cyclingwithelephants/kubediff/internal/gh"
	"github.com/cyclingwithelephants/kubediff/internal/utils"
	"github.com/cyclingwithelephants/kubediff/internal/yaml"
)

type Config struct {
	prDir                 string
	targetDir             string
	envsDir               string
	globLevels            int
	renderedYamlWriteRoot string
	tempPath              string
	renderedCommentPath   string
	githubOwner           string
	githubRepo            string
	githubPrNumber        int
	githubToken           string
	diffWithColour        bool
	diffContextLines      int
}

type Tool struct {
	config          Config
	logger          *log.Logger
	differ          Differ
	renderer        TemplateRenderer
	appFinder       AppFinder
	yamlBuilder     YamlBuilder
	chunker         Chunker
	githubCommenter GithubCommenter
}

type Differ interface {
	Diff(pathA, pathB string) (string, error)
	HasDiff(dir1, dir2 string) (bool, string, error)
}

type TemplateRenderer interface {
	Render(templateString string, replacements map[string]string) (string, error)
}

type AppFinder interface {
	GetAllAppPaths() (utils.Set, error)
}

type YamlBuilder interface {
	Build(path string) (yaml.BuiltYaml, error)
}

type Chunker interface {
	Chunk(diff string) (chunks []string)
}

type GithubCommenter interface {
	DeleteAllToolComments() error
	Comment(comments []string) error
}

func New() Tool {
	logger := log.Default()
	config := newConfig()
	differ := file.NewRealDiffer(logger, config.diffContextLines, config.diffWithColour)
	logger.Println("creating server with config:", fmt.Sprintf("%+v", config))
	return Tool{
		config:   config,
		logger:   logger,
		differ:   differ,
		renderer: file.NewTemplateRenderer(),
		appFinder: file.NewAppFinder(
			config.prDir,
			config.targetDir,
			config.envsDir,
			config.globLevels,
			logger,
		),
		yamlBuilder: yaml.NewBuilder(
			config.prDir,
			config.targetDir,
			config.envsDir,
			config.renderedYamlWriteRoot,
			logger,
		),
		chunker: utils.NewChunker(gh.MaxCommentLength),
		githubCommenter: gh.NewCommenter(
			config.githubOwner,
			config.githubRepo,
			config.githubPrNumber,
			config.githubToken,
			logger,
		),
	}
}

func newConfig() Config {
	return Config{
		prDir:                 utils.DefaultEnv("PR_BRANCH_DIR", "pr"),
		targetDir:             utils.DefaultEnv("TARGET_BRANCH_DIR", "target"),
		envsDir:               utils.MustGetEnv("ENVS_DIR"),
		globLevels:            utils.AsInt(utils.MustGetEnv("GLOB_LEVELS")),
		renderedYamlWriteRoot: utils.DefaultEnv("RENDERED_WRITE_PATH", "rendered"),
		tempPath:              utils.DefaultEnv("TEMP_PATH", "tmp"),
		renderedCommentPath:   utils.DefaultEnv("TEMP_PATH", "tmp"),
		githubOwner:           utils.MustGetEnv("GITHUB_OWNER"),
		githubRepo:            utils.MustGetEnv("GITHUB_REPO"),
		githubPrNumber:        utils.AsInt(utils.MustGetEnv("GITHUB_PR_NUMBER")),
		githubToken:           utils.MustGetEnv("GITHUB_TOKEN"),
		diffWithColour:        utils.AsBool(utils.DefaultEnv("DIFF_WITH_COLOUR", "true")),
		diffContextLines:      utils.AsInt(utils.DefaultEnv("DIFF_CONTEXT_LINES", "3")),
	}
}

func (S Tool) RunToCompletion() error {
	// clean up old comments
	// we do this first ti reduce likelihood of confusion with the new comments
	S.logger.Println("begin deleting all old comments")
	err := S.githubCommenter.DeleteAllToolComments()
	if err != nil {
		S.logger.Println("error deleting old comments:", err)
		return err
	}

	// finds all apps, regardless of which environment or branch they are in.
	allApps, err := S.appFinder.GetAllAppPaths()
	if err != nil {
		S.logger.Println("error finding all apps:", err)
		return err
	}

	diffPaths := make(map[string]struct{})
	for eachApp := range allApps {
		dir1 := filepath.Join(S.config.prDir, S.config.envsDir, eachApp)
		dir2 := filepath.Join(S.config.targetDir, S.config.envsDir, eachApp)
		hasDiff, reason, err := S.differ.HasDiff(dir1, dir2)
		if err != nil {
			S.logger.Println("error checking if diff exists:", err)
			return err
		}
		if hasDiff {
			diffPaths[eachApp] = struct{}{}
			S.logger.Println("diff found between branches for app: ", eachApp, "reason:", reason)
		}
	}
	// at each path, if the directory above has a kustomization.yaml, remove it from the list
	for diffPath := range diffPaths {
		fullPath := filepath.Join(S.config.prDir, S.config.envsDir, diffPath, "..", "kustomization.yaml")
		exists, err := utils.FileExists(fullPath)
		if err != nil {
			S.logger.Println("error checking if file exists:", err)
			return err
		}
		if exists {
			delete(diffPaths, diffPath)
		}
	}

	// render the yaml for each diffPath
	builtYamls := []yaml.BuiltYaml{}
	for diffPath := range diffPaths {
		S.logger.Println("building yaml for path:", diffPath)
		builtYaml, err := S.yamlBuilder.Build(diffPath)
		if err != nil {
			return err
		}
		builtYamls = append(builtYamls, builtYaml)
	}

	fileDiffs := []file.Diff{}
	for _, builtYaml := range builtYamls {
		diff, err := S.differ.Diff(builtYaml.YamlTargetBranch, builtYaml.YamlPrBranch)
		if err != nil {
			return err
		}
		fileDiffs = append(
			fileDiffs,
			file.Diff{
				AppPath: builtYaml.AppPath,
				Diff:    diff,
			},
		)
	}

	// chunk the diffs into comments
	chunkedDiffs := []file.Diff{}
	for _, fileDiff := range fileDiffs {
		chunks := S.chunker.Chunk(fileDiff.Diff)
		for _, chunk := range chunks {
			chunkedDiffs = append(chunkedDiffs, file.Diff{
				AppPath: fileDiff.AppPath,
				Diff:    chunk,
			})
		}
	}

	// render the comment templates
	renderedTemplates := []string{}
	for _, chunkedDiff := range chunkedDiffs {
		renderedTemplate, err := S.renderer.Render(
			gh.GitCommentTemplate,
			map[string]string{
				"TITLE": chunkedDiff.AppPath,
				"DIFF":  chunkedDiff.Diff,
			},
		)
		if err != nil {
			return err
		}
		renderedTemplates = append(renderedTemplates, renderedTemplate)
	}

	// create a PR comment for each rendered template
	err = S.githubCommenter.Comment(renderedTemplates)
	if err != nil {
		S.logger.Println("error commenting:", err)
		return err
	}

	return nil
}

func main() {
	tool := New()
	err := tool.RunToCompletion()
	if err != nil {
		tool.logger.Fatal(err)
	}
}
