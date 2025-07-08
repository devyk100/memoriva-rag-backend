package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"memoriva-backend/models"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type LLMService struct {
	deepSeekClient *openai.Client
	openAIClient   *openai.Client
}

func NewLLMService(deepSeekAPIKey, openAIAPIKey string) *LLMService {
	var deepSeekClient *openai.Client
	var openAIClient *openai.Client

	if deepSeekAPIKey != "" {
		config := openai.DefaultConfig(deepSeekAPIKey)
		config.BaseURL = "https://api.deepseek.com/v1"
		deepSeekClient = openai.NewClientWithConfig(config)
	}

	if openAIAPIKey != "" {
		openAIClient = openai.NewClient(openAIAPIKey)
	}

	return &LLMService{
		deepSeekClient: deepSeekClient,
		openAIClient:   openAIClient,
	}
}

func (s *LLMService) AnalyzeCardsForStudy(cards []models.CardWithMetadata, prompt string, maxCards int) ([]string, error) {
	// Create a prompt for the LLM to analyze and select cards
	systemPrompt := `You are an intelligent flashcard study assistant. Your task is to analyze flashcard data and select the most appropriate cards for study based on the user's request.

You will receive:
1. A collection of flashcards with their front/back content
2. SRS metadata including review counts (easy, hard, again) and repetition data
3. A user prompt describing what they want to study
4. A maximum card limit (but you can select FEWER cards if the user's request is specific)

Your job is to:
1. Understand the user's study intent from their prompt
2. Find cards semantically relevant to the user's prompt (prioritize relevance over quantity)
3. Analyze card weakness based on SRS data (high again/hard counts = weak cards)
4. Select the most appropriate cards - if the user asks for specific topics, only select cards related to those topics
5. If only 2 cards match the user's specific request, return only those 2 cards (don't pad with unrelated cards)
6. You can repeat very weak cards multiple times in the selection

IMPORTANT: Quality over quantity - better to return 2 highly relevant cards than 20 loosely related ones.

Return only a JSON array of flashcard IDs in the order they should be studied.`

	// Build the user prompt with card data
	userPrompt := fmt.Sprintf(`User wants to study: "%s"
Maximum cards: %d

Available flashcards:
`, prompt, maxCards)

	for i, cardData := range cards {
		if i >= 50 { // Limit to prevent token overflow
			break
		}

		weaknessScore := 0.0
		if cardData.Metadata != nil {
			total := cardData.Metadata.EasyReviewCount + cardData.Metadata.HardReviewCount + cardData.Metadata.AgainReviewCount
			if total > 0 {
				weaknessScore = float64(cardData.Metadata.AgainReviewCount) + float64(cardData.Metadata.HardReviewCount)*0.5
				weaknessScore = weaknessScore / float64(total)
			}
		}

		userPrompt += fmt.Sprintf(`
ID: %s
Front: %s
Back: %s
Weakness Score: %.2f (0=strong, 1=very weak)
Reviews: Easy=%d, Hard=%d, Again=%d
`, cardData.Card.ID, cardData.Card.Front, cardData.Card.Back, weaknessScore,
			getReviewCount(cardData.Metadata, "easy"),
			getReviewCount(cardData.Metadata, "hard"),
			getReviewCount(cardData.Metadata, "again"))
	}

	userPrompt += "\nReturn a JSON array of selected flashcard IDs:"

	// Try DeepSeek first, fallback to OpenAI
	var model string
	var client *openai.Client

	if s.deepSeekClient != nil {
		client = s.deepSeekClient
		model = "deepseek-chat" // Use DeepSeek model
	} else if s.openAIClient != nil {
		client = s.openAIClient
		model = openai.GPT3Dot5Turbo // Use OpenAI model
	} else {
		return nil, fmt.Errorf("no LLM client available")
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			MaxTokens:   1000,
			Temperature: 0.3,
		},
	)

	if err != nil {
		log.Printf("LLM API error: %v", err)
		return s.fallbackCardSelection(cards, maxCards), nil
	}

	if len(resp.Choices) == 0 {
		log.Printf("No response from LLM")
		return s.fallbackCardSelection(cards, maxCards), nil
	}

	// Parse the response to extract card IDs
	responseContent := resp.Choices[0].Message.Content
	log.Printf("LLM Response: %s", responseContent)

	// Try to parse JSON response
	selectedIDs, err := s.parseCardIDsFromResponse(responseContent)
	if err != nil {
		log.Printf("Failed to parse LLM response, using fallback: %v", err)
		return s.fallbackCardSelection(cards, maxCards), nil
	}

	// Validate that all selected IDs exist in the available cards
	validIDs := s.validateCardIDs(selectedIDs, cards)
	if len(validIDs) == 0 {
		log.Printf("No valid card IDs found in LLM response, using fallback")
		return s.fallbackCardSelection(cards, maxCards), nil
	}

	log.Printf("LLM selected %d cards: %v", len(validIDs), validIDs)
	return validIDs, nil
}

