# R2S Backend API Guide for Frontend

## Overview

R2S Backend는 LINE Mini dApp의 핵심 기능을 제공하는 RESTful API 서버입니다. 
모든 API는 JSON 형식으로 통신하며, 인증이 필요한 엔드포인트는 JWT Bearer 토큰을 사용합니다.

## Base URLs

- **Development**: `http://localhost:3001`
- **Production**: `https://api.r2s.io`

## Authentication Flow

### 1. Wallet Authentication (지갑 인증)

```typescript
// 1단계: Nonce 요청
const getNonce = async (walletAddress: string) => {
  const response = await fetch(`${API_BASE}/api/auth/nonce?address=${walletAddress}&chainId=1001`);
  const data = await response.json();
  return data;
  // Response: { nonce, message, requestId, expiresAt }
};

// 2단계: 메시지 서명 (Metamask/Kaikas 등 사용)
const signMessage = async (message: string) => {
  const provider = new ethers.BrowserProvider(window.ethereum);
  const signer = await provider.getSigner();
  const signature = await signer.signMessage(message);
  return signature;
};

// 3단계: 서명 검증 및 로그인
const verifySignature = async (address: string, signature: string, message: string, requestId: string) => {
  const response = await fetch(`${API_BASE}/api/auth/verify`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ address, signature, message, requestId })
  });
  const data = await response.json();
  return data;
  // Response: { success, accessToken, refreshToken, user }
};

// 전체 플로우 예시
const walletLogin = async () => {
  try {
    // 1. 지갑 연결
    const accounts = await window.ethereum.request({ 
      method: 'eth_requestAccounts' 
    });
    const walletAddress = accounts[0];
    
    // 2. Nonce 요청
    const { nonce, message, requestId } = await getNonce(walletAddress);
    
    // 3. 메시지 서명
    const signature = await signMessage(message);
    
    // 4. 로그인
    const { accessToken, refreshToken, user } = await verifySignature(
      walletAddress, 
      signature, 
      message, 
      requestId
    );
    
    // 5. 토큰 저장
    localStorage.setItem('accessToken', accessToken);
    localStorage.setItem('refreshToken', refreshToken);
    
    return user;
  } catch (error) {
    console.error('Wallet login failed:', error);
    throw error;
  }
};
```

### 2. LINE Authentication (LINE 인증)

```typescript
// LINE SDK 초기화 (LINE Mini dApp 환경)
const lineLogin = async () => {
  try {
    // LINE 로그인
    const result = await liff.login();
    
    // ID Token과 Access Token 획득
    const idToken = liff.getIDToken();
    const accessToken = liff.getAccessToken();
    
    // Backend 인증
    const response = await fetch(`${API_BASE}/api/auth/line`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ idToken, accessToken })
    });
    
    const data = await response.json();
    // Response: { success, token, user }
    
    localStorage.setItem('accessToken', data.token);
    return data.user;
  } catch (error) {
    console.error('LINE login failed:', error);
    throw error;
  }
};
```

### 3. Token Refresh (토큰 갱신)

```typescript
const refreshAccessToken = async () => {
  const refreshToken = localStorage.getItem('refreshToken');
  
  const response = await fetch(`${API_BASE}/api/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken })
  });
  
  const data = await response.json();
  // Response: { success, accessToken }
  
  localStorage.setItem('accessToken', data.accessToken);
  return data.accessToken;
};
```

## Authenticated Requests

인증이 필요한 모든 API 요청에는 Authorization 헤더를 포함해야 합니다:

```typescript
const authenticatedFetch = async (url: string, options?: RequestInit) => {
  const accessToken = localStorage.getItem('accessToken');
  
  const response = await fetch(url, {
    ...options,
    headers: {
      ...options?.headers,
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json'
    }
  });
  
  // 토큰 만료 시 자동 갱신
  if (response.status === 401) {
    const newToken = await refreshAccessToken();
    return fetch(url, {
      ...options,
      headers: {
        ...options?.headers,
        'Authorization': `Bearer ${newToken}`,
        'Content-Type': 'application/json'
      }
    });
  }
  
  return response;
};
```

## Campaign APIs

### Get Campaign List

```typescript
interface Campaign {
  id: string;
  title: string;
  description: string;
  imageUrl: string;
  basePrice: string;
  targetAmount: string;
  currentAmount: string;
  discountRate: number;
  status: 'recruiting' | 'reached' | 'fulfillment' | 'settled';
  progress: number;
  startTime: string;
  endTime: string;
}

const getCampaigns = async (status?: string, page = 1, limit = 20) => {
  const params = new URLSearchParams({
    ...(status && { status }),
    page: page.toString(),
    limit: limit.toString()
  });
  
  const response = await fetch(`${API_BASE}/api/campaigns?${params}`);
  const data = await response.json();
  
  return data;
  // Response: { 
  //   success: true,
  //   data: Campaign[],
  //   pagination: { page, limit, total, pages }
  // }
};

