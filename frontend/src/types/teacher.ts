export interface Teacher {
  id: number;
  name: string;
  user_id: number;
  business_id: number;
  salary: number;
  qualification: string;
  experience: string;
  description: string;
  status: number;
  created_on: string;
  updated_on: string;
  user?: {
    id: number;
    username: string;
    email: string;
    user_type: string;
    status: number;
    created_on: string;
  };
  business?: {
    id: number;
    name: string;
    slug: string;
    user_id: number;
    owner_name: string;
    email: string;
    phone: string;
    location: string;
    status: number;
    created_on: string;
  };
}

export interface CreateTeacherRequest {
  name: string;
  user_id: number;
  business_id: number;
  salary?: number;
  qualification?: string;
  experience?: string;
  description?: string;
}

export interface UpdateTeacherRequest {
  name?: string;
  salary?: number;
  qualification?: string;
  experience?: string;
  description?: string;
  status?: number;
}

export interface TeacherFilters {
  business_id?: number;
  status?: number;
  min_salary?: number;
  max_salary?: number;
  qualification?: string;
  search?: string;
  page?: number;
  limit?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface TeacherStats {
  total_teachers: number;
  active_teachers: number;
  inactive_teachers: number;
}

export interface SalaryStats {
  min_salary: number;
  max_salary: number;
  avg_salary: number;
}

export interface QualificationStats {
  [qualification: string]: number;
}

export interface TeacherSearchResult {
  teachers: Teacher[];
  search_term: string;
  total_found: number;
}

export interface BulkUpdateStatusRequest {
  teacher_ids: number[];
  status: number;
}

export interface BulkUpdateSalaryRequest {
  teacher_ids: number[];
  salary: number;
}

export interface TeachersResponse {
  data: {
    teachers: Teacher[];
    total: number;
    page: number;
    limit: number;
  };
}