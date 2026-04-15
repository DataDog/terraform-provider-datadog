package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// KubernetesLogsSourceModel represents the Terraform model for kubernetes_logs source configuration
type KubernetesLogsSourceModel struct {
	AutoPartialMerge             types.Bool                                        `tfsdk:"auto_partial_merge"`
	DelayDeletionMs              types.Int64                                       `tfsdk:"delay_deletion_ms"`
	ExcludePathsGlobPatterns     []types.String                                    `tfsdk:"exclude_paths_glob_patterns"`
	ExtraFieldSelector           types.String                                      `tfsdk:"extra_field_selector"`
	ExtraLabelSelector           types.String                                      `tfsdk:"extra_label_selector"`
	ExtraNamespaceLabelSelector  types.String                                      `tfsdk:"extra_namespace_label_selector"`
	IgnoreOlderSecs              types.Int64                                       `tfsdk:"ignore_older_secs"`
	IncludePathsGlobPatterns     []types.String                                    `tfsdk:"include_paths_glob_patterns"`
	InsertNamespaceFields        types.Bool                                        `tfsdk:"insert_namespace_fields"`
	KubeConfigFile               types.String                                      `tfsdk:"kube_config_file"`
	ReadFrom                     types.String                                      `tfsdk:"read_from"`
	Timezone                     types.String                                      `tfsdk:"timezone"`
	UseApiserverCache            types.Bool                                        `tfsdk:"use_apiserver_cache"`
	NamespaceAnnotationFields    []KubernetesLogsNamespaceAnnotationFieldsModel    `tfsdk:"namespace_annotation_fields"`
	NodeAnnotationFields         []KubernetesLogsNodeAnnotationFieldsModel         `tfsdk:"node_annotation_fields"`
	PodAnnotationFields          []KubernetesLogsPodAnnotationFieldsModel          `tfsdk:"pod_annotation_fields"`
}

// KubernetesLogsNamespaceAnnotationFieldsModel represents namespace annotation fields
type KubernetesLogsNamespaceAnnotationFieldsModel struct {
	NamespaceLabels types.String `tfsdk:"namespace_labels"`
}

// KubernetesLogsNodeAnnotationFieldsModel represents node annotation fields
type KubernetesLogsNodeAnnotationFieldsModel struct {
	NodeLabels types.String `tfsdk:"node_labels"`
}

// KubernetesLogsPodAnnotationFieldsModel represents pod annotation fields
type KubernetesLogsPodAnnotationFieldsModel struct {
	ContainerId      types.String `tfsdk:"container_id"`
	ContainerImage   types.String `tfsdk:"container_image"`
	ContainerImageId types.String `tfsdk:"container_image_id"`
	ContainerName    types.String `tfsdk:"container_name"`
	PodAnnotations   types.String `tfsdk:"pod_annotations"`
	PodIp            types.String `tfsdk:"pod_ip"`
	PodIps           types.String `tfsdk:"pod_ips"`
	PodLabels        types.String `tfsdk:"pod_labels"`
	PodName          types.String `tfsdk:"pod_name"`
	PodNamespace     types.String `tfsdk:"pod_namespace"`
	PodNodeName      types.String `tfsdk:"pod_node_name"`
	PodOwner         types.String `tfsdk:"pod_owner"`
	PodUid           types.String `tfsdk:"pod_uid"`
}

