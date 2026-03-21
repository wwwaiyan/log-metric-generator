package generators

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type MetricData struct {
	Timestamp  time.Time
	Namespace  string
	Dimensions map[string]string
	Metrics    []MetricPoint
}

type MetricPoint struct {
	Name      string
	Value     float64
	Unit      string
	Timestamp time.Time
}

type MetricsGenerator struct {
	namespace    string
	cpuBase      float64
	memoryBase   float64
	requestCount int
	latencyP50   int
	latencyP99   int
}

func NewMetricsGenerator(namespace string, cpuBase, memoryBase float64, requestCount, latencyP50, latencyP99 int) *MetricsGenerator {
	if namespace == "" {
		namespace = "TestApp/Metrics"
	}
	if cpuBase == 0 {
		cpuBase = 45.0
	}
	if memoryBase == 0 {
		memoryBase = 65.0
	}
	if requestCount == 0 {
		requestCount = 1000
	}
	if latencyP50 == 0 {
		latencyP50 = 120
	}
	if latencyP99 == 0 {
		latencyP99 = 500
	}

	return &MetricsGenerator{
		namespace:    namespace,
		cpuBase:      cpuBase,
		memoryBase:   memoryBase,
		requestCount: requestCount,
		latencyP50:   latencyP50,
		latencyP99:   latencyP99,
	}
}

func (g *MetricsGenerator) Generate(serviceName, instanceID string) MetricData {
	timestamp := time.Now().UTC()

	dimensions := map[string]string{
		"ServiceName": serviceName,
		"InstanceId":  instanceID,
	}

	metrics := []MetricPoint{
		g.generateCPU(timestamp),
		g.generateMemory(timestamp),
		g.generateRequestCount(timestamp),
		g.generateLatency(timestamp),
		g.generateErrorRate(timestamp),
		g.generateNetwork(timestamp),
		g.generateDisk(timestamp),
	}

	return MetricData{
		Timestamp:  timestamp,
		Namespace:  g.namespace,
		Dimensions: dimensions,
		Metrics:    metrics,
	}
}

func (g *MetricsGenerator) generateCPU(timestamp time.Time) MetricPoint {
	variance := (rand.Float64() - 0.5) * 20
	value := math.Max(5, math.Min(95, g.cpuBase+variance))

	spike := rand.Float64()
	if spike > 0.95 {
		value = math.Min(98, value+30)
	}

	return MetricPoint{
		Name:      "CPUUtilization",
		Value:     math.Round(value*100) / 100,
		Unit:      "Percent",
		Timestamp: timestamp,
	}
}

func (g *MetricsGenerator) generateMemory(timestamp time.Time) MetricPoint {
	variance := (rand.Float64() - 0.5) * 10
	value := math.Max(20, math.Min(90, g.memoryBase+variance))

	gradual := rand.Float64()
	if gradual > 0.9 {
		value = math.Min(95, value+15)
	}

	return MetricPoint{
		Name:      "MemoryUtilization",
		Value:     math.Round(value*100) / 100,
		Unit:      "Percent",
		Timestamp: timestamp,
	}
}

func (g *MetricsGenerator) generateRequestCount(timestamp time.Time) MetricPoint {
	variance := float64(g.requestCount) * (rand.Float64() - 0.5) * 0.5
	baseRequests := float64(g.requestCount) + variance

	burst := rand.Float64()
	if burst > 0.9 {
		baseRequests *= 2.5
	}

	return MetricPoint{
		Name:      "RequestCount",
		Value:     math.Max(0, math.Round(baseRequests)),
		Unit:      "Count",
		Timestamp: timestamp,
	}
}

func (g *MetricsGenerator) generateLatency(timestamp time.Time) MetricPoint {
	p50Variance := float64(g.latencyP50) * (rand.Float64() - 0.5) * 0.4
	p99Variance := float64(g.latencyP99) * (rand.Float64() - 0.5) * 0.4

	p50 := float64(g.latencyP50) + p50Variance
	p99 := float64(g.latencyP99) + p99Variance

	if p99 < p50 {
		p99 = p50 * 4
	}

	latency := (p50 + p99) / 2

	spike := rand.Float64()
	if spike > 0.95 {
		latency *= 3
	}

	return MetricPoint{
		Name:      "Latency",
		Value:     math.Max(10, math.Round(latency*100)/100),
		Unit:      "Milliseconds",
		Timestamp: timestamp,
	}
}

func (g *MetricsGenerator) generateErrorRate(timestamp time.Time) MetricPoint {
	baseErrorRate := 2.0
	variance := rand.Float64() * 3

	spike := rand.Float64()
	if spike > 0.95 {
		variance *= 5
	}

	errorRate := baseErrorRate + variance

	return MetricPoint{
		Name:      "ErrorRate",
		Value:     math.Max(0, math.Round(errorRate*100)/100),
		Unit:      "Percent",
		Timestamp: timestamp,
	}
}

func (g *MetricsGenerator) generateNetwork(timestamp time.Time) MetricPoint {
	bytesIn := float64(rand.Intn(1000000) + 100000)
	bytesOut := float64(rand.Intn(2000000) + 200000)

	totalBytes := bytesIn + bytesOut

	return MetricPoint{
		Name:      "NetworkBytes",
		Value:     math.Round(totalBytes),
		Unit:      "Bytes",
		Timestamp: timestamp,
	}
}

func (g *MetricsGenerator) generateDisk(timestamp time.Time) MetricPoint {
	readOps := float64(rand.Intn(5000) + 500)
	writeOps := float64(rand.Intn(3000) + 300)

	totalOps := readOps + writeOps

	return MetricPoint{
		Name:      "DiskIOPS",
		Value:     math.Round(totalOps),
		Unit:      "Count",
		Timestamp: timestamp,
	}
}

func (g *MetricsGenerator) ToEMF(metrics MetricData) string {
	emf := map[string]interface{}{
		"_aws": map[string]interface{}{
			"Timestamp": metrics.Timestamp.UnixMilli(),
			"CloudWatchMetrics": []map[string]interface{}{
				{
					"Namespace":  metrics.Namespace,
					"Dimensions": [][]string{flattenDimensions(metrics.Dimensions)},
					"Metrics":    buildMetricsArray(metrics.Metrics),
				},
			},
		},
	}

	for _, dim := range flattenDimensionsMap(metrics.Dimensions) {
		emf[dim.Key] = dim.Value
	}

	for _, m := range metrics.Metrics {
		emf[m.Name] = m.Value
	}

	return fmt.Sprintf("%+v", emf)
}

func flattenDimensions(dims map[string]string) []string {
	result := make([]string, 0, len(dims))
	for k := range dims {
		result = append(result, k)
	}
	return result
}

type dimPair struct {
	Key   string
	Value string
}

func flattenDimensionsMap(dims map[string]string) []dimPair {
	result := make([]dimPair, 0, len(dims))
	for k, v := range dims {
		result = append(result, dimPair{Key: k, Value: v})
	}
	return result
}

func buildMetricsArray(metrics []MetricPoint) []map[string]interface{} {
	result := make([]map[string]interface{}, len(metrics))
	for i, m := range metrics {
		result[i] = map[string]interface{}{
			"Name": m.Name,
			"Unit": m.Unit,
		}
	}
	return result
}
