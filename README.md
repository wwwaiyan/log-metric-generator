# CloudWatch Log/Metric Simulator

A lightweight Go application for simulating AWS CloudWatch logs and metrics. Designed for testing CloudWatch agent configurations, ECS deployments, and observability pipelines.

## Features

- **Realistic Log Generation**
  - Apache-style web server access logs
  - Error logs with configurable error rates
  - Custom JSON/CloudWatch format logs

- **Metric Simulation**
  - CPU/Memory utilization
  - Request counts and latency
  - Error rates and network metrics
  - CloudWatch EMF (Embedded Metric Format) compatible

- **Multiple Output Modes**
  - `stdout`: JSON format compatible with `awslogs` driver
  - `cloudwatch`: Direct CloudWatch Logs API integration

- **Container-Ready**
  - Multi-stage Alpine build (~15MB)
  - Scratch-based minimal image (~10MB)
  - AWS ECS Fargate compatible

## Quick Start

### Docker

```bash
# Build the image
docker build -t log-metric-generator .

# Run with stdout output (for awslogs driver)
docker run -e OUTPUT_MODE=stdout \
           -e SIMULATOR_LOG_GROUP=/ecs/simulator \
           log-metric-generator

# Run with CloudWatch direct output
docker run -e OUTPUT_MODE=cloudwatch \
           -e AWS_REGION=us-east-1 \
           -e AWS_ACCESS_KEY_ID=xxx \
           -e AWS_SECRET_ACCESS_KEY=xxx \
           -e SIMULATOR_LOG_GROUP=/ecs/simulator \
           log-metric-generator
```

### Docker Compose

```bash
# Local testing with stdout output
docker-compose up simulator

# All services
docker-compose up

# With AWS credentials for CloudWatch
AWS_ACCESS_KEY_ID=xxx AWS_SECRET_ACCESS_KEY=xxx docker-compose up simulator-cloudwatch
```

### AWS ECS

1. Push image to ECR:
```bash
aws ecr create-repository --repository-name log-metric-generator
docker build -t log-metric-generator .
docker tag log-metric-generator:latest YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/log-metric-generator:latest
docker push YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/log-metric-generator:latest
```

2. Register task definition:
```bash
aws ecs register-task-definition --cli-input-json file://ecs-task-definition.json
```

3. Run task:
```bash
aws ecs start-task --cluster YOUR_CLUSTER --task-definition log-metric-simulator --launch-type FARGATE --network-configuration "awsvpcConfiguration={subnets=[YOUR_SUBNET],securityGroups=[YOUR_SG]}"
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OUTPUT_MODE` | Output mode: `stdout` or `cloudwatch` | `stdout` |
| `SIMULATOR_LOG_GROUP` | CloudWatch Log Group name | `/ecs/simulator` |
| `SIMULATOR_LOG_STREAM` | CloudWatch Log Stream name | `app-{instance_id}` |
| `SIMULATOR_INSTANCE_ID` | Unique instance identifier | auto-generated |
| `GENERATOR_WEB_ENABLED` | Enable web server logs | `true` |
| `GENERATOR_WEB_RPS` | Requests per second | `50` |
| `GENERATOR_ERROR_ENABLED` | Enable error logs | `true` |
| `GENERATOR_ERROR_RATE` | Error rate (0.0-1.0) | `0.05` |
| `GENERATOR_CUSTOM_ENABLED` | Enable custom logs | `true` |
| `METRICS_NAMESPACE` | CloudWatch Metrics namespace | `TestApp/Metrics` |
| `METRICS_INTERVAL` | Metrics interval (seconds) | `60` |
| `AWS_REGION` | AWS region | `us-east-1` |

### config.yaml

See `config.yaml` for full configuration options. Environment variables override config file values.

## Architecture

```
┌─────────────────────────────────────────┐
│              Simulator                  │
├─────────────────────────────────────────┤
│  Log Generators     │  Metric Generator │
│  ├─ Web Server      │  ├─ CPU/Memory   │
│  ├─ Error Logs      │  ├─ Latency      │
│  └─ Custom JSON     │  └─ Request Count│
├─────────────────────────────────────────┤
│         Output Writers                 │
│  ├─ stdout (awslogs driver)            │
│  └─ CloudWatch API                    │
└─────────────────────────────────────────┘
```

## Log Formats

### Web Server Logs (Apache Format)
```
10.0.1.50 - - [22/Mar/2026:10:30:45 +0000] "GET /api/users/1234 HTTP/1.1" 200 1234 45.123 "Mozilla/5.0..." "request_id=abc123def456"
```

### Error Logs
```
[2026-03-22T10:30:45Z] ERROR: TIMEOUT | service=order-service request_id=xyz789
```

### CloudWatch EMF Metrics
```json
{"_aws":{"Timestamp":1679482245000,"CloudWatchMetrics":[{"Namespace":"TestApp/Metrics","Dimensions":[["ServiceName","InstanceId"]],"Metrics":[{"Name":"CPUUtilization","Unit":"Percent"}]}]},"CPUUtilization":45.23,"ServiceName":"app-service"}
```

## License

MIT
