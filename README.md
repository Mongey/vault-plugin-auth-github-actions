# vault-plugin-auth-github-actions
Authenticate with vault from your github actions.


## Setup

1. Download and decompress the latest plugin binary from the Releases tab on
GitHub. Alternatively you can compile the plugin from source.

2. Move the compiled plugin into Vault's configured `plugin_directory`:

  ```sh
  $ mv vault-plugin-auth-github-actions /etc/vault/plugins/vault-plugin-auth-github-actions
  ```

3. Calculate the SHA256 of the plugin and register it in Vault's plugin catalog.
If you are downloading the pre-compiled binary, it is highly recommended that
you use the published checksums to verify integrity.

  ```sh
  $ export SHA256=$(shasum -a 256 "/etc/vault/plugins/vault-plugin-auth-github-actions" | cut -d' ' -f1)

  $ vault write sys/plugins/catalog/auth/github-actions-auth-plugin \
      sha_256="${SHA256}" \
      command="vault-plugin-auth-github-actions"
  ```

4. Mount the auth method:

  ```sh
  $ vault auth enable \
      -path="github-actions" \
      -plugin-name="auth-github-actions" plugin
  ```

5. Configure the role your repository should assume
  ```sh
  $ vault write auth/github-actions/repositories/Mongey/vault-plugin-auth-github-actions policies=admin
  ```

6. Point your github action to import your secrets from Vault
```yaml
      - name: Import Secrets
        id: secrets
        uses: hashicorp/vault-action@v2.0.0
        with:
          url: https://my-vault-server.org:8200
          method: github-actions
          secrets: secret/data/ci npmToken | NPM_TOKEN
          authPayload: |
          '{
            "token": "${{ secrets.GITHUB_TOKEN }}",
            "run_id": "${{ github.run_id }}",
            "run_number": "${{ github.run_number }}",
            "owner": "${{ github.repository_owner }}",
            "repository": "${{ github.repository }}"
          }'
      - name: Print
        env:
          MY_VAR: Hello
          FOO: ${{ steps.secrets.outputs.NPM_TOKEN }}
        run: |
          echo $MY_VAR $FOO $NPM_TOKEN
```


### Assign a default policy to all repositories in your organization

```
$ vault write auth/github-actions/organizations/Mongey policies=admin
```
