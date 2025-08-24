import { Request, Response, NextFunction } from 'express';
import crypto from 'crypto';
import jwt from 'jsonwebtoken';
import { ethers } from 'ethers';
import axios from 'axios';
import { redis, setWithExpiry } from '@config/redis';
import { userRepository } from '@repositories/userRepository';
import { logger } from '@utils/logger';
import { AppError, AuthenticationError } from '@utils/errors';
import { v4 as uuidv4 } from 'uuid';

export class AuthController {
  async getNonce(req: Request, res: Response, next: NextFunction) {
    try {
      const { address, chainId } = req.query;
      
      if (!address || !ethers.isAddress(address as string)) {
        throw new AppError('Invalid wallet address', 400);
      }
      
      const nonce = crypto.randomBytes(16).toString('hex');
      const requestId = uuidv4();
      const issuedAt = new Date().toISOString();
      const expiresAt = new Date(Date.now() + 6 * 60 * 1000).toISOString();
      
      const domain = process.env.APP_URL || 'https://r2s.io';
      const message = `${domain} wants you to sign in with your wallet:
${address}

URI: ${domain}
Version: 1
Chain ID: ${chainId || '1001'}
Nonce: ${nonce}
Issued At: ${issuedAt}
Expiration Time: ${expiresAt}
Request ID: ${requestId}
Statement: Sign to authenticate with R2S platform.`;
      
      const nonceHash = crypto
        .createHash('sha256')
        .update(nonce)
        .digest('hex');
      
      await setWithExpiry(
        `nonce:${nonceHash}`,
        JSON.stringify({
          address: (address as string).toLowerCase(),
          chainId: chainId || '1001',
          requestId,
          expiresAt
        }),
        360
      );
      
      res.json({
        nonce,
        message,
        requestId,
        expiresAt
      });
    } catch (error) {
      next(error);
    }
  }
  
  async verifySignature(req: Request, res: Response, next: NextFunction) {
    try {
      const { address, signature, message, requestId } = req.body;
      
      if (!address || !signature || !message || !requestId) {
        throw new AppError('Missing required fields', 400);
      }
      
      const nonceMatch = message.match(/Nonce: ([a-f0-9]{32})/);
      if (!nonceMatch) {
        throw new AppError('Invalid message format', 400);
      }
      
      const nonce = nonceMatch[1];
      const nonceHash = crypto
        .createHash('sha256')
        .update(nonce)
        .digest('hex');
      
      const nonceData = await redis.get(`nonce:${nonceHash}`);
      if (!nonceData) {
        throw new AuthenticationError('Invalid or expired nonce');
      }
      
      const parsedNonceData = JSON.parse(nonceData);
      
      if (parsedNonceData.address !== address.toLowerCase()) {
        throw new AuthenticationError('Address mismatch');
      }
      
      if (new Date(parsedNonceData.expiresAt) < new Date()) {
        throw new AuthenticationError('Nonce expired');
      }
      
      const recoveredAddress = ethers.verifyMessage(message, signature);
      if (recoveredAddress.toLowerCase() !== address.toLowerCase()) {
        throw new AuthenticationError('Invalid signature');
      }
      
      await redis.del(`nonce:${nonceHash}`);
      
      let user = await userRepository.findByWalletAddress(address);
      if (!user) {
        user = await userRepository.create({
          wallet_address: address.toLowerCase(),
          metadata: {
            firstLogin: new Date().toISOString(),
            signupSource: 'wallet'
          }
        });
      } else {
        await userRepository.updateLastLogin(user.id);
      }
      
      const accessToken = jwt.sign(
        {
          userId: user.id,
          address: user.wallet_address,
          kycTier: user.kyc_tier
        },
        process.env.JWT_SECRET!,
        {
          expiresIn: process.env.JWT_EXPIRES_IN || '15m',
          issuer: 'r2s-auth',
          audience: 'r2s-api'
        }
      );
      
      const refreshToken = jwt.sign(
        {
          userId: user.id,
          address: user.wallet_address
        },
        process.env.JWT_REFRESH_SECRET!,
        {
          expiresIn: process.env.JWT_REFRESH_EXPIRES_IN || '7d',
          issuer: 'r2s-auth',
          audience: 'r2s-api'
        }
      );
      
      const sessionId = uuidv4();
      const tokenHash = crypto
        .createHash('sha256')
        .update(accessToken)
        .digest('hex');
      const refreshTokenHash = crypto
        .createHash('sha256')
        .update(refreshToken)
        .digest('hex');
      
      await userRepository.createSession({
        id: sessionId,
        user_id: user.id,
        token_hash: tokenHash,
        refresh_token_hash: refreshTokenHash,
        ip_address: req.ip,
        user_agent: req.headers['user-agent'],
        device_fingerprint: undefined,
        expires_at: new Date(Date.now() + 15 * 60 * 1000),
        refresh_expires_at: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
        created_at: new Date(),
        last_used_at: new Date()
      });
      
      logger.audit({
        user_id: user.id,
        action: 'LOGIN',
        resource_type: 'session',
        resource_id: sessionId,
        ip_address: req.ip,
        user_agent: req.headers['user-agent']
      });
      
      res.json({
        success: true,
        accessToken,
        refreshToken,
        user: {
          id: user.id,
          address: user.wallet_address,
          kycTier: user.kyc_tier,
          lineConnected: !!user.line_user_id
        }
      });
    } catch (error) {
      next(error);
    }
  }
  
