CREATE TABLE IF NOT EXISTS agreement_signatures
(
	id UUID       PRIMARY KEY DEFAULT gen_random_uuid(),
	agreement_id  TEXT NOT NULL,
  user_id       TEXT NOT NULL,
	ip_address    TEXT NOT NULL,
	created_at    TIMESTAMP NOT NULL DEFAULT now(),
	updated_at    TIMESTAMP NOT NULL DEFAULT now(),
	INDEX user_id_ind(user_id),
	CONSTRAINT fk_agreement
		FOREIGN KEY(agreement_id)
			REFERENCES agreements(id)
);