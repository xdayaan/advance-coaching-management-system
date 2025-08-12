'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { toast } from 'sonner';
import { apiService } from '@/services/api';
import { Package, PackageFilters, CreatePackageRequest } from '@/types/package';
import { Plus, Trash2, Search, Edit, DollarSign, Calendar, ToggleLeft, ToggleRight } from 'lucide-react';

export default function PackageManagement() {
  const [packages, setPackages] = useState<Package[]>([]);
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState<PackageFilters>({ page: 1, limit: 10 });
  const [pagination, setPagination] = useState<any>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [editingPackage, setEditingPackage] = useState<Package | null>(null);
  const [newPackage, setNewPackage] = useState<CreatePackageRequest>({
    name: '', 
    price: 0,
    validation_period: 30,
    description: '',
  });
  const [editPackage, setEditPackage] = useState({
    name: '',
    price: 0,
    validation_period: 30,
    description: '',
  });

  const fetchPackages = async () => {
    try {
      setLoading(true);
      const response = await apiService.packages.getPackages(filters);
      setPackages(response.data);
      setPagination(response.pagination);
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to fetch packages');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPackages();
  }, [filters]);

  const handleCreatePackage = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await apiService.packages.createPackage(newPackage);
      toast.success('Package created successfully');
      setCreateDialogOpen(false);
      setNewPackage({ name: '', price: 0, validation_period: 30, description: '' });
      fetchPackages();
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to create package');
    }
  };

  const handleEditPackage = (pkg: Package) => {
    setEditingPackage(pkg);
    setEditPackage({
      name: pkg.name,
      price: pkg.price,
      validation_period: pkg.validation_period,
      description: pkg.description || '',
    });
    setEditDialogOpen(true);
  };

  const handleUpdatePackage = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingPackage) return;

    try {
      await apiService.packages.updatePackage(editingPackage.id, editPackage);
      toast.success('Package updated successfully');
      setEditDialogOpen(false);
      setEditingPackage(null);
      setEditPackage({ name: '', price: 0, validation_period: 30, description: '' });
      fetchPackages();
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to update package');
    }
  };

  const handleDeletePackage = async (id: number) => {
    if (window.confirm('Are you sure you want to delete this package?')) {
      try {
        await apiService.packages.deletePackage(id);
        toast.success('Package deleted successfully');
        fetchPackages();
      } catch (error: any) {
        toast.error(error.response?.data?.message || 'Failed to delete package');
      }
    }
  };

  const handleToggleStatus = async (id: number, currentStatus: number) => {
    try {
      const newStatus = currentStatus === 1 ? 0 : 1;
      await apiService.packages.changePackageStatus(id, newStatus);
      toast.success(`Package ${newStatus === 1 ? 'activated' : 'deactivated'} successfully`);
      fetchPackages();
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to change package status');
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

  const handlePriceRangeFilter = (minPrice: string, maxPrice: string) => {
    setFilters({
      ...filters,
      min_price: minPrice ? parseFloat(minPrice) : undefined,
      max_price: maxPrice ? parseFloat(maxPrice) : undefined,
      page: 1,
    });
  };

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price);
  };

  const formatValidationPeriod = (days: number) => {
    if (days === 1) return '1 day';
    if (days < 30) return `${days} days`;
    if (days === 30) return '1 month';
    if (days < 365) {
      const months = Math.round(days / 30);
      return `${months} month${months > 1 ? 's' : ''}`;
    }
    const years = Math.round(days / 365);
    return `${years} year${years > 1 ? 's' : ''}`;
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Package Management</CardTitle>
        <CardDescription>Manage subscription packages in your system</CardDescription>
        
        <div className="flex flex-col sm:flex-row gap-4 mt-4">
          <div className="flex-1 relative">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search packages..."
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

          <div className="flex gap-2">
            <Input
              placeholder="Min price"
              type="number"
              className="w-24"
              onChange={(e) => {
                const maxPriceInput = document.getElementById('max-price') as HTMLInputElement;
                handlePriceRangeFilter(e.target.value, maxPriceInput?.value || '');
              }}
            />
            <Input
              id="max-price"
              placeholder="Max price"
              type="number"
              className="w-24"
              onChange={(e) => {
                const minPriceInput = document.querySelector('input[placeholder="Min price"]') as HTMLInputElement;
                handlePriceRangeFilter(minPriceInput?.value || '', e.target.value);
              }}
            />
          </div>

          <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="mr-2 h-4 w-4" />
                Add Package
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-md">
              <DialogHeader>
                <DialogTitle>Create New Package</DialogTitle>
              </DialogHeader>
              <form onSubmit={handleCreatePackage} className="space-y-4">
                <div>
                  <Label htmlFor="name">Package Name</Label>
                  <Input
                    id="name"
                    value={newPackage.name}
                    onChange={(e) => setNewPackage({ ...newPackage, name: e.target.value })}
                    required
                    placeholder="e.g., Basic Plan"
                  />
                </div>
                <div>
                  <Label htmlFor="price">Price ($)</Label>
                  <Input
                    id="price"
                    type="number"
                    step="0.01"
                    min="0"
                    value={newPackage.price}
                    onChange={(e) => setNewPackage({ ...newPackage, price: parseFloat(e.target.value) || 0 })}
                    required
                    placeholder="0.00"
                  />
                </div>
                <div>
                  <Label htmlFor="validation_period">Validation Period (days)</Label>
                  <Input
                    id="validation_period"
                    type="number"
                    min="1"
                    value={newPackage.validation_period}
                    onChange={(e) => setNewPackage({ ...newPackage, validation_period: parseInt(e.target.value) || 30 })}
                    required
                    placeholder="30"
                  />
                </div>
                <div>
                  <Label htmlFor="description">Description</Label>
                  <Textarea
                    id="description"
                    value={newPackage.description}
                    onChange={(e) => setNewPackage({ ...newPackage, description: e.target.value })}
                    placeholder="Package description..."
                    rows={3}
                  />
                </div>
                <Button type="submit" className="w-full">
                  Create Package
                </Button>
              </form>
            </DialogContent>
          </Dialog>

          {/* Edit Package Dialog */}
          <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
            <DialogContent className="max-w-md">
              <DialogHeader>
                <DialogTitle>Edit Package</DialogTitle>
              </DialogHeader>
              <form onSubmit={handleUpdatePackage} className="space-y-4">
                <div>
                  <Label htmlFor="edit-name">Package Name</Label>
                  <Input
                    id="edit-name"
                    value={editPackage.name}
                    onChange={(e) => setEditPackage({ ...editPackage, name: e.target.value })}
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="edit-price">Price ($)</Label>
                  <Input
                    id="edit-price"
                    type="number"
                    step="0.01"
                    min="0"
                    value={editPackage.price}
                    onChange={(e) => setEditPackage({ ...editPackage, price: parseFloat(e.target.value) || 0 })}
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="edit-validation_period">Validation Period (days)</Label>
                  <Input
                    id="edit-validation_period"
                    type="number"
                    min="1"
                    value={editPackage.validation_period}
                    onChange={(e) => setEditPackage({ ...editPackage, validation_period: parseInt(e.target.value) || 30 })}
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="edit-description">Description</Label>
                  <Textarea
                    id="edit-description"
                    value={editPackage.description}
                    onChange={(e) => setEditPackage({ ...editPackage, description: e.target.value })}
                    rows={3}
                  />
                </div>
                <Button type="submit" className="w-full">
                  Update Package
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
                  <TableHead>Package Name</TableHead>
                  <TableHead>Price</TableHead>
                  <TableHead>Validation Period</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                  {(packages ?? []).map((pkg) => (
                    <TableRow key={pkg.id}>
                      <TableCell className="font-medium">{pkg.name}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <DollarSign className="h-3 w-3" />
                          {formatPrice(pkg.price)}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <Calendar className="h-3 w-3" />
                          {formatValidationPeriod(pkg.validation_period)}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="max-w-xs truncate" title={pkg.description}>
                          {pkg.description || 'No description'}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge 
                          variant={pkg.status === 1 ? 'default' : 'secondary'}
                          className={pkg.status === 1 ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}
                        >
                          {pkg.status === 1 ? 'Active' : 'Inactive'}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleToggleStatus(pkg.id, pkg.status)}
                            title={pkg.status === 1 ? 'Deactivate package' : 'Activate package'}
                          >
                            {pkg.status === 1 ? (
                              <ToggleRight className="h-4 w-4 text-green-600" />
                            ) : (
                              <ToggleLeft className="h-4 w-4 text-gray-400" />
                            )}
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleEditPackage(pkg)}
                          >
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => handleDeletePackage(pkg.id)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
            </Table>

            {(packages ?? []).length === 0 && !loading && (
              <div className="text-center py-8 text-muted-foreground">
                No packages found. Create your first package to get started.
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
  );
}