# ATP2PC
# Distributed Workload Client

A lightweight, zero-trust distributed compute client for executing YAML-defined workloads across P2P networks. Built in Go with dual-mode operation: headless for pure compute nodes and interactive for workload management.

## 🎯 Overview

This client transforms any machine into a distributed compute node that can:
- Execute containerized workloads defined in YAML
- Operate in zero-trust environments with node identity verification
- Provide rich interactive dashboards for workload management
- Scale horizontally through aterna pub/sub messaging
- Report real-time metrics and health status

## 🏗️ Architecture

```
┌─────────────────────┐
│   Aterna Pub/Sub    │ ← Message Bus & Discovery
│                     │
└─────────┬───────────┘
          │
    ┌─────▼─────┐
    │  Client   │
    │           │
    ├───────────┤
    │ Headless  │ ← Pure compute execution
    │   Mode    │   • Zero UI overhead
    │           │   • Maximum throughput
    ├───────────┤
    │Interactive│ ← Management interface
    │   Mode    │   • Web dashboard
    │           │   • Real-time monitoring
    └───────────┘
```

## 🚀 Features

### Core Capabilities
- **YAML Workload Execution**: Parse and execute complex workload definitions
- **Real-time Status Reporting**: Health metrics and execution status via aterna
- **Zero-trust Security**: Node identity verification and capability announcement
- **Resource Management**: CPU/memory allocation and monitoring
- **Graceful Shutdown**: Clean workload termination and state preservation

### Dual Operating Modes

**🖥️ Headless Mode**
- Pure compute execution with no GUI overhead
- Ideal for edge devices, servers, and embedded systems
- Subscribes to workload topics via aterna pub/sub
- Reports execution results and health metrics
- Maximum resource efficiency for compute-intensive tasks

**🌐 Interactive Mode**
- Web-based dashboard on configurable port (default: 8080)
- Real-time workload submission and monitoring
- Live WebSocket updates for execution status
- Visual workload management and debugging
- Perfect for development and management nodes

## 🛠️ Installation & Setup

### Prerequisites
- Go 1.21 or later
- Access to aterna pub/sub infrastructure
- Docker/containerd (optional, for container execution)

### Quick Start

1. **Clone and build:**
```bash
git clone <repository>
cd distributed-client
go mod tidy
```

2. **Run headless compute node:**
```bash
go run main.go -node-id=worker-001 -aterna=your-aterna-endpoint:9999
```

3. **Run interactive management node:**
```bash
go run main.go -interactive -port=8080 -node-id=manager-001
```

4. **Access web interface:**
```
http://localhost:8080
```

## 📋 Usage

### Command Line Options

```bash
Usage: go run main.go [options]

Options:
  -interactive          Run in interactive mode with web GUI (default: false)
  -node-id string      Unique node identifier (auto-generated if empty)
  -aterna string       Aterna pub/sub address (default: "localhost:9999")
  -port string         Web server port for interactive mode (default: "8080")
```

### Examples

**Production Compute Farm:**
```bash
# Deploy multiple headless workers
go run main.go -node-id=gpu-worker-01 -aterna=prod-aterna:9999
go run main.go -node-id=cpu-worker-02 -aterna=prod-aterna:9999
go run main.go -node-id=edge-device-03 -aterna=prod-aterna:9999
```

**Development Environment:**
```bash
# Interactive node for testing
go run main.go -interactive -port=3000 -node-id=dev-manager
```

**Edge Computing:**
```bash
# Lightweight edge node
go run main.go -node-id=edge-$(hostname) -aterna=central-hub:9999
```

## 📄 YAML Workload Format

The client executes workloads defined in YAML format:

```yaml
# Basic workload structure
id: unique-workload-identifier
name: human-readable-name
image: container-image-name
command: ["executable", "arg1", "arg2"]
environment:
  ENV_VAR: "value"
  WORKER_ID: "node-001"
resources:
  cpu: "100m"      # CPU request (Kubernetes format)
  memory: "128Mi"  # Memory request
```

### Complete Example

```yaml
id: data-processing-task
name: log-analysis-worker
image: alpine:latest
command: 
  - "sh"
  - "-c"
  - |
    echo "Processing logs for $(date)"
    # Simulate data processing
    for i in $(seq 1 1000); do
      echo "Processing record $i" >> /tmp/results.log
    done
    echo "Analysis complete: $(wc -l < /tmp/results.log) records processed"
environment:
  TASK_TYPE: "analytics"
  LOG_LEVEL: "INFO"
  BATCH_SIZE: "1000"
resources:
  cpu: "200m"
  memory: "256Mi"
```

## 🌐 Web Interface (Interactive Mode)

When running in interactive mode, the client provides a rich web dashboard:

### Dashboard Features
- **Real-time Workload Status**: Live updates on execution progress
- **YAML Editor**: Submit new workloads with syntax highlighting
- **Execution History**: View completed, failed, and running workloads
- **Node Health**: CPU, memory usage, and connectivity status
- **WebSocket Updates**: Real-time data without page refreshes

### API Endpoints

```http
# Workload Management
GET  /api/workloads      # List all workloads
POST /api/submit         # Submit new workload (YAML body)

# Real-time Updates
GET  /ws                 # WebSocket connection for live updates

# Node Status
GET  /                   # Dashboard UI
```

## 🔗 Aterna Integration

The client integrates with your aterna pub/sub system for distributed coordination:

### Integration Points

**Client-side Methods** (implement with your aterna client):

