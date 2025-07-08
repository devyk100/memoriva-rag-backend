package services

import (
	"memoriva-backend/models"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func generateUUID() string {
	return uuid.New().String()
}

type DatabaseService struct {
	db *gorm.DB
}

func InitDatabase(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewDatabaseService(db *gorm.DB) *DatabaseService {
	return &DatabaseService{db: db}
}

func (s *DatabaseService) GetStudySession(sessionID string) (*models.StudySession, error) {
	var session models.StudySession
	err := s.db.Preload("User").Preload("Deck").First(&session, "id = ?", sessionID).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *DatabaseService) GetDeckCardsWithMetadata(deckID, userID string) ([]models.CardWithMetadata, error) {
	var cards []models.Flashcard
	err := s.db.Where("\"deckId\" = ?", deckID).Find(&cards).Error
	if err != nil {
		return nil, err
	}

	var result []models.CardWithMetadata
	for _, card := range cards {
		var metadata models.SRSCardMetadata
		err := s.db.Where("\"flashcardId\" = ? AND \"userId\" = ?", card.ID, userID).First(&metadata).Error

		cardWithMetadata := models.CardWithMetadata{
			Card: card,
		}

		if err == nil {
			cardWithMetadata.Metadata = &metadata
		}

		result = append(result, cardWithMetadata)
	}

	return result, nil
}

func (s *DatabaseService) UpdateStudySessionStatus(sessionID, status string) error {
	return s.db.Model(&models.StudySession{}).Where("id = ?", sessionID).Update("status", status).Error
}

func (s *DatabaseService) CreateStudySessionCards(sessionID string, cardIDs []string) error {
	// First, delete any existing cards for this session to avoid duplicates
	err := s.db.Where("\"studySessionId\" = ?", sessionID).Delete(&models.StudySessionCard{}).Error
	if err != nil {
		return err
	}

	var studySessionCards []models.StudySessionCard

	for i, cardID := range cardIDs {
		studySessionCards = append(studySessionCards, models.StudySessionCard{
			ID:             generateUUID(),
			StudySessionID: sessionID,
			FlashcardID:    cardID,
			Order:          i + 1,
		})
	}

	return s.db.Create(&studySessionCards).Error
}

func (s *DatabaseService) CompleteStudySession(sessionID string) error {
	return s.db.Model(&models.StudySession{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
		"status":      "READY",
		"completedAt": "NOW()",
	}).Error
}

func (s *DatabaseService) GetStudySessionStatus(sessionID string) (*models.StudySessionStatusResponse, error) {
	var session models.StudySession
	err := s.db.Preload("Cards").First(&session, "id = ?", sessionID).Error
	if err != nil {
		return nil, err
	}

	return &models.StudySessionStatusResponse{
		ID:          session.ID,
		Status:      session.Status,
		CardCount:   len(session.Cards),
		CompletedAt: session.CompletedAt,
	}, nil
}
