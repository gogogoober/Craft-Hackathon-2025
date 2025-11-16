package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"craft-hackathon/client"
)

const (
	// API endpoint from the documentation
	BaseURL = "https://connect.craft.do/links/3tXZdMX0EIe/api/v1"
)

// QueryRequest represents the incoming JSON payload
type QueryRequest struct {
	Query string `json:"query"`
}

// QueryResponse represents the response JSON
type QueryResponse struct {
	Status string `json:"status"`
	Query  string `json:"query"`
}

func main() {
	// Set up HTTP handler
	http.HandleFunc("/craft-hackathon", handleCraftHackathon)

	// Start server
	addr := "localhost:8080"
	fmt.Printf("Server starting on %s\n", addr)
	fmt.Println("Listening for POST requests on /craft-hackathon")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// handleCraftHackathon handles POST requests to /craft-hackathon
func handleCraftHackathon(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON body
	var req QueryRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Log the received query with timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] Received query: %s\n", timestamp, req.Query)

	// Create Craft API client
	c := client.NewClient(BaseURL)

	// Fetch the root document to get the actual root page ID
	root, err := c.FetchBlocks("", 0, false)
	if err != nil {
		log.Printf("Error fetching root: %v", err)
		http.Error(w, fmt.Sprintf("Failed to fetch document: %v", err), http.StatusInternalServerError)
		return
	}

	// Simply insert the query text as a block at the end of the document
	insertReq := client.InsertRequest{
		Markdown: req.Query,
		Position: client.Position{
			Position: "end",
			PageID:   root.ID, // Use actual root page ID
		},
	}

	insertedBlocks, err := c.InsertBlocks(insertReq)
	if err != nil {
		log.Printf("Error adding content: %v", err)
		http.Error(w, fmt.Sprintf("Failed to add content: %v", err), http.StatusInternalServerError)
		return
	}

	blockID := insertedBlocks[0].ID
	fmt.Printf("[%s] Added content to page %s with block ID: %s\n", timestamp, root.ID, blockID)

	// Prepare success response
	response := QueryResponse{
		Status: "created",
		Query:  req.Query,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// ============================================================
// COMMENTED OUT: Previous Craft API Explorer logic
// Uncomment when ready to integrate with Craft API
// ============================================================
/*
func mainOld() {
	// Create the client
	c := client.NewClient(BaseURL)

	fmt.Println("=== Craft API Explorer ===")
	fmt.Println()

	// 1. Fetch the root document structure (shallow, just top level)
	fmt.Println("1. Fetching root document structure (depth=1)...")
	root, err := c.FetchBlocks("", 1, false)
	if err != nil {
		log.Fatalf("Failed to fetch root: %v", err)
	}

	fmt.Printf("   Document Title: %s\n", root.Markdown)
	fmt.Printf("   Document Type: %s\n", root.Type)
	fmt.Printf("   Text Style: %s\n", root.TextStyle)
	fmt.Printf("   Number of top-level blocks: %d\n", len(root.Content))
	fmt.Println()

	// 2. List top-level blocks
	fmt.Println("2. Top-level blocks:")
	for i, block := range root.Content {
		fmt.Printf("   [%d] ID: %s | Type: %s | Style: %s\n", i, block.ID, block.Type, block.TextStyle)
		if block.Markdown != "" {
			// Truncate long markdown
			md := block.Markdown
			if len(md) > 80 {
				md = md[:77] + "..."
			}
			fmt.Printf("       Content: %s\n", md)
		}
	}
	fmt.Println()

	// 3. Fetch full document as JSON (for inspection)
	fmt.Println("3. Fetching full document structure...")
	fullDoc, err := c.FetchBlocks("", -1, false)
	if err != nil {
		log.Fatalf("Failed to fetch full document: %v", err)
	}

	// Count total blocks recursively
	totalBlocks := countBlocks(fullDoc)
	fmt.Printf("   Total blocks in document: %d\n", totalBlocks)
	fmt.Println()

	// 4. Test search functionality
	fmt.Println("4. Testing search (looking for 'TODO' or 'task')...")
	matches, err := c.Search("TODO|task", false, 2, 2)
	if err != nil {
		log.Printf("   Search failed: %v", err)
	} else {
		fmt.Printf("   Found %d matches\n", len(matches))
		for i, match := range matches {
			if i >= 3 { // Show first 3 only
				fmt.Printf("   ... and %d more\n", len(matches)-3)
				break
			}
			fmt.Printf("   [%d] Block %s: %s\n", i, match.BlockID, truncate(match.Markdown, 60))
		}
	}
	fmt.Println()

	// 5. Export full structure as JSON for inspection
	fmt.Println("5. Full document structure (JSON):")
	jsonData, err := json.MarshalIndent(fullDoc, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Just show first 2000 chars
	jsonStr := string(jsonData)
	if len(jsonStr) > 2000 {
		fmt.Println(jsonStr[:2000])
		fmt.Printf("\n... [truncated, full size: %d bytes]\n", len(jsonStr))
	} else {
		fmt.Println(jsonStr)
	}

	fmt.Println()
	fmt.Println("=== Exploration Complete ===")
}
*/

// countBlocks recursively counts all blocks
func countBlocks(block *client.Block) int {
	if block == nil {
		return 0
	}
	count := 1
	for i := range block.Content {
		count += countBlocks(&block.Content[i])
	}
	return count
}

// truncate shortens a string if needed
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
