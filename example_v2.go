// Example: Programmatic usage of AI Agent Team v2 with configurable teams
//
// This demonstrates the new v2 features including multi-agent teams,
// multi-round discussions, and custom configurations.
//
// Build and run: go run example_v2.go

package main

import (
	"fmt"
	"log"

	"github.com/yourusername/ai-agent-team/internal/llm"
	"github.com/yourusername/ai-agent-team/internal/llmfactory"
	"github.com/yourusername/ai-agent-team/internal/models"
	"github.com/yourusername/ai-agent-team/internal/orchestrator"
)

func main() {
	// Create LLM client (auto-detects backend from env vars)
	client, err := llmfactory.NewClientAuto("")
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v\nSet ANTHROPIC_API_KEY, LLMPROXY_KEY, OPENAI_API_KEY, or LLM_API_KEY.", err)
	}

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║   AI Agent Team v2 - Programmatic Example             ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Example 1: Using Extended configuration for deeper analysis
	runExtendedTeamExample(client)

	// Example 2: Custom configuration for specific use case
	// Uncomment to try:
	// runCustomTeamExample(client)
}

// Example 1: Extended team with 6 agents and 2 rounds
func runExtendedTeamExample(client llm.Client) {
	topic := "Innovative approaches to reduce plastic waste in urban environments"

	// Use the Extended preset: 6 agents, 2 rounds, deep dive mode
	config := models.ExtendedTeamConfig()

	fmt.Println("📊 Example 1: Extended Team Configuration")
	fmt.Printf("   Topic: %s\n", topic)
	fmt.Printf("   Team Size: %d agents\n", config.TeamSize())
	fmt.Printf("   Rounds: %d\n", config.MaxRounds)
	fmt.Printf("   Deep Dive: %v\n\n", config.DeepDive)

	orch := orchestrator.NewConfigurableOrchestrator(client, config)

	// Track progress
	roundsCompleted := 0
	orch.OnProgress = func(message string) {
		fmt.Println("📢", message)
		if message[:4] == "🔄 R" {
			roundsCompleted++
		}
	}

	// Run the discussion
	fmt.Println("Starting discussion...")
	if err := orch.StartDiscussion(topic); err != nil {
		log.Fatalf("Discussion failed: %v", err)
	}

	// Analyze results
	discussion := orch.GetDiscussion()
	printResults(discussion)

	// Save the idea sheet
	saveIdeaSheet(orch, "extended_team")
}

// Example 2: Custom team for a technical architecture decision
func runCustomTeamExample(client llm.Client) {
	topic := "Microservices vs Monolithic architecture for a new SaaS product"

	// Build a custom team focused on technical analysis
	config := &models.TeamConfig{
		IncludeTeamLeader:  true, // Required
		IncludeIdeation:    true, // For architecture ideas
		IncludeModerator:   true, // For evaluation
		IncludeResearcher:  true, // For technology research
		IncludeCritic:      true, // For identifying technical risks
		IncludeImplementer: true, // For practical implementation planning
		IncludeUICreator:   true, // For final visualization
		MaxRounds:          3,    // Deep exploration
		MinIdeas:           5,
		DeepDive:           true,
		MinScoreThreshold:  7.5, // High quality bar
	}

	fmt.Println("\n" + "═"*60)
	fmt.Println("📊 Example 2: Custom Team Configuration")
	fmt.Printf("   Topic: %s\n", topic)
	fmt.Printf("   Team: Full 7-agent team\n")
	fmt.Printf("   Rounds: %d (maximum depth)\n", config.MaxRounds)
	fmt.Println()

	orch := orchestrator.NewConfigurableOrchestrator(client, config)

	orch.OnProgress = func(message string) {
		fmt.Println("📢", message)
	}

	fmt.Println("Starting technical discussion...")
	if err := orch.StartDiscussion(topic); err != nil {
		log.Fatalf("Discussion failed: %v", err)
	}

	discussion := orch.GetDiscussion()
	printResults(discussion)
	saveIdeaSheet(orch, "custom_team")
}

// Print discussion results
func printResults(discussion *models.Discussion) {
	fmt.Println("\n" + "═"*60)
	fmt.Println("📊 Results:")
	fmt.Printf("   Rounds Completed: %d/%d\n", discussion.Round, discussion.MaxRounds)
	fmt.Printf("   Ideas Generated: %d\n", len(discussion.Ideas))
	fmt.Printf("   Messages Exchanged: %d\n", len(discussion.Messages))
	fmt.Printf("   Duration: %v\n", discussion.EndTime.Sub(discussion.StartTime))

	// Show all ideas with scores
	fmt.Println("\n💡 All Ideas Generated:")
	for i, idea := range discussion.Ideas {
		fmt.Printf("\n%d. %s", i+1, idea.Title)
		if idea.Validated {
			fmt.Printf(" [Score: %.1f/10]", idea.Score)
		}
		fmt.Printf("\n   %s\n", idea.Description)

		if idea.Validated && len(idea.Pros) > 0 {
			fmt.Println("   ✅ Strengths:", idea.Pros[0])
		}
		if idea.Validated && len(idea.Cons) > 0 {
			fmt.Println("   ⚠️  Concerns:", idea.Cons[0])
		}
	}

	// Show final selected idea
	if discussion.FinalIdea != nil {
		fmt.Println("\n⭐ FINAL SELECTED IDEA:")
		fmt.Printf("   %s\n", discussion.FinalIdea.Title)
		fmt.Printf("   Score: %.1f/10\n", discussion.FinalIdea.Score)
		fmt.Printf("\n   %s\n", discussion.FinalIdea.Description)

		if len(discussion.FinalIdea.Pros) > 0 {
			fmt.Println("\n   ✅ Pros:")
			for _, pro := range discussion.FinalIdea.Pros {
				fmt.Printf("      • %s\n", pro)
			}
		}

		if len(discussion.FinalIdea.Cons) > 0 {
			fmt.Println("\n   ❌ Cons:")
			for _, con := range discussion.FinalIdea.Cons {
				fmt.Printf("      • %s\n", con)
			}
		}
	}
}

// Save the HTML idea sheet
func saveIdeaSheet(orch *orchestrator.ConfigurableOrchestrator, prefix string) {
	html := orch.GetIdeaSheetHTML()
	if html == "" {
		return
	}

	filename := fmt.Sprintf("%s_idea_sheet.html", prefix)
	if err := os.WriteFile(filename, []byte(html), 0644); err != nil {
		log.Printf("Warning: Could not save idea sheet: %v", err)
		return
	}

	fmt.Printf("\n📄 Idea sheet saved: %s\n", filename)
	fmt.Println("   Open in your browser to view the beautiful visualization!")
}
