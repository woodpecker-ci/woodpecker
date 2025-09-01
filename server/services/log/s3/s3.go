package s3

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	logger "github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/log"
)

type logStore struct {
	client       *s3.Client
	bucket       string
	bucketFolder string
}

const (
	maxLineLength    = (pipeline.MaxLogLineLength/3)*4 + (64 * 1024)
	metadataFileName = "metadata"
)

func NewLogStore(bucket, folder string) (log.Service, error) {
	if bucket == "" {
		return nil, fmt.Errorf("S3 bucket name is required")
	}

	logger.Info().Str("bucket", bucket).Str("folder", folder).Msg("initializing S3 log store")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logger.Error().Err(err).Msg("failed to load AWS config")
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")

	return &logStore{
		client:       s3.NewFromConfig(cfg),
		bucket:       bucket,
		bucketFolder: folder,
	}, nil
}

func (l *logStore) stepPrefix(stepID int64) string {
	prefix := fmt.Sprintf("%d/", stepID)
	if l.bucketFolder != "" {
		prefix = l.bucketFolder + "/" + prefix
	}
	return prefix
}

func (l *logStore) fileKey(stepID int64, fileNum int) string {
	return fmt.Sprintf("%s%d.json", l.stepPrefix(stepID), fileNum)
}

func (l *logStore) LogFind(step *model.Step) ([]*model.LogEntry, error) {
	last, err := l.getLastFileNumber(step.ID)
	if err != nil || last == 0 {
		return []*model.LogEntry{}, nil
	}

	var all []*model.LogEntry
	for i := 1; i <= last; i++ {
		key := l.fileKey(step.ID, i)
		res, err := l.client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(l.bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			logger.Error().Err(err).Str("key", key).Msg("failed to get log file")
			continue
		}
		scanner := bufio.NewScanner(res.Body)
		scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxLineLength)
		for scanner.Scan() {
			var e model.LogEntry
			if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
				all = append(all, &e)
			}
		}
		res.Body.Close()
	}
	return all, nil
}

func (l *logStore) LogAppend(step *model.Step, entries []*model.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}
	last, err := l.getLastFileNumber(step.ID)
	if err != nil {
		return err
	}
	next := last + 1

	var buf bytes.Buffer
	for _, e := range entries {
		if b, err := json.Marshal(e); err == nil {
			buf.Write(b)
			buf.WriteByte('\n')
		}
	}

	key := l.fileKey(step.ID, next)
	_, err = l.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(l.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		logger.Error().Err(err).Str("key", key).Msg("failed to upload log file")
		return err
	}

	_ = l.saveLastFileNumber(step.ID, next) // non-fatal if fails
	return nil
}

func (l *logStore) LogDelete(step *model.Step) error {
	last, err := l.getLastFileNumber(step.ID)
	var nums []int
	if err != nil {
		// fallback: list all files
		nums, err = l.listFiles(step.ID)
		if err != nil {
			return err
		}
	} else {
		for i := 1; i <= last; i++ {
			nums = append(nums, i)
		}
	}

	for _, n := range nums {
		key := l.fileKey(step.ID, n)
		_, _ = l.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
			Bucket: aws.String(l.bucket),
			Key:    aws.String(key),
		})
	}
	_, _ = l.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(l.bucket),
		Key:    aws.String(l.stepPrefix(step.ID) + metadataFileName),
	})
	return nil
}

// --- private helpers ---

func (l *logStore) getLastFileNumber(stepID int64) (int, error) {
	key := l.stepPrefix(stepID) + metadataFileName
	res, err := l.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(l.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			return l.rebuildMetadata(stepID)
		}
		return 0, err
	}
	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)
	n, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return l.rebuildMetadata(stepID)
	}
	return n, nil
}

func (l *logStore) saveLastFileNumber(stepID int64, n int) error {
	key := l.stepPrefix(stepID) + metadataFileName
	_, err := l.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(l.bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(strconv.Itoa(n)),
	})
	return err
}

func (l *logStore) rebuildMetadata(stepID int64) (int, error) {
	nums, err := l.listFiles(stepID)
	if err != nil {
		return 0, err
	}
	last := 0
	if len(nums) > 0 {
		last = nums[len(nums)-1]
	}
	_ = l.saveLastFileNumber(stepID, last)
	return last, nil
}

func (l *logStore) listFiles(stepID int64) ([]int, error) {
	prefix := l.stepPrefix(stepID)
	p := s3.NewListObjectsV2Paginator(l.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(l.bucket),
		Prefix: aws.String(prefix),
	})

	var nums []int
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		for _, o := range page.Contents {
			if o.Key == nil || strings.HasSuffix(*o.Key, metadataFileName) {
				continue
			}
			name := strings.TrimPrefix(*o.Key, prefix)
			if n, err := strconv.Atoi(strings.TrimSuffix(name, ".json")); err == nil {
				nums = append(nums, n)
			}
		}
	}
	sort.Ints(nums)
	return nums, nil
}
