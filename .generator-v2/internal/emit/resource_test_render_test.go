package emit

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// twilioResourceView is the minimum ResourceView a destroy check needs: the
// API accessor and a read taking exactly the resource id.
func twilioResourceView() ResourceView {
	return ResourceView{
		TypeName:    "datadog_integration_twilio_account",
		GoName:      "IntegrationTwilioAccount",
		SDKPackage:  "datadogV2",
		APIAccessor: "GetTwilioIntegrationApiV2",
		Read: CRUDCallView{
			Method:    "GetTwilioIntegrationAccount",
			Arguments: []SDKArgumentView{{Expression: "state.Id.ValueString()", TFName: "id"}},
		},
		Schema: twilioSchemaView(),
	}
}

var _ = Describe("BuildResourceTestView", func() {
	build := func() resourceTestView {
		_, scenario := twilioExampleView()
		view, err := BuildResourceTestView(scenario, twilioResourceView())
		Expect(err).To(Succeed())
		return view
	}

	It("carries the common view through", func() {
		view := build()
		Expect(view.FuncName).To(Equal("TestAccDatadogIntegrationTwilioAccountOpenAPIExample"))
		Expect(view.Marker).To(Equal(model.GeneratedMarker))
		Expect(view.Steps).To(HaveLen(2))
		Expect(view.Steps[0].ConfigBody).To(ContainSubstring(`name = "twilio-prod"`))
	})

	It("names the destroy helper as a pair with the test, unexported", func() {
		view := build()
		Expect(view.DestroyCheckFunc).To(
			Equal("testAccDatadogIntegrationTwilioAccountOpenAPIExampleDestroy"))
		Expect(view.DestroyCheckFunc).NotTo(HavePrefix("Test"))
		Expect(view.DestroyCheckFunc).NotTo(ContainSubstring("Config"))
	})

	It("carries how the destroy check reaches the API", func() {
		view := build()
		Expect(view.APIAccessor).To(Equal("GetTwilioIntegrationApiV2"))
		Expect(view.ReadMethod).To(Equal("GetTwilioIntegrationAccount"))
		Expect(view.SDKPackage).To(Equal("datadogV2"))
	})

	It("labels the update step when the scenario has one", func() {
		view := build()
		Expect(view.UpdateStepIndex).To(Equal(2))
	})

	It("reports no update step when the scenario has none", func() {
		_, scenario := twilioExampleView()
		scenario.Steps = scenario.Steps[:1]
		view, err := BuildResourceTestView(scenario, twilioResourceView())
		Expect(err).To(Succeed())
		Expect(view.UpdateStepIndex).To(BeZero())
	})

	Describe("rejected views", func() {
		// Without a read there is nothing to ask, so the scenario's 404
		// interaction could never be consumed.
		It("reports a resource with no SDK read call", func() {
			_, scenario := twilioExampleView()
			view := twilioResourceView()
			view.Read.Method = ""
			_, err := BuildResourceTestView(scenario, view)
			Expect(err).To(MatchError(ContainSubstring("no SDK read call")))
		})

		// A CheckDestroy has only resource.Primary.ID to work from. Rendering
		// the test without the check would leave the 404 interaction
		// unconsumed, so the target is reported rather than half-built.
		It("reports a read needing more than the resource id", func() {
			_, scenario := twilioExampleView()
			view := twilioResourceView()
			view.Read.Arguments = append(view.Read.Arguments,
				SDKArgumentView{Expression: "state.Parent.ValueString()", TFName: "parent_id"})
			_, err := BuildResourceTestView(scenario, view)
			Expect(err).To(MatchError(ContainSubstring("has only the resource id")))
			Expect(err.Error()).To(ContainSubstring("takes 2 arguments"))
		})

		It("reports a read taking no arguments at all", func() {
			_, scenario := twilioExampleView()
			view := twilioResourceView()
			view.Read.Arguments = nil
			_, err := BuildResourceTestView(scenario, view)
			Expect(err).To(MatchError(ContainSubstring("takes 0 arguments")))
		})

		It("propagates a scenario that does not validate", func() {
			_, err := BuildResourceTestView(&model.GeneratedTestScenario{}, twilioResourceView())
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("destroyCheckFuncName", func() {
	It("lower-cases the leading Test and appends Destroy", func() {
		Expect(destroyCheckFuncName("TestAccDatadogWidgetOpenAPIExample")).
			To(Equal("testAccDatadogWidgetOpenAPIExampleDestroy"))
	})

	It("prefixes a name off-convention rather than slicing it", func() {
		Expect(destroyCheckFuncName("Odd")).To(Equal("testAccOddDestroy"))
	})
})
