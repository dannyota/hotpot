package job

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// JobData holds converted job data ready for Ent insertion.
type JobData struct {
	ResourceID      string // job name (full path)
	JobClass        string
	ProjectType     string
	IsBuildable     bool
	LastBuildNumber int
	LastBuildTime   *time.Time
	CollectedAt     time.Time

	// Child data
	Builds []BuildData
}

// BuildData holds converted build data.
type BuildData struct {
	BuildNumber      int
	Result           string
	Timestamp        *time.Time
	DurationMs       int64
	Version          string
	CheckCodeEnabled bool

	// Child data
	Repos []RepoData
}

// RepoData holds converted repo data.
type RepoData struct {
	RepoURL   string
	Branch    string
	CommitSHA string
}

// ConvertJob converts an API job summary to JobData.
func ConvertJob(apiJob APIJobSummary, jobClass string, isBuildable bool, collectedAt time.Time) *JobData {
	data := &JobData{
		ResourceID:  apiJob.Name,
		JobClass:    jobClass,
		ProjectType: classifyProjectType(jobClass),
		IsBuildable: isBuildable,
		CollectedAt: collectedAt,
	}

	if apiJob.LastBuild != nil {
		data.LastBuildNumber = apiJob.LastBuild.Number
		t := time.UnixMilli(apiJob.LastBuild.Timestamp)
		data.LastBuildTime = &t
	}

	return data
}

// ConvertBuild converts an API build detail to BuildData.
func ConvertBuild(apiBuild *APIBuildDetail) BuildData {
	data := BuildData{
		BuildNumber: apiBuild.Number,
		Result:      apiBuild.Result,
		DurationMs:  apiBuild.Duration,
	}

	if apiBuild.Timestamp > 0 {
		t := time.UnixMilli(apiBuild.Timestamp)
		data.Timestamp = &t
	}

	data.Version = extractVersion(apiBuild.Actions)
	data.CheckCodeEnabled = hasCheckCode(apiBuild.Actions)
	data.Repos = ExtractRepos(apiBuild.Actions)

	return data
}

// ExtractRepos extracts git repo information from build actions.
func ExtractRepos(actions []json.RawMessage) []RepoData {
	var repos []RepoData
	seen := map[string]bool{}

	for _, raw := range actions {
		var action APIBuildAction
		if err := json.Unmarshal(raw, &action); err != nil {
			continue
		}

		if action.Class != "hudson.plugins.git.util.BuildData" {
			continue
		}

		for _, repoURL := range action.RemoteURLs {
			if repoURL == "" {
				continue
			}

			branch := ""
			commitSHA := ""

			if action.LastBuiltRevision != nil {
				commitSHA = action.LastBuiltRevision.SHA1
				if len(action.LastBuiltRevision.Branch) > 0 {
					branch = normalizeBranch(action.LastBuiltRevision.Branch[0].Name)
				}
			}

			key := fmt.Sprintf("%s|%s|%s", repoURL, branch, commitSHA)
			if seen[key] {
				continue
			}
			seen[key] = true

			repos = append(repos, RepoData{
				RepoURL:   repoURL,
				Branch:    branch,
				CommitSHA: commitSHA,
			})
		}
	}

	return repos
}

// normalizeBranch strips common prefixes from branch names.
func normalizeBranch(branch string) string {
	branch = strings.TrimPrefix(branch, "refs/remotes/origin/")
	branch = strings.TrimPrefix(branch, "refs/heads/")
	return branch
}

// extractVersion looks for version info in build parameters.
func extractVersion(actions []json.RawMessage) string {
	versionParams := map[string]bool{
		"VERSION":     true,
		"APP_VERSION": true,
		"RELEASE":     true,
		"TAG":         true,
	}

	for _, raw := range actions {
		var action APIBuildAction
		if err := json.Unmarshal(raw, &action); err != nil {
			continue
		}

		for _, param := range action.Parameters {
			name := strings.ToUpper(param.Name)
			if versionParams[name] {
				if v, ok := param.Value.(string); ok && v != "" {
					return v
				}
			}
		}
	}

	return ""
}

// hasCheckCode checks if the build had SCM checkout actions.
func hasCheckCode(actions []json.RawMessage) bool {
	for _, raw := range actions {
		var action APIBuildAction
		if err := json.Unmarshal(raw, &action); err != nil {
			continue
		}
		if action.Class == "hudson.plugins.git.util.BuildData" {
			return true
		}
	}
	return false
}

// classifyProjectType maps Jenkins _class to a simplified project type.
func classifyProjectType(class string) string {
	switch {
	case strings.Contains(class, "FreeStyleProject"):
		return "freestyle"
	case strings.Contains(class, "WorkflowJob"):
		return "pipeline"
	case strings.Contains(class, "WorkflowMultiBranchProject"):
		return "multibranch"
	case strings.Contains(class, "MavenModuleSet"):
		return "maven"
	case strings.Contains(class, "MatrixProject"):
		return "matrix"
	default:
		return "other"
	}
}
