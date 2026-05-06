package hooks

import "testing"

func TestNoOpDataSourceHooks_SatisfiesInterface(t *testing.T) {
	var h DataSourceHooks = &NoOpDataSourceHooks{}
	if h == nil {
		t.Fatal("expected non-nil hooks")
	}
}

func TestNoOpDataSourceHooks_BeforeRead(t *testing.T) {
	h := &NoOpDataSourceHooks{}
	if err := h.BeforeRead(); err != nil {
		t.Errorf("BeforeRead() returned error: %v", err)
	}
}

func TestNoOpDataSourceHooks_AfterRead(t *testing.T) {
	h := &NoOpDataSourceHooks{}
	if err := h.AfterRead(); err != nil {
		t.Errorf("AfterRead() returned error: %v", err)
	}
}

func TestNoOpDataSourceHooks_ModifySchema(t *testing.T) {
	h := &NoOpDataSourceHooks{}
	h.ModifySchema() // Should not panic
}
