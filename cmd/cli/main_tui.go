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

	"github.com/yourusername/ai-agent-team/internal/llmfactory"
	"github.com/yourusername/ai-agent-team/internal/models"
	"github.com/yourusername/ai-agent-team/internal/tui"
)

func main() {
	// Print banner
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║   AI Agent Team TUI - Beautiful Terminal Interface    ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create LLM client (auto-detects backend from env vars)
	client, err := llmfactory.NewClientAuto("")
	if err != nil {
		// Fall back to interactive prompt
		fmt.Print("Enter your API key: ")
		reader := bufio.NewReader(os.Stdin)
		key, _ := reader.ReadString('\n')
		apiKey := strings.TrimSpace(key)
		if apiKey == "" {
			log.Fatal("API key is required. Set ANTHROPIC_API_KEY, LLMPROXY_KEY, OPENAI_API_KEY, or LLM_API_KEY.")
		}
		client, err = llmfactory.NewClientAuto(apiKey)
		if err != nil {
			log.Fatalf("Failed to create LLM client: %v", err)
		}
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

	fmt.Println("\n🚀 Launching AI Agent Team...\n")
	time.Sleep(500 * time.Millisecond) // Brief pause for effect

	// Run the TUI-based discussion
	discussion, err := tui.Run(client, config, topic)
	if err != nil {
		log.Fatalf("Discussion failed: %v", err)
	}

	// Extract the idea sheet HTML from the discussion results
	var html string
	if discussion != nil {
		for _, msg := range discussion.Messages {
			if msg.Type == "visualization" {
				html = msg.Content
				break
			}
		}
	}

	// Save the idea sheet HTML
	if html != "" {
		outputFile := filepath.Join(".", fmt.Sprintf("idea_sheet_%d.html", time.Now().Unix()))
		if err := os.WriteFile(outputFile, []byte(html), 0644); err != nil {
			log.Printf("Warning: Could not save idea sheet: %v", err)
		} else {
			fmt.Printf("\n📄 Idea sheet saved to: %s\n", outputFile)
		}
	}

	// Print final summary
	if discussion != nil && discussion.FinalIdea != nil {
		fmt.Println("\n" + strings.Repeat("═", 60))
		fmt.Printf("\n⭐ FINAL SELECTED IDEA:\n\n")
		fmt.Printf("   %s\n", discussion.FinalIdea.Title)
		fmt.Printf("   Score: %.1f/10\n\n", discussion.FinalIdea.Score)
		fmt.Printf("   %s\n", discussion.FinalIdea.Description)

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

		fmt.Println("\n" + strings.Repeat("═", 60))
	}

	fmt.Println("\n🎉 Thank you for using AI Agent Team!")
}

func selectTeamConfig() *models.TeamConfig {
	fmt.Println("🤖 Select Team Configuration:\n")
	fmt.Println("1. Standard (4 agents, 1 round) - Quick, focused ideation")
	fmt.Println("2. Extended (6 agents, 2 rounds) - Deeper analysis [Recommended]")
	fmt.Println("3. Full (7 agents, 3 rounds) - Maximum depth")
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
