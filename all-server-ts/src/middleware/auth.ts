import { Request, Response, NextFunction } from 'express';
import jwt from 'jsonwebtoken';
import crypto from 'crypto';
import { redis } from '@config/redis';
import { userRepository } from '@repositories/userRepository';
import { AuthenticationError, AuthorizationError } from '@utils/errors';
import { JWTPayload } from '@/types';

declare global {
  namespace Express {
    interface Request {
      user?: JWTPayload;
      session?: any;
    }
  }
}

export async function authenticate(
  req: Request,
  res: Response,
  next: NextFunction
) {
  try {
    const token = req.headers.authorization?.replace('Bearer ', '');
    
    if (!token) {
      throw new AuthenticationError('Authentication required');
    }
    
    const tokenHash = crypto
      .createHash('sha256')
      .update(token)
      .digest('hex');
    
    const isBlacklisted = await redis.get(`blacklist:${tokenHash}`);
    if (isBlacklisted) {
      throw new AuthenticationError('Token has been revoked');
    }
    
    const decoded = jwt.verify(token, process.env.JWT_SECRET!) as JWTPayload;
    
    const session = await userRepository.findSessionByToken(tokenHash);
    if (!session || session.user_id !== decoded.userId) {
      throw new AuthenticationError('Invalid session');
    }
    
    if (new Date(session.expires_at) < new Date()) {
      throw new AuthenticationError('Session expired');
    }
    
    req.user = {
      userId: decoded.userId,
      address: decoded.address,
      lineUserId: decoded.lineUserId,
      kycTier: decoded.kycTier || 0
    };
    
    req.session = session;
    
    await userRepository.updateSessionLastUsed(session.id);
    
    next();
  } catch (error: any) {
    if (error.name === 'JsonWebTokenError') {
      next(new AuthenticationError('Invalid token'));
    } else if (error.name === 'TokenExpiredError') {
      next(new AuthenticationError('Token expired'));
    } else {
      next(error);
    }
  }
}

export function requireKYC(minLevel: number) {
  return (req: Request, res: Response, next: NextFunction) => {
    if (!req.user) {
      return next(new AuthenticationError());
    }
    
    if (req.user.kycTier < minLevel) {
      return next(
        new AuthorizationError(
          `KYC level ${minLevel} required, current level: ${req.user.kycTier}`
        )
      );
    }
    
    next();
  };
}

export function requireRole(role: string) {
  return async (req: Request, res: Response, next: NextFunction) => {
    if (!req.user) {
      return next(new AuthenticationError());
    }
    
    const user = await userRepository.findById(req.user.userId);
    if (!user) {
      return next(new AuthenticationError('User not found'));
    }
    
    const userRoles = user.metadata?.roles || [];
    if (!userRoles.includes(role)) {
      return next(new AuthorizationError(`Role '${role}' required`));
    }
    
    next();
  };
}