// ExpandKubernetesLogsSource converts the Terraform model to the Datadog API model
func ExpandKubernetesLogsSource(src *KubernetesLogsSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineKubernetesLogsSourceWithDefaults()
	s.SetId(id)

	if !src.AutoPartialMerge.IsNull() {
		s.SetAutoPartialMerge(src.AutoPartialMerge.ValueBool())
	}
	if !src.DelayDeletionMs.IsNull() {
		s.SetDelayDeletionMs(src.DelayDeletionMs.ValueInt64())
	}
	if src.ExcludePathsGlobPatterns != nil {
		patterns := make([]string, len(src.ExcludePathsGlobPatterns))
		for i, p := range src.ExcludePathsGlobPatterns {
			patterns[i] = p.ValueString()
		}
		s.SetExcludePathsGlobPatterns(patterns)
	}
	if !src.ExtraFieldSelector.IsNull() {
		s.SetExtraFieldSelector(src.ExtraFieldSelector.ValueString())
	}
	if !src.ExtraLabelSelector.IsNull() {
		s.SetExtraLabelSelector(src.ExtraLabelSelector.ValueString())
	}
	if !src.ExtraNamespaceLabelSelector.IsNull() {
		s.SetExtraNamespaceLabelSelector(src.ExtraNamespaceLabelSelector.ValueString())
	}
	if !src.IgnoreOlderSecs.IsNull() {
		s.SetIgnoreOlderSecs(src.IgnoreOlderSecs.ValueInt64())
	}
	if src.IncludePathsGlobPatterns != nil {
		patterns := make([]string, len(src.IncludePathsGlobPatterns))
		for i, p := range src.IncludePathsGlobPatterns {
			patterns[i] = p.ValueString()
		}
		s.SetIncludePathsGlobPatterns(patterns)
	}
	if !src.InsertNamespaceFields.IsNull() {
		s.SetInsertNamespaceFields(src.InsertNamespaceFields.ValueBool())
	}
	if !src.KubeConfigFile.IsNull() {
		s.SetKubeConfigFile(src.KubeConfigFile.ValueString())
	}
	if !src.ReadFrom.IsNull() {
		s.SetReadFrom(datadogV2.ObservabilityPipelineKubernetesLogsSourceReadFrom(src.ReadFrom.ValueString()))
	}
	if !src.Timezone.IsNull() {
		s.SetTimezone(src.Timezone.ValueString())
	}
	if !src.UseApiserverCache.IsNull() {
		s.SetUseApiserverCache(src.UseApiserverCache.ValueBool())
	}

	if len(src.NamespaceAnnotationFields) > 0 {
		nsFields := datadogV2.NewObservabilityPipelineKubernetesLogsSourceNamespaceAnnotationFieldsWithDefaults()
		if !src.NamespaceAnnotationFields[0].NamespaceLabels.IsNull() {
			nsFields.SetNamespaceLabels(src.NamespaceAnnotationFields[0].NamespaceLabels.ValueString())
		}
		s.SetNamespaceAnnotationFields(*nsFields)
	}

	if len(src.NodeAnnotationFields) > 0 {
		nodeFields := datadogV2.NewObservabilityPipelineKubernetesLogsSourceNodeAnnotationFieldsWithDefaults()
		if !src.NodeAnnotationFields[0].NodeLabels.IsNull() {
			nodeFields.SetNodeLabels(src.NodeAnnotationFields[0].NodeLabels.ValueString())
		}
		s.SetNodeAnnotationFields(*nodeFields)
	}

	if len(src.PodAnnotationFields) > 0 {
		podFields := datadogV2.NewObservabilityPipelineKubernetesLogsSourcePodAnnotationFieldsWithDefaults()
		pod := src.PodAnnotationFields[0]
		if !pod.ContainerId.IsNull() {
			podFields.SetContainerId(pod.ContainerId.ValueString())
		}
		if !pod.ContainerImage.IsNull() {
			podFields.SetContainerImage(pod.ContainerImage.ValueString())
		}
		if !pod.ContainerImageId.IsNull() {
			podFields.SetContainerImageId(pod.ContainerImageId.ValueString())
		}
		if !pod.ContainerName.IsNull() {
			podFields.SetContainerName(pod.ContainerName.ValueString())
		}
		if !pod.PodAnnotations.IsNull() {
			podFields.SetPodAnnotations(pod.PodAnnotations.ValueString())
		}
		if !pod.PodIp.IsNull() {
			podFields.SetPodIp(pod.PodIp.ValueString())
		}
		if !pod.PodIps.IsNull() {
			podFields.SetPodIps(pod.PodIps.ValueString())
		}
		if !pod.PodLabels.IsNull() {
			podFields.SetPodLabels(pod.PodLabels.ValueString())
		}
		if !pod.PodName.IsNull() {
			podFields.SetPodName(pod.PodName.ValueString())
		}
		if !pod.PodNamespace.IsNull() {
			podFields.SetPodNamespace(pod.PodNamespace.ValueString())
		}
		if !pod.PodNodeName.IsNull() {
			podFields.SetPodNodeName(pod.PodNodeName.ValueString())
		}
		if !pod.PodOwner.IsNull() {
			podFields.SetPodOwner(pod.PodOwner.ValueString())
		}
		if !pod.PodUid.IsNull() {
			podFields.SetPodUid(pod.PodUid.ValueString())
		}
		s.SetPodAnnotationFields(*podFields)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineKubernetesLogsSource: s,
	}
}

