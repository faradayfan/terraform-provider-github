package github

import (
	"context"
	"log"
	"net/http"

	"github.com/google/go-github/v84/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubEnterpriseCustomProperties() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubEnterpriseCustomPropertiesCreate,
		Read:   resourceGithubEnterpriseCustomPropertiesRead,
		Update: resourceGithubEnterpriseCustomPropertiesUpdate,
		Delete: resourceGithubEnterpriseCustomPropertiesDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubEnterpriseCustomPropertiesImport,
		},

		Schema: map[string]*schema.Schema{
			"enterprise_slug": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The slug of the enterprise.",
			},
			"property_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the custom property.",
			},
			"value_type": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The type of the value for the property. Can be one of: 'string', 'single_select', 'multi_select', 'true_false', 'url'.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{string(github.PropertyValueTypeString), string(github.PropertyValueTypeSingleSelect), string(github.PropertyValueTypeMultiSelect), string(github.PropertyValueTypeTrueFalse), string(github.PropertyValueTypeURL)}, false)),
			},
			"required": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether the custom property is required.",
			},
			"default_value": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The default value of the custom property.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "A short description of the custom property.",
			},
			"allowed_values": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "An ordered list of allowed values for the property. Only applicable to 'single_select' and 'multi_select' types.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"values_editable_by": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "Who can edit the values of the property. Can be one of: 'org_actors', 'org_and_repo_actors'. Defaults to 'org_actors'.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"org_actors", "org_and_repo_actors"}, false)),
			},
		},
	}
}

func resourceGithubEnterpriseCustomPropertiesCreate(d *schema.ResourceData, meta any) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	enterpriseSlug := d.Get("enterprise_slug").(string)
	propertyName := d.Get("property_name").(string)

	property := buildEnterpriseCustomProperty(d)

	_, _, err := client.Enterprise.CreateOrUpdateCustomProperty(ctx, enterpriseSlug, propertyName, property)
	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(enterpriseSlug, propertyName))
	return resourceGithubEnterpriseCustomPropertiesRead(d, meta)
}

func resourceGithubEnterpriseCustomPropertiesRead(d *schema.ResourceData, meta any) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	enterpriseSlug, propertyName, err := parseTwoPartID(d.Id(), "enterprise_slug", "property_name")
	if err != nil {
		return err
	}

	property, resp, err := client.Enterprise.GetCustomProperty(ctx, enterpriseSlug, propertyName)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			log.Printf("[INFO] Removing enterprise custom property %s/%s from state because it no longer exists", enterpriseSlug, propertyName)
			d.SetId("")
			return nil
		}
		return err
	}

	defaultValue, _ := property.DefaultValueString()

	d.SetId(buildTwoPartID(enterpriseSlug, propertyName))
	_ = d.Set("enterprise_slug", enterpriseSlug)
	_ = d.Set("property_name", property.GetPropertyName())
	_ = d.Set("value_type", string(property.ValueType))
	_ = d.Set("required", property.GetRequired())
	_ = d.Set("default_value", defaultValue)
	_ = d.Set("description", property.GetDescription())
	_ = d.Set("allowed_values", property.AllowedValues)
	_ = d.Set("values_editable_by", property.GetValuesEditableBy())

	return nil
}

func resourceGithubEnterpriseCustomPropertiesUpdate(d *schema.ResourceData, meta any) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	enterpriseSlug, propertyName, err := parseTwoPartID(d.Id(), "enterprise_slug", "property_name")
	if err != nil {
		return err
	}

	property := buildEnterpriseCustomProperty(d)

	_, _, err = client.Enterprise.CreateOrUpdateCustomProperty(ctx, enterpriseSlug, propertyName, property)
	if err != nil {
		return err
	}

	return resourceGithubEnterpriseCustomPropertiesRead(d, meta)
}

func resourceGithubEnterpriseCustomPropertiesDelete(d *schema.ResourceData, meta any) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	enterpriseSlug, propertyName, err := parseTwoPartID(d.Id(), "enterprise_slug", "property_name")
	if err != nil {
		return err
	}

	_, err = client.Enterprise.RemoveCustomProperty(ctx, enterpriseSlug, propertyName)
	return err
}

func resourceGithubEnterpriseCustomPropertiesImport(d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	enterpriseSlug, propertyName, err := parseTwoPartID(d.Id(), "enterprise_slug", "property_name")
	if err != nil {
		return nil, err
	}

	_ = d.Set("enterprise_slug", enterpriseSlug)
	_ = d.Set("property_name", propertyName)

	return []*schema.ResourceData{d}, nil
}

func buildEnterpriseCustomProperty(d *schema.ResourceData) *github.CustomProperty {
	propertyName := d.Get("property_name").(string)
	valueType := github.PropertyValueType(d.Get("value_type").(string))
	required := d.Get("required").(bool)
	defaultValue := d.Get("default_value").(string)
	description := d.Get("description").(string)

	rawAllowedValues := d.Get("allowed_values").([]any)
	allowedValues := make([]string, 0, len(rawAllowedValues))
	for _, v := range rawAllowedValues {
		allowedValues = append(allowedValues, v.(string))
	}

	property := &github.CustomProperty{
		PropertyName:  &propertyName,
		ValueType:     valueType,
		Required:      &required,
		DefaultValue:  &defaultValue,
		Description:   &description,
		AllowedValues: allowedValues,
	}

	if val, ok := d.GetOk("values_editable_by"); ok {
		str := val.(string)
		property.ValuesEditableBy = &str
	}

	return property
}
