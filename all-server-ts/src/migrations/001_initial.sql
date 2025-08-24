-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  wallet_address VARCHAR(42) UNIQUE NOT NULL,
  line_user_id VARCHAR(255) UNIQUE,
  line_display_name VARCHAR(255),
  line_picture_url TEXT,
  email VARCHAR(255),
  kyc_tier SMALLINT DEFAULT 0 CHECK (kyc_tier >= 0 AND kyc_tier <= 3),
  status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  last_login_at TIMESTAMPTZ,
  metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_users_wallet ON users(wallet_address);
CREATE INDEX idx_users_line_id ON users(line_user_id);
CREATE INDEX idx_users_status ON users(status);

-- Campaigns table
CREATE TABLE campaigns (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chain_address VARCHAR(42) UNIQUE NOT NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  image_url TEXT,
  merchant_id UUID REFERENCES users(id),
  merchant_wallet VARCHAR(42) NOT NULL,
  base_price NUMERIC(36, 18) NOT NULL CHECK (base_price > 0),
  min_qty INTEGER NOT NULL CHECK (min_qty > 0),
  current_qty INTEGER DEFAULT 0,
  target_amount NUMERIC(36, 18) NOT NULL,
  current_amount NUMERIC(36, 18) DEFAULT 0,
  discount_rate INTEGER NOT NULL CHECK (discount_rate >= 0 AND discount_rate <= 10000),
  save_floor_bps INTEGER NOT NULL CHECK (save_floor_bps >= 0 AND save_floor_bps <= 10000),
  r_max_bps INTEGER NOT NULL CHECK (r_max_bps >= 0 AND r_max_bps <= 10000),
  merchant_fee_bps INTEGER DEFAULT 250,
  ops_fee_bps INTEGER DEFAULT 100,
  start_time TIMESTAMPTZ NOT NULL,
  end_time TIMESTAMPTZ NOT NULL,
  settlement_date TIMESTAMPTZ,
  status VARCHAR(20) DEFAULT 'draft' CHECK (status IN (
    'draft', 'recruiting', 'reached', 'fulfillment', 'settled', 'failed', 'cancelled'
  )),
  tx_hash VARCHAR(66),
  block_number BIGINT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  metadata JSONB DEFAULT '{}'::jsonb,
  CONSTRAINT check_time_order CHECK (start_time < end_time),
  CONSTRAINT check_settlement CHECK (settlement_date IS NULL OR settlement_date > end_time),
  CONSTRAINT check_rebate_order CHECK (save_floor_bps <= r_max_bps)
);

CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaigns_merchant ON campaigns(merchant_id);
CREATE INDEX idx_campaigns_chain_address ON campaigns(chain_address);
CREATE INDEX idx_campaigns_dates ON campaigns(start_time, end_time);

-- Participations table
CREATE TABLE participations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  campaign_id UUID REFERENCES campaigns(id) ON DELETE CASCADE,
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  wallet_address VARCHAR(42) NOT NULL,
  deposit_amount NUMERIC(36, 18) NOT NULL CHECK (deposit_amount > 0),
  joined_at TIMESTAMPTZ DEFAULT NOW(),
  cancel_pending NUMERIC(36, 18) DEFAULT 0,
  expected_rebate NUMERIC(36, 18) DEFAULT 0,
  actual_rebate NUMERIC(36, 18),
  status VARCHAR(20) DEFAULT 'active' CHECK (status IN (
    'active', 'pending_cancel', 'cancelled', 'settled', 'refunded'
  )),
  tx_hash VARCHAR(66),
  cancel_tx_hash VARCHAR(66),
  settlement_tx_hash VARCHAR(66),
  refund_tx_hash VARCHAR(66),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  metadata JSONB DEFAULT '{}'::jsonb,
  UNIQUE(campaign_id, user_id)
);

CREATE INDEX idx_participations_campaign ON participations(campaign_id);
CREATE INDEX idx_participations_user ON participations(user_id);
CREATE INDEX idx_participations_status ON participations(status);
CREATE INDEX idx_participations_wallet ON participations(wallet_address);

