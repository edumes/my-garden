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
  created_at: string;
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
  name: string;
  description?: string;
  size: number;
  water_level: number;
  soil_quality: number;
  fertilizer_level: number;
  has_sprinkler: boolean;
  has_greenhouse: boolean;
  has_composter: boolean;
  last_watered_at?: string;
  last_fertilized_at?: string;
  created_at: string;
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
  plant_type: PlantType;
}

export interface PlantType {
  id: string;
  name: string;
  description: string;
  icon: string;
  rarity: string;
  season: string;
  weather: string;
  growth_time: number;
  min_level: number;
  yield: number;
  harvest_value: number;
  experience_value: number;
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