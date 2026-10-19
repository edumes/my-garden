import {
  AuditActivityResponse,
  AuditLogsResponse,
  AuditStatsResponse,
  AuthResponse,
  BuySeedRequest,
  BuySeedResponse,
  CreateAccessLinkRequest,
  CreateDeliveryRequestRequest,
  CreatePlantListingRequest,
  CurrentWeatherResponse,
  DeliveryRequest,
  Garden,
  GardenAccessLink,
  GardenPermission,
  GardenShare,
  JoinGardenRequest,
  MarketplaceStats,
  Plant,
  PlantListing,
  PlantType,
  Reward,
  SeedInventory,
  SharedGardenResponse,
  UpdateProfileRequest,
  UpdateSharePermissionsRequest,
  User,
  Weather,
  WeatherForecast
} from "@/types/api";
import axios, { AxiosRequestConfig, AxiosResponse } from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';

class ApiService {
  private axiosInstance;

  constructor() {
    this.axiosInstance = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add request interceptor to include auth token
    this.axiosInstance.interceptors.request.use((config) => {
      const token = localStorage.getItem('auth_token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });

    // Add response interceptor for error handling
    this.axiosInstance.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response) {
          const errorMessage = error.response.data?.error || `API request failed: ${error.response.statusText}`;
          throw new Error(errorMessage);
        } else if (error.request) {
          throw new Error('Network error: No response received');
        } else {
          throw new Error(`Request error: ${error.message}`);
        }
      }
    );
  }

  private async request<T>(endpoint: string, config: AxiosRequestConfig = {}): Promise<T> {
    const response: AxiosResponse<T> = await this.axiosInstance.request({
      url: endpoint,
      ...config,
    });
    return response.data;
  }

  // Authentication
  async login(credentials: { username: string; password: string }) {
    const response = await this.request<AuthResponse>('/auth/login', {
      method: 'POST',
      data: credentials,
    });

    if (response.token) {
      localStorage.setItem('auth_token', response.token);
    }

    return response;
  }

  async register(userData: {
    username: string;
    email: string;
    password: string;
    first_name?: string;
    last_name?: string;
  }) {
    const response = await this.request<AuthResponse>('/auth/register', {
      method: 'POST',
      data: userData,
    });

    if (response.token) {
      localStorage.setItem('auth_token', response.token);
    }

    return response;
  }

  // Revokes every session of the user on the server. The local token is
  // dropped even when that fails, for example because it was already revoked.
  async logout() {
    try {
      await this.request('/auth/logout', { method: 'POST' });
    } finally {
      localStorage.removeItem('auth_token');
    }
  }

  // Gardens
  async getGardens() {
    return this.request<SharedGardenResponse>('/gardens');
  }

  async getGarden(id: string) {
    return this.request<{ garden: Garden; is_owner: boolean; permissions: GardenPermission[] }>(`/gardens/${id}`);
  }

  async createGarden(data: { name: string; description?: string }) {
    return this.request<{ garden: Garden }>('/gardens', {
      method: 'POST',
      data: data,
    });
  }

  async updateGarden(id: string, data: { name?: string; description?: string }) {
    return this.request<{ garden: Garden }>(`/gardens/${id}`, {
      method: 'PUT',
      data: data,
    });
  }

  async deleteGarden(id: string) {
    return this.request(`/gardens/${id}`, { method: 'DELETE' });
  }

  // Plants
  async plantSeed(gardenId: string, data: { plant_type_id: string; position: number }) {
    return this.request<{ plant: Plant; reward: Reward }>(`/gardens/${gardenId}/plants`, {
      method: 'POST',
      data: data,
    });
  }

  // Harvesting removes the plant; its grower receives the reward.
  async harvestPlant(gardenId: string, plantId: string) {
    return this.request<{ reward: Reward }>(`/gardens/${gardenId}/plants/${plantId}/harvest`, {
      method: 'POST',
    });
  }

  async removePlant(gardenId: string, plantId: string) {
    return this.request(`/gardens/${gardenId}/plants/${plantId}`, { method: 'DELETE' });
  }

  // Plant Types
  async getPlantTypes() {
    return this.request<{ plant_types: PlantType[] }>('/plants');
  }

  // Store
  async getStoreInventory() {
    return this.request<{ inventory: PlantType[] }>('/store/inventory');
  }

  async buySeed(request: BuySeedRequest) {
    return this.request<BuySeedResponse>('/store/buy', {
      method: 'POST',
      data: request,
    });
  }

  async getUserSeedInventory() {
    return this.request<{ seed_inventory: SeedInventory[] }>('/store/inventory/user');
  }

  // User Profile
  async getUserProfile() {
    return this.request<{ user: User }>('/users/profile');
  }