// FlattenKubernetesLogsSource converts the Datadog API model to the Terraform model
func FlattenKubernetesLogsSource(src *datadogV2.ObservabilityPipelineKubernetesLogsSource) *KubernetesLogsSourceModel {
	if src == nil {
		return nil
	}

	out := &KubernetesLogsSourceModel{}

	if v, ok := src.GetAutoPartialMergeOk(); ok {
		out.AutoPartialMerge = types.BoolValue(*v)
	}
	if v, ok := src.GetDelayDeletionMsOk(); ok {
		out.DelayDeletionMs = types.Int64Value(*v)
	}

	excludePatterns := []types.String{}
	for _, p := range src.GetExcludePathsGlobPatterns() {
		excludePatterns = append(excludePatterns, types.StringValue(p))
	}
	if len(excludePatterns) > 0 {
		out.ExcludePathsGlobPatterns = excludePatterns
	}

	if v, ok := src.GetExtraFieldSelectorOk(); ok {
		out.ExtraFieldSelector = types.StringValue(*v)
	}
	if v, ok := src.GetExtraLabelSelectorOk(); ok {
		out.ExtraLabelSelector = types.StringValue(*v)
	}
	if v, ok := src.GetExtraNamespaceLabelSelectorOk(); ok {
		out.ExtraNamespaceLabelSelector = types.StringValue(*v)
	}
	if v, ok := src.GetIgnoreOlderSecsOk(); ok {
		out.IgnoreOlderSecs = types.Int64Value(*v)
	}

	includePatterns := []types.String{}
	for _, p := range src.GetIncludePathsGlobPatterns() {
		includePatterns = append(includePatterns, types.StringValue(p))
	}
	if len(includePatterns) > 0 {
		out.IncludePathsGlobPatterns = includePatterns
	}

	if v, ok := src.GetInsertNamespaceFieldsOk(); ok {
		out.InsertNamespaceFields = types.BoolValue(*v)
	}
	if v, ok := src.GetKubeConfigFileOk(); ok {
		out.KubeConfigFile = types.StringValue(*v)
	}
	if v, ok := src.GetReadFromOk(); ok {
		out.ReadFrom = types.StringValue(string(*v))
	}
	if v, ok := src.GetTimezoneOk(); ok {
		out.Timezone = types.StringValue(*v)
	}
	if v, ok := src.GetUseApiserverCacheOk(); ok {
		out.UseApiserverCache = types.BoolValue(*v)
	}

	if nsFields, ok := src.GetNamespaceAnnotationFieldsOk(); ok {
		model := KubernetesLogsNamespaceAnnotationFieldsModel{}
		if v, ok := nsFields.GetNamespaceLabelsOk(); ok {
			model.NamespaceLabels = types.StringValue(*v)
		}
		out.NamespaceAnnotationFields = []KubernetesLogsNamespaceAnnotationFieldsModel{model}
	}

	if nodeFields, ok := src.GetNodeAnnotationFieldsOk(); ok {
		model := KubernetesLogsNodeAnnotationFieldsModel{}
		if v, ok := nodeFields.GetNodeLabelsOk(); ok {
			model.NodeLabels = types.StringValue(*v)
		}
		out.NodeAnnotationFields = []KubernetesLogsNodeAnnotationFieldsModel{model}
	}

	if podFields, ok := src.GetPodAnnotationFieldsOk(); ok {
		model := KubernetesLogsPodAnnotationFieldsModel{}
		if v, ok := podFields.GetContainerIdOk(); ok {
			model.ContainerId = types.StringValue(*v)
		}
		if v, ok := podFields.GetContainerImageOk(); ok {
			model.ContainerImage = types.StringValue(*v)
		}
		if v, ok := podFields.GetContainerImageIdOk(); ok {
			model.ContainerImageId = types.StringValue(*v)
		}
		if v, ok := podFields.GetContainerNameOk(); ok {
			model.ContainerName = types.StringValue(*v)
		}
		if v, ok := podFields.GetPodAnnotationsOk(); ok {
			model.PodAnnotations = types.StringValue(*v)
		}
		if v, ok := podFields.GetPodIpOk(); ok {
			model.PodIp = types.StringValue(*v)
		}
		if v, ok := podFields.GetPodIpsOk(); ok {
			model.PodIps = types.StringValue(*v)
		}
		if v, ok := podFields.GetPodLabelsOk(); ok {
			model.PodLabels = types.StringValue(*v)
		}
		if v, ok := podFields.GetPodNameOk(); ok {
			model.PodName = types.StringValue(*v)
		}
		if v, ok := podFields.GetPodNamespaceOk(); ok {
			model.PodNamespace = types.StringValue(*v)
		}
		if v, ok := podFields.GetPodNodeNameOk(); ok {
			model.PodNodeName = types.StringValue(*v)
		}
		if v, ok := podFields.GetPodOwnerOk(); ok {
			model.PodOwner = types.StringValue(*v)
		}
		if v, ok := podFields.GetPodUidOk(); ok {
			model.PodUid = types.StringValue(*v)
		}
		out.PodAnnotationFields = []KubernetesLogsPodAnnotationFieldsModel{model}
	}

	return out
}

