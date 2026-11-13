package customtypes

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.StringValuable = TagFilterExpressionStringValue{}
var _ basetypes.StringValuableWithSemanticEquals = TagFilterExpressionStringValue{}

type TagFilterExpressionStringValue struct {
	basetypes.StringValue
}

func (v TagFilterExpressionStringValue) Equal(o attr.Value) bool {
	other, ok := o.(TagFilterExpressionStringValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v TagFilterExpressionStringValue) Type(ctx context.Context) attr.Type {
	return TagFilterExpressionStringType{}
}

func (v TagFilterExpressionStringValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(TagFilterExpressionStringValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}
	return TagFilterExpressionsEqual(v.ValueString(), newValue.ValueString()), diags
}

// TagFilterExpressionsEqual reports whether two AWS tag filter expressions are equal after the
// normalization the API applies: "!" is returned as "NOT" and AND operands are reordered.
// Operand order is ignored only for expressions without OR or parentheses, where AND is
// commutative without precedence concerns.
func TagFilterExpressionsEqual(a, b string) bool {
	ta, tb := tokenizeTagFilterExpression(a), tokenizeTagFilterExpression(b)
	if !isPlainConjunction(ta) || !isPlainConjunction(tb) {
		return strings.Join(ta, " ") == strings.Join(tb, " ")
	}
	oa, ob := conjunctionOperands(ta), conjunctionOperands(tb)
	if len(oa) != len(ob) {
		return false
	}
	sort.Strings(oa)
	sort.Strings(ob)
	for i := range oa {
		if oa[i] != ob[i] {
			return false
		}
	}
	return true
}

func tokenizeTagFilterExpression(s string) []string {
	s = strings.NewReplacer("(", " ( ", ")", " ) ").Replace(s)
	var tokens []string
	for _, f := range strings.Fields(s) {
		for strings.HasPrefix(f, "!") {
			tokens = append(tokens, "NOT")
			f = f[1:]
		}
		if f == "" {
			continue
		}
		switch u := strings.ToUpper(f); u {
		case "AND", "OR", "NOT":
			tokens = append(tokens, u)
		default:
			tokens = append(tokens, f)
		}
	}
	return tokens
}

func isPlainConjunction(tokens []string) bool {
	for _, t := range tokens {
		if t == "OR" || t == "(" || t == ")" {
			return false
		}
	}
	return true
}

func conjunctionOperands(tokens []string) []string {
	var operands []string
	var cur []string
	for _, t := range tokens {
		if t == "AND" {
			operands = append(operands, strings.Join(cur, " "))
			cur = nil
			continue
		}
		cur = append(cur, t)
	}
	return append(operands, strings.Join(cur, " "))
}
