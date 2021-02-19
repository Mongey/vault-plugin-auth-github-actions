#!/usr/bin/env bash

set -ex

GOOS=linux go build

docker kill vaultplg 2>/dev/null || true
mkdir -p tmp
tmpdir=$(mktemp -d tmp/vaultplgXXXXXX)
mkdir "$tmpdir/data"
docker run --rm -d -p8200:8200 --name vaultplg -v "$(pwd)/$tmpdir/data":/data -v $(pwd):/example --cap-add=IPC_LOCK -e 'VAULT_LOCAL_CONFIG=
{
  "backend": {"file": {"path": "/data"}},
  "listener": [{"tcp": {"address": "0.0.0.0:8200", "tls_disable": true}}],
  "plugin_directory": "/example",
  "log_level": "debug",
  "ui": true,
  "disable_mlock": true,
  "api_addr": "http://localhost:8200"
}
' vault server
sleep 1

export VAULT_ADDR=http://localhost:8200

initoutput=$(vault operator init -key-shares=1 -key-threshold=1 -format=json)
vault operator unseal $(echo "$initoutput" | jq -r .unseal_keys_hex[0])

export VAULT_TOKEN=$(echo "$initoutput" | jq -r .root_token)

vault write sys/plugins/catalog/auth/github-actions-auth-plugin \
    sha_256=$(shasum -a 256 vault-plugin-auth-github-actions | cut -d' ' -f1) \
    command="vault-plugin-auth-github-actions"

export ADMIN_POLICY='path "secret/*" { capabilities = ["create", "read", "update", "delete", "list",
"sudo"] }'

vault auth enable \
    -path="github-actions" \
    -plugin-name="github-actions-auth-plugin" plugin

vault audit enable file file_path=stdout log_raw=true
echo "$ADMIN_POLICY" | vault policy write admin -
echo "$ADMIN_POLICY" | vault policy write admin2 -


vault write auth/github-actions/repositories/Mongey/vault-plugin-auth-github-actions policies=admin2
vault write auth/github-actions/organizations/Mongey policies=admin

echo $VAULT_TOKEN
vault secrets enable -path=secret -version=2 kv
vault kv put secret/ci npmToken=my-long-passcode
docker logs -f vaultplg
