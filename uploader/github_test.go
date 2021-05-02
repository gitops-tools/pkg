package uploader

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/zapr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v35/github"
	"go.uber.org/zap"
)

const testSHA = "be330f0f432100b5c4a4fc67963ca0952cedd7f6"

func TestWriteFiles(t *testing.T) {
	files := Files{
		"testing/file.yaml": []byte("test file 1"),
	}
	c := NewGitHubUploader(nil, zapr.NewLogger(zap.NewNop()))
	fc := newFakeGitHubClient(testSHA)
	c.client = fc

	_, err := c.WriteFiles(context.TODO(), files, RepoReference{Owner: "gitops-tools", Repo: "pkg", Branch: "testing"})
	if err != nil {
		t.Fatal(err)
	}
	want := []createdRef{
		{
			Owner: "gitops-tools",
			Repo:  "pkg",
			Ref: &github.Reference{
				Ref:    github.String("refs/heads/testing"),
				Object: &github.GitObject{SHA: github.String(testSHA)},
			},
		},
	}
	if diff := cmp.Diff(want, fc.createdRefs); diff != "" {
		t.Fatalf("failed to created refs:\n%s", diff)
	}
}

func newFakeGitHubClient(startSHA string) *fakeGitHubClient {
	return &fakeGitHubClient{
		startSHA: startSHA,
	}
}

type fakeGitHubClient struct {
	startSHA    string
	createdRefs []createdRef
}

// these are only exported fields for cmp
type createdRef struct {
	Owner, Repo string
	Ref         *github.Reference
}

func (f *fakeGitHubClient) CreateBlob(ctx context.Context, owner string, repo string, blob *github.Blob) (*github.Blob, *github.Response, error) {
	return nil, nil, nil
}

func (f *fakeGitHubClient) CreateRef(ctx context.Context, owner string, repo string, ref *github.Reference) (*github.Reference, *github.Response, error) {
	f.createdRefs = append(f.createdRefs, createdRef{Owner: owner, Repo: repo, Ref: ref})
	return nil, nil, nil
}

func (f *fakeGitHubClient) CreateTree(ctx context.Context, owner string, repo string, baseTree string, entries []*github.TreeEntry) (*github.Tree, *github.Response, error) {
	return nil, nil, nil
}

func (f *fakeGitHubClient) CreateCommit(ctx context.Context, owner string, repo string, commit *github.Commit) (*github.Commit, *github.Response, error) {
	return nil, nil, nil
}

func (f *fakeGitHubClient) GetCommitSHA1(ctx context.Context, owner, repo, ref, lastSHA string) (string, *github.Response, error) {
	if f.startSHA != "" {
		return f.startSHA, nil, nil
	}
	return "", nil, errors.New("failed to get commit sha for repo")
}

func (f *fakeGitHubClient) UpdateRef(ctx context.Context, owner string, repo string, ref *github.Reference, force bool) (*github.Reference, *github.Response, error) {
	return nil, nil, nil
}
