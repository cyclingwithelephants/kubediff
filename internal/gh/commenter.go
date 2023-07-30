package gh

import (
	"context"
	"fmt"
	"github.com/google/go-github/v41/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"log"
	"strings"
)

// Commenter is a struct for interacting with the GitHub API.
// client: The GitHub client used to interact with the API.
// ctx: The context used for API requests.
// owner: The owner of the repository where comments will be posted.
// repo: The repository where comments will be posted.
// prNumber: The prNumber of the pull request where comments will be posted.
type Commenter struct {
	client          *github.Client
	ctx             context.Context
	owner           string
	repo            string
	prNumber        int
	CommentIdPrefix string
	logger          *log.Logger
}

// NewCommenter is a constructor for the Commenter struct.
// It takes the owner of the repository, the repository name, and the prNumber of the pull request as parameters.
// It returns a new instance of Commenter.
func NewCommenter(owner string, repo string, number int, personalAccessToken string, logger *log.Logger) *Commenter {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: personalAccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return &Commenter{
		client:          client,
		ctx:             ctx,
		owner:           owner,
		repo:            repo,
		prNumber:        number,
		CommentIdPrefix: "bot-comment-kubediff",
		logger:          logger,
	}
}

// Comment posts comments on a specific pull request.
// It takes a slice of comments as input and posts each comment on the pull request.
// It returns an error if any occurs during the process.
func (c *Commenter) Comment(comments []string) error {
	for i, comment := range comments {
		commentId := fmt.Sprintf("%s-%d", c.CommentIdPrefix, i)
		//index := int64(i)
		// add tags to comment so we can easily delete it later
		title := ""
		if i == 0 && len(comments) > 1 {
			title = fmt.Sprintf("## Wow that's a lot of changes! They'll be split over %d comments", len(comments))
		}
		template := `
%s
%s
<!-- %s -->
`
		comment = fmt.Sprintf(template, title, comment, commentId)
		c.logger.Printf("Posting comment")
		commentObject, response, err := c.client.Issues.CreateComment(
			c.ctx,
			c.owner,
			c.repo,
			c.prNumber,
			&github.IssueComment{
				Body: &comment,
				//ID:     &index,
				//NodeID: &commentId,
			},
		)
		if err != nil {
			return err
		}
		if response.StatusCode != 201 {
			return errors.New("failed to post comment: " + fmt.Sprintf("%v", response))
		}
		//var id int64
		//id = *commentObject.ID
		c.logger.Printf("Posted comment, id:", *commentObject.ID, "and node_id:", *commentObject.NodeID)
	}
	return nil
}

// Delete all comments made by the previous run of this tool.
func (c *Commenter) DeleteAllToolComments() error {
	c.logger.Printf("Listing comments")
	comments, resp, err := c.client.Issues.ListComments(c.ctx, c.owner, c.repo, c.prNumber, &github.IssueListCommentsOptions{ListOptions: github.ListOptions{PerPage: 100}})
	c.logger.Printf("received response: %+v", resp.Response)
	if err != nil {
		return err
	}
	c.logger.Printf("found %d comments", len(comments))

	for i, comment := range comments {
		c.logger.Printf("found comment")
		if strings.Contains(*comment.Body, fmt.Sprintf("%s-%d", c.CommentIdPrefix, i)) {
			//if true {
			resp, err := c.client.Issues.DeleteComment(c.ctx, c.owner, c.repo, *comment.ID)
			c.logger.Printf("received response: %+v", resp.Response)
			if resp.StatusCode != 204 {
				c.logger.Println("Failed to delete comment: " + fmt.Sprintf("%v", resp))
				break
			}
			if err != nil {
				return err
			}
		} else {
			c.logger.Printf("Skipping comment: %s", *comment.Body)
		}
	}

	return nil

	return nil
}
