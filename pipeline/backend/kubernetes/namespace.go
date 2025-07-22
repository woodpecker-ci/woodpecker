package kubernetes

import (
	"context"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sNamespaceClient interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Namespace, error)
	Create(ctx context.Context, namespace *v1.Namespace, opts metav1.CreateOptions) (*v1.Namespace, error)
}

func mkNamespace(ctx context.Context, client K8sNamespaceClient, namespace string) error {
	_, err := client.Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		log.Trace().Str("namespace", namespace).Msg("Kubernetes namespace already exists")
		return nil
	}

	if !errors.IsNotFound(err) {
		log.Trace().Err(err).Str("namespace", namespace).Msg("failed to check Kubernetes namespace existence")
		return err
	}

	log.Trace().Str("namespace", namespace).Msg("creating Kubernetes namespace")

	_, err = client.Create(ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: namespace},
	}, metav1.CreateOptions{})
	if err != nil {
		log.Error().Err(err).Str("namespace", namespace).Msg("failed to create Kubernetes namespace")
		return err
	}

	log.Trace().Str("namespace", namespace).Msg("Kubernetes namespace created successfully")
	return nil
}
