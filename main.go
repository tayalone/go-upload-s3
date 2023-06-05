package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

func main() {
	awsSession, err := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_S3_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_S3_ACCESS_KEY_ID"), os.Getenv("AWS_S3_SECRET_ACCESS_KEY"), ""),
	})
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.New(awsSession)

	fmt.Println("Start")

	r := gin.Default()

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	// set new 8 MiB
	// r.MaxMultipartMemory = 8 << 20  // 8 MiB

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	/* Upload single file */
	r.POST("/single", func(c *gin.Context) {
		// single file
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, "Bad request")
			return
		}
		log.Println(file.Filename)

		src, err := file.Open()
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to open file")
			return
		}
		defer src.Close()

		/* do some logic */

		/*  --------- */
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	/* ------------------- */
	/* Upload Multi files */
	r.POST("/multi", func(c *gin.Context) {
		form, _ := c.MultipartForm()
		defer form.RemoveAll()
		files := form.File["upload[]"]
		for _, file := range files {
			src, err := file.Open()
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to open file")
				return
			}
			defer src.Close()
		}
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	/* ------------------- */

	r.POST("/single-s3", func(c *gin.Context) {
		// single file
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, "Bad request")
			return
		}
		log.Println(file.Filename)

		src, err := file.Open()
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to open file")
			return
		}
		defer src.Close()

		/* do some logic */

		key := fmt.Sprintf("test-go/%s", file.Filename) // Add sub-folder path to the object key

		params := &s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("AWS_S3_BUCKET_NAME")),
			Key:    aws.String(key),
			Body:   src,
		}

		_, err = s3Client.PutObject(params)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to upload file to S3")
			return
		}

		// Check if the file exists in S3
		headParams := &s3.HeadObjectInput{
			Bucket: aws.String(os.Getenv("AWS_S3_BUCKET_NAME")),
			Key:    aws.String(key),
		}

		_, err = s3Client.HeadObject(headParams)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
				c.String(http.StatusOK, fmt.Sprintf("'%s' does not exist in S3", file.Filename))
				return
			}
			c.String(http.StatusInternalServerError, "Failed to check file existence in S3")
			return
		}

		url := fmt.Sprintf("https://%s.s3-%s.amazonaws.com/%s", os.Getenv("AWS_S3_BUCKET_NAME"), os.Getenv("AWS_S3_REGION"), key)
		/*  --------- */

		// Remove the file from S3
		deleteParams := &s3.DeleteObjectInput{
			Bucket: aws.String(os.Getenv("AWS_S3_BUCKET_NAME")),
			Key:    aws.String(key),
		}

		_, err = s3Client.DeleteObject(deleteParams)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to remove file from S3")
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "OK", "url": url})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
