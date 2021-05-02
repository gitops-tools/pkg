package uploader

import (
	"context"
	"strings"
)

const DEBUG = 1

// Files is a mapping of filename within a repository to a byte slice of the
// content to be uploaded.
type Files map[string][]byte

// RepoReference represents a repository, e.g. "gitops-tools" and repo "pkg"
// along with a source branch.
type RepoReference struct {
	Owner  string
	Repo   string
	Branch string
}

func (r RepoReference) String() string {
	return strings.Join([]string{r.Owner, r.Repo}, "/") + ":" + r.Branch
}

// GitWriter implementations write files to Git.
type GitWriter interface {
	WriteFiles(ctx context.Context, files Files, src RepoReference, newBranch string) (string, error)
}
