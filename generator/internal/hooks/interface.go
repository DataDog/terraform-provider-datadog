package hooks

// DataSourceHooks defines extension points for generated data sources.
type DataSourceHooks interface {
	BeforeRead() error
	AfterRead() error
	ModifySchema()
}

// NoOpDataSourceHooks is a default implementation that does nothing.
type NoOpDataSourceHooks struct{}

func (n *NoOpDataSourceHooks) BeforeRead() error { return nil }
func (n *NoOpDataSourceHooks) AfterRead() error  { return nil }
func (n *NoOpDataSourceHooks) ModifySchema()     {}
