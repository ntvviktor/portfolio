package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/viktor/go-app/components"
	"github.com/viktor/go-app/services"
)

var ginLambda *ginadapter.GinLambdaV2

type ServerConfig struct {
	port string
	S3Client *s3.Client
}

var r = gin.Default()
func init() {
	gin.SetMode(gin.ReleaseMode)
	godotenv.Load()
	// Connect to s3 client
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	client := s3.NewFromConfig(cfg)
	serverCfg := ServerConfig {
		port: ":8080",
		S3Client: client,
	}

	r.HTMLRender = &services.TemplRender{}
	r.Static("/assets", "./assets")
	r.GET("/", func(c* gin.Context) {
		c.HTML(http.StatusOK, "", components.IndexPage())
	})

	r.GET("/blog", serverCfg.GetAllBlog)
	
	r.GET("/blog/:id", serverCfg.GetBlogByID)
	ginLambda = ginadapter.NewV2(r)
}

func main() {
	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		lambda.Start(Handler)
	} else {
		r.Run(":8080")
	}
}

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}