CREATE TABLE IF NOT EXISTS signed_agreements
(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	agreement_ids STRING[] NOT NULL,
  identity_id         TEXT NOT NULL,
	ip_address TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT now(),
	updated_at TIMESTAMP NOT NULL DEFAULT now()
);