package job

import (
	"context"
	"fmt"
	"time"

	entjenkins "danny.vn/hotpot/pkg/storage/ent/jenkins"
	"danny.vn/hotpot/pkg/storage/ent/jenkins/bronzehistoryjenkinsbuild"
	"danny.vn/hotpot/pkg/storage/ent/jenkins/bronzehistoryjenkinsbuildrepo"
	"danny.vn/hotpot/pkg/storage/ent/jenkins/bronzehistoryjenkinsjob"
)

// HistoryService handles history tracking for Jenkins jobs.
type HistoryService struct {
	entClient *entjenkins.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entjenkins.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new job and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entjenkins.Tx, data *JobData, now time.Time) error {
	jobHistCreate := h.buildJobHistoryCreate(tx, data, data.CollectedAt, now)

	jobHist, err := jobHistCreate.Save(ctx)
	if err != nil {
		return fmt.Errorf("create job history: %w", err)
	}

	return h.createBuildsHistory(ctx, tx, jobHist.ID, data.Builds, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entjenkins.Tx, old *entjenkins.BronzeJenkinsJob, new *JobData, diff *JobDiff, now time.Time) error {
	currentHist, err := tx.BronzeHistoryJenkinsJob.Query().
		Where(
			bronzehistoryjenkinsjob.ResourceID(old.ID),
			bronzehistoryjenkinsjob.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current job history: %w", err)
	}

	if diff.IsChanged {
		if err := tx.BronzeHistoryJenkinsJob.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("close job history: %w", err)
		}

		jobHistCreate := h.buildJobHistoryCreate(tx, new, old.FirstCollectedAt, now)
		jobHist, err := jobHistCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("create new job history: %w", err)
		}

		if err := h.closeBuildsHistory(ctx, tx, currentHist.ID, now); err != nil {
			return fmt.Errorf("close builds history: %w", err)
		}
		return h.createBuildsHistory(ctx, tx, jobHist.ID, new.Builds, now)
	}

	// Job unchanged, but new builds may have been added
	if diff.BuildsDiff.Changed {
		return h.createBuildsHistory(ctx, tx, currentHist.ID, new.Builds, now)
	}

	return nil
}

// CloseHistory closes history records for a deleted job.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entjenkins.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryJenkinsJob.Query().
		Where(
			bronzehistoryjenkinsjob.ResourceID(resourceID),
			bronzehistoryjenkinsjob.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entjenkins.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current job history: %w", err)
	}

	if err := tx.BronzeHistoryJenkinsJob.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close job history: %w", err)
	}

	return h.closeBuildsHistory(ctx, tx, currentHist.ID, now)
}

func (h *HistoryService) buildJobHistoryCreate(tx *entjenkins.Tx, data *JobData, firstCollectedAt time.Time, now time.Time) *entjenkins.BronzeHistoryJenkinsJobCreate {
	create := tx.BronzeHistoryJenkinsJob.Create().
		SetResourceID(data.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(firstCollectedAt).
		SetJobClass(data.JobClass).
		SetProjectType(data.ProjectType).
		SetIsBuildable(data.IsBuildable).
		SetLastBuildNumber(data.LastBuildNumber)

	if data.LastBuildTime != nil {
		create.SetLastBuildTime(*data.LastBuildTime)
	}

	return create
}

func (h *HistoryService) createBuildsHistory(ctx context.Context, tx *entjenkins.Tx, jobHistoryID uint, builds []BuildData, now time.Time) error {
	for _, build := range builds {
		buildCreate := tx.BronzeHistoryJenkinsBuild.Create().
			SetJobHistoryID(jobHistoryID).
			SetValidFrom(now).
			SetBuildNumber(build.BuildNumber).
			SetResult(build.Result).
			SetDurationMs(build.DurationMs).
			SetVersion(build.Version).
			SetCheckCodeEnabled(build.CheckCodeEnabled)

		if build.Timestamp != nil {
			buildCreate.SetTimestamp(*build.Timestamp)
		}

		buildHist, err := buildCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("create build history #%d: %w", build.BuildNumber, err)
		}

		for _, repo := range build.Repos {
			_, err := tx.BronzeHistoryJenkinsBuildRepo.Create().
				SetBuildHistoryID(buildHist.ID).
				SetValidFrom(now).
				SetRepoURL(repo.RepoURL).
				SetBranch(repo.Branch).
				SetCommitSha(repo.CommitSHA).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create build repo history: %w", err)
			}
		}
	}
	return nil
}

func (h *HistoryService) closeBuildsHistory(ctx context.Context, tx *entjenkins.Tx, jobHistoryID uint, now time.Time) error {
	// Close build repos first
	buildHists, err := tx.BronzeHistoryJenkinsBuild.Query().
		Where(
			bronzehistoryjenkinsbuild.JobHistoryID(jobHistoryID),
			bronzehistoryjenkinsbuild.ValidToIsNil(),
		).
		All(ctx)
	if err != nil {
		return fmt.Errorf("query build history: %w", err)
	}

	for _, buildHist := range buildHists {
		_, err := tx.BronzeHistoryJenkinsBuildRepo.Update().
			Where(
				bronzehistoryjenkinsbuildrepo.BuildHistoryID(buildHist.ID),
				bronzehistoryjenkinsbuildrepo.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("close build repo history: %w", err)
		}
	}

	// Close builds
	_, err = tx.BronzeHistoryJenkinsBuild.Update().
		Where(
			bronzehistoryjenkinsbuild.JobHistoryID(jobHistoryID),
			bronzehistoryjenkinsbuild.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close build history: %w", err)
	}

	return nil
}
