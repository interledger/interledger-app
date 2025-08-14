table "ledger_accounts" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "ledger_id" {
    null = false
    type = bigint
  }
  column "code" {
    null = false
    type = bigint
  }
  column "debits_must_not_exceed_credits" {
    null = false
    type = bool
    default = sql("false::BOOL")
  }
  column "credits_must_not_exceed_debits" {
    null = false
    type = bool
    default = sql("false::BOOL")
  }
  column "debits_pending" {
    null    = false
    type    = bigint
    default = sql("0::INT8")
  }
  column "debits_posted" {
    null    = false
    type    = bigint
    default = sql("0::INT8")
  }
  column "credits_pending" {
    null    = false
    type    = bigint
    default = sql("0::INT8")
  }
  column "credits_posted" {
    null    = false
    type    = bigint
    default = sql("0::INT8")
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
  foreign_key "fk_accounts_ledgers" {
    columns     = [column.ledger_id]
    ref_columns = [table.ledgers.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  check "ledger_accounts_credits_pending_non_negative" {
    expr = "credits_pending >= 0"
  }
  check "ledger_accounts_credits_posted_non_negative" {
    expr = "credits_posted >= 0"
  }
  check "ledger_accounts_debits_pending_non_negative" {
    expr = "debits_pending >= 0"
  }
  check "ledger_accounts_debits_posted_non_negative" {
    expr = "debits_posted >= 0"
  }

}
table "ledger_transfers" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "ledger_id" {
    null = false
    type = bigint
  }
  column "code" {
    null = false
    type = bigint
  }
  column "debit_account_id" {
    null = false
    type = uuid
  }
  column "credit_account_id" {
    null = false
    type = uuid
  }
  column "amount" {
    null = false
    type = bigint
  }
  column "provider_fee" {
    null = false
    type = bigint
    default = 0
  }
  column "state" {
    null = false
    type = smallint
  }
  column "timeout_at" {
    null = true
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
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_transfers_accounts_credits" {
    columns     = [column.credit_account_id]
    ref_columns = [table.ledger_accounts.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_transfers_accounts_debits" {
    columns     = [column.debit_account_id]
    ref_columns = [table.ledger_accounts.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
table "ledgers" {
  schema = schema.public
  column "id" {
    null = false
    type = bigint
  }
  column "name" {
    null = true
    type = text
  }
  column "asset" {
    null = false
    type = text
  }
  column "scale" {
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
}
schema "public" {
}
