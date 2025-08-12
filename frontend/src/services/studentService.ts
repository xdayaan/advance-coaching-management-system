import { AxiosInstance } from 'axios';
import {
  Student,
  CreateStudentRequest,
  UpdateStudentRequest,
  StudentFilters,
  StudentStats,
  GuardianStats,
  StudentSearchResult,
  BulkUpdateStudentStatusRequest,
  StudentsResponse,
} from '@/types/student';

export class StudentService {
  private api: AxiosInstance;

  constructor(api: AxiosInstance) {
    this.api = api;
  }

  // Basic CRUD operations
  async getStudents(filters: StudentFilters = {}): Promise<StudentsResponse> {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        params.append(key, value.toString());
      }
    });

    const response = await this.api.get(`/students?${params}`);
    return response.data;
  }

  async getStudent(id: number): Promise<Student> {
    const response = await this.api.get(`/students/${id}`);
    return response.data.data;
  }

  async createStudent(studentData: CreateStudentRequest): Promise<Student> {
    const response = await this.api.post('/students', studentData);
    return response.data.data;
  }

  async updateStudent(id: number, studentData: Partial<UpdateStudentRequest>): Promise<Student> {
    const response = await this.api.put(`/students/${id}`, studentData);
    return response.data.data;
  }

  async deleteStudent(id: number): Promise<void> {
    await this.api.delete(`/students/${id}`);
  }

  // Profile management (for student users)
  async getMyStudentProfile(): Promise<Student> {
    const response = await this.api.get('/my-student-profile');
    return response.data.data;
  }

  async updateMyStudentProfile(profileData: Partial<UpdateStudentRequest>): Promise<Student> {
    const response = await this.api.put('/my-student-profile', profileData);
    return response.data.data;
  }

  // Business-specific operations
  async getStudentsByBusiness(businessId: number, filters: StudentFilters = {}): Promise<StudentsResponse> {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        params.append(key, value.toString());
      }
    });

    const response = await this.api.get(`/businesses/${businessId}/students?${params}`);
    return response.data;
  }

  async getActiveStudentsByBusiness(businessId: number): Promise<Student[]> {
    const response = await this.api.get(`/businesses/${businessId}/students/active`);
    return response.data.data;
  }

  async getInactiveStudentsByBusiness(businessId: number): Promise<Student[]> {
    const response = await this.api.get(`/businesses/${businessId}/students/inactive`);
    return response.data.data;
  }

  // Status operations
  async changeStudentStatus(id: number, status: number): Promise<void> {
    await this.api.patch(`/students/${id}/status`, { status });
  }

  async getActiveStudents(): Promise<Student[]> {
    const response = await this.api.get('/students/active');
    return response.data.data;
  }

  async getInactiveStudents(): Promise<Student[]> {
    const response = await this.api.get('/students/inactive');
    return response.data.data;
  }

  // Search functionality
  async searchStudents(query: string, limit: number = 10, businessId?: number): Promise<StudentSearchResult> {
    const params = new URLSearchParams();
    params.append('q', query);
    params.append('limit', limit.toString());
    if (businessId) {
      params.append('business_id', businessId.toString());
    }

    const response = await this.api.get(`/students/search?${params}`);
    return response.data.data;
  }

  // Statistics
  async getStudentStats(businessId?: number): Promise<StudentStats> {
    const params = new URLSearchParams();
    if (businessId) {
      params.append('business_id', businessId.toString());
    }

    const response = await this.api.get(`/students/stats?${params}`);
    return response.data.data;
  }

  async getGuardianStats(businessId?: number): Promise<GuardianStats> {
    const params = new URLSearchParams();
    if (businessId) {
      params.append('business_id', businessId.toString());
    }

    const response = await this.api.get(`/students/stats/guardians?${params}`);
    return response.data.data;
  }

  // Bulk operations
  async bulkUpdateStudentStatus(studentIds: number[], status: number): Promise<void> {
    await this.api.post('/students/bulk/status', {
      student_ids: studentIds,
      status,
    });
  }

  // Guardian-specific methods
  async getStudentsByGuardianEmail(email: string, businessId?: number): Promise<Student[]> {
    const filters: StudentFilters = {
      guardian_email: email,
      limit: 100,
    };

    if (businessId) {
      filters.business_id = businessId;
    }

    const response = await this.getStudents(filters);
    return response.data.students;
  }

  async getStudentsByGuardianName(name: string, businessId?: number): Promise<Student[]> {
    const filters: StudentFilters = {
      guardian_name: name,
      limit: 100,
    };

    if (businessId) {
      filters.business_id = businessId;
    }

    const response = await this.getStudents(filters);
    return response.data.students;
  }

  // Information management methods
  async updateStudentInformation(id: number, information: Record<string, any>): Promise<Student> {
    return this.updateStudent(id, { information });
  }

  async getStudentInformation(id: number): Promise<Record<string, any>> {
    const student = await this.getStudent(id);
    return student.information || {};
  }

  async addStudentInformationField(id: number, key: string, value: any): Promise<Student> {
    const student = await this.getStudent(id);
    const updatedInfo = { ...student.information, [key]: value };
    return this.updateStudentInformation(id, updatedInfo);
  }

  async removeStudentInformationField(id: number, key: string): Promise<Student> {
    const student = await this.getStudent(id);
    const updatedInfo = { ...student.information };
    delete updatedInfo[key];
    return this.updateStudentInformation(id, updatedInfo);
  }

  // Export/Import methods
  async exportStudents(filters: StudentFilters = {}, format: 'csv' | 'excel' = 'csv'): Promise<Blob> {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        params.append(key, value.toString());
      }
    });
    params.append('format', format);

    const response = await this.api.get(`/students/export?${params}`, {
      responseType: 'blob'
    });
    return response.data;
  }

  async importStudents(file: File, businessId: number): Promise<{ success: number; errors: string[] }> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('business_id', businessId.toString());

    const response = await this.api.post('/students/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // Validation methods
  async validateStudentUserExists(userId: number, excludeStudentId?: number): Promise<{ available: boolean }> {
    const params = new URLSearchParams();
    params.append('user_id', userId.toString());
    if (excludeStudentId) {
      params.append('exclude_id', excludeStudentId.toString());
    }

    const response = await this.api.get(`/students/validate-user?${params}`);
    return response.data;
  }

  async validateGuardianEmail(email: string): Promise<{ valid: boolean }> {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return { valid: emailRegex.test(email) };
  }

  // Helper methods for filtering and sorting
  getAvailableSortFields(): Array<{ value: string; label: string }> {
    return [
      { value: 'name', label: 'Student Name' },
      { value: 'guardian_name', label: 'Guardian Name' },
      { value: 'guardian_email', label: 'Guardian Email' },
      { value: 'guardian_number', label: 'Guardian Number' },
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

  // Utility formatting methods
  formatGuardianContact(student: Student): string {
    const parts = [];
    if (student.guardian_name) parts.push(student.guardian_name);
    if (student.guardian_email) parts.push(student.guardian_email);
    if (student.guardian_number) parts.push(student.guardian_number);
    return parts.join(' • ') || 'No guardian info';
  }

  getStudentFullName(student: Student): string {
    return student.name || 'Unknown Student';
  }

  getStudentStatusLabel(status: number): string {
    return status === 1 ? 'Active' : 'Inactive';
  }

  getStudentStatusColor(status: number): string {
    return status === 1 ? 'green' : 'red';
  }

  hasGuardianContact(student: Student): boolean {
    return !!(student.guardian_email || student.guardian_number);
  }

  hasCompleteGuardianInfo(student: Student): boolean {
    return !!(student.guardian_name && (student.guardian_email || student.guardian_number));
  }

  // Information field management helpers
  getInformationFieldValue(student: Student, fieldKey: string, defaultValue: any = null): any {
    return student.information?.[fieldKey] ?? defaultValue;
  }

  hasInformationField(student: Student, fieldKey: string): boolean {
    return student.information && fieldKey in student.information;
  }

  getInformationFieldKeys(student: Student): string[] {
    return Object.keys(student.information || {});
  }

  // Contact validation helpers
  isValidEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }

  isValidPhoneNumber(phone: string): boolean {
    // Basic phone validation - adjust regex based on your requirements
    const phoneRegex = /^[\+]?[1-9][\d]{0,15}$/;
    return phoneRegex.test(phone.replace(/[\s\-\(\)]/g, ''));
  }

  formatPhoneNumber(phone: string): string {
    // Basic phone formatting - customize based on your locale
    const cleaned = phone.replace(/\D/g, '');
    if (cleaned.length === 10) {
      return `(${cleaned.slice(0, 3)}) ${cleaned.slice(3, 6)}-${cleaned.slice(6)}`;
    }
    return phone;
  }

  // Filter helpers
  getStudentsWithGuardianEmail(students: Student[]): Student[] {
    return students.filter(student => student.guardian_email && student.guardian_email.trim());
  }

  getStudentsWithGuardianPhone(students: Student[]): Student[] {
    return students.filter(student => student.guardian_number && student.guardian_number.trim());
  }

  getStudentsWithIncompleteInfo(students: Student[]): Student[] {
    return students.filter(student => !this.hasCompleteGuardianInfo(student));
  }

  // Grouping helpers
  groupStudentsByStatus(students: Student[]): { active: Student[]; inactive: Student[] } {
    return students.reduce(
      (acc, student) => {
        if (student.status === 1) {
          acc.active.push(student);
        } else {
          acc.inactive.push(student);
        }
        return acc;
      },
      { active: [], inactive: [] } as { active: Student[]; inactive: Student[] }
    );
  }

  groupStudentsByGuardian(students: Student[]): Record<string, Student[]> {
    return students.reduce((acc, student) => {
      const guardianKey = student.guardian_email || student.guardian_name || 'No Guardian';
      if (!acc[guardianKey]) {
        acc[guardianKey] = [];
      }
      acc[guardianKey].push(student);
      return acc;
    }, {} as Record<string, Student[]>);
  }

  // Date helpers
  getStudentAge(student: Student, birthdateField: string = 'birthdate'): number | null {
    const birthdate = this.getInformationFieldValue(student, birthdateField);
    if (!birthdate) return null;

    const birth = new Date(birthdate);
    const today = new Date();
    let age = today.getFullYear() - birth.getFullYear();
    const monthDiff = today.getMonth() - birth.getMonth();

    if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birth.getDate())) {
      age--;
    }

    return age;
  }

  formatStudentCreatedDate(student: Student): string {
    return new Date(student.created_on).toLocaleDateString();
  }

  getStudentsByDateRange(students: Student[], startDate: Date, endDate: Date): Student[] {
    return students.filter(student => {
      const createdDate = new Date(student.created_on);
      return createdDate >= startDate && createdDate <= endDate;
    });
  }
}