// 사용 예시
const activeCampaigns = await getCampaigns('recruiting');
```

### Get Campaign Details

```typescript
const getCampaignDetails = async (campaignId: string) => {
  const response = await fetch(`${API_BASE}/api/campaigns/${campaignId}`);
  const data = await response.json();
  return data; // Campaign object
};
```

## Payment APIs

### Create Payment

```typescript
interface PaymentRequest {
  campaignId: string;
  amount: string; // USDT amount (6 decimals)
  currency: 'USDT' | 'KAIA';
  mode: 'crypto' | 'stripe';
}

const createPayment = async (paymentData: PaymentRequest) => {
  const response = await authenticatedFetch(`${API_BASE}/api/payment/create`, {
    method: 'POST',
    body: JSON.stringify(paymentData)
  });
  
  const data = await response.json();
  return data;
  // Response: {
  //   success: true,
  //   data: {
  //     paymentId: string,
  //     paymentUrl: string,  // Redirect user to this URL
  //     expiresAt: string,
  //     amount: string,
  //     currency: string
  //   }
  // }
};

// 사용 예시
const initiatePayment = async () => {
  const payment = await createPayment({
    campaignId: 'campaign-uuid',
    amount: '100.000000',  // 100 USDT
    currency: 'USDT',
    mode: 'crypto'
  });
  
  // DappPortal 결제 페이지로 리다이렉트
  window.location.href = payment.data.paymentUrl;
};
```

### Check Payment Status

```typescript
const getPaymentStatus = async (paymentId: string) => {
  const response = await authenticatedFetch(
    `${API_BASE}/api/payment/${paymentId}/status`
  );
  
  const data = await response.json();
  return data;
  // Response: {
  //   success: true,
  //   data: {
  //     paymentId: string,
  //     status: 'pending' | 'completed' | 'failed',
  //     amount: string,
  //     currency: string,
  //     createdAt: string,
  //     completedAt?: string
  //   }
  // }
};
```

## User APIs

### Get User Profile

```typescript
const getUserProfile = async () => {
  const response = await authenticatedFetch(`${API_BASE}/api/users/profile`);
  const data = await response.json();
  return data;
  // Response: {
  //   id: string,
  //   walletAddress: string,
  //   lineUserId?: string,
  //   lineDisplayName?: string,
  //   email?: string,
  //   kycTier: number,
  //   status: string
  // }
};
```

### Get My Participations

```typescript
const getMyParticipations = async () => {
  const response = await authenticatedFetch(`${API_BASE}/api/participations/my`);
  const data = await response.json();
  return data;
  // Response: {
  //   success: true,
  //   data: [{
  //     id: string,
  //     campaign: Campaign,
  //     depositAmount: string,
  //     expectedRebate: string,
  //     actualRebate?: string,
  //     status: string,
  //     joinedAt: string
  //   }]
  // }
};
```

## WebSocket Events (Real-time Updates)

```typescript
// WebSocket 연결 (Socket.io 사용 예시)
import { io } from 'socket.io-client';

const socket = io(API_BASE, {
  auth: {
    token: localStorage.getItem('accessToken')
  }
});

// 캠페인 상태 업데이트 구독
socket.on('campaign:update', (data) => {
  console.log('Campaign updated:', data);
  // { campaignId, status, currentAmount, currentQty }
});

// 결제 상태 업데이트 구독
socket.on('payment:status', (data) => {
  console.log('Payment status:', data);
  // { paymentId, status, transactionHash }
});

// 참여 상태 업데이트 구독
socket.on('participation:update', (data) => {
  console.log('Participation updated:', data);
  // { participationId, status, rebateAmount }
});
```

## Error Handling

모든 API 에러는 동일한 형식으로 반환됩니다:

```typescript
interface ErrorResponse {
  success: false;
  error: string;
  statusCode: number;
}

// 에러 처리 예시
const handleApiError = (error: ErrorResponse) => {
  switch (error.statusCode) {
    case 400:
      // Validation error
      alert(`입력 오류: ${error.error}`);
      break;
    case 401:
      // Authentication error - redirect to login
      localStorage.clear();
      window.location.href = '/login';
      break;
    case 403:
      // Authorization error
      alert('권한이 없습니다');
      break;
    case 404:
      // Not found
      alert('요청한 리소스를 찾을 수 없습니다');
      break;
    case 429:
      // Rate limit
      alert('요청이 너무 많습니다. 잠시 후 다시 시도해주세요');
      break;
    default:
      // Server error
      alert('서버 오류가 발생했습니다');
  }
};
```

## LINE Mini dApp Integration

```typescript
// LIFF 초기화
import liff from '@line/liff';

