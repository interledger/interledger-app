# This file is here till we figure out why the Go sdk won't allow you to assign roles.

export BOUNDARY_ADDR=https://boundary.fynbos.dev

cat << EOF > ./temp.hcl
kms "awskms" {
  purpose = "recovery"
  key_id = "global_recovery"
  kms_key_id = "$KMS_KEY_ID"
}
EOF

# allow anon to list orgs and authentication methods so you can login via web interface.
boundary roles add-principals -id $GLOBAL_ANON_ROLE_ID \
  -recovery-config ./temp.hcl \
  -principal 'u_anon'

boundary roles add-principals -id $ORG_ANON_ROLE_ID \
  -recovery-config ./temp.hcl \
  -principal 'u_anon'


# assign org and project admin roles to admin group
boundary roles add-principals -id $ORG_ADMIN_ROLE_ID \
  -recovery-config ./temp.hcl \
  -principal $ADMIN_GROUP_ID

boundary roles add-principals -id $PROJECT_ADMIN_ROLE_ID \
-recovery-config ./temp.hcl \
-principal $ADMIN_GROUP_ID

rm ./temp.hcl