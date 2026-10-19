import {
  AcceptDeliveryRequestRequest,
  BlockchainLedger,
  BuyPlantListingRequest,
  BuySeedRequest,
  BuySeedResponse,
  CompleteDeliveryRequestRequest,
  CreateAccessLinkRequest,
  CreateDeliveryRequestRequest,
  CreatePlantListingRequest,
  CreateWalletRequest,
  DeliveryRequest,
  GardenAccessLink,
  GardenShare,
  JoinGardenRequest,
  MarketplaceStats,
  PlantListing,
  PlantType,
  SeedInventory,
  SharedGardenResponse,
  SignRequest,
  SignResponse,
  Transaction,
  TransferRequest,
  UpdateSharePermissionsRequest,
  Wallet
} from "@/types/api";
import axios, { AxiosRequestConfig, AxiosResponse } from 'axios';

const API_BASE_URL = 'http://localhost:8080/api/v1';

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
    const response = await this.request<any>('/auth/login', {
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
    const response = await this.request<any>('/auth/register', {
      method: 'POST',
      data: userData,
    });

    if (response.token) {
      localStorage.setItem('auth_token', response.token);
    }

    return response;
  }

  async logout() {
    await this.request('/auth/logout', { method: 'POST' });
    localStorage.removeItem('auth_token');
  }

  // Gardens
  async getGardens() {
    return this.request<SharedGardenResponse>('/gardens');
  }

  async getGarden(id: string) {
    return this.request<{ garden: any; permission?: string }>(`/gardens/${id}`);
  }

  async createGarden(data: { name: string; description?: string }) {
    return this.request<{ garden: any }>('/gardens', {
      method: 'POST',
      data: data,
    });
  }

  async updateGarden(id: string, data: { name?: string; description?: string }) {
    return this.request<{ garden: any }>(`/gardens/${id}`, {
      method: 'PUT',
      data: data,
    });
  }

  async deleteGarden(id: string) {
    return this.request(`/gardens/${id}`, { method: 'DELETE' });
  }

  // Plants
  async plantSeed(gardenId: string, data: { plant_type_id: string; position: number }) {
    return this.request<{ plant: any }>(`/gardens/${gardenId}/plants`, {
      method: 'POST',
      data: data,
    });
  }

  async harvestPlant(gardenId: string, plantId: string) {
    return this.request<{
      plant: any;
      harvest: {
        coins_earned: number;
        level_up: boolean;
        new_level: number;
      }
    }>(`/gardens/${gardenId}/plants/${plantId}/harvest`, {
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
    return this.request<{ user: any }>('/users/profile');
  }

  async updateUserProfile(data: any) {
    return this.request<{ user: any }>('/users/profile', {
      method: 'PUT',
      data: data,
    });
  }

  // Weather
  async getCurrentWeather() {
    return this.request<{ weather: any }>('/weather/current');
  }

  async getWeatherForecast() {
    return this.request<{ forecasts: any[] }>('/weather/forecast');
  }

  async getWeatherHistory() {
    return this.request<{ history: any[] }>('/weather/history');
  }

  // Garden Sharing
  async createAccessLink(request: CreateAccessLinkRequest) {
    return this.request<{ access_link: GardenAccessLink; share_url: string }>('/garden-shares/access-links', {
      method: 'POST',
      data: request,
    });
  }

  async joinGarden(request: JoinGardenRequest) {
    return this.request<{ message: string; garden: any; share: GardenShare }>('/garden-shares/join', {
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
    return this.request(`/audit/logs?${queryParams.toString()}`);
  }

  async getUserActivity(userId: string, limit?: number) {
    const queryParams = limit ? `?limit=${limit}` : '';
    return this.request(`/audit/users/${userId}/activity${queryParams}`);
  }

  async getGardenActivity(gardenId: string, limit?: number) {
    const queryParams = limit ? `?limit=${limit}` : '';
    return this.request(`/audit/gardens/${gardenId}/activity${queryParams}`);
  }

  async getAuditStats(days?: number) {
    const queryParams = days ? `?days=${days}` : '';
    return this.request(`/audit/stats${queryParams}`);
  }

  // Marketplace API Methods
  async createDeliveryRequest(request: CreateDeliveryRequestRequest) {
    return this.request<{ delivery_request: DeliveryRequest }>('/marketplace/delivery-requests', {
      method: 'POST',
      data: request,
    });
  }

  async getDeliveryRequests(params?: {
    status?: string;
    plant_type_id?: string;
    requester_id?: string;
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
    return this.request<{ delivery_requests: DeliveryRequest[]; total: number }>(`/marketplace/delivery-requests?${queryParams.toString()}`);
  }

  async acceptDeliveryRequest(id: string, request: AcceptDeliveryRequestRequest) {
    return this.request<{ delivery_request: DeliveryRequest; transaction: Transaction }>(`/marketplace/delivery-requests/${id}/accept`, {
      method: 'POST',
      data: request,
    });
  }

  async completeDeliveryRequest(id: string, request: CompleteDeliveryRequestRequest) {
    return this.request<{ delivery_request: DeliveryRequest; transaction: Transaction }>(`/marketplace/delivery-requests/${id}/complete`, {
      method: 'POST',
      data: request,
    });
  }

  async createPlantListing(request: CreatePlantListingRequest) {
    return this.request<{ listing: PlantListing }>('/marketplace/listings', {
      method: 'POST',
      data: request,
    });
  }

  async getPlantListings(params?: {
    status?: string;
    plant_type_id?: string;
    seller_id?: string;
    min_price?: number;
    max_price?: number;
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
    return this.request<{ listings: PlantListing[]; total: number }>(`/marketplace/listings?${queryParams.toString()}`);
  }

  async buyPlantListing(id: string, request: BuyPlantListingRequest) {
    return this.request<{ listing: PlantListing; transaction: Transaction }>(`/marketplace/listings/${id}/buy`, {
      method: 'POST',
      data: request,
    });
  }

  // Blockchain API Methods
  async submitTransaction(transaction: {
    sender_key: string;
    receiver_key: string;
    plant_id?: string;
    amount: number;
    tx_type: string;
    signature: string;
  }) {
    return this.request<{ transaction: Transaction }>('/blockchain/transactions', {
      method: 'POST',
      data: transaction,
    });
  }

  async getTransactionStatus(hash: string) {
    return this.request<{ transaction: Transaction }>(`/blockchain/transactions/${hash}`);
  }

  async getBlockchainLedger(params?: {
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
    return this.request<BlockchainLedger>(`/blockchain/ledger?${queryParams.toString()}`);
  }

  // Wallet API Methods
  async createWallet(request: CreateWalletRequest) {
    return this.request<{ wallet: Wallet }>('/wallet', {
      method: 'POST',
      data: request,
    });
  }

  async getWalletBalance() {
    return this.request<{ wallet: Wallet }>('/wallet/balance');
  }

  async transferCurrency(request: TransferRequest) {
    return this.request<{ transaction: Transaction }>('/wallet/transfer', {
      method: 'POST',
      data: request,
    });
  }

  async signData(request: SignRequest) {
    return this.request<SignResponse>('/wallet/sign', {
      method: 'POST',
      data: request,
    });
  }

  // Marketplace Stats
  async getMarketplaceStats() {
    return this.request<MarketplaceStats>('/marketplace/stats');
  }
}

export const apiService = new ApiService();