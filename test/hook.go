package test

import (
	"testing"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
)

// MakeHookFromFixture creates and returns a WebHook parsed from the provided
// fixture file, with the correct X-GitHub-Event type etc.
func MakeHookFromFixture(t *testing.T, filename, eventType string) scm.Webhook {
	t.Helper()
	req := MakeHookRequest(t, filename, eventType)
	scmClient, err := factory.NewClient("github", "", "")
	if err != nil {
		t.Fatal(err)
	}
	hook, err := scmClient.Webhooks.Parse(req, func(_ scm.Webhook) (string, error) {
		return "", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return hook
}
