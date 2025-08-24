import request from 'supertest';
import express from 'express';
import { ethers } from 'ethers';
import jwt from 'jsonwebtoken';
import { authController } from './authController';
import { authRouter } from './routes';
import { errorHandler } from '@middleware/errorHandler';
import { redis } from '@config/redis';
import { userRepository } from '@repositories/userRepository';

// Mock dependencies
jest.mock('@config/redis');
jest.mock('@repositories/userRepository');
jest.mock('@utils/logger');

const app = express();
app.use(express.json());
app.use('/api/auth', authRouter);
app.use(errorHandler);

describe('AuthController', () => {
  let wallet: ethers.Wallet;
  
  beforeAll(() => {
    process.env.JWT_SECRET = 'test-secret-key-minimum-32-characters';
    process.env.JWT_REFRESH_SECRET = 'test-refresh-secret-key';
    process.env.APP_URL = 'https://test.r2s.io';
    process.env.LINE_CHANNEL_ID = 'test-channel-id';
    
    wallet = ethers.Wallet.createRandom();
  });
  
  beforeEach(() => {
    jest.clearAllMocks();
  });
  
  describe('GET /api/auth/nonce', () => {
    it('should generate nonce for valid address', async () => {
      const mockSetWithExpiry = jest.fn().mockResolvedValue(undefined);
      (redis.setEx as jest.Mock) = mockSetWithExpiry;
      
      const response = await request(app)
        .get('/api/auth/nonce')
        .query({ address: wallet.address, chainId: '1001' });
      
      expect(response.status).toBe(200);
      expect(response.body).toHaveProperty('nonce');
      expect(response.body).toHaveProperty('message');
      expect(response.body).toHaveProperty('requestId');
      expect(response.body).toHaveProperty('expiresAt');
      expect(response.body.message).toContain(wallet.address);
    });
    
    it('should reject invalid address', async () => {
      const response = await request(app)
        .get('/api/auth/nonce')
        .query({ address: 'invalid-address' });
      
      expect(response.status).toBe(400);
      expect(response.body.success).toBe(false);
    });
    
    it('should reject missing address', async () => {
      const response = await request(app)
        .get('/api/auth/nonce');
      
      expect(response.status).toBe(400);
      expect(response.body.success).toBe(false);
    });
  });
  
  describe('POST /api/auth/verify', () => {
    it('should verify valid signature and return tokens', async () => {
      const nonce = 'a1b2c3d4e5f6789012345678901234567';
      const message = `https://test.r2s.io wants you to sign in with your wallet:
${wallet.address}

URI: https://test.r2s.io
Version: 1
Chain ID: 1001
Nonce: ${nonce}
Issued At: ${new Date().toISOString()}
Expiration Time: ${new Date(Date.now() + 300000).toISOString()}
Request ID: test-request-id
Statement: Sign to authenticate with R2S platform.`;
      
      const signature = await wallet.signMessage(message);
      
      const nonceData = {
        address: wallet.address.toLowerCase(),
        chainId: '1001',
        requestId: 'test-request-id',
        expiresAt: new Date(Date.now() + 300000).toISOString()
      };
      
      (redis.get as jest.Mock).mockResolvedValue(JSON.stringify(nonceData));
      (redis.del as jest.Mock).mockResolvedValue(1);
      
      const mockUser = {
        id: 'user-123',
        wallet_address: wallet.address.toLowerCase(),
        kyc_tier: 0,
        line_user_id: null
      };
      
      (userRepository.findByWalletAddress as jest.Mock).mockResolvedValue(null);
      (userRepository.create as jest.Mock).mockResolvedValue(mockUser);
      (userRepository.createSession as jest.Mock).mockResolvedValue({
        id: 'session-123'
      });
      
      const response = await request(app)
        .post('/api/auth/verify')
        .send({
          address: wallet.address,
          signature,
          message,
          requestId: 'test-request-id'
        });
      
      expect(response.status).toBe(200);
      expect(response.body.success).toBe(true);
      expect(response.body).toHaveProperty('accessToken');
      expect(response.body).toHaveProperty('refreshToken');
      expect(response.body.user).toEqual({
        id: 'user-123',
        address: wallet.address.toLowerCase(),
        kycTier: 0,
        lineConnected: false
      });
    });
    
    it('should reject invalid signature', async () => {
      const message = 'Invalid message';
      const signature = 'invalid-signature';
      
      const response = await request(app)
        .post('/api/auth/verify')
        .send({
          address: wallet.address,
          signature,
          message,
          requestId: 'test-request-id'
        });
      
      expect(response.status).toBe(400);
      expect(response.body.success).toBe(false);
    });
    
    it('should reject expired nonce', async () => {
      const nonce = 'a1b2c3d4e5f6789012345678901234567';
      const message = `Nonce: ${nonce}`;
      
      const nonceData = {
        address: wallet.address.toLowerCase(),
        chainId: '1001',
        requestId: 'test-request-id',
        expiresAt: new Date(Date.now() - 1000).toISOString() // Expired
      };
      
      (redis.get as jest.Mock).mockResolvedValue(JSON.stringify(nonceData));
      
      const response = await request(app)
        .post('/api/auth/verify')
        .send({
          address: wallet.address,
          signature: 'any-signature',
          message,
          requestId: 'test-request-id'
        });
      
      expect(response.status).toBe(401);
      expect(response.body.error).toContain('expired');
    });
  });
  
  describe('POST /api/auth/refresh', () => {
    it('should refresh access token with valid refresh token', async () => {
      const userId = 'user-123';
      const refreshToken = jwt.sign(
        { userId },
        process.env.JWT_REFRESH_SECRET!,
        { expiresIn: '7d' }
      );
      
      const mockSession = {
        id: 'session-123',
        user_id: userId,
        refresh_token_hash: 'hash'
      };
      
      const mockUser = {
        id: userId,
        wallet_address: wallet.address.toLowerCase(),
        kyc_tier: 1
      };
      
      (userRepository.findSessionByRefreshToken as jest.Mock).mockResolvedValue(mockSession);
      (userRepository.findById as jest.Mock).mockResolvedValue(mockUser);
      (userRepository.updateSession as jest.Mock).mockResolvedValue(undefined);
      
      const response = await request(app)
        .post('/api/auth/refresh')
        .send({ refreshToken });
      
      expect(response.status).toBe(200);
      expect(response.body.success).toBe(true);
      expect(response.body).toHaveProperty('accessToken');
      
      const decoded = jwt.verify(
        response.body.accessToken,
        process.env.JWT_SECRET!
      ) as any;
      expect(decoded.userId).toBe(userId);
      expect(decoded.kycTier).toBe(1);
    });
    
    it('should reject invalid refresh token', async () => {
      const response = await request(app)
        .post('/api/auth/refresh')
        .send({ refreshToken: 'invalid-token' });
      
      expect(response.status).toBe(401);
      expect(response.body.error).toContain('Invalid refresh token');
    });
    
    it('should reject missing refresh token', async () => {
      const response = await request(app)
        .post('/api/auth/refresh')
        .send({});
      
      expect(response.status).toBe(400);
      expect(response.body.error).toContain('Refresh token required');
    });
  });
  
  describe('POST /api/auth/logout', () => {
    it('should logout successfully', async () => {
      const token = jwt.sign(
        { userId: 'user-123' },
        process.env.JWT_SECRET!,
        { expiresIn: '15m' }
      );
      
      (userRepository.deleteSessionByToken as jest.Mock).mockResolvedValue(undefined);
      (redis.setEx as jest.Mock).mockResolvedValue(undefined);
      
      const response = await request(app)
        .post('/api/auth/logout')
        .set('Authorization', `Bearer ${token}`);
      
      expect(response.status).toBe(200);
      expect(response.body.success).toBe(true);
      expect(response.body.message).toBe('Logged out successfully');
    });
    
    it('should reject without token', async () => {
      const response = await request(app)
        .post('/api/auth/logout');
      
      expect(response.status).toBe(400);
      expect(response.body.error).toContain('Token required');
    });
  });
});