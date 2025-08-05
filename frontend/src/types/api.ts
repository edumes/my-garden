export interface User {
  id: string;
  username: string;
  email: string;
  first_name?: string;
  last_name?: string;
  level: number;
  experience: number;
  coins: number;
  avatar?: string;
  timezone?: string;
  language?: string;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
  achievements?: UserAchievement[];
}

export interface UserAchievement {
  id: string;
  achievement_id: string;
  unlocked_at: string;
  achievement: Achievement;
}

export interface Achievement {
  id: string;
  name: string;
  description: string;
  icon: string;
  points: number;
  category: string;
}

export interface AuthResponse {
  token: string;
  expires_at: string;
  user: User;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  first_name?: string;
  last_name?: string;
}

export interface Garden {
  id: string;
  user_id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
  user: User;
  plants?: Plant[];
}

export interface CreateGardenRequest {
  name: string;
  description?: string;
}

export interface UpdateGardenRequest {
  name?: string;
  description?: string;
}

export type PlantStage = 'seed' | 'sprout' | 'growing' | 'mature' | 'harvestable' | 'withered';

export interface Plant {
  id: string;
  garden_id: string;
  plant_type_id: string;
  position: number;
  stage: PlantStage;
  growth_progress: number;
  planted_at: string;
  harvested_at?: string;
  created_at: string;
  updated_at: string;
  garden: Garden;
  plant_type: PlantType;
}

export interface PlantType {
  id: string;
  name: string;
  description: string;
  icon: string;
  growth_time: number;
  yield: number;
  harvest_value: number;
  seed_price: number;
}

export interface SeedInventory {
  id: string;
  user_id: string;
  plant_type_id: string;
  quantity: number;
  purchased_at: string;
  plant_type: PlantType;
}

export interface BuySeedRequest {
  plant_type_id: string;
  quantity: number;
}

export interface BuySeedResponse {
  plant_type: PlantType;
  quantity: number;
  total_cost: number;
  user_coins: number;
}

export interface PlantRequest {
  plant_type_id: string;
  position: number;
}

export interface WaterPlantRequest {
  amount: number;
}

export interface FertilizePlantRequest {
  amount: number;
}

export interface Weather {
  current_temp: number;
  humidity: number;
  condition: string;
  season: string;
  timestamp: string;
}

export interface WeatherForecast {
  forecasts: Weather[];
}

export type GardenPermission = 'view' | 'plant' | 'harvest' | 'manage';

export interface GardenShare {
  id: string;
  garden_id: string;
  user_id: string;
  permissions: GardenPermission[];
  shared_by: string;
  shared_at: string;
  expires_at?: string;
  garden: Garden;
  user: User;
  shared_by_user: User;
}

export interface GardenAccessLink {
  id: string;
  garden_id: string;
  token: string;
  permissions: GardenPermission[];
  max_uses?: number;
  used_count: number;
  created_by: string;
  created_at: string;
  expires_at?: string;
  is_active: boolean;
  garden: Garden;
  created_by_user: User;
}

export interface CreateAccessLinkRequest {
  garden_id: string;
  permissions: GardenPermission[];
  max_uses?: number;
  expires_at?: string;
}

export interface JoinGardenRequest {
  token: string;
}

export interface UpdateSharePermissionsRequest {
  user_id: string;
  permissions: GardenPermission[];
}

export interface GardenWithPermission {
  garden: Garden;
  permission: GardenPermission;
}

export interface SharedGardenResponse {
  owned_gardens: Garden[];
  shared_gardens: {
    garden: Garden;
    permission: GardenPermission[];
    shared_by: User;
    shared_at: string;
  }[];
}

// Marketplace Types
export type RequestStatus = 'open' | 'accepted' | 'completed' | 'expired' | 'cancelled';
export type ListingStatus = 'active' | 'sold' | 'expired' | 'cancelled';
export type TransactionType = 'delivery_reward' | 'plant_sale' | 'transfer' | 'commission';
export type TransactionStatus = 'pending' | 'confirmed' | 'failed' | 'expired';

export interface DeliveryRequest {
  id: string;
  requester_id: string;
  plant_type_id: string;
  reward: number;
  deadline: string;
  status: RequestStatus;
  accepted_by?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
  requester: User;
  acceptor?: User;
  plant_type: PlantType;
}

export interface CreateDeliveryRequestRequest {
  plant_type_id: string;
  reward: number;
  deadline: string;
}

export interface AcceptDeliveryRequestRequest {
  signature: string;
}

export interface CompleteDeliveryRequestRequest {
  signature: string;
}

export interface PlantListing {
  id: string;
  seller_id: string;
  plant_id: string;
  price: number;
  status: ListingStatus;
  expires_at: string;
  created_at: string;
  updated_at: string;
  seller: User;
  plant: Plant;
}

export interface CreatePlantListingRequest {
  plant_id: string;
  price: number;
  expires_at: string;
}

export interface BuyPlantListingRequest {
  signature: string;
}

export interface Transaction {
  id: string;
  hash: string;
  sender_key: string;
  receiver_key: string;
  plant_id?: string;
  amount: number;
  tx_type: TransactionType;
  nonce: number;
  signature: string;
  status: TransactionStatus;
  block_height?: number;
  created_at: string;
  updated_at: string;
  plant?: Plant;
}

export interface Block {
  id: string;
  height: number;
  hash: string;
  previous_hash: string;
  transactions: Transaction[];
  timestamp: string;
  created_at: string;
}

export interface Wallet {
  id: string;
  user_id: string;
  public_key: string;
  balance: number;
  reputation: number;
  nonce: number;
  created_at: string;
  updated_at: string;
  user: User;
}

export interface ConsensusNode {
  id: string;
  node_id: string;
  public_key: string;
  is_active: boolean;
  last_seen: string;
  created_at: string;
}

export interface CreateWalletRequest {
  encrypted_private_key: string;
}

export interface TransferRequest {
  receiver_key: string;
  amount: number;
  signature: string;
}

export interface SignRequest {
  data: string;
}

export interface SignResponse {
  signature: string;
}

export interface BlockchainLedger {
  blocks: Block[];
  total_blocks: number;
  total_transactions: number;
  last_block_hash: string;
}

export interface MarketplaceStats {
  total_delivery_requests: number;
  active_delivery_requests: number;
  total_listings: number;
  active_listings: number;
  total_transactions: number;
  total_volume: number;
}