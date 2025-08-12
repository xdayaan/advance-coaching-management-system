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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Checkbox } from '@/components/ui/checkbox';
import { toast } from 'sonner';
import { apiService } from '@/services/api';
import { Business, BusinessFilters, CreateBusinessRequest, BusinessStats, LocationStats, PackageDistribution } from '@/types/business';
import { Package } from '@/types/package';
import { 
  Plus, Trash2, Search, Edit, User, Mail, Phone, MapPin, 
  ToggleLeft, ToggleRight, Package as PackageIcon, Building2,
  Users, Activity, TrendingUp, BarChart3, PieChart, Download,
  Upload, Filter, MoreHorizontal
} from 'lucide-react';

export default function BusinessManagement() {
  const [businesses, setBusinesses] = useState<Business[]>([]);
  const [packages, setPackages] = useState<Package[]>([]);
  const [locations, setLocations] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [statsLoading, setStatsLoading] = useState(false);
  const [filters, setFilters] = useState<BusinessFilters>({ page: 1, limit: 10 });
  const [pagination, setPagination] = useState<any>(null);
  const [selectedBusinesses, setSelectedBusinesses] = useState<number[]>([]);
  
  // Dialog states
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [bulkActionDialogOpen, setBulkActionDialogOpen] = useState(false);
  const [statsDialogOpen, setStatsDialogOpen] = useState(false);
  
  // Editing state
  const [editingBusiness, setEditingBusiness] = useState<Business | null>(null);
  
  // Form states
  const [newBusiness, setNewBusiness] = useState<CreateBusinessRequest>({
    name: '',
    owner_name: '',
    slug: '',
    email: '',
    phone: '',
    location: '',
    password: '',
  });
  
  const [editBusiness, setEditBusiness] = useState({
    name: '',
    slug: '',
    owner_name: '',
    email: '',
    phone: '',
    location: '',
    package_id: undefined as number | undefined,
  });

  // Stats states
  const [businessStats, setBusinessStats] = useState<BusinessStats | null>(null);
  const [locationStats, setLocationStats] = useState<LocationStats>({});
  const [packageDistribution, setPackageDistribution] = useState<PackageDistribution>({});

  const fetchBusinesses = async () => {
    try {
      setLoading(true);
      const response = await apiService.businesses.getBusinesses(filters);
      setBusinesses(response.data.businesses);
      setPagination({
        total: response.data.total,
        page: response.data.page,
        limit: response.data.limit,
        total_pages: Math.ceil(response.data.total / response.data.limit),
        has_next: response.data.page * response.data.limit < response.data.total,
        has_prev: response.data.page > 1,
      });
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to fetch businesses');
    } finally {
      setLoading(false);
    }
  };

  const fetchPackages = async () => {
    try {
      const response = await apiService.packages.getActivePackages();
      setPackages(response);
    } catch (error: any) {
      console.error('Failed to fetch packages:', error);
    }
  };

  const fetchLocations = async () => {
    try {
      const response = await apiService.businesses.getBusinessLocations();
      setLocations(response);
    } catch (error: any) {
      console.error('Failed to fetch locations:', error);
    }
  };

  const fetchStats = async () => {
    try {
      setStatsLoading(true);
      const [stats, locStats, pkgDist] = await Promise.all([
        apiService.businesses.getBusinessStats(),
        apiService.businesses.getLocationStats(),
        apiService.businesses.getPackageDistribution(),
      ]);
      setBusinessStats(stats);
      setLocationStats(locStats);
      setPackageDistribution(pkgDist);
    } catch (error: any) {
      toast.error('Failed to fetch statistics');
    } finally {
      setStatsLoading(false);
    }
  };

  useEffect(() => {
    fetchBusinesses();
  }, [filters]);

  useEffect(() => {
    fetchPackages();
    fetchLocations();
    fetchStats(); // Load stats on component mount
  }, []);

  const handleCreateBusiness = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await apiService.businesses.createBusiness(newBusiness);
      toast.success('Business created successfully');
      setCreateDialogOpen(false);
      setNewBusiness({
        name: '',
        owner_name: '',
        email: '',
        phone: '',
        location: '',
        password: '',
      });
      fetchBusinesses();
      fetchStats(); // Refresh stats
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to create business');
    }
  };

  const handleEditBusiness = (business: Business) => {
    setEditingBusiness(business);
    setEditBusiness({
      name: business.name,
      slug: business.slug,
      owner_name: business.owner_name,
      email: business.email,
      phone: business.phone,
      location: business.location,
      package_id: business.package_id,
    });
    setEditDialogOpen(true);
  };

  const handleUpdateBusiness = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingBusiness) return;

    try {
      await apiService.businesses.updateBusiness(editingBusiness.id, editBusiness);
      toast.success('Business updated successfully');
      setEditDialogOpen(false);
      setEditingBusiness(null);
      setEditBusiness({
        name: '',
        slug: '',
        owner_name: '',
        email: '',
        phone: '',
        location: '',
        package_id: undefined,
      });
      fetchBusinesses();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to update business');
    }
  };

  const handleDeleteBusiness = async (id: number) => {
    if (window.confirm('Are you sure you want to delete this business? This will also delete the associated user account.')) {
      try {
        await apiService.businesses.deleteBusiness(id);
        toast.success('Business deleted successfully');
        fetchBusinesses();
        fetchStats(); // Refresh stats
      } catch (error: any) {
        toast.error(error.response?.data?.error || 'Failed to delete business');
      }
    }
  };

  const handleToggleStatus = async (id: number, currentStatus: number) => {
    try {
      const newStatus = currentStatus === 1 ? 0 : 1;
      await apiService.businesses.changeBusinessStatus(id, newStatus);
      toast.success(`Business ${newStatus === 1 ? 'activated' : 'deactivated'} successfully`);
      fetchBusinesses();
      fetchStats(); // Refresh stats
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to change business status');
    }
  };

  const handleAssignPackage = async (businessId: number, packageId: number) => {
    try {
      await apiService.businesses.assignPackage(businessId, packageId);
      toast.success('Package assigned successfully');
      fetchBusinesses();
      fetchStats(); // Refresh stats
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to assign package');
    }
  };

  const handleRemovePackage = async (businessId: number) => {
    try {
      await apiService.businesses.removePackage(businessId);
      toast.success('Package removed successfully');
      fetchBusinesses();
      fetchStats(); // Refresh stats
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to remove package');
    }
  };

  const handleBulkAction = async (action: 'activate' | 'deactivate' | 'assign-package', data?: any) => {
    if (selectedBusinesses.length === 0) {
      toast.error('Please select businesses first');
      return;
    }

    try {
      if (action === 'activate' || action === 'deactivate') {
        const status = action === 'activate' ? 1 : 0;
        await apiService.businesses.bulkUpdateStatus(selectedBusinesses, status);
        toast.success(`${selectedBusinesses.length} businesses ${action}d successfully`);
      } else if (action === 'assign-package' && data?.packageId) {
        await apiService.businesses.bulkAssignPackage(selectedBusinesses, data.packageId);
        toast.success(`Package assigned to ${selectedBusinesses.length} businesses successfully`);
      }
      
      setSelectedBusinesses([]);
      setBulkActionDialogOpen(false);
      fetchBusinesses();
      fetchStats(); // Refresh stats
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Bulk action failed');
    }
  };

  const handleSearch = (search: string) => {
    setFilters({ ...filters, search, page: 1 });
  };

  const handleStatusFilter = (status: string) => {
    let statusValue: number | undefined;
    if (status === 'active') statusValue = 1;
    else if (status === 'inactive') statusValue = 0;
    else if (status === 'all') statusValue = undefined;
    setFilters({ ...filters, status: statusValue, page: 1 });
  };

  const handleLocationFilter = (location: string) => {
    setFilters({ 
      ...filters, 
      location: location === 'all' ? undefined : location, 
      page: 1 
    });
  };

  const handlePackageFilter = (packageId: string) => {
    let packageValue: number | undefined;
    if (packageId === 'all') packageValue = undefined;
    else if (packageId === 'no-package') packageValue = 0;
    else packageValue = parseInt(packageId);
    setFilters({ ...filters, package_id: packageValue, page: 1 });
  };

  const toggleBusinessSelection = (businessId: number) => {
    setSelectedBusinesses(prev => 
      prev.includes(businessId) 
        ? prev.filter(id => id !== businessId)
        : [...prev, businessId]
    );
  };

  const toggleAllBusinesses = () => {
    if (selectedBusinesses.length === businesses.length) {
      setSelectedBusinesses([]);
    } else {
      setSelectedBusinesses(businesses.map(b => b.id));
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center">
              <Building2 className="h-4 w-4 text-blue-600" />
              <div className="ml-2">
                <p className="text-sm font-medium text-muted-foreground">Total Businesses</p>
                <p className="text-2xl font-bold">{businessStats?.total_businesses || '-'}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center">
              <Activity className="h-4 w-4 text-green-600" />
              <div className="ml-2">
                <p className="text-sm font-medium text-muted-foreground">Active</p>
                <p className="text-2xl font-bold">{businessStats?.active_businesses || '-'}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center">
              <Users className="h-4 w-4 text-red-600" />
              <div className="ml-2">
                <p className="text-sm font-medium text-muted-foreground">Inactive</p>
                <p className="text-2xl font-bold">{businessStats?.inactive_businesses || '-'}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center">
              <PackageIcon className="h-4 w-4 text-purple-600" />
              <div className="ml-2">
                <p className="text-sm font-medium text-muted-foreground">With Package</p>
                <p className="text-2xl font-bold">{businessStats?.businesses_with_packages || '-'}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center">
              <TrendingUp className="h-4 w-4 text-orange-600" />
              <div className="ml-2">
                <p className="text-sm font-medium text-muted-foreground">No Package</p>
                <p className="text-2xl font-bold">{businessStats?.businesses_without_packages || '-'}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Main Content */}
      <Card>
        <CardHeader>
          <CardTitle>Business Management</CardTitle>
          <CardDescription>Manage businesses and their associated user accounts</CardDescription>
          
          <div className="flex flex-col sm:flex-row gap-4 mt-4">
            <div className="flex-1 relative">
              <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search businesses..."
                className="pl-8"
                onChange={(e) => handleSearch(e.target.value)}
              />
            </div>
            
            <Select onValueChange={handleStatusFilter} defaultValue="all">
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Filter by status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="active">Active</SelectItem>
                <SelectItem value="inactive">Inactive</SelectItem>
              </SelectContent>
            </Select>

            <Select onValueChange={handleLocationFilter} defaultValue="all">
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Filter by location" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Locations</SelectItem>
                {locations.map((location) => (
                  <SelectItem key={location} value={location}>{location}</SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select onValueChange={handlePackageFilter} defaultValue="all">
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Filter by package" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Packages</SelectItem>
                <SelectItem value="no-package">No Package</SelectItem>
                {packages.map((pkg) => (
                  <SelectItem key={pkg.id} value={pkg.id.toString()}>{pkg.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>

            <div className="flex gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  fetchStats();
                  setStatsDialogOpen(true);
                }}
              >
                <BarChart3 className="mr-2 h-4 w-4" />
                Stats
              </Button>

              {selectedBusinesses.length > 0 && (
                <Button
                  variant="outline"
                  onClick={() => setBulkActionDialogOpen(true)}
                >
                  Actions ({selectedBusinesses.length})
                </Button>
              )}

              <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
                <DialogTrigger asChild>
                  <Button>
                    <Plus className="mr-2 h-4 w-4" />
                    Add Business
                  </Button>
                </DialogTrigger>
                <DialogContent className="max-w-2xl">
                  <DialogHeader>
                    <DialogTitle>Create New Business</DialogTitle>
                  </DialogHeader>
                  <form onSubmit={handleCreateBusiness} className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label htmlFor="name">Business Name</Label>
                        <Input
                          id="name"
                          value={newBusiness.name}
                          onChange={(e) => setNewBusiness({ ...newBusiness, name: e.target.value })}
                          required
                          placeholder="e.g., ABC Corp"
                        />
                      </div>
                      <div>
                        <Label htmlFor="owner_name">Owner Name</Label>
                        <Input
                          id="owner_name"
                          value={newBusiness.owner_name}
                          onChange={(e) => setNewBusiness({ ...newBusiness, owner_name: e.target.value })}
                          required
                          placeholder="e.g., John Doe"
                        />
                      </div>
                    </div>
                    
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label htmlFor="email">Email</Label>
                        <Input
                          id="email"
                          type="email"
                          value={newBusiness.email}
                          onChange={(e) => setNewBusiness({ ...newBusiness, email: e.target.value })}
                          required
                          placeholder="john@business.com"
                        />
                      </div>
                      <div>
                        <Label htmlFor="phone">Phone</Label>
                        <Input
                          id="phone"
                          value={newBusiness.phone}
                          onChange={(e) => setNewBusiness({ ...newBusiness, phone: e.target.value })}
                          placeholder="+1234567890"
                        />
                      </div>
                    </div>

                    <div>
                      <Label htmlFor="location">Location</Label>
                      <Input
                        id="location"
                        value={newBusiness.location}
                        onChange={(e) => setNewBusiness({ ...newBusiness, location: e.target.value })}
                        placeholder="City, State"
                      />
                    </div>
                    <div>
                      <Label htmlFor="slug">Slug</Label>
                      <Input
                        id="slug"
                        value={newBusiness.slug}
                        onChange={(e) => setNewBusiness({ ...newBusiness, slug: e.target.value })}
                        placeholder="/business-name"
                      />
                    </div>

                    <div>
                      <Label htmlFor="password">Password</Label>
                      <Input
                        id="password"
                        type="password"
                        value={newBusiness.password}
                        onChange={(e) => setNewBusiness({ ...newBusiness, password: e.target.value })}
                        required
                        placeholder="Minimum 6 characters"
                        minLength={6}
                      />
                    </div>

                    <div>
                      <Label htmlFor="package_id">Package (Optional)</Label>
                      <Select
                        value={newBusiness.package_id?.toString() || 'none'}
                        onValueChange={(value) => setNewBusiness({ 
                          ...newBusiness, 
                          package_id: value === 'none' ? undefined : parseInt(value)
                        })}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="Select a package" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="none">No Package</SelectItem>
                          {packages.map((pkg) => (
                            <SelectItem key={pkg.id} value={pkg.id.toString()}>
                              {pkg.name} - ${pkg.price}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>

                    <Button type="submit" className="w-full">
                      Create Business
                    </Button>
                  </form>
                </DialogContent>
              </Dialog>

              {/* Edit Business Dialog */}
              <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
                <DialogContent className="max-w-2xl">
                  <DialogHeader>
                    <DialogTitle>Edit Business</DialogTitle>
                  </DialogHeader>
                  <form onSubmit={handleUpdateBusiness} className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label htmlFor="edit-name">Business Name</Label>
                        <Input
                          id="edit-name"
                          value={editBusiness.name}
                          onChange={(e) => setEditBusiness({ ...editBusiness, name: e.target.value })}
                          required
                        />
                      </div>
                      <div>
                        <Label htmlFor="edit-slug">Slug</Label>
                        <Input
                          id="edit-slug"
                          value={editBusiness.slug}
                          onChange={(e) => setEditBusiness({ ...editBusiness, slug: e.target.value })}
                          placeholder="business-slug"
                        />
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label htmlFor="edit-owner_name">Owner Name</Label>
                        <Input
                          id="edit-owner_name"
                          value={editBusiness.owner_name}
                                                    onChange={(e) => setEditBusiness({ ...editBusiness, owner_name: e.target.value })}
                          required
                        />
                      </div>
                      <div>
                        <Label htmlFor="edit-email">Email</Label>
                        <Input
                          id="edit-email"
                          type="email"
                          value={editBusiness.email}
                          onChange={(e) => setEditBusiness({ ...editBusiness, email: e.target.value })}
                          required
                        />
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label htmlFor="edit-phone">Phone</Label>
                        <Input
                          id="edit-phone"
                          value={editBusiness.phone}
                          onChange={(e) => setEditBusiness({ ...editBusiness, phone: e.target.value })}
                        />
                      </div>
                      <div>
                        <Label htmlFor="edit-location">Location</Label>
                        <Input
                          id="edit-location"
                          value={editBusiness.location}
                          onChange={(e) => setEditBusiness({ ...editBusiness, location: e.target.value })}
                        />
                      </div>
                    </div>

                    <div>
                      <Label htmlFor="edit-package_id">Package</Label>
                      <Select
                        value={editBusiness.package_id?.toString() || 'none'}
                        onValueChange={(value) => setEditBusiness({ 
                          ...editBusiness, 
                          package_id: value === 'none' ? undefined : parseInt(value)
                        })}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="Select a package" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="none">No Package</SelectItem>
                          {packages.map((pkg) => (
                            <SelectItem key={pkg.id} value={pkg.id.toString()}>
                              {pkg.name} - ${pkg.price}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>

                    <Button type="submit" className="w-full">
                      Update Business
                    </Button>
                  </form>
                </DialogContent>
              </Dialog>

              {/* Bulk Actions Dialog */}
              <Dialog open={bulkActionDialogOpen} onOpenChange={setBulkActionDialogOpen}>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>Bulk Actions</DialogTitle>
                  </DialogHeader>
                  <div className="space-y-4">
                    <p className="text-sm text-muted-foreground">
                      {selectedBusinesses.length} business(es) selected
                    </p>
                    
                    <div className="grid grid-cols-2 gap-2">
                      <Button
                        variant="outline"
                        onClick={() => handleBulkAction('activate')}
                      >
                        Activate Selected
                      </Button>
                      <Button
                        variant="outline"
                        onClick={() => handleBulkAction('deactivate')}
                      >
                        Deactivate Selected
                      </Button>
                    </div>

                    <div>
                      <Label>Assign Package</Label>
                      <div className="flex gap-2 mt-1">
                        <Select
                          onValueChange={(packageId) => {
                            if (packageId && packageId !== 'none') {
                              handleBulkAction('assign-package', { packageId: parseInt(packageId) });
                            }
                          }}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Select package" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="none">Select a package...</SelectItem>
                            {packages.map((pkg) => (
                              <SelectItem key={pkg.id} value={pkg.id.toString()}>
                                {pkg.name} - ${pkg.price}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                    </div>
                  </div>
                </DialogContent>
              </Dialog>

              {/* Stats Dialog */}
              <Dialog open={statsDialogOpen} onOpenChange={setStatsDialogOpen}>
                <DialogContent className="max-w-4xl">
                  <DialogHeader>
                    <DialogTitle>Business Statistics</DialogTitle>
                  </DialogHeader>
                  
                  {statsLoading ? (
                    <div className="flex justify-center py-8">
                      <div className="text-center">Loading statistics...</div>
                    </div>
                  ) : (
                    <Tabs defaultValue="overview" className="space-y-4">
                      <TabsList>
                        <TabsTrigger value="overview">Overview</TabsTrigger>
                        <TabsTrigger value="locations">Locations</TabsTrigger>
                        <TabsTrigger value="packages">Packages</TabsTrigger>
                      </TabsList>

                      <TabsContent value="overview" className="space-y-4">
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                          <Card>
                            <CardContent className="p-4 text-center">
                              <Building2 className="h-8 w-8 mx-auto text-blue-600 mb-2" />
                              <p className="text-2xl font-bold">{businessStats?.total_businesses}</p>
                              <p className="text-sm text-muted-foreground">Total Businesses</p>
                            </CardContent>
                          </Card>
                          
                          <Card>
                            <CardContent className="p-4 text-center">
                              <Activity className="h-8 w-8 mx-auto text-green-600 mb-2" />
                              <p className="text-2xl font-bold">{businessStats?.active_businesses}</p>
                              <p className="text-sm text-muted-foreground">Active</p>
                            </CardContent>
                          </Card>

                          <Card>
                            <CardContent className="p-4 text-center">
                              <PackageIcon className="h-8 w-8 mx-auto text-purple-600 mb-2" />
                              <p className="text-2xl font-bold">{businessStats?.businesses_with_packages}</p>
                              <p className="text-sm text-muted-foreground">With Packages</p>
                            </CardContent>
                          </Card>

                          <Card>
                            <CardContent className="p-4 text-center">
                              <Users className="h-8 w-8 mx-auto text-orange-600 mb-2" />
                              <p className="text-2xl font-bold">{businessStats?.businesses_without_packages}</p>
                              <p className="text-sm text-muted-foreground">No Package</p>
                            </CardContent>
                          </Card>
                        </div>
                      </TabsContent>

                      <TabsContent value="locations" className="space-y-4">
                        <Card>
                          <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                              <MapPin className="h-4 w-4" />
                              Businesses by Location
                            </CardTitle>
                          </CardHeader>
                          <CardContent>
                            {Object.keys(locationStats).length > 0 ? (
                              <div className="space-y-2">
                                {Object.entries(locationStats)
                                  .sort(([,a], [,b]) => b - a)
                                  .map(([location, count]) => (
                                  <div key={location} className="flex justify-between items-center">
                                    <span className="font-medium">{location}</span>
                                    <Badge variant="secondary">{count} businesses</Badge>
                                  </div>
                                ))}
                              </div>
                            ) : (
                              <p className="text-muted-foreground text-center py-4">No location data available</p>
                            )}
                          </CardContent>
                        </Card>
                      </TabsContent>

                      <TabsContent value="packages" className="space-y-4">
                        <Card>
                          <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                              <PieChart className="h-4 w-4" />
                              Package Distribution
                            </CardTitle>
                          </CardHeader>
                          <CardContent>
                            {Object.keys(packageDistribution).length > 0 ? (
                              <div className="space-y-2">
                                {Object.entries(packageDistribution)
                                  .sort(([,a], [,b]) => b - a)
                                  .map(([packageName, count]) => (
                                  <div key={packageName} className="flex justify-between items-center">
                                    <span className="font-medium">{packageName}</span>
                                    <Badge variant="secondary">{count} businesses</Badge>
                                  </div>
                                ))}
                              </div>
                            ) : (
                              <p className="text-muted-foreground text-center py-4">No package distribution data available</p>
                            )}
                          </CardContent>
                        </Card>
                      </TabsContent>
                    </Tabs>
                  )}
                </DialogContent>
              </Dialog>
            </div>
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
                    <TableHead className="w-12">
                      <Checkbox
                        checked={selectedBusinesses.length === businesses.length && businesses.length > 0}
                        onCheckedChange={toggleAllBusinesses}
                      />
                    </TableHead>
                    <TableHead>Business</TableHead>
                    <TableHead>Owner</TableHead>
                    <TableHead>Contact</TableHead>
                    <TableHead>Package</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {(businesses ?? []).map((business) => (
                    <TableRow key={business.id}>
                      <TableCell>
                        <Checkbox
                          checked={selectedBusinesses.includes(business.id)}
                          onCheckedChange={() => toggleBusinessSelection(business.id)}
                        />
                      </TableCell>
                      <TableCell>
                        <div>
                          <div className="font-medium flex items-center gap-2">
                            <Building2 className="h-3 w-3" />
                            {business.name}
                          </div>
                          {business.slug && (
                            <div className="text-xs text-muted-foreground">/{business.slug}</div>
                          )}
                          <div className="text-xs text-muted-foreground flex items-center gap-1">
                            <MapPin className="h-3 w-3" />
                            {business.location || 'No location'}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <User className="h-3 w-3" />
                          {business.owner_name}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="space-y-1">
                          <div className="flex items-center gap-1 text-xs">
                            <Mail className="h-3 w-3" />
                            {business.email}
                          </div>
                          {business.phone && (
                            <div className="flex items-center gap-1 text-xs">
                              <Phone className="h-3 w-3" />
                              {business.phone}
                            </div>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        {business.package ? (
                          <div className="flex items-center gap-1">
                            <PackageIcon className="h-3 w-3" />
                            <div>
                              <div className="font-medium text-xs">{business.package.name}</div>
                              <div className="text-xs text-muted-foreground">${business.package.price}</div>
                            </div>
                          </div>
                        ) : (
                          <Badge variant="outline" className="text-xs">No Package</Badge>
                        )}
                      </TableCell>
                      <TableCell>
                        <Badge 
                          variant={business.status === 1 ? 'default' : 'secondary'}
                          className={business.status === 1 ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}
                        >
                          {business.status === 1 ? 'Active' : 'Inactive'}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground">
                        {formatDate(business.created_on)}
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-1">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleToggleStatus(business.id, business.status)}
                            title={business.status === 1 ? 'Deactivate business' : 'Activate business'}
                          >
                            {business.status === 1 ? (
                              <ToggleRight className="h-3 w-3 text-green-600" />
                            ) : (
                              <ToggleLeft className="h-3 w-3 text-gray-400" />
                            )}
                          </Button>

                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleEditBusiness(business)}
                          >
                            <Edit className="h-3 w-3" />
                          </Button>

                          <Select onValueChange={(action) => {
                            const [actionType, ...params] = action.split('-');
                            
                            if (actionType === 'delete') {
                              handleDeleteBusiness(business.id);
                            } else if (actionType === 'assign' && params[0] === 'package') {
                              const packageId = parseInt(params[1]);
                              handleAssignPackage(business.id, packageId);
                            } else if (actionType === 'remove' && params[0] === 'package') {
                              handleRemovePackage(business.id);
                            }
                          }}>
                            <SelectTrigger className="w-8 h-8">
                              <MoreHorizontal className="h-3 w-3" />
                            </SelectTrigger>
                            <SelectContent>
                              {!business.package_id ? (
                                <>
                                  <SelectItem value="placeholder" disabled>
                                    Assign Package:
                                  </SelectItem>
                                  {packages.map((pkg) => (
                                    <SelectItem 
                                      key={`assign-package-${pkg.id}`}
                                      value={`assign-package-${pkg.id}`}
                                    >
                                      {pkg.name} (${pkg.price})
                                    </SelectItem>
                                  ))}
                                </>
                              ) : (
                                <SelectItem value="remove-package">
                                  <div className="flex items-center gap-2">
                                    <PackageIcon className="h-3 w-3" />
                                    Remove Package
                                  </div>
                                </SelectItem>
                              )}
                              <SelectItem value="delete" className="text-red-600">
                                <div className="flex items-center gap-2">
                                  <Trash2 className="h-3 w-3" />
                                  Delete Business
                                </div>
                              </SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              {(businesses ?? []).length === 0 && !loading && (
                <div className="text-center py-8 text-muted-foreground">
                  No businesses found. Create your first business to get started.
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