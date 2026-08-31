package spacelift

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

// The API models AI integrations as one object discriminated by a `provider`
// field, so this is one resource and the provider is expressed by the block
// that is set. Adding a provider means adding a block plus an entry in
// aiProviderBlocks and aiProviderBlockNames.
const (
	aiProviderAnthropic = "Anthropic"
	aiProviderBedrock   = "Bedrock"
	aiProviderGoogle    = "Google"
	aiProviderOpenAI    = "OpenAI"
)

// aiProviderBlocks maps the provider value the API expects onto its block.
var aiProviderBlocks = map[string]string{
	aiProviderAnthropic: "anthropic",
	aiProviderBedrock:   "bedrock",
	aiProviderGoogle:    "google",
	aiProviderOpenAI:    "openai",
}

// aiBlockSpacelift names no provider: it marks an integration Spacelift
// provides and shares with every account, which is read-only apart from
// `enabled`.
const aiBlockSpacelift = "spacelift"

const aiSpaceliftProvidedCreateError = "integrations provided by Spacelift cannot be created, only " +
	"imported: look the ID up with the spacelift_ai_integrations data source and import it"

// Sorted, so the generated ExactlyOneOf lists read consistently.
var aiProviderBlockNames = []string{"anthropic", "bedrock", "google", "openai", aiBlockSpacelift}

var aiProviderNames = []string{aiProviderAnthropic, aiProviderBedrock, aiProviderGoogle, aiProviderOpenAI}

// Attributes that belong to Spacelift on a Spacelift-provided integration.
var aiSpaceliftManagedAttributes = []string{"name", "description", "labels", "models", "space_id"}

func resourceAIIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_ai_integration` represents an integration with an AI/LLM provider, used by " +
			"Spacelift features that call out to a model.\n\n" +
			"The provider is chosen by setting exactly one provider block: `anthropic`, `bedrock`, " +
			"`google` (the Gemini API) or `openai`. Changing which block is set replaces the " +
			"integration, because the API does not allow an existing one to change provider.\n\n" +
			"A fifth block, `spacelift`, marks an integration that Spacelift provides and shares " +
			"with every account. Those can only be imported and toggled, so the block takes no " +
			"arguments and everything except `enabled` has to be left out of the configuration.",

		CreateContext: resourceAIIntegrationCreate,
		ReadContext:   resourceAIIntegrationRead,
		UpdateContext: resourceAIIntegrationUpdate,
		DeleteContext: resourceAIIntegrationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customizeAIIntegrationDiff,

		Schema: map[string]*schema.Schema{
			"name": {
				Type: schema.TypeString,
				Description: "Friendly name of the integration. Required, except on a " +
					"Spacelift-provided integration, whose name belongs to Spacelift and is read " +
					"back from the API.",
				// ConflictsWith cannot point at a Required attribute, so the
				// requirement is enforced in customizeAIIntegrationDiff instead.
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			aiBlockSpacelift: {
				Type: schema.TypeList,
				Description: "Marks a read-only integration provided by Spacelift and shared with " +
					"every account. The block takes no arguments because everything except " +
					"`enabled` belongs to Spacelift. These integrations cannot be created or " +
					"deleted, so the resource has to be imported, and removing it from the " +
					"configuration only drops it from the state.",
				Optional:      true,
				MaxItems:      1,
				ExactlyOneOf:  aiProviderBlockNames,
				ConflictsWith: aiSpaceliftManagedAttributes,
				Elem:          &schema.Resource{Schema: map[string]*schema.Schema{}},
			},
			"anthropic": aiProviderAPIKeyBlock("anthropic", "Anthropic", "Anthropic"),
			"google":    aiProviderAPIKeyBlock("google", "Google", "Gemini"),
			"openai":    aiProviderAPIKeyBlock("openai", "OpenAI", "OpenAI"),
			"bedrock": {
				Type: schema.TypeList,
				Description: "AWS Bedrock-specific configuration. Presence means this integration calls " +
					"Bedrock through an AWS integration rather than using an API key.",
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: aiProviderBlockNames,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"integration_id": {
							Type:             schema.TypeString,
							Description:      "ID of the AWS integration used to reach Bedrock",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"region": {
							Type:             schema.TypeString,
							Description:      "AWS region the Bedrock requests are routed through, for example `us-east-1`",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"profiles": {
							Type: schema.TypeList,
							Description: "Bedrock inference profile ARNs to expose. Spacelift validates these " +
								"against AWS through the attached integration when the integration is saved.",
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: validations.DisallowEmptyString,
							},
						},
						"integration_name": {
							Type:        schema.TypeString,
							Description: "Name of the AWS integration used to reach Bedrock",
							Computed:    true,
						},
					},
				},
			},
			"models": {
				Type: schema.TypeList,
				Description: "Model identifiers available on this integration, for example `gemini-2.5-pro`. " +
					"Leave unset to accept the provider's default model list, which is then recorded in " +
					"the state but never sent back, so removing the attribute later produces no diff. " +
					"Pinning is one-way: the API reports the models an integration resolves to rather " +
					"than the ones it was given, so it cannot report that nothing is pinned, and an " +
					"empty list is rejected rather than left to diff on every plan. " +
					"Not supported for Bedrock, which takes its models from `bedrock.profiles` instead.",
				Optional: true,
				Computed: true,
				// An empty list clears the pin, but the API then reports the
				// defaults it resolved, which can never match the empty list in
				// the configuration.
				MinItems: 1,
				// The API rejects models outright for Bedrock integrations.
				ConflictsWith: []string{"bedrock"},
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Free-form description of the integration",
				Optional:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels to set on the integration",
				Optional:    true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID of the space the integration belongs to",
				Optional:    true,
				Computed:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether the integration can be used for LLM calls. Defaults to `true`.",
				Optional:    true,
				Default:     true,
			},
			"ai_provider": {
				Type:        schema.TypeString,
				Description: "AI provider backing the integration, derived from the provider block that is set",
				Computed:    true,
			},
			"is_spacelift_provided": {
				Type:        schema.TypeBool,
				Description: "Whether this is a read-only, Spacelift-managed integration shared with every account",
				Computed:    true,
			},
		},
	}
}

// aiProviderAPIKeyBlock builds the block for an API-key based provider. Every
// provider except Bedrock takes the same shape, so they share one definition.
// api is how the provider names its own API, which is not always the provider
// name: Google's is the Gemini API.
func aiProviderAPIKeyBlock(name, label, api string) *schema.Schema {
	path := func(field string) string { return fmt.Sprintf("%s.0.%s", name, field) }

	return &schema.Schema{
		Type: schema.TypeList,
		Description: fmt.Sprintf(
			"%s-specific configuration. Presence means this integration uses the %s API.",
			label, api,
		),
		Optional:     true,
		MaxItems:     1,
		ExactlyOneOf: aiProviderBlockNames,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"api_key": {
					Type: schema.TypeString,
					Description: fmt.Sprintf(
						"%s API key. Note that once it's created, it will be just an empty string in the state due to security reasons.",
						api,
					),
					Optional:         true,
					Sensitive:        true,
					Deprecated:       "`api_key` is deprecated. Please use api_key_wo in combination with api_key_wo_version",
					DiffSuppressFunc: ignoreOnceCreated,
					ConflictsWith:    []string{path("api_key_wo"), path("api_key_wo_version")},
					// The API requires a key for every provider except Bedrock,
					// so catch an empty block at plan time rather than on apply.
					AtLeastOneOf: []string{path("api_key"), path("api_key_wo")},
					// An empty string would otherwise satisfy AtLeastOneOf.
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				"api_key_wo": {
					Type: schema.TypeString,
					Description: fmt.Sprintf(
						"%s API key. The key is not stored in the state. Modify api_key_wo_version to trigger an update. This field requires Terraform/OpenTofu 1.11+.",
						api,
					),
					Optional:         true,
					Sensitive:        true,
					WriteOnly:        true,
					ConflictsWith:    []string{path("api_key")},
					RequiredWith:     []string{path("api_key_wo_version")},
					AtLeastOneOf:     []string{path("api_key"), path("api_key_wo")},
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				"api_key_wo_version": {
					Type:          schema.TypeString,
					Description:   "Used together with api_key_wo to trigger an update to the API key. Increment this value when an update to api_key_wo is required. This field requires Terraform/OpenTofu 1.11+.",
					Optional:      true,
					ConflictsWith: []string{path("api_key")},
					RequiredWith:  []string{path("api_key_wo")},
					// An empty version satisfies RequiredWith but reads as unset
					// when the key is extracted, which would quietly send an
					// empty key instead of the one that was configured.
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				"base_url": {
					Type: schema.TypeString,
					Description: fmt.Sprintf(
						"Custom base URL for a self-hosted, %s-compatible gateway such as LiteLLM. "+
							"Leave unset to call the provider directly.",
						api,
					),
					Optional: true,
				},
			},
		},
	}
}

// otherAIProviderBlocks returns every provider block except the given one, used
// on read to clear the blocks that do not apply. Rejecting more than one block
// is left to ExactlyOneOf, which already covers it.
func otherAIProviderBlocks(block string) []string {
	others := make([]string, 0, len(aiProviderBlockNames))
	for _, name := range aiProviderBlockNames {
		if name != block {
			others = append(others, name)
		}
	}

	return others
}

func aiBlockSet(d *schema.ResourceDiff, name string) bool {
	list, ok := d.Get(name).([]any)

	return ok && len(list) > 0
}

// customizeAIIntegrationDiff keeps the block that is set and the integration it
// points at in step. aiIntegrationUpdate takes a provider argument but the API
// refuses to change it, so swapping one provider block for another replaces the
// integration; swapping into or out of `spacelift` cannot be planned at all,
// because those integrations can be neither created nor deleted.
func customizeAIIntegrationDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	spaceliftBlock := aiBlockSet(d, aiBlockSpacelift)

	// The raw config decides, because an unknown name reads back as empty.
	if !spaceliftBlock {
		if config := d.GetRawConfig(); !config.IsNull() && config.GetAttr("name").IsNull() {
			return errors.New("`name` is required unless the `spacelift` block is set")
		}
	}

	if d.Id() == "" {
		if spaceliftBlock {
			return errors.New(aiSpaceliftProvidedCreateError)
		}

		return nil
	}

	if spaceliftProvided := d.Get("is_spacelift_provided").(bool); spaceliftProvided != spaceliftBlock {
		if spaceliftProvided {
			return fmt.Errorf(
				"AI integration %s is provided by Spacelift, so it can only be managed through the "+
					"`spacelift` block", d.Id(),
			)
		}

		return fmt.Errorf(
			"AI integration %s is not provided by Spacelift, so the `spacelift` block does not "+
				"apply to it: use the block matching its provider instead", d.Id(),
		)
	}

	if spaceliftBlock {
		return nil
	}

	current := d.Get("ai_provider").(string)
	if current == "" {
		return nil
	}

	for _, name := range aiProviderBlockNames {
		if !aiBlockSet(d, name) {
			continue
		}

		if aiBlockProviders[name] != current {
			return d.ForceNew(name)
		}
	}

	return nil
}

// aiBlockProviders is aiProviderBlocks inverted.
var aiBlockProviders = func() map[string]string {
	inverted := make(map[string]string, len(aiProviderBlocks))
	for provider, block := range aiProviderBlocks {
		inverted[block] = provider
	}

	return inverted
}()

// aiIntegrationProvider derives the provider from the block that is set.
// ExactlyOneOf already rejects zero or several blocks at plan time; iterating
// the ordered names and counting means that if that ever slipped we fail
// loudly rather than picking a provider by Go's map ordering.
func aiIntegrationProvider(d *schema.ResourceData) (provider string, block string, err error) {
	var set []string

	for _, name := range aiProviderBlockNames {
		v, ok := d.GetOk(name)
		if !ok {
			continue
		}

		if list, ok := v.([]any); ok && len(list) > 0 {
			set = append(set, name)
		}
	}

	switch len(set) {
	case 1:
		return aiBlockProviders[set[0]], set[0], nil
	case 0:
		return "", "", fmt.Errorf("exactly one provider block must be set, one of: %v", aiProviderBlockNames)
	default:
		return "", "", fmt.Errorf("exactly one provider block may be set, got: %v", set)
	}
}

// aiIntegrationVariables builds the arguments shared by create and update.
// Every argument must be present because the list is baked into the query, so
// the ones a provider does not use are sent as null.
func aiIntegrationVariables(d *schema.ResourceData, provider, block string, creating bool) (map[string]any, diag.Diagnostics) {
	variables := map[string]any{
		"name":        toString(d.Get("name")),
		"description": toOptionalString(d.Get("description")),
		"labels":      setToOptionalStringList(d.Get("labels")),
		"provider":    graphql.String(provider),
		"space":       toOptionalID(d.Get("space_id")),

		// Only some providers use these.
		"apiKey":                (*graphql.String)(nil),
		"baseURL":               (*graphql.String)(nil),
		"models":                (*[]graphql.String)(nil),
		"providerIntegrationId": (*graphql.ID)(nil),
		"providerConfig":        (*structs.AIProviderConfigInput)(nil),
	}

	if provider == aiProviderBedrock {
		variables["providerIntegrationId"] = toOptionalID(d.Get(block + ".0.integration_id"))
		variables["providerConfig"] = &structs.AIProviderConfigInput{
			Bedrock: &structs.AWSBedrockConfigInput{
				Region:   toString(d.Get(block + ".0.region")),
				Profiles: listToStringList(d.Get(block + ".0.profiles")),
			},
		}

		return variables, nil
	}

	// The API treats any key it receives as a rotation, and with a custom
	// base_url that re-verifies the gateway and re-probes the models, so a
	// description-only apply must not resend one.
	if creating || d.HasChange(block+".0.api_key_wo_version") {
		apiKey, diags := internal.ExtractWriteOnlyFieldInBlock(block, "api_key", "api_key_wo", "api_key_wo_version", d)
		if diags != nil {
			return nil, diags
		}

		variables["apiKey"] = toOptionalString(apiKey)
	}

	variables["baseURL"] = toOptionalString(d.Get(block + ".0.base_url"))
	variables["models"] = aiIntegrationModels(d)

	return variables, nil
}

// aiIntegrationModels reads the models from the configuration rather than the
// state. The attribute is Optional and Computed, so the state holds whichever
// defaults the API filled in, and sending those back would pin the integration
// to a snapshot of them on the first unrelated update.
func aiIntegrationModels(d *schema.ResourceData) *[]graphql.String {
	config := d.GetRawConfig()
	if config.IsNull() {
		return nil
	}

	models := config.GetAttr("models")
	if models.IsNull() {
		return nil
	}

	if !models.IsWhollyKnown() {
		return listToOptionalStringList(d.Get("models"))
	}

	list := make([]graphql.String, 0, models.LengthInt())
	for _, model := range models.AsValueSlice() {
		list = append(list, graphql.String(model.AsString()))
	}

	return &list
}

func resourceAIIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	provider, block, err := aiIntegrationProvider(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if block == aiBlockSpacelift {
		return diag.Errorf("%s", aiSpaceliftProvidedCreateError)
	}

	variables, diags := aiIntegrationVariables(d, provider, block, true)
	if diags != nil {
		return diags
	}

	var mutation struct {
		AIIntegration structs.AIIntegration `graphql:"aiIntegrationCreate(name: $name, description: $description, labels: $labels, provider: $provider, apiKey: $apiKey, providerIntegrationId: $providerIntegrationId, providerConfig: $providerConfig, baseURL: $baseURL, models: $models, space: $space)"`
	}

	if err := meta.(*internal.Client).Mutate(ctx, "AIIntegrationCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create AI integration: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.AIIntegration.ID)

	// enabled is not a create argument, so reconcile it against what the API
	// actually returned rather than assuming a default.
	if desired := d.Get("enabled").(bool); desired != mutation.AIIntegration.Enabled {
		if err := setAIIntegrationEnabled(ctx, meta, d.Id(), desired); err != nil {
			return append(
				diag.Errorf("could not set enabled on the AI integration: %v", err),
				resourceAIIntegrationRead(ctx, d, meta)...,
			)
		}
	}

	return resourceAIIntegrationRead(ctx, d, meta)
}

func resourceAIIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var query struct {
		AIIntegration *structs.AIIntegration `graphql:"aiIntegration(id: $id)"`
	}

	variables := map[string]any{"id": toID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "AIIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for AI integration: %v", internal.FromSpaceliftError(err))
	}

	integration := query.AIIntegration
	if integration == nil {
		d.SetId("")
		return nil
	}

	d.SetId(integration.ID)
	d.Set("name", integration.Name)
	d.Set("ai_provider", integration.Provider)
	d.Set("models", integration.Models)
	d.Set("enabled", integration.Enabled)
	d.Set("is_spacelift_provided", integration.IsSpaceliftProvided)

	if integration.Space != nil {
		d.Set("space_id", integration.Space.ID)
	} else {
		d.Set("space_id", "")
	}

	// description and labels are the only attributes that are Optional without
	// being Computed, so recording them for an integration whose configuration
	// is not allowed to hold them would diff against that empty configuration
	// forever. The data sources expose them instead.
	if !integration.IsSpaceliftProvided {
		d.Set("description", integration.Description)

		labels := schema.NewSet(schema.HashString, []any{})
		for _, label := range integration.Labels {
			labels.Add(label)
		}
		d.Set("labels", labels)
	}

	if err := setAIIntegrationProviderBlock(d, integration); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// setAIIntegrationProviderBlock fills the block matching the integration's
// provider and clears the others.
func setAIIntegrationProviderBlock(d *schema.ResourceData, integration *structs.AIIntegration) error {
	if integration.IsSpaceliftProvided {
		for _, other := range otherAIProviderBlocks(aiBlockSpacelift) {
			d.Set(other, nil)
		}

		return d.Set(aiBlockSpacelift, []any{map[string]any{}})
	}

	block, ok := aiProviderBlocks[integration.Provider]
	if !ok {
		return fmt.Errorf(
			"AI integration %s uses provider %q, which this provider version does not support",
			integration.ID, integration.Provider,
		)
	}

	var entry map[string]any

	if integration.Provider == aiProviderBedrock {
		entry = map[string]any{
			"integration_id":   "",
			"integration_name": "",
			"region":           "",
			"profiles":         []string{},
		}

		if integration.ProviderIntegrationID != nil {
			entry["integration_id"] = *integration.ProviderIntegrationID
		}

		if integration.ProviderIntegrationName != nil {
			entry["integration_name"] = *integration.ProviderIntegrationName
		}

		if integration.ProviderConfig != nil {
			entry["region"] = integration.ProviderConfig.Bedrock.Region
			entry["profiles"] = integration.ProviderConfig.Bedrock.Profiles
		}
	} else {
		baseURL := ""
		if integration.BaseURL != nil {
			baseURL = *integration.BaseURL
		}

		entry = map[string]any{
			"base_url": baseURL,
			// The API never returns the key.
			"api_key": "",
			// Config-only, so carry it rather than clobbering it on refresh.
			"api_key_wo_version": d.Get(block + ".0.api_key_wo_version"),
		}
	}

	for _, other := range otherAIProviderBlocks(block) {
		d.Set(other, nil)
	}

	return d.Set(block, []any{entry})
}

func resourceAIIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Spacelift-provided integrations are read-only, but aiIntegrationToggle
	// explicitly works on them, so only reject changes to anything else.
	spaceliftProvided := d.Get("is_spacelift_provided").(bool)
	if spaceliftProvided && d.HasChangesExcept("enabled") {
		return diag.Errorf(
			"Spacelift-provided AI integrations are read-only, only `enabled` can be changed",
		)
	}

	var ret diag.Diagnostics

	if !spaceliftProvided && d.HasChangesExcept("enabled") {
		provider, block, err := aiIntegrationProvider(d)
		if err != nil {
			return diag.FromErr(err)
		}

		variables, diags := aiIntegrationVariables(d, provider, block, false)
		if diags != nil {
			return diags
		}
		variables["id"] = toID(d.Id())

		var mutation struct {
			AIIntegration structs.AIIntegration `graphql:"aiIntegrationUpdate(id: $id, name: $name, description: $description, labels: $labels, provider: $provider, apiKey: $apiKey, providerIntegrationId: $providerIntegrationId, providerConfig: $providerConfig, baseURL: $baseURL, models: $models, space: $space)"`
		}

		if err := meta.(*internal.Client).Mutate(ctx, "AIIntegrationUpdate", &mutation, variables); err != nil {
			ret = diag.Errorf("could not update AI integration: %v", internal.FromSpaceliftError(err))
		}
	}

	if d.HasChange("enabled") && !ret.HasError() {
		if err := setAIIntegrationEnabled(ctx, meta, d.Id(), d.Get("enabled").(bool)); err != nil {
			ret = append(ret, diag.Errorf("could not set enabled on the AI integration: %v", err)...)
		}
	}

	return append(ret, resourceAIIntegrationRead(ctx, d, meta)...)
}

func resourceAIIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// These cannot be deleted, and erroring would strand the resource in state
	// with `terraform state rm` the only way out.
	if d.Get("is_spacelift_provided").(bool) {
		d.SetId("")

		return diag.Diagnostics{{
			Severity: diag.Warning,
			Summary:  "Spacelift-provided AI integration removed from state only",
			Detail: "Spacelift-provided AI integrations are shared with every account and cannot be " +
				"deleted. The integration has been removed from the Terraform state and still exists " +
				"in Spacelift.",
		}}
	}

	var mutation struct {
		AIIntegration *struct {
			ID string `graphql:"id"`
		} `graphql:"aiIntegrationDelete(id: $id)"`
	}

	variables := map[string]any{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "AIIntegrationDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete AI integration: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

// setAIIntegrationEnabled flips the flag the create and update mutations do
// not accept.
func setAIIntegrationEnabled(ctx context.Context, meta any, id string, enabled bool) error {
	var mutation struct {
		AIIntegration struct {
			ID      string `graphql:"id"`
			Enabled bool   `graphql:"enabled"`
		} `graphql:"aiIntegrationToggle(id: $id, enabled: $enabled)"`
	}

	variables := map[string]any{
		"id":      toID(id),
		"enabled": toBool(enabled),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "AIIntegrationToggle", &mutation, variables); err != nil {
		return internal.FromSpaceliftError(err)
	}

	return nil
}
