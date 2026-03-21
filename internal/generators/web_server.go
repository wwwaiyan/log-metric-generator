package generators

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type WebServerLog struct {
	Timestamp  string
	RemoteAddr string
	Method     string
	Path       string
	StatusCode int
	BytesSent  int
	LatencyMs  float64
	UserAgent  string
	RequestID  string
}

type WebServerGenerator struct {
	paths       []string
	methods     []string
	statusCodes []int
	userAgents  []string
}

func NewWebServerGenerator(paths []string) *WebServerGenerator {
	if len(paths) == 0 {
		paths = []string{"/api/users", "/api/orders", "/health"}
	}

	return &WebServerGenerator{
		paths:   paths,
		methods: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		statusCodes: []int{
			200, 200, 200, 200, 200,
			201, 204,
			400, 401, 403, 404,
			500, 502, 503,
		},
		userAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
			"aws-cli/2.13.0 Python/3.11.5 exec-env/AWS_ECS Fargate",
			"APN/1.0 CloudWatchAgent/1.300037.1",
		},
	}
}

func (g *WebServerGenerator) Generate(latencyMs float64, statusCode int) WebServerLog {
	ip := generateRandomIP()
	method := g.methods[rand.Intn(len(g.methods))]
	path := g.paths[rand.Intn(len(g.paths))]
	userAgent := g.userAgents[rand.Intn(len(g.userAgents))]
	requestID := generateRequestID()

	if path == "/api/users" || path == "/api/orders" {
		id := rand.Intn(10000)
		path = fmt.Sprintf("%s/%d", path, id)
	}

	return WebServerLog{
		Timestamp:  time.Now().UTC().Format("02/Jan/2006:15:04:05 -0700"),
		RemoteAddr: ip,
		Method:     method,
		Path:       path,
		StatusCode: statusCode,
		BytesSent:  rand.Intn(5000) + 100,
		LatencyMs:  latencyMs,
		UserAgent:  userAgent,
		RequestID:  requestID,
	}
}

func (log WebServerLog) ToApacheFormat() string {
	return fmt.Sprintf(`%s - - [%s] "%s %s HTTP/1.1" %d %d %0.3f "%s" "%s" request_id=%s`,
		log.RemoteAddr,
		log.Timestamp,
		log.Method,
		log.Path,
		log.StatusCode,
		log.BytesSent,
		log.LatencyMs,
		"-",
		log.UserAgent,
		log.RequestID,
	)
}

func (log WebServerLog) ToJSON() string {
	return fmt.Sprintf(`{"timestamp":"%s","remote_addr":"%s","method":"%s","path":"%s","status":%d,"bytes":%d,"latency_ms":%.3f,"user_agent":"%s","request_id":"%s"}`,
		log.Timestamp,
		log.RemoteAddr,
		log.Method,
		log.Path,
		log.StatusCode,
		log.BytesSent,
		log.LatencyMs,
		strings.ReplaceAll(log.UserAgent, `"`, `\\"`),
		log.RequestID,
	)
}

func generateRandomIP() string {
	octets := []int{
		rand.Intn(223) + 1,
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
	}
	if octets[0] == 10 || (octets[0] == 172 && octets[1] >= 16 && octets[1] <= 31) || (octets[0] == 192 && octets[1] == 168) {
		return fmt.Sprintf("10.%d.%d.%d", octets[1], octets[2], octets[3])
	}
	return fmt.Sprintf("%d.%d.%d.%d", octets[0], octets[1], octets[2], octets[3])
}

func generateRequestID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 16)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
