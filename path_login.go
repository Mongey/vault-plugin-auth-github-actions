package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	RUN_IN_PROGRESS = "in_progress"
)

func (b *backend) pathLogin() *framework.Path {
	return &framework.Path{
		Pattern: "login",
		Fields: map[string]*framework.FieldSchema{
			"token": {
				Type: framework.TypeString,
			},
			"owner": {
				Type: framework.TypeString,
			},
			"repository": {
				Type: framework.TypeString,
			},
			"run_id": {
				Type: framework.TypeString,
			},
			"run_number": {
				Type: framework.TypeInt,
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathAuthLogin,
		},
	}
}

func (b *backend) pathAuthLogin(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	token := d.Get("token").(string)
	owner := d.Get("owner").(string)
	fullRepoName := d.Get("repository").(string)
	inputRunID := d.Get("run_id").(string)
	runNumber := d.Get("run_number").(int)

	runID, err := strconv.ParseInt(inputRunID, 10, 64)
	if err != nil {
		return nil, err
	}

	repository := repositoryName(fullRepoName, owner)
	client := githubClientFromToken(ctx, token)
	run, _, err := client.Actions.GetWorkflowRunByID(context.Background(), owner, repository, runID)
	if err != nil {
		return nil, err
	}

	if *run.Status != RUN_IN_PROGRESS && *run.RunNumber != runNumber {
		return nil, fmt.Errorf("Run is %s, expected '%s'", *run.Status, RUN_IN_PROGRESS)
	}

	var policies []string
	organizationEntry, err := b.Organization(ctx, req.Storage, owner)
	if err != nil {
		b.Logger().Warn(fmt.Sprintf("unable to retrieve %s: %s", owner, err.Error()))
	}

	if organizationEntry == nil {
		b.Logger().Debug(fmt.Sprintf("unable to find %s, does not currently exist", owner))
	}

	policies = append(policies, organizationEntry.Policies...)
	repositoryEntry, err := b.Repository(ctx, req.Storage, fullRepoName)
	if err != nil {
		b.Logger().Warn(fmt.Sprintf("unable to retrieve %s: %s", fullRepoName, err.Error()))
	}

	if repositoryEntry == nil {
		b.Logger().Debug(fmt.Sprintf("unable to find %s, does not currently exist", fullRepoName))
	}
	policies = append(policies, repositoryEntry.Policies...)

	return &logical.Response{
		Auth: &logical.Auth{
			InternalData: map[string]interface{}{
				"token":      token,
				"owner":      owner,
				"repository": repository,
				"run_id":     runID,
				"run_number": runNumber,
			},
			Policies: policies,
			Metadata: map[string]string{
				"owner": owner,
			},
			LeaseOptions: logical.LeaseOptions{
				TTL:       30 * time.Second,
				MaxTTL:    60 * time.Minute,
				Renewable: true,
			},
		},
	}, nil
}
