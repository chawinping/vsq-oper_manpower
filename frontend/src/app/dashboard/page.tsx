'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { authApi } from '@/lib/api/auth';
import { User } from '@/lib/api/auth';
import AppLayout from '@/components/layout/AppLayout';

export default function DashboardPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const userData = await authApi.getMe();
        setUser(userData);
      } catch (error) {
        router.push('/login');
      } finally {
        setLoading(false);
      }
    };

    fetchUser();
  }, [router]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-neutral-text-secondary">Loading...</div>
      </div>
    );
  }

  return (
    <AppLayout>
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">Dashboard</h1>
          <p className="text-sm text-neutral-text-secondary">Welcome back, {user?.username}</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
          <div className="card p-6">
            <h3 className="text-sm font-medium text-neutral-text-secondary mb-2">Total Staff</h3>
            <p className="text-2xl font-semibold text-neutral-text-primary">-</p>
          </div>
          <div className="card p-6">
            <h3 className="text-sm font-medium text-neutral-text-secondary mb-2">Active Branches</h3>
            <p className="text-2xl font-semibold text-neutral-text-primary">-</p>
          </div>
          <div className="card p-6">
            <h3 className="text-sm font-medium text-neutral-text-secondary mb-2">This Month</h3>
            <p className="text-2xl font-semibold text-neutral-text-primary">-</p>
          </div>
        </div>

        <div className="card p-6">
          <h2 className="text-lg font-semibold text-neutral-text-primary mb-4">Recent Activity</h2>
          <p className="text-sm text-neutral-text-secondary">Dashboard content will be implemented here.</p>
        </div>
      </div>
    </AppLayout>
  );
}

