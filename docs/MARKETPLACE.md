# Marketplace

## Overview

Players trade with the same coins they earn from harvests and spend in the store. There are no wallets, keys, or ledger: user coins are the only currency. See [Current Limitations](#current-limitations) for what the marketplace does not do yet.

## Features

### Delivery requests

- A player requests a plant type and offers a reward of at least 10 coins.
- The deadline must be at least one hour in the future.
- The requester must hold the reward when creating the request, but it is charged only when the request is completed.
- Another player accepts the request before its deadline; only one player can accept it.
- The accepting player completes the request and receives the reward and 10 XP. Completing after the deadline pays 80% of the reward.

### Plant listings

- A player lists a harvestable plant they grow, in their own or a shared garden, at a fixed price of at least 1 coin.
- A plant can have only one active listing. Listings expire after 7 days, after which the plant can be listed again.
- Listings can be filtered by plant type and price range.
- Buying a listing charges the buyer the full price and pays the seller the price minus a 0.5% commission, rounded down, plus 10 XP. The plant moves to the first free plot across the buyer's gardens, and the buyer becomes its grower, who earns its harvest. A buyer without a free plot gets `409 Conflict` and nothing changes.

### Integrity guarantees

- Purchases and delivery completions are single database transactions: coins and plants move together or not at all.
- Each command first claims its listing or request with a conditional update, so concurrent or repeated purchases, acceptances, and completions apply once. Later attempts receive `409 Conflict`.
- Balances cannot go negative: a buyer or requester who cannot pay gets an error and nothing changes.
- Listings and requests expose only public profile fields (ID, username, avatar, and level) of sellers, requesters, and acceptors.

## API Endpoints

All marketplace endpoints require authentication.

```http
POST /api/v1/marketplace/delivery-requests
{
  "plant_type_id": "uuid",
  "reward": 100,
  "deadline": "2026-01-15T18:00:00Z"
}

GET  /api/v1/marketplace/delivery-requests
POST /api/v1/marketplace/delivery-requests/{id}/accept
POST /api/v1/marketplace/delivery-requests/{id}/complete

POST /api/v1/marketplace/listings
{
  "plant_id": "uuid",
  "price": 50
}

GET  /api/v1/marketplace/listings?plant_type_id=uuid&min_price=10&max_price=100
POST /api/v1/marketplace/listings/{id}/buy

GET  /api/v1/marketplace/stats
```

`GET /delivery-requests` returns open requests whose deadline has not passed, and `GET /listings` returns active, unexpired listings. Completing a request returns the reward paid in `reward_paid`.

## Database Schema

- `delivery_requests`: requests with requester, plant type, reward, deadline, status (`open`, `accepted`, `completed`), and acceptor. A check constraint keeps rewards positive.
- `plant_listings`: listings with seller, plant, price, status (`active`, `sold`, `expired`), buyer, and expiry date. A partial unique index allows one active listing per plant, and a check constraint keeps prices positive.

Deleting a plant or a garden deletes the listings of its plants, sold or not.

## Current Limitations

The [marketplace phase](plans/05-marketplace.md) of the development plan addresses these.

- Completing a delivery request does not check or transfer any plant.
- Accepted requests disappear from `GET /delivery-requests`, so the interface cannot show a player the requests they accepted.
- Listings cannot be cancelled, and the plant can still be harvested or removed while listed; either deletes the listing.
- There is no economic history beyond the status and buyer of each listing and request.
- Reputation no longer exists.
