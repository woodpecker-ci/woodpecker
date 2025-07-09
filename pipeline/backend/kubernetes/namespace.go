package kubernetes

import (
	"context"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ensureNamespaceExists(ctx context.Context, engine *kube, namespace string) error {
	// Check if a namespace already exists
	_, err := engine.client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		// Namespace already exists
		log.Trace().Str("namespace", namespace).Msg("Kubernetes namespace already exists")
		return nil
	}

	// If the error is not "not found", return the error
	if !errors.IsNotFound(err) {
		log.Trace().Err(err).Str("namespace", namespace).Msg("failed to check Kubernetes namespace existence")
		return err
	}

	// Namespace doesn't exist, create it
	log.Trace().Str("namespace", namespace).Msg("creating Kubernetes namespace")

	_, err = engine.client.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		log.Error().Err(err).Str("namespace", namespace).Msg("failed to create Kubernetes namespace")
		return err
	}

	log.Trace().Str("namespace", namespace).Msg("Kubernetes namespace created successfully")
	return nil
}