  async lineAuth(req: Request, res: Response, next: NextFunction) {
    try {
      const { idToken, accessToken } = req.body;
      
      if (!idToken || !accessToken) {
        throw new AppError('Missing LINE tokens', 400);
      }
      
      const lineProfile = await this.verifyLineToken(idToken, accessToken);
      
      let user = await userRepository.findByLineUserId(lineProfile.userId);
      
      if (!user) {
        user = await userRepository.create({
          line_user_id: lineProfile.userId,
          line_display_name: lineProfile.displayName,
          line_picture_url: lineProfile.pictureUrl,
          email: lineProfile.email,
          metadata: {
            firstLogin: new Date().toISOString(),
            signupSource: 'line'
          }
        } as any);
      } else {
        await userRepository.updateLineProfile(user.id, {
          line_display_name: lineProfile.displayName,
          line_picture_url: lineProfile.pictureUrl
        });
      }
      
      const jwtToken = jwt.sign(
        {
          userId: user.id,
          lineUserId: user.line_user_id,
          kycTier: user.kyc_tier
        },
        process.env.JWT_SECRET!,
        {
          expiresIn: process.env.JWT_EXPIRES_IN || '15m'
        }
      );
      
      res.json({
        success: true,
        token: jwtToken,
        user: {
          id: user.id,
          lineUserId: user.line_user_id,
          displayName: user.line_display_name,
          pictureUrl: user.line_picture_url,
          walletConnected: !!user.wallet_address,
          kycTier: user.kyc_tier
        }
      });
    } catch (error) {
      next(error);
    }
  }
  
  private async verifyLineToken(idToken: string, accessToken: string) {
    try {
      const response = await axios.post(
        'https://api.line.me/oauth2/v2.1/verify',
        {
          id_token: idToken,
          client_id: process.env.LINE_CHANNEL_ID
        }
      );
      
      if (response.data.client_id !== process.env.LINE_CHANNEL_ID) {
        throw new AuthenticationError('Invalid LINE token');
      }
      
      const profileResponse = await axios.get(
        'https://api.line.me/v2/profile',
        {
          headers: {
            Authorization: `Bearer ${accessToken}`
          }
        }
      );
      
      return {
        userId: response.data.sub,
        displayName: profileResponse.data.displayName,
        pictureUrl: profileResponse.data.pictureUrl,
        email: response.data.email
      };
    } catch (error) {
      logger.error('LINE token verification failed:', error);
      throw new AuthenticationError('LINE authentication failed');
    }
  }
  
  async refreshToken(req: Request, res: Response, next: NextFunction) {
    try {
      const { refreshToken } = req.body;
      
      if (!refreshToken) {
        throw new AppError('Refresh token required', 400);
      }
      
      let decoded: any;
      try {
        decoded = jwt.verify(refreshToken, process.env.JWT_REFRESH_SECRET!);
      } catch (error) {
        throw new AuthenticationError('Invalid refresh token');
      }
      
      const refreshTokenHash = crypto
        .createHash('sha256')
        .update(refreshToken)
        .digest('hex');
      
      const session = await userRepository.findSessionByRefreshToken(refreshTokenHash);
      if (!session || session.user_id !== decoded.userId) {
        throw new AuthenticationError('Invalid session');
      }
      
      const user = await userRepository.findById(decoded.userId);
      if (!user) {
        throw new AppError('User not found', 404);
      }
      
      const newAccessToken = jwt.sign(
        {
          userId: user.id,
          address: user.wallet_address,
          kycTier: user.kyc_tier
        },
        process.env.JWT_SECRET!,
        {
          expiresIn: process.env.JWT_EXPIRES_IN || '15m'
        }
      );
      
      const tokenHash = crypto
        .createHash('sha256')
        .update(newAccessToken)
        .digest('hex');
      
      await userRepository.updateSession(session.id, {
        token_hash: tokenHash,
        expires_at: new Date(Date.now() + 15 * 60 * 1000),
        last_used_at: new Date()
      });
      
      res.json({
        success: true,
        accessToken: newAccessToken
      });
    } catch (error) {
      next(error);
    }
  }
  
  async logout(req: Request, res: Response, next: NextFunction) {
    try {
      const token = req.headers.authorization?.replace('Bearer ', '');
      
      if (!token) {
        throw new AppError('Token required', 400);
      }
      
      const tokenHash = crypto
        .createHash('sha256')
        .update(token)
        .digest('hex');
      
      await userRepository.deleteSessionByToken(tokenHash);
      
      const decoded = jwt.decode(token) as any;
      if (decoded?.exp) {
        const ttl = decoded.exp - Math.floor(Date.now() / 1000);
        if (ttl > 0) {
          await setWithExpiry(`blacklist:${tokenHash}`, '1', ttl);
        }
      }
      
      res.json({
        success: true,
        message: 'Logged out successfully'
      });
    } catch (error) {
      next(error);
    }
  }
}

export const authController = new AuthController();