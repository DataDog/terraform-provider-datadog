package observability_pipeline

// BaseProcessor interface defines common fields that all processors have
// Used for both flatten (Get methods) and expand (Set methods) operations
type BaseProcessor interface {
	// Get methods for flatten (API -> Terraform)
	GetId() string
	GetEnabled() bool
	GetInclude() string
	GetDisplayName() string
	GetDisplayNameOk() (*string, bool)
	// Set methods for expand (Terraform -> API)
	SetId(string)
	SetEnabled(bool)
	SetInclude(string)
	SetDisplayName(string)
}

// BaseProcessorFields holds the common fields shared by all processors
type BaseProcessorFields struct {
	Id          string
	Enabled     bool
	Include     string
	DisplayName *string
}

// ApplyTo sets the common fields on any processor
func (c BaseProcessorFields) ApplyTo(proc BaseProcessor) {
	proc.SetId(c.Id)
	proc.SetEnabled(c.Enabled)
	proc.SetInclude(c.Include)
	if c.DisplayName != nil {
		proc.SetDisplayName(*c.DisplayName)
	}
}
