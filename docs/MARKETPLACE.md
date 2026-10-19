# Virtual Garden Marketplace System

## Overview

The Virtual Garden Marketplace System is a decentralized plant marketplace with blockchain-inspired mechanics for the gardening game. It enables players to trade plants, request deliveries, and participate in a secure economic system with digital signatures and immutable records.

## Features

### 🏪 Two Marketplace Modes

1. **Service Orders (Delivery Requests)**
   - Players can request specific plants with delivery rewards
   - Time-based deadlines with penalties for late delivery
   - Reputation system for successful/failed deliveries

2. **Direct Sales Listings**
   - Players can list harvested plants for direct purchase
   - Fixed pricing with automatic expiration (7 days)
   - Real-time marketplace browsing with filters

### ⛓️ Blockchain Mechanics

- **Digitally Signed Transactions**: ECDSA signatures for all transactions
- **Immutable Transaction Ledger**: Public blockchain with transaction history
- **Simplified Consensus**: 3-node validation system
- **Double-Spend Prevention**: Nonce-based transaction verification
- **System Commission**: 0.5% fee on all marketplace transactions

### 💰 Economic System

- **In-Game Currency**: Secure wallet system with encrypted private keys
- **Reputation Scoring**: Player reputation affects transaction success rates
- **Penalties & Rewards**: Failed deliveries incur 20% penalty + reputation decrease
- **Cooling Periods**: Large transactions (>1000 currency) have 24-hour cooling period

## API Endpoints

### Marketplace

#### Delivery Requests
```http
POST /api/v1/marketplace/delivery-requests
{
  "plant_type_id": "uuid",
  "reward": 100,
  "deadline": "2024-01-15T18:00:00Z"
}

GET /api/v1/marketplace/delivery-requests

POST /api/v1/marketplace/delivery-requests/{id}/accept

POST /api/v1/marketplace/delivery-requests/{id}/complete
```

#### Plant Listings
```http
POST /api/v1/marketplace/listings
{
  "plant_id": "uuid",
  "price": 50
}

GET /api/v1/marketplace/listings?plant_type_id=uuid&min_price=10&max_price=100

POST /api/v1/marketplace/listings/{id}/buy
```

### Blockchain

```http
POST /api/v1/blockchain/transactions
{
  "sender_key": "public_key",
  "receiver_key": "public_key",
  "plant_id": "uuid",
  "amount": 100,
  "tx_type": "transfer",
  "nonce": 1234567890,
  "signature": "hex_signature"
}

GET /api/v1/blockchain/transactions/{hash}

GET /api/v1/blockchain/ledger?limit=10&offset=0
```

### Wallet

```http
POST /api/v1/wallet
{
  "password": "secure_password"
}

GET /api/v1/wallet/balance

POST /api/v1/wallet/transfer
{
  "to_public_key": "receiver_public_key",
  "amount": 100,
  "password": "wallet_password"
}

POST /api/v1/wallet/sign
{
  "receiver_key": "public_key",
  "amount": 100,
  "tx_type": "transfer",
  "nonce": 1234567890,
  "password": "wallet_password"
}
```

## Business Rules

### Transaction Processing
- **2-of-2 Signatures**: Marketplace transactions require both buyer and seller signatures
- **3-Node Consensus**: Blockchain transactions validated by 3 active consensus nodes
- **System Commission**: 0.5% fee automatically deducted from all transactions
- **Nonce Verification**: Prevents replay attacks and double-spending

### Penalties & Rewards
- **Failed Deliveries**: 20% reward penalty + 0.5 reputation decrease
- **Successful Transactions**: +0.1 reputation point per successful transaction
- **Listing Expiration**: Automatic cancellation after 7 days
- **Large Transaction Cooling**: 24-hour period for transactions >1000 currency

### Security Measures
- **AES-256-GCM Encryption**: Private keys encrypted with user passwords
- **ECDSA Signatures**: All transactions digitally signed
- **Transaction Nonces**: Prevents replay attacks
- **Public Key Verification**: Ensures transaction authenticity

## Workflows

### Direct Purchase Flow
1. **Seller lists plant** at fixed price
2. **Buyer initiates purchase** (signs TX1)
3. **System holds plant** in escrow
4. **Seller signs TX2** to confirm
5. **Transaction added** to blockchain after validation
6. **Plant transferred** to buyer, currency to seller

### Delivery Request Flow
1. **Player A creates** delivery request for rare plant
2. **Player B accepts** request (signs contract)
3. **Player B delivers** plant before deadline
4. **Both parties sign** completion transaction
5. **Reward transferred** after validation

## Database Schema

### Core Tables
- `delivery_requests`: Service orders with deadlines and rewards
- `plant_listings`: Active marketplace listings
- `transactions`: Blockchain transaction records
- `blocks`: Blockchain blocks with merkle roots
- `wallets`: User wallets with encrypted private keys
- `consensus_nodes`: Active validation nodes

### Key Relationships
- Users can have multiple delivery requests and listings
- Transactions reference sender/receiver wallets
- Blocks contain multiple transactions
- Wallets are linked to user accounts

## Security Considerations

### Private Key Management
- Private keys are **never stored in plain text**
- AES-256-GCM encryption with user passwords
- Keys are only decrypted temporarily for signing
- Password-based key derivation for additional security

### Transaction Security
- All transactions require valid ECDSA signatures
- Nonce-based replay protection
- Double-spend prevention through transaction ordering
- Public key verification for all operations

### Consensus Security
- 3-node validation requirement
- Merkle tree verification for block integrity
- Timestamp-based block ordering
- Validator reputation tracking

## Performance Considerations

### Database Optimization
- Indexed fields for fast queries
- Pagination support for large datasets
- Efficient joins for related data
- Automatic cleanup of expired items

### Blockchain Performance
- Batch transaction processing
- Efficient merkle tree calculations
- Optimized signature verification
- Cached public key lookups

## Monitoring & Maintenance

### Automated Tasks
- **Expired Listing Cleanup**: Daily cleanup of expired listings
- **Failed Request Cleanup**: Automatic expiration of overdue requests
- **Transaction Processing**: Background processing of pending transactions
- **Reputation Updates**: Automatic reputation adjustments

### Health Checks
- Database connectivity monitoring
- Blockchain consensus node status
- Transaction processing latency
- Wallet balance verification

## Future Enhancements

### Planned Features
- **Advanced Filtering**: More sophisticated marketplace search
- **Bulk Operations**: Batch listing and purchasing
- **Auction System**: Time-based bidding for rare plants
- **Guild Trading**: Group-based trading and reputation
- **Cross-Chain Integration**: Interoperability with other game systems

### Technical Improvements
- **Sharding**: Horizontal scaling for high transaction volumes
- **Zero-Knowledge Proofs**: Enhanced privacy for transactions
- **Smart Contracts**: Programmable trading conditions
- **Mobile SDK**: Native mobile wallet integration