export interface Package {
  id: number;
  name: string;
  price: number;
  validation_period: number;
  description?: string;
  status: number; // 1 = active, 0 = inactive
  created_on: string;
}

export interface CreatePackageRequest {
  name: string;
  price: number;
  validation_period: number;
  description?: string;
}

export interface UpdatePackageRequest {
  name?: string;
  price?: number;
  validation_period?: number;
  description?: string;
  status?: number;
}

export interface PackageFilters {
  page?: number;
  limit?: number;
  status?: number;
  min_price?: number;
  max_price?: number;
  min_period?: number;
  max_period?: number;
  search?: string;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface PackageStats {
  total_packages: number;
  active_packages: number;
  inactive_packages: number;
}

export interface PriceStatistics {
  min_price: number;
  max_price: number;
  avg_price: number;
}

export interface PackagesResponse {
  data: Package[];
  pagination: {
    has_next: boolean;
    has_prev: boolean;
    limit: number;
    page: number;
    total: number;
    total_pages: number;
  };
}