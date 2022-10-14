ALTER TABLE linked_accounts ADD CONSTRAINT  wallet_id_provider_provider_id_uniq UNIQUE (wallet_id, provider, provider_id);
