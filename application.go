package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/de1ux/deploy-tracker/integration"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type User struct {
	Name                   string `json:"name"`
	GithubHandle           string `yaml:"github_handle" json:"github_handle"`
	AwsAccessKey           string `yaml:"aws_access_key" json:"-"`
	AwsSecretAccessKeyId   string `yaml:"aws_secret_access_key_id" json:"-"`
	HasCloudfronts         bool   `json:"has_cloudfronts"`
	HasS3Buckets           bool   `json:"has_s3_buckets"`
	HasElasticBeanstalks   bool   `json:"has_elastic_beanstalks"`
	HasGitFrontendBlog     bool   `json:"has_git_frontend_blog"`
	HasGitBackendBlog      bool   `json:"has_git_backend_blog"`
	HasGitFrontendCapstone bool   `json:"has_git_frontend_capstone"`
	HasGitBackendCapstone  bool   `json:"has_git_backend_capstone"`
	HasRDS                 bool   `json:"has_rds"`
	Error                  string `json:"error"`
}

type UserCreds struct {
	Users []*User `json:"users"`
}

var mu sync.Mutex

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	b, err := os.ReadFile("creds.yaml")
	if err != nil {
		panic(err)
	}

	users := &UserCreds{}
	err = yaml.Unmarshal(b, users)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}
	if err != nil {
		panic(err)
	}

	go func(mu *sync.Mutex, users *UserCreds) {
		for {
			log.Printf("Updating users...")
			mu.Lock()

			wg := sync.WaitGroup{}
			for _, u := range users.Users {
				wg.Add(1)
				go func(user *User) {
					sess, err := session.NewSession(&aws.Config{
						Region:      aws.String("us-east-1"),
						Credentials: credentials.NewStaticCredentials(user.AwsAccessKey, user.AwsSecretAccessKeyId, ""),
					})
					if err != nil {
						user.Error = err.Error()
						log.Println(err.Error())
					}

					hasCloudFronts, err := integration.HasCloudfronts(sess)
					if err != nil {
						user.Error = err.Error()
						log.Println(err.Error())
					}

					hasS3Buckets, err := integration.HasS3Buckets(sess)
					if err != nil {
						user.Error = err.Error()
						log.Println(err.Error())
					}

					hasElasticBeanstalks, err := integration.HasElasticBeanstalks(sess)
					if err != nil {
						user.Error = err.Error()
						log.Println(err.Error())
					}

					hasRDS, err := integration.HasRDS(sess)
					if err != nil {
						user.Error = err.Error()
						log.Println(err.Error())
					}

					githubFrontendBlog, err := integration.DoesGithubRepoExist(user.GithubHandle, "blog-frontend")
					if err != nil {
						user.Error = err.Error()
						log.Println(err.Error())
					}

					githubBackendBlog, err := integration.DoesGithubRepoExist(user.GithubHandle, "blog-backend")
					if err != nil {
						user.Error = err.Error()
						log.Println(err.Error())
					}

					githubFrontendCapstone, err := integration.DoesGithubRepoExist(user.GithubHandle, "capstone-frontend")
					if err != nil {
						user.Error = err.Error()
						log.Println(err.Error())
					}

					githubBackendCapstone, err := integration.DoesGithubRepoExist(user.GithubHandle, "capstone-backend")
					if err != nil {
						user.Error = err.Error()
						log.Println(err.Error())
					}

					user.HasGitFrontendBlog = githubFrontendBlog
					user.HasGitBackendBlog = githubBackendBlog
					user.HasGitFrontendCapstone = githubFrontendCapstone
					user.HasGitBackendCapstone = githubBackendCapstone
					user.HasCloudfronts = hasCloudFronts
					user.HasS3Buckets = hasS3Buckets
					user.HasElasticBeanstalks = hasElasticBeanstalks
					user.HasRDS = hasRDS
					wg.Done()
				}(u)
			}
			wg.Wait()
			mu.Unlock()
			log.Printf("Updating users...done")
			log.Printf("Sleeping...")
			time.Sleep(time.Second * 30)
		}
	}(&mu, users)

	r := gin.Default()
	r.Use(CORSMiddleware())
	r.GET("/deploys", func(c *gin.Context) {
		mu.Lock()
		c.JSON(http.StatusOK, gin.H{
			"data": users,
		})
		mu.Unlock()
	})
	r.GET("/", func(c *gin.Context) {})

	port := "5000"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	println("Starting up server")
	if err := r.Run(":" + port); err != nil {
		panic(err)
	}
}
