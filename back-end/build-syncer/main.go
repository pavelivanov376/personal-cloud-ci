package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Build struct {
	ID            string `json:"id"`
	BuildNumber   int    `json:"buildNumber"`
	Timestamp     string `json:"timestamp"`
	Status        string `json:"status"`
	RepositoryUrl string `json:"repositoryUrl"`
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func fetchQueuedBuilds(baseURL string) ([]string, error) {
	resp, err := http.Get(baseURL + "/builds?status=QUEUED")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list queued: status %d: %s", resp.StatusCode, string(body))
	}
	var ids []string
	if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func fetchBuildSpecification(baseURL, id string) (*Build, error) {
	resp, err := http.Get(baseURL + "/builds/" + id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get build %s: status %d: %s", id, resp.StatusCode, string(body))
	}
	var b Build
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return nil, err
	}
	return &b, nil
}

func tick(baseURL string) {
	ids, err := fetchQueuedBuilds(baseURL)
	if err != nil {
		log.Printf("failed to list queued builds: %v", err)
		return
	}
	if len(ids) == 0 {
		log.Printf("no queued builds")
		return
	}
	log.Printf("found %d queued build(s)", len(ids))
	for _, id := range ids {
		b, err := fetchBuildSpecification(baseURL, id)
		if err != nil {
			log.Printf("failed to fetch build %s: %v", id, err)
			continue
		}
		log.Printf("build: id=%s number=%d status=%s repo=%s timestamp=%s",
			b.ID, b.BuildNumber, b.Status, b.RepositoryUrl, b.Timestamp)
	}
}

func main() {
	baseURL := getenv("GEAR_URL", "http://localhost:8081")
	interval, err := time.ParseDuration(getenv("POLL_INTERVAL", "10s"))
	if err != nil {
		log.Fatalf("invalid POLL_INTERVAL: %v", err)
	}

	log.Printf("build-syncer starting: gear=%s interval=%s", baseURL, interval)

	tick(baseURL)
	t := time.NewTicker(interval)
	defer t.Stop()
	for range t.C {
		tick(baseURL)
	}
}
