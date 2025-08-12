// services/business.ts
import { AxiosInstance } from 'axios';
import { Business, CreateBusinessRequest, UpdateBusinessRequest, BusinessFilters, BusinessStats, LocationStats, PackageDistribution } from '@/types/business';

export class BusinessService {

  private api: AxiosInstance;

  constructor(api: AxiosInstance) {
    this.api = api;
  }


  private getAuthHeader() {
    const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
    return token ? { Authorization: `Bearer ${token}` } : {};
  }

  async getBusinessBySlug(slug: string): Promise<Business> {
    console.log(slug)
    const response = await this.api.get(`/business/${slug}`);
    console.log("data: ", response)
    return response.data.data;
  }

  async getBusinesses(filters: BusinessFilters = {}): Promise<{
    data: { businesses: Business[]; total: number; page: number; limit: number };
  }> {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        params.append(key, value.toString());
      }
    });

    const response = await this.api.get(`/businesses?${params}`, { headers: this.getAuthHeader() });
    return response.data;
  }

  async getBusiness(id: number): Promise<Business> {
    const response = await this.api.get(`/businesses/${id}`, { headers: this.getAuthHeader() });
    return response.data.data;
  }

  async createBusiness(businessData: CreateBusinessRequest): Promise<Business> {
    const response = await this.api.post('/businesses', businessData, { headers: this.getAuthHeader() });
    return response.data.data;
  }

  async updateBusiness(id: number, businessData: Partial<UpdateBusinessRequest>): Promise<Business> {
    const response = await this.api.put(`/businesses/${id}`, businessData, { headers: this.getAuthHeader() });
    return response.data.data;
  }

  async deleteBusiness(id: number): Promise<void> {
    await this.api.delete(`/businesses/${id}`, { headers: this.getAuthHeader() });
  }

  async changeBusinessStatus(id: number, status: number): Promise<void> {
    await this.api.patch(`/businesses/${id}/status`, { status }, { headers: this.getAuthHeader() });
  }

  async assignPackage(id: number, packageId: number): Promise<void> {
    await this.api.post(`/businesses/${id}/assign-package`, { package_id: packageId }, { headers: this.getAuthHeader() });
  }

  async removePackage(id: number): Promise<void> {
    await this.api.delete(`/businesses/${id}/remove-package`, { headers: this.getAuthHeader() });
  }

  async getActiveBusinesses(): Promise<Business[]> {
    const response = await this.api.get('/businesses/active', { headers: this.getAuthHeader() });
    return response.data.data;
  }

  async getInactiveBusinesses(): Promise<Business[]> {
    const response = await this.api.get('/businesses/inactive', { headers: this.getAuthHeader() });
    return response.data.data;
  }

  async getBusinessStats(): Promise<BusinessStats> {
    const response = await this.api.get('/businesses/stats', { headers: this.getAuthHeader() });
    return response.data.data;
  }

  async getLocationStats(): Promise<LocationStats> {
    const response = await this.api.get('/businesses/stats/locations', { headers: this.getAuthHeader() });
    return response.data.data;
  }

  async getPackageDistribution(): Promise<PackageDistribution> {
    const response = await this.api.get('/businesses/stats/packages', { headers: this.getAuthHeader() });
    return response.data.data;
  }

  async searchBusinesses(query: string, limit: number = 10): Promise<Business[]> {
    const response = await this.api.get(`/businesses/search?q=${encodeURIComponent(query)}&limit=${limit}`, { headers: this.getAuthHeader() });
    return response.data.data.businesses;
  }

  async getBusinessesByPackage(packageId: number): Promise<Business[]> {
    const response = await this.api.get(`/businesses/package/${packageId}`, { headers: this.getAuthHeader() });
    return response.data.data;
  }

  async getBusinessesWithoutPackage(): Promise<Business[]> {
    const response = await this.api.get('/businesses/no-package', { headers: this.getAuthHeader() });
    return response.data.data;
  }

  async getBusinessesByLocation(location: string): Promise<Business[]> {
    const response = await this.api.get(`/businesses/by-location?location=${encodeURIComponent(location)}`, { headers: this.getAuthHeader() });
    return response.data.data;
  }

  async getBusinessLocations(): Promise<string[]> {
    const response = await this.api.get('/businesses/locations', { headers: this.getAuthHeader() });
    return response.data.data;
  }

  async bulkUpdateStatus(businessIds: number[], status: number): Promise<void> {
    await this.api.post('/businesses/bulk/status', {
      business_ids: businessIds,
      status,
    }, { headers: this.getAuthHeader() });
  }

  async bulkAssignPackage(businessIds: number[], packageId: number): Promise<void> {
    await this.api.post('/businesses/bulk/assign-package', {
      business_ids: businessIds,
      package_id: packageId,
    }, { headers: this.getAuthHeader() });
  }
}