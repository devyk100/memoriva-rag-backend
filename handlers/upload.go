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

func (h *UploadHandler) UploadToS3(c *gin.Context) {
	// Parse multipart form
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}
	defer file.Close()

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Try to detect from filename extension
		filename := header.Filename
		if len(filename) > 4 {
			ext := filename[len(filename)-4:]
			switch ext {
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			case ".png":
				contentType = "image/png"
			case ".gif":
				contentType = "image/gif"
			case ".webp":
				contentType = "image/webp"
			}
		}
	}

	validContentTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !validContentTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content type. Only image files are allowed."})
		return
	}

	// Upload to S3
	imageURL, err := h.s3Service.UploadFile(file, contentType)
	if err != nil {
		log.Printf("Failed to upload to S3: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"imageUrl": imageURL})
}