// KubernetesLogsSourceSchema returns the schema for kubernetes_logs source
func KubernetesLogsSourceSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `kubernetes_logs` source collects logs from Kubernetes pods running on the same node.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"auto_partial_merge": schema.BoolAttribute{
					Optional:    true,
					Description: "Whether to automatically merge partial events split by container runtime.",
				},
				"delay_deletion_ms": schema.Int64Attribute{
					Optional:    true,
					Description: "Milliseconds to delay removing pod metadata after a deletion event.",
				},
				"exclude_paths_glob_patterns": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
					Description: "Glob patterns to exclude from file reading.",
				},
				"extra_field_selector": schema.StringAttribute{
					Optional:    true,
					Description: "Field selector to filter pods.",
				},
				"extra_label_selector": schema.StringAttribute{
					Optional:    true,
					Description: "Label selector to filter pods.",
				},
				"extra_namespace_label_selector": schema.StringAttribute{
					Optional:    true,
					Description: "Label selector to filter namespaces.",
				},
				"ignore_older_secs": schema.Int64Attribute{
					Optional:    true,
					Description: "Ignore files older than this many seconds.",
				},
				"include_paths_glob_patterns": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
					Description: "Glob patterns to include for file reading.",
				},
				"insert_namespace_fields": schema.BoolAttribute{
					Optional:    true,
					Description: "Enrich logs with namespace fields.",
				},
				"kube_config_file": schema.StringAttribute{
					Optional:    true,
					Description: "Path to kubeconfig file.",
				},
				"read_from": schema.StringAttribute{
					Optional:    true,
					Description: "File read position.",
					Validators: []validator.String{
						stringvalidator.OneOf("beginning", "end"),
					},
				},
				"timezone": schema.StringAttribute{
					Optional:    true,
					Description: "Default timezone for log timestamps.",
				},
				"use_apiserver_cache": schema.BoolAttribute{
					Optional:    true,
					Description: "Use the kube-apiserver cache for pod metadata lookups.",
				},
			},
			Blocks: map[string]schema.Block{
				"namespace_annotation_fields": schema.ListNestedBlock{
					Description: "Controls how namespace metadata is attached to log events.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"namespace_labels": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with namespace labels.",
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
				},
				"node_annotation_fields": schema.ListNestedBlock{
					Description: "Controls how node metadata is attached to log events.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"node_labels": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with node labels.",
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
				},
				"pod_annotation_fields": schema.ListNestedBlock{
					Description: "Controls how pod metadata is attached to log events.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"container_id": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with the container ID.",
							},
							"container_image": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with the container image.",
							},
							"container_image_id": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with the container image ID.",
							},
							"container_name": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with the container name.",
							},
							"pod_annotations": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with pod annotations.",
							},
							"pod_ip": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with the pod IP.",
							},
							"pod_ips": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with the pod IPs.",
							},
							"pod_labels": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with pod labels.",
							},
							"pod_name": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with the pod name.",
							},
							"pod_namespace": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with the pod namespace.",
							},
							"pod_node_name": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with the pod node name.",
							},
							"pod_owner": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with the pod owner.",
							},
							"pod_uid": schema.StringAttribute{
								Optional:    true,
								Description: "The log event field to populate with the pod UID.",
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
