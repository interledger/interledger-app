CREATE TABLE IF NOT EXISTS openpayments_quoutes (
  id 				              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  send_payment_pointer_id UUID NOT NULL REFERENCES payment_pointers,
  recv_payment_pointer_id UUID NOT NULL REFERENCES payment_pointers,
  send_amount   BIGINT NOT NULL,
  send_asset    TEXT NOT NULL,
  send_scale    INT NOT NULL,
  recv_amount     BIGINT NOT NULL,
  recv_asset      TEXT NOT NULL,
  recv_scale      INT NOT NULL,
  expires_at   	TIMESTAMP NOT NULL,
  created_at   	TIMESTAMP NOT NULL DEFAULT now(),
  updated_at   	TIMESTAMP NOT NULL DEFAULT now()
);
