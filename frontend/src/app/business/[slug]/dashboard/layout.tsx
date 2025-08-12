'use client';

import Sidebar from '@/components/dashboard/business-sidebar';
import { Toaster } from '@/components/ui/sonner';
import { apiService } from '@/services/api';
import { redirect } from 'next/navigation';

interface BusinessDashboardLayoutProps {
  children: React.ReactNode;
  params: Promise<{ slug: string }>;
}

export default async function BusinessDashboardLayout({ children, params }: BusinessDashboardLayoutProps) {
  const { slug: businessSlug } = await params;

  // Server-side auth check (optional, can be improved)
  if (!apiService.isAuthenticated()) {
    redirect('/login');
  }

  return (
    <div className="flex h-screen bg-gray-50">
      <Sidebar businessSlug={businessSlug} />
      <main className="flex-1 overflow-auto">
        <div className="p-6">
          {children}
        </div>
      </main>
      <Toaster />
    </div>
  );
}