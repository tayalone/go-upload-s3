package bucket

import (
	"fmt"
	"log"
	"mime/multipart"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// UploadResp return from Upload File to bucket
type UploadResp struct {
	IsError bool
	Err     string
	Key     string
	Url     string
}

// Bucket for Handler Static File
type Bucket interface {
	Healtz() error
	Upload(file *multipart.FileHeader, prefix string) (UploadResp, error)
	FileExist(key string) error
	Remove(ket string) error
}

// Domain of Bucket
type Domain struct {
	region          string
	accessKeyID     string
	secretAccessKey string
	bucketName      string
	session         *session.Session // awsSession
	client          *s3.S3           // ses3Client
}

// Initialize Bucket
func Initialize(
	region string,
	accessKeyID string,
	secretAccessKey string,
	bucketName string,
) Bucket {
	awsSession, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		log.Fatal(err)
	}
	s3Client := s3.New(awsSession)

	return &Domain{
		region:          region,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		bucketName:      bucketName,
		session:         awsSession,
		client:          s3Client,
	}
}

// getFullUrl of file
func (d *Domain) getFullURL(key string) string {
	return fmt.Sprintf("https://%s.s3-%s.amazonaws.com/%s", d.bucketName, d.region, key)
}

// Healtz of bucket
func (d *Domain) Healtz() error {
	_, err := d.client.ListBuckets(nil)
	if err != nil {
		return err
	}
	return nil
}

// Upload of file to bucket
func (d *Domain) Upload(file *multipart.FileHeader, prefix string) (UploadResp, error) {
	src, err := file.Open()
	if err != nil {
		return UploadResp{
			IsError: true,
			Err:     err.Error(),
		}, err
	}
	defer src.Close()
	key := fmt.Sprintf("%s%s", prefix, file.Filename) // Add sub-folder path to the object key

	params := &s3.PutObjectInput{
		Bucket: aws.String(d.bucketName),
		Key:    aws.String(key),
		Body:   src,
	}
	_, err = d.client.PutObject(params)
	if err != nil {
		return UploadResp{
			IsError: true,
			Err:     err.Error(),
		}, err
	}

	return UploadResp{
		IsError: false,
		Err:     "",
		Key:     key,
		Url:     d.getFullURL(key),
	}, nil
}

// FileExist of file to bucket
func (d *Domain) FileExist(key string) error {

	// Check if the file exists in S3
	headParams := &s3.HeadObjectInput{
		Bucket: aws.String(d.bucketName),
		Key:    aws.String(key),
	}
	_, err := d.client.HeadObject(headParams)
	if err != nil {
		return err
	}
	return nil
}

// Remove of file to bucket
func (d *Domain) Remove(key string) error {

	err := d.FileExist(key)
	if err != nil {
		return err
	}

	// Remove the file from S3
	deleteParams := &s3.DeleteObjectInput{
		Bucket: aws.String(d.bucketName),
		Key:    aws.String(key),
	}
	_, err = d.client.DeleteObject(deleteParams)
	if err != nil {
		return err
	}

	return nil
}
