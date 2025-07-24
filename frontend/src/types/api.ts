export interface User {
  id: string;
  username: string;
  email: string;
  first_name?: string;
  last_name?: string;
  avatar?: string;
  level: number;
  experience: number;
  level_progress: number;
  coins: number;
  timezone: string;
  language: string;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
  achievements?: Achievement[];
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

export interface XPResponse {
  xp_earned?: number;
  level?: number;
  level_up?: boolean;
  coins_earned?: number;
}

export interface GardenResponse extends XPResponse {
  garden: Garden;
}

export interface PlantResponse extends XPResponse {
  plant: Plant;
}

export interface HarvestResponse extends XPResponse {
  plant: Plant;
  harvest: {
    coins_earned: number;
    xp_earned: number;
    level: number;
    level_up: boolean;
  };
}

// WebSocket Events
export type WebSocketEventType = 'plant' | 'harvest' | 'weather' | 'water' | 'progress' | 'presence' | 'action_lock' | 'chat';

export interface WebSocketEvent {
  type: WebSocketEventType;
  garden_id: string;
  user_id: string;
  username?: string; // Added for chat messages
  data: WebSocketEventData;
  timestamp: number;
}

export interface WebSocketPresenceEventData {
  users: {
    id: string;
    username: string;
    position?: number; // The plot they're currently interacting with
    action?: 'planting' | 'harvesting' | 'viewing';
  }[];
}

export interface WebSocketActionLockEventData {
  position: number;
  user_id: string;
  username: string;
  action: 'planting' | 'harvesting';
  locked: boolean; // true = lock acquired, false = lock released
}

export interface WebSocketPlantEventData {
  plant: Plant;
  position: number;
  plant_type: PlantType;
}

export interface WebSocketHarvestEventData {
  plant: Plant;
  coins_earned: number;
  xp_earned: number;
}

export interface WebSocketWeatherEventData extends Weather {}

export interface WebSocketProgressEventData {
  plant_id: string;
  progress: number;
  stage: PlantStage;
}

export interface WebSocketChatEventData {
  message: string;
  username: string;
  user_id: string;
  timestamp: number;
}

export type WebSocketEventData =
  | WebSocketPlantEventData
  | WebSocketHarvestEventData
  | WebSocketWeatherEventData
  | WebSocketProgressEventData
  | WebSocketPresenceEventData
  | WebSocketActionLockEventData
  | WebSocketChatEventData;