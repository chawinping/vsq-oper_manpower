'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/contexts/UserContext';
import BranchRotationTable from '@/components/rotation/BranchRotationTable';
import RotationStaffList from '@/components/rotation/RotationStaffList';
import { Staff } from '@/lib/api/staff';

export default function RotationSchedulingPage() {
  const router = useRouter();
  const { user, loading } = useUser();
  const [selectedRotationStaff, setSelectedRotationStaff] = useState<Staff[]>([]);

  useEffect(() => {
    // Check if user has permission - only admin and area_manager can assign rotation staff
    if (!loading && user && !['admin', 'area_manager'].includes(user.role)) {
      router.push('/dashboard');
    }
  }, [user, loading, router]);

  const handleAddRotationStaff = (staff: Staff) => {
    // Add staff if not already in the list
    if (!selectedRotationStaff.find(s => s.id === staff.id)) {
      setSelectedRotationStaff([...selectedRotationStaff, staff]);
    }
  };

  const handleRemoveRotationStaff = (staffId: string) => {
    setSelectedRotationStaff(selectedRotationStaff.filter(s => s.id !== staffId));
  };

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
          Rotation Staff Scheduling
          <span className="text-sm font-normal text-neutral-text-secondary ml-3">Manage rotation staff and assign them to branches</span>
        </h1>
      </div>

      {/* Rotation Staff List Table */}
      <RotationStaffList 
        onAddToAssignment={handleAddRotationStaff}
        selectedStaffIds={selectedRotationStaff.map(s => s.id)}
      />

      {/* Branch Rotation Assignment Table */}
      <BranchRotationTable 
        manuallyAddedStaff={selectedRotationStaff}
        onRemoveStaff={handleRemoveRotationStaff}
      />
    </div>
  );
}

