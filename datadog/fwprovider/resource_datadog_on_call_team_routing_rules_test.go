package fwprovider

import (
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnCallTeamRoutingRulesSlackCustomizableContentRequest(t *testing.T) {
	tests := map[string]struct {
		model    slackMessageModel
		expected string
	}{
		"all false omits customizable_content": {
			model: slackMessageModel{
				IncludeDescription: types.BoolValue(false),
				IncludeSource:      types.BoolValue(false),
				MirrorReplies:      types.BoolValue(false),
			},
			expected: `{"channel":"C1","type":"send_slack_message","workspace":"W1"}`,
		},
		"null flags omit customizable_content": {
			model:    slackMessageModel{},
			expected: `{"channel":"C1","type":"send_slack_message","workspace":"W1"}`,
		},
		"mixed flags": {
			model: slackMessageModel{
				IncludeDescription: types.BoolValue(true),
				IncludeSource:      types.BoolValue(false),
				MirrorReplies:      types.BoolValue(true),
			},
			expected: `{"channel":"C1","customizable_content":{"include_description":true,"include_source":false,"mirror_replies":true,"thread_timeline":false},"type":"send_slack_message","workspace":"W1"}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			action := datadogV2.NewSendSlackMessageActionWithDefaults()
			action.Channel = "C1"
			action.Workspace = "W1"
			setSlackCustomizableContent(action, &tc.model)

			body, err := json.Marshal(action)
			require.NoError(t, err)
			assert.JSONEq(t, tc.expected, string(body))
		})
	}
}

func TestOnCallTeamRoutingRulesSlackCustomizableContentState(t *testing.T) {
	tests := map[string]struct {
		payload                        string
		description, source, mirrorRep bool
	}{
		"absent object": {
			payload: `{"channel":"C1","type":"send_slack_message","workspace":"W1"}`,
		},
		"all true": {
			payload:     `{"channel":"C1","type":"send_slack_message","workspace":"W1","customizable_content":{"include_description":true,"include_source":true,"thread_timeline":false,"mirror_replies":true}}`,
			description: true, source: true, mirrorRep: true,
		},
		"partial object": {
			payload: `{"channel":"C1","type":"send_slack_message","workspace":"W1","customizable_content":{"include_source":true}}`,
			source:  true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var action datadogV2.SendSlackMessageAction
			require.NoError(t, json.Unmarshal([]byte(tc.payload), &action))
			require.Nil(t, action.UnparsedObject)

			assert.Equal(t, tc.description, slackCustomizableContentFlag(&action, "include_description").ValueBool())
			assert.Equal(t, tc.source, slackCustomizableContentFlag(&action, "include_source").ValueBool())
			assert.Equal(t, tc.mirrorRep, slackCustomizableContentFlag(&action, "mirror_replies").ValueBool())
		})
	}
}
