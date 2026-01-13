'use client';

import { useEffect, useState, useRef, useMemo } from 'react';
import { useUser } from '@/contexts/UserContext';
import { branchApi, Branch } from '@/lib/api/branch';
import MonthlyCalendar from '@/components/scheduling/MonthlyCalendar';

// Helper function to compare arrays
const arraysEqual = (a: string[], b: string[]): boolean => {
  if (a.length !== b.length) return false;
  return a.every((val, index) => val === b[index]);
};

export default function StaffSchedulingPage() {
  const { user } = useUser();
  const [branches, setBranches] = useState<Branch[]>([]);
  const [selectedBranchIds, setSelectedBranchIds] = useState<string[]>([]);
  const [showAllBranches, setShowAllBranches] = useState(false);
  const [loading, setLoading] = useState(true);
  const initializedRef = useRef(false);

  // Memoize filtered branches to prevent unnecessary re-renders
  const filteredBranches = useMemo(() => {
    return branches.filter(b => selectedBranchIds.includes(b.id));
  }, [branches, selectedBranchIds]);

  useEffect(() => {
    // Reset initialization flag when user changes
    initializedRef.current = false;
    
    const fetchData = async () => {
      try {
        // Load branches
        const branchesData = await branchApi.list();
        setBranches(branchesData || []);
        
        // If user is branch manager, set their branch
        if (user?.role === 'branch_manager') {
          const newBranchIds = user.branch_id 
            ? [user.branch_id]
            : (() => {
                // Fallback: try to find branch by matching username pattern
                const branchCode = user.username.toLowerCase().replace(/mgr$|amgr$/, '');
                const userBranch = branchesData?.find(b => b.code.toLowerCase() === branchCode);
                return userBranch ? [userBranch.id] : [];
              })();
          
          // Only update if different
          setSelectedBranchIds(prev => {
            if (!arraysEqual(prev, newBranchIds)) {
              return newBranchIds;
            }
            return prev;
          });
        } else if (user?.role === 'admin' || user?.role === 'district_manager') {
          // For admin and district manager, default to showing all branches
          const allBranchIds = branchesData?.map(b => b.id) || [];
          
          // Only update state if not already initialized or if values changed
          if (!initializedRef.current) {
            setShowAllBranches(true);
            setSelectedBranchIds(allBranchIds);
            initializedRef.current = true;
          } else {
            // Only update if branch IDs actually changed
            setSelectedBranchIds(prev => {
              if (!arraysEqual(prev, allBranchIds)) {
                return allBranchIds;
              }
              return prev;
            });
          }
        }
      } catch (error: any) {
        console.error('Failed to fetch data:', error);
      } finally {
        setLoading(false);
      }
    };

    if (user) {
      fetchData();
    }
  }, [user]);

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
        <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">Staff Scheduling</h1>
        <p className="text-sm text-neutral-text-secondary">Manage staff work schedules by branch</p>
      </div>

      {user?.role === 'admin' || user?.role === 'district_manager' ? (
        <div className="mb-6">
          <div className="flex items-center gap-4 mb-4">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={showAllBranches}
                onChange={(e) => {
                  setShowAllBranches(e.target.checked);
                  if (e.target.checked) {
                    setSelectedBranchIds(branches.map(b => b.id));
                  } else {
                    setSelectedBranchIds([]);
                  }
                }}
                className="w-4 h-4"
              />
              <span className="text-sm font-medium text-neutral-text-primary">Show All Branches</span>
            </label>
          </div>
          <div>
            <label htmlFor="branch-select" className="block text-sm font-medium text-neutral-text-primary mb-2">
              Select Branches (can select multiple)
            </label>
            <select
              id="branch-select"
              multiple
              value={selectedBranchIds}
              onChange={(e) => {
                const selected = Array.from(e.target.selectedOptions, option => option.value);
                setSelectedBranchIds(selected);
                setShowAllBranches(selected.length === branches.length);
              }}
              className="input-field w-full min-w-[300px] min-h-[150px]"
              disabled={showAllBranches}
            >
              {(branches || []).map((branch) => (
                <option key={branch.id} value={branch.id}>
                  {branch.code} - {branch.name}
                </option>
              ))}
            </select>
            <p className="mt-2 text-xs text-neutral-text-secondary">
              {showAllBranches 
                ? `Showing all ${branches.length} branches` 
                : `${selectedBranchIds.length} branch${selectedBranchIds.length !== 1 ? 'es' : ''} selected`}
            </p>
          </div>
        </div>
      ) : user?.role === 'area_manager' ? (
        <div className="mb-6">
          <label htmlFor="branch-select" className="block text-sm font-medium text-neutral-text-primary mb-2">
            Select Branch
          </label>
          <select
            id="branch-select"
            value={selectedBranchIds[0] || ''}
            onChange={(e) => setSelectedBranchIds(e.target.value ? [e.target.value] : [])}
            className="input-field w-auto min-w-[250px]"
          >
            <option value="">-- Select a branch --</option>
            {(branches || []).map((branch) => (
              <option key={branch.id} value={branch.id}>
                {branch.name} ({branch.code})
              </option>
            ))}
          </select>
        </div>
      ) : user?.role === 'branch_manager' ? (
        <div className="mb-6">
          {selectedBranchIds[0] && (() => {
            const branch = branches.find(b => b.id === selectedBranchIds[0]);
            return branch ? (
              <div>
                <h2 className="text-lg font-semibold text-neutral-text-primary mb-1">
                  {branch.name} ({branch.code})
                </h2>
                <p className="text-sm text-neutral-text-secondary">
                  Managing schedule for your branch
                </p>
              </div>
            ) : (
              <p className="text-sm text-neutral-text-secondary">
                Managing schedule for your branch
              </p>
            );
          })()}
        </div>
      ) : null}

      {selectedBranchIds.length > 0 ? (
        <div className="card">
          <MonthlyCalendar branchIds={selectedBranchIds} branches={filteredBranches} />
        </div>
      ) : (
        <div className="card p-8 text-center">
          <p className="text-neutral-text-secondary">Please select at least one branch to view the schedule</p>
        </div>
      )}
    </div>
  );
}

