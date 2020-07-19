# common

This is a shared repository for tooling for interacting with Git.

This is alpha code, just extracted from another project for reuse.

## updater

This provides functionality for updating YAML files with a single call,
including updating the file and optionally opening a PR for the change.

```go
package main

import (
	"context"
	"log"

	"github.com/gitops-tools/common/pkg/client"
	"github.com/gitops-tools/common/pkg/updater"
	"github.com/jenkins-x/go-scm/scm/factory"
	"go.uber.org/zap"
)

func main() {
	cli, err := factory.NewClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	scmClient := client.New(cli)

	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	u := updater.New(sugar, scmClient)

	input := updater.Input{
		Repo:               "my-org/my-repo",
		Filename:           "service/deployment.yaml",
		Branch:             "main",
		Key:                "metadata.annotations.reviewed",
		NewValue:           "test-user"
		BranchGenerateName: "test-branch-",
		CommitMessage:      "testing a common component library",
		PullRequest: updater.PullRequestInput{
			Title: "This is a test",
			Body:  "No, really, this is just a test",
		},
	}

	pr, err := u.UpdateYAML(context.Background(), &input)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("pr.Link = %s", pr.Link)
}
```
