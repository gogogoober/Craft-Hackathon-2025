package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Client represents the Craft API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Craft API client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// Block represents a content block in Craft
type Block struct {
	ID               string   `json:"id,omitempty"`
	Type             string   `json:"type"`
	TextStyle        string   `json:"textStyle,omitempty"`
	Markdown         string   `json:"markdown,omitempty"`
	Content          []Block  `json:"content,omitempty"`
	IndentationLevel int      `json:"indentationLevel,omitempty"`
	ListStyle        string   `json:"listStyle,omitempty"`
	Font             string   `json:"font,omitempty"`
	Color            string   `json:"color,omitempty"`
	URL              string   `json:"url,omitempty"`
	AltText          string   `json:"altText,omitempty"`
	Width            any      `json:"width,omitempty"` // Can be int or string like "auto"
	Height           int      `json:"height,omitempty"`
	FileName         string   `json:"fileName,omitempty"`
	MimeType         string   `json:"mimeType,omitempty"`
	FileSize         int64    `json:"fileSize,omitempty"`
}

// Position specifies where to insert blocks
type Position struct {
	Position  string `json:"position"`            // "start", "end", "before", "after"
	PageID    string `json:"pageId,omitempty"`    // For start/end positions
	SiblingID string `json:"siblingId,omitempty"` // For before/after positions
}

// InsertRequest represents a request to insert blocks
type InsertRequest struct {
	Blocks   []Block  `json:"blocks,omitempty"`
	Markdown string   `json:"markdown,omitempty"`
	Position Position `json:"position"`
}

// UpdateRequest represents a request to update blocks
type UpdateRequest struct {
	Blocks []Block `json:"blocks"`
}

// DeleteRequest represents a request to delete blocks
type DeleteRequest struct {
	BlockIDs []string `json:"blockIds"`
}

// MoveRequest represents a request to move blocks
type MoveRequest struct {
	BlockIDs []string `json:"blockIds"`
	Position Position `json:"position"`
}

// SearchMatch represents a single search result
type SearchMatch struct {
	BlockID       string            `json:"blockId"`
	Markdown      string            `json:"markdown"`
	PageBlockPath []PagePathElement `json:"pageBlockPath"`
	BeforeBlocks  []ContextBlock    `json:"beforeBlocks"`
	AfterBlocks   []ContextBlock    `json:"afterBlocks"`
}

// PagePathElement represents a page in the hierarchy path
type PagePathElement struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

// ContextBlock represents surrounding context in search results
type ContextBlock struct {
	BlockID  string `json:"blockId"`
	Markdown string `json:"markdown"`
}

// ItemsResponse is the generic response wrapper
type ItemsResponse struct {
	Items json.RawMessage `json:"items"`
}

// UploadLinkRequest represents a request to generate upload URL
type UploadLinkRequest struct {
	FileName string `json:"fileName"`
	MimeType string `json:"mimeType,omitempty"`
}

// UploadLinkResponse represents the response from upload-link endpoint
type UploadLinkResponse struct {
	UploadURL string `json:"uploadUrl"`
	RawURL    string `json:"rawUrl"`
}

// FetchBlocks retrieves blocks from the document
func (c *Client) FetchBlocks(id string, maxDepth int, fetchMetadata bool) (*Block, error) {
	reqURL := fmt.Sprintf("%s/blocks", c.BaseURL)

	params := url.Values{}
	if id != "" {
		params.Add("id", id)
	}
	if maxDepth != -1 {
		params.Add("maxDepth", strconv.Itoa(maxDepth))
	}
	if fetchMetadata {
		params.Add("fetchMetadata", "true")
	}

	if len(params) > 0 {
		reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var block Block
	if err := json.NewDecoder(resp.Body).Decode(&block); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &block, nil
}

// FetchBlocksMarkdown retrieves blocks as markdown
func (c *Client) FetchBlocksMarkdown(id string, maxDepth int) (string, error) {
	reqURL := fmt.Sprintf("%s/blocks", c.BaseURL)

	params := url.Values{}
	if id != "" {
		params.Add("id", id)
	}
	if maxDepth != -1 {
		params.Add("maxDepth", strconv.Itoa(maxDepth))
	}

	if len(params) > 0 {
		reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "text/markdown")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	return string(body), nil
}

// InsertBlocks adds new blocks to the document
func (c *Client) InsertBlocks(req InsertRequest) ([]Block, error) {
	reqURL := fmt.Sprintf("%s/blocks", c.BaseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var itemsResp ItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	var blocks []Block
	if err := json.Unmarshal(itemsResp.Items, &blocks); err != nil {
		return nil, fmt.Errorf("unmarshaling blocks: %w", err)
	}

	return blocks, nil
}

// UpdateBlocks modifies existing blocks
func (c *Client) UpdateBlocks(req UpdateRequest) ([]Block, error) {
	reqURL := fmt.Sprintf("%s/blocks", c.BaseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("PUT", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var itemsResp ItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	var blocks []Block
	if err := json.Unmarshal(itemsResp.Items, &blocks); err != nil {
		return nil, fmt.Errorf("unmarshaling blocks: %w", err)
	}

	return blocks, nil
}

// DeleteBlocks removes blocks from the document
func (c *Client) DeleteBlocks(blockIDs []string) ([]string, error) {
	reqURL := fmt.Sprintf("%s/blocks", c.BaseURL)

	req := DeleteRequest{BlockIDs: blockIDs}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("DELETE", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != 207 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var itemsResp ItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	var deletedItems []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(itemsResp.Items, &deletedItems); err != nil {
		return nil, fmt.Errorf("unmarshaling deleted IDs: %w", err)
	}

	ids := make([]string, len(deletedItems))
	for i, item := range deletedItems {
		ids[i] = item.ID
	}

	return ids, nil
}

// MoveBlocks repositions blocks in the document
func (c *Client) MoveBlocks(req MoveRequest) ([]string, error) {
	reqURL := fmt.Sprintf("%s/blocks/move", c.BaseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("PUT", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != 207 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var itemsResp ItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	var movedItems []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(itemsResp.Items, &movedItems); err != nil {
		return nil, fmt.Errorf("unmarshaling moved IDs: %w", err)
	}

	ids := make([]string, len(movedItems))
	for i, item := range movedItems {
		ids[i] = item.ID
	}

	return ids, nil
}

// Search finds blocks matching a pattern
func (c *Client) Search(pattern string, caseSensitive bool, beforeCount, afterCount int) ([]SearchMatch, error) {
	reqURL := fmt.Sprintf("%s/blocks/search", c.BaseURL)

	params := url.Values{}
	params.Add("pattern", pattern)
	if caseSensitive {
		params.Add("caseSensitive", "true")
	}
	if beforeCount > 0 {
		params.Add("beforeBlockCount", strconv.Itoa(beforeCount))
	}
	if afterCount > 0 {
		params.Add("afterBlockCount", strconv.Itoa(afterCount))
	}

	reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var itemsResp ItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	var matches []SearchMatch
	if err := json.Unmarshal(itemsResp.Items, &matches); err != nil {
		return nil, fmt.Errorf("unmarshaling search results: %w", err)
	}

	return matches, nil
}

// GenerateUploadURL creates a pre-signed S3 URL for file upload
func (c *Client) GenerateUploadURL(fileName, mimeType string) (*UploadLinkResponse, error) {
	reqURL := fmt.Sprintf("%s/upload-link", c.BaseURL)

	req := UploadLinkRequest{
		FileName: fileName,
		MimeType: mimeType,
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var uploadResp UploadLinkResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &uploadResp, nil
}
