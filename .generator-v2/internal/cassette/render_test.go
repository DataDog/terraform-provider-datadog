package cassette

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	govcr "gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/yaml.v3"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func twilioScenario() *model.GeneratedTestScenario {
	scenario, err := BuildResourceScenario(twilioTarget())
	Expect(err).To(Succeed())
	return scenario
}

var _ = Describe("RenderCassette", func() {
	// The decisive property: go-vcr must be able to load what we emit. The
	// format is its contract, so this asserts against its own loader rather
	// than against a shape we believe it wants.
	It("produces a cassette go-vcr itself can load and replay in order", func() {
		scenario := twilioScenario()
		out, err := RenderCassette(scenario)
		Expect(err).To(Succeed())

		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "TestAccGenerated.yaml")
		Expect(os.WriteFile(path, out, 0o644)).To(Succeed())

		loaded, err := govcr.Load(strings.TrimSuffix(path, ".yaml"))
		Expect(err).To(Succeed())
		Expect(loaded.Version).To(Equal(2))
		Expect(loaded.Interactions).To(HaveLen(len(scenario.Interactions)))

		for i, interaction := range loaded.Interactions {
			Expect(interaction.ID).To(Equal(i), "interaction ids must be dense and ordered")
		}
		Expect(loaded.Interactions[0].Request.Method).To(Equal("POST"))
		last := loaded.Interactions[len(loaded.Interactions)-1]
		Expect(last.Response.Code).To(Equal(404), "the trace must end on the destroy verification")
	})

	It("carries the document separator and version a recorded cassette has", func() {
		out, err := RenderCassette(twilioScenario())
		Expect(err).To(Succeed())
		Expect(string(out)).To(HavePrefix("---\nversion: 2\ninteractions:"))
	})

	Describe("one interaction's recorded shape", func() {
		var first, del map[string]any
		BeforeEach(func() {
			out, err := RenderCassette(twilioScenario())
			Expect(err).To(Succeed())
			var doc struct {
				Interactions []map[string]any `yaml:"interactions"`
			}
			Expect(yaml.Unmarshal(out, &doc)).To(Succeed())
			first = doc.Interactions[0]
			// Located by method rather than position: this test is about the
			// recorded shape of an interaction, not how many reads precede it.
			for _, interaction := range doc.Interactions {
				if interaction["request"].(map[string]any)["method"] == "DELETE" {
					del = interaction
				}
			}
			Expect(del).NotTo(BeNil(), "no DELETE interaction in the rendered cassette")
		})

		It("records the request the way the recorder would", func() {
			request := first["request"].(map[string]any)
			Expect(request["method"]).To(Equal("POST"))
			Expect(request["host"]).To(Equal("api.datadoghq.com"))
			Expect(request["proto"]).To(Equal("HTTP/1.1"))
			Expect(request["content_length"]).To(Equal(len(request["body"].(string))))
			// Empty rather than null, matching a recorded interaction.
			Expect(request["form"]).To(BeEmpty())
			Expect(request["trailer"]).To(BeEmpty())
		})

		It("reports a streamed body as an unknown length, already decompressed", func() {
			response := first["response"].(map[string]any)
			Expect(response["content_length"]).To(Equal(-1))
			Expect(response["uncompressed"]).To(BeTrue())
			Expect(response["status"]).To(Equal("201 Created"))
			Expect(response["code"]).To(Equal(201))
		})

		// A declared bodyless response reports neither a length nor a
		// decompressed body, which is what the recorder writes for a 204.
		It("reports a bodyless response as zero length and not decompressed", func() {
			response := del["response"].(map[string]any)
			Expect(del["request"].(map[string]any)["method"]).To(Equal("DELETE"))
			Expect(response["body"]).To(BeEmpty())
			Expect(response["content_length"]).To(Equal(0))
			Expect(response["uncompressed"]).To(BeFalse())
		})

		// A recorded latency would make regeneration non-deterministic while
		// changing nothing about replay.
		It("records every duration as zero", func() {
			scenario := twilioScenario()
			out, err := RenderCassette(scenario)
			Expect(err).To(Succeed())
			Expect(strings.Count(string(out), "duration: 0s")).
				To(Equal(len(scenario.Interactions)))
		})
	})

	It("produces identical bytes on repeated renders", func() {
		first, err := RenderCassette(twilioScenario())
		Expect(err).To(Succeed())
		second, err := RenderCassette(twilioScenario())
		Expect(err).To(Succeed())
		Expect(first).To(Equal(second))
	})

	It("never writes a declared secret into the fixture", func() {
		out, err := RenderCassette(twilioScenario())
		Expect(err).To(Succeed())
		Expect(string(out)).NotTo(ContainSubstring("twilio-basic-auth-secret"))
		Expect(string(out)).To(ContainSubstring(model.RedactedPlaceholder))
	})

	Describe("rejected input", func() {
		It("propagates a scenario that does not validate", func() {
			_, err := RenderCassette(&model.GeneratedTestScenario{})
			Expect(err).To(HaveOccurred())
		})

		// A recorded request carries its host alongside the URL, so a URL
		// without one cannot be rendered.
		It("reports a request URL carrying no host", func() {
			scenario := twilioScenario()
			scenario.Interactions[0].Request.URL = "/api/v2/relative"
			_, err := RenderCassette(scenario)
			Expect(err).To(MatchError(ContainSubstring("carries no host")))
			Expect(err.Error()).To(ContainSubstring("interaction 0"))
		})
	})
})

var _ = Describe("RenderFreeze", func() {
	// restoreClock parses the companion with time.RFC3339Nano, so the written
	// form has to round-trip through exactly that.
	It("writes the fixed instant in the form the harness parses", func() {
		out, err := RenderFreeze(twilioScenario())
		Expect(err).To(Succeed())

		parsed, err := time.Parse(time.RFC3339Nano, string(out))
		Expect(err).To(Succeed())
		Expect(parsed.UTC()).To(Equal(frozen))
	})

	It("writes no trailing newline, matching what setClock writes", func() {
		out, err := RenderFreeze(twilioScenario())
		Expect(err).To(Succeed())
		Expect(string(out)).NotTo(HaveSuffix("\n"))
	})

	It("propagates a scenario that does not validate", func() {
		_, err := RenderFreeze(&model.GeneratedTestScenario{})
		Expect(err).To(HaveOccurred())
	})
})
