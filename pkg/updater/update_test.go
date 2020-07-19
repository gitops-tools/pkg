package updater

import (
	"context"
	"errors"
	"testing"

	"github.com/bigkevmcd/common/pkg/client/mock"
	"github.com/jenkins-x/go-scm/scm"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	testGitHubRepo = "testorg/testrepo"
	testFilePath   = "environments/test/services/service-a/test.yaml"
	testBranch     = "main"
)

func TestApplyUpdateToFile(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, testBranch, []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, testBranch, testSHA)
	updater := New(zap.New(), m, NameGenerator(stubNameGenerator{"a"}))
	newBody := []byte("new content")

	branch, err := updater.ApplyUpdateToFile(context.Background(), makeCommitInput(), func([]byte) ([]byte, error) {
		return newBody, nil
	})

	if err != nil {
		t.Fatal(err)
	}
	if branch != "test-branch-a" {
		t.Fatalf("newly created branch, got %#v, want %#v", branch, "test-branch-a")
	}
	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	if s := string(updated); s != string(newBody) {
		t.Fatalf("update failed, got %#v, want %#v", s, newBody)
	}
	m.AssertBranchCreated(testGitHubRepo, "test-branch-a", testSHA)
	m.AssertNoPullRequestsCreated()
}

func TestApplyUpdateToFileMissingFile(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, testBranch, []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, testBranch, testSHA)
	updater := New(zap.New(), m, NameGenerator(stubNameGenerator{"a"}))
	testErr := errors.New("missing file")
	m.GetFileErr = testErr

	_, err := updater.ApplyUpdateToFile(context.Background(), makeCommitInput(), func([]byte) ([]byte, error) {
		return []byte("testing"), nil
	})

	if err != testErr {
		t.Fatalf("got %s, want %s", err, testErr)
	}
	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	if s := string(updated); s != "" {
		t.Fatalf("update failed, got %#v, want %#v", s, "")
	}
	m.AssertNoBranchesCreated()
	m.AssertNoPullRequestsCreated()
}

func TestApplyUpdateToFileWithBranchCreationFailure(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, testBranch, []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, testBranch, testSHA)
	updater := New(zap.New(), m, NameGenerator(stubNameGenerator{"a"}))
	testErr := errors.New("can't create branch")
	m.CreateBranchErr = testErr

	_, err := updater.ApplyUpdateToFile(context.Background(), makeCommitInput(), func([]byte) ([]byte, error) {
		return []byte("testing"), nil
	})

	if err.Error() != "failed to create branch: can't create branch" {
		t.Fatalf("got %s, want %s", err, "failed to create branch: can't create branch")
	}
	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	if s := string(updated); s != "" {
		t.Fatalf("update failed, got %#v, want %#v", s, "")
	}
	m.AssertNoBranchesCreated()
	m.AssertNoPullRequestsCreated()
}

func TestCreatePullRequest(t *testing.T) {
	m := mock.New(t)
	updater := New(zap.New(), m, NameGenerator(stubNameGenerator{"a"}))
	input := makePullRequestInput()

	pr, err := updater.CreatePR(context.Background(), input)

	if err != nil {
		t.Fatal(err)
	}
	m.AssertPullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Title: input.Title,
		Body:  input.Body,
		Head:  "test-branch-a",
		Base:  testBranch,
	})
	if pr.Link != "https://example.com/pull-request/1" {
		t.Fatalf("link to PR is incorrect: got %#v, want %#v", pr.Link, "https://example.com/pull-request/1")
	}
}

func TestCreatePullRequestHandlingErrors(t *testing.T) {
	m := mock.New(t)
	updater := New(zap.New(), m, NameGenerator(stubNameGenerator{"a"}))
	input := makePullRequestInput()
	testErr := errors.New("can't create pull-request")
	m.CreatePullRequestErr = testErr

	_, err := updater.CreatePR(context.Background(), input)

	if err.Error() != "failed to create a pull request: can't create pull-request" {
		t.Fatalf("got %s, want %s", err, "failed to create a pull request: can't create pull-request")
	}
}

type stubNameGenerator struct {
	name string
}

func (s stubNameGenerator) PrefixedName(p string) string {
	return p + s.name
}

func makeCommitInput() CommitInput {
	return CommitInput{
		Repo:               testGitHubRepo,
		Filename:           testFilePath,
		Branch:             testBranch,
		BranchGenerateName: "test-branch-",
		CommitMessage:      "just a test commit",
	}
}

func makePullRequestInput() PullRequestInput {
	return PullRequestInput{
		Repo:         testGitHubRepo,
		NewBranch:    "test-branch-a",
		SourceBranch: testBranch,
		Title:        "This is a test PR",
		Body:         "This is the body",
	}
}
