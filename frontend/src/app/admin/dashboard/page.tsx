import { StatsCards } from '@/components/dashboard/stats-cards';
import { UserManagement } from '@/components/dashboard/user-management';

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">Welcome to your admin dashboard</p>
      </div>
      
      <StatsCards />
      
    </div>
  );
}