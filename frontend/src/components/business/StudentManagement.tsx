'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Textarea } from '@/components/ui/textarea';
import { toast } from 'sonner';
import { apiService } from '@/services/api';
import { Student, StudentFilters, CreateStudentRequest, StudentStats, GuardianStats } from '@/types/student';
import { User } from '@/types/user';
import { Plus, Trash2, Search, Edit, ToggleLeft, ToggleRight, Users, Phone, Mail, Download, Upload } from 'lucide-react';

// Define pagination interface
interface Pagination {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
  has_prev: boolean;
  has_next: boolean;
}

interface StudentManagementProps {
  businessId: number;
}

interface NewInfoField {
  key: string;
  value: string;
}

interface EditStudentData {
  name: string;
  guardian_name: string;
  guardian_number: string;
  guardian_email: string;
  information: Record<string, any>;
  status: number;
}

export default function StudentManagement({ businessId }: StudentManagementProps) {
  const [students, setStudents] = useState<Student[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [filters, setFilters] = useState<StudentFilters>({ business_id: businessId, page: 1, limit: 10 });
  const [pagination, setPagination] = useState<Pagination | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState<boolean>(false);
  const [editDialogOpen, setEditDialogOpen] = useState<boolean>(false);
  const [editingStudent, setEditingStudent] = useState<Student | null>(null);
  const [stats, setStats] = useState<StudentStats | null>(null);
  const [guardianStats, setGuardianStats] = useState<GuardianStats | null>(null);
  const [searchQuery, setSearchQuery] = useState<string>('');
  
  const [newStudent, setNewStudent] = useState<CreateStudentRequest>({
    name: '',
    user_id: 0,
    business_id: businessId,
    guardian_name: '',
    guardian_number: '',
    guardian_email: '',
    information: {},
  });
  
  const [editStudent, setEditStudent] = useState<EditStudentData>({
    name: '',
    guardian_name: '',
    guardian_number: '',
    guardian_email: '',
    information: {},
    status: 1,
  });

  // Additional information fields management
  const [newInfoField, setNewInfoField] = useState<NewInfoField>({ key: '', value: '' });
  const [editInfoField, setEditInfoField] = useState<NewInfoField>({ key: '', value: '' });

  const fetchStudents = async (): Promise<void> => {
    try {
      setLoading(true);
      const response = await apiService.students.getStudentsByBusiness(businessId, filters);
      setStudents(response.data.students);
      setPagination({
        page: response.data.page,
        limit: response.data.limit,
        total: response.data.total,
        total_pages: Math.ceil(response.data.total / response.data.limit),
        has_prev: response.data.page > 1,
        has_next: response.data.page < Math.ceil(response.data.total / response.data.limit),
      });
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to fetch students');
    } finally {
      setLoading(false);
    }
  };

  const fetchUsers = async (): Promise<void> => {
    try {
      // Fetch available users that can be students (not already assigned)
      const response = await apiService.users.getUsers({ role: 'student', status: 1, limit: 100 });
      setUsers(response.data);
    } catch (error: any) {
      console.error('Failed to fetch users:', error);
    }
  };

  const fetchStats = async (): Promise<void> => {
    try {
      const [studentStats, guardianStats] = await Promise.all([
        apiService.students.getStudentStats(businessId),
        apiService.students.getGuardianStats(businessId)
      ]);
      setStats(studentStats);
      setGuardianStats(guardianStats);
    } catch (error) {
      console.error('Failed to fetch student stats:', error);
    }
  };

  useEffect(() => {
    fetchStudents();
    fetchStats();
  }, [filters]);

  useEffect(() => {
    fetchUsers();
  }, []);

  const handleCreateStudent = async (e: React.FormEvent<HTMLFormElement>): Promise<void> => {
    e.preventDefault();
    
    // Validate guardian email if provided
    if (newStudent.guardian_email && !apiService.students.isValidEmail(newStudent.guardian_email)) {
      toast.error('Please enter a valid guardian email address');
      return;
    }

    // Validate guardian phone if provided
    if (newStudent.guardian_number && !apiService.students.isValidPhoneNumber(newStudent.guardian_number)) {
      toast.error('Please enter a valid guardian phone number');
      return;
    }

    try {
      await apiService.students.createStudent(newStudent);
      toast.success('Student created successfully');
      setCreateDialogOpen(false);
      setNewStudent({
        name: '',
        user_id: 0,
        business_id: businessId,
        guardian_name: '',
        guardian_number: '',
        guardian_email: '',
        information: {},
      });
      await fetchStudents();
      await fetchStats();
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to create student');
    }
  };

  const handleEditStudent = (student: Student): void => {
    setEditingStudent(student);
    setEditStudent({
      name: student.name,
      guardian_name: student.guardian_name,
      guardian_number: student.guardian_number,
      guardian_email: student.guardian_email,
      information: student.information || {},
      status: student.status,
    });
    setEditDialogOpen(true);
  };

  const handleUpdateStudent = async (e: React.FormEvent<HTMLFormElement>): Promise<void> => {
    e.preventDefault();
    if (!editingStudent) return;

    // Validate guardian email if provided
    if (editStudent.guardian_email && !apiService.students.isValidEmail(editStudent.guardian_email)) {
      toast.error('Please enter a valid guardian email address');
      return;
    }

    // Validate guardian phone if provided
    if (editStudent.guardian_number && !apiService.students.isValidPhoneNumber(editStudent.guardian_number)) {
      toast.error('Please enter a valid guardian phone number');
      return;
    }

    try {
      await apiService.students.updateStudent(editingStudent.id, editStudent);
      toast.success('Student updated successfully');
      setEditDialogOpen(false);
      setEditingStudent(null);
      setEditStudent({
        name: '',
        guardian_name: '',
        guardian_number: '',
        guardian_email: '',
        information: {},
        status: 1,
      });
      await fetchStudents();
      await fetchStats();
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to update student');
    }
  };

  const handleDeleteStudent = async (id: number): Promise<void> => {
    if (window.confirm('Are you sure you want to delete this student?')) {
      try {
        await apiService.students.deleteStudent(id);
        toast.success('Student deleted successfully');
        await fetchStudents();
        await fetchStats();
      } catch (error: any) {
        toast.error(error.response?.data?.message || 'Failed to delete student');
      }
    }
  };

  const handleToggleStatus = async (id: number, currentStatus: number): Promise<void> => {
    try {
      const newStatus = currentStatus === 1 ? 0 : 1;
      await apiService.students.changeStudentStatus(id, newStatus);
      toast.success(`Student ${newStatus === 1 ? 'activated' : 'deactivated'} successfully`);
      await fetchStudents();
      await fetchStats();
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to change student status');
    }
  };

  const handleSearch = (search: string): void => {
    setSearchQuery(search);
    setFilters({ ...filters, search, page: 1 });
  };

  const handleStatusFilter = (status: string): void => {
    let statusValue: number | undefined;
    if (status === 'active') statusValue = 1;
    else if (status === 'inactive') statusValue = 0;
    else statusValue = undefined;
    setFilters({ ...filters, status: statusValue, page: 1 });
  };

  const handleSortChange = (sortBy: string): void => {
    setFilters({ ...filters, sort_by: sortBy, page: 1 });
  };

  const handleExportStudents = async (format: 'csv' | 'excel' = 'csv'): Promise<void> => {
    try {
      const blob = await apiService.students.exportStudents({ business_id: businessId }, format);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `students-${new Date().toISOString().split('T')[0]}.${format === 'csv' ? 'csv' : 'xlsx'}`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
      toast.success('Students exported successfully');
    } catch (error) {
      toast.error('Failed to export students');
    }
  };

  // Search functionality with debouncing
  const handleQuickSearch = async (): Promise<void> => {
    if (searchQuery.trim()) {
      try {
        const searchResult = await apiService.students.searchStudents(searchQuery, 50, businessId);
        setStudents(searchResult.students);
        setPagination(null); // Disable pagination for search results
      } catch (error) {
        toast.error('Search failed');
      }
    } else {
      await fetchStudents(); // Reset to normal list
    }
  };

  // Information field management functions
  const addNewInfoField = (): void => {
    if (newInfoField.key && newInfoField.value) {
      setNewStudent({
        ...newStudent,
        information: {
          ...newStudent.information,
          [newInfoField.key]: newInfoField.value
        }
      });
      setNewInfoField({ key: '', value: '' });
    }
  };

  const removeNewInfoField = (key: string): void => {
    const { [key]: removed, ...rest } = newStudent.information;
    setNewStudent({ ...newStudent, information: rest });
  };

  const addEditInfoField = (): void => {
    if (editInfoField.key && editInfoField.value) {
      setEditStudent({
        ...editStudent,
        information: {
          ...editStudent.information,
          [editInfoField.key]: editInfoField.value
        }
      });
      setEditInfoField({ key: '', value: '' });
    }
  };

  const removeEditInfoField = (key: string): void => {
    const { [key]: removed, ...rest } = editStudent.information;
    setEditStudent({ ...editStudent, information: rest });
  };

  const formatPhoneNumber = (phone: string): string => {
    return apiService.students.formatPhoneNumber(phone);
  };

  const handlePaginationChange = (newPage: number): void => {
    setFilters({ ...filters, page: newPage });
  };

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      {stats && guardianStats && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center">
                <Users className="h-8 w-8 text-blue-600" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600">Total Students</p>
                  <p className="text-2xl font-bold">{stats.total_students}</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center">
                <ToggleRight className="h-8 w-8 text-green-600" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600">Active Students</p>
                  <p className="text-2xl font-bold">{stats.active_students}</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center">
                <Mail className="h-8 w-8 text-purple-600" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600">With Guardian Email</p>
                  <p className="text-2xl font-bold">{guardianStats.students_with_guardian_email}</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center">
                <Phone className="h-8 w-8 text-orange-600" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600">With Guardian Phone</p>
                  <p className="text-2xl font-bold">{guardianStats.students_with_guardian_phone}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Student Management</CardTitle>
          <CardDescription>Manage students for your business</CardDescription>
          
          <div className="flex flex-col sm:flex-row gap-4 mt-4">
            <div className="flex-1 relative">
              <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search students..."
                className="pl-8"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleQuickSearch()}
              />
            </div>
            
            <Select onValueChange={handleStatusFilter}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Filter by status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                {apiService.students.getStatusOptions().map((option) => (
                  <SelectItem key={option.value} value={option.value === 1 ? 'active' : 'inactive'}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select onValueChange={handleSortChange}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Sort by..." />
              </SelectTrigger>
              <SelectContent>
                {apiService.students.getAvailableSortFields().map((field) => (
                  <SelectItem key={field.value} value={field.value}>
                    {field.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Button variant="outline" onClick={() => handleExportStudents('csv')}>
              <Download className="mr-2 h-4 w-4" />
              Export CSV
            </Button>

            <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
              <DialogTrigger asChild>
                <Button>
                  <Plus className="mr-2 h-4 w-4" />
                  Add Student
                </Button>
              </DialogTrigger>
              <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
                <DialogHeader>
                  <DialogTitle>Create New Student</DialogTitle>
                </DialogHeader>
                <form onSubmit={handleCreateStudent} className="space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="user">Select User</Label>
                      <Select value={newStudent.user_id.toString()} onValueChange={(value) => setNewStudent({ ...newStudent, user_id: parseInt(value) })}>
                        <SelectTrigger>
                          <SelectValue placeholder="Select a user" />
                        </SelectTrigger>
                        <SelectContent>
                          {users.map((user) => (
                            <SelectItem key={user.id} value={user.id.toString()}>
                              {user.name} ({user.email})
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <div>
                      <Label htmlFor="name">Student Name</Label>
                      <Input
                        id="name"
                        value={newStudent.name}
                        onChange={(e) => setNewStudent({ ...newStudent, name: e.target.value })}
                        required
                        placeholder="Enter student name"
                      />
                    </div>
                    <div>
                      <Label htmlFor="guardian-name">Guardian Name</Label>
                      <Input
                        id="guardian-name"
                        value={newStudent.guardian_name}
                        onChange={(e) => setNewStudent({ ...newStudent, guardian_name: e.target.value })}
                        placeholder="Enter guardian name"
                      />
                    </div>
                    <div>
                      <Label htmlFor="guardian-email">Guardian Email</Label>
                      <Input
                        id="guardian-email"
                        type="email"
                        value={newStudent.guardian_email}
                        onChange={(e) => setNewStudent({ ...newStudent, guardian_email: e.target.value })}
                        placeholder="Enter guardian email"
                      />
                    </div>
                    <div className="md:col-span-2">
                      <Label htmlFor="guardian-number">Guardian Phone</Label>
                      <Input
                        id="guardian-number"
                        value={newStudent.guardian_number}
                        onChange={(e) => setNewStudent({ ...newStudent, guardian_number: e.target.value })}
                        placeholder="Enter guardian phone number"
                      />
                    </div>
                  </div>

                  {/* Additional Information Section */}
                  <div className="space-y-4">
                    <Label>Additional Information</Label>
                    
                    {/* Add new information field */}
                    <div className="flex gap-2">
                      <Input
                        placeholder="Field name"
                        value={newInfoField.key}
                        onChange={(e) => setNewInfoField({ ...newInfoField, key: e.target.value })}
                      />
                      <Input
                        placeholder="Field value"
                        value={newInfoField.value}
                        onChange={(e) => setNewInfoField({ ...newInfoField, value: e.target.value })}
                      />
                      <Button type="button" onClick={addNewInfoField} variant="outline" size="sm">
                        Add
                      </Button>
                    </div>

                    {/* Display existing information fields */}
                    {Object.entries(newStudent.information).map(([key, value]) => (
                      <div key={key} className="flex items-center gap-2 p-2 bg-gray-50 rounded">
                        <span className="font-medium">{key}:</span>
                        <span>{String(value)}</span>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => removeNewInfoField(key)}
                          className="ml-auto"
                        >
                          Remove
                        </Button>
                      </div>
                    ))}
                  </div>

                  <Button type="submit" className="w-full">
                    Create Student
                  </Button>
                </form>
              </DialogContent>
            </Dialog>

            {/* Edit Student Dialog */}
            <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
              <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
                <DialogHeader>
                  <DialogTitle>Edit Student</DialogTitle>
                </DialogHeader>
                <form onSubmit={handleUpdateStudent} className="space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="edit-name">Student Name</Label>
                      <Input
                        id="edit-name"
                        value={editStudent.name}
                        onChange={(e) => setEditStudent({ ...editStudent, name: e.target.value })}
                        required
                      />
                    </div>
                    <div>
                      <Label htmlFor="edit-status">Status</Label>
                      <Select value={editStudent.status.toString()} onValueChange={(value) => setEditStudent({ ...editStudent, status: parseInt(value) })}>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {apiService.students.getStatusOptions().map((option) => (
                            <SelectItem key={option.value} value={option.value.toString()}>
                              {option.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <div>
                      <Label htmlFor="edit-guardian-name">Guardian Name</Label>
                      <Input
                        id="edit-guardian-name"
                        value={editStudent.guardian_name}
                        onChange={(e) => setEditStudent({ ...editStudent, guardian_name: e.target.value })}
                      />
                    </div>
                    <div>
                      <Label htmlFor="edit-guardian-email">Guardian Email</Label>
                      <Input
                        id="edit-guardian-email"
                        type="email"
                        value={editStudent.guardian_email}
                        onChange={(e) => setEditStudent({ ...editStudent, guardian_email: e.target.value })}
                      />
                    </div>
                    <div className="md:col-span-2">
                      <Label htmlFor="edit-guardian-number">Guardian Phone</Label>
                      <Input
                        id="edit-guardian-number"
                        value={editStudent.guardian_number}
                        onChange={(e) => setEditStudent({ ...editStudent, guardian_number: e.target.value })}
                      />
                    </div>
                  </div>

                  {/* Additional Information Section */}
                  <div className="space-y-4">
                    <Label>Additional Information</Label>
                    
                    {/* Add new information field */}
                    <div className="flex gap-2">
                      <Input
                        placeholder="Field name"
                        value={editInfoField.key}
                        onChange={(e) => setEditInfoField({ ...editInfoField, key: e.target.value })}
                      />
                      <Input
                        placeholder="Field value"
                        value={editInfoField.value}
                        onChange={(e) => setEditInfoField({ ...editInfoField, value: e.target.value })}
                      />
                      <Button type="button" onClick={addEditInfoField} variant="outline" size="sm">
                        Add
                      </Button>
                    </div>

                    {/* Display existing information fields */}
                    {Object.entries(editStudent.information).map(([key, value]) => (
                      <div key={key} className="flex items-center gap-2 p-2 bg-gray-50 rounded">
                        <span className="font-medium">{key}:</span>
                        <span>{String(value)}</span>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => removeEditInfoField(key)}
                          className="ml-auto"
                        >
                          Remove
                        </Button>
                      </div>
                    ))}
                                    </div>

                  <Button type="submit" className="w-full">
                    Update Student
                  </Button>
                </form>
              </DialogContent>
            </Dialog>
          </div>
        </CardHeader>
        
        <CardContent>
          {loading ? (
            <div className="text-center py-4">Loading...</div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Student Name</TableHead>
                    <TableHead>Guardian Info</TableHead>
                    <TableHead>User Account</TableHead>
                    <TableHead>Additional Info</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {students.map((student) => (
                    <TableRow key={student.id}>
                      <TableCell className="font-medium">
                        {apiService.students.getStudentFullName(student)}
                      </TableCell>
                      <TableCell>
                        <div className="text-sm">
                          {apiService.students.formatGuardianContact(student)}
                        </div>
                        {student.guardian_number && (
                          <div className="text-xs text-gray-500 mt-1">
                            {formatPhoneNumber(student.guardian_number)}
                          </div>
                        )}
                      </TableCell>
                      <TableCell>{student.user?.email || 'N/A'}</TableCell>
                      <TableCell>
                        <div className="text-sm">
                          {apiService.students.getInformationFieldKeys(student).length > 0 ? (
                            <div>
                              {Object.entries(student.information || {}).slice(0, 2).map(([key, value]) => (
                                <div key={key} className="truncate">
                                  <span className="font-medium">{key}:</span> {String(value)}
                                </div>
                              ))}
                              {apiService.students.getInformationFieldKeys(student).length > 2 && (
                                <div className="text-gray-500">
                                  +{apiService.students.getInformationFieldKeys(student).length - 2} more
                                </div>
                              )}
                            </div>
                          ) : (
                            'No additional info'
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge 
                          variant={student.status === 1 ? 'default' : 'secondary'}
                          className={`${apiService.students.getStudentStatusColor(student.status) === 'green' 
                            ? 'bg-green-100 text-green-800' 
                            : 'bg-red-100 text-red-800'}`}
                        >
                          {apiService.students.getStudentStatusLabel(student.status)}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleToggleStatus(student.id, student.status)}
                            title={student.status === 1 ? 'Deactivate student' : 'Activate student'}
                          >
                            {student.status === 1 ? (
                              <ToggleRight className="h-4 w-4 text-green-600" />
                            ) : (
                              <ToggleLeft className="h-4 w-4 text-gray-400" />
                            )}
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleEditStudent(student)}
                          >
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => handleDeleteStudent(student.id)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              {students.length === 0 && !loading && (
                <div className="text-center py-8 text-muted-foreground">
                  {searchQuery ? 'No students found matching your search.' : 'No students found. Create your first student to get started.'}
                </div>
              )}

              {pagination && (
                <div className="flex items-center justify-between mt-4">
                  <div className="text-sm text-muted-foreground">
                    Showing {((pagination.page - 1) * pagination.limit) + 1} to {Math.min(pagination.page * pagination.limit, pagination.total)} of {pagination.total} results
                  </div>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handlePaginationChange(pagination.page - 1)}
                      disabled={!pagination.has_prev}
                    >
                      Previous
                    </Button>
                    <span className="flex items-center px-3 text-sm">
                      Page {pagination.page} of {pagination.total_pages}
                    </span>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handlePaginationChange(pagination.page + 1)}
                      disabled={!pagination.has_next}
                    >
                      Next
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Additional Analytics Section */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Student Overview</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <span>Active Students:</span>
                  <span className="font-semibold text-green-600">{stats.active_students}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span>Inactive Students:</span>
                  <span className="font-semibold text-red-600">{stats.inactive_students}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span>Total Students:</span>
                  <span className="font-semibold">{stats.total_students}</span>
                </div>
                {stats.active_students > 0 && stats.total_students > 0 && (
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-green-600 h-2 rounded-full" 
                      style={{ width: `${(stats.active_students / stats.total_students) * 100}%` }}
                    />
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Guardian Contact Info</CardTitle>
            </CardHeader>
            <CardContent>
              {guardianStats && (
                <div className="space-y-4">
                  <div className="flex justify-between items-center">
                    <span>With Email:</span>
                    <span className="font-semibold text-purple-600">{guardianStats.students_with_guardian_email}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span>With Phone:</span>
                    <span className="font-semibold text-orange-600">{guardianStats.students_with_guardian_phone}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span>Complete Contact Info:</span>
                    <span className="font-semibold text-blue-600">
                      {stats ? stats.total_students - apiService.students.getStudentsWithIncompleteInfo(students).length : 0}
                    </span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span>Missing Contact Info:</span>
                    <span className="font-semibold text-gray-600">
                      {apiService.students.getStudentsWithIncompleteInfo(students).length}
                    </span>
                  </div>
                  {guardianStats.students_with_guardian_email > 0 && stats && stats.total_students > 0 && (
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div 
                        className="bg-purple-600 h-2 rounded-full" 
                        style={{ width: `${(guardianStats.students_with_guardian_email / stats.total_students) * 100}%` }}
                      />
                    </div>
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      {/* Quick Actions Section */}
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-4">
            <Button 
              variant="outline" 
              onClick={() => handleExportStudents('excel')}
              className="flex items-center gap-2"
            >
              <Download className="h-4 w-4" />
              Export Excel
            </Button>
            
            <Button 
              variant="outline"
              onClick={() => {
                const activeStudents = apiService.students.groupStudentsByStatus(students).active;
                if (activeStudents.length === 0) {
                  toast.info('No active students to export');
                  return;
                }
                handleExportStudents('csv');
              }}
              className="flex items-center gap-2"
            >
              <Users className="h-4 w-4" />
              Export Active Students
            </Button>

            <Button 
              variant="outline"
              onClick={() => {
                const incompleteStudents = apiService.students.getStudentsWithIncompleteInfo(students);
                if (incompleteStudents.length === 0) {
                  toast.info('All students have complete guardian information');
                  return;
                }
                toast.info(`${incompleteStudents.length} students have incomplete guardian information`);
              }}
              className="flex items-center gap-2"
            >
              <Mail className="h-4 w-4" />
              Check Incomplete Info
            </Button>

            <Button 
              variant="outline"
              onClick={() => {
                if (searchQuery) {
                  setSearchQuery('');
                  fetchStudents();
                } else {
                  handleQuickSearch();
                }
              }}
              className="flex items-center gap-2"
            >
              <Search className="h-4 w-4" />
              {searchQuery ? 'Clear Search' : 'Quick Search'}
            </Button>

            <Button 
              variant="outline"
              onClick={async () => {
                try {
                  const studentsWithEmail = apiService.students.getStudentsWithGuardianEmail(students);
                  if (studentsWithEmail.length === 0) {
                    toast.info('No students with guardian email found');
                    return;
                  }
                  setStudents(studentsWithEmail);
                  setPagination(null);
                  toast.success(`Found ${studentsWithEmail.length} students with guardian email`);
                } catch (error) {
                  toast.error('Failed to filter students');
                }
              }}
              className="flex items-center gap-2"
            >
              <Mail className="h-4 w-4" />
              Show Students with Email
            </Button>

            <Button 
              variant="outline"
              onClick={async () => {
                try {
                  const studentsWithPhone = apiService.students.getStudentsWithGuardianPhone(students);
                  if (studentsWithPhone.length === 0) {
                    toast.info('No students with guardian phone found');
                    return;
                  }
                  setStudents(studentsWithPhone);
                  setPagination(null);
                  toast.success(`Found ${studentsWithPhone.length} students with guardian phone`);
                } catch (error) {
                  toast.error('Failed to filter students');
                }
              }}
              className="flex items-center gap-2"
            >
              <Phone className="h-4 w-4" />
              Show Students with Phone
            </Button>

            {/* Bulk Actions */}
            <Button 
              variant="outline"
              onClick={async () => {
                const selectedStudentIds = students
                  .filter(student => student.status === 0)
                  .map(student => student.id);
                
                if (selectedStudentIds.length === 0) {
                  toast.info('No inactive students to activate');
                  return;
                }

                if (window.confirm(`Activate ${selectedStudentIds.length} inactive students?`)) {
                  try {
                    await apiService.students.bulkUpdateStudentStatus(selectedStudentIds, 1);
                    toast.success(`Successfully activated ${selectedStudentIds.length} students`);
                    await fetchStudents();
                    await fetchStats();
                  } catch (error) {
                    toast.error('Failed to bulk activate students');
                  }
                }
              }}
              className="flex items-center gap-2"
            >
              <ToggleRight className="h-4 w-4" />
              Activate All Inactive
            </Button>

            <Button 
              variant="outline"
              onClick={async () => {
                setSearchQuery('');
                setFilters({ business_id: businessId, page: 1, limit: 10 });
                await fetchStudents();
                toast.success('Filters cleared and data refreshed');
              }}
              className="flex items-center gap-2"
            >
              <Search className="h-4 w-4" />
              Reset All Filters
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Student Summary by Guardian */}
      {students.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Guardian Summary</CardTitle>
            <CardDescription>Students grouped by guardian contact</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4 max-h-96 overflow-y-auto">
              {Object.entries(apiService.students.groupStudentsByGuardian(students))
                .slice(0, 10) // Limit to first 10 guardians to avoid UI overflow
                .map(([guardianKey, guardianStudents]) => (
                <div key={guardianKey} className="flex justify-between items-center p-3 bg-gray-50 rounded">
                  <div className="flex-1">
                    <div className="font-medium text-sm truncate">{guardianKey}</div>
                    <div className="text-xs text-gray-500">
                      {guardianStudents.length} student{guardianStudents.length !== 1 ? 's' : ''}
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm font-medium">
                      {guardianStudents.map(s => s.name).join(', ')}
                    </div>
                    <div className="text-xs text-gray-500">
                      Active: {guardianStudents.filter(s => s.status === 1).length} / 
                      Inactive: {guardianStudents.filter(s => s.status === 0).length}
                    </div>
                  </div>
                </div>
              ))}
              {Object.keys(apiService.students.groupStudentsByGuardian(students)).length > 10 && (
                <div className="text-center text-sm text-gray-500 p-2">
                  ... and {Object.keys(apiService.students.groupStudentsByGuardian(students)).length - 10} more guardians
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* File Import Section */}
      <Card>
        <CardHeader>
          <CardTitle>Import Students</CardTitle>
          <CardDescription>Upload a CSV or Excel file to import multiple students</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-4">
            <div>
              <Input
                type="file"
                accept=".csv,.xlsx,.xls"
                onChange={async (e) => {
                  const file = e.target.files?.[0];
                  if (file) {
                    try {
                      const result = await apiService.students.importStudents(file, businessId);
                      toast.success(`Successfully imported ${result.success} students`);
                      if (result.errors.length > 0) {
                        toast.warning(`${result.errors.length} rows had errors`);
                        console.error('Import errors:', result.errors);
                      }
                      await fetchStudents();
                      await fetchStats();
                    } catch (error: any) {
                      toast.error(error.response?.data?.message || 'Failed to import students');
                    }
                  }
                }}
              />
            </div>
            <div className="text-sm text-gray-500">
              <p>Supported formats: CSV, Excel (.xlsx, .xls)</p>
              <p>Required columns: name, guardian_name, guardian_email (optional), guardian_number (optional)</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}