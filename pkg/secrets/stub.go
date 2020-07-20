package secrets

import (
	"context"
	"errors"
	"strings"

	"k8s.io/apimachinery/pkg/types"
)

var _ SecretGetter = (*SecretsStub)(nil)

// NewSecretsStub creates and returns a new SecretsStub.
func NewSecretsStub() *SecretsStub {
	return &SecretsStub{
		secrets: map[string]string{},
	}
}

// SecretsStub is an implementation of the SecretGetter interface.
type SecretsStub struct {
	tokenErr error
	secrets  map[string]string
}

// SecretToken is an implementation of the SecretGetter interface.
func (s SecretsStub) SecretToken(ctx context.Context, id types.NamespacedName, key string) (string, error) {
	token, ok := s.secrets[stubKey(id, key)]
	if !ok {
		return "", errors.New("not found")
	}
	return token, s.tokenErr
}

// StubSecret stubs the return value for an id/key combination.
func (s *SecretsStub) StubSecret(id types.NamespacedName, key, token string) {
	s.secrets[stubKey(id, key)] = token
}

// StubError stubs the return value as an error.
func (s *SecretsStub) StubError(err error) {
	s.tokenErr = err
}

func stubKey(id types.NamespacedName, key string) string {
	return strings.Join([]string{id.Name, id.Namespace, key}, ":")
}
