package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/yourusername/ai-agent-team/internal/llmfactory"
	"github.com/yourusername/ai-agent-team/internal/models"
	"github.com/yourusername/ai-agent-team/internal/orchestrator"
)

// sessionState tracks per-discussion agent state for the War Room UI
type sessionState struct {
	Discussion *models.Discussion
	Agents     map[string]*webAgentState
	Phase      string
	PhaseIcon  string
	Log        []string
}

type webAgentState struct {
	Role    string `json:"role"`
	Name    string `json:"name"`
	Icon    string `json:"icon"`
	Tagline string `json:"tagline"`
	Color   string `json:"color"`
	Status  string `json:"status"` // idle, thinking, done
	Speech  string `json:"speech"`
}

var agentPersonas = map[string]struct {
	Name, Icon, Tagline, Color string
}{
	"team_leader": {"Commander Bleep", "🤖", "beep-boop, let's rally!", "#FF6B6B"},
	"ideation":    {"Sparx", "💡", "zapping up wild ideas", "#51E898"},
	"moderator":   {"Balancebot", "🔮", "keeping the circuits aligned", "#7B68EE"},
	"researcher":  {"Digger-3000", "🔍", "scanning all known databases", "#00D4FF"},
	"critic":      {"Glitchy", "👾", "poking the logic boards", "#FFD93D"},
	"implementer": {"Bolt", "🔧", "tightening the bolts", "#FF8C42"},
	"ui_creator":  {"Doodlebot", "🎨", "painting pixels with love", "#FF6BC1"},
}

