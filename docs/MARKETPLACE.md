# Virtual Garden Marketplace System

## Overview

The Virtual Garden Marketplace System is a plant marketplace with blockchain-inspired mechanics for the gardening game. It enables players to trade plants, request deliveries, and hold coins in wallets backed by ECDSA key pairs. See [Current Limitations](#current-limitations) for the parts of the design that are not implemented.

## Features

### 🏪 Two Marketplace Modes

1. **Service Orders (Delivery Requests)**
   - Players request a plant type and offer a reward of at least 10 coins
   - Deadlines must be at least one hour in the future; completing late costs 20% of the reward
   - Completing a request raises the deliverer's reputation; completing late lowers it

2. **Direct Sales Listings**
   - Players list plants they own at a fixed price
   - Listings carry a 7-day expiry date
   - Listings can be filtered by plant type and price range

### ⛓️ Blockchain Mechanics

- **Digitally Signed Transactions**: transactions submitted to `POST /blockchain/transactions` must carry a valid ECDSA signature
- **Replay Protection**: a submitted transaction's nonce must be greater than the sender wallet's last nonce
- **Transaction Records**: purchases and delivery rewards are recorded as pending transactions
- **System Commission**: 0.5% fee on direct sales

### 💰 Economic System

- **In-Game Currency**: wallets hold coin balances
- **Reputation Scoring**: completed trades and deliveries raise reputation
- **Penalties**: late deliveries lose 20% of the reward and 0.5 reputation

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

### Penalties & Rewards
- **Late Deliveries**: completing after the deadline removes 20% of the reward and 0.5 reputation
- **Successful Trades**: +0.1 reputation for the buyer, the seller, and the deliverer
- **Listing Expiration**: listings carry a 7-day expiry date
- **Commission**: 0.5% of the listing price is withheld from the seller

### Security Measures
- **AES-256-GCM Encryption**: private keys of wallets created through `POST /wallet` are encrypted with a key derived from the wallet password
- **ECDSA Signatures**: required for transactions submitted to `POST /blockchain/transactions`
- **Transaction Nonces**: prevent replaying submitted transactions

## Workflows

### Direct Purchase Flow
1. **Seller lists a plant** they own at a fixed price
2. **Buyer purchases** the listing, and the price leaves the buyer's wallet immediately
3. **Seller receives** the price minus the 0.5% commission
4. **Plant moves** to the buyer's garden; a garden is created if the buyer has none
5. **A pending transaction** is recorded and the listing is marked sold

### Delivery Request Flow
1. **Player A creates** a delivery request for a plant type with a reward and a deadline
2. **Player B accepts** the request
3. **Player B marks it complete**, and the reward moves from A's wallet to B's, minus the late penalty if the deadline passed
4. **A pending reward transaction** is recorded

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
- Private keys of wallets created through `POST /wallet` are stored encrypted with AES-256-GCM
- The encryption key is a single SHA-256 hash of the wallet password
- Keys are decrypted only to sign a transfer or a transaction

### Transaction Security
- Transactions submitted to the blockchain endpoint require valid ECDSA signatures
- Nonce checks reject replayed submissions

## Performance Considerations

### Database Optimization
- The ledger endpoint supports `limit` and `offset` pagination

## Current Limitations

These parts of the original design are not implemented. The [marketplace phase](plans/05-marketplace.md) of the development plan addresses them.

- Pending transactions are never validated or added to blocks: nothing calls `ProcessPendingTransactions`, so the ledger stays empty and there is no consensus step.
- Purchases settle immediately, with no escrow, seller confirmation (2-of-2 signatures), or cooling period for large amounts.
- Completing a delivery request does not check or transfer any plant.
- Expired listings and overdue requests are never cleaned up: nothing calls `CleanupExpiredItems`.
- Wallets created implicitly by a purchase or delivery start with 100 coins and store no private key.
- Wallet passwords are hashed once with SHA-256 instead of a slow key-derivation function.
- Reputation does not influence any outcome.

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