package spacelift

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"

	"github.com/fluxio/multierror"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func resourceWorkerPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkerPoolCreate,
		Read:   resourceWorkerPoolRead,
		Update: resourceWorkerPoolUpdate,
		Delete: resourceWorkerPoolDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"config": &schema.Schema{
				Type:        schema.TypeString,
				Description: "credentials necessary to connect WorkerPool's workers to the control plane",
				Computed:    true,
				Sensitive:   true,
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "name of the worker pool",
				Required:    true,
			},
			"csr": &schema.Schema{
				Type:        schema.TypeString,
				Description: "certificate signing request in base64",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "description of the worker pool",
				Optional:    true,
			},
			"private_key": &schema.Schema{
				Type:        schema.TypeString,
				Description: "private key in base64",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceWorkerPoolCreate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)

	var mutation struct {
		WorkerPool *structs.WorkerPool `graphql:"workerPoolCreate(name: $name, certificateSigningRequest: $csr, description: $description)"`
	}

	variables := map[string]interface{}{
		"name":        graphql.String(name),
		"description": (*graphql.String)(nil),
	}

	if desc, ok := d.GetOk("description"); ok {
		variables["description"] = graphql.String(desc.(string))
	}

	if desc, ok := d.GetOk("csr"); ok {
		variables["csr"] = graphql.String(desc.(string))
	} else {
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return errors.Wrap(err, "couldn't generate private key")
		}

		subj := pkix.Name{
			CommonName: "workers.spacelift.io",
		}

		asn1Subj, err := asn1.Marshal(subj.ToRDNSequence())
		if err != nil {
			return errors.Wrap(err, "couldn't marshal certificate subject")
		}
		template := x509.CertificateRequest{
			RawSubject:         asn1Subj,
			SignatureAlgorithm: x509.SHA256WithRSA,
		}

		csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
		if err != nil {
			return errors.Wrap(err, "couldn't create certificate request")
		}

		cert := base64.StdEncoding.EncodeToString(
			pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes}),
		)

		privASN1, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return errors.Wrap(err, "could not pkcs8 marshal private key")
		}

		d.Set("csr", cert)
		d.Set("private_key", base64.StdEncoding.EncodeToString(
			pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privASN1}),
		))

		variables["csr"] = graphql.String(cert)
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create worker pool")
	}

	d.SetId(mutation.WorkerPool.ID)
	d.Set("config", mutation.WorkerPool.Config)
	d.Set("name", mutation.WorkerPool.Name)

	if description := mutation.WorkerPool.Description; description != nil {
		d.Set("description", *description)
	}

	return nil
}

func resourceWorkerPoolRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		WorkerPool *structs.WorkerPool `graphql:"workerPool(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}
	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for worker pool")
	}

	workerPool := query.WorkerPool
	if workerPool == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", workerPool.Name)
	if description := workerPool.Description; description != nil {
		d.Set("description", *description)
	}

	return nil
}

func resourceWorkerPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)

	var mutation struct {
		WorkerPool structs.WorkerPool `graphql:"workerPoolUpdate(id: $id, name: $name, description: $description)"`
	}

	variables := map[string]interface{}{
		"id":          toID(d.Id()),
		"name":        graphql.String(name),
		"description": (*graphql.String)(nil),
	}

	if desc, ok := d.GetOk("description"); ok {
		variables["description"] = graphql.String(desc.(string))
	}

	var acc multierror.Accumulator

	acc.Push(errors.Wrap(meta.(*Client).Mutate(&mutation, variables), "could not update worker pool"))
	acc.Push(errors.Wrap(resourceWorkerPoolRead(d, meta), "could not read the current state"))

	return acc.Error()
}

func resourceWorkerPoolDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		WorkerPool *structs.WorkerPool `graphql:"workerPoolDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete worker pool")
	}

	d.SetId("")

	return nil
}
