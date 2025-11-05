package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"
)

// Workload represents a YAML-defined task
type Workload struct {
	ID          string            `yaml:"id" json:"id"`
	Name        string            `yaml:"name" json:"name"`
	Image       string            `yaml:"image" json:"image"`
	Command     []string          `yaml:"command" json:"command"`
	Environment map[string]string `yaml:"environment" json:"environment"`
	Resources   struct {
		CPU    string `yaml:"cpu" json:"cpu"`
		Memory string `yaml:"memory" json:"memory"`
	} `yaml:"resources" json:"resources"`
	Status string `json:"status"`
}

// AternaClient handles pub/sub communication
type AternaClient struct {
	nodeID   string
	endpoint string
	// Add your aterna-specific fields here
}

// WorkloadManager handles execution and lifecycle
type WorkloadManager struct {
	workloads   map[string]*Workload
	mu          sync.RWMutex
	aterna      *AternaClient
	interactive bool
}

// WebSocketUpgrader for real-time communication
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for localhost
	},
}

func main() {
	var (
		interactive = flag.Bool("interactive", false, "Run in interactive mode with GUI")
		aternaAddr  = flag.String("aterna", "localhost:9999", "Aterna pub/sub address")
		nodeID      = flag.String("node-id", "", "Unique node identifier")
		port        = flag.String("port", "8080", "Web server port for interactive mode")
	)
	flag.Parse()

	if *nodeID == "" {
		*nodeID = fmt.Sprintf("node-%d", time.Now().Unix())
	}

	// Initialize components
	aterna := &AternaClient{
		nodeID:   *nodeID,
		endpoint: *aternaAddr,
	}

	manager := &WorkloadManager{
		workloads:   make(map[string]*Workload),
		aterna:      aterna,
		interactive: *interactive,
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start workload manager
	go manager.Start(ctx)

	if *interactive {
		log.Printf("Starting interactive mode on port %s", *port)
		go startWebServer(*port, manager)
		
		// In production, you'd launch webview here:
		// browser := webview.New(false)
		// browser.SetTitle("Distributed Workload Manager")
		// browser.Navigate(fmt.Sprintf("http://localhost:%s", *port))
		// browser.SetSize(1200, 800, webview.HintNone)
		// defer browser.Destroy()
		// browser.Run()
		
		log.Printf("Web interface available at http://localhost:%s", *port)
	} else {
		log.Printf("Starting headless mode for node %s", *nodeID)
	}

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down...")
	cancel()
}

func (wm *WorkloadManager) Start(ctx context.Context) {
	// Connect to aterna pub/sub
	if err := wm.aterna.Connect(); err != nil {
		log.Fatalf("Failed to connect to aterna: %v", err)
	}

	// Subscribe to workload topics
	go wm.subscribeToWorkloads(ctx)
	
	// Publish node capabilities
	go wm.publishCapabilities(ctx)

	// Health check loop
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			wm.publishHealth()
		}
	}
}

func (wm *WorkloadManager) subscribeToWorkloads(ctx context.Context) {
	// Mock aterna subscription - replace with your actual aterna client
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Simulate receiving workload from aterna
			time.Sleep(5 * time.Second)
			
			// Example workload
			workload := &Workload{
				ID:      fmt.Sprintf("task-%d", time.Now().Unix()),
				Name:    "example-task",
				Image:   "alpine:latest",
				Command: []string{"echo", "Hello from distributed client"},
				Status:  "pending",
			}
			
			log.Printf("Received workload: %s", workload.ID)
			wm.executeWorkload(workload)
		}
	}
}

func (wm *WorkloadManager) executeWorkload(workload *Workload) {
	wm.mu.Lock()
	workload.Status = "running"
	wm.workloads[workload.ID] = workload
	wm.mu.Unlock()

	go func() {
		defer func() {
			wm.mu.Lock()
			workload.Status = "completed"
			wm.mu.Unlock()
		}()

		// Execute the workload (simplified - you'd use proper containerization)
		cmd := exec.Command(workload.Command[0], workload.Command[1:]...)
		
		// Set environment variables
		env := os.Environ()
		for k, v := range workload.Environment {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env

		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Workload %s failed: %v", workload.ID, err)
			workload.Status = "failed"
		} else {
			log.Printf("Workload %s completed: %s", workload.ID, string(output))
		}

		// Report back to aterna
		wm.reportWorkloadStatus(workload)
	}()
}

