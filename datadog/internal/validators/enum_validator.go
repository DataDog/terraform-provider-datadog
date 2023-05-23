package validators

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func NewEnumValidator[T any](i interface{}) enumValidator[T] {
	f := reflect.TypeOf(i)
	enum := reflect.New(f.Out(0)).Elem()
	allowedValues := enum.MethodByName("GetAllowedValues").Call([]reflect.Value{})[0].Interface()

	return enumValidator[T]{
		enumFunc:          i,
		AllowedEnumValues: allowedValues,
	}
}

type enumValidator[T any] struct {
	enumFunc          interface{}
	AllowedEnumValues interface{}
}

func (v enumValidator[T]) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v enumValidator[T]) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value must be valid one of: %v", v.AllowedEnumValues)
}

func (v enumValidator[T]) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	err := v.validateHelper(req.ConfigValue.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("invalid value", fmt.Sprintf("invalid value for \"%s\". valid values are %v", req.Path.String(), v.AllowedEnumValues))
	}
}

func (v enumValidator[T]) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	err := v.validateHelper(req.ConfigValue.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("invalid value", fmt.Sprintf("invalid value for \"%s\". valid values are %v", req.Path.String(), v.AllowedEnumValues))
	}
}

func (v enumValidator[T]) validateHelper(value interface{}) error {
	argT := reflect.TypeOf(v.enumFunc).In(0)
	outs := reflect.ValueOf(v.enumFunc).Call([]reflect.Value{reflect.ValueOf(value).Convert(argT)})
	if err := outs[1].Interface(); err != nil {
		return err.(error)
	}

	return nil
}
