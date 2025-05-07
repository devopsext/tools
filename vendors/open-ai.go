package vendors

import (
	"context"
	"fmt"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIOptions struct {
	APIKey   string
	Timeout  int
	Model    string
	Messages []map[string]string
}

type OpenAI struct {
	options OpenAIOptions
	client  *openai.Client
}

func (ai *OpenAI) CreateChatCompletion(options OpenAIOptions) ([]byte, error) {
	timeout := options.Timeout
	if timeout == 0 {
		timeout = 30
	}

	// Set default model if not provided
	model := options.Model
	if model == "" {
		model = openai.GPT3Dot5Turbo // Default model
	}

	messages := []openai.ChatCompletionMessage{}
	for _, msg := range options.Messages {
		role, hasRole := msg["role"]
		content, hasContent := msg["content"]

		if hasRole && hasContent {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    role,
				Content: content,
			})
		}
	}

	// If no messages provided, add a default one
	if len(messages) == 0 {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: "Hello, how can I help you?",
		})
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	resp, err := ai.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: 0.2, // Lower temperature for more deterministic outputs
			TopP:        0.9,
			MaxTokens:   1000, // Increased token limit for complex report analysis
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to complete OpenAI chat: %w", err)
	}

	if len(resp.Choices) > 0 {
		response := resp.Choices[0].Message.Content
		return []byte(response), nil
	}

	return []byte("No response generated"), nil
}

func NewOpenAI(options OpenAIOptions) *OpenAI {
	ai := &OpenAI{
		options: options,
		client:  openai.NewClient(options.APIKey),
	}
	return ai
}
