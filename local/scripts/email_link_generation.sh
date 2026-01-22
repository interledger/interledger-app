#!/bin/bash

PASSW_RECOVERY_PATH="/recovery"
EMAIL_CONF_PATH="/verification"

if [[ -z "$1" ]]; then
  echo "Usage: sh email_link_generation.sh <passowrd or recovery> <email>" >&2
  exit 1
elif [[ -z "$2" ]]; then
  echo "Usage: sh email_link_generation.sh <passowrd or recovery> <email>" >&2
  exit 1
fi

flow_id=$(docker exec local-postgres-1 psql -U postgres -d kratos -t -A -c "SELECT selfservice_recovery_flow_id FROM public.identity_recovery_tokens ORDER BY issued_at DESC LIMIT 1;")
token=$(docker logs  local-kratos-1 2>&1 \
  | grep $2 \
  | sed -n 's/.*recovery_link_token=\([^ ]*\).*/\1/p' \
  | tail -n 1)

PATH=""
if [[ "$1" == "recovery" ]]; then
  PATH=$PASSW_RECOVERY_PATH
elif [[ "$1" == "password" ]]; then
  PATH=$EMAIL_CONF_PATH
else
  echo "Error: invalid first argument. Usage: sh email_link_generation.sh <passowrd or recovery> <email>" >&2
  exit 1
fi

echo "👇🏼 Link 👇🏼"
echo "https://interledger.test/self-service$PATH?flow=$flow_id&token=$token"
echo "// Make sure you just requested the confirmation, the token is always the last token without considering the user"

