package emit

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// CreateInitializationView reuses Update's request mapping after Create, while
// obtaining any secret values through the Create request's configuration.
type CreateInitializationView struct {
	Call   CRUDCallView
	Fields []CreateInitializationField
}

type CreateInitializationField struct {
	GoField          string
	Condition        string
	ResponseNullExpr string
}

func buildCreateInitialization(attrs []*model.Attribute, create, update *model.Schema, call CRUDCallView, secrets []WriteOnlySecretView, unsupported *[]UnsupportedNode) *CreateInitializationView {
	if update == nil {
		return nil
	}
	var fields []CreateInitializationField
	for _, attr := range attrs {
		if !attr.Required && !attr.Optional {
			continue
		}
		updateChild, updateDeclared := roleChild(update, attr)
		if !updateDeclared {
			continue
		}
		createChild, createDeclared := roleChild(create, attr)
		if create == nil {
			createDeclared = false
		}
		if createDeclared {
			checkNestedUpdateOnly(attr.Children, createChild, updateChild, unsupported)
			continue
		}
		if attr.WriteOnlySecret {
			// validateWriteOnlySecrets diagnoses update-only secrets separately.
			continue
		}
		field := model.SdkName(tfNameOf(attr.Path))
		configured := "configured." + field
		condition := configured + " != nil"
		responseNullExpr := ""
		switch attr.TfType {
		case "schema.StringAttribute", "schema.BoolAttribute", "schema.Int64Attribute", "schema.Float64Attribute", "schema.ListAttribute", "schema.SetAttribute", "schema.MapAttribute":
			condition = notNullOrUnknown(configured)
			if attr.InResponse {
				// State starts from the plan. Clear its planned value before the
				// Create response is mapped, so equality only means the server
				// actually returned this value rather than omitted the field.
				responseNullExpr = computedNullExpr(attr)
				condition = fmt.Sprintf("%s && !%s.Equal(state.%s)", condition, configured, field)
			}
		}
		fields = append(fields, CreateInitializationField{GoField: field, Condition: condition, ResponseNullExpr: responseNullExpr})
	}
	if len(fields) == 0 {
		return nil
	}
	call.WriteOnlySecrets = writeOnlySecretsForFields(secrets, call.Envelope.Fields, "GetSecretForCreate(ctx, &request.Config)")
	return &CreateInitializationView{Call: call, Fields: fields}
}

// A nested field cannot simply be overlaid on a possibly absent response
// parent. Fail closed until initialization can preserve that parent's complete
// Terraform value, rather than silently accepting a value Create never sends.
func checkNestedUpdateOnly(attrs []*model.Attribute, create, update *model.Schema, unsupported *[]UnsupportedNode) {
	if create == nil || update == nil {
		return
	}
	for _, attr := range attrs {
		if !attr.Required && !attr.Optional {
			continue
		}
		updateChild, updateDeclared := roleChild(update, attr)
		if !updateDeclared {
			continue
		}
		createChild, createDeclared := roleChild(create, attr)
		if !createDeclared {
			*unsupported = append(*unsupported, UnsupportedNode{
				Path: attr.Path, Reason: "update-only fields nested in a shared create/update attribute require unsupported post-create initialization",
			})
			continue
		}
		checkNestedUpdateOnly(attr.Children, createChild, updateChild, unsupported)
	}
}
