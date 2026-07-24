package spacelift

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

const (
	repoFilePath          = "path"
	repoFileContent       = "content"
	repoFileMode          = "file_mode"
	repoFileEncrypt       = "encrypt"
	repoFileCommitMessage = "commit_message"
	repoFileAuthorName    = "author_name"
	repoFileAuthorEmail   = "author_email"
	repoFileRevisionSHA   = "revision_sha"
	repoFileSizeBytes     = "size_bytes"

	repoFileDefaultMode          = "0644"
	repoFileDefaultCommitMessage = "Managed by Terraform"
	repoFileDefaultAuthorName    = "Terraform"
)

func resourceRepoFile() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_repo_file` represents a single file in a " +
			"[Spacelift repo](repo.md). Every change commits a new revision, " +
			"which triggers a tracked run on each stack attached to the repo.",

		CreateContext: resourceRepoFileCreate,
		ReadContext:   resourceRepoFileRead,
		UpdateContext: resourceRepoFileUpdate,
		DeleteContext: resourceRepoFileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRepoFileImport,
		},

		Schema: map[string]*schema.Schema{
			repoID: {
				Type:             schema.TypeString,
				Description:      "ID (slug) of the repo the file belongs to",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			repoFilePath: {
				Type:             schema.TypeString,
				Description:      "Path of the file relative to the repo root, without a leading slash, for example `modules/vpc/main.tf`",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			repoFileContent: {
				Type:        schema.TypeString,
				Description: "Contents of the file. Text only - binary files are not supported.",
				Required:    true,
			},
			repoFileMode: {
				Type:        schema.TypeString,
				Description: "Octal file permissions, for example `0755`. Defaults to `0644`.",
				Optional:    true,
				Default:     repoFileDefaultMode,
			},
			repoFileEncrypt: {
				Type: schema.TypeBool,
				Description: "Encrypt the file at rest. Spacelift never returns the contents of an encrypted file, " +
					"so Terraform cannot detect changes made to it outside of Terraform. Defaults to `false`.",
				Optional: true,
				Default:  false,
			},
			repoFileCommitMessage: {
				Type:        schema.TypeString,
				Description: "Message of the commit created for each change to the file. Defaults to `Managed by Terraform`.",
				Optional:    true,
				Default:     repoFileDefaultCommitMessage,
			},
			repoFileAuthorName: {
				Type:        schema.TypeString,
				Description: "Author name recorded on the commit. Defaults to `Terraform`.",
				Optional:    true,
				Default:     repoFileDefaultAuthorName,
			},
			repoFileAuthorEmail: {
				Type:        schema.TypeString,
				Description: "Author email recorded on the commit",
				Optional:    true,
			},
			repoFileRevisionSHA: {
				Type:        schema.TypeString,
				Description: "SHA of the revision that last changed the file",
				Computed:    true,
			},
			repoFileSizeBytes: {
				Type:        schema.TypeInt,
				Description: "Size of the file in bytes",
				Computed:    true,
			},
		},
	}
}

func resourceRepoFileCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	revision, err := commitRepoFile(ctx, d, meta, "CREATE")
	if err != nil {
		return diag.Errorf("could not create the repo file: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(repoFileID(d.Get(repoID).(string), d.Get(repoFilePath).(string)))
	d.Set(repoFileRevisionSHA, revision.SHA)

	return resourceRepoFileRead(ctx, d, meta)
}

func resourceRepoFileRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var query struct {
		RepoFileHistory struct {
			Edges []struct {
				Node structs.RepoFile `graphql:"node"`
			} `graphql:"edges"`
		} `graphql:"repoFileHistory(repoID: $repoID, filePath: $filePath, first: $first)"`
	}

	first := graphql.Int(1)
	variables := map[string]any{
		"repoID":   toID(d.Get(repoID)),
		"filePath": toString(d.Get(repoFilePath)),
		"first":    &first,
	}

	if err := meta.(*internal.Client).Query(ctx, "RepoFileRead", &query, variables); err != nil {
		return diag.Errorf("could not query for repo file: %v", err)
	}

	edges := query.RepoFileHistory.Edges
	if len(edges) == 0 || edges[0].Node.IsDeleted {
		d.SetId("")
		return nil
	}

	file := edges[0].Node

	d.Set(repoFileRevisionSHA, file.Revision.SHA)

	mode := repoFileDefaultMode
	if file.FileMode != nil {
		mode = *file.FileMode
	}
	d.Set(repoFileMode, mode)

	if file.SizeBytes != nil {
		d.Set(repoFileSizeBytes, *file.SizeBytes)
	}

	if content := file.Content; content != nil {
		d.Set(repoFileEncrypt, content.IsEncrypted)

		// Spacelift withholds the contents of encrypted files, so the state
		// keeps whatever Terraform last wrote.
		if content.Content != nil {
			decoded, err := base64.StdEncoding.DecodeString(*content.Content)
			if err != nil {
				return diag.Errorf("could not decode contents of repo file %s: %v", file.FilePath, err)
			}
			d.Set(repoFileContent, string(decoded))
		}
	}

	return nil
}

func resourceRepoFileUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	revision, err := commitRepoFile(ctx, d, meta, "UPDATE")
	if err != nil {
		return diag.Errorf("could not update the repo file: %v", internal.FromSpaceliftError(err))
	}

	d.Set(repoFileRevisionSHA, revision.SHA)

	return resourceRepoFileRead(ctx, d, meta)
}

func resourceRepoFileDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if _, err := commitRepoFile(ctx, d, meta, "DELETE"); err != nil {
		return diag.Errorf("could not delete the repo file: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func commitRepoFile(ctx context.Context, d *schema.ResourceData, meta any, action string) (*structs.Revision, error) {
	var mutation struct {
		CreateRevision structs.Revision `graphql:"revisionCreate(input: $input)"`
	}

	file := structs.RevisionFileInput{
		Path:   toString(d.Get(repoFilePath)),
		Action: toOptionalString(action),
	}

	// A deletion carries no content, and sending the mode would make the
	// backend treat it as a change to a file that is going away.
	if action != "DELETE" {
		encoded := graphql.String(base64.StdEncoding.EncodeToString([]byte(d.Get(repoFileContent).(string))))
		encrypt := graphql.Boolean(d.Get(repoFileEncrypt).(bool))

		file.Content = &encoded
		file.Encrypt = &encrypt
		file.FileMode = toOptionalString(d.Get(repoFileMode))
	}

	input := structs.RevisionCreateInput{
		RepoID:     toID(d.Get(repoID)),
		Message:    toString(d.Get(repoFileCommitMessage)),
		AuthorName: toString(d.Get(repoFileAuthorName)),
		Files:      []structs.RevisionFileInput{file},
	}

	if email, ok := d.GetOk(repoFileAuthorEmail); ok {
		input.AuthorEmail = toOptionalString(email)
	}

	variables := map[string]any{"input": input}

	if err := meta.(*internal.Client).Mutate(ctx, "RevisionCreate", &mutation, variables); err != nil {
		return nil, err
	}

	return &mutation.CreateRevision, nil
}

func resourceRepoFileImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	repoSlug, path, found := strings.Cut(d.Id(), "/")
	if !found || repoSlug == "" || path == "" {
		return nil, fmt.Errorf("expected an ID of the form $repoID/$path, got %q", d.Id())
	}

	d.Set(repoID, repoSlug)
	d.Set(repoFilePath, path)

	return []*schema.ResourceData{d}, nil
}

func repoFileID(repoSlug, path string) string {
	return repoSlug + "/" + path
}
