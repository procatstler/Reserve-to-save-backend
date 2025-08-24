import { pool } from '@config/database';
import { User, Session } from '@/types';
import { v4 as uuidv4 } from 'uuid';

export class UserRepository {
  async findById(id: string): Promise<User | null> {
    const query = 'SELECT * FROM users WHERE id = $1';
    const result = await pool.query(query, [id]);
    return result.rows[0] || null;
  }

  async findByWalletAddress(address: string): Promise<User | null> {
    const query = 'SELECT * FROM users WHERE LOWER(wallet_address) = LOWER($1)';
    const result = await pool.query(query, [address]);
    return result.rows[0] || null;
  }

  async findByLineUserId(lineUserId: string): Promise<User | null> {
    const query = 'SELECT * FROM users WHERE line_user_id = $1';
    const result = await pool.query(query, [lineUserId]);
    return result.rows[0] || null;
  }

  async create(data: Partial<User>): Promise<User> {
    const id = uuidv4();
    const query = `
      INSERT INTO users (
        id, wallet_address, line_user_id, line_display_name, 
        line_picture_url, email, kyc_tier, status, metadata
      ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
      RETURNING *
    `;
    const values = [
      id,
      data.wallet_address?.toLowerCase(),
      data.line_user_id,
      data.line_display_name,
      data.line_picture_url,
      data.email,
      data.kyc_tier || 0,
      data.status || 'active',
      JSON.stringify(data.metadata || {})
    ];
    const result = await pool.query(query, values);
    return result.rows[0];
  }

  async update(id: string, data: Partial<User>): Promise<User | null> {
    const fields = [];
    const values = [];
    let index = 1;

    Object.entries(data).forEach(([key, value]) => {
      if (value !== undefined && key !== 'id') {
        fields.push(`${key} = $${index}`);
        values.push(key === 'metadata' ? JSON.stringify(value) : value);
        index++;
      }
    });

    if (fields.length === 0) return null;

    values.push(id);
    const query = `
      UPDATE users 
      SET ${fields.join(', ')}
      WHERE id = $${index}
      RETURNING *
    `;
    const result = await pool.query(query, values);
    return result.rows[0] || null;
  }

  async updateLastLogin(id: string): Promise<void> {
    const query = 'UPDATE users SET last_login_at = NOW() WHERE id = $1';
    await pool.query(query, [id]);
  }

  async updateLineProfile(id: string, data: Partial<User>): Promise<void> {
    const query = `
      UPDATE users 
      SET line_display_name = $2, line_picture_url = $3, updated_at = NOW()
      WHERE id = $1
    `;
    await pool.query(query, [id, data.line_display_name, data.line_picture_url]);
  }

  async createSession(data: Session): Promise<Session> {
    const query = `
      INSERT INTO sessions (
        id, user_id, token_hash, refresh_token_hash,
        ip_address, user_agent, device_fingerprint,
        expires_at, refresh_expires_at
      ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
      RETURNING *
    `;
    const values = [
      data.id,
      data.user_id,
      data.token_hash,
      data.refresh_token_hash,
      data.ip_address,
      data.user_agent,
      data.device_fingerprint,
      data.expires_at,
      data.refresh_expires_at
    ];
    const result = await pool.query(query, values);
    return result.rows[0];
  }

  async findSessionByToken(tokenHash: string): Promise<Session | null> {
    const query = 'SELECT * FROM sessions WHERE token_hash = $1';
    const result = await pool.query(query, [tokenHash]);
    return result.rows[0] || null;
  }

  async findSessionByRefreshToken(refreshTokenHash: string): Promise<Session | null> {
    const query = 'SELECT * FROM sessions WHERE refresh_token_hash = $1';
    const result = await pool.query(query, [refreshTokenHash]);
    return result.rows[0] || null;
  }

  async updateSession(id: string, data: Partial<Session>): Promise<void> {
    const query = `
      UPDATE sessions 
      SET token_hash = $2, expires_at = $3, last_used_at = $4
      WHERE id = $1
    `;
    await pool.query(query, [id, data.token_hash, data.expires_at, data.last_used_at]);
  }

  async updateSessionLastUsed(id: string): Promise<void> {
    const query = 'UPDATE sessions SET last_used_at = NOW() WHERE id = $1';
    await pool.query(query, [id]);
  }

  async deleteSessionByToken(tokenHash: string): Promise<void> {
    const query = 'DELETE FROM sessions WHERE token_hash = $1';
    await pool.query(query, [tokenHash]);
  }

  async deleteExpiredSessions(): Promise<void> {
    const query = 'DELETE FROM sessions WHERE expires_at < NOW()';
    await pool.query(query);
  }
}

export const userRepository = new UserRepository();