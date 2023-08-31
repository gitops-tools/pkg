package kubeconfig

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var (
	runtimeScheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(runtimeScheme))
}

// RESTOptions provides additional information for creating REST config during the client
// secret parsing.
type RESTOptions struct {
	Impersonate *rest.ImpersonationConfig
}

// ClientOptions can provide optional options when creating a client.
type ClientOptions struct {
	// REST is a set of RESTOptions.
	REST RESTOptions
	// Key is the key within the loaded secret to lookup the KubeConfig in.
	// If this key does not exist, an error is returned.
	Key string
}

// ClientFromSecret loads a secret from the provided name, and returns a parsed
// client.Client based on parsing the rest.Config from the secret.
func ClientFromSecret(ctx context.Context, cl client.Client, name types.NamespacedName, opts ClientOptions) (client.Client, error) {
	// TODO: take a logger?
	cfg, err := restConfigFromSecret(ctx, cl, name, opts)
	if err != nil {
		return nil, err
	}
	httpClient, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create an HTTP client for cluster: %w", err)
	}
	mapper, err := apiutil.NewDynamicRESTMapper(cfg, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST mapper: %w", err)
	}
	parsedClient, err := client.New(cfg, client.Options{Mapper: mapper})
	if err != nil {
		return nil, fmt.Errorf("failed to create a client: %w", err)
	}

	return parsedClient, nil
}

func restConfigFromSecret(ctx context.Context, cl client.Client, name types.NamespacedName, opts ClientOptions) (*rest.Config, error) {
	var secret corev1.Secret
	if err := cl.Get(ctx, name, &secret); err != nil {
		return nil, fmt.Errorf("unable to read KubeConfig secret %q error: %w", name, err)
	}

	b := secret.Data[opts.Key]
	if len(b) == 0 {
		return nil, fmt.Errorf("KubeConfig secret %q doesn't contain a KubeConfig, missing key %q", name, opts.Key)
	}

	cfg, err := clientcmd.RESTConfigFromKubeConfig(b)
	if err != nil {
		return nil, fmt.Errorf("failed to parse REST Config from secret data: %w", err)
	}
	if opts.REST.Impersonate != nil {
		cfg.Impersonate = *opts.REST.Impersonate
	}
	return cfg, nil
}
