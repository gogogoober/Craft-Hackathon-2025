# Agent Reference Guide - Craft Hackathon 2025

This document serves as a comprehensive reference for building features in this Craft API integration project. Use this guide when implementing new functionality.

## Table of Contents
1. [Project Architecture](#project-architecture)
2. [Development Workflow](#development-workflow)
3. [Craft API Patterns](#craft-api-patterns)
4. [Best Practices](#best-practices)
5. [Common Use Cases](#common-use-cases)
6. [Error Handling](#error-handling)
7. [Testing & Safety](#testing--safety)

---

## Project Architecture

### Directory Structure
```
Craft-Hackathon-2025/
├── main.go              # HTTP server & application entry point
├── client/
│   └── client.go        # Craft API client library
├── docs/
│   ├── craft-docs.md    # Official Craft API documentation
│   └── agent-reference.md # This file
└── go.mod               # Go module definition
```

### Current Components

#### 1. HTTP Server (main.go)
- **Purpose**: Local server to receive voice-dictated queries
- **Endpoint**: `POST localhost:8080/craft-hackathon`
- **Input**: `{"query": "dictated text"}`
- **Output**: `{"status": "received", "query": "echoed text"}`
- **Location**: [main.go](../main.go)

#### 2. Craft API Client (client/client.go)
- **Purpose**: Complete Go client for Craft Blocks API
- **Base URL**: `https://connect.craft.do/links/JwH02Yc5RGk/api/v1`
- **Available Methods**:
  - `FetchBlocks()` - Get document structure
  - `FetchBlocksMarkdown()` - Get as markdown
  - `InsertBlocks()` - Add new content
  - `UpdateBlocks()` - Modify existing content
  - `DeleteBlocks()` - Remove content
  - `MoveBlocks()` - Reposition blocks
  - `Search()` - Find content with regex
  - `GenerateUploadURL()` - Get S3 upload link

---

## Development Workflow

### When Building New Features

1. **Always consult [docs/craft-docs.md](craft-docs.md)** for:
   - API endpoint specifications
   - Request/response formats
   - Block structure details
   - Position semantics

2. **Use the existing client package** - Don't write raw HTTP calls:
   ```go
   // Good
   client := client.NewClient(BaseURL)
   blocks, err := client.FetchBlocks("", -1, false)

   // Bad - Don't do this
   resp, err := http.Get("https://...")
   ```

3. **Follow the hierarchical block model**:
   - Everything is a Block with an ID
   - Pages contain nested content
   - Preserve hierarchy when possible

4. **Test with real API calls**:
   - No mocking or simulation
   - Verify responses match expectations
   - See [Testing & Safety](#testing--safety) section

---

## Craft API Patterns

### Block Hierarchy

```
Root Page (id: "0")
└── Content Blocks
    ├── Text Block (id: "1")
    ├── Page Block (id: "2")
    │   └── Nested Content
    │       ├── Text Block (id: "3")
    │       └── Image Block (id: "4")
    └── Text Block (id: "5")
```

### Key Concepts

#### Block Types
- `text` - Markdown content with optional styling
- `page` - Container for hierarchical content
- `image` - Image with URL, alt text, dimensions
- `video` - Video with URL
- `file` - File attachment

#### Text Styles
- `page` - Page title
- `card` - Card/section title
- `h1`, `h2`, `h3` - Headers
- `body` - Normal text
- `title`, `subtitle`, `caption` - Other styles

#### Position Semantics
```go
// Add to end of a page
Position{Position: "end", PageID: "0"}

// Add to start of a page
Position{Position: "start", PageID: "0"}

// Insert before another block
Position{Position: "before", SiblingID: "5"}

// Insert after another block
Position{Position: "after", SiblingID: "5"}
```

### Common Operations

#### 1. Fetch Document Structure
```go
// Get everything (expensive)
root, err := client.FetchBlocks("", -1, false)

// Get shallow (recommended for exploration)
root, err := client.FetchBlocks("", 1, false)

// Get specific block and children
block, err := client.FetchBlocks("block-id", 2, false)
```

**Reference**: [craft-docs.md](craft-docs.md#fetch-blocks) lines 36-133

#### 2. Insert Text Content
```go
// Using markdown directly (simple)
req := client.InsertRequest{
    Markdown: "## New Section\n\nSome content here",
    Position: client.Position{
        Position: "end",
        PageID:   "0",
    },
}
blocks, err := client.InsertBlocks(req)

// Using structured blocks (more control)
req := client.InsertRequest{
    Blocks: []client.Block{
        {
            Type:     "text",
            Markdown: "## Header Text",
        },
    },
    Position: client.Position{
        Position: "end",
        PageID:   "0",
    },
}
blocks, err := client.InsertBlocks(req)
```

**Reference**: [craft-docs.md](craft-docs.md#insert-blocks) lines 172-336

#### 3. Search Content
```go
// Basic search
matches, err := client.Search("TODO", false, 0, 0)

// With context (before/after blocks)
matches, err := client.Search("important", false, 2, 2)

// Case-sensitive regex search
matches, err := client.Search("^# .*Header", true, 1, 1)
```

**Reference**: [craft-docs.md](craft-docs.md#search-in-document) lines 547-711

#### 4. Update Existing Content
```go
req := client.UpdateRequest{
    Blocks: []client.Block{
        {
            ID:       "5",
            Markdown: "## Updated Header",
        },
    },
}
updated, err := client.UpdateBlocks(req)
```

**Reference**: [craft-docs.md](craft-docs.md#update-blocks) lines 385-445

#### 5. File Upload (3-Step Process)
```go
// Step 1: Get upload URL
uploadResp, err := client.GenerateUploadURL("photo.jpg", "image/jpeg")

// Step 2: Upload to S3 (using curl or http.Put)
// curl -T /path/to/file "uploadResp.UploadURL" -H "Content-Type: image/jpeg"

// Step 3: Insert block with URL
req := client.InsertRequest{
    Blocks: []client.Block{
        {
            Type:    "image",
            URL:     uploadResp.RawURL,
            AltText: "Description",
            Width:   "auto",
        },
    },
    Position: client.Position{Position: "end", PageID: "0"},
}
blocks, err := client.InsertBlocks(req)
```

**Reference**: [craft-docs.md](craft-docs.md#generate-upload-url) lines 449-494

---

## Best Practices

### 1. Preserve Hierarchy
When working with existing content, maintain the document structure:
- Fetch with appropriate depth
- Understand parent-child relationships
- Use Move operations when reordering

### 2. Use Markdown Wisely
- For simple text, use `markdown` field directly
- For complex structures, use block arrays
- Paragraph-level markdown only (single block worth)

### 3. Handle Errors Gracefully
```go
blocks, err := client.FetchBlocks("", 1, false)
if err != nil {
    log.Printf("Failed to fetch: %v", err)
    // Return error to user, don't crash
    return
}
```

### 4. Log Operations
```go
timestamp := time.Now().Format("2006-01-02 15:04:05")
fmt.Printf("[%s] Inserted block: %s\n", timestamp, blocks[0].ID)
```

### 5. Batch When Possible
```go
// Good - single request
req := client.InsertRequest{
    Blocks: []client.Block{block1, block2, block3},
    Position: position,
}

// Less efficient - multiple requests
client.InsertBlocks(req1)
client.InsertBlocks(req2)
client.InsertBlocks(req3)
```

---

## Common Use Cases

### Voice-to-Document Integration

**Goal**: Take dictated text, process it, insert into Craft

```go
func handleCraftHackathon(w http.ResponseWriter, r *http.Request) {
    // 1. Parse incoming query
    var req QueryRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 2. Process the query (your logic here)
    processedContent := processQuery(req.Query)

    // 3. Insert into Craft
    c := client.NewClient(BaseURL)
    insertReq := client.InsertRequest{
        Markdown: processedContent,
        Position: client.Position{
            Position: "end",
            PageID:   "0", // Root page
        },
    }
    blocks, err := c.InsertBlocks(insertReq)

    // 4. Return success
    response := QueryResponse{
        Status: "inserted",
        Query:  req.Query,
    }
    json.NewEncoder(w).Encode(response)
}
```

### Intelligent Content Organization

**Goal**: Parse voice command, create structured hierarchy

```go
// Input: "Create a new section called Projects with subsections for Web and Mobile"
// Output: Nested page structure

// 1. Create main section
mainBlock := client.Block{
    Type:      "page",
    TextStyle: "card",
    Markdown:  "<card>Projects</card>",
}

// 2. Insert subsections
// After inserting mainBlock, you get its ID, then add children
```

### Search and Update

**Goal**: Find and update specific content

```go
// 1. Search for content
matches, err := client.Search("TODO:", false, 0, 0)

// 2. Update each match
for _, match := range matches {
    // Parse TODO and mark as done
    newContent := strings.Replace(match.Markdown, "TODO:", "DONE:", 1)

    req := client.UpdateRequest{
        Blocks: []client.Block{
            {ID: match.BlockID, Markdown: newContent},
        },
    }
    client.UpdateBlocks(req)
}
```

---

## Error Handling

### API Error Responses
- `400` - Bad request (invalid JSON, malformed data)
- `404` - Block not found
- `405` - Wrong HTTP method
- `207` - Partial success (some operations failed)

### Client Error Handling Pattern
```go
blocks, err := client.SomeOperation(...)
if err != nil {
    // Check if it's an HTTP error
    if strings.Contains(err.Error(), "status 400") {
        // Handle validation error
    } else if strings.Contains(err.Error(), "status 404") {
        // Handle not found
    } else {
        // Handle general error
    }
    return err
}
```

### Validation Before API Calls
```go
// Validate markdown doesn't create multiple blocks
if strings.Count(markdown, "\n\n") > 3 {
    // Too complex, might create multiple blocks
    // Split or simplify
}

// Ensure block IDs exist before updating
if blockID == "" {
    return errors.New("block ID required for update")
}
```

---

## Testing & Safety

### IMPORTANT: Production Data Warning
From [craft-docs.md](craft-docs.md) lines 18-25:

> **This is a production server connected to real user data.**
>
> Only perform testing operations that can be safely rolled back:
> - Safe: Reading data, creating test content you delete immediately
> - Safe: Modifying content if you can restore it
> - Safe: Moving blocks if you can move them back
> - Unsafe: Permanent deletions, modifications without backup

### Testing Checklist

1. **Before Writing**:
   ```go
   // Fetch and backup original state
   original, err := client.FetchBlocks(blockID, -1, false)
   // Store original for rollback
   ```

2. **After Writing**:
   ```go
   // Verify the change worked
   updated, err := client.FetchBlocks(blockID, -1, false)
   // Compare with expected
   ```

3. **Cleanup**:
   ```go
   // Delete test blocks
   deletedIDs, err := client.DeleteBlocks([]string{testBlockID})
   // Verify deletion
   ```

### Testing Pattern
```go
func testInsertAndRollback() error {
    // 1. Insert test content
    insertReq := client.InsertRequest{
        Markdown: "TEST CONTENT - DELETE ME",
        Position: client.Position{Position: "end", PageID: "0"},
    }
    blocks, err := client.InsertBlocks(insertReq)
    if err != nil {
        return err
    }
    testID := blocks[0].ID

    // 2. Verify insert worked
    fetched, err := client.FetchBlocks(testID, 0, false)
    if err != nil || fetched.Markdown != "TEST CONTENT - DELETE ME" {
        return errors.New("verification failed")
    }

    // 3. Clean up
    _, err = client.DeleteBlocks([]string{testID})
    return err
}
```

---

## Quick Reference Cheat Sheet

### Essential Client Methods

| Method | Purpose | Key Parameters | Returns |
|--------|---------|----------------|---------|
| `FetchBlocks()` | Get document structure | id, maxDepth, fetchMetadata | `*Block` |
| `FetchBlocksMarkdown()` | Get as markdown | id, maxDepth | `string` |
| `InsertBlocks()` | Add new content | InsertRequest | `[]Block` |
| `UpdateBlocks()` | Modify content | UpdateRequest | `[]Block` |
| `DeleteBlocks()` | Remove content | blockIDs | `[]string` |
| `MoveBlocks()` | Reposition | MoveRequest | `[]string` |
| `Search()` | Find with regex | pattern, caseSensitive, context | `[]SearchMatch` |
| `GenerateUploadURL()` | Get S3 upload link | fileName, mimeType | `*UploadLinkResponse` |

### Position Quick Reference
```go
// End of page
Position{Position: "end", PageID: "page-id"}

// Start of page
Position{Position: "start", PageID: "page-id"}

// Before sibling
Position{Position: "before", SiblingID: "sibling-id"}

// After sibling
Position{Position: "after", SiblingID: "sibling-id"}
```

### Block Type Quick Reference
```go
// Text block
Block{Type: "text", Markdown: "content"}

// Page block
Block{Type: "page", TextStyle: "card", Markdown: "<card>Title</card>"}

// Image block
Block{Type: "image", URL: "https://...", AltText: "desc", Width: "auto"}

// File block
Block{Type: "file", URL: "https://...", FileName: "doc.pdf"}
```

---

## Integration with Current Server

### Current State (main.go)
- HTTP server on `localhost:8080`
- Receives POST to `/craft-hackathon`
- Parses `{"query": "string"}`
- Logs to console
- Returns success response

### Next Steps for Integration

1. **Parse Intent from Query**:
   - Identify if query is a command vs. content
   - Extract action (add, search, update, etc.)
   - Extract target location

2. **Execute Against Craft API**:
   - Use appropriate client method
   - Handle errors gracefully
   - Log operations

3. **Return Rich Response**:
   - Include block IDs in response
   - Show what was created/modified
   - Provide confirmation

### Example Enhancement
```go
func handleCraftHackathon(w http.ResponseWriter, r *http.Request) {
    var req QueryRequest
    json.NewDecoder(r.Body).Decode(&req)

    timestamp := time.Now().Format("2006-01-02 15:04:05")
    fmt.Printf("[%s] Received query: %s\n", timestamp, req.Query)

    // NEW: Process with Craft API
    c := client.NewClient(BaseURL)

    // Determine intent and execute
    action, content := parseIntent(req.Query)
    var resultBlockID string

    switch action {
    case "add":
        blocks, err := c.InsertBlocks(client.InsertRequest{
            Markdown: content,
            Position: client.Position{Position: "end", PageID: "0"},
        })
        if err == nil {
            resultBlockID = blocks[0].ID
        }
    case "search":
        matches, _ := c.Search(content, false, 1, 1)
        // Process matches...
    }

    // Enhanced response
    response := QueryResponse{
        Status:  "processed",
        Query:   req.Query,
        BlockID: resultBlockID, // NEW field
    }

    json.NewEncoder(w).Encode(response)
}
```

---

## Resources

- **API Documentation**: [docs/craft-docs.md](craft-docs.md)
- **Client Implementation**: [client/client.go](../client/client.go)
- **Server Code**: [main.go](../main.go)
- **API Base URL**: `https://connect.craft.do/links/JwH02Yc5RGk/api/v1`

---

## Agent Instructions

When asked to implement features:

1. **First**, read this document to understand patterns
2. **Then**, consult [craft-docs.md](craft-docs.md) for API specifics
3. **Always** use the existing `client` package
4. **Never** hardcode or mock API responses
5. **Test** with real API calls
6. **Preserve** document hierarchy
7. **Log** all operations with timestamps
8. **Handle** errors gracefully
9. **Clean up** test data
10. **Document** new patterns in this file

---

*Last Updated: 2025-11-16*
*Version: 1.0*
