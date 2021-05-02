package uploader

import (
	"context"

	"github.com/google/go-github/v35/github"
)

// githubClient provides a minimal GitHub client API suitable for testing.
type githubClient interface {
	CreateBlob(ctx context.Context, owner string, repo string, blob *github.Blob) (*github.Blob, *github.Response, error)
	CreateRef(ctx context.Context, owner string, repo string, ref *github.Reference) (*github.Reference, *github.Response, error)
	CreateTree(ctx context.Context, owner string, repo string, baseTree string, entries []*github.TreeEntry) (*github.Tree, *github.Response, error)
	CreateCommit(ctx context.Context, owner string, repo string, commit *github.Commit) (*github.Commit, *github.Response, error)
	GetCommitSHA1(ctx context.Context, owner, repo, ref, lastSHA string) (string, *github.Response, error)
	UpdateRef(ctx context.Context, owner string, repo string, ref *github.Reference, force bool) (*github.Reference, *github.Response, error)
}

type githubUploaderClient struct {
	c *github.Client
}

func (g githubUploaderClient) GetCommitSHA1(ctx context.Context, owner, repo, ref, lastSHA string) (string, *github.Response, error) {
	return g.c.Repositories.GetCommitSHA1(ctx, owner, repo, ref, lastSHA)
}

func (g githubUploaderClient) CreateRef(ctx context.Context, owner string, repo string, ref *github.Reference) (*github.Reference, *github.Response, error) {
	return g.c.Git.CreateRef(ctx, owner, repo, ref)
}

func (g githubUploaderClient) CreateCommit(ctx context.Context, owner string, repo string, commit *github.Commit) (*github.Commit, *github.Response, error) {
	return g.c.Git.CreateCommit(ctx, owner, repo, commit)
}

func (g githubUploaderClient) CreateTree(ctx context.Context, owner string, repo string, baseTree string, entries []*github.TreeEntry) (*github.Tree, *github.Response, error) {
	return g.c.Git.CreateTree(ctx, owner, repo, baseTree, entries)

}
func (g githubUploaderClient) UpdateRef(ctx context.Context, owner string, repo string, ref *github.Reference, force bool) (*github.Reference, *github.Response, error) {
	return g.c.Git.UpdateRef(ctx, owner, repo, ref, force)
}

func (g githubUploaderClient) CreateBlob(ctx context.Context, owner string, repo string, blob *github.Blob) (*github.Blob, *github.Response, error) {
	return g.c.Git.CreateBlob(ctx, owner, repo, blob)
}
