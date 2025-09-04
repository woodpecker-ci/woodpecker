package s3

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	logger "github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/log"
)

const (
	// Add base64 overhead and space for other JSON fields (just to be safe).
	maxLineLength = (pipeline.MaxLogLineLength/3)*4 + (64 * 1024)
)

type logStore struct {
	client       *s3.Client
	bucket       string
	bucketFolder string
	dbStore      log.Service
}

func (l *logStore) logPath(stepID int64) string {
	return "/" + path.Join(l.bucketFolder, fmt.Sprintf("%d.json", stepID))
}

func NewLogStore(bucket, folder string, dbStore log.Service) (log.Service, error) {
	if bucket == "" {
		return nil, fmt.Errorf("S3 bucket name is required")
	}

	logger.Info().Str("bucket", bucket).Str("folder", folder).Msg("initializing S3 log store")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load S3 config: %w", err)
	}

	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")

	s3Client := s3.NewFromConfig(cfg)

	return &logStore{
		client:       s3Client,
		bucket:       bucket,
		bucketFolder: folder,
		dbStore:      dbStore,
	}, nil
}

func (l *logStore) LogFind(step *model.Step) ([]*model.LogEntry, error) {
	logPath := l.logPath(step.ID)

	response, err := l.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(l.bucket),
		Key:    aws.String(logPath),
	})

	if err == nil {
		defer response.Body.Close()

		var entries []*model.LogEntry
		scanner := bufio.NewScanner(response.Body)
		scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxLineLength)

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var entry model.LogEntry
			if err := json.Unmarshal(line, &entry); err == nil {
				entries = append(entries, &entry)
			} else {
				logger.Warn().Err(err).Msg("failed to unmarshal log entry from S3")
			}
		}

		if scanErr := scanner.Err(); scanErr != nil {
			logger.Error().Err(scanErr).Str("logPath", logPath).Msg("error reading from S3")
			return []*model.LogEntry{}, fmt.Errorf("error reading S3 content: %w", scanErr)
		}

		logger.Debug().Int64("stepID", step.ID).Int("entryCount", len(entries)).Msg("downloaded logs from S3")
		return entries, nil
	} else {
		logger.Debug().Err(err).Str("logPath", logPath).Int64("stepID", step.ID).Msg("S3 download failed, falling back to database")
		return l.dbStore.LogFind(step)
	}
}

func (l *logStore) LogAppend(step *model.Step, entries []*model.LogEntry) error {
	// First append to database
	if err := l.dbStore.LogAppend(step, entries); err != nil {
		return err
	}

	// Check if step is completed (Finished timestamp indicates completion)
	if step.Finished != 0 {
		logger.Info().Int64("stepID", step.ID).Msg("step completed, uploading logs to S3")

		// Get all logs for this step from database
		dbEntries, err := l.dbStore.LogFind(step)
		if err != nil {
			logger.Error().Err(err).Int64("stepID", step.ID).Msg("failed to retrieve logs from DB for S3 upload")
			return nil // Don't fail the append, logs are still in DB
		}

		if len(dbEntries) > 0 {
			logPath := l.logPath(step.ID)

			// Marshal all log entries to JSON lines
			var buf strings.Builder
			encoder := json.NewEncoder(&buf)
			for _, entry := range dbEntries {
				if err := encoder.Encode(entry); err != nil {
					logger.Warn().Err(err).Msg("failed to encode log entry")
					continue
				}
			}

			_, err := l.client.PutObject(context.TODO(), &s3.PutObjectInput{
				Bucket: aws.String(l.bucket),
				Key:    aws.String(logPath),
				Body:   strings.NewReader(buf.String()),
				Metadata: map[string]string{
					"step-id":     fmt.Sprintf("%d", step.ID),
					"entry-count": fmt.Sprintf("%d", len(dbEntries)),
					"uploaded-at": fmt.Sprintf("%d", time.Now().Unix()),
				},
			})

			// Always delete logs from database to avoid orphaned logs on last step log
			if deleteErr := l.dbStore.LogDelete(step); deleteErr != nil {
				logger.Error().Err(deleteErr).Int64("stepID", step.ID).Msg("failed to cleanup database")
			}

			if err != nil {
				return fmt.Errorf("S3 upload failed: %w", err)
			} else {
				logger.Debug().Int64("stepID", step.ID).Msg("successfully uploaded logs to S3 and cleaned up DB")
			}
		}
	}
	return nil
}

func (l *logStore) LogDelete(step *model.Step) error {
	if err := l.dbStore.LogDelete(step); err != nil {
		logger.Error().Err(err).Int64("stepID", step.ID).Msg("failed to delete from database")
		return fmt.Errorf("failed to delete from database: %w", err)
	}

	logPath := l.logPath(step.ID)
	if _, err := l.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(l.bucket),
		Key:    aws.String(logPath),
	}); err != nil {
		logger.Debug().Err(err).Str("logPath", logPath).Msg("failed to delete log from S3 (may not exist)")
	}

	return nil
}
