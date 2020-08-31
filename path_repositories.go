package main

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/policyutil"
	"github.com/hashicorp/vault/sdk/logical"
)

const REPO_PATH_PREFIX = "repositories"

func (b *backend) pathRepositoriesList() *framework.Path {
	return &framework.Path{
		Pattern: REPO_PATH_PREFIX + "/?$",

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{
				Callback: b.pathRepositoryList,
			},
		},

		HelpSynopsis:    pathRepositoryHelpSyn,
		HelpDescription: pathRepositoryHelpDesc,
	}
}

func (b *backend) pathRepositories() *framework.Path {
	return &framework.Path{
		Pattern: REPO_PATH_PREFIX + `/(?P<name>.+)`,
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Name of the Github repository.",
			},

			"policies": {
				Type:        framework.TypeCommaStringSlice,
				Description: "Comma-separated list of policies associated to the repository.",
			},
		},

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathRepositoryDelete,
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathRepositoryRead,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathRepositoryWrite,
			},
		},

		HelpSynopsis:    pathRepositoryHelpSyn,
		HelpDescription: pathRepositoryHelpDesc,
	}
}

func (b *backend) Repository(ctx context.Context, s logical.Storage, n string) (*RepositoryEntry, error) {
	entry, err := s.Get(ctx, "repository/"+n)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var result RepositoryEntry
	if err := entry.DecodeJSON(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (b *backend) pathRepositoryDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete(ctx, "repository/"+d.Get("name").(string))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathRepositoryRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	repository, err := b.Repository(ctx, req.Storage, d.Get("name").(string))
	if err != nil {
		return nil, err
	}
	if repository == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"policies": repository.Policies,
		},
	}, nil
}

func (b *backend) pathRepositoryWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	// Store it
	entry, err := logical.StorageEntryJSON("repository/"+d.Get("name").(string), &RepositoryEntry{
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

func (b *backend) pathRepositoryList(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	repositories, err := req.Storage.List(ctx, "repository/")
	if err != nil {
		return nil, err
	}
	return logical.ListResponse(repositories), nil
}

type RepositoryEntry struct {
	Policies []string
}

const pathRepositoryHelpSyn = `
Manage repositories allowed to authenticate.
`

const pathRepositoryHelpDesc = `
This endpoint allows you to create, read, update, and delete configuration
for repositories that are allowed to authenticate, and associate policies to
them.
`
