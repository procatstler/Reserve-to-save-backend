# R2S Demo Setup Guide

This demo provides a complete testing environment for the R2S (Reserve to Save) backend system, including mock data, API endpoints, and a frontend integration example.

## Components

### 1. Demo Data Seeding (`seed.go`)
Seeds the database with test data including:
- Demo users (Alice, Bob, Carol)
- Demo merchants (Starbucks, CU)
- Demo campaigns (coffee, lunch boxes, snacks)
- Demo participations

### 2. Demo API Server (`demo_api.go`)
Provides REST API endpoints for testing:
- `GET /demo/users` - List demo users
- `GET /demo/campaigns` - List demo campaigns
- `GET /demo/auth?wallet=0x...` - Get demo auth token
- `GET /demo/stats` - Get system statistics
- `GET /demo/health` - Health check

### 3. Frontend Integration (`frontend-integration.html`)
Interactive web interface for testing all APIs:
- View demo users and campaigns
- Build blockchain transactions
- Test authentication flow
- Monitor system statistics

## Quick Start

### Step 1: Set up environment
Create a `.env` file in the root directory:
```env
DATABASE_URL=postgresql://postgres:password@localhost:5432/r2s_dev?sslmode=disable
DEMO_PORT=3008
TX_HELPER_PORT=3006

# For TX Helper
BLOCKCHAIN_RPC_URL=https://public-en.node.kaia.io
CAMPAIGN_FACTORY_ADDRESS=0x1234567890123456789012345678901234567890
USDT_ADDRESS=0x0987654321098765432109876543210987654321
```

### Step 2: Seed demo data
```bash
cd demo
go run seed.go
```

Expected output:
```
Connected to database
Clearing existing demo data...
Inserting demo users...
Inserting demo campaigns...
Inserting demo participations...
Demo data seeded successfully!

Demo Users:
- Alice: 0xAb8483F64d9C6d1EcF9b849Ae677dD3315835cb2
- Bob:   0x4B20993Bc481177ec7E8f571ceCaE8A9e22C02db
- Carol: 0x78731D3Ca6b7E34aC0F824c42a7cC18A495cabaB

Demo Campaigns:
- Starbucks Americano (recruiting)
- CU Lunch Box (recruiting)
- GS25 Snacks (reached)
```

### Step 3: Start the demo API server
```bash
go run demo_api.go
```

The server will start on port 3008 (or the port specified in DEMO_PORT).

### Step 4: Start the TX Helper service
```bash
cd ../tx-helper
go run main.go
```

The TX Helper will start on port 3006.

### Step 5: Open the frontend demo
Open `demo/frontend-integration.html` in your web browser.

## Testing Flows

### 1. Authentication Flow
```bash
# Get demo auth token for Alice
curl "http://localhost:3008/demo/auth?wallet=0xAb8483F64d9C6d1EcF9b849Ae677dD3315835cb2"
```

### 2. Campaign Participation Flow
1. Open the frontend demo
2. Select "Approve USDT" in Transaction Builder
3. Enter user address and campaign address
4. Click "Build Transaction" to get the transaction data
5. Sign and send using MetaMask or similar wallet

### 3. View Campaign Info
```bash
# Get campaign information from blockchain
curl "http://localhost:3006/tx/campaign-info?address=0x1234567890123456789012345678901234567890"
```

### 4. Check System Stats
```bash
# Get overall system statistics
curl http://localhost:3008/demo/stats
```

## Demo User Accounts

| Name  | Wallet Address                             | Email              | KYC Tier |
|-------|--------------------------------------------|--------------------|----------|
| Alice | 0xAb8483F64d9C6d1EcF9b849Ae677dD3315835cb2 | alice@example.com  | 1        |
| Bob   | 0x4B20993Bc481177ec7E8f571ceCaE8A9e22C02db | bob@example.com    | 1        |
| Carol | 0x78731D3Ca6b7E34aC0F824c42a7cC18A495cabaB | carol@example.com  | 0        |

## Demo Campaigns

| Title                      | Status     | Discount | Progress |
|----------------------------|------------|----------|----------|
| Starbucks Americano - 30% OFF | recruiting | 30%      | 70%      |
| CU Lunch Box Special       | recruiting | 25%      | 70%      |
| GS25 Snack Bundle          | reached    | 20%      | 100%     |

## API Response Examples

### Get Demo Users
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "wallet": "0xAb8483F64d9C6d1EcF9b849Ae677dD3315835cb2",
      "lineUserId": "U123456789",
      "displayName": "Alice Kim",
      "email": "alice@example.com",
      "kycTier": 1,
      "testPassword": "demo123"
    }
  ],
  "message": "Use these demo accounts for testing"
}
```

### Build Transaction
```json
{
  "success": true,
  "data": {
    "transaction": {
      "to": "0x1234567890123456789012345678901234567890",
      "from": "0xAb8483F64d9C6d1EcF9b849Ae677dD3315835cb2",
      "data": "0x...",
      "value": "0",
      "gasLimit": 300000,
      "gasPrice": "25000000000",
      "nonce": 0,
      "chainId": "8217"
    },
    "message": "Sign and send this transaction to join the campaign"
  }
}
```

## Troubleshooting

### Database Connection Error
- Ensure PostgreSQL is running
- Check DATABASE_URL in .env file
- Verify database exists: `createdb r2s_dev`

### Port Already in Use
- Change DEMO_PORT or TX_HELPER_PORT in .env
- Check for running processes: `lsof -i :3008`

### Transaction Building Fails
- Verify BLOCKCHAIN_RPC_URL is correct
- Check contract addresses are valid
- Ensure network connection to blockchain node

## Notes

- This is a demo environment with simplified authentication
- Demo tokens are for testing only and not secure
- All example.com emails are reserved for demo users
- Transaction signing must be done client-side with actual wallets