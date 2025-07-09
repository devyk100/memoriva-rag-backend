package handlers

import (
	"memoriva-backend/models"
	"memoriva-backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StudyHandler struct {
	queueService *services.QueueService
	dbService    *services.DatabaseService
}

func NewStudyHandler(queueService *services.QueueService, dbService *services.DatabaseService) *StudyHandler {
	return &StudyHandler{
		queueService: queueService,
		dbService:    dbService,
	}
}

func (h *StudyHandler) ProcessStudySession(c *gin.Context) {
	var req models.ProcessStudySessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that the study session exists before enqueuing
	_, err := h.dbService.GetStudySession(req.SessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Study session not found",
		})
		return
	}

	// Enqueue the study session for processing
	if err := h.queueService.EnqueueStudySession(req.SessionID); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Queue is full, please try again later",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":   "Study session processing started",
		"sessionId": req.SessionID,
	})
}

func (h *StudyHandler) GetStudySessionStatus(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID is required"})
		return
	}

	// Get the actual session status from database
	session, err := h.dbService.GetStudySession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     sessionID,
		"status": session.Status,
	})
}
