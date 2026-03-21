package output

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

type CloudWatchWriter struct {
	client        *cloudwatchlogs.Client
	logGroup      string
	logStream     string
	sequenceToken *string
	mu            sync.Mutex
	batchSize     int
	batchBuffer   []types.InputLogEvent
	flushInterval time.Duration
	lastFlush     time.Time
	useHTTP       bool
	httpEndpoint  string
	httpClient    *http.Client
	awsRegion     string
}

type CloudWatchConfig struct {
	Region          string
	LogGroup        string
	LogStream       string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseHTTP         bool
}

func NewCloudWatchWriter(cfg CloudWatchConfig) (*CloudWatchWriter, error) {
	writer := &CloudWatchWriter{
		logGroup:      cfg.LogGroup,
		logStream:     cfg.LogStream,
		batchSize:     100,
		batchBuffer:   make([]types.InputLogEvent, 0, 100),
		flushInterval: 5 * time.Second,
		lastFlush:     time.Now(),
		useHTTP:       cfg.UseHTTP,
		httpEndpoint:  cfg.Endpoint,
		awsRegion:     cfg.Region,
	}

	if !cfg.UseHTTP {
		var awsCfg aws.Config
		var err error

		ctx := context.Background()

		if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
			awsCfg, err = config.LoadDefaultConfig(ctx,
				config.WithRegion(cfg.Region),
				config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
					cfg.AccessKeyID,
					cfg.SecretAccessKey,
					"",
				)),
			)
		} else {
			awsCfg, err = config.LoadDefaultConfig(ctx,
				config.WithRegion(cfg.Region),
			)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}

		writer.client = cloudwatchlogs.NewFromConfig(awsCfg)

		if err := writer.ensureLogGroup(ctx); err != nil {
			return nil, err
		}

		if err := writer.ensureLogStream(ctx); err != nil {
			return nil, err
		}
	} else {
		writer.httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return writer, nil
}

func (w *CloudWatchWriter) ensureLogGroup(ctx context.Context) error {
	_, err := w.client.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(w.logGroup),
	})
	if err != nil {
		var alreadyExists *types.ResourceAlreadyExistsException
		if ok := errorAs(err, &alreadyExists); !ok {
			return fmt.Errorf("failed to create log group: %w", err)
		}
	}

	_, err = w.client.PutRetentionPolicy(ctx, &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    aws.String(w.logGroup),
		RetentionInDays: aws.Int32(1),
	})
	if err != nil {
		var alreadyExists *types.ResourceAlreadyExistsException
		if ok := errorAs(err, &alreadyExists); !ok {
		}
	}

	return nil
}

func (w *CloudWatchWriter) ensureLogStream(ctx context.Context) error {
	_, err := w.client.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(w.logGroup),
		LogStreamName: aws.String(w.logStream),
	})
	if err != nil {
		var alreadyExists *types.ResourceAlreadyExistsException
		if ok := errorAs(err, &alreadyExists); !ok {
			return fmt.Errorf("failed to create log stream: %w", err)
		}
	}

	describeResp, err := w.client.DescribeLogStreams(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(w.logGroup),
		LogStreamNamePrefix: aws.String(w.logStream),
	})
	if err == nil && len(describeResp.LogStreams) > 0 {
		w.sequenceToken = describeResp.LogStreams[0].UploadSequenceToken
	}

	return nil
}

func (w *CloudWatchWriter) WriteWebServerLog(message string) error {
	timestamp := time.Now().UnixMilli()
	return w.writeEvent(timestamp, message)
}

func (w *CloudWatchWriter) WriteErrorLog(message string) error {
	timestamp := time.Now().UnixMilli()
	return w.writeEvent(timestamp, message)
}

func (w *CloudWatchWriter) WriteCustomLog(message string) error {
	timestamp := time.Now().UnixMilli()
	return w.writeEvent(timestamp, message)
}

func (w *CloudWatchWriter) WriteMetricEMF(emf string) error {
	timestamp := time.Now().UnixMilli()
	return w.writeEvent(timestamp, emf)
}

func (w *CloudWatchWriter) writeEvent(timestamp int64, message string) error {
	if w.useHTTP {
		return w.writeViaHTTP(timestamp, message)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.batchBuffer = append(w.batchBuffer, types.InputLogEvent{
		Timestamp: aws.Int64(timestamp),
		Message:   aws.String(message),
	})

	if len(w.batchBuffer) >= w.batchSize || time.Since(w.lastFlush) >= w.flushInterval {
		if err := w.flush(); err != nil {
			return err
		}
	}

	return nil
}

func (w *CloudWatchWriter) flush() error {
	if len(w.batchBuffer) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	input := &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(w.logGroup),
		LogStreamName: aws.String(w.logStream),
		LogEvents:     w.batchBuffer,
	}

	if w.sequenceToken != nil {
		input.SequenceToken = w.sequenceToken
	}

	resp, err := w.client.PutLogEvents(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put log events: %w", err)
	}

	w.batchBuffer = w.batchBuffer[:0]
	w.sequenceToken = resp.NextSequenceToken
	w.lastFlush = time.Now()

	return nil
}

func (w *CloudWatchWriter) writeViaHTTP(timestamp int64, message string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	payload := map[string]interface{}{
		"log_group":  w.logGroup,
		"log_stream": w.logStream,
		"timestamp":  timestamp,
		"message":    message,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal log event: %w", err)
	}

	req, err := http.NewRequest("POST", w.httpEndpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send log event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send log event: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (w *CloudWatchWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.flush()
}

func (w *CloudWatchWriter) Close() error {
	return w.Flush()
}

func errorAs(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	_, ok := target.(*types.ResourceAlreadyExistsException)
	return ok
}
