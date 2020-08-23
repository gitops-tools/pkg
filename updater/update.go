package updater

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-logr/logr"
	"github.com/jenkins-x/go-scm/scm"

	"github.com/gitops-tools/pkg/client"
	"github.com/gitops-tools/pkg/names"
)

// ContentUpdater takes an existing body, it should transform it, and return the
// updated body.
type ContentUpdater func([]byte) ([]byte, error)

// UpdaterFunc is an option for for creating new Updaters.
type UpdaterFunc func(u *Updater)

// CommitInput is used to configure the commit and pull request.
type CommitInput struct {
	Repo               string // e.g. my-org/my-repo
	Filename           string // relative path to the file in the repository
	Branch             string // e.g. main
	BranchGenerateName string // e.g. update-image-
	CommitMessage      string // This is used for the commit when updating the file
}

// PullRequestInput provides configuration for the PullRequest to be opened.
type PullRequestInput struct {
	SourceBranch string // e.g. 'main'
	NewBranch    string
	Repo         string // e.g. my-org/my-repo
	Title        string
	Body         string
}

var timeSeed = rand.New(rand.NewSource(time.Now().UnixNano()))

// NameGenerator is an option func for the Updater creation function.
func NameGenerator(g names.Generator) UpdaterFunc {
	return func(u *Updater) {
		u.nameGenerator = g
	}
}

// New creates and returns a new Updater.
func New(l logr.Logger, c client.GitClient, opts ...UpdaterFunc) *Updater {
	u := &Updater{gitClient: c, nameGenerator: names.New(timeSeed), log: l}
	for _, o := range opts {
		o(u)
	}
	return u
}

// Updater can update a Git repo with an updated version of a file.
type Updater struct {
	gitClient     client.GitClient
	nameGenerator names.Generator
	log           logr.Logger
}

// ApplyUpdateToFile does the job of fetching the existing file, passing it to a
// user-provided function, and optionally creating a PR.
func (u *Updater) ApplyUpdateToFile(ctx context.Context, input CommitInput, f ContentUpdater) (string, error) {
	current, err := u.gitClient.GetFile(ctx, input.Repo, input.Branch, input.Filename)
	if err != nil {
		u.log.Info("failed to get file from repo", "err", err)
		return "", err
	}
	u.log.Info("got existing file", "sha", current.Sha)
	updated, err := f(current.Data)
	if err != nil {
		return "", err
	}
	return u.applyUpdate(ctx, input, current.Sha, updated)
}

func (u *Updater) applyUpdate(ctx context.Context, input CommitInput, currentSHA string, newBody []byte) (string, error) {
	branchRef, err := u.gitClient.GetBranchHead(ctx, input.Repo, input.Branch)
	if err != nil {
		return "", fmt.Errorf("failed to get branch head: %v", err)
	}
	newBranchName, err := u.createBranchIfNecessary(ctx, input, branchRef)
	if err != nil {
		return "", err
	}
	err = u.gitClient.UpdateFile(ctx, input.Repo, newBranchName, input.Filename, input.CommitMessage, currentSHA, newBody)
	if err != nil {
		return "", fmt.Errorf("failed to update file: %w", err)
	}
	u.log.Info("updated file", "filename", input.Filename)
	return newBranchName, nil
}

func (u *Updater) createBranchIfNecessary(ctx context.Context, input CommitInput, sourceRef string) (string, error) {
	if input.BranchGenerateName == "" {
		u.log.Info("no branchGenerateName configured, reusing source branch", "branch", input.Branch)
		return input.Branch, nil
	}

	newBranchName := u.nameGenerator.PrefixedName(input.BranchGenerateName)
	u.log.Info("generating new branch", "name", newBranchName)
	err := u.gitClient.CreateBranch(ctx, input.Repo, newBranchName, sourceRef)
	if err != nil {
		return "", fmt.Errorf("failed to create branch: %w", err)
	}
	u.log.Info("created branch", "branch", newBranchName, "ref", sourceRef)
	return newBranchName, nil
}

func (u *Updater) CreatePR(ctx context.Context, input PullRequestInput) (*scm.PullRequest, error) {
	pr, err := u.gitClient.CreatePullRequest(ctx, input.Repo, &scm.PullRequestInput{
		Title: input.Title,
		Body:  input.Body,
		Head:  input.NewBranch,
		Base:  input.SourceBranch,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create a pull request: %w", err)
	}
	u.log.Info("created PullRequest", "number", pr.Number)
	return pr, nil
}