var (
	sessions = make(map[string]*sessionState)
	mu       sync.RWMutex
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
	fmt.Println("║   🤖 IdeaArmy — The Idea Factory Server                ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Printf("\n🌐 Server starting on http://localhost:%s\n", port)
	fmt.Println()

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func newAgentState(role string) *webAgentState {
	p, ok := agentPersonas[role]
	if !ok {
		return &webAgentState{Role: role, Name: role, Icon: "🤖", Tagline: "booting up", Color: "#8892A0", Status: "idle"}
	}
	return &webAgentState{Role: role, Name: p.Name, Icon: p.Icon, Tagline: p.Tagline, Color: p.Color, Status: "idle"}
}

// parseProgress extracts agent speech and phase changes from orchestrator messages
func parseProgress(ss *sessionState, msg string) {
	trimmed := strings.TrimSpace(msg)

	// Detect speech: "📣 [role] content"
	if strings.Contains(trimmed, "📣") {
		if start := strings.Index(trimmed, "["); start >= 0 {
			if end := strings.Index(trimmed[start:], "]"); end > 0 {
				role := trimmed[start+1 : start+end]
				speech := strings.TrimSpace(trimmed[start+end+1:])
				if a, ok := ss.Agents[role]; ok {
					a.Speech = speech
					a.Status = "done"
				}
			}
		}
	}

	// Detect agent starting work: "🗣️ <name> contributing..."
	if strings.Contains(trimmed, "contributing...") {
		// Set all to idle first, mark the active one
		for _, a := range ss.Agents {
			if a.Status == "thinking" {
				a.Status = "idle"
			}
		}
		for role, a := range ss.Agents {
			if strings.Contains(trimmed, a.Name) || strings.Contains(strings.ToLower(trimmed), role) {
				a.Status = "thinking"
			}
		}
	}

	// Detect phase changes
	if strings.Contains(trimmed, "Phase 1:") || strings.Contains(trimmed, "Kickoff") {
		ss.Phase = "Team Leader Kickoff"
		ss.PhaseIcon = "📋"
		if a, ok := ss.Agents["team_leader"]; ok {
			a.Status = "thinking"
		}
	} else if strings.Contains(trimmed, "Exploration Round") {
		ss.Phase = trimmed
		ss.PhaseIcon = "💡"
	} else if strings.Contains(trimmed, "Final Validation") {
		ss.Phase = "Final Validation"
		ss.PhaseIcon = "🔍"
		if a, ok := ss.Agents["moderator"]; ok {
			a.Status = "thinking"
		}
	} else if strings.Contains(trimmed, "Final Selection") {
		ss.Phase = "Final Selection"
		ss.PhaseIcon = "🎯"
		if a, ok := ss.Agents["team_leader"]; ok {
			a.Status = "thinking"
		}
	} else if strings.Contains(trimmed, "Visual Idea Sheet") {
		ss.Phase = "Creating Idea Sheet"
		ss.PhaseIcon = "🎨"
		if a, ok := ss.Agents["ui_creator"]; ok {
			a.Status = "thinking"
		}
	} else if strings.Contains(trimmed, "completed successfully") {
		ss.Phase = "Idea Factory Complete!"
		ss.PhaseIcon = "🙌"
		for _, a := range ss.Agents {
			if a.Status == "thinking" {
				a.Status = "done"
			}
		}
	}

	// Detect synthesizing
	if strings.Contains(trimmed, "synthesizing") {
		if a, ok := ss.Agents["team_leader"]; ok {
			a.Status = "thinking"
		}
	}

	ss.Log = append(ss.Log, trimmed)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>🤖 The Idea Factory — IdeaArmy</title>
    <link href="https://fonts.googleapis.com/css2?family=Nunito:wght@400;600;700;800&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-dark: #1A1B2E;
            --bg-card: #232440;
            --bg-desk: #2D2E4A;
            --border: #7B68EE44;
            --text: #e2e8f0;
            --text-dim: #8892A0;
            --accent: #FF6BC1;
            --accent-glow: rgba(255, 107, 193, 0.3);
            --neon-mint: #51E898;
            --electric-cyan: #00D4FF;
            --bright-yellow: #FFD93D;
            --coral: #FF6B6B;
            --tangerine: #FF8C42;
            --slate-purple: #7B68EE;
        }
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Nunito', 'SF Mono', sans-serif;
            background: var(--bg-dark);
            color: var(--text);
            min-height: 100vh;
            overflow-x: hidden;
        }

        /* ── Header ── */
        .war-room-header {
            text-align: center;
            padding: 24px 20px 16px;
            background: linear-gradient(135deg, rgba(255,107,193,0.15) 0%, rgba(81,232,152,0.1) 50%, rgba(0,212,255,0.15) 100%);
            border-bottom: 1px solid var(--border);
        }
        .war-room-header h1 {
            font-size: 2rem;
            font-weight: 800;
            background: linear-gradient(90deg, var(--accent), var(--electric-cyan), var(--neon-mint));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            letter-spacing: 2px;
        }
        .war-room-header .subtitle { color: var(--text-dim); font-size: 0.85rem; margin-top: 4px; }

        /* ── Setup Form ── */
        .setup-panel {
            max-width: 700px;
            margin: 40px auto;
            padding: 30px;
            background: var(--bg-card);
            border-radius: 12px;
            border: 1px solid var(--border);
        }
        .setup-panel h2 { color: var(--accent); margin-bottom: 20px; text-align: center; }
        .field { margin-bottom: 18px; }
        .field label { display: block; color: var(--text-dim); font-size: 0.8rem; margin-bottom: 6px; text-transform: uppercase; letter-spacing: 1px; }
        .field input, .field textarea, .field select {
            width: 100%; padding: 10px 14px; background: var(--bg-desk); border: 1px solid var(--border);
            border-radius: 8px; color: var(--text); font-family: inherit; font-size: 0.9rem;
        }
        .field input:focus, .field textarea:focus, .field select:focus { outline: none; border-color: var(--accent); box-shadow: 0 0 0 3px var(--accent-glow); }
        .field input, .field textarea, .field select {
            background: var(--bg-desk); border: 1px solid var(--border);
        }
        .field textarea { resize: vertical; min-height: 80px; }
        .field small { color: var(--text-dim); font-size: 0.75rem; display: block; margin-top: 4px; }

        .team-options { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }
        .team-opt {
            padding: 12px; border-radius: 8px; cursor: pointer; text-align: center;
            background: var(--bg-desk); border: 2px solid transparent; transition: all 0.2s;
        }
        .team-opt:hover { border-color: var(--border); }
        .team-opt.selected { border-color: var(--accent); background: rgba(255,107,193,0.1); }
        .team-opt .t-icon { font-size: 1.5rem; display: block; margin-bottom: 4px; }
        .team-opt .t-name { font-weight: 700; font-size: 0.85rem; }
        .team-opt .t-desc { color: var(--text-dim); font-size: 0.7rem; margin-top: 2px; }

        .btn-launch {
            display: block; width: 100%; padding: 14px;
            background: linear-gradient(135deg, var(--accent), var(--slate-purple));
            color: white; border: none; border-radius: 8px;
            font-family: inherit; font-size: 1rem; font-weight: 800;
            cursor: pointer; letter-spacing: 1px; transition: all 0.2s; margin-top: 10px;
            animation: pinkGlow 2s ease-in-out infinite;
        }
        .btn-launch:hover { transform: translateY(-2px); box-shadow: 0 4px 25px var(--accent-glow); }
        @keyframes pinkGlow {
            0%, 100% { box-shadow: 0 0 10px var(--accent-glow); }
            50% { box-shadow: 0 0 25px var(--accent-glow), 0 0 50px rgba(255,107,193,0.1); }
        }

        /* ── War Room Layout ── */
        .war-room { display: none; }
        .war-room.active { display: block; }

        .phase-banner {
            text-align: center; padding: 12px;
            background: var(--bg-card); border-bottom: 1px solid var(--border);
        }
        .phase-banner .phase-text { font-size: 1.1rem; color: var(--bright-yellow); font-weight: 700; }

        .room-grid {
            display: grid;
            grid-template-columns: 1fr 320px;
            gap: 0;
            min-height: calc(100vh - 140px);
        }

        /* ── Agent Desks ── */
        .desks-area {
            padding: 20px;
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
            gap: 16px;
            align-content: start;
        }

        .desk {
            background: var(--bg-card);
            border-radius: 12px;
            border: 2px solid var(--border);
            padding: 16px;
            position: relative;
            transition: all 0.4s;
        }
        .desk.thinking {
            border-color: var(--accent);
            box-shadow: 0 0 20px var(--accent-glow);
            animation: glow 2s ease-in-out infinite;
        }
        .desk.done { border-color: var(--neon-mint); }

        @keyframes glow {
            0%, 100% { box-shadow: 0 0 15px var(--accent-glow); }
            50% { box-shadow: 0 0 30px var(--accent-glow), 0 0 60px rgba(255,107,193,0.1); }
        }
        .desk.thinking .avatar { animation: botBounce 0.6s ease-in-out infinite; }
        @keyframes botBounce {
            0%, 100% { transform: translateY(0); }
            50% { transform: translateY(-4px); }
        }

        .desk-header { display: flex; align-items: center; gap: 10px; margin-bottom: 10px; }
        .avatar {
            width: 48px; height: 48px; border-radius: 50%;
            display: flex; align-items: center; justify-content: center;
            font-size: 1.5rem; flex-shrink: 0;
            position: relative;
        }
        .avatar::after {
            content: ''; position: absolute; bottom: 2px; right: 2px;
            width: 10px; height: 10px; border-radius: 50%;
            border: 2px solid var(--bg-card);
        }
        .desk.idle .avatar::after { background: var(--text-dim); }
        .desk.thinking .avatar::after { background: var(--bright-yellow); animation: blink 1s infinite; }
        .desk.done .avatar::after { background: var(--neon-mint); }

        @keyframes blink { 0%, 100% { opacity: 1; } 50% { opacity: 0.3; } }

        .agent-info { flex: 1; min-width: 0; }
        .agent-name { font-weight: 700; font-size: 0.95rem; }
        .agent-role { color: var(--text-dim); font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.5px; }

        .status-badge {
            font-size: 0.65rem; padding: 2px 8px; border-radius: 10px;
            text-transform: uppercase; font-weight: 700; letter-spacing: 0.5px;
        }
        .status-idle { background: var(--bg-desk); color: var(--text-dim); }
        .status-thinking { background: rgba(255,217,61,0.15); color: var(--bright-yellow); }
        .status-done { background: rgba(81,232,152,0.15); color: var(--neon-mint); }

        .speech-bubble {
            background: var(--bg-desk);
            border-radius: 8px;
            padding: 10px 12px;
            font-size: 0.8rem;
            line-height: 1.5;
            color: var(--text);
            position: relative;
            min-height: 40px;
            max-height: 120px;
            overflow-y: auto;
        }
        .speech-bubble::before {
            content: ''; position: absolute; top: -6px; left: 20px;
            border-left: 6px solid transparent; border-right: 6px solid transparent;
            border-bottom: 6px solid var(--bg-desk);
        }
        .speech-bubble.empty { color: var(--text-dim); font-style: italic; }

        .typing-indicator span {
            display: inline-block; width: 6px; height: 6px; border-radius: 50%;
            background: var(--text-dim); margin: 0 2px;
            animation: typing 1.4s infinite both;
        }
        .typing-indicator span:nth-child(2) { animation-delay: 0.2s; }
        .typing-indicator span:nth-child(3) { animation-delay: 0.4s; }
        @keyframes typing { 0%, 100% { opacity: 0.3; } 50% { opacity: 1; } }

        /* ── Sidebar ── */
        .sidebar {
            background: var(--bg-card);
            border-left: 1px solid var(--border);
            display: flex; flex-direction: column;
            overflow: hidden;
        }
        .sidebar-section { padding: 16px; border-bottom: 1px solid var(--border); }
        .sidebar-section h3 { color: var(--electric-cyan); font-size: 0.8rem; text-transform: uppercase; letter-spacing: 1px; margin-bottom: 10px; }

        .idea-card {
            background: var(--bg-desk); border-radius: 8px; padding: 10px 12px; margin-bottom: 8px;
            border-left: 3px solid var(--neon-mint);
        }
        .idea-card .idea-title { font-weight: 700; font-size: 0.85rem; margin-bottom: 2px; }
        .idea-card .idea-score { color: var(--bright-yellow); font-size: 0.75rem; }
        .idea-card.winner { border-left-color: var(--bright-yellow); background: rgba(255,217,61,0.05); }

        .activity-feed {
            flex: 1; overflow-y: auto; padding: 12px 16px;
            font-size: 0.75rem; line-height: 1.7; color: var(--text-dim);
        }
        .activity-feed .feed-entry { margin-bottom: 3px; word-break: break-word; }

        /* ── Result Overlay ── */
        .result-overlay { display: none; position: fixed; inset: 0; background: rgba(0,0,0,0.8); z-index: 100; padding: 30px; overflow-y: auto; }
        .result-overlay.active { display: flex; align-items: start; justify-content: center; }
        .result-card {
            background: white; border-radius: 12px; width: 100%; max-width: 1100px;
            overflow: hidden; margin-top: 20px; position: relative;
        }
        .result-card .result-header {
            background: linear-gradient(135deg, var(--accent), var(--slate-purple)); color: white;
            padding: 16px 24px; display: flex; justify-content: space-between; align-items: center;
        }
        .result-card iframe { width: 100%; height: 70vh; border: none; }
        .btn-close {
            background: rgba(255,255,255,0.2); border: none; color: white;
            padding: 6px 16px; border-radius: 6px; cursor: pointer; font-family: inherit; font-size: 0.85rem;
        }
        .btn-close:hover { background: rgba(255,255,255,0.3); }
        .btn-new {
            display: inline-block; padding: 8px 20px; margin-top: 12px;
            background: var(--accent); color: white; border: none; border-radius: 6px;
            cursor: pointer; font-family: inherit; font-size: 0.85rem;
        }
        .btn-new:hover { background: var(--slate-purple); }

        /* ── Sparkle celebration ── */
        @keyframes sparkle {
            0% { opacity: 1; transform: scale(1) translateY(0); }
            100% { opacity: 0; transform: scale(1.5) translateY(-40px); }
        }
        .sparkle-emoji {
            position: fixed; font-size: 1.5rem; pointer-events: none; z-index: 200;
            animation: sparkle 1.2s ease-out forwards;
        }

        /* ── Responsive ── */
        @media (max-width: 900px) {
            .room-grid { grid-template-columns: 1fr; }
            .sidebar { border-left: none; border-top: 1px solid var(--border); max-height: 300px; }
            .team-options { grid-template-columns: 1fr; }
        }
    </style>
</head>
<body>
    <div class="war-room-header">
        <h1>🤖 THE IDEA FACTORY</h1>
        <div class="subtitle">A Playful Robot Army for Collaborative Brainstorming</div>
    </div>

    <!-- ── Setup Form ── -->
    <div id="setup" class="setup-panel">
        <h2>🤖 Bot Deployment Orders</h2>

        <div class="field">
            <label>Your LLM Token <span style="color:#ef4444">*</span></label>
            <input type="password" id="apiKey" placeholder="user=yourname&amp;key=sk_..." required>
            <small>Your personal LLM proxy key (format: user=name&amp;key=sk_...) or OpenAI API key. Each user must provide their own token.</small>
        </div>

        <div class="field">
            <label>Squad Configuration</label>
            <div class="team-options">
                <div class="team-opt selected" onclick="selectTeam(this,'standard')">
                    <span class="t-icon">⚡</span>
                    <div class="t-name">Starter Pack</div>
                    <div class="t-desc">4 bots · 1 round</div>
                </div>
                <div class="team-opt" onclick="selectTeam(this,'extended')">
                    <span class="t-icon">🔍</span>
                    <div class="t-name">Explorer Squad</div>
                    <div class="t-desc">6 bots · 2 rounds</div>
                </div>
                <div class="team-opt" onclick="selectTeam(this,'full')">
                    <span class="t-icon">🤖</span>
                    <div class="t-name">Full Robot Army</div>
                    <div class="t-desc">7 bots · 3 rounds</div>
                </div>
            </div>
            <input type="hidden" id="teamConfig" value="standard">
        </div>

        <div class="field">
            <label>What should the bots brainstorm?</label>
            <textarea id="topic" placeholder="Describe the topic you want the bots to explore..."></textarea>
        </div>

        <button class="btn-launch" onclick="launchMission()">🤖 DEPLOY THE BOTS!</button>
    </div>

    <!-- ── War Room ── -->
    <div id="warRoom" class="war-room">
        <div class="phase-banner">
            <span class="phase-text" id="phaseText">⚡ Powering up the bots...</span>
        </div>
        <div class="room-grid">
            <div class="desks-area" id="desksArea"></div>
            <div class="sidebar">
                <div class="sidebar-section">
                    <h3>💡 Idea Conveyor Belt</h3>
                    <div id="ideasBoard"><div style="color:var(--text-dim);font-size:0.8rem;font-style:italic">Warming up idea generators...</div></div>
                </div>
                <div class="sidebar-section" style="flex-shrink:0">
                    <h3>📡 Bot Chatter</h3>
                </div>
                <div class="activity-feed" id="activityFeed"></div>
            </div>
        </div>
    </div>

    <!-- ── Result Overlay ── -->
    <div id="resultOverlay" class="result-overlay">
        <div class="result-card">
            <div class="result-header">
                <span>✨ See What They Built!</span>
                <button class="btn-close" onclick="closeResult()">✕ Close</button>
            </div>
            <iframe id="resultFrame"></iframe>
        </div>
    </div>

    <script>
        let discussionId = null;
        let pollTimer = null;
        let selectedTeam = 'standard';
        let seenLogCount = 0;
        let agentRoles = [];

        const TEAM_AGENTS = {
            standard: ['team_leader','ideation','moderator','ui_creator'],
            extended: ['team_leader','ideation','moderator','researcher','critic','ui_creator'],
            full:     ['team_leader','ideation','moderator','researcher','critic','implementer','ui_creator']
        };

        function selectTeam(el, team) {
            selectedTeam = team;
            document.getElementById('teamConfig').value = team;
            document.querySelectorAll('.team-opt').forEach(e => e.classList.remove('selected'));
            el.classList.add('selected');
        }

        function buildDesks(roles) {
            agentRoles = roles;
            const area = document.getElementById('desksArea');
            area.innerHTML = '';
            roles.forEach(role => {
                // We'll get real persona data from the first status poll
                const div = document.createElement('div');
                div.className = 'desk idle';
                div.id = 'desk-' + role;
                div.innerHTML = ` + "`" + `
                    <div class="desk-header">
                        <div class="avatar" id="avatar-${role}" style="background:var(--bg-desk)">🤖</div>
                        <div class="agent-info">
                            <div class="agent-name" id="name-${role}">${role}</div>
                            <div class="agent-role">${role.replace('_',' ')}</div>
                        </div>
                        <span class="status-badge status-idle" id="badge-${role}">snoozing</span>
                    </div>
                    <div class="speech-bubble empty" id="speech-${role}">Charging batteries...</div>
                ` + "`" + `;
                area.appendChild(div);
            });
        }

        function updateDesk(agent) {
            const desk = document.getElementById('desk-' + agent.role);
            if (!desk) return;

            // Update persona info
            document.getElementById('avatar-' + agent.role).textContent = agent.icon;
            document.getElementById('avatar-' + agent.role).style.background = agent.color + '22';
            document.getElementById('name-' + agent.role).textContent = agent.name;

            // Update status
            desk.className = 'desk ' + agent.status;
            const badge = document.getElementById('badge-' + agent.role);
            badge.className = 'status-badge status-' + agent.status;
            badge.textContent = agent.status === 'thinking' ? 'buzzing' : agent.status === 'done' ? 'high-five!' : 'snoozing';

            // Update speech
            const bubble = document.getElementById('speech-' + agent.role);
            if (agent.status === 'thinking' && !agent.speech) {
                bubble.className = 'speech-bubble';
                bubble.innerHTML = '<span class="typing-indicator"><span></span><span></span><span></span></span>';
            } else if (agent.speech) {
                bubble.className = 'speech-bubble';
                bubble.textContent = agent.speech;
            }
        }

        async function launchMission() {
            const apiKey = document.getElementById('apiKey').value.trim();
            const topic = document.getElementById('topic').value.trim();

            if (!apiKey) { alert('Please provide your LLM token'); return; }
            if (!topic) { alert('Please provide a mission objective'); return; }

            document.getElementById('setup').style.display = 'none';
            document.getElementById('warRoom').classList.add('active');

            buildDesks(TEAM_AGENTS[selectedTeam]);

            try {
                const res = await fetch('/api/start', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({ api_key: apiKey, topic: topic, team_config: selectedTeam })
                });
                const data = await res.json();
                if (data.error) throw new Error(data.error);

                discussionId = data.discussion_id;
                pollTimer = setInterval(pollStatus, 1500);
            } catch(e) {
                alert('Launch failed: ' + e.message);
                resetToSetup();
            }
        }

        async function pollStatus() {
            if (!discussionId) return;
            try {
                const res = await fetch('/api/status/' + discussionId);
                const data = await res.json();

                // Update phase
                if (data.phase) {
                    document.getElementById('phaseText').textContent = (data.phase_icon || '⏳') + ' ' + data.phase;
                }

                // Update agent desks
                if (data.agents) {
                    Object.values(data.agents).forEach(updateDesk);
                }

                // Update ideas board
                if (data.ideas && data.ideas.length > 0) {
                    const board = document.getElementById('ideasBoard');
                    board.innerHTML = '';
                    data.ideas.forEach((idea, i) => {
                        const card = document.createElement('div');
                        card.className = 'idea-card' + (data.final_idea === idea.title ? ' winner' : '');
                        card.innerHTML = (data.final_idea === idea.title ? '⭐ ' : '💡 ')
                            + '<span class="idea-title">' + escHtml(idea.title) + '</span>'
                            + (idea.score > 0 ? ' <span class="idea-score">' + idea.score.toFixed(1) + '/10</span>' : '');
                        board.appendChild(card);
                    });
                }

                // Update activity feed
                if (data.log && data.log.length > seenLogCount) {
                    const feed = document.getElementById('activityFeed');
                    for (let i = seenLogCount; i < data.log.length; i++) {
                        const div = document.createElement('div');
                        div.className = 'feed-entry';
                        div.textContent = data.log[i];
                        feed.appendChild(div);
                    }
                    seenLogCount = data.log.length;
                    feed.scrollTop = feed.scrollHeight;
                }

                // Check completion
                if (data.status === 'completed') {
                    clearInterval(pollTimer);
                    document.getElementById('phaseText').innerHTML =
                        '🙌 Bots nailed it! <button class="btn-new" onclick="showResult()">✨ See What They Built!</button> <button class="btn-new" onclick="resetToSetup()">🤖 Deploy Again!</button>';
                    celebrateSparkles();
                    showResult();
                } else if (data.status === 'failed') {
                    clearInterval(pollTimer);
                    document.getElementById('phaseText').textContent = '❌ Bots hit a glitch — ' + (data.error || 'Unknown error');
                }
            } catch(e) {
                console.error('Poll error:', e);
            }
        }

        async function showResult() {
            try {
                const res = await fetch('/api/result/' + discussionId);
                const data = await res.json();
                if (data.html) {
                    document.getElementById('resultFrame').srcdoc = data.html;
                    document.getElementById('resultOverlay').classList.add('active');
                }
            } catch(e) { console.error(e); }
        }

        function closeResult() {
            document.getElementById('resultOverlay').classList.remove('active');
        }

        function resetToSetup() {
            clearInterval(pollTimer);
            discussionId = null;
            seenLogCount = 0;
            document.getElementById('warRoom').classList.remove('active');
            document.getElementById('resultOverlay').classList.remove('active');
            document.getElementById('setup').style.display = 'block';
        }

        function celebrateSparkles() {
            const emojis = ['✨','🤖','💡','🎉','⭐','🔧','🙌','💫'];
            for (let i = 0; i < 20; i++) {
                setTimeout(() => {
                    const span = document.createElement('span');
                    span.className = 'sparkle-emoji';
                    span.textContent = emojis[Math.floor(Math.random() * emojis.length)];
                    span.style.left = Math.random() * 100 + 'vw';
                    span.style.top = (50 + Math.random() * 40) + 'vh';
                    document.body.appendChild(span);
                    setTimeout(() => span.remove(), 1500);
                }, i * 100);
            }
        }

        function escHtml(s) { const d = document.createElement('div'); d.textContent = s; return d.innerHTML; }
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

	if req.Topic == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Topic is required"})
		return
	}

	if req.APIKey == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "LLM token is required — each user must provide their own"})
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

	client, err := llmfactory.NewClientAuto(req.APIKey)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Failed to create LLM client: %v", err)})
		return
	}

	orch := orchestrator.NewConfigurableOrchestrator(client, config)

	// Build agent states for this team
	agentStates := make(map[string]*webAgentState)
	for _, role := range config.GetActiveAgentRoles() {
		agentStates[string(role)] = newAgentState(string(role))
	}

	// Create session state
	ss := &sessionState{
		Agents:    agentStates,
		Phase:     "Powering up the bots...",
		PhaseIcon: "⚡",
	}

	// Wire up progress callback to parse agent updates
	orch.OnProgress = func(message string) {
		log.Println(message)
		mu.Lock()
		parseProgress(ss, message)
		mu.Unlock()
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

	time.Sleep(500 * time.Millisecond)

	discussion := orch.GetDiscussion()
	if discussion == nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create discussion"})
		return
	}

	ss.Discussion = discussion

	mu.Lock()
	sessions[discussion.ID] = ss
	mu.Unlock()

	respondJSON(w, http.StatusOK, map[string]string{
		"discussion_id": discussion.ID,
	})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/status/"):]

	mu.RLock()
	ss, exists := sessions[id]
	mu.RUnlock()

	if !exists {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "Discussion not found"})
		return
	}

	// Build ideas list
	type ideaInfo struct {
		Title string  `json:"title"`
		Score float64 `json:"score"`
	}
	ideas := make([]ideaInfo, 0)
	for _, idea := range ss.Discussion.Ideas {
		ideas = append(ideas, ideaInfo{Title: idea.Title, Score: idea.Score})
	}

	// Final idea title
	finalIdea := ""
	if ss.Discussion.FinalIdea != nil {
		finalIdea = ss.Discussion.FinalIdea.Title
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":     ss.Discussion.Status,
		"phase":      ss.Phase,
		"phase_icon": ss.PhaseIcon,
		"agents":     ss.Agents,
		"ideas":      ideas,
		"final_idea": finalIdea,
		"round":      ss.Discussion.Round,
		"log":        ss.Log,
	})
}

func handleResult(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/result/"):]

	mu.RLock()
	ss, exists := sessions[id]
	mu.RUnlock()

	if !exists {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "Discussion not found"})
		return
	}

	var html string
	for _, msg := range ss.Discussion.Messages {
		if msg.Type == "visualization" {
			html = msg.Content
			break
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"html": html,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
