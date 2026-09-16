package fwprovider

import "github.com/hashicorp/terraform-plugin-framework/resource"

// generatedResources holds the resources produced by the generator-v2 emit
// pipeline. tfgen owns this file: each run merges the constructors it produced
// into the existing set (union, sorted) so a scoped --include run never drops
// resources it did not regenerate. Do not edit by hand.
//
// FrameworkProvider.Resources registers this slice alongside the hand-written
// Resources.
var generatedResources = []func() resource.Resource{}