-- Payments table
CREATE TABLE payments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  payment_id VARCHAR(255) UNIQUE NOT NULL,
  campaign_id UUID REFERENCES campaigns(id),
  user_id UUID REFERENCES users(id),
  participation_id UUID REFERENCES participations(id),
  amount NUMERIC(36, 18) NOT NULL,
  currency VARCHAR(10) NOT NULL CHECK (currency IN ('USDT', 'KAIA', 'KRW', 'USD')),
  mode VARCHAR(20) NOT NULL CHECK (mode IN ('crypto', 'stripe')),
  status VARCHAR(20) DEFAULT 'pending' CHECK (status IN (
    'pending', 'processing', 'completed', 'failed', 'refunded'
  )),
  transaction_hash VARCHAR(66),
  provider_response JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  completed_at TIMESTAMPTZ,
  failed_at TIMESTAMPTZ,
  refunded_at TIMESTAMPTZ,
  metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_payments_payment_id ON payments(payment_id);
CREATE INDEX idx_payments_user ON payments(user_id);
CREATE INDEX idx_payments_campaign ON payments(campaign_id);
CREATE INDEX idx_payments_status ON payments(status);

-- Chain events table
CREATE TABLE chain_events (
  id BIGSERIAL PRIMARY KEY,
  block_number BIGINT NOT NULL,
  tx_hash VARCHAR(66) NOT NULL,
  log_index INTEGER NOT NULL,
  contract_address VARCHAR(42) NOT NULL,
  event_name VARCHAR(100) NOT NULL,
  event_data JSONB NOT NULL,
  decoded_data JSONB,
  chain_timestamp TIMESTAMPTZ,
  processed BOOLEAN DEFAULT FALSE,
  processed_at TIMESTAMPTZ,
  ingested_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(tx_hash, log_index)
);

CREATE INDEX idx_chain_events_block ON chain_events(block_number);
CREATE INDEX idx_chain_events_contract ON chain_events(contract_address);
CREATE INDEX idx_chain_events_event ON chain_events(event_name);
CREATE INDEX idx_chain_events_processed ON chain_events(processed);

-- Receipts table
CREATE TABLE receipts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  campaign_id UUID REFERENCES campaigns(id),
  user_id UUID REFERENCES users(id),
  participation_id UUID REFERENCES participations(id),
  type VARCHAR(20) NOT NULL CHECK (type IN ('settlement', 'refund', 'cancel')),
  file_url TEXT NOT NULL,
  file_hash VARCHAR(64) NOT NULL,
  metadata JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_receipts_campaign ON receipts(campaign_id);
CREATE INDEX idx_receipts_user ON receipts(user_id);
CREATE INDEX idx_receipts_type ON receipts(type);

-- Sessions table
CREATE TABLE sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  token_hash VARCHAR(64) UNIQUE NOT NULL,
  refresh_token_hash VARCHAR(64) UNIQUE,
  ip_address INET,
  user_agent TEXT,
  device_fingerprint VARCHAR(255),
  expires_at TIMESTAMPTZ NOT NULL,
  refresh_expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  last_used_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token_hash);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

-- Webhook logs table
CREATE TABLE webhook_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id VARCHAR(255) UNIQUE NOT NULL,
  event_type VARCHAR(100) NOT NULL,
  payload JSONB NOT NULL,
  signature VARCHAR(255),
  processed BOOLEAN DEFAULT FALSE,
  retry_count INTEGER DEFAULT 0,
  error_message TEXT,
  received_at TIMESTAMPTZ DEFAULT NOW(),
  processed_at TIMESTAMPTZ
);

CREATE INDEX idx_webhook_logs_event_id ON webhook_logs(event_id);
CREATE INDEX idx_webhook_logs_processed ON webhook_logs(processed);

-- Audit logs table
CREATE TABLE audit_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID,
  action VARCHAR(100) NOT NULL,
  resource_type VARCHAR(50),
  resource_id VARCHAR(255),
  ip_address INET,
  user_agent TEXT,
  request_body JSONB,
  response_status INTEGER,
  error_message TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);

-- Trigger function for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_campaigns_updated_at BEFORE UPDATE ON campaigns
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_participations_updated_at BEFORE UPDATE ON participations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();