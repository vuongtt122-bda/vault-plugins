$ cd vault-plugin-accounts-mock
$ go build -o vault/plugins/vault-plugin-accounts-mock cmd/vault-plugin-accounts-mock/main.go

$ vault server -dev -dev-root-token-id=root -dev-plugin-dir=./vault/plugins
$ export VAULT_ADDR='http://127.0.0.1:8200'
$ vault login root
$ vault secrets enable -path=mock-accounts vault-plugin-accounts-mock

$ vault secrets list

$ vault write mock-accounts/hello message="Hello World"
vault write mock-accounts/[key] message="[value]"

$ vault read mock-accounts/hello
vault read mock-accounts/[key]

$ vault delete mock-accounts/hello
vault delete mock-accounts/[key]

$ vault secrets disable mock-accounts

