'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { toast } from 'sonner';
import { Loader2 } from 'lucide-react';
import { apiService } from '@/services/api';

export function LoginForm() {
  const [formData, setFormData] = useState({
    email: '',
    password: ''
  });
  const [loading, setLoading] = useState(false);
  const [redirecting, setRedirecting] = useState(false);
  const router = useRouter();

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.email || !formData.password) {
      toast.error('Please fill in all fields');
      return;
    }

    setLoading(true);

    try {
      const response = await apiService.login(formData);
      
      if (response.user) {
        const { role, businessSlug } = response.user;
        
        // Show role-specific success message
        const roleMessages = {
          admin: 'Welcome back, Admin! Redirecting to admin dashboard...',
          business: 'Welcome back! Redirecting to your business dashboard...',
          teacher: 'Welcome back, Teacher! Redirecting to teacher dashboard...',
          student: 'Welcome back, Student! Redirecting to student dashboard...'
        };
        
        toast.success(roleMessages[role as keyof typeof roleMessages] || 'Login successful! Redirecting...');
        
        // Show redirecting state
        setRedirecting(true);
        
        // Small delay to show the success message before redirect
        setTimeout(() => {
          const dashboardRoute = response.dashboardRoute || getDashboardRoute(response.user);
          router.push(dashboardRoute);
        }, 1000);
      }
    } catch (error: any) {
      console.error('Login error:', error);
      
      // Handle specific error cases
      if (error.response?.status === 401) {
        toast.error('Invalid email or password');
      } else if (error.response?.status === 403) {
        toast.error('Account access denied. Please contact support.');
      } else if (error.response?.status === 422) {
        toast.error('Please check your email format');
      } else if (error.response?.data?.message) {
        toast.error(error.response.data.message);
      } else if (error.code === 'NETWORK_ERROR') {
        toast.error('Network error. Please check your connection.');
      } else {
        toast.error('Login failed. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  // Helper function to determine dashboard route
  const getDashboardRoute = (user: any) => {
    const { role, businessSlug } = user;

    switch (role) {
      case 'admin':
        return '/dashboard/admin';
      case 'business':
        return `/business/${businessSlug}/dashboard`;
      case 'teacher':
        return `/business/${businessSlug}/teacher/dashboard`;
      case 'student':
        return `/business/${businessSlug}/student/dashboard`;
      default:
        return '/dashboard/login';
    }
  };

  const isFormDisabled = loading || redirecting;

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl font-bold text-center">Login</CardTitle>
          <CardDescription className="text-center">
            Enter your credentials to access your dashboard
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                name="email"
                type="email"
                value={formData.email}
                onChange={handleInputChange}
                placeholder="Enter your email"
                required
                disabled={isFormDisabled}
                autoComplete="email"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                name="password"
                type="password"
                value={formData.password}
                onChange={handleInputChange}
                placeholder="Enter your password"
                required
                disabled={isFormDisabled}
                autoComplete="current-password"
              />
            </div>
            <Button type="submit" className="w-full" disabled={isFormDisabled}>
              {loading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Signing in...
                </>
              ) : redirecting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Redirecting...
                </>
              ) : (
                'Sign in'
              )}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}