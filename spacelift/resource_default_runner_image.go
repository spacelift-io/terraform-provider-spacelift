package spacelift

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func resourceDefaultRunnerImage() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_default_runner_image` represents the default runner images " +
			"that will be used for public and private worker pools.",

		CreateContext: resourceDefaultRunnerImageCreate,
		ReadContext:   resourceDefaultRunnerImageRead,
		UpdateContext: resourceDefaultRunnerImageUpdate,
		DeleteContext: resourceDefaultRunnerImageDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"private": {
				Type:         schema.TypeString,
				Description:  "Runner image to be used for private worker pools",
				Optional:     true,
				AtLeastOneOf: []string{"private", "public"},
			},
			"public": {
				Type:         schema.TypeString,
				Description:  "Runner image to be used for public worker pools",
				Optional:     true,
				AtLeastOneOf: []string{"private", "public"},
			},
		},
	}
}

type AccountUpdateDefaultWorkerPoolRunnerImagesResult struct {
	PrivateWorkerPoolRunnerImage *string `graphql:"privateWorkerPoolRunnerImage"`
	PublicWorkerPoolRunnerImage  *string `graphql:"publicWorkerPoolRunnerImage"`
}

func resourceDefaultRunnerImageCreate(ctx context.Context, data *schema.ResourceData, i any) diag.Diagnostics {
	var mutation struct {
		Result *AccountUpdateDefaultWorkerPoolRunnerImagesResult `graphql:"accountUpdateDefaultWorkerPoolRunnerImages(privateWorkerPoolRunnerImage: $privateWorkerPoolRunnerImage, publicWorkerPoolRunnerImage: $publicWorkerPoolRunnerImage)"`
	}

	variables := map[string]any{
		"privateWorkerPoolRunnerImage": graphql.String(""),
		"publicWorkerPoolRunnerImage":  graphql.String(""),
	}
	if v, ok := data.GetOk("private"); ok {
		variables["privateWorkerPoolRunnerImage"] = graphql.String(v.(string))
	}
	if v, ok := data.GetOk("public"); ok {
		variables["publicWorkerPoolRunnerImage"] = graphql.String(v.(string))
	}

	if err := i.(*internal.Client).Mutate(ctx, "AccountUpdateDefaultRunnerImage", &mutation, variables); err != nil {
		return diag.Errorf("could not set default runner image for account: %v", err)
	}

	data.SetId(time.Now().String())

	return resourceDefaultRunnerImageRead(ctx, data, i)
}

func resourceDefaultRunnerImageRead(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var query struct {
		DefaultPublicWorkerPoolRunnerImage  *string `graphql:"defaultPublicWorkerPoolRunnerImage"`
		DefaultPrivateWorkerPoolRunnerImage *string `graphql:"defaultPrivateWorkerPoolRunnerImage"`
	}
	if err := i.(*internal.Client).Query(ctx, "DefaultRunnerImage", &query, nil); err != nil {
		return diag.Errorf("could not query for default runner image: %v", err)
	}

	// Check if both values are nil or empty - if so, the resource doesn't exist
	privateEmpty := query.DefaultPrivateWorkerPoolRunnerImage == nil || *query.DefaultPrivateWorkerPoolRunnerImage == ""
	publicEmpty := query.DefaultPublicWorkerPoolRunnerImage == nil || *query.DefaultPublicWorkerPoolRunnerImage == ""

	if privateEmpty && publicEmpty {
		data.SetId("")
		return nil
	}

	if !publicEmpty {
		data.Set("public", *query.DefaultPublicWorkerPoolRunnerImage)
	} else {
		data.Set("public", "")
	}

	if !privateEmpty {
		data.Set("private", *query.DefaultPrivateWorkerPoolRunnerImage)
	} else {
		data.Set("private", "")
	}

	return nil
}

func resourceDefaultRunnerImageUpdate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var mutation struct {
		Result *AccountUpdateDefaultWorkerPoolRunnerImagesResult `graphql:"accountUpdateDefaultWorkerPoolRunnerImages(privateWorkerPoolRunnerImage: $privateWorkerPoolRunnerImage, publicWorkerPoolRunnerImage: $publicWorkerPoolRunnerImage)"`
	}

	variables := map[string]any{
		"privateWorkerPoolRunnerImage": graphql.String(""),
		"publicWorkerPoolRunnerImage":  graphql.String(""),
	}
	if v, ok := data.GetOk("private"); ok {
		variables["privateWorkerPoolRunnerImage"] = graphql.String(v.(string))
	}
	if v, ok := data.GetOk("public"); ok {
		variables["publicWorkerPoolRunnerImage"] = graphql.String(v.(string))
	}

	if err := i.(*internal.Client).Mutate(ctx, "AccountUpdateDefaultRunnerImage", &mutation, variables); err != nil {
		return diag.Errorf("could not set default runner image for account: %v", err)
	}

	return resourceDefaultRunnerImageRead(ctx, data, i)
}

func resourceDefaultRunnerImageDelete(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var mutation struct {
		Result *AccountUpdateDefaultWorkerPoolRunnerImagesResult `graphql:"accountUpdateDefaultWorkerPoolRunnerImages(privateWorkerPoolRunnerImage: $privateWorkerPoolRunnerImage, publicWorkerPoolRunnerImage: $publicWorkerPoolRunnerImage)"`
	}

	variables := map[string]any{
		"privateWorkerPoolRunnerImage": graphql.String(""),
		"publicWorkerPoolRunnerImage":  graphql.String(""),
	}

	if err := i.(*internal.Client).Mutate(ctx, "AccountUpdateDefaultRunnerImage", &mutation, variables); err != nil {
		return diag.Errorf("could not unset default runner image for account: %v", err)
	}

	data.SetId("")

	return nil
}