  async updateUserProfile(data: UpdateProfileRequest) {
    return this.request<{ user: User }>('/users/profile', {
      method: 'PUT',
      data: data,
    });
  }

  // Weather
  async getCurrentWeather() {
    return this.request<CurrentWeatherResponse>('/weather/current');
  }

  async getWeatherForecast() {
    return this.request<{ forecasts: WeatherForecast[]; generated_at: string }>('/weather/forecast');
  }

  async getWeatherHistory() {
    return this.request<{ history: Weather[]; count: number }>('/weather/history');
  }

  // Garden Sharing
  async createAccessLink(request: CreateAccessLinkRequest) {
    return this.request<{ access_link: GardenAccessLink; share_url: string }>('/garden-shares/access-links', {
      method: 'POST',
      data: request,
    });
  }

  async joinGarden(request: JoinGardenRequest) {
    return this.request<{ message: string; garden: Garden; share: GardenShare }>('/garden-shares/join', {
      method: 'POST',
      data: request,
    });
  }

  async getSharedGardens() {
    return this.request<{ shared_gardens: GardenShare[] }>('/garden-shares/shared-with-me');
  }

  async getGardenShares(gardenId: string) {
    return this.request<{ garden_shares: GardenShare[] }>(`/garden-shares/garden/${gardenId}`);
  }

  async updateSharePermissions(gardenId: string, request: UpdateSharePermissionsRequest) {
    return this.request<{ garden_share: GardenShare }>(`/garden-shares/garden/${gardenId}/permissions`, {
      method: 'PUT',
      data: request,
    });
  }

  async removeGardenShare(gardenId: string, userId: string) {
    return this.request(`/garden-shares/garden/${gardenId}/user/${userId}`, { method: 'DELETE' });
  }

  async getAccessLinks(gardenId: string) {
    return this.request<{ access_links: GardenAccessLink[] }>(`/garden-shares/garden/${gardenId}/access-links`);
  }

  async deactivateAccessLink(gardenId: string, linkId: string) {
    return this.request(`/garden-shares/garden/${gardenId}/access-links/${linkId}/deactivate`, { method: 'POST' });
  }

  // Audit
  async getAuditLogs(params?: {
    user_id?: string;
    action?: string;
    resource?: string;
    resource_id?: string;
    status?: string;
    start_date?: string;
    end_date?: string;
    limit?: number;
    offset?: number;
  }) {
    const queryParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          queryParams.append(key, value.toString());
        }
      });
    }
    return this.request<AuditLogsResponse>(`/audit/logs?${queryParams.toString()}`);
  }

  async getUserActivity(userId: string, limit?: number) {
    const queryParams = limit ? `?limit=${limit}` : '';
    return this.request<AuditActivityResponse>(`/audit/users/${userId}/activity${queryParams}`);
  }

  async getGardenActivity(gardenId: string, limit?: number) {
    const queryParams = limit ? `?limit=${limit}` : '';
    return this.request<AuditActivityResponse>(`/audit/gardens/${gardenId}/activity${queryParams}`);
  }

  async getAuditStats(days?: number) {
    const queryParams = days ? `?days=${days}` : '';
    return this.request<AuditStatsResponse>(`/audit/stats${queryParams}`);
  }

  // Marketplace
  async createDeliveryRequest(request: CreateDeliveryRequestRequest) {
    return this.request<DeliveryRequest>('/marketplace/delivery-requests', {
      method: 'POST',
      data: request,
    });
  }

  async getDeliveryRequests() {
    return this.request<{ delivery_requests: DeliveryRequest[]; total: number }>('/marketplace/delivery-requests');
  }

  async acceptDeliveryRequest(id: string) {
    return this.request<{ message: string }>(`/marketplace/delivery-requests/${id}/accept`, { method: 'POST' });
  }

  async completeDeliveryRequest(id: string) {
    return this.request<{ message: string; reward_paid: number }>(`/marketplace/delivery-requests/${id}/complete`, {
      method: 'POST',
    });
  }

  async createPlantListing(request: CreatePlantListingRequest) {
    return this.request<PlantListing>('/marketplace/listings', {
      method: 'POST',
      data: request,
    });
  }

  async getPlantListings(params?: { plant_type_id?: string; min_price?: number; max_price?: number }) {
    const queryParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          queryParams.append(key, value.toString());
        }
      });
    }
    return this.request<{ listings: PlantListing[]; total: number }>(`/marketplace/listings?${queryParams.toString()}`);
  }

  async buyPlantListing(id: string) {
    return this.request<{ message: string }>(`/marketplace/listings/${id}/buy`, { method: 'POST' });
  }

  async getMarketplaceStats() {
    return this.request<MarketplaceStats>('/marketplace/stats');
  }
}

export const apiService = new ApiService();