func (s *LLMService) parseCardIDsFromResponse(response string) ([]string, error) {
	var cardIDs []string

	// Try to find JSON array in the response
	// Look for patterns like ["id1", "id2", "id3"]
	startIdx := strings.Index(response, "[")
	endIdx := strings.LastIndex(response, "]")

	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return nil, fmt.Errorf("no JSON array found in response")
	}

	jsonStr := response[startIdx : endIdx+1]

	// Try to parse as JSON
	err := json.Unmarshal([]byte(jsonStr), &cardIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return cardIDs, nil
}

func (s *LLMService) validateCardIDs(selectedIDs []string, availableCards []models.CardWithMetadata) []string {
	// Create a map of available card IDs for fast lookup
	availableMap := make(map[string]bool)
	for _, card := range availableCards {
		availableMap[card.Card.ID] = true
	}

	// Filter out invalid IDs
	var validIDs []string
	for _, id := range selectedIDs {
		if availableMap[id] {
			validIDs = append(validIDs, id)
		}
	}

	return validIDs
}

func (s *LLMService) fallbackCardSelection(cards []models.CardWithMetadata, maxCards int) []string {
	// Improved fallback: prioritize weak cards, but be flexible with count
	var selectedIDs []string

	// If we have very few cards, just return all of them
	if len(cards) <= 3 {
		for _, card := range cards {
			selectedIDs = append(selectedIDs, card.Card.ID)
		}
		return selectedIDs
	}

	// Sort by weakness score
	weakCards := make([]models.CardWithMetadata, 0)
	normalCards := make([]models.CardWithMetadata, 0)

	for _, card := range cards {
		if card.Metadata != nil {
			total := card.Metadata.EasyReviewCount + card.Metadata.HardReviewCount + card.Metadata.AgainReviewCount
			if total > 0 {
				weaknessScore := float64(card.Metadata.AgainReviewCount+card.Metadata.HardReviewCount) / float64(total)
				if weaknessScore > 0.3 { // Consider weak if >30% hard/again
					weakCards = append(weakCards, card)
				} else {
					normalCards = append(normalCards, card)
				}
			} else {
				normalCards = append(normalCards, card)
			}
		} else {
			normalCards = append(normalCards, card)
		}
	}

	// Calculate optimal card count (flexible based on available cards)
	optimalCount := maxCards
	if len(weakCards) > 0 && len(weakCards) < maxCards/2 {
		// If we have few weak cards, adjust the total count
		optimalCount = len(weakCards) + min(len(normalCards), maxCards/2)
	}

	// Select weak cards first
	for i, card := range weakCards {
		if len(selectedIDs) >= optimalCount {
			break
		}
		selectedIDs = append(selectedIDs, card.Card.ID)

		// Repeat very weak cards
		if card.Metadata != nil {
			total := card.Metadata.EasyReviewCount + card.Metadata.HardReviewCount + card.Metadata.AgainReviewCount
			if total > 0 {
				weaknessScore := float64(card.Metadata.AgainReviewCount) / float64(total)
				if weaknessScore > 0.5 && len(selectedIDs) < optimalCount {
					selectedIDs = append(selectedIDs, card.Card.ID) // Repeat very weak card
				}
			}
		}

		if i >= optimalCount/2 { // Don't use more than half slots for weak cards
			break
		}
	}

	// Fill remaining slots with normal cards
	for _, card := range normalCards {
		if len(selectedIDs) >= optimalCount {
			break
		}
		selectedIDs = append(selectedIDs, card.Card.ID)
	}

	return selectedIDs
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getReviewCount(metadata *models.SRSCardMetadata, reviewType string) int {
	if metadata == nil {
		return 0
	}

	switch reviewType {
	case "easy":
		return metadata.EasyReviewCount
	case "hard":
		return metadata.HardReviewCount
	case "again":
		return metadata.AgainReviewCount
	default:
		return 0
	}
}
