package datadog

import (
	"context"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogUser() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing user to use it in an other resources.",
		ReadContext: dataSourceDatadogUserRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"filter": {
					Description: "Filter all users by the given string.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"exact_match": {
					Description: "When true, `filter` string is exact matched against the user's `email`, followed by `name` attribute.",
					Type:        schema.TypeBool,
					Default:     false,
					Optional:    true,
				},
				"exclude_service_accounts": {
					Description: "When true, service accounts are excluded from the result.",
					Type:        schema.TypeBool,
					Default:     false,
					Optional:    true,
				},
				// Computed values
				"created_at": {
					Description: "The time when the user was created (RFC3339 format).",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"disabled": {
					Description: "Indicates whether the user is disabled.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"email": {
					Description: "Email of the user.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"handle": {
					Description: "The user's handle.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"icon": {
					Description: "The URL where the user's icon is located.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"mfa_enabled": {
					Description: "Indicates whether the user has enabled MFA.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"modified_at": {
					Description: "The time at which the user was last updated (RFC3339 format).",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"name": {
					Description: "Name of the user.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"service_account": {
					Description: "Indicates whether the user is a service account.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"status": {
					Description: "The user's status.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"title": {
					Description: "The user's title.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"verified": {
					Description: "Indicates whether the user is verified.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
			}
		},
	}
}

func dataSourceDatadogUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	filter := d.Get("filter").(string) // string | Filter all users by the given string. Defaults to no filtering. (optional) // string | Filter on status attribute. Comma separated list, with possible values `Active`, `Pending`, and `Disabled`. Defaults to no filtering. (optional)
	exactMatch := d.Get("exact_match").(bool)
	excludeServiceAccounts := d.Get("exclude_service_accounts").(bool)
	optionalParams := datadogV2.ListUsersOptionalParameters{
		Filter: &filter,
	}

	res, httpresp, err := apiInstances.GetUsersApiV2().ListUsers(auth, optionalParams)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying user")
	}

	users := res.GetData()
	errorPrefix := ""
	if excludeServiceAccounts {
		errorPrefix = "after excluding service accounts, "
		filteredUsers := make([]datadogV2.User, 0)
		for _, user := range users {
			if !user.Attributes.GetServiceAccount() {
				filteredUsers = append(filteredUsers, user)
			}
		}
		users = filteredUsers
	}

	if len(users) > 1 && !exactMatch {
		return diag.Errorf("%syour query returned more than one result for filter \"%s\", please try a more specific search criteria",
			errorPrefix,
			filter,
		)
	} else if len(users) == 0 {
		return diag.Errorf("%sdidn't find any user matching filter string \"%s\"", errorPrefix, filter)
	}

	matchedUser := users[0]
	if exactMatch {
		matchCount := 0
		for _, user := range users {
			if user.Attributes.GetEmail() == filter {
				matchedUser = user
				matchCount++
				continue
			}
			if user.Attributes.GetName() == filter {
				matchedUser = user
				matchCount++
				continue
			}
		}
		if matchCount > 1 {
			return diag.Errorf("%syour query returned more than one result for filter with exact match \"%s\", please try a more specific search criteria",
				errorPrefix,
				filter,
			)
		}
		if matchCount == 0 {
			return diag.Errorf("%sdidn't find any user matching filter string with exact match \"%s\"", errorPrefix, filter)
		}
	}

	if err := utils.CheckForUnparsed(matchedUser); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(matchedUser.GetId())
	mapAttrString := map[string]func() string{
		"created_at":  func() string { return matchedUser.Attributes.GetCreatedAt().Format(time.RFC3339) },
		"email":       matchedUser.Attributes.GetEmail,
		"handle":      matchedUser.Attributes.GetHandle,
		"icon":        matchedUser.Attributes.GetIcon,
		"modified_at": func() string { return matchedUser.Attributes.GetModifiedAt().Format(time.RFC3339) },
		"name":        matchedUser.Attributes.GetName,
		"status":      matchedUser.Attributes.GetStatus,
		"title":       matchedUser.Attributes.GetTitle,
	}
	for key, value := range mapAttrString {
		if err := d.Set(key, value()); err != nil {
			return diag.FromErr(err)
		}
	}
	mapAttrBool := map[string]func() bool{
		"disabled":        matchedUser.Attributes.GetDisabled,
		"mfa_enabled":     matchedUser.Attributes.GetMfaEnabled,
		"service_account": matchedUser.Attributes.GetServiceAccount,
	}
	for key, value := range mapAttrBool {
		if err := d.Set(key, value()); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
