export interface User {
  id: string;
  username: string;
  email: string;
  first_name?: string;
  last_name?: string;
  level: number;
  experience: number;
  // Experience at which the current level started and the next one starts.
  level_experience: number;
  next_level_experience: number;
  garden_limit: number;
  next_garden_level: number;
  coins: number;
  avatar?: string;
  timezone?: string;
  language?: string;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
  achievements?: UserAchievement[];
}

// The part of another player's profile the API exposes.
export interface PublicUser {
  id: string;
  username: string;
  avatar: string;
  level: number;
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

export interface UpdateProfileRequest {
  first_name?: string;
  last_name?: string;
  avatar?: string;
  timezone?: string;
  language?: string;
}

export interface Garden {
  id: string;
  user_id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
  user: PublicUser;
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

export type PlantStage = 'seed' | 'sprout' | 'growing' | 'mature' | 'harvestable';

export interface Plant {
  id: string;
  garden_id: string;
  plant_type_id: string;
  position: number;
  // Earns the harvest: whoever planted the seed or bought the plant.
  grower_id: string;
  stage: PlantStage;
  growth_progress: number;
  planted_at: string;
  grown_at: string;
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
  unlock_level: number;
}

// What an action paid and to whom.
export interface Reward {
  user_id: string;
  coins: number;
  experience: number;
  level: number;
  level_up: boolean;
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
  id: string;
  condition: string;
  temperature: number;
  humidity: number;
  wind_speed: number;
  pressure: number;
  growth_multiplier: number;
  water_evaporation_rate: number;
  created_at: string;
  valid_until: string;
}

export interface CurrentWeatherResponse {
  weather: Weather;
  season: string;
  updated_at: string;
}

export interface WeatherForecast {
  id: string;
  condition: string;
  temperature: number;
  humidity: number;
  probability: number;
  forecast_for: string;
  created_at: string;
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
  user: PublicUser;
  shared_by_user: PublicUser;
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
  created_by_user: PublicUser;
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
    shared_by: PublicUser;
    shared_at: string;
  }[];
}

// Marketplace Types
export type RequestStatus = 'open' | 'accepted' | 'completed';
export type ListingStatus = 'active' | 'sold';

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
  requester: PublicUser;
  acceptor?: PublicUser;
  plant_type: PlantType;
}

export interface CreateDeliveryRequestRequest {
  plant_type_id: string;
  reward: number;
  deadline: string;
}

export interface PlantListing {
  id: string;
  seller_id: string;
  plant_id: string;
  price: number;
  status: ListingStatus;
  buyer_id?: string;
  expires_at: string;
  created_at: string;
  updated_at: string;
  seller: PublicUser;
  plant: Plant;
}

export interface CreatePlantListingRequest {
  plant_id: string;
  price: number;
}

export interface MarketplaceStats {
  total_delivery_requests: number;
  active_delivery_requests: number;
  total_listings: number;
  active_listings: number;
  total_transactions: number;
  total_volume: number;
}

// Audit Types
export interface AuditLog {
  id: string;
  user_id: string | null;
  action: string;
  resource: string;
  resource_id: string | null;
  ip_address: string;
  status: string;
  created_at: string;
}

export interface AuditLogsResponse {
  logs: AuditLog[];
  total: number;
  limit: number;
  offset: number;
  has_more: boolean;
}

export interface AuditActivityResponse {
  activity: AuditLog[];
  count: number;
}

export interface AuditStats {
  total_actions: number;
  successful_actions: number;
  failed_actions: number;
  error_actions: number;
  top_actions: { action: string; count: number }[];
  top_resources: { resource: string; count: number }[];
}

export interface AuditStatsResponse {
  period_days: number;
  start_date: string;
  end_date: string;
  stats: AuditStats;
}
