package models

// TeamConfig defines the configuration for the agent team
type TeamConfig struct {
	// Core agents (always included)
	IncludeTeamLeader bool
	IncludeUICreator  bool

	// Discussion agents (configurable)
	IncludeIdeation    bool
	IncludeModerator   bool
	IncludeResearcher  bool
	IncludeCritic      bool
	IncludeImplementer bool

	// Discussion settings
	MaxRounds          int     // Number of discussion rounds
	MinIdeas           int     // Minimum ideas to generate
	DeepDive           bool    // Enable deep dive mode with more back-and-forth

	// Quality settings
	MinScoreThreshold  float64 // Minimum score for ideas to be considered
}

// DefaultTeamConfig returns a standard team configuration
func DefaultTeamConfig() *TeamConfig {
	return &TeamConfig{
		IncludeTeamLeader:  true,
		IncludeUICreator:   true,
		IncludeIdeation:    true,
		IncludeModerator:   true,
		IncludeResearcher:  false,
		IncludeCritic:      false,
		IncludeImplementer: false,
		MaxRounds:          1,
		MinIdeas:           3,
		DeepDive:           false,
		MinScoreThreshold:  6.0,
	}
}

// StandardTeamConfig returns a 4-agent team (original)
func StandardTeamConfig() *TeamConfig {
	return &TeamConfig{
		IncludeTeamLeader:  true,
		IncludeUICreator:   true,
		IncludeIdeation:    true,
		IncludeModerator:   true,
		IncludeResearcher:  false,
		IncludeCritic:      false,
		IncludeImplementer: false,
		MaxRounds:          1,
		MinIdeas:           3,
		DeepDive:           false,
		MinScoreThreshold:  6.0,
	}
}

// ExtendedTeamConfig returns a 6-agent team for deeper analysis
func ExtendedTeamConfig() *TeamConfig {
	return &TeamConfig{
		IncludeTeamLeader:  true,
		IncludeUICreator:   true,
		IncludeIdeation:    true,
		IncludeModerator:   true,
		IncludeResearcher:  true,
		IncludeCritic:      true,
		IncludeImplementer: false,
		MaxRounds:          2,
		MinIdeas:           4,
		DeepDive:           true,
		MinScoreThreshold:  7.0,
	}
}

// FullTeamConfig returns all 7 agents for maximum depth
func FullTeamConfig() *TeamConfig {
	return &TeamConfig{
		IncludeTeamLeader:  true,
		IncludeUICreator:   true,
		IncludeIdeation:    true,
		IncludeModerator:   true,
		IncludeResearcher:  true,
		IncludeCritic:      true,
		IncludeImplementer: true,
		MaxRounds:          3,
		MinIdeas:           5,
		DeepDive:           true,
		MinScoreThreshold:  7.5,
	}
}

// GetActiveAgentRoles returns a list of active agent roles based on config
func (c *TeamConfig) GetActiveAgentRoles() []AgentRole {
	var roles []AgentRole

	if c.IncludeTeamLeader {
		roles = append(roles, RoleTeamLeader)
	}
	if c.IncludeIdeation {
		roles = append(roles, RoleIdeation)
	}
	if c.IncludeModerator {
		roles = append(roles, RoleModerator)
	}
	if c.IncludeResearcher {
		roles = append(roles, RoleResearcher)
	}
	if c.IncludeCritic {
		roles = append(roles, RoleCritic)
	}
	if c.IncludeImplementer {
		roles = append(roles, RoleImplementer)
	}
	if c.IncludeUICreator {
		roles = append(roles, RoleUICreator)
	}

	return roles
}

// TeamSize returns the number of agents in the team
func (c *TeamConfig) TeamSize() int {
	return len(c.GetActiveAgentRoles())
}
