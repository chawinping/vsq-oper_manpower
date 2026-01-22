'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/contexts/UserContext';
import RotationStaffList from '@/components/rotation/RotationStaffList';

export default function RotationStaffProfilePage() {
  const router = useRouter();
  const { user, loading } = useUser();

  useEffect(() => {
    // Check if user has permission - only admin, area_manager, and district_manager can view rotation staff profiles
    if (!loading && user && !['admin', 'area_manager', 'district_manager'].includes(user.role)) {
      router.push('/dashboard');
    }
  }, [user, loading, router]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-neutral-text-secondary">Loading...</div>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-4">
      <div className="mb-4">
        <h1 className="text-2xl font-semibold text-neutral-text-primary inline">
          Rotation Staff Profile
          <span className="text-sm font-normal text-neutral-text-secondary ml-3">
            Manage rotation staff profiles, zones, branches, and travel parameters
          </span>
        </h1>
      </div>

      {/* Rotation Staff List Table */}
      <RotationStaffList />
    </div>
  );
}
