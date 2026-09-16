// Package ai talks to the Python AI service's HTTP API (ai/api/) for
// requests that need a real-time response - currently chat question
// answering. File processing does not go through here: the Go backend
// enqueues those jobs to SQS directly and the worker writes status
// straight to Postgres, both independent of this client.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type ChatSource struct {
	FileID     string `json:"file_id"`
	ChunkIndex int    `json:"chunk_index"`
}

type AnswerChatQuestionResult struct {
	Answer  string       `json:"answer"`
	Sources []ChatSource `json:"sources"`
}

// AnswerChatQuestion calls POST {baseURL}/api/v1/chat, which embeds the
// question, retrieves the project's most relevant document chunks from
// Qdrant, and asks the LLM to answer grounded in them.
func (c *Client) AnswerChatQuestion(
	ctx context.Context,
	projectID string,
	question string,
) (*AnswerChatQuestionResult, error) {
	body, err := json.Marshal(map[string]string{
		"project_id": projectID,
		"question":   question,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal chat question request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/api/v1/chat",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("build chat question request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call AI service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"AI service returned %d: %s",
			resp.StatusCode,
			string(respBody),
		)
	}

	var result AnswerChatQuestionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode chat question response: %w", err)
	}

	return &result, nil
}
