package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/yourusername/ai-agent-team/internal/models"
	"github.com/yourusername/ai-agent-team/internal/orchestrator"
)

var (
	discussions = make(map[string]*models.Discussion)
	mu          sync.RWMutex
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/api/start", handleStart)
	http.HandleFunc("/api/status/", handleStatus)
	http.HandleFunc("/api/result/", handleResult)

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║   AI Agent Team v2 - Web Server                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Printf("\n🌐 Server starting on http://localhost:%s\n", port)
	fmt.Println()

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI Agent Team v2 - Collaborative Ideation</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container { max-width: 900px; margin: 0 auto; }
        .header { text-align: center; color: white; margin-bottom: 40px; }
        .header h1 { font-size: 2.5rem; margin-bottom: 10px; }
        .header p { font-size: 1.1rem; opacity: 0.9; }
        .card {
            background: white;
            border-radius: 12px;
            padding: 30px;
            box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
            margin-bottom: 20px;
        }
        .input-group { margin-bottom: 20px; }
        .input-group label { display: block; margin-bottom: 8px; font-weight: 600; color: #333; }
        .input-group input, .input-group textarea, .input-group select {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 1rem;
            transition: border-color 0.3s;
        }
        .input-group input:focus, .input-group textarea:focus, .input-group select:focus {
            outline: none;
            border-color: #667eea;
        }
        .input-group textarea { resize: vertical; min-height: 100px; }
        .btn {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 14px 32px;
            font-size: 1.1rem;
            border-radius: 8px;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
            width: 100%;
        }
        .btn:hover { transform: translateY(-2px); box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4); }
        .btn:disabled { opacity: 0.6; cursor: not-allowed; transform: none; }
        .progress { display: none; }
        .progress.active { display: block; }
        .progress-bar { background: #e0e0e0; border-radius: 8px; height: 8px; overflow: hidden; margin-bottom: 15px; }
        .progress-fill {
            background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
            height: 100%;
            width: 0%;
            transition: width 0.3s;
            animation: pulse 2s infinite;
        }
        @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.7; } }
        .progress-text { color: #666; font-size: 0.9rem; margin-bottom: 10px; }
        .log {
            background: #f8f9fa;
            border-radius: 8px;
            padding: 15px;
            max-height: 400px;
            overflow-y: auto;
            font-family: 'Courier New', monospace;
            font-size: 0.85rem;
            line-height: 1.6;
        }
        .log-entry { margin-bottom: 5px; }
        .result { display: none; }
        .result.active { display: block; }
        .result iframe { width: 100%; height: 600px; border: none; border-radius: 8px; }
        .team-option {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 10px;
            cursor: pointer;
            border: 2px solid transparent;
            transition: all 0.3s;
        }
        .team-option:hover { background: #e9ecef; }
        .team-option.selected { border-color: #667eea; background: #e7e9fc; }
        .team-option h4 { margin-bottom: 5px; color: #333; }
        .team-option p { color: #666; font-size: 0.9rem; margin-bottom: 5px; }
        .team-option .agents { color: #667eea; font-size: 0.85rem; }
        small { color: #666; display: block; margin-top: 5px; font-size: 0.85rem; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🤖 AI Agent Team v2</h1>
            <p>Configurable Multi-Agent Collaborative Ideation</p>
        </div>

        <div class="card">
            <div id="inputForm">
                <div class="input-group">
                    <label for="apiKey">Anthropic API Key</label>
                    <input type="password" id="apiKey" placeholder="sk-ant-...">
                    <small>Your API key is only used for this session and never stored</small>
                </div>

                <div class="input-group">
                    <label>Team Configuration</label>
                    <div class="team-option selected" onclick="selectTeam('standard')">
                        <h4>⚡ Standard (4 agents, 1 round)</h4>
                        <p>Quick, focused ideation</p>
                        <div class="agents">Leader • Ideation • Moderator • UI Creator</div>
                    </div>
                    <div class="team-option" onclick="selectTeam('extended')">
                        <h4>🔬 Extended (6 agents, 2 rounds)</h4>
                        <p>Deeper analysis with research and critical thinking</p>
                        <div class="agents">+ Researcher • Critic | Multi-round refinement</div>
                    </div>
                    <div class="team-option" onclick="selectTeam('full')">
                        <h4>🚀 Full (7 agents, 3 rounds)</h4>
                        <p>Maximum depth with implementation planning</p>
                        <div class="agents">+ Implementer | Extensive exploration</div>
                    </div>
                    <input type="hidden" id="teamConfig" value="standard">
                </div>

                <div class="input-group">
                    <label for="topic">Discussion Topic</label>
                    <textarea id="topic" placeholder="Enter the topic you want the AI team to explore and develop ideas around..."></textarea>
                </div>

                <button class="btn" onclick="startDiscussion()">
                    🚀 Start AI Team Discussion
                </button>
            </div>

            <div id="progress" class="progress">
                <div class="progress-bar">
                    <div id="progressFill" class="progress-fill"></div>
                </div>
                <div id="progressText" class="progress-text">Initializing...</div>
                <div class="log" id="log"></div>
            </div>
        </div>

        <div id="result" class="result">
            <div class="card">
                <h2 style="margin-bottom: 20px;">✨ Idea Sheet</h2>
                <iframe id="resultFrame"></iframe>
            </div>
        </div>
    </div>

    <script>
        let discussionId = null;
        let checkInterval = null;
        let selectedTeam = 'standard';

        function selectTeam(team) {
            selectedTeam = team;
            document.getElementById('teamConfig').value = team;

            document.querySelectorAll('.team-option').forEach(el => {
                el.classList.remove('selected');
            });
            event.currentTarget.classList.add('selected');
        }

        async function startDiscussion() {
            const apiKey = document.getElementById('apiKey').value.trim();
            const topic = document.getElementById('topic').value.trim();
            const teamConfig = document.getElementById('teamConfig').value;

            if (!apiKey || !topic) {
                alert('Please provide both API key and topic');
                return;
            }

            document.getElementById('inputForm').style.display = 'none';
            document.getElementById('progress').classList.add('active');
            document.getElementById('result').classList.remove('active');
            document.getElementById('log').innerHTML = '';

            try {
                const response = await fetch('/api/start', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        api_key: apiKey,
                        topic: topic,
                        team_config: teamConfig
                    })
                });

                const data = await response.json();

                if (data.error) {
                    throw new Error(data.error);
                }

                discussionId = data.discussion_id;
                updateProgress(20, 'Discussion started...');
                checkInterval = setInterval(checkStatus, 2000);

            } catch (error) {
                alert('Error: ' + error.message);
                resetForm();
            }
        }

        async function checkStatus() {
            if (!discussionId) return;

            try {
                const response = await fetch('/api/status/' + discussionId);
                const data = await response.json();

                if (data.status === 'completed') {
                    clearInterval(checkInterval);
                    updateProgress(100, 'Completed! Loading results...');
                    await loadResult();
                } else if (data.status === 'failed') {
                    clearInterval(checkInterval);
                    alert('Discussion failed: ' + (data.error || 'Unknown error'));
                    resetForm();
                } else {
                    const progress = Math.min(90, 20 + (data.round || 0) * 30 + (data.messages || 0) * 3);
                    updateProgress(progress, 'Round ' + (data.round || 1) + ' - AI agents working...');

                    if (data.latest_message) {
                        addLogEntry(data.latest_message);
                    }
                }
            } catch (error) {
                console.error('Status check error:', error);
            }
        }

        async function loadResult() {
            try {
                const response = await fetch('/api/result/' + discussionId);
                const data = await response.json();

                if (data.html) {
                    document.getElementById('resultFrame').srcdoc = data.html;
                    document.getElementById('progress').classList.remove('active');
                    document.getElementById('result').classList.add('active');
                }
            } catch (error) {
                alert('Error loading result: ' + error.message);
                resetForm();
            }
        }

        function updateProgress(percent, text) {
            document.getElementById('progressFill').style.width = percent + '%';
            document.getElementById('progressText').textContent = text;
        }

        function addLogEntry(message) {
            const log = document.getElementById('log');
            const entry = document.createElement('div');
            entry.className = 'log-entry';
            entry.textContent = message;
            log.appendChild(entry);
            log.scrollTop = log.scrollHeight;
        }

        function resetForm() {
            document.getElementById('inputForm').style.display = 'block';
            document.getElementById('progress').classList.remove('active');
            document.getElementById('result').classList.remove('active');
            discussionId = null;
            if (checkInterval) {
                clearInterval(checkInterval);
                checkInterval = null;
            }
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		APIKey     string `json:"api_key"`
		Topic      string `json:"topic"`
		TeamConfig string `json:"team_config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	if req.APIKey == "" || req.Topic == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "API key and topic are required"})
		return
	}

	// Get team configuration
	var config *models.TeamConfig
	switch req.TeamConfig {
	case "extended":
		config = models.ExtendedTeamConfig()
	case "full":
		config = models.FullTeamConfig()
	default:
		config = models.StandardTeamConfig()
	}

	// Create orchestrator
	orch := orchestrator.NewConfigurableOrchestrator(req.APIKey, config)

	// Set up progress logging
	orch.OnProgress = func(message string) {
		log.Println(message)
	}

	// Start discussion in background
	go func() {
		if err := orch.StartDiscussion(req.Topic); err != nil {
			log.Printf("Discussion failed: %v", err)
			mu.Lock()
			if orch.GetDiscussion() != nil {
				orch.GetDiscussion().Status = "failed"
			}
			mu.Unlock()
		}
	}()

	// Wait a moment for discussion to initialize
	time.Sleep(500 * time.Millisecond)

	discussion := orch.GetDiscussion()
	if discussion == nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create discussion"})
		return
	}

	mu.Lock()
	discussions[discussion.ID] = discussion
	mu.Unlock()

	respondJSON(w, http.StatusOK, map[string]string{
		"discussion_id": discussion.ID,
	})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/status/"):]

	mu.RLock()
	discussion, exists := discussions[id]
	mu.RUnlock()

	if !exists {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "Discussion not found"})
		return
	}

	// Get latest message
	var latestMsg string
	if len(discussion.Messages) > 0 {
		latestMsg = fmt.Sprintf("[%s] %s",
			discussion.Messages[len(discussion.Messages)-1].From,
			discussion.Messages[len(discussion.Messages)-1].Content[:min(100, len(discussion.Messages[len(discussion.Messages)-1].Content))])
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":         discussion.Status,
		"messages":       len(discussion.Messages),
		"ideas":          len(discussion.Ideas),
		"round":          discussion.Round,
		"latest_message": latestMsg,
	})
}

func handleResult(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/result/"):]

	mu.RLock()
	discussion, exists := discussions[id]
	mu.RUnlock()

	if !exists {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "Discussion not found"})
		return
	}

	var html string
	for _, msg := range discussion.Messages {
		if msg.Type == "visualization" {
			html = msg.Content
			break
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"discussion": discussion,
		"html":       html,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
