table "agreement_signatures" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "agreement_id" {
    null = false
    type = text
  }
  column "user_id" {
    null = false
    type = text
  }
  column "ip_address" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_agreement" {
    columns     = [column.agreement_id]
    ref_columns = [table.agreements.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "user_id_ind" {
    columns = [column.user_id]
  }
}
table "agreements" {
  schema = schema.public
  column "id" {
    null = false
    type = text
  }
  column "name" {
    null = false
    type = text
  }
  column "version" {
    null = false
    type = text
  }
  column "content" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
}
table "individual_kyc_details" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "revision" {
    null = false
    type = bigint
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "country_code" {
    null = false
    type = text
  }
  column "first_name" {
    null = false
    type = text
  }
  column "last_name" {
    null = false
    type = text
  }
  column "gender" {
    null = false
    type = bigint
  }
  column "date_of_birth" {
    null = true
    type = date
  }
  column "ip_address" {
    null = false
    type = text
  }
  column "place_of_birth" {
    null = true
    type = text
  }
  column nationality {
    null = true
    type = text
  }
  column "address" {
    null = true
    type = jsonb
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "wallet_id_ind" {
    columns = [column.wallet_id]
  }
  index "wallet_id_revision_uniq" {
    unique  = true
    columns = [column.wallet_id, column.revision]
  }
}
table "linked_accounts" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "name" {
    null = false
    type = text
  }
  column "mask" {
    null = false
    type = text
  }
  column "provider" {
    null = false
    type = text
  }
  column "provider_id" {
    null = false
    type = text
  }
  column "type" {
    null = false
    type = text
  }
  column "nickname" {
    null = false
    type = text
    default = ""
  }
  column "state" {
    null = false
    type = text
    default = ""
  }
  column "can_send" {
    type = bool
    default = false
  }
  column "can_receive" {
    type = bool
    default = false
  }
  column "default_receive" {
    type = bool
    default = false
  }
  column "default_send" {
    type = bool
    default = false
  }
  column "send_country" {
    null = false
    type = text
    default = ""
  }
  column "send_currency" {
    null = false
    type = text
    default = ""
  }
  column "send_network" {
    null = false
    type = text
    default = ""
  }
  column "send_availability" {
    null = false
    type = text
    default = ""
  }
  column "receive_country" {
    null = false
    type = text
    default = ""
  }
  column "receive_currency" {
    null = false
    type = text
    default = ""
  }
  column "receive_network" {
    null = false
    type = text
    default = ""
  }
  column "receive_availability" {
    null = false
    type = text
    default = ""
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "deleted_at" {
    null = true
    type = timestamp
  }
  primary_key {
    columns = [column.id]
  }
  index "wallet_id_provider_provider_id_uniq" {
    unique  = true
    columns = [column.wallet_id, column.provider, column.provider_id]
  }
  index "linked_accounts_wallet_id" {
    unique  = false
    columns = [column.wallet_id]
  }
  index "wallet_can_receive_state" {
    unique  = false
    columns = [column.wallet_id, column.can_receive, column.state]
  }
}
table "linked_account_reviews" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "linked_account_id" {
    null = false
    type = uuid
  }
  column "reason" {
    null = false
    type = text
    default = ""
  }
  column "state" {
    null = false
    type = text
    default = ""
  }
  column "new_state" {
    null = false
    type = text
    default = ""
  }
  column "reviewed_by" {
    null = false
    type = text
    default = ""
  }
  index "review_linked_account_index" {
    unique  = false
    columns = [column.linked_account_id]
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "completed_at" {
    null = true
    type = timestamp
  }
  primary_key {
    columns = [column.id]
  }
}
table "basistheory_cards" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "token_id" {
    null = false
    type = text
  }
  column "expiration_month" {
    null = false
    type = text
  }
  column "expiration_year" {
    null = false
    type = text
  }
  column "tokenized_number" {
    null = false
    type = text
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "fingerprint" {
    null = false
    type = text
    default = ""
  }
  column "bin" {
    default = ""
    null = false
    type = text
  }
  column "pull_network" {
    default = ""
    null = false
    type = text
  }
  column "pull_enabled" {
    default = false
    null = false
    type = bool
  }
  column "pull_type" {
    default = ""
    null = false
    type = text
  }
  column "pull_country" {
    default = ""
    null = false
    type = text
  }
  column "push_network" {
    default = ""
    null = false
    type = text
  }
  column "push_enabled" {
    default = false
    null = false
    type = bool
  }
  column "push_type" {
    default = ""
    null = false
    type = text
  }
  column "push_availability" {
    default = ""
    null = false
    type = text
  }
  column "push_country" {
    default = ""
    null = false
    type = text
  }  
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "deleted_at" {
    null = true
    type = timestamp
  }
  primary_key {
    columns = [column.id]
  }
  index "token_id_wallet_id_uniq" {
    unique  = true
    columns = [column.token_id, column.wallet_id]
  }
}
table "openpayments_incoming_payment" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "payment_pointer_id" {
    null = true
    type = uuid
  }
  column "receiver_wallet_address" {
    null = true
    type = text
  }
  column "asset_code" {
    null = true
    type = text
  }
  column "asset_scale" {
    null = true
    type = bigint
  }
  column "incoming_amount" {
    null = false
    type = bigint
  }
  column "received_amount" {
    null = false
    type = bigint
  }
  column "completed" {
    null    = true
    type    = boolean
    default = false
  }
  column "expires_at" {
    null = true
    type = timestamp
  }
  column "external_ref" {
    null = true
    type = text
  }
  column "ilp_stream_id" {
    null = true
    type = text
  }
  column "ilp_address" {
    null = true
    type = text
  }
  column "ilp_shared_secret" {
    null = true
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "from_payment_pointer_id" {
    null = true
    type = uuid
  }
  column "sender_wallet_address" {
    null = true
    type = text
  }
  column "description" {
    null = true
    type = text
  }
  column "created_by" {
    null = true
    type = text
  }
  primary_key {
    columns = [column.id]
  }
}
table "openpayments_outgoing_payment" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "quote_id" {
    null = false
    type = uuid
  }
  column "failed" {
    null    = false
    type    = boolean
    default = false
  }
  column "description" {
    null = false
    type = text
  }
  column "sent_amount" {
    null = false
    type = bigint
  }
  column "sent_asset" {
    null = false
    type = text
  }
  column "sent_scale" {
    null = false
    type = bigint
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "to_payment_pointer_id" {
    null = true
    type = uuid
  }
  column "receiver_wallet_address" {
    null = true
    type = text
  }
  column "sender_wallet_address" {
    null = true
    type = text
  }
  column "completed" {
    null    = false
    type    = boolean
    default = false
  }
  column "created_by" {
    null = true
    type = text
  }
  primary_key {
    columns = [column.id]
  }
  index "outgoing_quote_id" {
    unique  = true
    columns = [column.quote_id]
  }
  foreign_key "fk_quote_id_ref_openpayments_quotes" {
    columns     = [column.quote_id]
    ref_columns = [table.openpayments_quotes.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
table "openpayments_quotes" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "sender_wallet_address" {
    null = true
    type = text
  }
  column "receiver_wallet_address" {
    null = true
    type = text
  }
  column "incoming_payment_id" {
    null = false
    type = uuid
  }
  column "send_amount" {
    null = false
    type = bigint
  }
  column "send_asset" {
    null = false
    type = text
  }
  column "send_scale" {
    null = false
    type = bigint
  }
  column "recv_amount" {
    null = false
    type = bigint
  }
  column "recv_asset" {
    null = false
    type = text
  }
  column "recv_scale" {
    null = false
    type = bigint
  }
  column "expires_at" {
    null = false
    type = timestamp
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "send_linked_acc_id" {
    null = true
    type = uuid
  }
  column "created_by" {
    null = true
    type = text
  }
  column "recv_identity_type" {
    null = true
    type = text
  }
  column "recv_identity" {
    null = true
    type = text
  }
  column "otp_required" {
    null = true
    type = boolean
  }
  column "otp_validated" {
    null = true
    type = boolean
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_incoming_payment_id_ref_openpayments_incoming_payment" {
    columns     = [column.incoming_payment_id]
    ref_columns = [table.openpayments_incoming_payment.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_send_linked_acc_id_ref_linked_accounts" {
    columns     = [column.send_linked_acc_id]
    ref_columns = [table.linked_accounts.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
table "schema_lock" {
  schema = schema.public
  column "lock_id" {
    null = false
    type = bigint
  }
  primary_key {
    columns = [column.lock_id]
  }
}
table "schema_migrations" {
  schema = schema.public
  column "version" {
    null = false
    type = bigint
  }
  column "dirty" {
    null = false
    type = boolean
  }
  primary_key {
    columns = [column.version]
  }
}
table "signups" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "first_name" {
    null = true
    type = text
  }
  column "last_name" {
    null = true
    type = text
  }
  column "country_code" {
    null = true
    type = character_varying(2)
  }
  column "email" {
    null = true
    type = text
  }
  column "mobile_number" {
    null = true
    type = text
  }
  column "user_id" {
    null = true
    type = uuid
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
}
table "user_wallets" {
  schema = schema.public
  column "user_id" {
    null = false
    type = uuid
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  primary_key {
    columns = [column.user_id, column.wallet_id]
  }
  foreign_key "fk_user_wallets_wallets" {
    columns     = [column.wallet_id]
    ref_columns = [table.wallets.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "user_wallets_wallet_id_ind" {
    unique  = true
    columns = [column.wallet_id]
  }
}
table "waitlist_signups" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "email" {
    null = false
    type = text
  }
  column "country_code" {
    null = false
    type = character(2)
  }
  column "full_name" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "notified_at" {
    null = true
    type = timestamp
  }
  column "notified_count" {
    null = true
    type = bigint
  }
  column "beta_opt_in" {
    null    = false
    type    = boolean
    default = false
  }
  column "mug_id" {
    null = true
    type = text
  }
  column "can_signup" {
    null    = false
    type    = boolean
    default = false
  }
  column "user_id" {
    null = true
    type = uuid
  }
  primary_key {
    columns = [column.id]
  }
  index "email_country_code" {
    unique  = true
    columns = [column.email, column.country_code]
  }
  index "waitlist_signups_mug_id" {
    unique  = true
    columns = [column.mug_id]
  }
}
table "wallets" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "name" {
    null = false
    type = text
  }
  column "country" {
    null = false
    type = text
    default = ""
  }
  column "exceeded_limits" {
    null = false
    type = boolean
    default = false
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "wallets_name" {
    unique  = false
    columns = [column.name]
  }
}
table "wallet_addresses" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "url" {
    null = false
    type = text
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "wallet_addresses_wallet_id_ind" {
    columns = [column.wallet_id]
  }
  index "wallet_address_url_ind" {
    columns = [column.url]
  }
}
table "transactions" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "foreign_id" {
    null = true
    type = text
  }
  column "grant_id" {
    null = true
    type = uuid
  }
  column "type" {
    null = false
    type = text
  }
  column "state" {
    null = false
    type = text
  }
  column "provider" {
    null = false
    type = text
  }
  column "source" {
    null = true
    type = text
  }
  column "destination" {
    null = true
    type = text
  }
  column "title" {
    null = true
    type = text
  }
  column "note" {
    null = true
    type = text
  }
  column "amount" {
    null = false
    type = bigint
  }
  column "asset_code" {
    null = false
    type = text
  }
  column "asset_scale" {
    null = false
    type = bigint
  }
  column "linked_account_title" {
    null = true
    type = text
  }
  column "destination_identity_type" {
    null = true
    type = text
  }
  column "destination_identity" {
    null = true
    type = text
  }
  column "reference" {
    null = true
    type = text
  }
  column "refund_state" {
    null = false
    type = int
    default = 0
  }
  column "payment_protection_fee_percentage" {
    null = false
    type = float
    default = 0
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "grant_id_limits_ind" {
    columns = [column.grant_id, column.wallet_id, column.created_at, column.state]
  }
  index "grant_id_limits_overall_ind" {
    columns = [column.grant_id, column.wallet_id, column.state]
  }
  index "transactions_wallet_id_ind" {
    columns = [column.wallet_id]
  }
  index "wallet_id_state_ind" {
    columns = [column.wallet_id, column.state]
  }
  index "wallet_id_state_updated_ind" {
    columns = [column.wallet_id, column.state, column.updated_at]
  }
  index "wallet_id_state_created_ind" {
    columns = [column.wallet_id, column.state, column.created_at]
  }
  index "wallet_id_foreign_id_ind" {
    columns = [column.wallet_id, column.foreign_id]
  }
}
table "transfers" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "transaction_id" {
    null = false
    type = uuid
  }
  column "linked_acc_id" {
    null = true
    type = uuid
  }
  column "foreign_id" {
    null = true
    type = text
  }
  column "type" {
    null = false
    type = text
  }
  column "state" {
    null = false
    type = text
  }
  column "amount" {
    null = false
    type = bigint
  }
  column "asset_code" {
    null = false
    type = text
  }
  column "asset_scale" {
    null = false
    type = bigint
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_transfers_wallets" {
    columns     = [column.transaction_id]
    ref_columns = [table.transactions.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_transfers_linked_accounts" {
    columns     = [column.linked_acc_id]
    ref_columns = [table.linked_accounts.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
table "authorisation_clients" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "url" {
    null    = false
    type    = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "url_uniq" {
    unique  = true
    columns = [column.url]
  }
}
table "authorisation_grants" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "client_id" {
    null = false
    type = uuid
  }
  column "state" {
    null = false
    type = text
  }
  column "continue_token" {
    null = false
    type = text
  }
  column "wait" {
    null = true
    type = bigint
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "continue_token_uniq" {
    unique  = true
    columns = [column.continue_token]
  }
  foreign_key "fk_grants_clients" {
    columns     = [column.client_id]
    ref_columns = [table.authorisation_clients.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "client_id_ind" {
    columns = [column.client_id]
  }
}
table "authorisation_tokens" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "grant_id" {
    null = false
    type = uuid
  }
  column "state" {
    null = false
    type = text
  }
  column "label" {
    null = false
    type = text
  }
  column "token" {
    null = false
    type = text
  }
  column "expires_in" {
    null = false
    type = bigint
  }
  column "revoked_at" {
    null    = true
    type    = timestamp
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "token_uniq" {
    unique  = true
    columns = [column.token]
  }
  foreign_key "fk_token_grant" {
    columns     = [column.grant_id]
    ref_columns = [table.authorisation_grants.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
table "authorisation_token_access" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "token_id" {
    null = false
    type = uuid
  }
  column "type" {
    null = false
    type = text
  }
  column "actions" {
    null = true
    type = sql("text[]")
  }
  column "identifier" {
    null = false
    type = text
  }
  column "locations" {
    null = true
    type = sql("text[]")
  }
  column "data_types" {
    null = true
    type = sql("text[]")
  }
  column "interval" {
    null = true
    type = text
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_access_token" {
    columns     = [column.token_id]
    ref_columns = [table.authorisation_tokens.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}

table "contacts" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "payment_pointer" {
    null = false
    type = text
  }
  column "name" {
    null = false
    type = text
  }
  column "last_paid_at" {
    null = true
    type = timestamp
  }
  primary_key {
    columns = [column.id]
  }
  index "wallet_payment_pointer_unique" {
    unique  = true
    columns = [column.wallet_id, column.payment_pointer]
  }
}
table "identities" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "platform" {
    null = false
    type = text
  }
  column "state" {
    null = false
    type = text
  }
  column "public" {
    null = false
    type = boolean
  }
  column "key_id" {
    null = false
    type = text
  }
  column "identifier" {
    null = false
    type = text
  }
  column "proof" {
    null = false
    type = text
  }
  column "signature" {
    null = false
    type = bytea
  }
  column "signature_hash" {
    null = false
    type = bytea
  }
  column "verified_at" {
    null    = true
    type    = timestamp
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "platform_identifier_state_ind" {
    unique  = false
    columns = [column.platform, column.identifier, column.state]
  }
  index "identities_wallet_id_ind" {
    unique  = false
    columns = [column.wallet_id]
  }
  index "public_state" {
    unique=false
    columns = [column.public, column.state]
  }
}

table "authorisation_limits" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "foreign_id" {
    null = false
    type = uuid
  }
  column "type" {
    null = false
    type = text
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "currency" {
    null = false
    type = text
  }
  column "daily" {
    null = false
    type = bigint
  }
  column "monthly" {
    null = false
    type = bigint
  }
  column "overall" {
    null = false
    type = bigint
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "auth_limits_wallet_id_ind" {
    unique  = false
    columns = [column.wallet_id]
  }
  index "wallet_id_foreign_id_uniq" {
    unique  = true
    columns = [column.wallet_id, column.foreign_id, column.currency]
  }
}

table "wallet_kyc_status" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "status" {
    null = false
    type = int
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "wallet_id" {
    unique  = true
    columns = [column.wallet_id]
  }
}

table "gmt_users" {
  schema = schema.public
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "sender_id" {
    null = true
    type = bigint
  }
  column "receiver_id" {
    null = true
    type = bigint
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.wallet_id]
  }
  index "gmt_users_wallet_id_ind" {
    columns = [column.wallet_id]
    unique  = true
  }
}
table "gmt_receipts" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "external_id" {
    null = true
    type = text
  }
  column "receipt" {
    null = true
    type = text
  }
  column "licence" {
    null = true
    type = text
  }
  column "right_to_return" {
    null = true
    type = text
  }
  column "error_msg" {
    null = true
    type = text
  }
  column "contact" {
    null = true
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
}
table "gmt_workflow_refs" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "external_id" {
    null = false
    type = text
  }
  column "workflow_id" {
    null = false
    type = text
  }
  column "workflow_run_id" {
    null = false
    type = text
  }
  column "activity_name" {
    null = false
    type = text
  }
  column "completed" {
    null = false
    type = boolean
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "ref_completed" {
    unique  = false
    columns = [column.completed]
  }
}
table "wallet_keys" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "key_type" {
    null = false
    type = text
  }
  column "location" {
    null = false
    type = text
  }
  column "reference" {
    null = true
    type = text
  }
  column "name" {
    null = true
    type = text
  }
  column "public_key" {
    null = false
    type = text
    default = ""
  }
  column "key_id" {
    null = false
    type = text
    default = ""
  }
  index "wallet_keys_wallet_id_ind" {
    unique  = false
    columns = [column.wallet_id]
  }
  index "wallet_key_type_reference_ind" {
    unique  = true
    columns = [column.wallet_id, column.key_type, column.reference]
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "deleted_at" {
    null    = true
    type    = timestamp
  }
  primary_key {
    columns = [column.id]
  }
}
table "kyc_persona_inquiries" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "external_id" {
    null = false
    type = text
  }
  column "state" {
    null = false
    type = text
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = true
    type    = timestamp
  }
  primary_key {
    columns = [column.id]
  }
  index "kyc_person_inquiries_external_id_ind" {
    unique  = true
    columns = [column.external_id]
  }
  index "kyc_persona_wallet_id_ind" {
    unique  = false
    columns = [column.wallet_id]
  }
}
table "kyc_persona_accounts" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "external_id" {
    null = false
    type = text
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "kyc_persona_accounts_wallet_id_ind" {
    unique  = true
    columns = [column.wallet_id]
  }
}
table "admin_audit_log" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "admin_user" {
    null = false
    type = text
  }
  column "wallet_id" {
    null = true
    type = uuid
  }
  column "operation" {
    null = false
    type = text
  }
  column "parameters" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "audit_wallet_id_ind" {
    columns = [column.wallet_id]
  }
}
table "tabapay_3ds_sessions" {
  schema = schema.public
  column "id" {
    null = false
    type = text
  }
  column "order_id" {
    null = false
    type = uuid
  }
  column "card_id" {
    null = false
    type = text
  }
  column "amount" {
    null = false
    type = text
  }
  column "currency" {
    null = false
    type = text
  }
  column "version" {
    null = false
    type = text
  }
  column "revision" {
    null = false
    type = int
  }
  column "enrolled" {
    null = false
    type = text
  }
  column "processor_transaction_id" {
    null = false
    type = text
  }
  column "ds_transaction_id" {
    null = false
    type = text
  }
  column "status" {
    null = false
    type = text
  }
  column "eci" {
    null = false
    type = text
  }
  column "ucaf" {
    null = false
    type = text
  }
  column "xid" {
    null = false
    type = text
  }
  column "challenge_url" {
    null = false
    type = text
  }
  column "payload" {
    null = false
    type = text
  }
  column "init_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "lookup_at" {
    null    = true
    type    = timestamp
  }
  column "authenticated_at" {
    null    = true
    type    = timestamp
  }
  primary_key {
    columns = [column.id, column.revision]
  }
}

table "tabapay_fx_rates" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "currency_code" {
    null = false
    type = text
  }
  column "network" {
    null = false
    type = text
  }
  column "buy_rate"{
    null = false
    type = float
  }
  column "buy_rate_inverted"{
    null = false
    type = float
  }
  column "sell_rate"{
    null = false
    type = float
  }
  column "sell_rate_inverted"{
    null = false
    type = float
  }
  column "file_name" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "tabapay_fx_rates_file_currency_uniq_ind" {
    unique = true
    columns = [column.currency_code, column.network, column.file_name]
  }
  index "tabapay_fx_rates_latest_currency_ind" {
    columns = [column.currency_code, column.network, column.created_at]
  }
}

table "wallet_features" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "send_enabled" {
    null = false
    type = boolean
    default = false
  }
  column "receive_enabled" {
    null = false
    type = boolean
    default = false
  }
  column "linked_accounts_enabled" {
    null = false
    type = boolean
    default = false
  }
  column "interac_enabled" {
    null = false
    type = boolean
    default = false
  }
  column "cards_enabled" {
    null = false
    type = boolean
    default = false
  }
  column "banks_enabled" {
    null = false
    type = boolean
    default = false
  }
  column "identities_enabled" {
    null = false
    type = boolean
    default = false
  }
  column "twitter_enabled" {
    null = false
    type = boolean
    default = false
  }
  column "add_cards_enabled" {
    null = false
    type = boolean
    default = false
  }
  column "manage_wallet_cards_enabled" {
    null = false
    type = boolean
    default = false
  }
  column "account_enabled" {
    null = false
    type = boolean
    default = false
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "wallet_features_wallet_id_ind" {
    unique  = true
    columns = [column.wallet_id]
  }
}

table "external_api_logs" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "provider" {
    null = false
    type = text
  }
  column "context" {
    null = false
    type = text
    default = ""
  }
  column "method" {
    null = false
    type = text
    default = ""
  }
  column "request_body" {
    null = false
    type = text
    default= ""
  }
  column "request_path" {
    null = false
    type = text
  }
  column "response_body" {
    null = false
    type = text
    default= ""
  }
  column "response_status" {
    null = false
    type = text
    default= ""
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
}

table "twitter_authorizations" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "client_id" {
    null = false
    type = text
  }
  column "state" {
    null = false
    type = text
  }
  column "scopes" {
    null = false
    type = sql("text[]")
  }
  column "redirect_url" {
    null = false
    type = text
  }
  column "code_verifier" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "twitter_auth_state_ind" {
    unique  = true
    columns = [column.state]
  }
}

table "twitter_connections" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "user_id" {
    null = false
    type = text
  }
  column "scopes" {
    null = false
    type = sql("text[]")
  }
  column "username" {
    null = false
    type = text
  }
  column "client_id" {
    null = false
    type = text
  }
  column "access_token" {
    null = false
    type = text
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "refresh_token" {
    null = false
    type = text
  }
  column "expiry" {
    null = false
    type = timestamp
  }
  column "token_type" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "twitter_wallet_user_ind" {
    unique  = true
    columns = [column.wallet_id, column.user_id]
  }
}

table "payments" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "public_id" {
    null    = false
    type    = text
  }
  column "state" {
    null    = false
    type    = int
  }
  column "sender_id" {
    null = false
    type = text
  }
  column "sender_id_type" {
    null = false
    type = text
  }
  column "sender_amount" {
    null = false
    type = bigint
  }
  column "sender_currency" {
    null = false
    type = text
  }
  column "sender_account" {
    null = true
    type = text
  }
  column "send_transaction_id" {
    null = true
    type = text
  }
  column "receiver_id" {
    null = false
    type = text
  }
  column "receiver_id_type" {
    null = false
    type = text
  }
  column "fx_rate" {
    null = true
    type = float
  }
  column "fx_fee_percentage" {
    null = true
    type = float
  }
  column "receiver_amount" {
    null = false
    type = bigint
  }
  column "receiver_currency" {
    null = false
    type = text
  }
  column "receiver_account" {
    null = true
    type = text
  }
  column "receive_transaction_id" {
    null = true
    type = text
  }
  column "action_three_ds_required" {
    null = false
    type = boolean
    default = false
  }
  column "action_three_ds_id" {
    null = true
    type = text
  }
  column "note" {
    null = true
    type = text
  }
  column "action_otp_required" {
    null = false
    type = boolean
    default = false
  }
  column "action_otp" {
    null = true
    type = text
  }
  column "ip_address" {
    null = true
    type = text
  }
  column "protection_fee_percentage" {
    null = false
    type = float
    default = 0
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "type" {
    null    = false
    type    = int
    default = 1
  }
  column "revision" {
    null    = false
    type    = int
    default = 1
  }
  column "astra_correlation_id" {
    null    = true
    type    = text
  }
  primary_key {
    columns = [column.id]
  }
  index "payments_public_id_ind" {
    columns = [column.public_id]
    unique  = true
  }
  index "payments_astra_correlation_id_ind" {
    columns = [column.astra_correlation_id]
    unique  = true
  }
}

table "payments_workflow_refs" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "payment_id" {
    null = false
    type = text
  }
  column "identifier" {
    null = false
    type = text
  }
  column "wallet_id" {
    null = false
    type = text
  }
  column "workflow_id" {
    null = false
    type = text
  }
  column "workflow_run_id" {
    null = false
    type = text
  }
  column "completed" {
    null = false
    type = boolean
    default = false
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "payments_workflow_refs_completed_ind" {
    unique  = false
    columns = [column.completed]
  }
  index "payments_workflow_refs_identifier_ind" {
    unique  = false
    columns = [column.identifier]
  }
  index "payments_workflow_refs_payment_id_ind" {
    unique  = false
    columns = [column.payment_id]
  }
  index "payments_workflow_refs_wallet_id_ind" {
    unique  = false
    columns = [column.wallet_id]
  }
}

table "discord_authorizations" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "client_id" {
    null = false
    type = text
  }
  column "state" {
    null = false
    type = text
  }
  column "scopes" {
    null = false
    type = sql("text[]")
  }
  column "redirect_url" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "discord_auth_state_ind" {
    unique  = true
    columns = [column.state]
  }
}

table "discord_connections" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "user_id" {
    null = false
    type = text
  }
  column "scopes" {
    null = false
    type = sql("text[]")
  }
  column "username" {
    null = false
    type = text
  }
  column "global_name" {
    null = false
    type = text
  }
  column "client_id" {
    null = false
    type = text
  }
  column "access_token" {
    null = false
    type = text
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "refresh_token" {
    null = false
    type = text
  }
  column "expiry" {
    null = false
    type = timestamp
  }
  column "token_type" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "discord_wallet_user_ind" {
    unique  = true
    columns = [column.wallet_id, column.user_id]
  }
}

table "dynamic_forms" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = true
    type = uuid
  }
  column "form_id" {
    null = false
    type = text
  }
  column "data" {
    null = false
    type = jsonb
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
}

table "slack_bot_installs" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "access_token" {
    null = false
    type = text
  }
  column "scopes" {
    null = false
    type = sql("text[]")
  }
  column "team_id" {
    null = false
    type = text
  }
  column "bot_id" {
    null = false
    type = text
  }
  column "app_id" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  index "slack_bot_installs_team_ind" {
    unique  = true
    columns = [column.team_id]
  }
  primary_key {
    columns = [column.id]
  }
}

table "slack_unsignedup_payments" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "payment_id" {
    null = false
    type = uuid
  }
  column "team_id" {
    null = false
    type = text
  }
  column "user_id" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  index "slack_unsignedup_payments_user_ind" {
    columns = [column.team_id, column.user_id]
  }
  primary_key {
    columns = [column.id]
  }
}

table "slack_authorizations" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "client_id" {
    null = false
    type = text
  }
  column "nonce" {
    null = false
    type = text
  }
  column "state" {
    null = false
    type = text
  }
  column "scopes" {
    null = false
    type = sql("text[]")
  }
  column "redirect_url" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "slack_auth_state_ind" {
    unique  = true
    columns = [column.state]
  }
}

table "slack_connections" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "user_id" {
    null = false
    type = text
  }
  column "username" {
    null = false
    type = text
  }
  column "team_name" {
    null = false
    type = text
  }
  column "team_domain" {
    null = false
    type = text
  }
  column "team_id" {
    null = false
    type = text
  }
  column "scopes" {
    null = false
    type = sql("text[]")
  }
  column "client_id" {
    null = false
    type = text
  }
  column "access_token" {
    null = false
    type = text
  }
  column "refresh_token" {
    null = false
    type = text
  }
  column "expiry" {
    null = false
    type = timestamp
  }
  column "token_type" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "slack_wallet_user_team_ind" {
    unique  = true
    columns = [column.wallet_id, column.user_id, column.team_domain]
  }
  index "slack_user_team_ind" {
    columns = [column.user_id, column.team_domain]
  }
}

table "discord_payment_interactions" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "payment_id" {
    null = false
    type = text
  }
  column "notified_processing" {
    type    = bool
    default = false
  }
  column "notified_receiver" {
    type    = bool
    default = false
  }
  column "interaction" {
    type = text
    default = ""
  }
  column "receiver_discord_user_id" {
    type = text
    default = ""
  }
  column "sender_discord_username" {
    type = text
    default = ""
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "expired_at" {
    null    = true
    type    = timestamp
  }
  primary_key {
    columns = [column.id]
  }
}

table "rafiki_payment_pointers" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "payment_pointer_id" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "rafiki_payment_pointers_id_uniq_idx" {
    unique  = true
    columns = [column.payment_pointer_id]
  }
  index "rafiki_payment_pointers_wallet_idx" {
    unique  = true
    columns = [column.wallet_id]
  }
}

table "rafiki_wallet_keys" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "external_id" {
    type = text
    null = false
  }
  column "internal_id" {
    null = false
    type = uuid
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
}

table "rafiki_outgoing_payments" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "completed" {
    null    = false
    type    = boolean
    default = false
  }
  column "payment_id" {
    null = true
    type = uuid
  }
  column "event_id" {
    null = true
    type = uuid
  }
  column "from_wallet" {
    null = false
    type = uuid
  }
  column "to_wallet" {
    null = false
    type = uuid
  }
  column "amount" {
    null = false
    type = bigint
  }
  column "amount_asset" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "rafiki_outgoing_payments_from_wallet_idx" {
    columns = [column.from_wallet]
  }
  index "rafiki_outgoing_payments_to_wallet_idx" {
    columns = [column.to_wallet]
  }
}

table "rafiki_incoming_payments" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "payment_pointer_id" {
    null = false
    type = text
  }
  column "completed" {
    null    = false
    type    = boolean
    default = false
  }
  column "received_amount" {
    null = false
    type = bigint
  }
  column "received_amount_asset" {
    null = false
    type = text
  }
  column "payment_id" {
    null = true
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "rafiki_incoming_payments_payment_pointers_id_idx" {
    columns = [column.payment_pointer_id]
  }
}

table "xago_sub_accounts" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "account_id" {
    null    = false
    type    = uuid
  }
  column "deposit_address" {
    null = false
    type = text
  }
  column "deposit_tag" {
    null = false
    type = text
  }
  column "deposit_reference" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "xago_sub_accounts_wallet_id_idx" {
    unique  = true
    columns = [column.wallet_id]
  }
}

table "xago_deposit_details" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "sub_account_id" {
    null    = false
    type    = uuid
  }
  column "currency" {
    null    = false
    type    = text
  }
  column "bank_name" {
    null = false
    type = text
  }
  column "account_name" {
    null = false
    type = text
  }
  column "account_number" {
    null = false
    type = text
  }
  column "branch_code" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "xago_deposit_details_wallet_id_idx" {
    columns = [column.wallet_id]
  }
  index "xago_deposit_details_uniq_idx" {
    unique = true
    columns = [column.sub_account_id, column.currency, column.bank_name]
  }
}

table "xago_beneficiaries" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "address" {
    null    = false
    type    = text
  }
  column "reference" {
    null    = false
    type    = text
  }
  column "status" {
    null    = false
    type    = text
  }
  column "currency" {
    null    = false
    type    = text
  }
  column "scope" {
    null    = false
    type    = text
  }
  column "name" {
    null    = false
    type    = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "xago_beneficiaries_wallet_id_idx" {
    columns = [column.wallet_id]
  }
}

table "xago_transactions" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "linked_account_id" {
    null    = false
    type    = text
  }
  column "transaction_id" {
    null    = false
    type    = text
  }
  column "amount" {
    null = false
    type = bigint
  }
  column "currency" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "xago_transactions_wallet_id_idx" {
    columns = [column.wallet_id]
  }
}

table "xago_deposits" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "transaction_id" {
    null = false
    type = uuid
  }
  column "origin_amount" {
    null = false
    type = float
  }
  column "amount" {
    null = false
    type = float
  }
  column "status" {
    null = false
    type = text
  }
  column "account_id" {
    null = false
    type = uuid
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "xago_deposits_tx_id_idx" {
    unique = true
    columns = [column.transaction_id]
  }
}

table "xago_access_token" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "token" {
    null = false
    type = text
  }
  column "expires_at" {
    null    = false
    type    = timestamp
  }
  primary_key {
    columns = [column.id]
  }
}

table "pti_users" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "external_id" {
    null = false
    type = text
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "status" {
    null = false
    type = text
    default = ""
  }
  column "assessment_status" {
    null = false
    type = text
    default = ""
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "pti_users_wallet_id_idx" {
    unique = true
    columns = [column.wallet_id]
  }
}
table "astra_user_intents" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "intent_id" {
    null = false
    type = text
  }
  column "status" {
    null = false
    type = text
  }
  column "user_id" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "astra_user_intents_wallet_id_idx" {
    unique = true
    columns = [column.wallet_id]
  }
}

table "astra_access_tokens" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "token" {
    null = false
    type = text
  }
  column "expires_at" {
    null    = false
    type    = timestamp
  }
  column "refresh_token" {
    null = false
    type = text
  }
  column "refresh_expires_at" {
    null    = false
    type    = timestamp
  }

  primary_key {
    columns = [column.id]
  }
  index "astra_access_tokens_wallet_id_idx" {
    unique = true
    columns = [column.wallet_id]
  }
  index "astra_access_tokens_refresh_idx" {
    columns = [column.refresh_expires_at]
  }
}

