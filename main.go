package main

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/tayalone/go-upload-s3/bucket"
)

func main() {

	fmt.Println("Start")

	myBucket := bucket.Initialize(
		os.Getenv("AWS_S3_REGION"),
		os.Getenv("AWS_S3_ACCESS_KEY_ID"),
		os.Getenv("AWS_S3_SECRET_ACCESS_KEY"),
		os.Getenv("AWS_S3_BUCKET_NAME"),
	)

	err := myBucket.Healtz()
	if err != nil {
		fmt.Println("bucket err", err.Error())
	} else {
		fmt.Println("bucket ok!!")
	}

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

	/* single upload file s3 */
	r.POST("/single-s3", func(c *gin.Context) {
		// single file
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, "Bad request")
			return
		}

		resp, err := myBucket.Upload(file, "test-go/")
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		fmt.Println("upload to s3 success")

		err = myBucket.FileExist(resp.Key)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		fmt.Println("Check File Exist  s3 success")

		err = myBucket.Remove(resp.Key)

		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		fmt.Println("Remove  s3 success")

		c.JSON(http.StatusOK, gin.H{"message": "OK", "url": resp.Url})
	})
	/* -------------- */
	r.POST("/multi-s3", func(c *gin.Context) {
		form, _ := c.MultipartForm()
		defer form.RemoveAll()
		files := form.File["upload[]"]
		lists := make([]string, 0, len(files))

		var wg sync.WaitGroup
		wg.Add(len(files))

		for _, file := range files {
			go func(file *multipart.FileHeader) {
				defer wg.Done()

				// Upload each file to S3
				resp, err := myBucket.Upload(file, "test-go/")
				if err != nil {
					lists = append(lists, fmt.Sprintf("Failed to upload file: %s", file.Filename))
				} else {
					lists = append(lists, resp.Url)
				}
			}(file)
		}
		wg.Wait()

		c.JSON(http.StatusOK, gin.H{
			"message": "OK",
			"lists":   lists,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
