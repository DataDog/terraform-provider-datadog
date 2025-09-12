package fwutils

import (
	"context"
	"sort"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToTerraformStr(v any, ok bool) types.String {
	if !ok || v == nil {
		return types.StringNull()
	}
	switch t := v.(type) {
	case *int32:
		return types.StringValue(strconv.FormatInt(int64(*t), 10))
	case *float64:
		return types.StringValue(strconv.FormatFloat(*t, 'f', -1, 64))
	case *datadog.NullableFloat64:
		if !t.IsSet() || t.Get() == nil {
			return types.StringNull()
		}
		return ToTerraformStr(t.Get(), true)
	case *datadog.NullableString:
		if !t.IsSet() || t.Get() == nil {
			return types.StringNull()
		}
		return ToTerraformStr(t.Get(), true)
	case *string:
		return types.StringValue(*t)
	}
	return types.StringNull()
}

func ToTerraformBool(v *bool, ok bool) types.Bool {
	if ok && v != nil {
		return types.BoolValue(*v)
	}
	return types.BoolNull()
}

func ToTerraformInt32(v *int32, ok bool) types.Int32 {
	if ok && v != nil {
		return types.Int32Value(*v)
	}
	return types.Int32Null()
}

func ToTerraformInt64(v *int64, ok bool) types.Int64 {
	if ok && v != nil {
		return types.Int64Value(*v)
	}
	return types.Int64Null()
}

func ToTerraformSetString(ctx context.Context, get func() (*[]string, bool)) types.Set {
	if v, ok := get(); ok && v != nil {
		result, _ := types.SetValueFrom(ctx, types.StringType, v)
		return result
	}
	return types.SetNull(types.StringType)
}

func SetOptString(s types.String, set func(string)) {
	if !s.IsNull() && !s.IsUnknown() {
		set(s.ValueString())
	}
}

func SetOptInt32(i types.Int32, set func(int32)) {
	if !i.IsNull() && !i.IsUnknown() {
		set(i.ValueInt32())
	}
}

func SetOptInt64(i types.Int64, set func(int64)) {
	if !i.IsNull() && !i.IsUnknown() {
		set(i.ValueInt64())
	}
}

func SetOptBool(b types.Bool, set func(bool)) {
	if !b.IsNull() && !b.IsUnknown() {
		set(b.ValueBool())
	}
}

func SetOptStringList(typeCollection any, set func([]string), ctx context.Context) {
	var strList []string
	switch t := typeCollection.(type) {
	case types.Set:
		if !t.IsNull() && !t.IsUnknown() {
			t.ElementsAs(ctx, &strList, false)
			sort.Strings(strList)
			set(strList)
		}
	case types.List:
		if !t.IsNull() && !t.IsUnknown() {
			t.ElementsAs(ctx, &strList, false)
			set(strList)
		}
	}
}
