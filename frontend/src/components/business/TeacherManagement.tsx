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
import { Teacher, TeacherFilters, CreateTeacherRequest } from '@/types/teacher';
import { User } from '@/types/user';
import { Plus, Trash2, Search, Edit, ToggleLeft, ToggleRight, DollarSign, GraduationCap } from 'lucide-react';

interface TeacherManagementProps {
  businessId: number;
}

export default function TeacherManagement({ businessId }: TeacherManagementProps) {
  const [teachers, setTeachers] = useState<Teacher[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState<TeacherFilters>({ business_id: businessId, page: 1, limit: 10 });
  const [pagination, setPagination] = useState<any>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [editingTeacher, setEditingTeacher] = useState<Teacher | null>(null);
  const [stats, setStats] = useState<any>(null);
  
  const [newTeacher, setNewTeacher] = useState<CreateTeacherRequest>({
    name: '',
    user_id: 0,
    business_id: businessId,
    salary: 0,
    qualification: '',
    experience: '',
    description: '',
  });
  
  const [editTeacher, setEditTeacher] = useState({
    name: '',
    salary: 0,
    qualification: '',
    experience: '',
    description: '',
    status: 1,
  });

  const fetchTeachers = async () => {
    try {
      setLoading(true);
      const response = await apiService.teachers.getTeachersByBusiness(businessId, filters);
      setTeachers(response.data.teachers);
      setPagination({
        page: response.data.page,
        limit: response.data.limit,
        total: response.data.total,
        total_pages: Math.ceil(response.data.total / response.data.limit),
        has_prev: response.data.page > 1,
        has_next: response.data.page < Math.ceil(response.data.total / response.data.limit),
      });
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to fetch teachers');
    } finally {
      setLoading(false);
    }
  };

  const fetchUsers = async () => {
    try {
      // Fetch available users that can be teachers (not already assigned)
      const response = await apiService.users.getUsers({ role: 'teacher', status: 1, limit: 100 });
      setUsers(response.data);
    } catch (error: any) {
      console.error('Failed to fetch users:', error);
    }
  };

  const fetchStats = async () => {
    try {
      const stats = await apiService.teachers.getTeacherStats(businessId);
      setStats(stats);
    } catch (error) {
      console.error('Failed to fetch teacher stats:', error);
    }
  };

  useEffect(() => {
    fetchTeachers();
    fetchStats();
  }, [filters]);

  useEffect(() => {
    fetchUsers();
  }, []);

  const handleCreateTeacher = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await apiService.teachers.createTeacher(newTeacher);
      toast.success('Teacher created successfully');
      setCreateDialogOpen(false);
      setNewTeacher({
        name: '',
        user_id: 0,
        business_id: businessId,
        salary: 0,
        qualification: '',
        experience: '',
        description: '',
      });
      fetchTeachers();
      fetchStats();
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to create teacher');
    }
  };

  const handleEditTeacher = (teacher: Teacher) => {
    setEditingTeacher(teacher);
    setEditTeacher({
      name: teacher.name,
      salary: teacher.salary,
      qualification: teacher.qualification,
      experience: teacher.experience,
      description: teacher.description,
      status: teacher.status,
    });
    setEditDialogOpen(true);
  };

  const handleUpdateTeacher = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingTeacher) return;

    try {
      await apiService.teachers.updateTeacher(editingTeacher.id, editTeacher);
      toast.success('Teacher updated successfully');
      setEditDialogOpen(false);
      setEditingTeacher(null);
      setEditTeacher({
        name: '',
        salary: 0,
        qualification: '',
        experience: '',
        description: '',
        status: 1,
      });
      fetchTeachers();
      fetchStats();
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to update teacher');
    }
  };

  const handleDeleteTeacher = async (id: number) => {
    if (window.confirm('Are you sure you want to delete this teacher?')) {
      try {
        await apiService.teachers.deleteTeacher(id);
        toast.success('Teacher deleted successfully');
        fetchTeachers();
        fetchStats();
      } catch (error: any) {
        toast.error(error.response?.data?.message || 'Failed to delete teacher');
      }
    }
  };

  const handleToggleStatus = async (id: number, currentStatus: number) => {
    try {
      const newStatus = currentStatus === 1 ? 0 : 1;
      await apiService.teachers.changeTeacherStatus(id, newStatus);
      toast.success(`Teacher ${newStatus === 1 ? 'activated' : 'deactivated'} successfully`);
      fetchTeachers();
      fetchStats();
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to change teacher status');
    }
  };

  const handleSearch = (search: string) => {
    setFilters({ ...filters, search, page: 1 });
  };

  const handleStatusFilter = (status: string) => {
    let statusValue: number | undefined;
    if (status === 'active') statusValue = 1;
    else if (status === 'inactive') statusValue = 0;
    setFilters({ ...filters, status: statusValue, page: 1 });
  };

  const handleSalaryFilter = (range: string) => {
    let minSalary: number | undefined;
    let maxSalary: number | undefined;
    
    switch (range) {
      case '0-30000':
        minSalary = 0;
        maxSalary = 30000;
        break;
      case '30000-60000':
        minSalary = 30000;
        maxSalary = 60000;
        break;
      case '60000+':
        minSalary = 60000;
        break;
      default:
        minSalary = undefined;
        maxSalary = undefined;
    }
    
    setFilters({ ...filters, min_salary: minSalary, max_salary: maxSalary, page: 1 });
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(amount);
  };

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center">
                <GraduationCap className="h-8 w-8 text-blue-600" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600">Total Teachers</p>
                  <p className="text-2xl font-bold">{stats.total_teachers}</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center">
                <ToggleRight className="h-8 w-8 text-green-600" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600">Active Teachers</p>
                  <p className="text-2xl font-bold">{stats.active_teachers}</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center">
                <ToggleLeft className="h-8 w-8 text-red-600" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600">Inactive Teachers</p>
                  <p className="text-2xl font-bold">{stats.inactive_teachers}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Teacher Management</CardTitle>
          <CardDescription>Manage teachers for your business</CardDescription>
          
          <div className="flex flex-col sm:flex-row gap-4 mt-4">
            <div className="flex-1 relative">
              <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search teachers..."
                className="pl-8"
                onChange={(e) => handleSearch(e.target.value)}
              />
            </div>
            
            <Select onValueChange={handleStatusFilter}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Filter by status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="active">Active</SelectItem>
                <SelectItem value="inactive">Inactive</SelectItem>
              </SelectContent>
            </Select>

            <Select onValueChange={handleSalaryFilter}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Filter by salary" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Salaries</SelectItem>
                <SelectItem value="0-30000">$0 - $30,000</SelectItem>
                <SelectItem value="30000-60000">$30,000 - $60,000</SelectItem>
                <SelectItem value="60000+">$60,000+</SelectItem>
              </SelectContent>
            </Select>

            <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
              <DialogTrigger asChild>
                <Button>
                  <Plus className="mr-2 h-4 w-4" />
                  Add Teacher
                </Button>
              </DialogTrigger>
              <DialogContent className="max-w-md max-h-[80vh] overflow-y-auto">
                <DialogHeader>
                  <DialogTitle>Create New Teacher</DialogTitle>
                </DialogHeader>
                <form onSubmit={handleCreateTeacher} className="space-y-4">
                  <div>
                    <Label htmlFor="user">Select User</Label>
                    <Select value={newTeacher.user_id.toString()} onValueChange={(value) => setNewTeacher({ ...newTeacher, user_id: parseInt(value) })}>
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
                    <Label htmlFor="name">Name</Label>
                    <Input
                      id="name"
                      value={newTeacher.name}
                      onChange={(e) => setNewTeacher({ ...newTeacher, name: e.target.value })}
                      required
                      placeholder="Enter teacher name"
                    />
                  </div>
                  <div>
                    <Label htmlFor="salary">Salary</Label>
                    <Input
                      id="salary"
                      type="number"
                      value={newTeacher.salary}
                      onChange={(e) => setNewTeacher({ ...newTeacher, salary: parseFloat(e.target.value) || 0 })}
                      placeholder="Enter salary"
                    />
                  </div>
                  <div>
                    <Label htmlFor="qualification">Qualification</Label>
                    <Input
                      id="qualification"
                      value={newTeacher.qualification}
                      onChange={(e) => setNewTeacher({ ...newTeacher, qualification: e.target.value })}
                      placeholder="Enter qualification"
                    />
                  </div>
                  <div>
                    <Label htmlFor="experience">Experience</Label>
                    <Textarea
                      id="experience"
                      value={newTeacher.experience}
                      onChange={(e) => setNewTeacher({ ...newTeacher, experience: e.target.value })}
                      placeholder="Enter experience details"
                      rows={3}
                    />
                  </div>
                  <div>
                    <Label htmlFor="description">Description</Label>
                    <Textarea
                      id="description"
                      value={newTeacher.description}
                      onChange={(e) => setNewTeacher({ ...newTeacher, description: e.target.value })}
                      placeholder="Enter description"
                      rows={3}
                    />
                  </div>
                  <Button type="submit" className="w-full">
                    Create Teacher
                  </Button>
                </form>
              </DialogContent>
            </Dialog>

            {/* Edit Teacher Dialog */}
            <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
              <DialogContent className="max-w-md max-h-[80vh] overflow-y-auto">
                <DialogHeader>
                  <DialogTitle>Edit Teacher</DialogTitle>
                </DialogHeader>
                <form onSubmit={handleUpdateTeacher} className="space-y-4">
                  <div>
                    <Label htmlFor="edit-name">Name</Label>
                    <Input
                      id="edit-name"
                      value={editTeacher.name}
                      onChange={(e) => setEditTeacher({ ...editTeacher, name: e.target.value })}
                      required
                    />
                  </div>
                  <div>
                    <Label htmlFor="edit-salary">Salary</Label>
                    <Input
                      id="edit-salary"
                      type="number"
                      value={editTeacher.salary}
                      onChange={(e) => setEditTeacher({ ...editTeacher, salary: parseFloat(e.target.value) || 0 })}
                    />
                  </div>
                  <div>
                    <Label htmlFor="edit-qualification">Qualification</Label>
                    <Input
                      id="edit-qualification"
                      value={editTeacher.qualification}
                      onChange={(e) => setEditTeacher({ ...editTeacher, qualification: e.target.value })}
                    />
                  </div>
                  <div>
                    <Label htmlFor="edit-experience">Experience</Label>
                    <Textarea
                      id="edit-experience"
                      value={editTeacher.experience}
                      onChange={(e) => setEditTeacher({ ...editTeacher, experience: e.target.value })}
                      rows={3}
                    />
                  </div>
                  <div>
                    <Label htmlFor="edit-description">Description</Label>
                    <Textarea
                      id="edit-description"
                      value={editTeacher.description}
                      onChange={(e) => setEditTeacher({ ...editTeacher, description: e.target.value })}
                      rows={3}
                    />
                  </div>
                  <div>
                    <Label htmlFor="edit-status">Status</Label>
                    <Select value={editTeacher.status.toString()} onValueChange={(value) => setEditTeacher({ ...editTeacher, status: parseInt(value) })}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="1">Active</SelectItem>
                        <SelectItem value="0">Inactive</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <Button type="submit" className="w-full">
                    Update Teacher
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
                    <TableHead>Name</TableHead>
                    <TableHead>Qualification</TableHead>
                    <TableHead>Salary</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>User</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {teachers.map((teacher) => (
                    <TableRow key={teacher.id}>
                      <TableCell className="font-medium">{teacher.name}</TableCell>
                      <TableCell>{teacher.qualification || 'N/A'}</TableCell>
                      <TableCell className="font-medium">{formatCurrency(teacher.salary)}</TableCell>
                      <TableCell>
                        <Badge 
                          variant={teacher.status === 1 ? 'default' : 'secondary'}
                          className={teacher.status === 1 ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}
                        >
                          {teacher.status === 1 ? 'Active' : 'Inactive'}
                        </Badge>
                      </TableCell>
                      <TableCell>{teacher.user?.email || 'N/A'}</TableCell>
                      <TableCell>
                        <div className="flex gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleToggleStatus(teacher.id, teacher.status)}
                            title={teacher.status === 1 ? 'Deactivate teacher' : 'Activate teacher'}
                          >
                            {teacher.status === 1 ? (
                              <ToggleRight className="h-4 w-4 text-green-600" />
                            ) : (
                              <ToggleLeft className="h-4 w-4 text-gray-400" />
                            )}
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleEditTeacher(teacher)}
                          >
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => handleDeleteTeacher(teacher.id)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              {teachers.length === 0 && !loading && (
                <div className="text-center py-8 text-muted-foreground">
                  No teachers found. Create your first teacher to get started.
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
                      onClick={() => setFilters({ ...filters, page: filters.page! - 1 })}
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
                      onClick={() => setFilters({ ...filters, page: filters.page! + 1 })}
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
    </div>
  );
}