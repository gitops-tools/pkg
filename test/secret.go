package test

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretOptFunc func(*corev1.Secret)

// CreateSecret creates and returns a new Secret with the provided key in the data.
//
// A slice of SecretOptFuncs can be used to modify the returned secret.
func CreateSecret(key string, opts ...SecretOptFunc) *corev1.Secret {
	s := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "testing",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			key: []byte(`secret-token`),
		},
	}
	for _, o := range opts {
		o(s)
	}
	return s
}
