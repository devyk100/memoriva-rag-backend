package services

import (
	"context"
	"fmt"
	"io"
	"time"

	memoriva_config "memoriva-backend/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Service struct {
	client            *s3.Client
	bucketName        string
	cloudFrontBaseURL string
}

type PresignedUploadResponse struct {
	UploadURL   string `json:"uploadUrl"`
	ImageURL    string `json:"imageUrl"`
	Key         string `json:"key"`
	ContentType string `json:"contentType"`
}

func NewS3Service(cfg *memoriva_config.Config) (*S3Service, error) {
	// Create AWS config with static credentials
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.AWSRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID,
			cfg.AWSSecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsConfig)

	return &S3Service{
		client:            client,
		bucketName:        cfg.S3BucketName,
		cloudFrontBaseURL: cfg.CloudFrontBaseURL,
	}, nil
}

func (s *S3Service) GeneratePresignedUploadURL(contentType string) (*PresignedUploadResponse, error) {
	// Generate unique key for the image
	imageID := uuid.New().String()
	var extension string

	// Determine file extension based on content type
	switch contentType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/gif":
		extension = ".gif"
	case "image/webp":
		extension = ".webp"
	default:
		extension = ".jpg" // Default to jpg
	}

	key := fmt.Sprintf("images/%s%s", imageID, extension)

	// Create presigned URL for PUT operation
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		// Removed ACL as it's deprecated and can cause CORS issues
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(15 * time.Minute) // URL expires in 15 minutes
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// Generate the CloudFront URL for accessing the image
	imageURL := fmt.Sprintf("%s/%s", s.cloudFrontBaseURL, key)

	return &PresignedUploadResponse{
		UploadURL:   request.URL,
		ImageURL:    imageURL,
		Key:         key,
		ContentType: contentType,
	}, nil
}

func (s *S3Service) UploadFile(file io.Reader, contentType string) (string, error) {
	// Generate unique key for the image
	imageID := uuid.New().String()
	var extension string

	// Determine file extension based on content type
	switch contentType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/gif":
		extension = ".gif"
	case "image/webp":
		extension = ".webp"
	default:
		extension = ".jpg" // Default to jpg
	}

	key := fmt.Sprintf("images/%s%s", imageID, extension)

	// Upload file directly to S3
	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Generate the CloudFront URL for accessing the image
	imageURL := fmt.Sprintf("%s/%s", s.cloudFrontBaseURL, key)

	return imageURL, nil
}
