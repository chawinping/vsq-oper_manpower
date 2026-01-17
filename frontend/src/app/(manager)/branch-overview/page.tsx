'use client';

import { useState, useEffect } from 'react';
import { useUser } from '@/contexts/UserContext';
import { branchApi, Branch } from '@/lib/api/branch';
import BranchOverviewCalendar from '@/components/allocation/BranchOverviewCalendar';

export default function BranchOverviewPage() {
  const { user } = useUser();
  const [branch, setBranch] = useState<Branch | null>(null);
  const [year, setYear] = useState(new Date().getFullYear());
  const [month, setMonth] = useState(new Date().getMonth() + 1);

  useEffect(() => {
    loadBranch();
  }, []);

  const loadBranch = async () => {
    if (user?.branch_id) {
      try {
        const branches = await branchApi.list();
        const userBranch = branches.find(b => b.id === user.branch_id);
        if (userBranch) {
          setBranch(userBranch);
        }
      } catch (error) {
        console.error('Failed to load branch:', error);
      }
    }
  };

  if (!branch) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading branch information...</div>
      </div>
    );
  }

  return (
    <div className="w-full p-6">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-3xl font-bold">Branch Overview</h1>
        <div className="flex gap-2">
          <input
            type="month"
            value={`${year}-${String(month).padStart(2, '0')}`}
            onChange={(e) => {
              const [y, m] = e.target.value.split('-').map(Number);
              setYear(y);
              setMonth(m);
            }}
            className="px-3 py-2 border border-gray-300 rounded-md"
          />
        </div>
      </div>

      <BranchOverviewCalendar branchId={branch.id} year={year} month={month} />
    </div>
  );
}
