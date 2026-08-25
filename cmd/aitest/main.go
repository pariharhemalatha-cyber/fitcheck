package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ashokparihar/fitcheck/internal/ai"
	"github.com/ashokparihar/fitcheck/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("FAIL config: %v\n", err)
		os.Exit(1)
	}

	client, err := ai.NewClientFromConfig(cfg)
	if err != nil {
		fmt.Printf("FAIL init: %v\n", err)
		os.Exit(1)
	}
	if client == nil {
		fmt.Println("NO_KEY")
		fmt.Println("Add GEMINI_API_KEY to .env — get a free key at https://aistudio.google.com/apikey")
		os.Exit(2)
	}

	reply, err := ai.TestConnection(context.Background(), client)
	if err != nil {
		fmt.Printf("FAIL api (%s): %v\n", client.Provider, err)
		os.Exit(1)
	}

	fmt.Printf("OK provider=%s reply=%q\n", client.Provider, reply)

	// Vision test with existing upload if present
	testImage := "uploads/49226443-86ec-49a4-8ab0-bb86844d6dcd.jpeg"
	if _, err := os.Stat(testImage); err == nil {
		if verr := ai.AnalyzeItemVisionError(context.Background(), client, testImage); verr != nil {
			raw, _ := ai.AnalyzeItemVisionRaw(context.Background(), client, testImage)
			fmt.Printf("WARN vision failed: %v\n", verr)
			if raw != "" {
				fmt.Printf("     raw response (first 300 chars): %q\n", truncate(raw, 300))
			}
			fmt.Println("     (uploads still work via heuristics; large photos may need resize)")
		} else {
			attrs, _ := ai.AnalyzeItem(context.Background(), client, testImage)
			fmt.Printf("OK vision name=%q category=%q color=%q material=%q formality=%d\n",
				attrs.Name, attrs.Category, attrs.MainColor, attrs.Material, attrs.Formality)
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