const initializeLiff = async () => {
  await liff.init({ 
    liffId: process.env.NEXT_PUBLIC_LIFF_ID 
  });
  
  if (!liff.isLoggedIn()) {
    liff.login();
  }
  
  // 사용자 프로필 가져오기
  const profile = await liff.getProfile();
  
  // Backend와 연동
  const idToken = liff.getIDToken();
  const accessToken = liff.getAccessToken();
  
  // LINE 인증으로 Backend 로그인
  const response = await fetch(`${API_BASE}/api/auth/line`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ idToken, accessToken })
  });
  
  const { token, user } = await response.json();
  localStorage.setItem('accessToken', token);
  
  return user;
};
```

## Testing with Mock Data

개발 중 Mock 데이터 사용:

```typescript
// Mock API Client
class MockApiClient {
  async getCampaigns() {
    return {
      success: true,
      data: [
        {
          id: 'mock-campaign-1',
          title: '스타벅스 아메리카노',
          description: '매일 마시는 커피, 더 저렴하게!',
          imageUrl: 'https://example.com/coffee.jpg',
          basePrice: '5000000', // 5 USDT
          targetAmount: '100000000', // 100 USDT
          currentAmount: '75000000', // 75 USDT
          discountRate: 1000, // 10%
          status: 'recruiting',
          progress: 75,
          startTime: '2024-01-01T00:00:00Z',
          endTime: '2024-02-01T00:00:00Z'
        }
      ],
      pagination: {
        page: 1,
        limit: 20,
        total: 1,
        pages: 1
      }
    };
  }
  
  async createPayment(data: PaymentRequest) {
    return {
      success: true,
      data: {
        paymentId: 'mock-payment-123',
        paymentUrl: 'https://mock-payment.example.com',
        expiresAt: new Date(Date.now() + 30 * 60 * 1000).toISOString(),
        amount: data.amount,
        currency: data.currency
      }
    };
  }
}

// 환경에 따라 실제/Mock API 사용
const apiClient = process.env.NODE_ENV === 'development' 
  ? new MockApiClient() 
  : new ApiClient();
```

## TypeScript Types

```typescript
// types/api.ts
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  statusCode?: number;
}

export interface PaginatedResponse<T> extends ApiResponse<T[]> {
  pagination: {
    page: number;
    limit: number;
    total: number;
    pages: number;
  };
}

export interface User {
  id: string;
  walletAddress?: string;
  lineUserId?: string;
  lineDisplayName?: string;
  linePictureUrl?: string;
  email?: string;
  kycTier: number;
  status: 'active' | 'suspended' | 'deleted';
}

export interface Campaign {
  id: string;
  chainAddress: string;
  title: string;
  description?: string;
  imageUrl?: string;
  merchantWallet: string;
  basePrice: string;
  minQty: number;
  currentQty: number;
  targetAmount: string;
  currentAmount: string;
  discountRate: number;
  saveFloorBps: number;
  rMaxBps: number;
  startTime: string;
  endTime: string;
  status: CampaignStatus;
  progress: number;
}

export type CampaignStatus = 
  | 'draft' 
  | 'recruiting' 
  | 'reached' 
  | 'fulfillment' 
  | 'settled' 
  | 'failed' 
  | 'cancelled';

export interface Participation {
  id: string;
  campaign: Campaign;
  depositAmount: string;
  expectedRebate: string;
  actualRebate?: string;
  status: ParticipationStatus;
  joinedAt: string;
}

export type ParticipationStatus = 
  | 'active' 
  | 'pending_cancel' 
  | 'cancelled' 
  | 'settled' 
  | 'refunded';

export interface Payment {
  paymentId: string;
  campaignId: string;
  amount: string;
  currency: 'USDT' | 'KAIA' | 'KRW' | 'USD';
  mode: 'crypto' | 'stripe';
  status: PaymentStatus;
  transactionHash?: string;
  createdAt: string;
  completedAt?: string;
}

export type PaymentStatus = 
  | 'pending' 
  | 'processing' 
  | 'completed' 
  | 'failed' 
  | 'refunded';
```

## Rate Limiting

- Default: 100 requests per minute per IP
- Authenticated users: 200 requests per minute
- 초과 시 429 에러와 함께 `X-RateLimit-Reset` 헤더에 리셋 시간 반환

## Additional Resources

- **Swagger Documentation**: `https://api.r2s.io/api-docs`
- **Postman Collection**: [Download](https://api.r2s.io/postman-collection.json)
- **GraphQL Endpoint**: `https://api.r2s.io/graphql` (Coming soon)

## Support

- Email: dev@r2s.io
- Discord: [R2S Developer Community](https://discord.gg/r2s)
- GitHub Issues: [github.com/r2s/backend/issues](https://github.com/r2s/backend/issues)