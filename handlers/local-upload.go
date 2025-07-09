package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LocalUploadHandler struct {
	uploadDir string
	baseURL   string
}

func NewLocalUploadHandler() *LocalUploadHandler {
	// Create uploads directory if it doesn't exist
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create upload directory: %v", err))
	}

	return &LocalUploadHandler{
		uploadDir: uploadDir,
		baseURL:   "http://localhost:8080",
	}
}

func (h *LocalUploadHandler) UploadImage(c *gin.Context) {
	// Parse multipart form
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only images are allowed."})
		return
	}

	// Generate unique filename
	ext := getFileExtension(contentType)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filepath := filepath.Join(h.uploadDir, filename)

	// Create the file
	dst, err := os.Create(filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Return the URL
	imageURL := fmt.Sprintf("%s/uploads/%s", h.baseURL, filename)
	c.JSON(http.StatusOK, gin.H{
		"imageUrl": imageURL,
		"filename": filename,
	})
}

func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

func getFileExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}
