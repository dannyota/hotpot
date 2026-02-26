package job

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client wraps the Jenkins REST API.
type Client struct {
	baseURL        string
	username       string
	apiToken       string
	httpClient     *http.Client
	excludeRepos   []string
	maxBuildsPerJob int
}

// NewClient creates a new Jenkins API client.
func NewClient(baseURL, username, apiToken string, maxBuildsPerJob int, excludeRepos []string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:        strings.TrimRight(baseURL, "/"),
		username:       username,
		apiToken:       apiToken,
		httpClient:     httpClient,
		excludeRepos:   excludeRepos,
		maxBuildsPerJob: maxBuildsPerJob,
	}
}

// APIJobSummary is a job from the list/tree API (minimal fields).
type APIJobSummary struct {
	Name      string           `json:"name"`
	URL       string           `json:"url"`
	Class     string           `json:"_class"`
	LastBuild *APIBuildSummary `json:"lastBuild"`
	Jobs      []APIJobSummary  `json:"jobs"` // sub-jobs for folders
}

// APIBuildSummary is minimal build info from the job list.
type APIBuildSummary struct {
	Number    int   `json:"number"`
	Timestamp int64 `json:"timestamp"` // millis since epoch
}

// APIJobDetail is the full job detail response.
type APIJobDetail struct {
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Class     string            `json:"_class"`
	Buildable bool              `json:"buildable"`
	Builds    []APIBuildSummary `json:"builds"`
}

// APIBuildDetail is the full build detail response.
type APIBuildDetail struct {
	Number    int              `json:"number"`
	Result    string           `json:"result"`
	Timestamp int64            `json:"timestamp"` // millis since epoch
	Duration  int64            `json:"duration"`   // millis
	Actions   []json.RawMessage `json:"actions"`
}

// APIBuildAction represents a parsed build action.
type APIBuildAction struct {
	Class      string              `json:"_class"`
	Parameters []APIBuildParameter `json:"parameters"`
	BuildsByBranchName map[string]APIBranchBuild `json:"buildsByBranchName"`
	RemoteURLs []string `json:"remoteUrls"`
	LastBuiltRevision *APIRevision `json:"lastBuiltRevision"`
}

// APIBuildParameter is a build parameter.
type APIBuildParameter struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

// APIBranchBuild maps branch name to build info.
type APIBranchBuild struct {
	BuildNumber int          `json:"buildNumber"`
	Revision    APIRevision  `json:"revision"`
}

// APIRevision holds git revision info.
type APIRevision struct {
	SHA1    string       `json:"SHA1"`
	Branch  []APIBranch  `json:"branch"`
}

// APIBranch holds branch info from git plugin.
type APIBranch struct {
	SHA1 string `json:"SHA1"`
	Name string `json:"name"`
}

// folderClasses are Jenkins job classes that represent folders.
var folderClasses = map[string]bool{
	"com.cloudbees.hudson.plugins.folder.Folder":                             true,
	"jenkins.branch.OrganizationFolder":                                      true,
	"org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject": true,
}

// ListAllJobs lists all jobs recursively, filtering by since date.
func (c *Client) ListAllJobs(since time.Time) ([]APIJobSummary, error) {
	return c.listJobsRecursive("", "", since)
}

func (c *Client) listJobsRecursive(path, apiPath string, since time.Time) ([]APIJobSummary, error) {
	endpoint := apiPath + "/api/json"
	params := url.Values{}
	params.Set("tree", "jobs[name,url,_class,lastBuild[number,timestamp],jobs[name,url,_class,lastBuild[number,timestamp],jobs[name,url,_class,lastBuild[number,timestamp]]]]")

	body, err := c.doRequest("GET", endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("list jobs at %s: %w", path, err)
	}

	var response struct {
		Jobs []APIJobSummary `json:"jobs"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse jobs response: %w", err)
	}

	var result []APIJobSummary
	for _, j := range response.Jobs {
		fullName := j.Name
		if path != "" {
			fullName = path + "/" + j.Name
		}

		if folderClasses[j.Class] {
			// Recurse into folder
			folderPath := apiPath + "/job/" + url.PathEscape(j.Name)
			subJobs, err := c.listJobsRecursive(fullName, folderPath, since)
			if err != nil {
				slog.Warn("jenkins: failed to list folder", "folder", fullName, "error", err)
				continue
			}
			result = append(result, subJobs...)
			continue
		}

		// Filter by since: skip jobs whose last build is older than since
		if !since.IsZero() && j.LastBuild != nil {
			lastBuildTime := time.UnixMilli(j.LastBuild.Timestamp)
			if lastBuildTime.Before(since) {
				continue
			}
		}

		j.Name = fullName
		result = append(result, j)
	}

	return result, nil
}

// GetJobDetails fetches full job details including build list (last 100).
func (c *Client) GetJobDetails(jobName string) (*APIJobDetail, error) {
	endpoint := c.jobAPIPath(jobName) + "/api/json"
	params := url.Values{}
	params.Set("tree", "name,url,_class,buildable,builds[number,url]")

	body, err := c.doRequest("GET", endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("get job details %s: %w", jobName, err)
	}

	var detail APIJobDetail
	if err := json.Unmarshal(body, &detail); err != nil {
		return nil, fmt.Errorf("parse job details: %w", err)
	}

	return &detail, nil
}

// GetBuildDetails fetches full build details including actions.
func (c *Client) GetBuildDetails(jobName string, buildNumber int) (*APIBuildDetail, error) {
	endpoint := fmt.Sprintf("%s/%d/api/json", c.jobAPIPath(jobName), buildNumber)

	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("get build %s#%d: %w", jobName, buildNumber, err)
	}

	var detail APIBuildDetail
	if err := json.Unmarshal(body, &detail); err != nil {
		return nil, fmt.Errorf("parse build details: %w", err)
	}

	return &detail, nil
}

// jobAPIPath converts a job name like "FolderA/FolderB/MyJob" to "/job/FolderA/job/FolderB/job/MyJob".
func (c *Client) jobAPIPath(jobName string) string {
	parts := strings.Split(jobName, "/")
	var segments []string
	for _, p := range parts {
		segments = append(segments, "job", url.PathEscape(p))
	}
	return "/" + strings.Join(segments, "/")
}

func (c *Client) doRequest(method, endpoint string, params url.Values) ([]byte, error) {
	requestURL := c.baseURL + endpoint
	if params != nil {
		requestURL += "?" + params.Encode()
	}

	start := time.Now()

	req, err := http.NewRequest(method, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.apiToken)
	req.Header.Set("Accept", "application/json")

	slog.Debug("jenkins api request", "method", method, "endpoint", endpoint)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("jenkins api request failed", "method", method, "endpoint", endpoint, "error", err, "durationMs", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	slog.Info("jenkins api response", "method", method, "endpoint", endpoint, "status", resp.StatusCode, "responseBytes", len(body), "durationMs", time.Since(start).Milliseconds())

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("authentication failed (status: %d)", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
