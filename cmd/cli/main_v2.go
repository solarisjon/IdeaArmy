package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/ai-agent-team/internal/models"
	"github.com/yourusername/ai-agent-team/internal/orchestrator"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║   AI Agent Team v2 - Configurable Multi-Agent System  ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Get API key - check both ANTHROPIC_API_KEY and ANTHROPIC_KEY
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_KEY")
	}

	if apiKey == "" {
		fmt.Print("Enter your Anthropic API key: ")
		reader := bufio.NewReader(os.Stdin)
		key, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(key)
	}

	if apiKey == "" {
		log.Fatal("API key is required. Set ANTHROPIC_API_KEY or ANTHROPIC_KEY environment variable, or enter it when prompted.")
	}

	// Choose team configuration
	config := selectTeamConfig()

	// Get topic from user
	fmt.Println("\n📝 What topic would you like the AI team to explore?")
	fmt.Print("> ")

	reader := bufio.NewReader(os.Stdin)
	topic, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}
	topic = strings.TrimSpace(topic)

	if topic == "" {
		log.Fatal("Topic cannot be empty")
	}

	fmt.Println("\n🚀 Starting AI agent team discussion...\n")
	printTeamComposition(config)
	fmt.Println()

	// Create orchestrator with selected configuration
	orch := orchestrator.NewConfigurableOrchestrator(apiKey, config)

	// Set up progress callback
	orch.OnProgress = func(message string) {
		fmt.Println(message)
	}

	// Run the discussion
	startTime := time.Now()
	if err := orch.StartDiscussion(topic); err != nil {
		log.Fatalf("Discussion failed: %v", err)
	}

	duration := time.Since(startTime)

	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Printf("✅ Discussion completed in %.1f seconds\n", duration.Seconds())

	// Get the idea sheet HTML
	html := orch.GetIdeaSheetHTML()
	if html != "" {
		// Save to file
		outputFile := filepath.Join(".", fmt.Sprintf("idea_sheet_%d.html", time.Now().Unix()))
		if err := os.WriteFile(outputFile, []byte(html), 0644); err != nil {
			log.Printf("Warning: Could not save idea sheet: %v", err)
		} else {
			fmt.Printf("📄 Idea sheet saved to: %s\n", outputFile)
			fmt.Println("   Open this file in your browser to view the results!")
		}
	}

	// Print summary
	discussion := orch.GetDiscussion()
	if discussion != nil {
		fmt.Println("\n📊 Discussion Summary:")
		fmt.Printf("   Topic: %s\n", discussion.Topic)
		fmt.Printf("   Team Size: %d agents\n", config.TeamSize())
		fmt.Printf("   Rounds Completed: %d\n", discussion.Round)
		fmt.Printf("   Ideas Generated: %d\n", len(discussion.Ideas))
		fmt.Printf("   Messages Exchanged: %d\n", len(discussion.Messages))

		if discussion.FinalIdea != nil {
			fmt.Printf("\n⭐ Final Selected Idea:\n")
			fmt.Printf("   Title: %s\n", discussion.FinalIdea.Title)
			fmt.Printf("   Score: %.1f/10\n", discussion.FinalIdea.Score)
			fmt.Printf("   Description: %s\n", discussion.FinalIdea.Description)

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

		fmt.Println("\n" + strings.Repeat("═", 60))
		fmt.Println("🎉 Thank you for using AI Agent Team!")
	}
}

func selectTeamConfig() *models.TeamConfig {
	fmt.Println("🤖 Select Team Configuration:\n")
	fmt.Println("1. Standard (4 agents, 1 round) - Quick, focused ideation")
	fmt.Println("   Team: Leader, Ideation, Moderator, UI Creator")
	fmt.Println()
	fmt.Println("2. Extended (6 agents, 2 rounds) - Deeper analysis")
	fmt.Println("   Team: + Researcher, Critic")
	fmt.Println("   Multiple rounds with iterative refinement")
	fmt.Println()
	fmt.Println("3. Full (7 agents, 3 rounds) - Maximum depth")
	fmt.Println("   Team: + Implementer")
	fmt.Println("   Extensive multi-round exploration")
	fmt.Println()
	fmt.Println("4. Custom - Configure your own team")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Choose configuration (1-4) [default: 2]: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice == "" {
		choice = "2"
	}

	switch choice {
	case "1":
		return models.StandardTeamConfig()
	case "2":
		return models.ExtendedTeamConfig()
	case "3":
		return models.FullTeamConfig()
	case "4":
		return customTeamConfig(reader)
	default:
		fmt.Println("Invalid choice, using Extended configuration")
		return models.ExtendedTeamConfig()
	}
}

func customTeamConfig(reader *bufio.Reader) *models.TeamConfig {
	config := models.DefaultTeamConfig()

	fmt.Println("\n🔧 Custom Team Configuration")
	fmt.Println("(Press Enter to accept default)")
	fmt.Println()

	// Core agents (required)
	config.IncludeTeamLeader = true
	config.IncludeUICreator = true

	// Optional agents
	config.IncludeIdeation = askYesNo(reader, "Include Ideation Agent? (Y/n)", true)
	config.IncludeModerator = askYesNo(reader, "Include Moderator Agent? (Y/n)", true)
	config.IncludeResearcher = askYesNo(reader, "Include Researcher Agent? (y/N)", false)
	config.IncludeCritic = askYesNo(reader, "Include Critic Agent? (y/N)", false)
	config.IncludeImplementer = askYesNo(reader, "Include Implementer Agent? (y/N)", false)

	// Rounds
	fmt.Print("Number of discussion rounds (1-5) [2]: ")
	roundsStr, _ := reader.ReadString('\n')
	roundsStr = strings.TrimSpace(roundsStr)
	if roundsStr == "" {
		config.MaxRounds = 2
	} else {
		rounds, err := strconv.Atoi(roundsStr)
		if err != nil || rounds < 1 || rounds > 5 {
			config.MaxRounds = 2
		} else {
			config.MaxRounds = rounds
		}
	}

	// Deep dive
	config.DeepDive = askYesNo(reader, "Enable deep dive mode? (Y/n)", true)

	return config
}

func askYesNo(reader *bufio.Reader, prompt string, defaultYes bool) bool {
	fmt.Print(prompt + " ")
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))

	if response == "" {
		return defaultYes
	}

	return response == "y" || response == "yes"
}

func printTeamComposition(config *models.TeamConfig) {
	fmt.Println("👥 Team Composition:")
	if config.IncludeTeamLeader {
		fmt.Println("   🎯 Team Leader - Orchestrates discussion and makes decisions")
	}
	if config.IncludeIdeation {
		fmt.Println("   💡 Ideation Agent - Generates creative ideas")
	}
	if config.IncludeModerator {
		fmt.Println("   🔍 Moderator - Validates and scores ideas")
	}
	if config.IncludeResearcher {
		fmt.Println("   📚 Researcher - Provides facts and real-world context")
	}
	if config.IncludeCritic {
		fmt.Println("   🤔 Critic - Challenges assumptions and identifies risks")
	}
	if config.IncludeImplementer {
		fmt.Println("   🔧 Implementer - Plans practical execution")
	}
	if config.IncludeUICreator {
		fmt.Println("   🎨 UI Creator - Generates beautiful visualizations")
	}

	fmt.Printf("\n📊 Discussion Settings:\n")
	fmt.Printf("   Rounds: %d\n", config.MaxRounds)
	fmt.Printf("   Deep Dive: %v\n", config.DeepDive)
	fmt.Printf("   Total Agents: %d\n", config.TeamSize())
}
