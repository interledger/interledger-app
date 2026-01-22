#!/bin/bash

PASSW_RECOVERY_PATH="recovery"
EMAIL_CONF_PATH="verification"

if [[ -z "$1" || -z "$2" ]]; then
  echo "Usage: sh email_link_generation.sh <password or recovery> <email>" >&2
  exit 1
fi

FLOW_PATH=""
if [[ "$1" == "recovery" ]]; then
  FLOW_PATH=$PASSW_RECOVERY_PATH
elif [[ "$1" == "password" ]]; then
  FLOW_PATH=$EMAIL_CONF_PATH
else
  echo "Error: invalid first argument. Usage: sh email_link_generation.sh <password or recovery> <email>" >&2
  exit 1
fi

# Fetch Flow ID and Token
flow_id=$(docker exec local-postgres-1 psql -U postgres -d kratos -t -A -c "SELECT selfservice_${FLOW_PATH}_flow_id FROM public.identity_${FLOW_PATH}_tokens ORDER BY issued_at DESC LIMIT 1;")
token=$(docker logs local-kratos-1 2>&1 \
  | grep "$2" \
  | sed -n "s/.*${FLOW_PATH}_link_token=\([^ ]*\).*/\1/p" \
  | tail -n 1)

# Display Link
echo "👇🏼 Link 👇🏼"
echo "https://interledger.test/self-service/$FLOW_PATH?flow=$flow_id&token=$token"
echo "// Make sure you just requested the confirmation, the token is always the last token without considering the user"
