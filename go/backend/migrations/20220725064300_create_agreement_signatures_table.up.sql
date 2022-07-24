CREATE TABLE IF NOT EXISTS agreement_signatures
(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	agreement_id TEXT NOT NULL,
  identity_id         TEXT NOT NULL,
	ip_address TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT now(),
	updated_at TIMESTAMP NOT NULL DEFAULT now(),
	CONSTRAINT fk_agreement
		FOREIGN KEY(agreement_id)
			REFERENCES agreements(id)
);