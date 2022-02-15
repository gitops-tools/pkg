package kubeconfig

import (
	"context"
	"testing"

	"github.com/gitops-tools/pkg/test"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestClientFromSecret(t *testing.T) {
	testEnv := &envtest.Environment{}
	testEnv.ControlPlane.GetAPIServer().Configure().Append("--authorization-mode=RBAC")
	testCfg, err := testEnv.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer testEnv.Stop()

	cl, err := client.New(testCfg, client.Options{Scheme: runtimeScheme})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("parsing valid secret", func(t *testing.T) {
		secretName := types.NamespacedName{
			Namespace: "default",
			Name:      "test-kubeconfig",
		}
		b := generateKubeConfig(t, "default", testCfg)
		testWriteSecret(t, b, cl, secretName)

		parsedClient, err := ClientFromSecret(context.TODO(), cl, secretName, ClientOptions{Key: "kube.config"})
		if err != nil {
			t.Fatal(err)
		}
		assertSecretReadable(t, parsedClient, secretName)
	})

	t.Run("missing secret", func(t *testing.T) {
		secretName := types.NamespacedName{
			Namespace: "default",
			Name:      "test-kubeconfig",
		}

		_, err = ClientFromSecret(context.TODO(), cl, secretName, ClientOptions{Key: "kube.config"})
		if !apierrors.IsNotFound(err) {
			t.Fatalf("got %s, want not found", err)
		}
	})

	t.Run("missing key in secret", func(t *testing.T) {
		secretName := types.NamespacedName{
			Namespace: "default",
			Name:      "test-kubeconfig",
		}
		b := generateKubeConfig(t, "default", testCfg)
		testWriteSecret(t, b, cl, secretName)

		_, err = ClientFromSecret(context.TODO(), cl, secretName, ClientOptions{Key: "kubeconfig"})
		assertErrorMatch(t, `secret.*doesn't contain a KubeConfig, missing key "kubeconfig"`, err)
	})

	t.Run("invalid secret", func(t *testing.T) {
		secretName := types.NamespacedName{
			Namespace: "default",
			Name:      "test-kubeconfig",
		}
		testWriteSecret(t, []byte("not a real secret"), cl, secretName)

		_, err = ClientFromSecret(context.TODO(), cl, secretName, ClientOptions{Key: "kube.config"})
		assertErrorMatch(t, "failed to parse REST Config from secret data", err)
	})

	t.Run("impersonating", func(t *testing.T) {
		secretName := types.NamespacedName{
			Namespace: "default",
			Name:      "test-kubeconfig",
		}
		b := generateKubeConfig(t, "default", testCfg)
		testWriteSecret(t, b, cl, secretName)
		writeRBACForSecret(t, cl, secretName)

		parsedClient, err := ClientFromSecret(context.TODO(), cl, secretName,
			ClientOptions{
				Key:  "kube.config",
				REST: RESTOptions{Impersonate: &rest.ImpersonationConfig{UserName: "test-user"}}})
		if err != nil {
			t.Fatal(err)
		}
		assertSecretReadable(t, cl, secretName)
		assertSecretReadable(t, parsedClient, secretName)
	})
}

func assertErrorMatch(t *testing.T, s string, e error) {
	if !test.MatchError(t, s, e) {
		t.Fatalf("failed to match error %s against %s", e, s)
	}
}

func writeRBACForSecret(t *testing.T, cl client.Client, obj types.NamespacedName) {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: "test-role", Namespace: obj.Namespace},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{""},
				Resources:     []string{"secrets"},
				Verbs:         []string{"get"},
				ResourceNames: []string{obj.Name},
			},
		},
	}
	if err := cl.Create(context.TODO(), role); err != nil {
		t.Fatalf("failed to write role: %s", err)
	}
	binding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-role-binding", Namespace: obj.Namespace},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "User",
				Name:     "test-user",
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     role.Name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	if err := cl.Create(context.TODO(), binding); err != nil {
		t.Fatalf("failed to write role-binding: %s", err)
	}
}

func testWriteSecret(t *testing.T, b []byte, cl client.Client, name types.NamespacedName) {
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: name.Namespace,
			Name:      name.Name,
		},
		Data: map[string][]byte{
			"kube.config": b,
		},
	}
	if err := cl.Create(context.TODO(), &secret); err != nil {
		t.Fatalf("failed to write secret: %s", err)
	}
	t.Cleanup(func() {
		if err := cl.Delete(context.TODO(), &secret); err != nil {
			t.Fatalf("failed to delete secret %s: %s", name, err)
		}
	})
}

func assertSecretReadable(t *testing.T, cl client.Client, name types.NamespacedName) {
	t.Helper()
	var secret corev1.Secret
	if err := cl.Get(context.TODO(), name, &secret); err != nil {
		t.Fatalf("failed to read secret %s", name)
	}
}

func generateKubeConfig(t *testing.T, name string, rc *rest.Config) []byte {
	t.Helper()
	cluster := clientcmdapi.NewCluster()
	cluster.CertificateAuthorityData = rc.CAData
	cluster.Server = rc.Host

	authinfo := clientcmdapi.NewAuthInfo()
	authinfo.AuthProvider = rc.AuthProvider
	authinfo.ClientCertificateData = rc.CertData
	authinfo.ClientKeyData = rc.KeyData
	authinfo.Username = rc.Username
	authinfo.Password = rc.Password
	authinfo.Token = rc.BearerToken

	context := clientcmdapi.NewContext()
	context.Cluster = name
	context.AuthInfo = name

	clientConfig := clientcmdapi.NewConfig()
	clientConfig.Clusters[name] = cluster
	clientConfig.Contexts[name] = context
	clientConfig.AuthInfos[name] = authinfo
	clientConfig.CurrentContext = name

	b, err := clientcmd.Write(*clientConfig)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