```go
// Connection Management
func (ac *AternaClient) Connect() error {
    // Establish connection to aterna cluster
    // Handle authentication and encryption
}

// Workload Subscription
func (wm *WorkloadManager) subscribeToWorkloads(ctx context.Context) {
    // Subscribe to: workloads.{node-id}
    // Subscribe to: workloads.broadcast
    // Parse incoming YAML and execute
}

// Status Publishing
func (wm *WorkloadManager) publishCapabilities(ctx context.Context) {
    // Publish to: nodes.capabilities
    // Include: CPU cores, memory, availability
}

func (wm *WorkloadManager) reportWorkloadStatus(workload *Workload) {
    // Publish to: workloads.status.{workload-id}
    // Include: status, node-id, timestamps
}
```

### Message Topics

| Topic | Direction | Purpose |
|-------|-----------|---------|
| `workloads.{node-id}` | Subscribe | Receive targeted workloads |
| `workloads.broadcast` | Subscribe | Receive cluster-wide workloads |
| `nodes.capabilities` | Publish | Announce node capabilities |
| `nodes.health.{node-id}` | Publish | Health check reports |
| `workloads.status.{id}` | Publish | Execution status updates |

## 🔧 Configuration

### Environment Variables

```bash
export ATERNA_ENDPOINT="prod-cluster:9999"
export NODE_ID="worker-$(hostname)"
export MAX_CONCURRENT_WORKLOADS="5"
export HEALTH_CHECK_INTERVAL="30s"
```

### Runtime Configuration

The client automatically configures based on system resources:

```go
// Auto-detected capabilities
capabilities := map[string]interface{}{
    "cpu_cores":   runtime.NumCPU(),
    "memory_gb":   getTotalMemoryGB(),
    "platform":    runtime.GOOS + "/" + runtime.GOARCH,
    "available":   true,
}
```

## 📊 Monitoring & Metrics

### Health Metrics

The client reports comprehensive health data:

```json
{
  "node_id": "worker-001",
  "timestamp": "2025-11-05T10:30:00Z",
  "running_workloads": 3,
  "total_workloads": 47,
  "cpu_usage": 65.2,
  "memory_usage": 78.5,
  "uptime_seconds": 86400
}
```

### Workload Status

Each workload reports detailed execution information:

```json
{
  "id": "data-processing-task",
  "status": "running",
  "assigned_node": "worker-001",
  "started_at": "2025-11-05T10:25:00Z",
  "progress": 65,
  "output_preview": "Processing record 650...",
  "resource_usage": {
    "cpu_percent": 45.2,
    "memory_mb": 156
  }
}
```

## 🛡️ Security & Zero Trust

### Node Identity
- Each client generates or is assigned a unique node ID
- Capabilities are cryptographically signed (when aterna crypto is enabled)
- Health reports include node authentication tokens

### Workload Security
- YAML workloads can include signature verification
- Environment variables support encrypted values
- Container isolation prevents workload interference

### Network Security
- All aterna communication uses TLS encryption
- Node discovery requires valid certificates
- Workload dispatch includes anti-replay tokens

## 🚨 Troubleshooting

### Common Issues

**Connection Problems:**
```bash
# Check aterna connectivity
go run main.go -aterna=your-endpoint:9999 -node-id=test-connection

# Expected output:
# Connecting to aterna at your-endpoint:9999 as node test-connection
# Scheduler operational
```

**Workload Execution Failures:**
```bash
# Check logs for execution details
tail -f /var/log/workload-client.log

# Common issues:
# - Missing container runtime
# - Insufficient resources
# - Network connectivity
```

**Interactive Mode Issues:**
```bash
# Verify web server is running
curl http://localhost:8080/api/workloads

# Check WebSocket connection
# Browser dev tools -> Network -> WS tab
```

### Debug Mode

Enable verbose logging:

```bash
go run main.go -interactive -node-id=debug-node -log-level=debug
```

## 📚 Advanced Usage

### Custom Execution Environments

Replace the default execution engine:

```go
// Custom workload executor
func (wm *WorkloadManager) executeWorkload(workload *Workload) {
    // Option 1: Docker execution
    cmd := exec.Command("docker", "run", "--rm", workload.Image, workload.Command...)
    
    // Option 2: Kubernetes Job
    createKubernetesJob(workload)
    
    // Option 3: Custom runtime
    executeInSandbox(workload)
}
```

### Multi-tenant Workloads

```yaml
id: tenant-workload
name: customer-analytics
image: analytics:latest
command: ["python", "analyze.py"]
environment:
  TENANT_ID: "customer-123"
  DATA_SOURCE: "s3://customer-data/"
resources:
  cpu: "1000m"
  memory: "2Gi"
```

### GPU Workloads

```yaml
id: ml-training-task
name: neural-network-training
image: tensorflow/tensorflow:latest-gpu
command: ["python", "train_model.py"]
environment:
  CUDA_VISIBLE_DEVICES: "0,1"
  MODEL_PATH: "/workspace/models"
resources:
  cpu: "4000m"
  memory: "8Gi"
  gpu: "2"
```

## 🔮 Roadmap

- **Container Runtime Integration**: Native Docker/containerd support
- **Kubernetes CRD**: Deploy as Kubernetes DaemonSet
- **WebAssembly Support**: Execute WASM workloads for ultra-fast startup
- **ML Acceleration**: Built-in support for TensorFlow/PyTorch workloads
- **Edge AI**: Optimized execution for edge AI inference
- **Auto-scaling**: Dynamic resource adjustment based on workload demands

## 🤝 Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

-** No Support for now. Use at your own peril.

---

**Built for the distributed future. Execute workloads anywhere, anytime. 🚀**
