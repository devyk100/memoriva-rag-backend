package services

import (
	"fmt"
	"log"
	"memoriva-backend/models"
)

type RAGService struct {
	dbService        *DatabaseService
	llmService       *LLMService
	embeddingService *EmbeddingService
}

func NewRAGService(dbService *DatabaseService, llmService *LLMService, embeddingService *EmbeddingService) *RAGService {
	return &RAGService{
		dbService:        dbService,
		llmService:       llmService,
		embeddingService: embeddingService,
	}
}

func (s *RAGService) ProcessStudySession(sessionID string) error {
	// Get study session details first to validate it exists
	session, err := s.dbService.GetStudySession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Update status to PROCESSING
	err = s.dbService.UpdateStudySessionStatus(sessionID, "PROCESSING")
	if err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	// Get deck cards with metadata
	cards, err := s.dbService.GetDeckCardsWithMetadata(session.DeckID, session.UserID)
	if err != nil {
		s.dbService.UpdateStudySessionStatus(sessionID, "FAILED")
		return fmt.Errorf("failed to get deck cards: %w", err)
	}

	if len(cards) == 0 {
		s.dbService.UpdateStudySessionStatus(sessionID, "FAILED")
		return fmt.Errorf("no cards found in deck")
	}

	// Use LLM to analyze and select cards
	selectedCardIDs, err := s.llmService.AnalyzeCardsForStudy(cards, session.Prompt, session.MaxCards)
	if err != nil {
		log.Printf("LLM analysis failed, using fallback: %v", err)
		// Use fallback selection if LLM fails
		selectedCardIDs = s.fallbackSelection(cards, session.MaxCards)
	}

	// Create study session cards
	err = s.dbService.CreateStudySessionCards(sessionID, selectedCardIDs)
	if err != nil {
		s.dbService.UpdateStudySessionStatus(sessionID, "FAILED")
		return fmt.Errorf("failed to create session cards: %w", err)
	}

	// Mark session as complete
	err = s.dbService.CompleteStudySession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to complete session: %w", err)
	}

	log.Printf("Successfully processed study session %s with %d cards", sessionID, len(selectedCardIDs))
	return nil
}

func (s *RAGService) fallbackSelection(cards []models.CardWithMetadata, maxCards int) []string {
	var selectedIDs []string

	// Simple fallback: prioritize cards with metadata (reviewed cards) and weak cards
	reviewedCards := make([]models.CardWithMetadata, 0)
	newCards := make([]models.CardWithMetadata, 0)

	for _, card := range cards {
		if card.Metadata != nil {
			reviewedCards = append(reviewedCards, card)
		} else {
			newCards = append(newCards, card)
		}
	}

	// Add reviewed cards first (prioritizing weak ones)
	for _, card := range reviewedCards {
		if len(selectedIDs) >= maxCards {
			break
		}

		selectedIDs = append(selectedIDs, card.Card.ID)

		// Check if this is a very weak card (high again count)
		if card.Metadata != nil {
			total := card.Metadata.EasyReviewCount + card.Metadata.HardReviewCount + card.Metadata.AgainReviewCount
			if total > 0 {
				againRatio := float64(card.Metadata.AgainReviewCount) / float64(total)
				if againRatio > 0.5 && len(selectedIDs) < maxCards {
					// Repeat very weak cards
					selectedIDs = append(selectedIDs, card.Card.ID)
				}
			}
		}
	}

	// Fill remaining slots with new cards
	for _, card := range newCards {
		if len(selectedIDs) >= maxCards {
			break
		}
		selectedIDs = append(selectedIDs, card.Card.ID)
	}

	return selectedIDs
}
