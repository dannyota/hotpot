package jenkins

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeJenkinsJob represents a Jenkins job in the bronze layer.
type BronzeJenkinsJob struct {
	ent.Schema
}

func (BronzeJenkinsJob) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeJenkinsJob) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Jenkins job name (full path for folder jobs)"),
		field.String("job_class").
			Optional().
			Comment("Jenkins job _class (e.g., hudson.model.FreeStyleProject)"),
		field.String("project_type").
			Optional().
			Comment("Simplified project type (freestyle, pipeline, multibranch, etc.)"),
		field.Bool("is_buildable").
			Default(false),
		field.Int("last_build_number").
			Default(0),
		field.Time("last_build_time").
			Optional().
			Nillable(),
	}
}

func (BronzeJenkinsJob) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("builds", BronzeJenkinsBuild.Type),
	}
}

func (BronzeJenkinsJob) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("is_buildable"),
	}
}

func (BronzeJenkinsJob) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "jenkins_jobs"},
	}
}

// BronzeJenkinsBuild represents a Jenkins build in the bronze layer.
type BronzeJenkinsBuild struct {
	ent.Schema
}

func (BronzeJenkinsBuild) Fields() []ent.Field {
	return []ent.Field{
		field.Int("build_number"),
		field.String("result").
			Optional().
			Comment("Build result: SUCCESS, FAILURE, UNSTABLE, ABORTED, NOT_BUILT"),
		field.Time("timestamp").
			Optional().
			Nillable().
			Comment("Build start time"),
		field.Int64("duration_ms").
			Default(0),
		field.String("version").
			Optional().
			Comment("Version extracted from build parameters"),
		field.Bool("check_code_enabled").
			Default(false).
			Comment("Whether build had SCM checkout"),
	}
}

func (BronzeJenkinsBuild) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("job", BronzeJenkinsJob.Type).
			Ref("builds").
			Unique().
			Required(),
		edge.To("repos", BronzeJenkinsBuildRepo.Type),
	}
}

func (BronzeJenkinsBuild) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("build_number"),
		index.Fields("result"),
	}
}

func (BronzeJenkinsBuild) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "jenkins_builds"},
	}
}

// BronzeJenkinsBuildRepo represents a git repository associated with a Jenkins build.
type BronzeJenkinsBuildRepo struct {
	ent.Schema
}

func (BronzeJenkinsBuildRepo) Fields() []ent.Field {
	return []ent.Field{
		field.String("repo_url").
			Optional(),
		field.String("branch").
			Optional(),
		field.String("commit_sha").
			Optional(),
	}
}

func (BronzeJenkinsBuildRepo) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("build", BronzeJenkinsBuild.Type).
			Ref("repos").
			Unique().
			Required(),
	}
}

func (BronzeJenkinsBuildRepo) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "jenkins_build_repos"},
	}
}
