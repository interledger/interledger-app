#!/bin/bash

PASSW_RECOVERY_PATH="/recovery"
EMAIL_CONF_PATH="/verification"

if [ -z "$1" ] || [ -z "$2" ]; then
  echo "Usage: ./email_link_generation.sh <password|recovery> <email>" >&2
  exit 1
fi

mode="$1"
email="$2"

query=""
link_path=""

case "$mode" in
  recovery)
    link_path=$PASSW_RECOVERY_PATH
    query="SELECT irt.selfservice_recovery_flow_id AS flow_id, irt.token AS token FROM public.identity_recovery_tokens irt JOIN public.identity_recovery_addresses ira ON irt.identity_recovery_address_id = ira.id WHERE ira.value = '$email' ORDER BY irt.issued_at DESC LIMIT 1;"
    ;;
  password)
    link_path=$EMAIL_CONF_PATH
    query="SELECT ivt.selfservice_verification_flow_id AS flow_id, ivt.token AS token FROM public.identity_verification_tokens ivt JOIN public.identity_verifiable_addresses iva ON ivt.identity_verifiable_address_id = iva.id WHERE iva.value = '$email' ORDER BY ivt.issued_at DESC LIMIT 1;"
    ;;
  *)
    echo "Error: invalid first argument. Usage: ./email_link_generation.sh <password|recovery> <email>" >&2
    exit 1
    ;;
esac

result=$(docker exec local-postgres-1 psql -U postgres -d kratos -t -A -F '|' -c "$query")
flow_id="${result%%|*}"
token="${result#*|}"

if [ -z "$flow_id" ] || [ -z "$token" ]; then
  echo "No flow/token found. Trigger a new $mode email for $email and try again." >&2
  exit 1
fi

echo "👇🏼 Link 👇🏼"
echo "https://interledger.test/self-service${link_path}?flow=$flow_id&token=$token"
echo "// Make sure you just requested the confirmation; the link uses the most recent token for $email"

