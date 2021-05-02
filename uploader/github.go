package uploader

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v35/github"
)

var (
	fileMode = github.String("100644")
	blobType = github.String("blob")
)

// GitHubUploader is an implementation of the GitUploader
type GitHubUploader struct {
	client githubClient
	logger logr.Logger
}

// NewGitHubUploader creates and returns a GitHub uploader using the provided
// client and logger.
func NewGitHubUploader(c *github.Client, l logr.Logger) *GitHubUploader {
	gu := &GitHubUploader{
		client: githubUploaderClient{c: c},
		logger: l,
	}
	return gu
}

func (g *GitHubUploader) WriteFiles(ctx context.Context, files Files, ref RepoReference) (string, error) {
	l := g.logger.WithValues("repo", ref)
	if len(files) == 0 {
		l.V(DEBUG).Info("no files to write")
		return "", nil
	}

	repoSHA, _, err := g.client.GetCommitSHA1(ctx, ref.Owner, ref.Repo, "main", "")
	if err != nil {
		return "", fmt.Errorf("failed to GetCommitSHA1 for repo:branch %s: %w", ref, err)
	}
	l.V(DEBUG).Info("got commit", "sha", repoSHA)

	reference := &github.Reference{
		Ref: github.String("refs/heads/" + ref.Branch),
		Object: &github.GitObject{
			SHA: github.String(repoSHA),
		},
	}

	if _, _, err := g.client.CreateRef(ctx, ref.Owner, ref.Repo, reference); err != nil {
		return "", fmt.Errorf("failed to CreateRef for SHA %s in %s: %w", ref, *reference.Object.SHA, err)
	}

	var entries []*github.TreeEntry
	for filename, body := range files {
		sha, err := g.createBlob(ctx, ref, body)
		if err != nil {
			return "", fmt.Errorf("failed to createBlob for %s in %s: %w", filename, ref, err)
		}
		entry := treeEntryBlob(filename, sha)
		entries = append(entries, entry)
	}

	tree, _, err := g.client.CreateTree(ctx, ref.Owner, ref.Repo, repoSHA, entries)
	if err != nil {
		return "", fmt.Errorf("failed to CreateTree for SHA %s in  %s: %w", repoSHA, ref, err)
	}
	l.V(DEBUG).Info("tree created", "entries", len(entries))
	commit := &github.Commit{
		Message: github.String("just a commit"),
		Tree:    tree,
		Parents: []*github.Commit{&github.Commit{SHA: github.String(repoSHA)}},
	}

	newCommit, _, err := g.client.CreateCommit(ctx, ref.Owner, ref.Repo, commit)
	if err != nil {
		return "", fmt.Errorf("failed to CreateCommit in %s: %w", ref, err)
	}
	l.V(DEBUG).Info("commit created", "newCommit", newCommit)

	commitSHA := newCommit.GetSHA()
	branchRef := &github.Reference{
		Ref: github.String(fmt.Sprintf("refs/heads/%s", ref.Branch)),
		Object: &github.GitObject{
			Type: github.String("commit"),
			SHA:  github.String(commitSHA),
		},
	}
	_, _, err = g.client.UpdateRef(ctx, ref.Owner, ref.Repo, branchRef, false)
	if err != nil {
		return "", fmt.Errorf("failed to UpdateRef for ref %s: %w", ref, err)
	}
	return commitSHA, nil
}

func (g *GitHubUploader) createBlob(ctx context.Context, ref RepoReference, data []byte) (string, error) {
	newBlob := &github.Blob{
		Encoding: github.String("base64"),
		Content:  github.String(base64.StdEncoding.EncodeToString(data)),
	}
	blob, _, err := g.client.CreateBlob(ctx, ref.Owner, ref.Repo, newBlob)
	if err != nil {
		return "", fmt.Errorf("failed to create blob in %s: %w", ref, err)
	}
	return blob.GetSHA(), nil
}

func treeEntryBlob(path string, fileSHA string) *github.TreeEntry {
	return &github.TreeEntry{
		SHA:  github.String(fileSHA),
		Path: github.String(path),
		Mode: fileMode,
		Type: blobType,
	}
}
