'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { authApi, User } from '@/lib/api/auth';
import RotationAssignmentView from '@/components/rotation/RotationAssignmentView';
import AppLayout from '@/components/layout/AppLayout';

export default function RotationSchedulingPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const userData = await authApi.getMe();
        setUser(userData);
        
        // Check if user has permission
        if (!['admin', 'area_manager', 'district_manager'].includes(userData.role)) {
          router.push('/dashboard');
        }
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
          <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">Rotation Staff Scheduling</h1>
          <p className="text-sm text-neutral-text-secondary">Assign rotation staff to branches</p>
        </div>

        <div className="card">
          <RotationAssignmentView />
        </div>
      </div>
    </AppLayout>
  );
}

