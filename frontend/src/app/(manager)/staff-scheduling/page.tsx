'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { authApi, User } from '@/lib/api/auth';
import { branchApi, Branch } from '@/lib/api/branch';
import MonthlyCalendar from '@/components/scheduling/MonthlyCalendar';
import AppLayout from '@/components/layout/AppLayout';

export default function StaffSchedulingPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [selectedBranchId, setSelectedBranchId] = useState<string>('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const userData = await authApi.getMe();
        setUser(userData);
        
        // Load branches
        const branchesData = await branchApi.list();
        setBranches(branchesData || []);
        
        // If user is branch manager, set their branch
        if (userData.role === 'branch_manager' && branchesData.length > 0) {
          // Note: In a real app, you'd get the branch_id from the user object
          setSelectedBranchId(branchesData[0].id);
        }
      } catch (error) {
        router.push('/login');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
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
          <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">Staff Scheduling</h1>
          <p className="text-sm text-neutral-text-secondary">Manage staff work schedules by branch</p>
        </div>

        {user?.role === 'admin' || user?.role === 'area_manager' || user?.role === 'district_manager' ? (
          <div className="mb-6">
            <label htmlFor="branch-select" className="block text-sm font-medium text-neutral-text-primary mb-2">
              Select Branch
            </label>
            <select
              id="branch-select"
              value={selectedBranchId}
              onChange={(e) => setSelectedBranchId(e.target.value)}
              className="input-field w-auto min-w-[250px]"
            >
              <option value="">-- Select a branch --</option>
              {branches.map((branch) => (
                <option key={branch.id} value={branch.id}>
                  {branch.name} ({branch.code})
                </option>
              ))}
            </select>
          </div>
        ) : null}

        {selectedBranchId ? (
          <div className="card">
            <MonthlyCalendar branchId={selectedBranchId} />
          </div>
        ) : (
          <div className="card p-8 text-center">
            <p className="text-neutral-text-secondary">Please select a branch to view the schedule</p>
          </div>
        )}
      </div>
    </AppLayout>
  );
}

