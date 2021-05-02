module github.com/gitops-tools/pkg

go 1.14

require (
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.0
	github.com/google/go-cmp v0.4.0
	github.com/google/go-github/v35 v35.1.0
	github.com/jenkins-x/go-scm v1.5.157
	github.com/tidwall/sjson v1.1.1
	go.uber.org/zap v1.15.0
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	gopkg.in/h2non/gock.v1 v1.0.15
	k8s.io/api v0.18.4
	k8s.io/apimachinery v0.18.4
	sigs.k8s.io/controller-runtime v0.6.1
	sigs.k8s.io/yaml v1.2.0
)
