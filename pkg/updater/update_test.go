package updater

import (
	"context"
	"errors"
	"testing"

	"github.com/bigkevmcd/common/pkg/client/mock"
	"github.com/jenkins-x/go-scm/scm"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const (
	testQuayRepo   = "mynamespace/repository"
	testGitHubRepo = "testorg/testrepo"
	testFilePath   = "environments/test/services/service-a/test.yaml"
	testBranch     = "main"
)

func TestUpdateYAML(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, testBranch, []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, testBranch, testSHA)
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)).Sugar()

	updater := New(logger, m, NameGenerator(stubNameGenerator{"a"}))
	input := makeUpdateYAMLInput()
	pr, err := updater.UpdateYAML(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}

	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	want := "test:\n  image: test/my-test-image\n"
	if s := string(updated); s != want {
		t.Fatalf("update failed, got %#v, want %#v", s, want)
	}
	m.AssertBranchCreated(testGitHubRepo, "test-branch-a", testSHA)
	m.AssertPullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Title: input.PullRequest.Title,
		Body:  input.PullRequest.Body,
		Head:  "test-branch-a",
		Base:  testBranch,
	})
	if pr.Link != "https://example.com/pull-request/1" {
		t.Fatalf("link to PR is incorrect: got %#v, want %#v", pr.Link, "https://example.com/pull-request/1")
	}
}

func TestUpdateYAMLWithMissingFile(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, testBranch, []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, testBranch, testSHA)
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)).Sugar()

	updater := New(logger, m, NameGenerator(stubNameGenerator{"a"}))
	testErr := errors.New("missing file")
	m.GetFileErr = testErr

	_, err := updater.UpdateYAML(context.Background(), makeUpdateYAMLInput())

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

func TestUpdateYAMLWithBranchCreationFailure(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, testBranch, []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, testBranch, testSHA)
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)).Sugar()

	updater := New(logger, m, NameGenerator(stubNameGenerator{"a"}))
	testErr := errors.New("can't create branch")
	m.CreateBranchErr = testErr

	_, err := updater.UpdateYAML(context.Background(), makeUpdateYAMLInput())

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

func TestUpdateYAMLWithUpdateFileFailure(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, testBranch, []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, testBranch, testSHA)
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)).Sugar()

	updater := New(logger, m, NameGenerator(stubNameGenerator{"a"}))
	testErr := errors.New("can't update file")
	m.UpdateFileErr = testErr
	input := makeUpdateYAMLInput()

	_, err := updater.UpdateYAML(context.Background(), input)

	if err.Error() != "failed to update file: can't update file" {
		t.Fatalf("got %s, want %s", err, "failed to update file: can't update file")
	}
	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	if s := string(updated); s != "" {
		t.Fatalf("update failed, got %#v, want %#v", s, "")
	}
	m.AssertBranchCreated(testGitHubRepo, "test-branch-a", testSHA)
	m.AssertNoPullRequestsCreated()
}

func TestUpdateYAMLWithCreatePullRequestFailure(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, testBranch, []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, testBranch, testSHA)
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)).Sugar()

	updater := New(logger, m, NameGenerator(stubNameGenerator{"a"}))
	testErr := errors.New("can't create pull-request")
	m.CreatePullRequestErr = testErr
	input := makeUpdateYAMLInput()

	_, err := updater.UpdateYAML(context.Background(), input)

	if err.Error() != "failed to create a pull request: can't create pull-request" {
		t.Fatalf("got %s, want %s", err, "failed to create a pull request: can't create pull-request")
	}
	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	want := "test:\n  image: test/my-test-image\n"
	if s := string(updated); s != want {
		t.Fatalf("update failed, got %#v, want %#v", s, want)
	}
	m.AssertBranchCreated(testGitHubRepo, "test-branch-a", testSHA)
	m.AssertNoPullRequestsCreated()
}

func TestUpdateYAMLWithNonMainSourceBranch(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, "staging", []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, "staging", testSHA)
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)).Sugar()

	input := makeUpdateYAMLInput()
	input.Branch = "staging"
	updater := New(logger, m, NameGenerator(stubNameGenerator{"a"}))

	_, err := updater.UpdateYAML(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}

	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	want := "test:\n  image: test/my-test-image\n"
	if s := string(updated); s != want {
		t.Fatalf("update failed, got %#v, want %#v", s, want)
	}
	m.AssertBranchCreated(testGitHubRepo, "test-branch-a", testSHA)
	m.AssertPullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Title: input.PullRequest.Title,
		Body:  input.PullRequest.Body,
		Head:  "test-branch-a",
		Base:  "staging",
	})
}

func TestUpdateFile(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, testBranch, []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, testBranch, testSHA)
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)).Sugar()

	updater := New(logger, m, NameGenerator(stubNameGenerator{"a"}))
	input := makeUpdateFileInput()
	pr, err := updater.UpdateFile(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}

	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	if s := string(updated); s != string(input.Body) {
		t.Fatalf("update failed, got %#v, want %#v", s, input.Body)
	}
	m.AssertBranchCreated(testGitHubRepo, "test-branch-a", testSHA)
	m.AssertPullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Title: input.PullRequest.Title,
		Body:  input.PullRequest.Body,
		Head:  "test-branch-a",
		Base:  testBranch,
	})
	if pr.Link != "https://example.com/pull-request/1" {
		t.Fatalf("link to PR is incorrect: got %#v, want %#v", pr.Link, "https://example.com/pull-request/1")
	}
}

func TestApplyUpdateToFile(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, testBranch, []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, testBranch, testSHA)
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)).Sugar()

	updater := New(logger, m, NameGenerator(stubNameGenerator{"a"}))
	input := makeCommitInput()
	newBody := []byte("new content")
	pr, err := updater.ApplyUpdateToFile(context.Background(), input, func([]byte) ([]byte, error) {
		return newBody, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	if s := string(updated); s != string(newBody) {
		t.Fatalf("update failed, got %#v, want %#v", s, newBody)
	}
	m.AssertBranchCreated(testGitHubRepo, "test-branch-a", testSHA)
	m.AssertPullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Title: input.PullRequest.Title,
		Body:  input.PullRequest.Body,
		Head:  "test-branch-a",
		Base:  testBranch,
	})
	if pr.Link != "https://example.com/pull-request/1" {
		t.Fatalf("link to PR is incorrect: got %#v, want %#v", pr.Link, "https://example.com/pull-request/1")
	}
}

type stubNameGenerator struct {
	name string
}

func (s stubNameGenerator) PrefixedName(p string) string {
	return p + s.name
}

func makeUpdateYAMLInput() *UpdateYAMLInput {
	return &UpdateYAMLInput{
		Key:         "test.image",
		NewValue:    "test/my-test-image",
		CommitInput: makeCommitInput(),
	}
}

func makeUpdateFileInput() *UpdateFileInput {
	return &UpdateFileInput{
		Body:        []byte("this is the new content"),
		CommitInput: makeCommitInput(),
	}
}

func makeCommitInput() CommitInput {
	return CommitInput{
		Repo:               testGitHubRepo,
		Filename:           testFilePath,
		Branch:             testBranch,
		BranchGenerateName: "test-branch-",
		CommitMessage:      "just a test commit",
		PullRequest: PullRequestInput{
			Title: "test pull-request",
			Body:  "test pull-request body",
		},
	}
}
