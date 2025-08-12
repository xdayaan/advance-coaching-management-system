import axios, { AxiosInstance } from 'axios';
import { UserService } from './userService';
import { PackageService } from './packageService';
import { BusinessService } from './businessService';
import { StudentService } from './studentService';
import { TeacherService } from './teacherService';

type UserRole = 'admin' | 'business' | 'teacher' | 'student';

interface User {
  role: UserRole;
  businessSlug?: string;
  [key: string]: any;
}

class APIService {
  private api: AxiosInstance;
  public users: UserService;
  public packages: PackageService;
  public businesses: BusinessService;
  public teachers: TeacherService;
  public students: StudentService;

  constructor() {
    this.api = axios.create({
      baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api',
      timeout: 10000,
    });

    // Initialize service modules
    this.users = new UserService(this.api);
    this.packages = new PackageService(this.api);
    this.businesses = new BusinessService(this.api);
    this.students = new StudentService(this.api);
    this.teachers = new TeacherService(this.api);

    this.setupInterceptors();
  }

  private setupInterceptors() {
    // Request interceptor
    this.api.interceptors.request.use(
      (config) => {
        // Only access localStorage in browser environment
        if (typeof window !== 'undefined') {
          const token = localStorage.getItem('token');
          if (token) {
            config.headers.Authorization = `Bearer ${token}`;
          }
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Response interceptor
    this.api.interceptors.response.use(
      (response) => response,
      async (error) => {
        // Only handle client-side redirects in browser environment
        if (typeof window !== 'undefined' && error.response?.status === 401) {
          localStorage.removeItem('token');
          localStorage.removeItem('user');
          
          // Use Next.js router for better navigation
          const { default: Router } = await import('next/router');
          Router.push('/dashboard/login');
        }
        return Promise.reject(error);
      }
    );
  }


  // Auth methods
  async login(data: any) {
    try {
      // Ensure email and password are sent as plain strings, not nested objects
      const response = await this.api.post('/login', { 
        email: String(data.email), 
        password: String(data.password) 
      });
      
      const { token, user } = response.data;
      
      if (token && user) {
        localStorage.setItem('token', token);
        localStorage.setItem('user', JSON.stringify(user));
        
        // Get the appropriate dashboard route based on user role

      }
      
      return { token, user };
    } catch (error) {
      throw error;
    }
  }

  async register(userData: any) {
    try {
      const response = await this.api.post('/register', userData);
      const { token, user } = response.data;
      
      if (token && user) {
        localStorage.setItem('token', token);
        localStorage.setItem('user', JSON.stringify(user));
        

        
        // Redirect to role-specific dashboard
        if (typeof window !== 'undefined') {
          const { default: Router } = await import('next/router');

        }
      }
      
      return { token, user };
    } catch (error) {
      throw error;
    }
  }

  async logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    
    // Redirect to login page
    if (typeof window !== 'undefined') {
      const { default: Router } = await import('next/router');
      Router.push('/dashboard/login');
    }
  }

  async getProfile() {
    const response = await this.api.get('/profile');
    return response.data?.data;
  }

  async updateProfile(profileData: any) {
    const response = await this.api.put('/profile', profileData);
    return response.data?.data;
  }

  // Utility methods
  isAuthenticated(): boolean {
    if (typeof window === 'undefined') return false;
    return Boolean(localStorage.getItem('token'));
  }

  getCurrentUser(): User | null {
    if (typeof window === 'undefined') return null;
    
    const user = localStorage.getItem('user');
    try {
      return user ? JSON.parse(user) : null;
    } catch {
      return null;
    }
  }

  getCurrentUserRole(): UserRole | null {
    const user = this.getCurrentUser();
    return user?.role || null;
  }

  getDashboardRouteForCurrentUser() {
    const user = this.getCurrentUser();
    if (!user) return '/dashboard/login';

  }

  // Role checking utilities
  isAdmin(): boolean {
    return this.getCurrentUserRole() === 'admin';
  }

  isBusiness(): boolean {
    return this.getCurrentUserRole() === 'business';
  }

  isTeacher(): boolean {
    return this.getCurrentUserRole() === 'teacher';
  }

  isStudent(): boolean {
    return this.getCurrentUserRole() === 'student';
  }

  // Route protection utility
  canAccessRoute(route: string): boolean {
    const user = this.getCurrentUser();
    if (!user) return false;

    const userRole = user.role;
    const businessSlug = user.businessSlug;

    // Admin can access admin routes
    if (route.startsWith('/dashboard/admin')) {
      return userRole === 'admin';
    }

    // Business routes
    if (route.includes('/business/')) {
      if (userRole === 'admin') return true; // Admin can access all business routes
      
      // Extract business slug from route
      const routeBusinessSlug = route.split('/business/')[1]?.split('/')[0];
      
      // Check if user belongs to this business
      if (businessSlug !== routeBusinessSlug) return false;
      
      // Check role-specific access
      if (route.includes('/teacher/dashboard')) {
        return userRole === 'teacher' || userRole === 'business';
      }
      
      if (route.includes('/student/dashboard')) {
        return userRole === 'student' || userRole === 'business';
      }
      
      // Business dashboard access
      if (route.endsWith('/dashboard')) {
        return userRole === 'business';
      }
    }

    return false;
  }
}

export const apiService = new APIService();