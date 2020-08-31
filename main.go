package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/sdk/plugin"
	"golang.org/x/oauth2"
)

func main() {
	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	flags.Parse(os.Args[1:])

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	if err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: Factory,
		TLSProviderFunc:    tlsProviderFunc,
	}); err != nil {
		log.Fatal(err)
	}
}

func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
	b := Backend(c)
	if err := b.Setup(ctx, c); err != nil {
		return nil, err
	}
	return b, nil
}

type backend struct {
	*framework.Backend
}

func Backend(c *logical.BackendConfig) *backend {
	var b backend

	paths := []*framework.Path{
		b.pathLogin(),
		b.pathOrganizations(),
		b.pathRepositories(),
	}

	b.Backend = &framework.Backend{
		BackendType: logical.TypeCredential,
		AuthRenew:   b.pathAuthRenew,
		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{"login"},
		},
		Paths: paths,
	}

	return &b
}

func githubClientFromToken(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func repositoryName(fullRepoName, owner string) string {
	return strings.Replace(fullRepoName, fmt.Sprintf("%s/", owner), "", 1)
}

func (b *backend) pathAuthRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	if req.Auth == nil {
		return nil, errors.New("request auth was nil")
	}

	token := req.Auth.InternalData["token"].(string)
	owner := req.Auth.InternalData["owner"].(string)
	repository := req.Auth.InternalData["repository"].(string)
	runID := req.Auth.InternalData["run_id"].(int64)
	runNumber := req.Auth.InternalData["run_number"].(int)

	client := githubClientFromToken(ctx, token)

	run, _, err := client.Actions.GetWorkflowRunByID(context.Background(), owner, repository, runID)
	if err != nil {
		return nil, err
	}

	if *run.Status != "in_progress" && *run.RunNumber != runNumber {
		return nil, fmt.Errorf("Run is %s, expected 'in_progress'", *run.Status)
	}

	return framework.LeaseExtend(30*time.Second, 60*time.Minute, b.System())(ctx, req, d)
}
