import { AxiosInstance } from 'axios';
import { User, CreateUserRequest, UpdateUserRequest, UsersFilters, UserStats } from '@/types/user';

export class UserService {
  private api: AxiosInstance;

  constructor(api: AxiosInstance) {
    this.api = api;
  }

  async getUsers(filters: UsersFilters = {}): Promise<{
    data: User[];
    pagination: any;
  }> {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        params.append(key, value.toString());
      }
    });

    const response = await this.api.get(`/users?${params}`);
    return response.data;
  }

  async getUser(id: number): Promise<User> {
    const response = await this.api.get(`/users/${id}`);
    return response.data.data;
  }

  async createUser(userData: CreateUserRequest): Promise<User> {
    const response = await this.api.post('/users', userData);
    return response.data.data;
  }

  async updateUser(id: number, userData: Partial<UpdateUserRequest>): Promise<User> {
    const response = await this.api.put(`/users/${id}`, userData);
    return response.data.data;
  }

  async deleteUser(id: number): Promise<void> {
    await this.api.delete(`/users/${id}`);
  }

  async getUsersByRole(role: string): Promise<User[]> {
    const response = await this.api.get(`/users/role/${role}`);
    return response.data.data;
  }

  async getUserStats(): Promise<UserStats> {
    const response = await this.api.get('/users/stats/roles');
    return response.data.data;
  }

  async promoteUser(id: number, role: string): Promise<void> {
    await this.api.post(`/users/${id}/promote`, { role });
  }

  async changeUserStatus(id: number, status: number): Promise<void> {
    await this.api.patch(`/users/${id}/status`, { status });
  }

  async bulkUpdateUserStatus(userIds: number[], status: number): Promise<void> {
    await this.api.patch('/users/bulk/status', {
      user_ids: userIds,
      status,
    });
  }

  async searchUsers(query: string, limit: number = 10): Promise<User[]> {
    const response = await this.api.get(`/users/search?q=${encodeURIComponent(query)}&limit=${limit}`);
    return response.data.data;
  }
}