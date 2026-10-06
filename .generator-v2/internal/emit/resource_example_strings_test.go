package emit

import (
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestExampleWithBackticksRendersValidGo(t *testing.T) {
	scenario := collectionScenario(t, &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"}, "`code` ${var.literal}")
	view := ResourceView{SDKPackage: "datadogV2", APIAccessor: "GetWidgetsApiV2", Read: CRUDCallView{Method: "GetWidget", Arguments: []SDKArgumentView{{TFName: "id", GoType: "string"}}}, Schema: SchemaView{Attributes: []AttrView{{TFName: "field", Required: true}}}}
	if _, err := RenderResourceExampleTest(scenario, view, map[string]string{"field": "field"}); err != nil {
		t.Fatal(err)
	}
}
