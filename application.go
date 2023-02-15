package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var mu sync.Mutex

var githubUsers = []string{
	"de1ux",
	"NicoleBroadnax",
	"lgoodman320",
	"ChristiHarlow",
	"dougMR",
	"WayneStB",
	"UniqueCre8tion85",
	"Sccotyiab",
	"chazdickerson1428",
	"gsnunez",
	"zhafner",
	"shantinaperez",
	"SMaeweather",
	"Itismywinningseason",
}

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
	var deploys []map[string]interface{}

	go func(mu *sync.Mutex) {
		for {
			log.Printf("Fetching deploys...")
			newDeploys := getDeploys()
			log.Printf("Fetching deploys...done")

			log.Printf("Acquiring lock to overwrite results...")
			mu.Lock()
			deploys = newDeploys
			mu.Unlock()
			log.Printf("Acquiring lock to overwrite results...done")
			log.Printf("Sleeping...")
			time.Sleep(time.Second * 30)
		}
	}(&mu)

	r := gin.Default()
	r.Use(CORSMiddleware())
	r.GET("/deploys", func(c *gin.Context) {
		mu.Lock()
		c.JSON(http.StatusOK, gin.H{
			"data": deploys,
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

func getDeploys() []map[string]interface{} {
	results := make(chan map[string]interface{}, len(githubUsers))

	wg := sync.WaitGroup{}
	for _, user := range githubUsers {
		wg.Add(1)
		go func(u string) {
			results <- getDeploysByUsername(u)
			wg.Done()
		}(user)
	}
	wg.Wait()
	close(results)

	var deploys []map[string]interface{}
	for result := range results {
		deploys = append(deploys, result)
	}

	sort.Slice(deploys, func(i, j int) bool {
		return strings.ToLower(deploys[i]["username"].(string)) < strings.ToLower(deploys[j]["username"].(string))
	})

	return deploys
}

func getDeploysByUsername(username string) map[string]interface{} {
	return map[string]interface{}{
		"username":                 username,
		"git_frontend_blog":        doesGithubRepoExist(username, "blog-frontend"),
		"git_backend_blog":         doesGithubRepoExist(username, "blog-backend"),
		"heroku_frontend_blog":     doesHerokuDeployExist(username, "blog-frontend"),
		"heroku_backend_blog":      doesHerokuDeployExist(username, "blog-backend"),
		"git_frontend_capstone":    doesGithubRepoExist(username, "capstone-frontend"),
		"git_backend_capstone":     doesGithubRepoExist(username, "capstone-backend"),
		"heroku_frontend_capstone": doesHerokuDeployExist(username, "capstone-frontend"),
		"heroku_backend_capstone":  doesHerokuDeployExist(username, "capstone-backend"),
	}

}

func getHttpClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
	}
}

func doesHttpStatus(url string, status int) (bool, error) {
	resp, err := getHttpClient().Get(url)
	if err != nil {
		log.Printf("failed check for status %d, got err: %s\n", status, err)
		return false, err
	}
	if resp.StatusCode != status {
		log.Printf("Got %d while looking for %d on url %s", resp.StatusCode, status, url)
	}
	return resp.StatusCode == status, nil
}

func doesUrl404(url string) (bool, error) {
	return doesHttpStatus(url, http.StatusNotFound)
}

func doesUrl200(url string) (bool, error) {
	return doesHttpStatus(url, http.StatusOK)
}

func doesGithubRepoExist(githubUsername, repoName string) bool {
	does200, _ := doesUrl200(fmt.Sprintf("https://github.com/%s/%s", githubUsername, repoName))
	return does200
}

func doesHerokuDeployExist(githubUsername, deployName string) bool {
	url := fmt.Sprintf("https://%s-%s.herokuapp.com", githubUsername, deployName)
	if strings.Contains(deployName, "backend") {
		did404, err := doesUrl404(url)
		if err != nil {
			return false
		}
		return !did404
	}
	does200, _ := doesUrl200(url)
	return does200
}
