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

-- 예시 데이터 INSERT
-- commnet below if you don't want example data

-- Users 예시 데이터 (5개)
INSERT INTO users (wallet_address, line_uid, status) VALUES
  (decode('a1b2c3d4e5f6789012345678901234567890abcd', 'hex'), 'line_user_001', 1),
  (decode('b2c3d4e5f6789012345678901234567890abcdef', 'hex'), 'line_user_002', 1),
  (decode('c3d4e5f6789012345678901234567890abcdef01', 'hex'), 'line_user_003', 1),
  (decode('d4e5f6789012345678901234567890abcdef0123', 'hex'), 'line_user_004', 1),
  (decode('e5f6789012345678901234567890abcdef012345', 'hex'), 'line_user_005', 1);

-- Merchants 예시 데이터 (2개)
INSERT INTO merchants (wallet_address, name) VALUES
  (decode('1234567890abcdef1234567890abcdef12345678', 'hex'), 'merchant1'),
  (decode('234567890abcdef1234567890abcdef123456789', 'hex'), 'merchant2');

-- Campaigns 예시 데이터 (2개) - merchant_id는 위에서 생성된 merchants를 참조
INSERT INTO campaigns (
  address, 
  merchant_id, 
  base_price, 
  min_qty, 
  lock_start, 
  lock_end, 
  rmax_bps, 
  savefloor_bps, 
  merchant_fee_bps, 
  ops_fee_bps, 
  state, 
  metadata_uri
) VALUES
  (
    decode('1111111111111111111111111111111111111111', 'hex'),
    1, -- merchant1의 ID
    10.500000,
    100,
    '2024-01-15 09:00:00+00',
    '2024-02-15 18:00:00+00',
    500,  -- 5% rmax
    200,  -- 2% savefloor
    300,  -- 3% merchant fee
    100,  -- 1% ops fee
    1,    -- active state
    'https://example.com/campaign1-metadata.json'
  ),
  (
    decode('2222222222222222222222222222222222222222', 'hex'),
    2, -- merchant2의 ID
    25.000000,
    50,
    '2024-01-20 10:00:00+00',
    '2024-03-20 20:00:00+00',
    600,  -- 6% rmax
    250,  -- 2.5% savefloor
    400,  -- 4% merchant fee
    150,  -- 1.5% ops fee
    1,    -- active state
    'https://example.com/campaign2-metadata.json'
  );
