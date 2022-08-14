package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

var mu sync.Mutex

var githubUsers = []string{
	"de1ux",
	"NicoleeeBroadnax",
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
	if err := r.Run(); err != nil {
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

func doesUrl200(url string) bool {
	resp, err := getHttpClient().Get(url)
	if err != nil {
		log.Printf("failed check url 200 request: %s\n", err)
		return false
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Got %d for %s", resp.StatusCode, url)
	}
	return resp.StatusCode == http.StatusOK
}

func doesGithubRepoExist(githubUsername, repoName string) bool {
	return doesUrl200(fmt.Sprintf("https://github.com/%s/%s", githubUsername, repoName))
}

func doesHerokuDeployExist(githubUsername, deployName string) bool {
	return doesUrl200(fmt.Sprintf("https://%s-%s.herokuapp.com", githubUsername, deployName))
}
