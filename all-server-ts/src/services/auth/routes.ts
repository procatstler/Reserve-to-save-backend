import { Router } from 'express';
import { authController } from './authController';
import { validateRequest } from '@middleware/validation';
import { z } from 'zod';

const router = Router();

// Validation schemas
const getNonceSchema = z.object({
  query: z.object({
    address: z.string().regex(/^0x[a-fA-F0-9]{40}$/),
    chainId: z.string().optional()
  })
});

const verifySignatureSchema = z.object({
  body: z.object({
    address: z.string().regex(/^0x[a-fA-F0-9]{40}$/),
    signature: z.string(),
    message: z.string(),
    requestId: z.string().uuid()
  })
});

const lineAuthSchema = z.object({
  body: z.object({
    idToken: z.string(),
    accessToken: z.string()
  })
});

const refreshTokenSchema = z.object({
  body: z.object({
    refreshToken: z.string()
  })
});

// Routes
router.get('/nonce', validateRequest(getNonceSchema), authController.getNonce);
router.post('/verify', validateRequest(verifySignatureSchema), authController.verifySignature);
router.post('/line', validateRequest(lineAuthSchema), authController.lineAuth);
router.post('/refresh', validateRequest(refreshTokenSchema), authController.refreshToken);
router.post('/logout', authController.logout);

export const authRouter = router;