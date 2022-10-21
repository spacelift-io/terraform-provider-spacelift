package predicates

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func BooleanField(description string, maxItems int) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: description,
		Optional:    true,
		MaxItems:    maxItems,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"equals": {Type: schema.TypeBool},
			},
		},
	}
}

func StringField(description string, maxItems int) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: description,
		Optional:    true,
		MaxItems:    maxItems,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"any_of": {
					Type:     schema.TypeList,
					Required: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}
