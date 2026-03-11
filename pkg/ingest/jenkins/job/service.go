package job

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	entjenkins "danny.vn/hotpot/pkg/storage/ent/jenkins"
	"danny.vn/hotpot/pkg/storage/ent/jenkins/bronzejenkinsbuild"
	"danny.vn/hotpot/pkg/storage/ent/jenkins/bronzejenkinsbuildrepo"
	"danny.vn/hotpot/pkg/storage/ent/jenkins/bronzejenkinsjob"
)

// Service handles Jenkins job ingestion.
type Service struct {
	client    *Client
	entClient *entjenkins.Client
	history   *HistoryService
}

// NewService creates a new job ingestion service.
func NewService(client *Client, entClient *entjenkins.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of job ingestion.
type IngestResult struct {
	JobCount       int
	BuildCount     int
	RepoCount      int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all jobs and their builds from Jenkins.
func (s *Service) Ingest(ctx context.Context, since time.Time, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// 1. List all jobs (filtered by since)
	jobs, err := s.client.ListAllJobs(since)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}

	slog.Info("jenkins jobs listed", "jobCount", len(jobs))

	if heartbeat != nil {
		heartbeat()
	}

	totalBuilds := 0
	totalRepos := 0

	// 2. For each job, fetch details and new builds
	for i, apiJob := range jobs {
		jobData, newBuilds, newRepos, err := s.fetchJobWithBuilds(ctx, apiJob, collectedAt)
		if err != nil {
			slog.Error("jenkins: failed to fetch job", "job", apiJob.Name, "error", err)
			continue
		}

		if err := s.saveJob(ctx, jobData); err != nil {
			return nil, fmt.Errorf("save job %s: %w", apiJob.Name, err)
		}

		totalBuilds += newBuilds
		totalRepos += newRepos

		if (i+1)%10 == 0 {
			slog.Info("jenkins jobs progress", "processed", i+1, "total", len(jobs), "builds", totalBuilds)
			if heartbeat != nil {
				heartbeat()
			}
		}
	}

	return &IngestResult{
		JobCount:       len(jobs),
		BuildCount:     totalBuilds,
		RepoCount:      totalRepos,
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// fetchJobWithBuilds fetches job details and new builds since last ingestion.
func (s *Service) fetchJobWithBuilds(ctx context.Context, apiJob APIJobSummary, collectedAt time.Time) (*JobData, int, int, error) {
	// Get full job details (includes last 100 builds)
	detail, err := s.client.GetJobDetails(apiJob.Name)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("get job details: %w", err)
	}

	jobData := ConvertJob(apiJob, detail.Class, detail.Buildable, collectedAt)

	// Query DB for last known build number
	lastBuildNumber := 0
	existing, err := s.entClient.BronzeJenkinsJob.Query().
		Where(bronzejenkinsjob.ID(apiJob.Name)).
		First(ctx)
	if err != nil && !entjenkins.IsNotFound(err) {
		return nil, 0, 0, fmt.Errorf("query existing job: %w", err)
	}
	if existing != nil {
		lastBuildNumber = existing.LastBuildNumber
	}

	// Filter builds to only new ones (number > lastBuildNumber)
	var newBuildNumbers []int
	for _, b := range detail.Builds {
		if b.Number > lastBuildNumber {
			newBuildNumbers = append(newBuildNumbers, b.Number)
		}
	}

	// Sort descending to keep newest builds when capping
	sort.Sort(sort.Reverse(sort.IntSlice(newBuildNumbers)))

	// Apply safety cap
	if len(newBuildNumbers) > s.client.maxBuildsPerJob {
		slog.Warn("jenkins: capping builds per job", "job", apiJob.Name, "total", len(newBuildNumbers), "cap", s.client.maxBuildsPerJob)
		newBuildNumbers = newBuildNumbers[:s.client.maxBuildsPerJob]
	}

	totalRepos := 0
	for _, buildNum := range newBuildNumbers {
		buildDetail, err := s.client.GetBuildDetails(apiJob.Name, buildNum)
		if err != nil {
			slog.Warn("jenkins: failed to fetch build", "job", apiJob.Name, "build", buildNum, "error", err)
			continue
		}

		buildData := ConvertBuild(buildDetail)

		// Filter excluded repos
		if len(s.client.excludeRepos) > 0 {
			buildData.Repos = filterRepos(buildData.Repos, s.client.excludeRepos)
		}

		totalRepos += len(buildData.Repos)
		jobData.Builds = append(jobData.Builds, buildData)
	}

	// Update last build number to the highest we've seen
	if len(newBuildNumbers) > 0 {
		maxBuild := newBuildNumbers[0]
		for _, n := range newBuildNumbers[1:] {
			if n > maxBuild {
				maxBuild = n
			}
		}
		jobData.LastBuildNumber = maxBuild
		if apiJob.LastBuild != nil {
			t := time.UnixMilli(apiJob.LastBuild.Timestamp)
			jobData.LastBuildTime = &t
		}
	}

	return jobData, len(jobData.Builds), totalRepos, nil
}

// filterRepos removes repos matching excluded patterns.
func filterRepos(repos []RepoData, excludePatterns []string) []RepoData {
	if len(excludePatterns) == 0 {
		return repos
	}
	var filtered []RepoData
	for _, repo := range repos {
		excluded := false
		for _, pattern := range excludePatterns {
			if repo.RepoURL == pattern {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, repo)
		}
	}
	return filtered
}

func (s *Service) saveJob(ctx context.Context, data *JobData) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	existing, err := tx.BronzeJenkinsJob.Query().
		Where(bronzejenkinsjob.ID(data.ResourceID)).
		First(ctx)
	if err != nil && !entjenkins.IsNotFound(err) {
		tx.Rollback()
		return fmt.Errorf("load existing job %s: %w", data.ResourceID, err)
	}

	diff := DiffJobData(existing, data)

	if !diff.HasAnyChange() && existing != nil {
		if err := tx.BronzeJenkinsJob.UpdateOneID(data.ResourceID).
			SetCollectedAt(data.CollectedAt).
			Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("update collected_at for job %s: %w", data.ResourceID, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit transaction: %w", err)
		}
		return nil
	}

	var savedJob *entjenkins.BronzeJenkinsJob
	if existing == nil {
		create := tx.BronzeJenkinsJob.Create().
			SetID(data.ResourceID).
			SetJobClass(data.JobClass).
			SetProjectType(data.ProjectType).
			SetIsBuildable(data.IsBuildable).
			SetLastBuildNumber(data.LastBuildNumber).
			SetCollectedAt(data.CollectedAt).
			SetFirstCollectedAt(data.CollectedAt)

		if data.LastBuildTime != nil {
			create.SetLastBuildTime(*data.LastBuildTime)
		}

		savedJob, err = create.Save(ctx)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("create job %s: %w", data.ResourceID, err)
		}
	} else {
		update := tx.BronzeJenkinsJob.UpdateOneID(data.ResourceID).
			SetJobClass(data.JobClass).
			SetProjectType(data.ProjectType).
			SetIsBuildable(data.IsBuildable).
			SetLastBuildNumber(data.LastBuildNumber).
			SetCollectedAt(data.CollectedAt)

		if data.LastBuildTime != nil {
			update.SetLastBuildTime(*data.LastBuildTime)
		}

		savedJob, err = update.Save(ctx)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("update job %s: %w", data.ResourceID, err)
		}
	}

	// Create new builds and their repos
	if err := s.createBuildsAndRepos(ctx, tx, savedJob, data.Builds); err != nil {
		tx.Rollback()
		return fmt.Errorf("create builds for job %s: %w", data.ResourceID, err)
	}

	// Track history
	if diff.IsNew {
		if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("create history for job %s: %w", data.ResourceID, err)
		}
	} else {
		if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("update history for job %s: %w", data.ResourceID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *Service) createBuildsAndRepos(ctx context.Context, tx *entjenkins.Tx, job *entjenkins.BronzeJenkinsJob, builds []BuildData) error {
	for _, build := range builds {
		buildCreate := tx.BronzeJenkinsBuild.Create().
			SetJob(job).
			SetBuildNumber(build.BuildNumber).
			SetResult(build.Result).
			SetDurationMs(build.DurationMs).
			SetVersion(build.Version).
			SetCheckCodeEnabled(build.CheckCodeEnabled)

		if build.Timestamp != nil {
			buildCreate.SetTimestamp(*build.Timestamp)
		}

		savedBuild, err := buildCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("create build #%d: %w", build.BuildNumber, err)
		}

		for _, repo := range build.Repos {
			_, err := tx.BronzeJenkinsBuildRepo.Create().
				SetBuild(savedBuild).
				SetRepoURL(repo.RepoURL).
				SetBranch(repo.Branch).
				SetCommitSha(repo.CommitSHA).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create build repo: %w", err)
			}
		}
	}
	return nil
}

