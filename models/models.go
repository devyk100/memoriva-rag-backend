package models

import (
	"time"
)

// Database models matching the Prisma schema
type User struct {
	ID       string  `gorm:"primaryKey;column:id"`
	Name     *string `gorm:"column:name"`
	Email    string  `gorm:"column:email;unique"`
	AuthType string  `gorm:"column:authType"`
	Image    *string `gorm:"column:image"`
}

type FlashcardDeck struct {
	ID         string      `gorm:"primaryKey;column:id"`
	Name       string      `gorm:"column:name"`
	Flashcards []Flashcard `gorm:"foreignKey:DeckID"`
}

type Flashcard struct {
	ID     string        `gorm:"primaryKey;column:id"`
	Front  string        `gorm:"column:front"`
	Back   string        `gorm:"column:back"`
	DeckID string        `gorm:"column:deckId"`
	Deck   FlashcardDeck `gorm:"foreignKey:DeckID"`
}

type SRSCardMetadata struct {
	ID               string     `gorm:"primaryKey;column:id"`
	UserID           string     `gorm:"column:userId"`
	FlashcardID      string     `gorm:"column:flashcardId"`
	EaseFactor       float64    `gorm:"column:easeFactor;default:1.3"`
	Interval         int64      `gorm:"column:interval;default:1"`
	Repetitions      int        `gorm:"column:repetitions;default:-1"`
	LastReviewed     *time.Time `gorm:"column:lastReviewed"`
	NextReview       *time.Time `gorm:"column:nextReview"`
	EasyReviewCount  int        `gorm:"column:easyReviewCount;default:0"`
	HardReviewCount  int        `gorm:"column:hardReviewCount;default:0"`
	AgainReviewCount int        `gorm:"column:againReviewCount;default:0"`
	User             User       `gorm:"foreignKey:UserID"`
	Flashcard        Flashcard  `gorm:"foreignKey:FlashcardID"`
}

type StudySession struct {
	ID          string             `gorm:"primaryKey;column:id"`
	UserID      string             `gorm:"column:userId"`
	DeckID      string             `gorm:"column:deckId"`
	Prompt      string             `gorm:"column:prompt"`
	MaxCards    int                `gorm:"column:maxCards"`
	Status      string             `gorm:"column:status;default:PENDING"`
	CreatedAt   time.Time          `gorm:"column:createdAt;default:CURRENT_TIMESTAMP"`
	CompletedAt *time.Time         `gorm:"column:completedAt"`
	User        User               `gorm:"foreignKey:UserID"`
	Deck        FlashcardDeck      `gorm:"foreignKey:DeckID"`
	Cards       []StudySessionCard `gorm:"foreignKey:StudySessionID"`
}

func (StudySession) TableName() string {
	return "StudySession"
}

type StudySessionCard struct {
	ID             string       `gorm:"primaryKey;column:id"`
	StudySessionID string       `gorm:"column:studySessionId"`
	FlashcardID    string       `gorm:"column:flashcardId"`
	Order          int          `gorm:"column:order"`
	StudySession   StudySession `gorm:"foreignKey:StudySessionID"`
	Flashcard      Flashcard    `gorm:"foreignKey:FlashcardID"`
}

func (StudySessionCard) TableName() string {
	return "StudySessionCard"
}

func (User) TableName() string {
	return "User"
}

func (FlashcardDeck) TableName() string {
	return "FlashcardDeck"
}

func (Flashcard) TableName() string {
	return "Flashcard"
}

func (SRSCardMetadata) TableName() string {
	return "SRSCardMetadata"
}

// API request/response models
type ProcessStudySessionRequest struct {
	SessionID string `json:"sessionId" binding:"required"`
}

type StudySessionStatusResponse struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	CardCount   int        `json:"cardCount"`
	CompletedAt *time.Time `json:"completedAt"`
}

// RAG processing models
type CardWithMetadata struct {
	Card     Flashcard
	Metadata *SRSCardMetadata
}

type CardScore struct {
	Card          Flashcard
	WeaknessScore float64
	SemanticScore float64
	CombinedScore float64
}

type RAGResult struct {
	SelectedCards []CardScore
	TotalCards    int
	WeakCards     int
	SemanticCards int
}
