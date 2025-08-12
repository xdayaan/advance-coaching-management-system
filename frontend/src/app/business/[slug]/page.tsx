import { apiService } from '@/services/api';
import { Business } from '@/types/business';
import React from 'react';

interface BusinessPageProps {
  params: Promise<{ slug: string }>; // params is a Promise
}

export default async function BusinessPage({ params }: BusinessPageProps) {
  let business: Business | null = null;
  let error = '';
  
  // Await params before accessing its properties
  const { slug } = await params;
  
  try {
    business = await apiService.businesses.getBusinessBySlug(slug);
    console.log(business)
  } catch (e: any) {
    console.log(e)
    error = e?.response?.data?.message || 'Business not found';
  }

  if (error) {
    return <div className="p-8 text-red-600">{error}</div>;
  }

  if (!business) {
    return <div className="p-8">Loading...</div>;
  }

  return (
    <div className="p-8 max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-4">{business.name}</h1>
      <div className="mb-2"><strong>Owner:</strong> {business.owner_name}</div>
      <div className="mb-2"><strong>Email:</strong> {business.email}</div>
      <div className="mb-2"><strong>Phone:</strong> {business.phone}</div>
      <div className="mb-2"><strong>Location:</strong> {business.location}</div>
      <div className="mb-2"><strong>Status:</strong> {business.status === 1 ? 'Active' : 'Inactive'}</div>
      <div className="mb-2"><strong>Package:</strong> {business.package?.name || 'None'}</div>
      <div className="mb-2"><strong>Created On:</strong> {new Date(business.created_on).toLocaleString()}</div>
      {/* Add more fields as needed */}
    </div>
  );
}