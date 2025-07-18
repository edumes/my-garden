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
      throw new Error(`API request failed: ${response.statusText}`);
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
    return this.request<{ gardens: any[] }>('/gardens');
  }

  async getGarden(id: string) {
    return this.request<{ garden: any }>(`/gardens/${id}`);
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
    return this.request<{ harvest: any }>(`/gardens/${gardenId}/plants/${plantId}/harvest`, {
      method: 'POST',
    });
  }

  async removePlant(gardenId: string, plantId: string) {
    return this.request(`/gardens/${gardenId}/plants/${plantId}`, { method: 'DELETE' });
  }

  // Plant Types
  async getPlantTypes() {
    return this.request<{ plant_types: import('../types/api').PlantType[] }>('/plants');
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
}

export const apiService = new ApiService();