// DeleteStale removes jobs that were not collected in the latest run.
func (s *Service) DeleteStale(ctx context.Context, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	stale, err := tx.BronzeJenkinsJob.Query().
		Where(bronzejenkinsjob.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, job := range stale {
		if err := s.history.CloseHistory(ctx, tx, job.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for job %s: %w", job.ID, err)
		}

		if err := s.deleteJobChildren(ctx, tx, job.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete children for job %s: %w", job.ID, err)
		}

		if err := tx.BronzeJenkinsJob.DeleteOne(job).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete job %s: %w", job.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *Service) deleteJobChildren(ctx context.Context, tx *entjenkins.Tx, jobID string) error {
	// Get all builds for this job
	builds, err := tx.BronzeJenkinsBuild.Query().
		Where(bronzejenkinsbuild.HasJobWith(bronzejenkinsjob.ID(jobID))).
		All(ctx)
	if err != nil {
		return fmt.Errorf("query builds: %w", err)
	}

	// Delete repos for each build
	for _, build := range builds {
		_, err := tx.BronzeJenkinsBuildRepo.Delete().
			Where(bronzejenkinsbuildrepo.HasBuildWith(bronzejenkinsbuild.ID(build.ID))).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("delete repos for build %d: %w", build.ID, err)
		}
	}

	// Delete builds
	_, err = tx.BronzeJenkinsBuild.Delete().
		Where(bronzejenkinsbuild.HasJobWith(bronzejenkinsjob.ID(jobID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete builds: %w", err)
	}

	return nil
}
