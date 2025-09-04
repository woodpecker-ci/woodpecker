package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	logger "github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/log"
)

type logStore struct {
	client       *s3.Client
	bucket       string
	bucketFolder string
	dbStore      log.Service
	uploader     *manager.Uploader
	downloader   *manager.Downloader
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
	uploader := manager.NewUploader(s3Client)
	downloader := manager.NewDownloader(s3Client)

	return &logStore{
		client:       s3Client,
		bucket:       bucket,
		bucketFolder: folder,
		dbStore:      dbStore,
		uploader:     uploader,
		downloader:   downloader,
	}, nil
}

func (l *logStore) logPath(stepID int64) string {
	return "/" + path.Join(l.bucketFolder, fmt.Sprintf("%d.json", stepID))
}

func (l *logStore) LogFind(step *model.Step) ([]*model.LogEntry, error) {
	dbEntries, dbErr := l.dbStore.LogFind(step)

	// If logs found in database, return them (step still active or upload failed)
	if dbErr == nil && len(dbEntries) > 0 {
		return dbEntries, nil
	}

	// No logs in database, try S3 (step completed and uploaded)
	logPath := l.logPath(step.ID)

	buf := manager.NewWriteAtBuffer(nil)
	_, err := l.downloader.Download(context.TODO(), buf, &s3.GetObjectInput{
		Bucket: aws.String(l.bucket),
		Key:    aws.String(logPath),
	})

	if err != nil {
		logger.Debug().Err(err).Str("logPath", logPath).Msg("failed to download from S3")
		return []*model.LogEntry{}, nil
	}

	var entries []*model.LogEntry
	data := buf.Bytes()
	lines := bytes.Split(data, []byte("\n"))

	for _, line := range lines {
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

	logger.Debug().Int64("stepID", step.ID).Int("entryCount", len(entries)).Msg("downloaded logs from S3")
	return entries, nil
}

func (l *logStore) LogAppend(step *model.Step, entries []*model.LogEntry) error {
	// First append to database
	if err := l.dbStore.LogAppend(step, entries); err != nil {
		return err
	}

	// Check if this batch contains an exit code entry (indicates step completion)
	hasExitCode := false
	for _, entry := range entries {
		if entry.Type == model.LogEntryExitCode {
			hasExitCode = true
			break
		}
	}

	// If step is completed, upload all logs to S3 and cleanup DB
	if hasExitCode {
		logger.Info().Int64("stepID", step.ID).Msg("step completed, uploading logs to S3")

		// Get all logs for this step from database
		dbEntries, err := l.dbStore.LogFind(step)
		if err != nil {
			logger.Error().Err(err).Int64("stepID", step.ID).Msg("failed to retrieve logs from DB for S3 upload")
			return nil // Don't fail the append, logs are still in DB
		}

		if len(dbEntries) > 0 {
			var buf bytes.Buffer
			for _, entry := range dbEntries {
				if jsonBytes, err := json.Marshal(entry); err == nil {
					buf.Write(jsonBytes)
					buf.WriteByte('\n')
				} else {
					logger.Warn().Err(err).Msg("failed to marshal log entry")
				}
			}

			if buf.Len() > 0 {
				logPath := l.logPath(step.ID)
				_, err := l.uploader.Upload(context.TODO(), &s3.PutObjectInput{
					Bucket: aws.String(l.bucket),
					Key:    aws.String(logPath),
					Body:   bytes.NewReader(buf.Bytes()),
					Metadata: map[string]string{
						"step-id":     fmt.Sprintf("%d", step.ID),
						"entry-count": fmt.Sprintf("%d", len(dbEntries)),
						"uploaded-at": fmt.Sprintf("%d", time.Now().Unix()),
					},
				})

				if err != nil {
					logger.Warn().Err(err).Int64("stepID", step.ID).Msg("S3 upload failed, keeping in DB")
					return nil // Don't fail, logs are safe in DB
				}

				// Clean up database after successful upload
				if err := l.dbStore.LogDelete(step); err != nil {
					logger.Error().Err(err).Int64("stepID", step.ID).Msg("failed to cleanup DB after S3 upload")
				} else {
					logger.Debug().Int64("stepID", step.ID).Msg("successfully uploaded logs to S3 and cleaned up DB")
				}
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
