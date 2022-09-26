package utils

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ResourceDataKey structure holds Terraform's *schema.ResourceData structure
// and a slice of strings that hold a "key" to somewhere in the stored data.
// For example, []string{"foo", "0", "bar"} would point to "foo.0.bar". This
// allows for easy extraction of the data while adding/removing parts of the "key".
type ResourceDataKey struct {
	parts []string
	d     *schema.ResourceData
}

// NewResourceDataKey creates a new ResourceDataKey with given initial key (can be
// given with dotted notation, e.g. "foo.0.bar").
func NewResourceDataKey(d *schema.ResourceData, initial string) *ResourceDataKey {
	rdk := &ResourceDataKey{}
	rdk.Add(initial)
	rdk.d = d
	return rdk
}

// interfaceToPartsStr converts an interface to string, assuming it's a string or int,
// it panics otherwise.
func (k *ResourceDataKey) interfaceToPartsStr(i interface{}, method string) string {
	switch i := i.(type) {
	case string:
		return i
	case int:
		return fmt.Sprintf("%d", i)
	}

	panic(fmt.Sprintf("ResourceDataKey.%s only accepts string and int argument, got %T\n", method, i))
}

// Add adds new parts to the stored key. The argument can either be a string using
// dotted notation (e.g. "foo.0.bar") or an int (assumed to be a list index and
// converted to string).
func (k *ResourceDataKey) Add(newParts interface{}) *ResourceDataKey {
	newPartsStr := k.interfaceToPartsStr(newParts, "Add")
	newPartsStr = strings.Trim(newPartsStr, ".")

	if len(newPartsStr) > 0 {
		k.parts = append(k.parts, strings.Split(newPartsStr, ".")...)
	}
	return k
}

// Remove like Add, but removes given key parts from the end of the key. Panics if the
// key doesn't end with given parts.
func (k *ResourceDataKey) Remove(parts interface{}) *ResourceDataKey {
	removePartsStr := strings.Trim(k.interfaceToPartsStr(parts, "Remove"), ".")
	splitParts := strings.Split(removePartsStr, ".")
	// represents index in k.parts which should match the first item of splitParts
	kPartsIndexStart := len(k.parts) - 1 - (len(splitParts) - 1)
	for i, p := range splitParts {
		if k.parts[kPartsIndexStart+i] != p {
			panic(fmt.Sprintf("%v doesn't end with %v\n", k.parts, splitParts))
		}
	}
	k.parts = k.parts[:kPartsIndexStart]
	return k
}

// Pop will remove given count of key parts from the end of the key. Panics if given count
// is bigger than number of key parts.
func (k *ResourceDataKey) Pop(removeCount int) *ResourceDataKey {
	if len(k.parts) < removeCount {
		panic(fmt.Sprintf("Trying to remove %d components from %s\n", removeCount, k.S()))
	} else {
		k.parts = k.parts[:len(k.parts)-removeCount]
	}
	return k
}

// S returns string representation of the key, e.g. "foo.0.bar".
func (k *ResourceDataKey) S() string {
	return strings.Join(k.parts, ".")
}

// With returns string representation of the key (much like S) with given parts added,
// but doesn't add the parts permanently.
func (k *ResourceDataKey) With(parts interface{}) string {
	partsStr := k.interfaceToPartsStr(parts, "With")
	popCount := 0
	if len(partsStr) > 0 {
		popCount = strings.Count(partsStr, ".") + 1
	}
	k.Add(partsStr)
	defer k.Pop(popCount)

	return k.S()
}

// Get calls the "Get" method on the stored ResourceData structure with
// the current key (obtained by S).
func (k *ResourceDataKey) Get() interface{} {
	return k.d.Get(k.S())
}

// GetWith calls the "Get" method on the stored ResourceData structure with
// the current key plus given parts (obtained by With, meaning that the parts
// are not added permanently).
func (k *ResourceDataKey) GetWith(parts interface{}) interface{} {
	return k.d.Get(k.With(parts))
}

// GetOk calls the "GetOk" method on the stored ResourceData structure with
// the current key (obtained by S).
func (k *ResourceDataKey) GetOk() (interface{}, bool) {
	return k.d.GetOk(k.S())
}

// GetOkWith calls the "GetOk" method on the stored ResourceData structure with
// the current key plus given parts (obtained by With, meaning that the parts
// are not added permanently).
func (k *ResourceDataKey) GetOkWith(parts interface{}) (interface{}, bool) {
	return k.d.GetOk(k.With(parts))
}
