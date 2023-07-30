package gh

import (
	_ "embed"
)

// const MaxGithubCommentLength = 65536 // characters

// despite the prNumber officially being 65536, it seems this is
// the actual limit, at least for how my code is counting characters
const MaxGithubCommentLength = 253336 // characters

// embeds can't be constants
//
//go:embed git-diff-template.txt
var GitCommentTemplate string

// 50 is a buffer for the rest of the comment, like the header and footer
var MaxCommentLength = MaxGithubCommentLength - len(GitCommentTemplate) - 50

// var MaxCommentLength = 1000

//
