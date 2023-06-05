package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
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

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
