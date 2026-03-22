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
	mu      sync.Mutex
	useJSON bool
}

func NewStdoutWriter() *StdoutWriter {
	return &StdoutWriter{
		useJSON: true,
	}
}

func (w *StdoutWriter) WriteWebServerLog(line string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	if w.useJSON {
		fmt.Println(logToJSON(timestamp, "INFO", line))
	} else {
		fmt.Println(logToPlainText(timestamp, "INFO", line))
	}
	return nil
}

func (w *StdoutWriter) WriteErrorLog(message string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	fmt.Println(logToJSON(timestamp, "ERROR", message))
	return nil
}

func (w *StdoutWriter) WriteCustomLog(message string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	fmt.Println(logToJSON(timestamp, "DEBUG", message))
	return nil
}

func (w *StdoutWriter) WriteMetricEMF(emf string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	fmt.Println(logToJSON(timestamp, "INFO", emf))
	return nil
}

func (w *StdoutWriter) Flush() error {
	return nil
}

func (w *StdoutWriter) Close() error {
	return nil
}

func logToJSON(timestamp, level, message string) string {
	entry := map[string]interface{}{
		"timestamp": timestamp,
		"level":     level,
		"message":   message,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Sprintf(`{"timestamp":"%s","level":"%s","message":"%s"}`,
			timestamp, level, message)
	}
	return string(data)
}

func logToPlainText(timestamp, level, message string) string {
	return fmt.Sprintf("[%s] [%s] %s", timestamp, level, message)
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
