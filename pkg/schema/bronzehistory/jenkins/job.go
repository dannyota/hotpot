package jenkins

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "danny.vn/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryJenkinsJob stores historical snapshots of Jenkins jobs.
type BronzeHistoryJenkinsJob struct {
	ent.Schema
}

func (BronzeHistoryJenkinsJob) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryJenkinsJob) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze job by resource_id (job name)"),
		field.String("job_class").
			Optional(),
		field.String("project_type").
			Optional(),
		field.Bool("is_buildable").
			Default(false),
		field.Int("last_build_number").
			Default(0),
		field.Time("last_build_time").
			Optional().
			Nillable(),
	}
}

func (BronzeHistoryJenkinsJob) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryJenkinsJob) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "jenkins_jobs_history"},
	}
}

// BronzeHistoryJenkinsBuild stores historical snapshots of Jenkins builds.
type BronzeHistoryJenkinsBuild struct {
	ent.Schema
}

func (BronzeHistoryJenkinsBuild) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.Uint("job_history_id").
			Comment("Links to parent BronzeHistoryJenkinsJob"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),
		field.Int("build_number"),
		field.String("result").
			Optional(),
		field.Time("timestamp").
			Optional().
			Nillable(),
		field.Int64("duration_ms").
			Default(0),
		field.String("version").
			Optional(),
		field.Bool("check_code_enabled").
			Default(false),
	}
}

func (BronzeHistoryJenkinsBuild) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("job_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryJenkinsBuild) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "jenkins_builds_history"},
	}
}

// BronzeHistoryJenkinsBuildRepo stores historical snapshots of Jenkins build repos.
type BronzeHistoryJenkinsBuildRepo struct {
	ent.Schema
}

func (BronzeHistoryJenkinsBuildRepo) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.Uint("build_history_id").
			Comment("Links to parent BronzeHistoryJenkinsBuild"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),
		field.String("repo_url").
			Optional(),
		field.String("branch").
			Optional(),
		field.String("commit_sha").
			Optional(),
	}
}

func (BronzeHistoryJenkinsBuildRepo) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("build_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryJenkinsBuildRepo) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "jenkins_build_repos_history"},
	}
}
