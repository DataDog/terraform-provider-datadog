# Import existing UI mappings before applying this resource to avoid creating duplicates.
# Import an existing mapping using teams[].mapping_id from /rules,
# not the grouped rule ID.
terraform import datadog_rum_team_ownership.checkout "<mapping_id>"
