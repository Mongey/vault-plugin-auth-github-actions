package main

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/policyutil"
	"github.com/hashicorp/vault/sdk/logical"
)

const PATH_PREFIX = "organizations"

func (b *backend) pathOrganizationsList() *framework.Path {
	return &framework.Path{
		Pattern: PATH_PREFIX + "/?$",

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{
				Callback: b.pathOrganizationList,
			},
		},

		HelpSynopsis:    pathOrganizationHelpSyn,
		HelpDescription: pathOrganizationHelpDesc,
	}
}

func (b *backend) pathOrganizations() *framework.Path {
	return &framework.Path{
		Pattern: PATH_PREFIX + `/(?P<name>.+)`,
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Name of the Github organization.",
			},

			"policies": {
				Type:        framework.TypeCommaStringSlice,
				Description: "Comma-separated list of policies associated to the organization.",
			},
		},

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathOrganizationDelete,
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathOrganizationRead,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathOrganizationWrite,
			},
		},

		HelpSynopsis:    pathOrganizationHelpSyn,
		HelpDescription: pathOrganizationHelpDesc,
	}
}

func (b *backend) Organization(ctx context.Context, s logical.Storage, n string) (*OrganizationEntry, error) {
	entry, err := s.Get(ctx, "organization/"+n)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var result OrganizationEntry
	if err := entry.DecodeJSON(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (b *backend) pathOrganizationDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete(ctx, "organization/"+d.Get("name").(string))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathOrganizationRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	organization, err := b.Organization(ctx, req.Storage, d.Get("name").(string))
	if err != nil {
		return nil, err
	}
	if organization == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"policies": organization.Policies,
		},
	}, nil
}

func (b *backend) pathOrganizationWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	// Store it
	entry, err := logical.StorageEntryJSON("organization/"+d.Get("name").(string), &OrganizationEntry{
		Policies: policyutil.ParsePolicies(d.Get("policies")),
	})
	if err != nil {
		return nil, err
	}
	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathOrganizationList(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	organizations, err := req.Storage.List(ctx, "organization/")
	if err != nil {
		return nil, err
	}
	return logical.ListResponse(organizations), nil
}

type OrganizationEntry struct {
	Policies []string
}

const pathOrganizationHelpSyn = `
Manage organizations allowed to authenticate.
`

const pathOrganizationHelpDesc = `
This endpoint allows you to create, read, update, and delete configuration
for organizations that are allowed to authenticate, and associate policies to
them. `
