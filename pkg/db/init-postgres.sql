-- PostgreSQL schema (핵심만 발췌)

CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  wallet_address BYTEA UNIQUE NOT NULL, -- lowercased hex -> bytea
  line_uid TEXT,
  created_at TIMESTAMPTZ DEFAULT now(),
  status SMALLINT DEFAULT 1
);

CREATE TABLE merchants (
  id BIGSERIAL PRIMARY KEY,
  wallet_address BYTEA UNIQUE NOT NULL,
  name TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE campaigns (
  id BIGSERIAL PRIMARY KEY,
  address BYTEA UNIQUE NOT NULL,
  merchant_id BIGINT REFERENCES merchants(id),
  base_price NUMERIC(20,6) NOT NULL,
  min_qty BIGINT NOT NULL,
  lock_start TIMESTAMPTZ NOT NULL,
  lock_end TIMESTAMPTZ NOT NULL,
  rmax_bps INTEGER NOT NULL,
  savefloor_bps INTEGER NOT NULL,
  merchant_fee_bps INTEGER NOT NULL,
  ops_fee_bps INTEGER NOT NULL,
  state SMALLINT NOT NULL,
  metadata_uri TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE participants (
  id BIGSERIAL PRIMARY KEY,
  campaign_id BIGINT REFERENCES campaigns(id),
  user_id BIGINT REFERENCES users(id),
  deposit NUMERIC(20,6) NOT NULL,
  joined_at TIMESTAMPTZ NOT NULL,
  status SMALLINT NOT NULL,
  UNIQUE (campaign_id, user_id)
);

CREATE TABLE sponsor_allocations (
  id BIGSERIAL PRIMARY KEY,
  campaign_id BIGINT REFERENCES campaigns(id),
  amount NUMERIC(20,6) NOT NULL,
  sponsor_wallet BYTEA NOT NULL,
  allocated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE settlements (
  id BIGSERIAL PRIMARY KEY,
  campaign_id BIGINT REFERENCES campaigns(id) UNIQUE,
  snapshot_time TIMESTAMPTZ NOT NULL,
  total_amount NUMERIC(20,6) NOT NULL,
  rebate_paid NUMERIC(20,6) NOT NULL,
  merchant_payout NUMERIC(20,6) NOT NULL,
  ops_fee NUMERIC(20,6) NOT NULL,
  sponsor_consumed NUMERIC(20,6) NOT NULL,
  state SMALLINT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE rebates (
  id BIGSERIAL PRIMARY KEY,
  settlement_id BIGINT REFERENCES settlements(id),
  user_id BIGINT REFERENCES users(id),
  amount NUMERIC(20,6) NOT NULL,
  sponsor_part NUMERIC(20,6) NOT NULL,
  yield_part NUMERIC(20,6) NOT NULL
);

CREATE INDEX idx_campaign_state ON campaigns(state, lock_end);
CREATE INDEX idx_participants_user ON participants(user_id, campaign_id);
