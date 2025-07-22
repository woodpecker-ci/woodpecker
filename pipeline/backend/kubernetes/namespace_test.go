package kubernetes

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type mockNamespaceClient struct {
	getError     error
	createError  error
	getCalled    bool
	createCalled bool
	createdNS    *v1.Namespace
}

func (m *mockNamespaceClient) Get(_ context.Context, name string, _ metav1.GetOptions) (*v1.Namespace, error) {
	m.getCalled = true
	if m.getError != nil {
		return nil, m.getError
	}
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}, nil
}

func (m *mockNamespaceClient) Create(_ context.Context, ns *v1.Namespace, _ metav1.CreateOptions) (*v1.Namespace, error) {
	m.createCalled = true
	m.createdNS = ns
	return ns, m.createError
}

func TestMkNamespace(t *testing.T) {
	tests := []struct {
		name               string
		namespace          string
		setupMock          func(*mockNamespaceClient)
		expectError        bool
		errorContains      string
		expectGetCalled    bool
		expectCreateCalled bool
	}{
		{
			name:      "should succeed when namespace already exists",
			namespace: "existing-namespace",
			setupMock: func(m *mockNamespaceClient) {
				m.getError = nil // namespace exists
			},
			expectError:        false,
			expectGetCalled:    true,
			expectCreateCalled: false,
		},
		{
			name:      "should create namespace when it doesn't exist",
			namespace: "new-namespace",
			setupMock: func(m *mockNamespaceClient) {
				m.getError = k8serrors.NewNotFound(schema.GroupResource{Resource: "namespaces"}, "new-namespace")
				m.createError = nil
			},
			expectError:        false,
			expectGetCalled:    true,
			expectCreateCalled: true,
		},
		{
			name:      "should fail when Get namespace returns generic error",
			namespace: "error-namespace",
			setupMock: func(m *mockNamespaceClient) {
				m.getError = errors.New("api server unavailable")
			},
			expectError:        true,
			errorContains:      "api server unavailable",
			expectGetCalled:    true,
			expectCreateCalled: false,
		},
		{
			name:      "should fail when Create namespace returns error",
			namespace: "create-fail-namespace",
			setupMock: func(m *mockNamespaceClient) {
				m.getError = k8serrors.NewNotFound(schema.GroupResource{Resource: "namespaces"}, "create-fail-namespace")
				m.createError = errors.New("insufficient permissions")
			},
			expectError:        true,
			errorContains:      "insufficient permissions",
			expectGetCalled:    true,
			expectCreateCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockNamespaceClient{}
			tt.setupMock(client)

			err := mkNamespace(t.Context(), client, tt.namespace)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectGetCalled, client.getCalled, "Get call expectation")
			assert.Equal(t, tt.expectCreateCalled, client.createCalled, "Create call expectation")

			if tt.expectCreateCalled && client.createCalled {
				assert.NotNil(t, client.createdNS, "Created namespace should not be nil")
				assert.Equal(t, tt.namespace, client.createdNS.Name, "Created namespace should have correct name")
			}
		})
	}
}
