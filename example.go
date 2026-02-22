// Example: Programmatic usage of the AI Agent Team system
//
// This file demonstrates how to use the orchestrator programmatically
// in your own Go applications.
//
// Build and run: go run example.go

package main

import (
	"fmt"
	"log"

	"github.com/yourusername/ai-agent-team/internal/llmfactory"
	"github.com/yourusername/ai-agent-team/internal/orchestrator"
)

func main() {
	// Create LLM client (auto-detects backend from env vars)
	client, err := llmfactory.NewClientAuto("")
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v\nSet ANTHROPIC_API_KEY, LLMPROXY_KEY, OPENAI_API_KEY, or LLM_API_KEY.", err)
	}

	// Create the orchestrator
	orch := orchestrator.NewOrchestrator(client)

	// Set up a progress callback to monitor the discussion
	orch.OnProgress = func(message string) {
		fmt.Println("📢", message)
	}

	// Define your topic
	topic := "Innovative approaches to reduce single-use plastics in daily life"

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║     AI Agent Team - Programmatic Example              ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Printf("\nTopic: %s\n\n", topic)

	// Start the discussion
	if err := orch.StartDiscussion(topic); err != nil {
		log.Fatalf("Discussion failed: %v", err)
	}

	// Access the results
	discussion := orch.GetDiscussion()

	fmt.Println("\n" + "═"*60)
	fmt.Println("Results:")
	fmt.Printf("  Ideas Generated: %d\n", len(discussion.Ideas))
	fmt.Printf("  Messages Exchanged: %d\n", len(discussion.Messages))

	// Print all ideas with scores
	fmt.Println("\n💡 All Ideas:")
	for i, idea := range discussion.Ideas {
		fmt.Printf("\n%d. %s\n", i+1, idea.Title)
		fmt.Printf("   %s\n", idea.Description)
		if idea.Validated {
			fmt.Printf("   Score: %.1f/10\n", idea.Score)
		}
	}

	// Print the final selected idea
	if discussion.FinalIdea != nil {
		fmt.Println("\n⭐ Final Selected Idea:")
		fmt.Printf("   Title: %s\n", discussion.FinalIdea.Title)
		fmt.Printf("   Description: %s\n", discussion.FinalIdea.Description)
		fmt.Printf("   Score: %.1f/10\n", discussion.FinalIdea.Score)

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

	// Get and save the HTML idea sheet
	html := orch.GetIdeaSheetHTML()
	if html != "" {
		outputFile := "example_idea_sheet.html"
		if err := os.WriteFile(outputFile, []byte(html), 0644); err != nil {
			log.Printf("Warning: Could not save idea sheet: %v", err)
		} else {
			fmt.Printf("\n📄 Idea sheet saved to: %s\n", outputFile)
		}
	}

	fmt.Println("\n✅ Done!")
}
