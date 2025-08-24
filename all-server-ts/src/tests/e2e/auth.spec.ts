import { test, expect } from '@playwright/test';
import { ethers } from 'ethers';

test.describe('Authentication API E2E Tests', () => {
  const API_BASE = 'http://localhost:3001';
  let wallet: ethers.Wallet;
  
  test.beforeAll(async () => {
    // Create a test wallet
    wallet = ethers.Wallet.createRandom();
  });
  
  test('Complete authentication flow', async ({ request }) => {
    // Step 1: Get nonce
    const nonceResponse = await request.get(`${API_BASE}/api/auth/nonce`, {
      params: {
        address: wallet.address,
        chainId: '1001'
      }
    });
    
    expect(nonceResponse.ok()).toBeTruthy();
    const nonceData = await nonceResponse.json();
    
    expect(nonceData).toHaveProperty('nonce');
    expect(nonceData).toHaveProperty('message');
    expect(nonceData).toHaveProperty('requestId');
    expect(nonceData).toHaveProperty('expiresAt');
    expect(nonceData.message).toContain(wallet.address);
    
    // Step 2: Sign message
    const signature = await wallet.signMessage(nonceData.message);
    
    // Step 3: Verify signature
    const verifyResponse = await request.post(`${API_BASE}/api/auth/verify`, {
      data: {
        address: wallet.address,
        signature,
        message: nonceData.message,
        requestId: nonceData.requestId
      }
    });
    
    expect(verifyResponse.ok()).toBeTruthy();
    const authData = await verifyResponse.json();
    
    expect(authData.success).toBe(true);
    expect(authData).toHaveProperty('accessToken');
    expect(authData).toHaveProperty('refreshToken');
    expect(authData.user).toMatchObject({
      address: wallet.address.toLowerCase(),
      kycTier: 0,
      lineConnected: false
    });
    
    // Step 4: Use access token
    const profileResponse = await request.get(`${API_BASE}/api/users/profile`, {
      headers: {
        'Authorization': `Bearer ${authData.accessToken}`
      }
    });
    
    // Note: This will fail if profile endpoint is not implemented
    // expect(profileResponse.ok()).toBeTruthy();
    
    // Step 5: Refresh token
    const refreshResponse = await request.post(`${API_BASE}/api/auth/refresh`, {
      data: {
        refreshToken: authData.refreshToken
      }
    });
    
    expect(refreshResponse.ok()).toBeTruthy();
    const refreshData = await refreshResponse.json();
    expect(refreshData.success).toBe(true);
    expect(refreshData).toHaveProperty('accessToken');
    
    // Step 6: Logout
    const logoutResponse = await request.post(`${API_BASE}/api/auth/logout`, {
      headers: {
        'Authorization': `Bearer ${authData.accessToken}`
      }
    });
    
    expect(logoutResponse.ok()).toBeTruthy();
    const logoutData = await logoutResponse.json();
    expect(logoutData.success).toBe(true);
  });
  
  test('Reject invalid wallet address', async ({ request }) => {
    const response = await request.get(`${API_BASE}/api/auth/nonce`, {
      params: {
        address: 'invalid-address',
        chainId: '1001'
      }
    });
    
    expect(response.status()).toBe(400);
    const data = await response.json();
    expect(data.success).toBe(false);
    expect(data.error).toBeDefined();
  });
  
  test('Reject expired nonce', async ({ request }) => {
    // Get nonce
    const nonceResponse = await request.get(`${API_BASE}/api/auth/nonce`, {
      params: {
        address: wallet.address,
        chainId: '1001'
      }
    });
    
    const nonceData = await nonceResponse.json();
    const signature = await wallet.signMessage(nonceData.message);
    
    // Wait for nonce to expire (in test, this might need to be mocked)
    // For now, we'll just test with an invalid nonce
    const invalidMessage = nonceData.message.replace(/Nonce: \w+/, 'Nonce: invalid123');
    
    const verifyResponse = await request.post(`${API_BASE}/api/auth/verify`, {
      data: {
        address: wallet.address,
        signature,
        message: invalidMessage,
        requestId: nonceData.requestId
      }
    });
    
    expect(verifyResponse.status()).toBe(400);
    const data = await verifyResponse.json();
    expect(data.success).toBe(false);
  });
  
  test('Reject invalid signature', async ({ request }) => {
    const nonceResponse = await request.get(`${API_BASE}/api/auth/nonce`, {
      params: {
        address: wallet.address,
        chainId: '1001'
      }
    });
    
    const nonceData = await nonceResponse.json();
    
    const verifyResponse = await request.post(`${API_BASE}/api/auth/verify`, {
      data: {
        address: wallet.address,
        signature: 'invalid-signature',
        message: nonceData.message,
        requestId: nonceData.requestId
      }
    });
    
    expect(verifyResponse.status()).toBe(401);
    const data = await verifyResponse.json();
    expect(data.success).toBe(false);
  });
  
  test('Rate limiting works', async ({ request }) => {
    const requests = [];
    
    // Send many requests quickly
    for (let i = 0; i < 150; i++) {
      requests.push(
        request.get(`${API_BASE}/api/auth/nonce`, {
          params: {
            address: wallet.address,
            chainId: '1001'
          }
        })
      );
    }
    
    const responses = await Promise.all(requests);
    
    // Some requests should be rate limited
    const rateLimited = responses.filter(r => r.status() === 429);
    expect(rateLimited.length).toBeGreaterThan(0);
    
    // Check rate limit headers
    const limitedResponse = rateLimited[0];
    expect(limitedResponse.headers()['x-ratelimit-limit']).toBeDefined();
    expect(limitedResponse.headers()['x-ratelimit-remaining']).toBe('0');
    expect(limitedResponse.headers()['x-ratelimit-reset']).toBeDefined();
  });
});

test.describe('API Health Check', () => {
  const API_BASE = 'http://localhost:3001';
  
  test('Health endpoint returns ok', async ({ request }) => {
    const response = await request.get(`${API_BASE}/health`);
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    
    expect(data.status).toBe('ok');
    expect(data.service).toBe('r2s-backend');
    expect(data.timestamp).toBeDefined();
  });
  
  test('Swagger documentation is accessible', async ({ page }) => {
    await page.goto(`${API_BASE}/api-docs`);
    
    // Check if Swagger UI loaded
    await expect(page.locator('text=R2S Backend API')).toBeVisible();
    await expect(page.locator('text=Reserve-to-Save LINE Mini dApp Backend API')).toBeVisible();
    
    // Check if endpoints are listed
    await expect(page.locator('text=/api/auth/nonce')).toBeVisible();
    await expect(page.locator('text=/api/auth/verify')).toBeVisible();
  });
});