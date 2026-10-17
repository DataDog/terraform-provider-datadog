package fwprovider

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// githubOIDCClaimMatchersModel mirrors datadogV2.GitHubOIDCClaimPatterns: the
// required `sub` claim plus the optional GitHub Actions OIDC claim regexes.
// It is shared by the GitHub OIDC persona and intake mapping resources.
type githubOIDCClaimMatchersModel struct {
	Actor                types.String `tfsdk:"actor"`
	ActorId              types.String `tfsdk:"actor_id"`
	Enterprise           types.String `tfsdk:"enterprise"`
	EnterpriseId         types.String `tfsdk:"enterprise_id"`
	Environment          types.String `tfsdk:"environment"`
	EventName            types.String `tfsdk:"event_name"`
	JobWorkflowRef       types.String `tfsdk:"job_workflow_ref"`
	Ref                  types.String `tfsdk:"ref"`
	RefType              types.String `tfsdk:"ref_type"`
	Repository           types.String `tfsdk:"repository"`
	RepositoryId         types.String `tfsdk:"repository_id"`
	RepositoryOwner      types.String `tfsdk:"repository_owner"`
	RepositoryOwnerId    types.String `tfsdk:"repository_owner_id"`
	RepositoryVisibility types.String `tfsdk:"repository_visibility"`
	RunnerEnvironment    types.String `tfsdk:"runner_environment"`
	Sub                  types.String `tfsdk:"sub"`
	Workflow             types.String `tfsdk:"workflow"`
	WorkflowRef          types.String `tfsdk:"workflow_ref"`
}

// githubOIDCClaimMatchersAttribute returns the claim matchers schema shared by
// the GitHub OIDC persona and intake mapping resources. Mappings cannot be
// updated, so every claim pattern forces replacement.
func githubOIDCClaimMatchersAttribute() schema.SingleNestedAttribute {
	attributes := map[string]schema.Attribute{
		"sub": schema.StringAttribute{
			Description: "Regular expression matched against the required `sub` claim of the GitHub Actions OIDC token. It must start with the literal `repo:<OWNER>/` prefix; the owner cannot be a regex.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
	}
	for _, claim := range []string{
		"actor", "actor_id", "enterprise", "enterprise_id", "environment",
		"event_name", "job_workflow_ref", "ref", "ref_type", "repository",
		"repository_id", "repository_owner", "repository_owner_id",
		"repository_visibility", "runner_environment", "workflow", "workflow_ref",
	} {
		attributes[claim] = schema.StringAttribute{
			Description: fmt.Sprintf("Regular expression matched against the %#q claim of the GitHub Actions OIDC token.", claim),
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		}
	}
	return schema.SingleNestedAttribute{
		Description: "GitHub Actions OIDC claims to match against. Each attribute is a regular expression; `sub` is required and all other claims are optional. A token matches only when all provided patterns match simultaneously (AND semantics).",
		Attributes:  attributes,
		Required:    true,
	}
}

func (m githubOIDCClaimMatchersModel) toSDK() datadogV2.GitHubOIDCClaimPatterns {
	patterns := datadogV2.NewGitHubOIDCClaimPatterns(m.Sub.ValueString())
	setClaimPattern(m.Actor, patterns.SetActor)
	setClaimPattern(m.ActorId, patterns.SetActorId)
	setClaimPattern(m.Enterprise, patterns.SetEnterprise)
	setClaimPattern(m.EnterpriseId, patterns.SetEnterpriseId)
	setClaimPattern(m.Environment, patterns.SetEnvironment)
	setClaimPattern(m.EventName, patterns.SetEventName)
	setClaimPattern(m.JobWorkflowRef, patterns.SetJobWorkflowRef)
	setClaimPattern(m.Ref, patterns.SetRef)
	setClaimPattern(m.RefType, patterns.SetRefType)
	setClaimPattern(m.Repository, patterns.SetRepository)
	setClaimPattern(m.RepositoryId, patterns.SetRepositoryId)
	setClaimPattern(m.RepositoryOwner, patterns.SetRepositoryOwner)
	setClaimPattern(m.RepositoryOwnerId, patterns.SetRepositoryOwnerId)
	setClaimPattern(m.RepositoryVisibility, patterns.SetRepositoryVisibility)
	setClaimPattern(m.RunnerEnvironment, patterns.SetRunnerEnvironment)
	setClaimPattern(m.Workflow, patterns.SetWorkflow)
	setClaimPattern(m.WorkflowRef, patterns.SetWorkflowRef)
	return *patterns
}

func setClaimPattern(value types.String, set func(string)) {
	if !value.IsNull() && !value.IsUnknown() {
		set(value.ValueString())
	}
}

func githubOIDCClaimMatchersFromSDK(patterns datadogV2.GitHubOIDCClaimPatterns) githubOIDCClaimMatchersModel {
	claimValue := func(value *string, ok bool) types.String {
		if !ok || value == nil {
			return types.StringNull()
		}
		return types.StringValue(*value)
	}
	return githubOIDCClaimMatchersModel{
		Actor:                claimValue(patterns.GetActorOk()),
		ActorId:              claimValue(patterns.GetActorIdOk()),
		Enterprise:           claimValue(patterns.GetEnterpriseOk()),
		EnterpriseId:         claimValue(patterns.GetEnterpriseIdOk()),
		Environment:          claimValue(patterns.GetEnvironmentOk()),
		EventName:            claimValue(patterns.GetEventNameOk()),
		JobWorkflowRef:       claimValue(patterns.GetJobWorkflowRefOk()),
		Ref:                  claimValue(patterns.GetRefOk()),
		RefType:              claimValue(patterns.GetRefTypeOk()),
		Repository:           claimValue(patterns.GetRepositoryOk()),
		RepositoryId:         claimValue(patterns.GetRepositoryIdOk()),
		RepositoryOwner:      claimValue(patterns.GetRepositoryOwnerOk()),
		RepositoryOwnerId:    claimValue(patterns.GetRepositoryOwnerIdOk()),
		RepositoryVisibility: claimValue(patterns.GetRepositoryVisibilityOk()),
		RunnerEnvironment:    claimValue(patterns.GetRunnerEnvironmentOk()),
		Sub:                  claimValue(patterns.GetSubOk()),
		Workflow:             claimValue(patterns.GetWorkflowOk()),
		WorkflowRef:          claimValue(patterns.GetWorkflowRefOk()),
	}
}
