package generators

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type ErrorType string

const (
	ErrorTypeTimeout            ErrorType = "TIMEOUT"
	ErrorTypeConnection         ErrorType = "CONNECTION_REFUSED"
	ErrorTypeInternal           ErrorType = "INTERNAL_ERROR"
	ErrorTypeDatabase           ErrorType = "DATABASE_ERROR"
	ErrorTypeAuth               ErrorType = "AUTH_FAILED"
	ErrorTypeRateLimit          ErrorType = "RATE_LIMIT_EXCEEDED"
	ErrorTypeValidation         ErrorType = "VALIDATION_ERROR"
	ErrorTypeNotFound           ErrorType = "NOT_FOUND"
	ErrorTypeServiceUnavailable ErrorType = "SERVICE_UNAVAILABLE"
)

type ErrorLog struct {
	Timestamp  string
	Level      string
	ErrorType  string
	Message    string
	RequestID  string
	StackTrace string
	Service    string
}

type ErrorGenerator struct {
	errorTypes []ErrorType
	services   []string
}

func NewErrorGenerator() *ErrorGenerator {
	return &ErrorGenerator{
		errorTypes: []ErrorType{
			ErrorTypeTimeout,
			ErrorTypeConnection,
			ErrorTypeInternal,
			ErrorTypeDatabase,
			ErrorTypeAuth,
			ErrorTypeRateLimit,
			ErrorTypeValidation,
			ErrorTypeNotFound,
			ErrorTypeServiceUnavailable,
		},
		services: []string{
			"user-service",
			"order-service",
			"payment-service",
			"inventory-service",
			"notification-service",
		},
	}
}

func (g *ErrorGenerator) Generate(errorType ErrorType) ErrorLog {
	requestID := generateRequestID()
	service := g.services[rand.Intn(len(g.services))]

	if errorType == "" {
		errorType = g.errorTypes[rand.Intn(len(g.errorTypes))]
	}

	level := "ERROR"
	if errorType == ErrorTypeTimeout || errorType == ErrorTypeRateLimit {
		level = "WARN"
	}

	return ErrorLog{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Level:      level,
		ErrorType:  string(errorType),
		Message:    g.generateMessage(errorType),
		RequestID:  requestID,
		StackTrace: g.generateStackTrace(service),
		Service:    service,
	}
}

func (g *ErrorGenerator) generateMessage(errorType ErrorType) string {
	switch errorType {
	case ErrorTypeTimeout:
		return fmt.Sprintf("Request to downstream service timed out after %ds", rand.Intn(10)+5)
	case ErrorTypeConnection:
		return "Connection refused: remote server unavailable"
	case ErrorTypeInternal:
		return "Internal server error: unexpected condition encountered"
	case ErrorTypeDatabase:
		return fmt.Sprintf("Database query failed: connection pool exhausted (pool_size=%d)", rand.Intn(50)+10)
	case ErrorTypeAuth:
		return "Authentication failed: invalid credentials or expired token"
	case ErrorTypeRateLimit:
		return fmt.Sprintf("Rate limit exceeded: %d requests per minute allowed", rand.Intn(100)+100)
	case ErrorTypeValidation:
		return "Request validation failed: required field 'email' is missing"
	case ErrorTypeNotFound:
		return "Resource not found: /api/v1/users/999999"
	case ErrorTypeServiceUnavailable:
		return "Service temporarily unavailable: maintenance in progress"
	default:
		return "Unknown error occurred"
	}
}

func (g *ErrorGenerator) generateStackTrace(service string) string {
	frames := []string{
		fmt.Sprintf("    at %s.handleRequest (/%s/handler.js:%d)", service, service, rand.Intn(500)+100),
		fmt.Sprintf("    at %s.process (/%s/middleware.js:%d)", service, service, rand.Intn(300)+50),
		fmt.Sprintf("    at async %s.execute (/%s/router.js:%d)", service, service, rand.Intn(200)+20),
		fmt.Sprintf("    at module.exports (/%s/index.js:%d)", service, service, rand.Intn(100)+1),
	}
	return strings.Join(frames, "\n")
}

func (log ErrorLog) ToJSON() string {
	return fmt.Sprintf(`{"timestamp":"%s","level":"%s","error_type":"%s","message":"%s","request_id":"%s","service":"%s","stack_trace":"%s"}`,
		log.Timestamp,
		log.Level,
		log.ErrorType,
		strings.ReplaceAll(log.Message, `"`, `\\"`),
		log.RequestID,
		log.Service,
		strings.ReplaceAll(log.StackTrace, `"`, `\\"`),
	)
}

func (log ErrorLog) ToPlainText() string {
	return fmt.Sprintf("[%s] %s: %s | service=%s request_id=%s\n%s",
		log.Timestamp,
		log.Level,
		log.ErrorType,
		log.Message,
		log.Service,
		log.RequestID,
		log.StackTrace,
	)
}
