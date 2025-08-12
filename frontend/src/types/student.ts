export interface Student {
  id: number;
  name: string;
  user_id: number;
  business_id: number;
  guardian_name: string;
  guardian_number: string;
  guardian_email: string;
  information: Record<string, any>;
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

export interface CreateStudentRequest {
  name: string;
  user_id: number;
  business_id: number;
  guardian_name?: string;
  guardian_number?: string;
  guardian_email?: string;
  information?: Record<string, any>;
}

export interface UpdateStudentRequest {
  name?: string;
  guardian_name?: string;
  guardian_number?: string;
  guardian_email?: string;
  information?: Record<string, any>;
  status?: number;
}

export interface StudentFilters {
  business_id?: number;
  status?: number;
  guardian_name?: string;
  guardian_email?: string;
  search?: string;
  page?: number;
  limit?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface StudentStats {
  total_students: number;
  active_students: number;
  inactive_students: number;
}

export interface GuardianStats {
  students_with_guardian_email: number;
  students_with_guardian_phone: number;
}

export interface StudentSearchResult {
  students: Student[];
  search_term: string;
  total_found: number;
}

export interface BulkUpdateStudentStatusRequest {
  student_ids: number[];
  status: number;
}

export interface StudentsResponse {
  data: {
    students: Student[];
    total: number;
    page: number;
    limit: number;
  };
}