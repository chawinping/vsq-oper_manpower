'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/contexts/UserContext';
import RotationStaffCalendar from '@/components/rotation/RotationStaffCalendar';

export default function RotationSchedulingPage() {
  const router = useRouter();
  const { user, loading } = useUser();

  useEffect(() => {
    // Check if user has permission - only admin, area_manager, and district_manager can view rotation staff schedules
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
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">
          Rotation Staff Scheduling
        </h1>
        <p className="text-sm text-neutral-text-secondary">
          Set leave days, working days, and schedule status for rotation staff
        </p>
      </div>

      <RotationStaffCalendar />
    </div>
  );
}

