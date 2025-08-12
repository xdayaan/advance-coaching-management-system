export interface User {
  id: number;
  name: string;
  email: string;
  phone: string;
  role: 'admin' | 'business' | 'teacher' | 'student';
  status: number;
  created_on: string;
  businessSlug?: string;
}

export interface CreateUserRequest {
  name: string;
  email: string;
  phone: string;
  role: 'admin' | 'business' | 'teacher' | 'student';
  password: string;
}

export interface UpdateUserRequest {
  name?: string;
  email?: string;
  phone?: string;
  role?: 'admin' | 'business' | 'teacher' | 'student';
  password?: string;
  status?: number;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  message: string;
  token: string;
  user: User;
}

export interface UsersResponse {
  data: User[];
  pagination: {
    has_next: boolean;
    has_prev: boolean;
    limit: number;
    page: number;
    total: number;
    total_pages: number;
  };
}

export interface UserStats {
  admin: number;
  business: number;
  teacher: number;
  student: number;
  total: number;
}

export interface UsersFilters {
  page?: number;
  limit?: number;
  role?: string;
  status?: number;
  search?: string;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface ApiError {
  message: string;
  [key: string]: any;
}