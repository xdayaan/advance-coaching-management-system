// types/business.ts
export interface Business {
  id: number;
  name: string;
  slug: string;
  user_id: number;
  owner_name: string;
  package_id?: number;
  email: string;
  phone: string;
  location: string;
  status: number;
  created_on: string;
  user?: {
    id: number;
    name: string;
    email: string;
    phone: string;
    role: string;
    status: number;
    created_on: string;
  };
  package?: {
    id: number;
    name: string;
    price: number;
    validation_period: number;
    description: string;
    status: number;
    created_on: string;
  };
}

export interface CreateBusinessRequest {
  name: string;
  slug?: string;
  owner_name: string;
  email: string;
  phone: string;
  location: string;
  password: string;
  package_id?: number;
}

export interface UpdateBusinessRequest {
  name?: string;
  slug?: string;
  owner_name?: string;
  email?: string;
  phone?: string;
  location?: string;
  password?: string;
  package_id?: number;
  status?: number;
}

export interface BusinessFilters {
  page?: number;
  limit?: number;
  status?: number;
  package_id?: number;
  location?: string;
  search?: string;
  sort_by?: string;
  sort_order?: string;
}

export interface BusinessStats {
  total_businesses: number;
  active_businesses: number;
  inactive_businesses: number;
  businesses_with_packages: number;
  businesses_without_packages: number;
}

export interface LocationStats {
  [location: string]: number;
}

export interface PackageDistribution {
  [packageName: string]: number;
}