func (wm *WorkloadManager) publishCapabilities(ctx context.Context) {
	capabilities := map[string]interface{}{
		"node_id": wm.aterna.nodeID,
		"cpu_cores": 4, // Get from runtime
		"memory_gb": 8, // Get from runtime
		"available": true,
	}
	
	// Mock aterna publish - replace with actual implementation
	log.Printf("Publishing capabilities: %+v", capabilities)
}

func (wm *WorkloadManager) publishHealth() {
	wm.mu.RLock()
	runningCount := 0
	for _, w := range wm.workloads {
		if w.Status == "running" {
			runningCount++
		}
	}
	wm.mu.RUnlock()

	health := map[string]interface{}{
		"node_id": wm.aterna.nodeID,
		"timestamp": time.Now(),
		"running_workloads": runningCount,
		"total_workloads": len(wm.workloads),
	}
	
	log.Printf("Health check: %+v", health)
}

func (wm *WorkloadManager) reportWorkloadStatus(workload *Workload) {
	// Report to aterna pub/sub
	log.Printf("Reporting workload %s status: %s", workload.ID, workload.Status)
}

// Aterna client methods (mock implementation)
func (ac *AternaClient) Connect() error {
	log.Printf("Connecting to aterna at %s as node %s", ac.endpoint, ac.nodeID)
	// Your actual aterna connection logic here
	return nil
}

// Web server for interactive mode
func startWebServer(port string, manager *WorkloadManager) {
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/api/workloads", manager.handleWorkloads)
	http.HandleFunc("/api/submit", manager.handleSubmitWorkload)
	http.HandleFunc("/ws", manager.handleWebSocket)
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>Distributed Workload Manager</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .workload { border: 1px solid #ccc; margin: 10px 0; padding: 10px; }
        .running { background-color: #e8f5e8; }
        .completed { background-color: #e8e8f5; }
        .failed { background-color: #f5e8e8; }
        textarea { width: 100%; height: 200px; }
    </style>
</head>
<body>
    <h1>Distributed Workload Manager</h1>
    
    <h2>Submit New Workload</h2>
    <form id="workloadForm">
        <textarea id="yamlInput" placeholder="Enter YAML workload definition...">
id: my-task
name: example-workload
image: alpine:latest
command: ["echo", "Hello World"]
environment:
  ENV_VAR: "test"
resources:
  cpu: "100m"
  memory: "128Mi"
        </textarea>
        <br><br>
        <button type="submit">Submit Workload</button>
    </form>

    <h2>Active Workloads</h2>
    <div id="workloads"></div>

    <script>
        const ws = new WebSocket('ws://localhost:` + port + `/ws');
        
        ws.onmessage = function(event) {
            const workloads = JSON.parse(event.data);
            updateWorkloads(workloads);
        };

        function updateWorkloads(workloads) {
            const container = document.getElementById('workloads');
            container.innerHTML = '';
            
            for (const [id, workload] of Object.entries(workloads)) {
                const div = document.createElement('div');
                div.className = 'workload ' + workload.status;
                div.innerHTML = '<h3>' + workload.name + ' (' + id + ')</h3>' +
                               '<p>Status: ' + workload.status + '</p>' +
                               '<p>Command: ' + workload.command.join(' ') + '</p>';
                container.appendChild(div);
            }
        }

        document.getElementById('workloadForm').onsubmit = function(e) {
            e.preventDefault();
            const yaml = document.getElementById('yamlInput').value;
            
            fetch('/api/submit', {
                method: 'POST',
                headers: {'Content-Type': 'text/plain'},
                body: yaml
            }).then(response => {
                if (response.ok) {
                    alert('Workload submitted!');
                } else {
                    alert('Failed to submit workload');
                }
            });
        };
        
        // Request initial workload list
        fetch('/api/workloads').then(r => r.json()).then(updateWorkloads);
    </script>
</body>
</html>`
	
	t, _ := template.New("index").Parse(tmpl)
	t.Execute(w, nil)
}

func (wm *WorkloadManager) handleWorkloads(w http.ResponseWriter, r *http.Request) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wm.workloads)
}

func (wm *WorkloadManager) handleSubmitWorkload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var workload Workload
	if err := yaml.NewDecoder(r.Body).Decode(&workload); err != nil {
		http.Error(w, "Invalid YAML: "+err.Error(), http.StatusBadRequest)
		return
	}

	workload.Status = "pending"
	log.Printf("Received workload submission: %s", workload.ID)
	
	wm.executeWorkload(&workload)
	w.WriteHeader(http.StatusOK)
}

func (wm *WorkloadManager) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Send real-time updates
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wm.mu.RLock()
			if err := conn.WriteJSON(wm.workloads); err != nil {
				wm.mu.RUnlock()
				return
			}
			wm.mu.RUnlock()
		}
	}
}
