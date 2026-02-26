package job

import (
	entjenkins "github.com/dannyota/hotpot/pkg/storage/ent/jenkins"
)

// JobDiff represents changes between old and new job states.
type JobDiff struct {
	IsNew       bool
	IsChanged   bool
	BuildsDiff  ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// HasAnyChange returns true if any part of the job changed.
func (d *JobDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.BuildsDiff.Changed
}

// DiffJobData compares old Ent entity and new data.
func DiffJobData(old *entjenkins.BronzeJenkinsJob, new *JobData) *JobDiff {
	if old == nil {
		return &JobDiff{
			IsNew:      true,
			BuildsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &JobDiff{}
	diff.IsChanged = hasJobFieldsChanged(old, new)
	// Builds are append-only (new builds added each run), so we always check
	diff.BuildsDiff = ChildDiff{Changed: len(new.Builds) > 0}

	return diff
}

func hasJobFieldsChanged(old *entjenkins.BronzeJenkinsJob, new *JobData) bool {
	return old.JobClass != new.JobClass ||
		old.ProjectType != new.ProjectType ||
		old.IsBuildable != new.IsBuildable ||
		old.LastBuildNumber != new.LastBuildNumber
}
