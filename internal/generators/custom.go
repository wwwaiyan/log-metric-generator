package generators

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

type CustomLogFormat string

const (
	FormatJSON       CustomLogFormat = "json"
	FormatCloudWatch CustomLogFormat = "cloudwatch"
	FormatSyslog     CustomLogFormat = "syslog"
)

type CustomLogGenerator struct {
	format     CustomLogFormat
	eventTypes []string
	severities []string
	components []string
}

func NewCustomLogGenerator(format string) *CustomLogGenerator {
	if format == "" {
		format = "json"
	}

	return &CustomLogGenerator{
		format: CustomLogFormat(format),
		eventTypes: []string{
			"USER_LOGIN",
			"USER_LOGOUT",
			"ORDER_CREATED",
			"ORDER_COMPLETED",
			"PAYMENT_PROCESSED",
			"PAYMENT_FAILED",
			"INVENTORY_UPDATED",
			"EMAIL_SENT",
			"CONFIG_CHANGED",
			"CACHE_HIT",
			"CACHE_MISS",
		},
		severities: []string{
			"DEBUG",
			"INFO",
			"INFO",
			"INFO",
			"WARN",
			"ERROR",
		},
		components: []string{
			"api-gateway",
			"auth-service",
			"order-service",
			"payment-service",
			"notification-service",
			"inventory-service",
		},
	}
}

func (g *CustomLogGenerator) Generate() string {
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	eventType := g.eventTypes[rand.Intn(len(g.eventTypes))]
	severity := g.severities[rand.Intn(len(g.severities))]
	component := g.components[rand.Intn(len(g.components))]
	requestID := generateRequestID()

	switch g.format {
	case FormatCloudWatch:
		return g.generateCloudWatchFormat(timestamp, eventType, severity, component, requestID)
	case FormatSyslog:
		return g.generateSyslogFormat(timestamp, eventType, severity, component, requestID)
	default:
		return g.generateJSONFormat(timestamp, eventType, severity, component, requestID)
	}
}

func (g *CustomLogGenerator) generateJSONFormat(timestamp, eventType, severity, component, requestID string) string {
	log := map[string]interface{}{
		"@timestamp": timestamp,
		"level":      severity,
		"event":      eventType,
		"component":  component,
		"request_id": requestID,
		"trace_id":   generateRequestID(),
		"span_id":    generateShortID(),
		"user_id":    rand.Intn(10000),
		"session_id": generateRequestID(),
		"metadata": map[string]interface{}{
			"region":      randomRegion(),
			"environment": "production",
			"version":     "1.2.3",
			"instance_id": fmt.Sprintf("i-%s", generateShortID()),
		},
	}

	if eventType == "ORDER_CREATED" || eventType == "ORDER_COMPLETED" {
		log["order"] = map[string]interface{}{
			"order_id":   fmt.Sprintf("ORD-%08d", rand.Intn(100000)),
			"amount":     float64(rand.Intn(10000)) / 100.0,
			"currency":   "USD",
			"item_count": rand.Intn(10) + 1,
		}
	}

	if eventType == "PAYMENT_PROCESSED" || eventType == "PAYMENT_FAILED" {
		statusValue := "failed"
		if eventType == "PAYMENT_PROCESSED" {
			statusValue = "success"
		}
		log["payment"] = map[string]interface{}{
			"payment_id": fmt.Sprintf("PAY-%s", generateShortID()),
			"method":     randomPaymentMethod(),
			"amount":     float64(rand.Intn(10000)) / 100.0,
			"status":     statusValue,
			"processor":  randomPaymentProcessor(),
		}
	}

	data, _ := json.Marshal(log)
	return string(data)
}

func (g *CustomLogGenerator) generateCloudWatchFormat(timestamp, eventType, severity, component, requestID string) string {
	log := map[string]interface{}{
		"timestamp": timestamp,
		"level":     severity,
		"message":   fmt.Sprintf("[%s] %s: Event %s processed by %s", severity, component, eventType, requestID),
		"logger":    component,
		"thread":    fmt.Sprintf("pool-%d-thread-%d", rand.Intn(10), rand.Intn(20)),
		"context": map[string]interface{}{
			"requestId": requestID,
			"traceId":   generateRequestID(),
			"spanId":    generateShortID(),
		},
	}

	data, _ := json.Marshal(log)
	return string(data)
}

func (g *CustomLogGenerator) generateSyslogFormat(timestamp, eventType, severity, component, requestID string) string {
	priority := calculateSyslogPriority(severity)
	return fmt.Sprintf("<%d>%s %s %s[%d]: [%s] %s - request_id=%s event=%s",
		priority,
		timestamp,
		"localhost",
		component,
		rand.Intn(65000),
		severity,
		eventType,
		requestID,
		eventType,
	)
}

func calculateSyslogPriority(severity string) int {
	sevMap := map[string]int{
		"EMERGENCY": 0,
		"ALERT":     1,
		"CRITICAL":  2,
		"ERROR":     3,
		"WARN":      4,
		"NOTICE":    5,
		"INFO":      6,
		"DEBUG":     7,
	}
	if sev, ok := sevMap[severity]; ok {
		return 16*23 + sev
	}
	return 16*23 + 6
}

func randomRegion() string {
	regions := []string{
		"us-east-1", "us-west-2", "eu-west-1",
		"ap-southeast-1", "ap-northeast-1",
	}
	return regions[rand.Intn(len(regions))]
}

func randomPaymentMethod() string {
	methods := []string{"credit_card", "debit_card", "paypal", "bank_transfer", "crypto"}
	return methods[rand.Intn(len(methods))]
}

func randomPaymentProcessor() string {
	processors := []string{"stripe", "paypal", "square", "adyen", "braintree"}
	return processors[rand.Intn(len(processors))]
}

func generateShortID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 8)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