table "astra_accounts" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "account_id" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }

  primary_key {
    columns = [column.id]
  }
  index "astra_accounts_wallet_id_idx" {
    unique = true
    columns = [column.wallet_id]
  }
}

table "pti_transactions" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "payment_id" {
    null = false
    type = uuid
  }
  column "external_request_id" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
}

table "basis_theory_linked_accounts" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "basis_theory_card_id" {
    null = false
    type = uuid
  }
  column "linked_account_id" {
    null = false
    type = uuid
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }

  index "basis_theory_linked_account_id_idx" {
    unique = true
    columns = [column.basis_theory_card_id, column.linked_account_id]
  }
  primary_key {
    columns = [column.id]
  }
}

table "gatehub_users" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "external_id" {
    null = false
    type = text
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "external_customer_id" {
    null = true
    type = text
  }
  column "external_account_id" {
    null = true
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "gatehub_users_wallet_id_idx" {
    unique = true
    columns = [column.wallet_id, column.external_id]
  }
}

table "gatehub_transactions" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "external_id" {
    null = false
    type = text
  }
  column "payment_id" {
    null = false
    type = uuid
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "gatehub_transactions_external_id_payment_id" {
    unique = true
    columns = [column.payment_id, column.external_id]
  }
}

table "chi_money_wallets" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "external_id" {
    null = false
    type = text
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "chi_money_wallets_wallet_id" {
    unique = true
    columns = [column.wallet_id]
  }
}

table "chi_money_interac_emails" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "email" {
    null = false
    type = text
  }
  column "wallet_id" {
    null = false
    type = uuid
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
  primary_key {
    columns = [column.id]
  }
  index "chi_money_interac_emails_wallet_id" {
    unique = true
    columns = [column.wallet_id]
  }
}

table "atlas_schema_history" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "hash" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
}
schema "public" {
}
