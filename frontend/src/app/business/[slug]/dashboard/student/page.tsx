import { apiService } from '@/services/api';
import { Business } from '@/types/business';
import StudentManagement from '@/components/business/StudentManagement';
import React from 'react';

interface BusinessPageProps {
  params: Promise<{ slug: string }>;
}

export default async function StudentManagementPage({ params }: BusinessPageProps) {
  let business: Business | null = null;
  let error = '';
  
  // Await params before accessing its properties
  const { slug } = await params;
  
  try {
    business = await apiService.businesses.getBusinessBySlug(slug);
  } catch (e: any) {
    console.log(e);
    error = e?.response?.data?.message || 'Business not found';
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-red-600 mb-4">Error</h1>
          <p className="text-gray-600">{error}</p>
        </div>
      </div>
    );
  }

  if (!business) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 px-4">
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Student Management</h1>
        <p className="text-gray-600 mt-2">
          Manage students for <span className="font-semibold">{business.name}</span>
        </p>
      </div>
      
      <StudentManagement businessId={business.id} />
    </div>
  );
}