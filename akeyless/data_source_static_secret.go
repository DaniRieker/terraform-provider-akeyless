package akeyless

import (
	"context"
	"errors"
	"fmt"

	"github.com/akeylesslabs/akeyless-go/v3"	
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceStaticSecret() *schema.Resource {
	return &schema.Resource{
		Description: "Static secret data source",
		Read:        dataSourceStaticSecretRead,
		Schema: map[string]*schema.Schema{
			"path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The path where the secret is stored. Defaults to the latest version.",
			},
			"get_metadata": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Should the metadata like description and tags be given back? Defaults to false.",
			},
			"version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The version of the secret.",
			},
			"value": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The secret contents.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the object",
			},
			"tags": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of the tags attached to this secret. To specify multiple tags use argument multiple times: -t Tag1 -t Tag2",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceStaticSecretRead(d *schema.ResourceData, m interface{}) error {
	provider := m.(providerMeta)
	client := *provider.client
	token := *provider.token

	path := d.Get("path").(string)
	getDescription := d.Get("get_metadata").(bool)

	var apiErr akeyless.GenericOpenAPIError
	ctx := context.Background()
	gsvBody := akeyless.GetSecretValue{
		Names: []string{path},
		Token: &token,
	}
	version := int32(d.Get("version").(int))

	if version != 0 {
		gsvBody.Version = &version
	}

	gsvOut, _, err := client.GetSecretValue(ctx).Body(gsvBody).Execute()
	if err != nil {
		if errors.As(err, &apiErr) {
			return fmt.Errorf("can't get Secret value: %v", string(apiErr.Body()))
		}
		return fmt.Errorf("can't get Secret value: %v", err)
	}

	err = d.Set("version", version)
	if err != nil {
		return err
	}
	err = d.Set("value", gsvOut[path])
	if err != nil {
		return err
	}

	d.SetId(path)

	if getDescription {
		gsdBody := akeyless.DescribeItem{
			Name:         path,
			ShowVersions: akeyless.PtrBool(true),
			Token:        &token,
		}
	
		itemOut, _, err := client.DescribeItem(ctx).Body(gsdBody).Execute()
		if err != nil {
			return err
		}
	
		if itemOut.ItemMetadata != nil {
			err = d.Set("description", *itemOut.ItemMetadata)
			if err != nil {
				return err
			}
		}

		if itemOut.ItemTags != nil {
			err = d.Set("tags", *itemOut.ItemTags)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}
