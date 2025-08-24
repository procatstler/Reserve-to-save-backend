export interface User {
  id: string;
  wallet_address: string;
  line_user_id?: string;
  line_display_name?: string;
  line_picture_url?: string;
  email?: string;
  kyc_tier: number;
  status: 'active' | 'suspended' | 'deleted';
  created_at: Date;
  updated_at: Date;
  last_login_at?: Date;
  metadata: Record<string, any>;
}

export interface Campaign {
  id: string;
  chain_address: string;
  title: string;
  description?: string;
  image_url?: string;
  merchant_id?: string;
  merchant_wallet: string;
  base_price: string;
  min_qty: number;
  current_qty: number;
  target_amount: string;
  current_amount: string;
  discount_rate: number;
  save_floor_bps: number;
  r_max_bps: number;
  merchant_fee_bps: number;
  ops_fee_bps: number;
  start_time: Date;
  end_time: Date;
  settlement_date?: Date;
  status: 'draft' | 'recruiting' | 'reached' | 'fulfillment' | 'settled' | 'failed' | 'cancelled';
  tx_hash?: string;
  block_number?: number;
  created_at: Date;
  updated_at: Date;
  metadata: Record<string, any>;
}

export interface Participation {
  id: string;
  campaign_id: string;
  user_id: string;
  wallet_address: string;
  deposit_amount: string;
  joined_at: Date;
  cancel_pending: string;
  expected_rebate: string;
  actual_rebate?: string;
  status: 'active' | 'pending_cancel' | 'cancelled' | 'settled' | 'refunded';
  tx_hash?: string;
  cancel_tx_hash?: string;
  settlement_tx_hash?: string;
  refund_tx_hash?: string;
  created_at: Date;
  updated_at: Date;
  metadata: Record<string, any>;
}

export interface Payment {
  id: string;
  payment_id: string;
  campaign_id?: string;
  user_id?: string;
  participation_id?: string;
  amount: string;
  currency: 'USDT' | 'KAIA' | 'KRW' | 'USD';
  mode: 'crypto' | 'stripe';
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'refunded';
  transaction_hash?: string;
  provider_response?: Record<string, any>;
  created_at: Date;
  completed_at?: Date;
  failed_at?: Date;
  refunded_at?: Date;
  metadata: Record<string, any>;
}

export interface Session {
  id: string;
  user_id: string;
  token_hash: string;
  refresh_token_hash?: string;
  ip_address?: string;
  user_agent?: string;
  device_fingerprint?: string;
  expires_at: Date;
  refresh_expires_at?: Date;
  created_at: Date;
  last_used_at: Date;
}

export interface ChainEvent {
  id: number;
  block_number: number;
  tx_hash: string;
  log_index: number;
  contract_address: string;
  event_name: string;
  event_data: Record<string, any>;
  decoded_data?: Record<string, any>;
  chain_timestamp?: Date;
  processed: boolean;
  processed_at?: Date;
  ingested_at: Date;
}

export interface WebhookLog {
  id: string;
  event_id: string;
  event_type: string;
  payload: Record<string, any>;
  signature?: string;
  processed: boolean;
  retry_count: number;
  error_message?: string;
  received_at: Date;
  processed_at?: Date;
}

export interface JWTPayload {
  userId: string;
  address?: string;
  lineUserId?: string;
  kycTier: number;
}

export interface AuthRequest extends Express.Request {
  user?: JWTPayload;
  session?: Session;
}