package spacelift

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceWorkerPool() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_worker_pool` represents a worker pool assigned to the " +
			"Spacelift account.",

		CreateContext: resourceWorkerPoolCreate,
		ReadContext:   resourceWorkerPoolRead,
		UpdateContext: resourceWorkerPoolUpdate,
		DeleteContext: resourceWorkerPoolDelete,
		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, _ any) error {
			// Force the config to be recomputed if the CSR changes. Otherwise, it will ignore changes event if we `Set` new value.
			if diff.HasChange("csr") {
				diff.SetNewComputed("config")
			}

			return nil
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"config": {
				Type:        schema.TypeString,
				Description: "credentials necessary to connect WorkerPool's workers to the control plane",
				Computed:    true,
				Sensitive:   true,
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "name of the worker pool",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"csr": {
				Type:        schema.TypeString,
				Description: "certificate signing request in base64. Changing this value will trigger a token reset.",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "description of the worker pool",
				Optional:    true,
			},
			"private_key": {
				Type:        schema.TypeString,
				Description: "private key in base64",
				Computed:    true,
				Sensitive:   true,
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the worker pool is in",
				Optional:    true,
				Computed:    true,
			},
			"labels": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
		},
	}
}

func resourceWorkerPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	var mutation struct {
		WorkerPool *structs.WorkerPool `graphql:"workerPoolCreate(name: $name, certificateSigningRequest: $csr, description: $description, labels: $labels, space: $space)"`
	}

	variables := map[string]interface{}{
		"name":        graphql.String(name),
		"description": (*graphql.String)(nil),
		"labels":      []graphql.String(nil),
		"space":       (*graphql.ID)(nil),
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
		variables["labels"] = &labels
	}

	if spaceID, ok := d.GetOk("space_id"); ok {
		variables["space"] = graphql.NewID(spaceID)
	}

	if desc, ok := d.GetOk("description"); ok {
		variables["description"] = graphql.String(desc.(string))
	}

	if csrValue, ok := d.GetOk("csr"); ok {
		variables["csr"] = graphql.String(csrValue.(string))
	} else {
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return diag.Errorf("couldn't generate private key: %v", err)
		}

		subj := pkix.Name{
			CommonName: "workers.spacelift.io",
		}

		asn1Subj, err := asn1.Marshal(subj.ToRDNSequence())
		if err != nil {
			return diag.Errorf("couldn't marshal certificate subject: %v", err)
		}
		template := x509.CertificateRequest{
			RawSubject:         asn1Subj,
			SignatureAlgorithm: x509.SHA256WithRSA,
		}

		csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
		if err != nil {
			return diag.Errorf("couldn't create certificate request: %b", err)
		}

		cert := base64.StdEncoding.EncodeToString(
			pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes}),
		)

		privASN1, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return diag.Errorf("could not pkcs8 marshal private key: %v", err)
		}

		d.Set("csr", cert)
		d.Set("private_key", base64.StdEncoding.EncodeToString(
			pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privASN1}),
		))

		variables["csr"] = graphql.String(cert)
	}

	if err := meta.(*internal.Client).Mutate(ctx, "WorkerPoolCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create worker pool: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.WorkerPool.ID)
	d.Set("config", mutation.WorkerPool.Config)
	d.Set("name", mutation.WorkerPool.Name)
	d.Set("space_id", mutation.WorkerPool.Space)

	if description := mutation.WorkerPool.Description; description != nil {
		d.Set("description", *description)
	}

	return nil
}

func resourceWorkerPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		WorkerPool *structs.WorkerPool `graphql:"workerPool(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "WorkerPoolRead", &query, variables); err != nil {
		return diag.Errorf("could not query for worker pool: %v", err)
	}

	workerPool := query.WorkerPool
	if workerPool == nil {
		d.SetId("")
		return nil
	}

	d.Set("config", workerPool.Config)
	d.Set("name", workerPool.Name)

	if description := workerPool.Description; description != nil {
		d.Set("description", *description)
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range query.WorkerPool.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)
	d.Set("space_id", workerPool.Space)

	return nil
}

func resourceWorkerPoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	// If CSR has changed, use workerPoolReset mutation
	if d.HasChange("csr") {
		var resetMutation struct {
			WorkerPool structs.WorkerPool `graphql:"workerPoolReset(id: $id, certificateSigningRequest: $csr)"`
		}

		csrString := d.Get("csr").(string)
		resetVariables := map[string]interface{}{
			"id":  toID(d.Id()),
			"csr": graphql.String(csrString),
		}

		if err := meta.(*internal.Client).Mutate(ctx, "WorkerPoolReset", &resetMutation, resetVariables); err != nil {
			return diag.Errorf("could not reset worker pool token: %v", internal.FromSpaceliftError(err))
		}

		d.Set("config", resetMutation.WorkerPool.Config)

		if !d.HasChangeExcept("csr") {
			return resourceWorkerPoolRead(ctx, d, meta)
		}
	}

	// Handle regular update for other attributes
	var mutation struct {
		WorkerPool structs.WorkerPool `graphql:"workerPoolUpdate(id: $id, name: $name, description: $description, labels: $labels, space: $space)"`
	}

	variables := map[string]interface{}{
		"id":          toID(d.Id()),
		"name":        graphql.String(name),
		"description": (*graphql.String)(nil),
		"labels":      []graphql.String(nil),
		"space":       (*graphql.ID)(nil),
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
		variables["labels"] = &labels
	}

	if desc, ok := d.GetOk("description"); ok {
		variables["description"] = graphql.String(desc.(string))
	}

	if spaceID, ok := d.GetOk("space_id"); ok {
		variables["space"] = graphql.NewID(spaceID)
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "WorkerPoolUpdate", &mutation, variables); err != nil {
		ret = diag.Errorf("could not update worker pool: %v", internal.FromSpaceliftError(err))
	}

	return append(ret, resourceWorkerPoolRead(ctx, d, meta)...)
}

func resourceWorkerPoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		WorkerPool *structs.WorkerPool `graphql:"workerPoolDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "WorkerPoolDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete worker pool: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
