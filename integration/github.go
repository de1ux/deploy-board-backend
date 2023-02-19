package integration

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

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

func DoesGithubRepoExist(githubUsername, repoName string) (bool, error) {
	return doesUrl200(fmt.Sprintf("https://github.com/%s/%s", githubUsername, repoName))
}
