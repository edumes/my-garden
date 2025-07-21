import {
  BuySeedRequest,
  BuySeedResponse,
  CreateAccessLinkRequest,
  GardenAccessLink,
  GardenShare,
  JoinGardenRequest,
  PlantType,
  SeedInventory,
  SharedGardenResponse,
  UpdateSharePermissionsRequest
} from "@/types/api";

const API_BASE_URL = 'http://localhost:8080/api/v1';

class ApiService {
  private getAuthHeaders(): HeadersInit {
    const token = localStorage.getItem('auth_token');
    return {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
    };
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${API_BASE_URL}${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        ...this.getAuthHeaders(),
        ...options.headers,
      },
    });

    if (!response.ok) {
      try {
        const errorData = await response.json();
        throw new Error(errorData.error || `API request failed: ${response.statusText}`);
      } catch {
        throw new Error(`API request failed: ${response.statusText}`);
      }
    }

    return response.json();
  }

  // Authentication
  async login(credentials: { username: string; password: string }) {
    const response = await this.request<any>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
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
      body: JSON.stringify(userData),
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
      body: JSON.stringify(data),
    });
  }

  async updateGarden(id: string, data: { name?: string; description?: string }) {
    return this.request<{ garden: any }>(`/gardens/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteGarden(id: string) {
    return this.request(`/gardens/${id}`, { method: 'DELETE' });
  }

  // Plants
  async plantSeed(gardenId: string, data: { plant_type_id: string; position: number }) {
    return this.request<{ plant: any }>(`/gardens/${gardenId}/plants`, {
      method: 'POST',
      body: JSON.stringify(data),
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
      body: JSON.stringify(request),
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
      body: JSON.stringify(data),
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
      body: JSON.stringify(request),
    });
  }

  async joinGarden(request: JoinGardenRequest) {
    return this.request<{ message: string; garden: any; share: GardenShare }>('/garden-shares/join', {
      method: 'POST',
      body: JSON.stringify(request),
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
      body: JSON.stringify(request),
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
}

export const apiService = new ApiService();