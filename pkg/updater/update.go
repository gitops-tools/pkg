package updater

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/bigkevmcd/common/pkg/client"
	"github.com/bigkevmcd/common/pkg/names"
	"github.com/bigkevmcd/common/pkg/syaml"
	"github.com/jenkins-x/go-scm/scm"
)

// TODO: split this into separate commit / pull request logic.

// UpdateYAMLInput is the configuration for updating a file in a repository.
type UpdateYAMLInput struct {
	Key      string // key - the key within the YAML file to be updated, use a dotted path
	NewValue interface{}
	CommitInput
}

// UpdateFileInput is the configuration for updating a file in a repository.
type UpdateFileInput struct {
	Body []byte // replacement file contents
	CommitInput
}

// ContentUpdater takes an existing body, it should transform it, and return the
// updated body.
type ContentUpdater func([]byte) ([]byte, error)

// UpdaterFunc is an option for for creating new Updaters.
type UpdaterFunc func(u *Updater)

// CommitInput is used to configure the commit and pull request.
type CommitInput struct {
	Repo               string           // e.g. my-org/my-repo
	Filename           string           // relative path to the file in the repository
	Branch             string           // e.g. main
	BranchGenerateName string           // e.g. update-image-
	CommitMessage      string           // This will be used for the commit to update the file
	PullRequest        PullRequestInput // This is used to create the pull request.
}

// PullRequestInput provides configuration for the PullRequest to be opened.
type PullRequestInput struct {
	Title string
	Body  string
}

var timeSeed = rand.New(rand.NewSource(time.Now().UnixNano()))

type logger interface {
	Errorw(msg string, keysAndValues ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
}

// NameGenerator is an option func for the Updater creation function.
func NameGenerator(g names.Generator) UpdaterFunc {
	return func(u *Updater) {
		u.nameGenerator = g
	}
}

// New creates and returns a new Updater.
func New(l logger, c client.GitClient, opts ...UpdaterFunc) *Updater {
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
	log           logger
}

// UpdateYAML does the job of fetching the existing file, updating it, and
// then optionally creating a PR.
func (u *Updater) UpdateYAML(ctx context.Context, input *UpdateYAMLInput) (*scm.PullRequest, error) {
	return u.ApplyUpdateToFile(ctx, input.CommitInput, func(b []byte) ([]byte, error) {
		return syaml.SetBytes(b, input.Key, input.NewValue)
	})
}

// ApplyUpdateToFile does the job of fetching the existing file, passing it to a
// user-provided function, and optionally creating a PR.
func (u *Updater) ApplyUpdateToFile(ctx context.Context, input CommitInput, f ContentUpdater) (*scm.PullRequest, error) {
	current, err := u.gitClient.GetFile(ctx, input.Repo, input.Branch, input.Filename)
	if err != nil {
		u.log.Errorw("failed to get file from repo", "error", err)
		return nil, err
	}
	u.log.Debugw("got existing file", "sha", current.Sha)
	updated, err := f(current.Data)
	if err != nil {
		return nil, err
	}
	return u.applyUpdate(ctx, input, current.Sha, updated)
}

func (u *Updater) applyUpdate(ctx context.Context, input CommitInput, currentSHA string, newBody []byte) (*scm.PullRequest, error) {
	branchRef, err := u.gitClient.GetBranchHead(ctx, input.Repo, input.Branch)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch head: %v", err)
	}
	newBranchName, err := u.createBranchIfNecessary(ctx, input, branchRef)
	if err != nil {
		return nil, err
	}
	err = u.gitClient.UpdateFile(ctx, input.Repo, newBranchName, input.Filename, input.CommitMessage, currentSHA, newBody)
	if err != nil {
		return nil, fmt.Errorf("failed to update file: %w", err)
	}
	u.log.Debugw("updated file", "filename", input.Filename)
	return u.createPRIfNecessary(ctx, input, newBranchName)
}

// UpdateFile does the job of fetching the existing file, updating it, and
// then optionally creating a PR.
func (u *Updater) UpdateFile(ctx context.Context, input *UpdateFileInput) (*scm.PullRequest, error) {
	current, err := u.gitClient.GetFile(ctx, input.Repo, input.Branch, input.Filename)
	if err != nil {
		u.log.Errorw("failed to get file from repo", "error", err)
		return nil, err
	}
	return u.applyUpdate(ctx, input.CommitInput, current.Sha, input.Body)
}

func (u *Updater) createBranchIfNecessary(ctx context.Context, input CommitInput, sourceRef string) (string, error) {
	if input.BranchGenerateName == "" {
		u.log.Debugw("no branchGenerateName configured, reusing source branch", "branch", input.Branch)
		return input.Branch, nil
	}

	newBranchName := u.nameGenerator.PrefixedName(input.BranchGenerateName)
	u.log.Debugw("generating new branch", "name", newBranchName)
	err := u.gitClient.CreateBranch(ctx, input.Repo, newBranchName, sourceRef)
	if err != nil {
		return "", fmt.Errorf("failed to create branch: %w", err)
	}
	u.log.Debugw("created branch", "branch", newBranchName, "ref", sourceRef)
	return newBranchName, nil
}

func (u *Updater) createPRIfNecessary(ctx context.Context, input CommitInput, newBranchName string) (*scm.PullRequest, error) {
	if input.Branch == newBranchName {
		return nil, nil
	}
	pr, err := u.gitClient.CreatePullRequest(ctx, input.Repo, &scm.PullRequestInput{
		Title: input.PullRequest.Title,
		Body:  input.PullRequest.Body,
		Head:  newBranchName,
		Base:  input.Branch,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create a pull request: %w", err)
	}
	u.log.Debugw("created PullRequest", "number", pr.Number)
	return pr, nil
}
