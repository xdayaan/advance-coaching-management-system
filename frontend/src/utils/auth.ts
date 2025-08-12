import { User } from "@/types/user";
// utils/auth.ts
export const getRedirectPath = (user: User): string => {
  switch (user.role) {
    case 'admin':
      return '/dashboard/admin';
    
    case 'business':
      return `/business/${user.businessSlug}/dashboard`;
    
    case 'teacher':
      return `/business/${user.businessSlug}/teacher/dashboard`;
    
    case 'student':
      return `/business/${user.businessSlug}/student/dashboard`;
    
    default:
      return '/';
  }
};