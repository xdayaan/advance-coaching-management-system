import { AxiosInstance } from 'axios';
import { Package, CreatePackageRequest, UpdatePackageRequest, PackageFilters, PackageStats, PriceStatistics } from '@/types/package';

export class PackageService {
  private api: AxiosInstance;

  constructor(api: AxiosInstance) {
    this.api = api;
  }

  async getPackages(filters: PackageFilters = {}): Promise<{
    data: Package[];
    pagination: any;
  }> {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        params.append(key, value.toString());
      }
    });

    const response = await this.api.get(`/packages?${params}`);
    return response.data;
  }

  async getPackage(id: number): Promise<Package> {
    const response = await this.api.get(`/packages/${id}`);
    return response.data.data;
  }

  async createPackage(packageData: CreatePackageRequest): Promise<Package> {
    const response = await this.api.post('/packages', packageData);
    return response.data.data;
  }

  async updatePackage(id: number, packageData: Partial<UpdatePackageRequest>): Promise<Package> {
    const response = await this.api.put(`/packages/${id}`, packageData);
    return response.data.data;
  }

  async deletePackage(id: number): Promise<void> {
    await this.api.delete(`/packages/${id}`);
  }

  async changePackageStatus(id: number, status: number): Promise<void> {
    await this.api.patch(`/packages/${id}/status`, { status });
  }

  async getActivePackages(): Promise<Package[]> {
    const response = await this.api.get('/packages/active');
    return response.data.data;
  }

  async getInactivePackages(): Promise<Package[]> {
    const response = await this.api.get('/packages/inactive');
    return response.data.data;
  }

  async getPackageStats(): Promise<PackageStats> {
    const response = await this.api.get('/packages/stats');
    return response.data.data;
  }

  async getPriceStatistics(): Promise<PriceStatistics> {
    const response = await this.api.get('/packages/stats/prices');
    return response.data.data;
  }

  async searchPackages(query: string, limit: number = 10): Promise<Package[]> {
    const response = await this.api.get(`/packages/search?q=${encodeURIComponent(query)}&limit=${limit}`);
    return response.data.data;
  }

  async getPackagesByPriceRange(minPrice?: number, maxPrice?: number): Promise<Package[]> {
    const params = new URLSearchParams();
    if (minPrice !== undefined) params.append('min_price', minPrice.toString());
    if (maxPrice !== undefined) params.append('max_price', maxPrice.toString());
    
    const response = await this.api.get(`/packages/price-range?${params}`);
    return response.data.data;
  }

  async getPackagesByValidationPeriod(minDays?: number, maxDays?: number): Promise<Package[]> {
    const params = new URLSearchParams();
    if (minDays !== undefined) params.append('min_period', minDays.toString());
    if (maxDays !== undefined) params.append('max_period', maxDays.toString());
    
    const response = await this.api.get(`/packages/period-range?${params}`);
    return response.data.data;
  }

  async bulkUpdatePackageStatus(packageIds: number[], status: number): Promise<void> {
    await this.api.patch('/packages/bulk/status', {
      package_ids: packageIds,
      status,
    });
  }

  async bulkDeletePackages(packageIds: number[]): Promise<void> {
    await this.api.delete('/packages/bulk', {
      data: { package_ids: packageIds }
    });
  }

  async getRecentPackages(limit: number = 10): Promise<Package[]> {
    const response = await this.api.get(`/packages/recent?limit=${limit}`);
    return response.data.data;
  }

  async duplicatePackage(id: number, newName: string): Promise<Package> {
    const response = await this.api.post(`/packages/${id}/duplicate`, { name: newName });
    return response.data.data;
  }

  async getPackagesByDateRange(startDate: string, endDate: string): Promise<Package[]> {
    const params = new URLSearchParams();
    params.append('start_date', startDate);
    params.append('end_date', endDate);
    
    const response = await this.api.get(`/packages/date-range?${params}`);
    return response.data.data;
  }

  async validatePackageName(name: string, excludeId?: number): Promise<{ available: boolean }> {
    const params = new URLSearchParams();
    params.append('name', name);
    if (excludeId) params.append('exclude_id', excludeId.toString());
    
    const response = await this.api.get(`/packages/validate-name?${params}`);
    return response.data;
  }

  async exportPackages(format: 'csv' | 'excel' = 'csv'): Promise<Blob> {
    const response = await this.api.get(`/packages/export?format=${format}`, {
      responseType: 'blob'
    });
    return response.data;
  }

  async importPackages(file: File): Promise<{ success: number; errors: string[] }> {
    const formData = new FormData();
    formData.append('file', file);
    
    const response = await this.api.post('/packages/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }
}