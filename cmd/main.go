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
	HasDiff(dir1, dir2 string) (bool, error)
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
	differ := file.NewRealDiffer(logger)
	config := newConfig()
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
		globLevels:            utils.MustGetEnvAsInt("GLOB_LEVELS"),
		renderedYamlWriteRoot: utils.DefaultEnv("RENDERED_WRITE_PATH", "rendered"),
		tempPath:              utils.DefaultEnv("TEMP_PATH", "tmp"),
		renderedCommentPath:   utils.DefaultEnv("TEMP_PATH", "tmp"),
		githubOwner:           utils.MustGetEnv("GITHUB_OWNER"),
		githubRepo:            utils.MustGetEnv("GITHUB_REPO"),
		githubPrNumber:        utils.MustGetEnvAsInt("GITHUB_PR_NUMBER"),
		githubToken:           utils.MustGetEnv("GITHUB_TOKEN"),
	}
}

func (S Tool) RunToCompletion() error {
	// finds all apps, regardless of which environment or branch they are in.
	allApps, err := S.appFinder.GetAllAppPaths()
	if err != nil {
		return err
	}

	diffPaths := make(map[string]struct{})
	for eachApp := range allApps {
		dir1 := filepath.Join(S.config.prDir, S.config.envsDir, eachApp)
		dir2 := filepath.Join(S.config.targetDir, S.config.envsDir, eachApp)
		hasDiff, err := S.differ.HasDiff(dir1, dir2)
		if err != nil {
			return err
		}
		if hasDiff {
			diffPaths[eachApp] = struct{}{}
			S.logger.Println("diff found between branches for app: ", eachApp)
		}
	}

	// render the yaml for each diffPath
	builtYamls := []yaml.BuiltYaml{}
	for diffPath := range diffPaths {
		builtYaml, err := S.yamlBuilder.Build(diffPath)
		if err != nil {
			return err
		}
		builtYamls = append(builtYamls, builtYaml)
	}

	fileDiffs := []file.Diff{}
	for _, builtYaml := range builtYamls {
		diff, err := S.differ.Diff(builtYaml.YamlPrBranch, builtYaml.YamlTargetBranch)
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
		// If we have more than 2 chunks for a single app, that's a bit silly and we shouldn't be making that many comments
		// I'm still playing with this number, but I think 2 is a good number for now.
		// TODO: list the limitation that this isn't really very good for adding new apps
		// TODO: perhaps we can list all of the resources that are going to be created instead of the actual diffs
		//if len(chunks) > 2 {
		//	S.logger.Println("too many chunks to render. Built ", len(chunks), " chunks for app: ", fileDiff.AppPath)
		//	os.Exit(0)
		//}
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

	// clean up old comments
	S.logger.Println("deleting all old comments")
	err = S.githubCommenter.DeleteAllToolComments()
	if err != nil {
		return err
	}
	// create a PR comment for each rendered template
	err = S.githubCommenter.Comment(renderedTemplates)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	Tool := New()
	err := Tool.RunToCompletion()
	if err != nil {
		log.Fatal(err)
	}
}
