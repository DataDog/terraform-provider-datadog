package datadog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/Masterminds/semver/v3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

const catalogPath = "/api/v2/catalog/entity"

func resourceDatadogCatalogEntity() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Software Catalog Entity resource. This can be used to create and manage entities in Datadog Software Catalog using the YAML/JSON definition.",
		CreateContext: resourceEntityCreate,
		ReadContext:   resourceEntityRead,
		UpdateContext: resourceEntityUpdate,
		DeleteContext: resourceEntityDelete,
		CustomizeDiff: customdiff.ForceNewIfChange("software_catalog", func(ctx context.Context, old, new, meta any) bool {
			// we use this function to compute what's considered a new entity.
			// if the entity's reference is changed, then it's a new entity. eg. from service:myservice to service:otherservice
			// else if the entity's attributes, such as owner, are changed, then we are updating it.
			oldEntity, errO := entityFromYAML(old.(string))
			if errO != nil {
				return false
			}
			newEntity, errN := entityFromYAML(new.(string))
			if errN != nil {
				return false
			}
			// reference is a unique key. if it's changed, then we force to create a new entity.
			return oldEntity.reference().equal(*newEntity.reference())
		}),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"entity": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validEntity,
					StateFunc: func(v any) string {
						e, _ := entityFromYAML(v.(string))
						prepEntityResource(e)
						res, _ := e.toYAML()
						return res
					},
					Description: "The YAML/JSON definition of entity",
				},
				// in the future, we should extend the schema to support schema specific fields. similar to how our SDKs work.
			}
		},
	}
}

func validEntity(in any, k string) (warnings []string, errors []error) {
	// verify one yaml document 
	inYAML, ok := in.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("Input is not string %q", k))
	}
	e, err := entityFromYAML(inYAML)
	if err != nil {
		errors = append(errors, fmt.Errorf("Error while parsing input %q, err %s", k, err))
	}

	// verify apiVersion is v3 or above 
	if e.APIVersion == "" {
		errors = append(errors, fmt.Errorf("Missing apiVersion %q", k))
	}
	v, err := semver.NewVersion(e.APIVersion)
	if err != nil {
		errors = append(errors, fmt.Errorf("Invalid apiVersion string %q, err %s", k, err))
	}
	if v.Major() < 3 {
		errors = append(errors, fmt.Errorf("apiVersion v3 or above is required %q", k))
	}
	
	// verify name and kind are present 
	if e.Kind == "" {
		errors = append(errors, fmt.Errorf("Missing kind %q", k))
	}
	if e.Metadata != nil && e.Metadata.name() == "" {
		errors = append(errors, fmt.Errorf("Missing name %q", k))
	}
	return nil, nil
}

func prepEntityResource(e Entity) Entity {
	if e.Metadata != nil {
		// this is generated field, let's remove it
		delete(e.Metadata, "managed")
	}
	return e
}

type UpsertEntityResponse struct {
	Included []Included `json:"included"`
}

type Included struct {
	ID         string
	Attributes *IncludedAttributes `json:"attributes"`
}

type IncludedAttributes struct {
	Schema *Entity `json:"schema"`
}

func resourceEntityRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	// id represents an entity's reference, e.g. service:default/myservice
	id := d.Id()
	respByte, resp, err := utils.SendRequest(auth, apiInstances.HttpClient, "GET", catalogPath+"?filter[ref]="+id, nil)
	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return utils.TranslateClientErrorDiag(err, resp, fmt.Sprintf("error retrieving entity with reference: %s", id))
	}

	var response UpsertEntityResponse

	err = json.Unmarshal(respByte, &response)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(response.Included) != 1 || response.Included[0].Attributes == nil || response.Included[0].Attributes.Schema == nil {
		return diag.FromErr(errors.New("error retrieving data from response"))
	}

	e := response.Included[0].Attributes.Schema

	return updateEntityState(d, *e)
}

func resourceEntityCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	e, diag := resourceEntityUpsert(d, meta, "creating entity")
	if diag != nil {
		return diag
	}
	d.SetId(e.reference().String())
	return updateEntityState(d, e)
}

func resourceEntityUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	e, diag := resourceEntityUpsert(d, meta, "updating entity")
	if diag != nil {
		return diag
	}
	return updateEntityState(d, e)
}

func resourceEntityUpsert(d *schema.ResourceData, meta any, action string) (Entity, diag.Diagnostics) {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	entity := d.Get("entity").(string)

	respByte, resp, err := utils.SendRequest(auth, apiInstances.HttpClient, "POST", catalogPath, &entity)
	if err != nil {
		return Entity{}, utils.TranslateClientErrorDiag(err, resp, "error "+action)
	}

	var response UpsertEntityResponse
	err = json.Unmarshal(respByte, &response)
	if err != nil {
		return Entity{}, diag.FromErr(err)
	}

	if len(response.Included) != 1 || response.Included[0].Attributes == nil || response.Included[0].Attributes.Schema == nil {
		return Entity{}, diag.FromErr(errors.New("error retrieving data from response"))
	}
	e := response.Included[0].Attributes.Schema
	return *e, nil
}

func resourceEntityDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	id := d.Id()
	_, resp, err := utils.SendRequest(auth, apiInstances.HttpClient, "DELETE", catalogPath+"/"+id, nil)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, resp, "error deleting entity")
	}

	return nil
}

func updateEntityState(d *schema.ResourceData, e Entity) diag.Diagnostics {
	cleanedEntity := prepEntityResource(e)
	entityYAML, err := cleanedEntity.toYAML()
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("entity", entityYAML); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func entityFromYAML(inYAML string) (Entity, error) {
	var e Entity
	err := yaml.Unmarshal([]byte(inYAML), &e)
	return e, err
}

type Reference struct {
	Kind      string
	Name      string
	Namespace string
}

func (r Reference) equal(o Reference) bool {
	return r.Kind == o.Kind && r.Name == o.Name && r.Namespace == o.Namespace
}

func (r Reference) String() string {
	return string(r.Kind) + ":" + r.Namespace + "/" + r.Name
}

type Entity struct {
	APIVersion   string         `yaml:"apiVersion" json:"apiVersion"`
	Kind         string         `yaml:"kind" json:"kind"`
	Metadata     Metadata       `yaml:"metadata" json:"metadata"`
	Spec         map[string]any `yaml:"spec" json:"spec"`
	Integrations map[string]any `yaml:"integrations" json:"integrations"`
	Datadog      map[string]any `yaml:"datadog" json:"datadog"`
}

func (e *Entity) reference() *Reference {
	if e.Kind == "" {
		return nil
	}
	if e.Metadata == nil {
		return nil
	}
	if e.Metadata.name() == "" {
		return nil
	}
	return &Reference{
		Kind:      e.Kind,
		Name:      e.Metadata.name(),
		Namespace: e.Metadata.namespace(),
	}
}

func (e *Entity) toYAML() (string, error) {
	result, err := yaml.Marshal(e)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

type Metadata map[string]any

func (m Metadata) name() string {
	name, _ := m["name"].(string)
	return name
}

func (m Metadata) namespace() string {
	if namespace, ok := m["namespace"].(string); ok {
		return namespace
	}
	return "default"
}
