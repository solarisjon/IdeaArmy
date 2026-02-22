package models

import "time"

// Message represents a communication between agents
type Message struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "idea", "validation", "question", "response", "summary"
}

// Idea represents a generated idea
type Idea struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Pros        []string `json:"pros"`
	Cons        []string `json:"cons"`
	Category    string   `json:"category"`
	CreatedBy   string   `json:"created_by"`
	Validated   bool     `json:"validated"`
	Score       float64  `json:"score"` // validation score 0-10
}

// Discussion represents the complete discussion session
type Discussion struct {
	ID        string    `json:"id"`
	Topic     string    `json:"topic"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Messages  []Message `json:"messages"`
	Ideas     []Idea    `json:"ideas"`
	FinalIdea *Idea     `json:"final_idea"`
	Summary   string    `json:"summary"`
	Status    string    `json:"status"`     // "running", "completed", "failed"
	Round     int       `json:"round"`      // Current discussion round
	MaxRounds int       `json:"max_rounds"` // Maximum rounds to run
}

// AgentRole defines the role of an agent
type AgentRole string

const (
	RoleTeamLeader  AgentRole = "team_leader"
	RoleIdeation    AgentRole = "ideation"
	RoleModerator   AgentRole = "moderator"
	RoleUICreator   AgentRole = "ui_creator"
	RoleResearcher  AgentRole = "researcher"
	RoleCritic      AgentRole = "critic"
	RoleImplementer AgentRole = "implementer"
)

// AgentResponse represents an agent's response
type AgentResponse struct {
	AgentRole AgentRole              `json:"agent_role"`
	Content   string                 `json:"content"`
	Ideas     []Idea                 `json:"ideas,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
