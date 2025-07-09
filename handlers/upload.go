package handlers

import (
	"log"
	"net/http"

	"memoriva-backend/services"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	s3Service *services.S3Service
}

func NewUploadHandler(s3Service *services.S3Service) *UploadHandler {
	return &UploadHandler{
		s3Service: s3Service,
	}
}

type GeneratePresignedURLRequest struct {
	ContentType string `json:"contentType" binding:"required"`
}

func (h *UploadHandler) GeneratePresignedURL(c *gin.Context) {
	var req GeneratePresignedURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate content type
	validContentTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !validContentTypes[req.ContentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content type. Only image files are allowed."})
		return
	}

	response, err := h.s3Service.GeneratePresignedUploadURL(req.ContentType)
	if err != nil {
		log.Printf("Failed to generate presigned URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate upload URL"})
		return
	}

	c.JSON(http.StatusOK, response)
}
