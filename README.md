# R2S Backend - Reserve-to-Save LINE Mini dApp

## Overview

R2S Backend는 LINE Mini dApp의 핵심 비즈니스 로직과 블록체인 상호작용을 담당하는 Node.js/TypeScript 기반 백엔드 시스템입니다.

## Quick Start

### Prerequisites
- Node.js 20+ LTS
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (optional)

### Installation

```bash
# Clone repository
git clone https://github.com/procatstler/Reserve-to-save-backend.git

# Copy environment variables
cp .env.example .env

```

### Using Docker

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop services
docker-compose down
```

## Architecture

### System Architecture

```mermaid
graph TD;
Client[Client/Frontend] --> Gateway[API Server<br/>Express/Fastify]

Gateway -->|/auth/*| Auth[Auth Service<br/>JWT/Wallet Auth]
Gateway -->|/api/*| API[API Service<br/>Business Logic]
Gateway -->|/payment/*| Payment[Payment Service<br/>DappPortal Integration]

Auth --> Redis[(Redis<br/>Sessions/Cache)]
API --> Redis
Payment --> PG[(PostgreSQL<br/>Main Database)]
API --> PG

EventReceiver[Event Receiver<br/>Blockchain Events] --> PG
EventReceiver --> Queue[BullMQ<br/>Job Queue]

Batch[Batch Service<br/>Settlement/Jobs] --> PG
Batch --> Blockchain[Kaia Blockchain<br/>Smart Contracts]

Queue --> Batch

classDef gateway fill:#e1f5fe
classDef service fill:#f3e5f5
classDef database fill:#e8f5e8
classDef blockchain fill:#fff3e0

class Gateway gateway
class Auth,API,Payment,EventReceiver,Batch service
class Redis,PG,Queue database
class Blockchain blockchain
```

### API Flow

```mermaid
sequenceDiagram
    participant U as User
    participant F as Frontend
    participant A as API Server
    participant DB as Database
    participant BC as Blockchain
    participant DP as DappPortal

    U->>F: Login with Wallet
    F->>A: GET /api/auth/nonce
    A-->>F: Return nonce & message
    F->>U: Sign message
    U-->>F: Signature
    F->>A: POST /api/auth/verify
    A->>DB: Create/Update user
    A-->>F: JWT tokens
    
    F->>A: GET /api/campaigns
    A->>DB: Query campaigns
    A-->>F: Campaign list
    
    U->>F: Join campaign
    F->>A: POST /api/payment/create
    A->>DP: Create payment
    DP-->>A: Payment URL
    A-->>F: Payment details
    F->>U: Redirect to payment
    
    U->>DP: Complete payment
    DP->>A: Webhook notification
    A->>DB: Update payment status
    A->>BC: Verify transaction
    A->>DB: Create participation
```

## Project Structure

```
src/
├── config/              # Configuration files
│   ├── database.ts     # Database config
│   ├── redis.ts        # Redis config
│   └── blockchain.ts   # Blockchain config
├── services/           # Microservices
│   ├── auth/          # Authentication service
│   ├── api/           # Main API service
│   ├── payment/       # Payment service
│   └── events/        # Event processing
├── models/            # Data models
├── repositories/      # Data access layer
├── middleware/        # Express middleware
├── utils/             # Utility functions
├── migrations/        # DB migrations
├── tests/             # Test files
│   ├── unit/         # Unit tests
│   ├── integration/  # Integration tests
│   └── e2e/          # E2E tests with Playwright
└── index.ts          # Application entry point
```

## API Documentation

### Authentication

```typescript
// Get nonce for wallet authentication
GET /api/auth/nonce?address=0x...&chainId=1001

// Verify signature and login
POST /api/auth/verify
{
  "address": "0x...",
  "signature": "0x...",
  "message": "...",
  "requestId": "uuid"
}

// LINE authentication
POST /api/auth/line
{
  "idToken": "...",
  "accessToken": "..."
}

// Refresh access token
POST /api/auth/refresh
{
  "refreshToken": "..."
}
```

### Campaigns

```typescript
// List campaigns
GET /api/campaigns?status=recruiting&page=1&limit=20

// Get campaign details
GET /api/campaigns/:id

// Create campaign (merchant only)
POST /api/campaigns
Authorization: Bearer <token>
{
  "title": "Campaign Title",
  "basePrice": "100",
  "targetAmount": "10000",
  ...
}
```

### Payments

```typescript
// Create payment
POST /api/payment/create
Authorization: Bearer <token>
{
  "campaignId": "uuid",
  "amount": "100",
  "currency": "USDT",
  "mode": "crypto"
}

// Check payment status
GET /api/payment/:paymentId/status
Authorization: Bearer <token>
```

### Full API Documentation

- **Swagger UI**: http://localhost:3001/api-docs
- **OpenAPI Spec**: [src/swagger/swagger.json](src/swagger/swagger.json)
- **Frontend Guide**: [FRONTEND_API_GUIDE.md](FRONTEND_API_GUIDE.md)

## Testing

### Run Tests

```bash
# Unit tests
npm test

# Watch mode
npm run test:watch

# Coverage report
npm run test:coverage

# E2E tests with Playwright
npm run test:e2e

# E2E tests with UI
npx playwright test --ui
```

### Test Structure

```typescript
// Unit test example
describe('AuthController', () => {
  it('should generate valid nonce', async () => {
    const response = await request(app)
      .get('/api/auth/nonce')
      .query({ address: '0x...' });
    
    expect(response.status).toBe(200);
    expect(response.body).toHaveProperty('nonce');
  });
});

// E2E test example
test('Complete authentication flow', async ({ request }) => {
  // Get nonce
  const nonceResponse = await request.get('/api/auth/nonce');
  
  // Sign message
  const signature = await wallet.signMessage(message);
  
  // Verify signature
  const authResponse = await request.post('/api/auth/verify', {
    data: { address, signature, message, requestId }
  });
  
  expect(authResponse.ok()).toBeTruthy();
});
```

## Environment Variables

```env
# Server
NODE_ENV=development
PORT=3001

# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/r2s

# Redis
REDIS_URL=redis://localhost:6379

# Blockchain
BLOCKCHAIN_RPC_URL=https://public-en-kairos.node.kaia.io
BLOCKCHAIN_CHAIN_ID=1001

# JWT
JWT_SECRET=your-secret-key-minimum-32-characters
JWT_EXPIRES_IN=15m

# DappPortal
DAPPPORTAL_API_KEY=your-api-key
DAPPPORTAL_CLIENT_SECRET=your-secret

# LINE
LINE_CHANNEL_ID=2006543970
LINE_CHANNEL_SECRET=your-secret
```

## Database Schema

### Main Tables

- **users**: User accounts and profiles
- **campaigns**: Campaign information
- **participations**: User participations in campaigns
- **payments**: Payment records
- **sessions**: User sessions
- **chain_events**: Blockchain event logs

### Migrations

```bash
# Run migrations
npm run db:migrate

# Rollback
npm run db:rollback

# Create new migration
npm run db:migrate:create -- --name add_new_table
```

## Deployment

### Production Build

```bash
# Build TypeScript
npm run build

# Start production server
npm start
```

### Docker Deployment

```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build
EXPOSE 3001
CMD ["node", "dist/index.js"]
```

### Health Checks

```bash
# Health endpoint
GET /health

# Response
{
  "status": "ok",
  "timestamp": "2024-01-01T00:00:00Z",
  "service": "r2s-backend"
}
```

## Monitoring

### Logging

- Application logs: `./logs/combined.log`
- Error logs: `./logs/error.log`
- Audit logs: `./logs/audit.log`

### Metrics

- Request rate
- Response time
- Error rate
- Database connection pool
- Redis memory usage

## Development

### Code Style

```bash
# Lint
npm run lint

# Format
npm run format

# Type check
npm run typecheck
```

### Git Hooks

```bash
# Install husky
npm run prepare

# Pre-commit
- Lint staged files
- Run tests
- Type check
```

## Resources

- [Frontend Integration Guide](FRONTEND_API_GUIDE.md)
- [Backend Complete Guide](R2S-BACKEND-COMPLETE-GUIDE.md)

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## License

MIT License - see [LICENSE](LICENSE) file for details

## Support

- GitHub Issues: [Issues](https://github.com/procatstler/Reserve-to-save-backend/issues)
