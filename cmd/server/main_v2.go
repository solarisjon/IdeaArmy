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
	"team_leader": {"Captain Rex", "🎖️", "rallying the troops", "#FFD700"},
	"ideation":    {"Sparky", "⚡", "igniting ideas", "#10B981"},
	"moderator":   {"The Judge", "⚖️", "keeping order", "#3B82F6"},
	"researcher":  {"Doc Sage", "📖", "digging deep", "#8B5CF6"},
	"critic":      {"Nitpick", "🧐", "poking holes", "#F59E0B"},
	"implementer": {"Wrench", "🔩", "making it real", "#06B6D4"},
	"ui_creator":  {"Pixel", "🎨", "painting the vision", "#EC4899"},
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
	fmt.Println("║   AI Agent Team v2 - War Room Server                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Printf("\n🌐 Server starting on http://localhost:%s\n", port)
	fmt.Println()

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func newAgentState(role string) *webAgentState {
	p, ok := agentPersonas[role]
	if !ok {
		return &webAgentState{Role: role, Name: role, Icon: "🤖", Tagline: "working", Color: "#6B7280", Status: "idle"}
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
		ss.Phase = "Mission Complete!"
		ss.PhaseIcon = "✅"
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
    <title>⚔️ The War Room — AI Agent Team</title>
    <style>
        :root {
            --bg-dark: #0f172a;
            --bg-card: #1e293b;
            --bg-desk: #334155;
            --border: #475569;
            --text: #e2e8f0;
            --text-dim: #94a3b8;
            --accent: #7c3aed;
            --accent-glow: rgba(124, 58, 237, 0.3);
            --gold: #ffd700;
            --green: #10b981;
            --blue: #3b82f6;
            --pink: #ec4899;
        }
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'SF Mono', 'Fira Code', 'Cascadia Code', 'Consolas', monospace;
            background: var(--bg-dark);
            color: var(--text);
            min-height: 100vh;
            overflow-x: hidden;
        }

        /* ── Header ── */
        .war-room-header {
            text-align: center;
            padding: 24px 20px 16px;
            background: linear-gradient(180deg, rgba(124,58,237,0.15) 0%, transparent 100%);
            border-bottom: 1px solid var(--border);
        }
        .war-room-header h1 {
            font-size: 2rem;
            background: linear-gradient(90deg, var(--gold), #fbbf24, var(--gold));
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
        .setup-panel h2 { color: var(--gold); margin-bottom: 20px; text-align: center; }
        .field { margin-bottom: 18px; }
        .field label { display: block; color: var(--text-dim); font-size: 0.8rem; margin-bottom: 6px; text-transform: uppercase; letter-spacing: 1px; }
        .field input, .field textarea, .field select {
            width: 100%; padding: 10px 14px; background: var(--bg-desk); border: 1px solid var(--border);
            border-radius: 8px; color: var(--text); font-family: inherit; font-size: 0.9rem;
        }
        .field input:focus, .field textarea:focus, .field select:focus { outline: none; border-color: var(--accent); box-shadow: 0 0 0 3px var(--accent-glow); }
        .field textarea { resize: vertical; min-height: 80px; }
        .field small { color: var(--text-dim); font-size: 0.75rem; display: block; margin-top: 4px; }

        .team-options { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }
        .team-opt {
            padding: 12px; border-radius: 8px; cursor: pointer; text-align: center;
            background: var(--bg-desk); border: 2px solid transparent; transition: all 0.2s;
        }
        .team-opt:hover { border-color: var(--border); }
        .team-opt.selected { border-color: var(--accent); background: rgba(124,58,237,0.1); }
        .team-opt .t-icon { font-size: 1.5rem; display: block; margin-bottom: 4px; }
        .team-opt .t-name { font-weight: 700; font-size: 0.85rem; }
        .team-opt .t-desc { color: var(--text-dim); font-size: 0.7rem; margin-top: 2px; }

        .btn-launch {
            display: block; width: 100%; padding: 14px;
            background: linear-gradient(135deg, var(--accent), #9333ea);
            color: white; border: none; border-radius: 8px;
            font-family: inherit; font-size: 1rem; font-weight: 700;
            cursor: pointer; letter-spacing: 1px; transition: all 0.2s; margin-top: 10px;
        }
        .btn-launch:hover { transform: translateY(-1px); box-shadow: 0 4px 20px var(--accent-glow); }

        /* ── War Room Layout ── */
        .war-room { display: none; }
        .war-room.active { display: block; }

        .phase-banner {
            text-align: center; padding: 12px;
            background: var(--bg-card); border-bottom: 1px solid var(--border);
        }
        .phase-banner .phase-text { font-size: 1.1rem; color: var(--gold); }

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
        .desk.done { border-color: var(--green); }

        @keyframes glow {
            0%, 100% { box-shadow: 0 0 15px var(--accent-glow); }
            50% { box-shadow: 0 0 30px var(--accent-glow), 0 0 60px rgba(124,58,237,0.1); }
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
        .desk.thinking .avatar::after { background: #fbbf24; animation: blink 1s infinite; }
        .desk.done .avatar::after { background: var(--green); }

        @keyframes blink { 0%, 100% { opacity: 1; } 50% { opacity: 0.3; } }

        .agent-info { flex: 1; min-width: 0; }
        .agent-name { font-weight: 700; font-size: 0.95rem; }
        .agent-role { color: var(--text-dim); font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.5px; }

        .status-badge {
            font-size: 0.65rem; padding: 2px 8px; border-radius: 10px;
            text-transform: uppercase; font-weight: 700; letter-spacing: 0.5px;
        }
        .status-idle { background: var(--bg-desk); color: var(--text-dim); }
        .status-thinking { background: rgba(251,191,36,0.15); color: #fbbf24; }
        .status-done { background: rgba(16,185,129,0.15); color: var(--green); }

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
        .sidebar-section h3 { color: var(--gold); font-size: 0.8rem; text-transform: uppercase; letter-spacing: 1px; margin-bottom: 10px; }

        .idea-card {
            background: var(--bg-desk); border-radius: 8px; padding: 10px 12px; margin-bottom: 8px;
            border-left: 3px solid var(--green);
        }
        .idea-card .idea-title { font-weight: 700; font-size: 0.85rem; margin-bottom: 2px; }
        .idea-card .idea-score { color: var(--gold); font-size: 0.75rem; }
        .idea-card.winner { border-left-color: var(--gold); background: rgba(255,215,0,0.05); }

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
            background: linear-gradient(135deg, var(--accent), #9333ea); color: white;
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
        .btn-new:hover { background: #9333ea; }

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
        <h1>⚔️ THE WAR ROOM</h1>
        <div class="subtitle">Multi-Agent Collaborative Ideation Command Center</div>
    </div>

    <!-- ── Setup Form ── -->
    <div id="setup" class="setup-panel">
        <h2>🎯 Mission Briefing</h2>

        <div class="field">
            <label>API Key</label>
            <input type="password" id="apiKey" placeholder="Leave blank to use server environment">
            <small>Optional — falls back to LLMPROXY_KEY / OPENAI_API_KEY from server env</small>
        </div>

        <div class="field">
            <label>Squad Configuration</label>
            <div class="team-options">
                <div class="team-opt selected" onclick="selectTeam(this,'standard')">
                    <span class="t-icon">⚡</span>
                    <div class="t-name">Strike Team</div>
                    <div class="t-desc">4 agents · 1 round</div>
                </div>
                <div class="team-opt" onclick="selectTeam(this,'extended')">
                    <span class="t-icon">🔬</span>
                    <div class="t-name">Recon Squad</div>
                    <div class="t-desc">6 agents · 2 rounds</div>
                </div>
                <div class="team-opt" onclick="selectTeam(this,'full')">
                    <span class="t-icon">🚀</span>
                    <div class="t-name">Full Battalion</div>
                    <div class="t-desc">7 agents · 3 rounds</div>
                </div>
            </div>
            <input type="hidden" id="teamConfig" value="standard">
        </div>

        <div class="field">
            <label>Mission Objective</label>
            <textarea id="topic" placeholder="Describe the topic you want the team to explore..."></textarea>
        </div>

        <button class="btn-launch" onclick="launchMission()">🚀 LAUNCH MISSION</button>
    </div>

    <!-- ── War Room ── -->
    <div id="warRoom" class="war-room">
        <div class="phase-banner">
            <span class="phase-text" id="phaseText">⏳ Assembling team...</span>
        </div>
        <div class="room-grid">
            <div class="desks-area" id="desksArea"></div>
            <div class="sidebar">
                <div class="sidebar-section">
                    <h3>💡 Ideas Board</h3>
                    <div id="ideasBoard"><div style="color:var(--text-dim);font-size:0.8rem;font-style:italic">Waiting for ideas...</div></div>
                </div>
                <div class="sidebar-section" style="flex-shrink:0">
                    <h3>📡 Activity Feed</h3>
                </div>
                <div class="activity-feed" id="activityFeed"></div>
            </div>
        </div>
    </div>

    <!-- ── Result Overlay ── -->
    <div id="resultOverlay" class="result-overlay">
        <div class="result-card">
            <div class="result-header">
                <span>✨ Mission Report — Idea Sheet</span>
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
                        <span class="status-badge status-idle" id="badge-${role}">standby</span>
                    </div>
                    <div class="speech-bubble empty" id="speech-${role}">Awaiting orders...</div>
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
            badge.textContent = agent.status === 'thinking' ? 'active' : agent.status === 'done' ? 'done' : 'standby';

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
                        '✅ Mission Complete! <button class="btn-new" onclick="showResult()">📄 View Report</button> <button class="btn-new" onclick="resetToSetup()">🔄 New Mission</button>';
                    showResult();
                } else if (data.status === 'failed') {
                    clearInterval(pollTimer);
                    document.getElementById('phaseText').textContent = '❌ Mission Failed — ' + (data.error || 'Unknown error');
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
		Phase:     "Assembling team...",
		PhaseIcon: "⏳",
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
