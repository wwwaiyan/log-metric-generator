package output

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

type StdoutWriter struct {
	mu        sync.Mutex
	logGroup  string
	logStream string
	useJSON   bool
}

func NewStdoutWriter(logGroup, logStream string) *StdoutWriter {
	return &StdoutWriter{
		logGroup:  logGroup,
		logStream: logStream,
		useJSON:   true,
	}
}

func (w *StdoutWriter) WriteWebServerLog(line string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	if w.useJSON {
		fmt.Println(logToJSON(w.logGroup, w.logStream, timestamp, line))
	} else {
		fmt.Println(logToPlainText(w.logGroup, w.logStream, timestamp, line))
	}
	return nil
}

func (w *StdoutWriter) WriteErrorLog(message string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	fmt.Println(logToJSON(w.logGroup, w.logStream, timestamp, message))
	return nil
}

func (w *StdoutWriter) WriteCustomLog(message string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	fmt.Println(logToJSON(w.logGroup, w.logStream, timestamp, message))
	return nil
}

func (w *StdoutWriter) WriteMetricEMF(emf string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	fmt.Println(logToJSON(w.logGroup, w.logStream, timestamp, emf))
	return nil
}

func (w *StdoutWriter) Flush() error {
	return nil
}

func (w *StdoutWriter) Close() error {
	return nil
}

func logToJSON(logGroup, logStream, timestamp, message string) string {
	entry := map[string]interface{}{
		"log_group":  logGroup,
		"log_stream": logStream,
		"timestamp":  timestamp,
		"message":    message,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Sprintf(`{"log_group":"%s","log_stream":"%s","message":"%s"}`,
			logGroup, logStream, message)
	}
	return string(data)
}

func logToPlainText(logGroup, logStream, timestamp, message string) string {
	return fmt.Sprintf("[%s] [%s] [%s] %s",
		timestamp, logGroup, logStream, message)
}

type FileWriter struct {
	mu        sync.Mutex
	file      *os.File
	logGroup  string
	logStream string
}

func NewFileWriter(path, logGroup, logStream string) (*FileWriter, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &FileWriter{
		file:      file,
		logGroup:  logGroup,
		logStream: logStream,
	}, nil
}

func (w *FileWriter) WriteWebServerLog(line string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	_, err := w.file.WriteString(line + "\n")
	return err
}

func (w *FileWriter) WriteErrorLog(line string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	_, err := w.file.WriteString(line + "\n")
	return err
}

func (w *FileWriter) WriteCustomLog(message string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	line := fmt.Sprintf("[%s] %s\n", timestamp, message)
	_, err := w.file.WriteString(line)
	return err
}

func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}
