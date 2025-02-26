if [ -f configuration/vault-tokens.txt ]; then
  export VAULT_TOKEN=$(cat configuration/vault-tokens.txt | grep 'Root' | cut -d':' -f2 | xargs)
else
    echo "No vault-tokens.txt found"
fi