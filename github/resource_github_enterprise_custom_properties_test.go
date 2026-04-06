package github

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGithubEnterpriseCustomPropertiesValidation(t *testing.T) {
	t.Run("rejects invalid values_editable_by value", func(t *testing.T) {
		config := fmt.Sprintf(`
		resource "github_enterprise_custom_property" "test" {
			enterprise_slug    = "%s"
			property_name      = "TestInvalidValuesEditableBy"
			value_type         = "string"
			required           = false
			description        = "Test invalid values_editable_by"
			values_editable_by = "invalid_value"
		}`, testAccConf.enterpriseSlug)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile("invalid_value"),
				},
			},
		})
	})

	t.Run("rejects invalid value_type", func(t *testing.T) {
		config := fmt.Sprintf(`
		resource "github_enterprise_custom_property" "test" {
			enterprise_slug = "%s"
			property_name   = "TestInvalidValueType"
			value_type      = "invalid_type"
		}`, testAccConf.enterpriseSlug)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile("invalid_type"),
				},
			},
		})
	})
}

func TestAccGithubEnterpriseCustomProperties(t *testing.T) {
	t.Run("creates custom property without error", func(t *testing.T) {
		config := fmt.Sprintf(`
		resource "github_enterprise_custom_property" "test" {
			enterprise_slug = "%s"
			allowed_values  = ["Test"]
			description     = "Test Description"
			default_value   = "Test"
			property_name   = "TerraformAccTest"
			required        = true
			value_type      = "single_select"
		}`, testAccConf.enterpriseSlug)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_enterprise_custom_property.test", "property_name", "TerraformAccTest",
			),
			resource.TestCheckResourceAttr(
				"github_enterprise_custom_property.test", "value_type", "single_select",
			),
			resource.TestCheckResourceAttr(
				"github_enterprise_custom_property.test", "required", "true",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})

	t.Run("creates and updates a custom property", func(t *testing.T) {
		configBefore := fmt.Sprintf(`
		resource "github_enterprise_custom_property" "test" {
			enterprise_slug = "%s"
			allowed_values  = ["one"]
			description     = "Test Description"
			property_name   = "TerraformAccTestUpdate"
			value_type      = "single_select"
		}`, testAccConf.enterpriseSlug)

		configAfter := fmt.Sprintf(`
		resource "github_enterprise_custom_property" "test" {
			enterprise_slug = "%s"
			allowed_values  = ["one", "two"]
			description     = "Test Description Updated"
			property_name   = "TerraformAccTestUpdate"
			value_type      = "single_select"
		}`, testAccConf.enterpriseSlug)

		const resourceName = "github_enterprise_custom_property.test"

		checkBefore := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "allowed_values.#", "1"),
			resource.TestCheckResourceAttr(resourceName, "description", "Test Description"),
		)
		checkAfter := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "allowed_values.#", "2"),
			resource.TestCheckResourceAttr(resourceName, "description", "Test Description Updated"),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: configBefore,
					Check:  checkBefore,
				},
				{
					Config: configAfter,
					Check:  checkAfter,
				},
			},
		})
	})

	t.Run("imports enterprise custom property without error", func(t *testing.T) {
		description := "Test Description Import"
		propertyName := "TerraformAccTestImport"
		valueType := "string"

		config := fmt.Sprintf(`
		resource "github_enterprise_custom_property" "test" {
			enterprise_slug = "%s"
			description     = "%s"
			property_name   = "%s"
			value_type      = "%s"
		}`, testAccConf.enterpriseSlug, description, propertyName, valueType)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_enterprise_custom_property.test", "description", description,
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
				{
					ResourceName:      "github_enterprise_custom_property.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("creates custom property with values_editable_by", func(t *testing.T) {
		config := fmt.Sprintf(`
		resource "github_enterprise_custom_property" "test" {
			enterprise_slug    = "%s"
			property_name      = "TerraformAccTestEditableBy"
			value_type         = "string"
			required           = false
			description        = "Test property for values_editable_by"
			values_editable_by = "org_and_repo_actors"
		}`, testAccConf.enterpriseSlug)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_enterprise_custom_property.test", "property_name", "TerraformAccTestEditableBy",
			),
			resource.TestCheckResourceAttr(
				"github_enterprise_custom_property.test", "values_editable_by", "org_and_repo_actors",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})

	t.Run("backward compatibility - property without values_editable_by defaults correctly", func(t *testing.T) {
		config := fmt.Sprintf(`
		resource "github_enterprise_custom_property" "test" {
			enterprise_slug = "%s"
			property_name   = "TerraformAccTestBackwardCompat"
			value_type      = "string"
			required        = false
			description     = "Test property without values_editable_by"
		}`, testAccConf.enterpriseSlug)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_enterprise_custom_property.test", "property_name", "TerraformAccTestBackwardCompat",
			),
			resource.TestCheckResourceAttr(
				"github_enterprise_custom_property.test", "values_editable_by", "org_actors",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})
}
