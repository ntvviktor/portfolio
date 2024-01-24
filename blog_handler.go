package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/viktor/go-app/components"
)

var listOfPosts = make(map[string]components.Post)

func (cfg *ServerConfig) GetAllBlog(c *gin.Context) {
	var s3BucketName = os.Getenv("S3_BUCKET_NAME")

	var posts []components.Post

	output, err := cfg.S3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(s3BucketName),
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, object := range output.Contents {
		_, month, day := object.LastModified.Date()
		date := fmt.Sprintf("%d %s", day, month.String())
		id, title := splitIDTitle(aws.ToString(object.Key))
		post := components.Post {
			ID: id,
			Date: date,
			Title: title,
		}
		posts = append(posts, post)
		_, ok := listOfPosts[id]
		if !ok {
			listOfPosts[id] = post
		}
	}
	c.HTML(http.StatusOK, "", components.BlogPage(posts))
}

func splitIDTitle(name string) (id string, title string) {
	s := strings.Split(name, "_")
	s1 := strings.TrimRight(s[1], ".html") 
	return s[0], s1
}

func concatString(id string, title string) (objectKey string) {
	return id + "_" + title + ".html";
}

func (cfg *ServerConfig) GetBlogByID(c *gin.Context) {
	var s3BucketName = os.Getenv("S3_BUCKET_NAME")
	var msg struct {
		Message string
	}
	msg.Message = "Content Not Found"
	id := c.Param("id")
	post, ok := listOfPosts[id]
	if !ok {
		c.HTML(http.StatusNotFound, "", msg)
		panic("Not Found")
	}
	objectKey := concatString(post.ID, post.Title)
	content, err := GetBlogFromS3(cfg.S3Client, s3BucketName, objectKey)
	if err != nil {
		c.HTML(http.StatusNotFound, "", msg)
	}
	c.HTML(http.StatusOK, "", components.ContentPage(post.Title, Unsafe(content)))
}

func Unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, html)
		return
	})
}

func GetBlogFromS3(s3Client *s3.Client, bucketName string, objectKey string) (string, error){
	result, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		log.Printf("Couldn't get object %v:%v. Here's why: %v\n", bucketName, objectKey, err)
		return "", err
	}
	defer result.Body.Close()
	
	body, err := io.ReadAll(result.Body)
	if err != nil {
		log.Printf("Couldn't read object body from %v. Here's why: %v\n", objectKey, err)
		return "", err
	}
	return string(body), nil
}