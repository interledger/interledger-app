CREATE TABLE IF NOT EXISTS openpayments_incoming_payment
(
  id                  UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
  payment_pointer_id  UUID REFERENCES payment_pointers,
  asset_code          TEXT NOT NULL,
  asset_scale         INT NOT NULL,
  incoming_amount     BIGINT NOT NULL,
  received_amount     BIGINT NOT NULL,
  completed           BOOLEAN DEFAULT FALSE,
  expires_at          TIMESTAMP,
  external_ref        TEXT,
  ilp_stream_id       TEXT,
  ilp_address         TEXT,
  ilp_shared_secret   TEXT,
  created_at          TIMESTAMP NOT NULL DEFAULT now(),
  updated_at          TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS openpayments_quoutes
(
  id 				              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  send_payment_pointer_id UUID NOT NULL REFERENCES payment_pointers,
  recv_payment_pointer_id UUID NOT NULL REFERENCES payment_pointers,
  incoming_payment_id     UUID NOT NULL REFERENCES openpayments_incoming_payment,
  send_amount             BIGINT NOT NULL,
  send_asset              TEXT NOT NULL,
  send_scale              INT NOT NULL,
  recv_amount             BIGINT NOT NULL,
  recv_asset              TEXT NOT NULL,
  recv_scale              INT NOT NULL,
  expires_at   	          TIMESTAMP NOT NULL,
  created_at   	          TIMESTAMP NOT NULL DEFAULT now(),
  updated_at   	          TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS openpayments_outgoing_payment
(
  id 				              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  quote_id                UUID NOT NULL REFERENCES openpayments_incoming_payment,
  failed                  BOOLEAN NOT NULL DEFAULT FALSE,
  description             TEXT NOT NULL,
  sent_amount             BIGINT NOT NULL,
  sent_asset              TEXT NOT NULL,
  sent_scale              INT NOT NULL,
  created_at   	          TIMESTAMP NOT NULL DEFAULT now(),
  updated_at   	          TIMESTAMP NOT NULL DEFAULT now()
);