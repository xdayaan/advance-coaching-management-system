import { AxiosInstance } from 'axios';
import {
  Teacher,
  CreateTeacherRequest,
  UpdateTeacherRequest,
  TeacherFilters,
  TeacherStats,
  SalaryStats,
  QualificationStats,
  TeacherSearchResult,
  BulkUpdateStatusRequest,
  BulkUpdateSalaryRequest,
  TeachersResponse,
} from '@/types/teacher';

export class TeacherService {
  private api: AxiosInstance;

  constructor(api: AxiosInstance) {
    this.api = api;
  }

  // Basic CRUD operations
  async getTeachers(filters: TeacherFilters = {}): Promise<TeachersResponse> {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        params.append(key, value.toString());
      }
    });

    const response = await this.api.get(`/teachers?${params}`);
    return response.data;
  }

  async getTeacher(id: number): Promise<Teacher> {
    const response = await this.api.get(`/teachers/${id}`);
    return response.data.data;
  }

  async createTeacher(teacherData: CreateTeacherRequest): Promise<Teacher> {
    const response = await this.api.post('/teachers', teacherData);
    return response.data.data;
  }

  async updateTeacher(id: number, teacherData: Partial<UpdateTeacherRequest>): Promise<Teacher> {
    const response = await this.api.put(`/teachers/${id}`, teacherData);
    return response.data.data;
  }

  async deleteTeacher(id: number): Promise<void> {
    await this.api.delete(`/teachers/${id}`);
  }

  // Profile management (for teacher users)
  async getMyTeacherProfile(): Promise<Teacher> {
    const response = await this.api.get('/my-teacher-profile');
    return response.data.data;
  }

  async updateMyTeacherProfile(profileData: Partial<UpdateTeacherRequest>): Promise<Teacher> {
    const response = await this.api.put('/my-teacher-profile', profileData);
    return response.data.data;
  }

  // Business-specific operations
  async getTeachersByBusiness(businessId: number, filters: TeacherFilters = {}): Promise<TeachersResponse> {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        params.append(key, value.toString());
      }
    });

    const response = await this.api.get(`/businesses/${businessId}/teachers?${params}`);
    return response.data;
  }

  async getActiveTeachersByBusiness(businessId: number): Promise<Teacher[]> {
    const response = await this.api.get(`/businesses/${businessId}/teachers/active`);
    return response.data.data;
  }

  async getInactiveTeachersByBusiness(businessId: number): Promise<Teacher[]> {
    const response = await this.api.get(`/businesses/${businessId}/teachers/inactive`);
    return response.data.data;
  }

  // Status operations
  async changeTeacherStatus(id: number, status: number): Promise<void> {
    await this.api.patch(`/teachers/${id}/status`, { status });
  }

  async getActiveTeachers(): Promise<Teacher[]> {
    const response = await this.api.get('/teachers/active');
    return response.data.data;
  }

  async getInactiveTeachers(): Promise<Teacher[]> {
    const response = await this.api.get('/teachers/inactive');
    return response.data.data;
  }

  // Search functionality
  async searchTeachers(query: string, limit: number = 10, businessId?: number): Promise<TeacherSearchResult> {
    const params = new URLSearchParams();
    params.append('q', query);
    params.append('limit', limit.toString());
    if (businessId) {
      params.append('business_id', businessId.toString());
    }

    const response = await this.api.get(`/teachers/search?${params}`);
    return response.data.data;
  }

  // Statistics
  async getTeacherStats(businessId?: number): Promise<TeacherStats> {
    const params = new URLSearchParams();
    if (businessId) {
      params.append('business_id', businessId.toString());
    }

    const response = await this.api.get(`/teachers/stats?${params}`);
    return response.data.data;
  }

  async getSalaryStats(businessId?: number): Promise<SalaryStats> {
    const params = new URLSearchParams();
    if (businessId) {
      params.append('business_id', businessId.toString());
    }

    const response = await this.api.get(`/teachers/stats/salary?${params}`);
    return response.data.data;
  }

  async getQualificationStats(businessId?: number): Promise<QualificationStats> {
    const params = new URLSearchParams();
    if (businessId) {
      params.append('business_id', businessId.toString());
    }

    const response = await this.api.get(`/teachers/stats/qualifications?${params}`);
    return response.data.data;
  }

  // Bulk operations
  async bulkUpdateTeacherStatus(teacherIds: number[], status: number): Promise<void> {
    await this.api.post('/teachers/bulk/status', {
      teacher_ids: teacherIds,
      status,
    });
  }

  async bulkUpdateSalary(teacherIds: number[], salary: number): Promise<void> {
    await this.api.post('/teachers/bulk/salary', {
      teacher_ids: teacherIds,
      salary,
    });
  }

  // Utility methods
  async getTeachersByQualification(qualification: string, businessId?: number): Promise<Teacher[]> {
    const filters: TeacherFilters = {
      qualification,
      limit: 100, // Get all matching
    };
    
    if (businessId) {
      filters.business_id = businessId;
    }

    const response = await this.getTeachers(filters);
    return response.data.teachers;
  }

  async getTeachersBySalaryRange(minSalary?: number, maxSalary?: number, businessId?: number): Promise<Teacher[]> {
    const filters: TeacherFilters = {
      min_salary: minSalary,
      max_salary: maxSalary,
      limit: 100,
    };

    if (businessId) {
      filters.business_id = businessId;
    }

    const response = await this.getTeachers(filters);
    return response.data.teachers;
  }

  async exportTeachers(filters: TeacherFilters = {}, format: 'csv' | 'excel' = 'csv'): Promise<Blob> {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        params.append(key, value.toString());
      }
    });
    params.append('format', format);

    const response = await this.api.get(`/teachers/export?${params}`, {
      responseType: 'blob'
    });
    return response.data;
  }

  async importTeachers(file: File, businessId: number): Promise<{ success: number; errors: string[] }> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('business_id', businessId.toString());

    const response = await this.api.post('/teachers/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // Validation methods
  async validateTeacherUserExists(userId: number, excludeTeacherId?: number): Promise<{ available: boolean }> {
    const params = new URLSearchParams();
    params.append('user_id', userId.toString());
    if (excludeTeacherId) {
      params.append('exclude_id', excludeTeacherId.toString());
    }

    const response = await this.api.get(`/teachers/validate-user?${params}`);
    return response.data;
  }

  // Helper methods for filtering and sorting
  getAvailableSortFields(): Array<{ value: string; label: string }> {
    return [
      { value: 'name', label: 'Name' },
      { value: 'salary', label: 'Salary' },
      { value: 'qualification', label: 'Qualification' },
      { value: 'experience', label: 'Experience' },
      { value: 'status', label: 'Status' },
      { value: 'created_on', label: 'Created Date' },
      { value: 'updated_on', label: 'Updated Date' },
    ];
  }

  getStatusOptions(): Array<{ value: number; label: string; color: string }> {
    return [
      { value: 1, label: 'Active', color: 'green' },
      { value: 0, label: 'Inactive', color: 'red' },
    ];
  }

  formatSalary(amount: number): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'INR',
    }).format(amount);
  }

  getTeacherFullName(teacher: Teacher): string {
    return teacher.name || 'Unknown Teacher';
  }

  getTeacherStatusLabel(status: number): string {
    return status === 1 ? 'Active' : 'Inactive';
  }

  getTeacherStatusColor(status: number): string {
    return status === 1 ? 'green' : 'red';
  }
}