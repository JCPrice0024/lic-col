package lic

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Repo is a struct used to get the License from the githubapi
type Repo struct {
	Limit     int
	Remaining int
	Reset     time.Time
	License   License `json:"license"`
}

// MinimumRem is the minimum amount of remaining requests before the program no longer requests from the githubapi
const MinimumRem = 400

// License is a struct used to get only the License name from the githubapi
type License struct {
	Name string `json:"name"`
}

// GetRepoInfo makes a http.Request to the github api and gets a License name, if available,
// from the repo currently being scanned. This can only be used if the user provides a
// git-token and git-username.
func GetRepoInfo(owner, repoName, username, token string) (Repo, error) {
	repo := Repo{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repoName), nil)
	if err != nil {
		return repo, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	if len(token) > 0 && len(username) > 0 {
		req.SetBasicAuth(username, token)
	}
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return repo, err
	}
	if isBadResp(resp) {
		resp.Body.Close()
		return repo, fmt.Errorf("received invalid status code: %v", resp.StatusCode)
	}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&repo)

	limit := resp.Header.Values("X-RateLimit-Limit")
	remaining := resp.Header.Values("X-RateLimit-Remaining")
	reset := resp.Header.Values("X-RateLimit-Reset")

	if len(limit) > 0 && len(remaining) > 0 && len(reset) > 0 {
		repo.Limit = int(parseIntLogging("X-RateLimit-Limit", limit[0]))

		repo.Remaining = int(parseIntLogging("X-RateLimit-Remaining", remaining[0]))

		tmpReset := parseIntLogging("X-RateLimit-Reset", reset[0])

		repo.Reset = time.Unix(tmpReset, int64(0))
		if repo.Reset.IsZero() {
			log.Println("unable to convert X-RateLimit-Reset to time", reset[0])
		}
	}

	resp.Body.Close()
	return repo, err
}

func isBadResp(resp *http.Response) bool {
	return resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest
}

func parseIntLogging(field, input string) int64 {
	tmpReset, err := strconv.ParseInt(input, 0, 64)
	if err != nil {
		log.Println("unable to convert", field, "to int", input)
	}
	return tmpReset
}

// CalcGitApiSleep is a function that prevents overloading the gitapi with too many requests
// in an hour. DO NOT REMOVE THIS FUNCTION!
func (r *Repo) CalcGitApiSleep() (nearLimit bool) {
	if r.Remaining > r.Limit/2 {
		time.Sleep(time.Millisecond * 50)
		return false
	}
	if r.Remaining < MinimumRem {
		log.Printf("TOKEN REQUEST NEARING LIMIT STOPPING GITHUB API CALL TRY AGAIN AT: %v", r.Reset)
		return true
	}
	time.Sleep(time.Second * 2)
	return false
}
