package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourusername/ai-agent-team/internal/orchestrator"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║        AI Agent Team - Collaborative Ideation          ║")
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

	// Create orchestrator
	orch := orchestrator.NewOrchestrator(apiKey)

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
