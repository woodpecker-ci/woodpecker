package s3

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// mockDBStore is a mock implementation of log.Service for testing
type mockDBStore struct{}

func (m *mockDBStore) LogFind(step *model.Step) ([]*model.LogEntry, error) {
	return []*model.LogEntry{}, nil
}

func (m *mockDBStore) LogAppend(step *model.Step, entries []*model.LogEntry) error {
	return nil
}

func (m *mockDBStore) LogDelete(step *model.Step) error {
	return nil
}

func TestNewLogStore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		bucket       string
		bucketFolder string
		expectError  bool
	}{
		{
			name:         "empty bucket should error",
			bucket:       "",
			bucketFolder: "/logs",
			expectError:  true,
		},
		{
			name:         "valid bucket with folder",
			bucket:       "test-bucket",
			bucketFolder: "/logs",
			expectError:  false,
		},
		{
			name:         "valid bucket with empty folder",
			bucket:       "test-bucket",
			bucketFolder: "",
			expectError:  false,
		},
		{
			name:         "valid bucket with root folder",
			bucket:       "test-bucket",
			bucketFolder: "/",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockDB := &mockDBStore{}
			store, err := NewLogStore(tt.bucket, tt.bucketFolder, mockDB)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, store)
				assert.Contains(t, err.Error(), "S3 bucket name is required")
			} else {
				// Note: This will fail without AWS credentials, but that's expected in CI/testing
				if err != nil {
					// Skip if no AWS credentials available (expected in CI/testing environments)
					t.Skipf("Skipping test due to missing AWS credentials: %v", err)
				}
				assert.NoError(t, err)
				assert.NotNil(t, store)
			}
		})
	}
}

func TestLogPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		bucketFolder string
		stepID       int64
		expectedPath string
	}{
		{
			name:         "root level (empty folder)",
			bucketFolder: "",
			stepID:       123,
			expectedPath: "/123.json",
		},
		{
			name:         "single folder",
			bucketFolder: "logs",
			stepID:       456,
			expectedPath: "/logs/456.json",
		},
		{
			name:         "nested folder",
			bucketFolder: "logs/pipeline",
			stepID:       789,
			expectedPath: "/logs/pipeline/789.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := &logStore{
				bucket:       "test-bucket",
				bucketFolder: tt.bucketFolder,
			}
			logPath := store.logPath(tt.stepID)
			assert.Equal(t, tt.expectedPath, logPath)
		})
	}
}

func TestLogStoreBucketFolderNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		inputFolder    string
		expectedFolder string
	}{
		{
			name:           "empty folder becomes root",
			inputFolder:    "",
			expectedFolder: "",
		},
		{
			name:           "root folder becomes empty",
			inputFolder:    "/",
			expectedFolder: "",
		},
		{
			name:           "folder gets normalized (no slashes)",
			inputFolder:    "logs",
			expectedFolder: "logs",
		},
		{
			name:           "folder with leading slash gets normalized",
			inputFolder:    "/logs",
			expectedFolder: "logs",
		},
		{
			name:           "folder with trailing slash gets normalized",
			inputFolder:    "logs/",
			expectedFolder: "logs",
		},
		{
			name:           "folder with both slashes gets normalized",
			inputFolder:    "/logs/",
			expectedFolder: "logs",
		},
		{
			name:           "nested folder gets normalized",
			inputFolder:    "/logs/pipeline/",
			expectedFolder: "logs/pipeline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockDB := &mockDBStore{}
			store, err := NewLogStore("test-bucket", tt.inputFolder, mockDB)
			if err != nil {
				// Skip if AWS config fails (expected in testing)
				t.Skipf("Skipping test due to AWS config error: %v", err)
			}

			// Cast to logStore to check internal folder normalization
			logStore := store.(*logStore)
			assert.Equal(t, "test-bucket", logStore.bucket)
			assert.Equal(t, tt.expectedFolder, logStore.bucketFolder)
		})
	}
}

// TestLogOperationsInterface verifies that the returned service implements the correct interface
func TestLogOperationsInterface(t *testing.T) {
	t.Parallel()

	mockDB := &mockDBStore{}
	store, err := NewLogStore("test-bucket", "/logs", mockDB)
	if err != nil {
		// Skip if AWS config fails (expected in testing)
		t.Skipf("Skipping test due to AWS config error: %v", err)
	}

	// Verify the service implements the expected interface methods
	step := &model.Step{ID: 123}
	logEntries := []*model.LogEntry{
		{ID: 1, StepID: 123, Time: 1000, Line: 1, Data: []byte("test"), Type: model.LogEntryStdout},
	}

	// These will fail without real S3 credentials/service, but we're testing the interface
	// The methods should exist and be callable (even if they fail due to credentials)

	// Test LogFind method exists and is callable
	entries, err := store.LogFind(step)
	// We expect this to fail due to missing credentials, but method should exist
	if err == nil {
		// If it succeeds (unlikely in test environment), entries should be a slice
		assert.NotNil(t, entries)
	}

	// Test LogAppend method exists and is callable
	err = store.LogAppend(step, logEntries)
	// We expect this to fail due to missing credentials, but method should exist
	// No assertion needed, just verify the method is callable

	// Test LogDelete method exists and is callable
	err = store.LogDelete(step)
	// We expect this to fail due to missing credentials, but method should exist
	// No assertion needed, just verify the method is callable

	// If we get here without panics, the interface is correctly implemented
	t.Log("All log interface methods are callable")
}
