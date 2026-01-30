'use client';

import { useState, useEffect, useMemo } from 'react';
import { format } from 'date-fns';
import { branchApi, Branch } from '@/lib/api/branch';
import BranchCard from '@/components/allocation/BranchCard';
import BranchDetailDrawer from '@/components/allocation/BranchDetailDrawer';
import FilterBar from '@/components/allocation/FilterBar';
import SummaryStats from '@/components/allocation/SummaryStats';
import BranchSelector from '@/components/allocation/BranchSelector';

interface BranchSummary {
  branch: Branch;
  currentStaffCount: number;
  preferredStaffCount: number;
  minimumStaffCount: number;
}

export default function AllocateStaffPage() {
  const [selectedDate, setSelectedDate] = useState(format(new Date(), 'yyyy-MM-dd'));
  const [allBranchesSelected, setAllBranchesSelected] = useState(false);
  const [selectedBranchIds, setSelectedBranchIds] = useState<string[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [branchSummaries, setBranchSummaries] = useState<Map<string, BranchSummary>>(new Map());
  const [selectedBranchId, setSelectedBranchId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [filter, setFilter] = useState({
    status: 'all' as 'all' | 'needs_attention' | 'critical' | 'ok',
    priority: 'all' as 'all' | 'high' | 'medium' | 'low',
    search: '',
  });

  useEffect(() => {
    loadBranches();
  }, []);

  useEffect(() => {
    if (branches.length > 0) {
      loadBranchSummaries();
    }
  }, [selectedDate, allBranchesSelected, selectedBranchIds, branches]);

  const loadBranches = async () => {
    try {
      const branchesData = await branchApi.list();
      setBranches(branchesData || []);
      // Default to no branches selected (empty array)
      // User can manually select branches or check "All Branches"
    } catch (error) {
      console.error('Failed to load branches:', error);
    }
  };

  const loadBranchSummaries = async () => {
    try {
      setLoading(true);
      const branchIds = allBranchesSelected 
        ? branches.map(b => b.id)
        : selectedBranchIds;

      if (branchIds.length === 0) {
        setBranchSummaries(new Map());
        return;
      }

      // Create summaries for selected branches
      // TODO: Load actual staff counts from API when available
      const summariesMap = new Map<string, BranchSummary>();
      
      branchIds.forEach(branchId => {
        const branch = branches.find(b => b.id === branchId);
        if (!branch) return;

        summariesMap.set(branchId, {
          branch,
          currentStaffCount: 0, // TODO: Get from API
          preferredStaffCount: 0, // TODO: Get from API
          minimumStaffCount: 0, // TODO: Get from API
        });
      });

      setBranchSummaries(summariesMap);
    } catch (error) {
      console.error('Failed to load branch summaries:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleAllBranchesToggle = (checked: boolean) => {
    setAllBranchesSelected(checked);
    if (checked) {
      setSelectedBranchIds(branches.map(b => b.id));
    } else {
      // When unchecking "All Branches", clear selection
      setSelectedBranchIds([]);
    }
  };

  const handleBranchSelect = (branchId: string, selected: boolean) => {
    if (selected) {
      setSelectedBranchIds([...selectedBranchIds, branchId]);
    } else {
      setSelectedBranchIds(selectedBranchIds.filter(id => id !== branchId));
    }
  };

  const filteredBranches = useMemo(() => {
    const displayBranches = allBranchesSelected 
      ? branches 
      : branches.filter(b => selectedBranchIds.includes(b.id));

    return displayBranches.filter(branch => {
      // Search filter
      if (filter.search) {
        const searchLower = filter.search.toLowerCase();
        return (
          branch.name.toLowerCase().includes(searchLower) ||
          branch.code.toLowerCase().includes(searchLower)
        );
      }

      return true;
    });
  }, [branches, allBranchesSelected, selectedBranchIds, filter]);

  const summaryStats = useMemo(() => {
    const total = filteredBranches.length;
    return { total, needsAttention: 0, critical: 0, ok: total };
  }, [filteredBranches]);

  const handleBranchClick = (branchId: string) => {
    setSelectedBranchId(branchId);
  };

  const handleCloseDrawer = () => {
    setSelectedBranchId(null);
  };

  const handleAssignmentSuccess = () => {
    loadBranchSummaries();
    setSelectedBranchId(null);
  };

  return (
    <div className="w-full p-6">
      <div className="mb-6">
        <h1 className="text-3xl font-bold mb-2">Allocate Staff</h1>
        <p className="text-gray-600">
          Allocate rotation staff to branches
        </p>
      </div>

      {/* Top Control Bar */}
      <div className="mb-4 flex flex-wrap gap-4 items-center">
        <div className="flex items-center gap-2">
          <label className="text-sm font-medium">📅 Date:</label>
          <input
            type="date"
            value={selectedDate}
            onChange={(e) => setSelectedDate(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md"
          />
        </div>
        <div className="flex items-center gap-2">
          <label className="text-sm font-medium">🌳 Branches:</label>
          <BranchSelector
            branches={branches}
            allBranchesSelected={allBranchesSelected}
            selectedBranchIds={selectedBranchIds}
            onAllBranchesToggle={handleAllBranchesToggle}
            onBranchIdsChange={setSelectedBranchIds}
          />
        </div>
      </div>

      {/* Summary Stats */}
      <SummaryStats stats={summaryStats} />

      {/* Filter Bar */}
      <FilterBar filter={filter} onFilterChange={setFilter} />

      {/* Branch Grid */}
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="text-lg text-gray-600">Loading branches...</div>
        </div>
      ) : (allBranchesSelected ? branches.length === 0 : selectedBranchIds.length === 0) ? (
        <div className="flex flex-col items-center justify-center py-12">
          <div className="text-lg text-gray-600 mb-2">No branches selected</div>
          <div className="text-sm text-gray-500">
            Please select branches using the branch selector above, or check "All Branches" to view all {branches.length} branches.
          </div>
        </div>
      ) : filteredBranches.length === 0 ? (
        <div className="flex items-center justify-center py-12">
          <div className="text-lg text-gray-600">No branches match the current filters</div>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4 mt-6">
          {filteredBranches.map(branch => {
            const summary = branchSummaries.get(branch.id);
            return (
              <BranchCard
                key={branch.id}
                branch={branch}
                summary={summary}
                onClick={() => handleBranchClick(branch.id)}
              />
            );
          })}
        </div>
      )}

      {/* Branch Detail Drawer */}
      {selectedBranchId && (
        <BranchDetailDrawer
          isOpen={!!selectedBranchId}
          branchId={selectedBranchId}
          date={selectedDate}
          onClose={handleCloseDrawer}
          onSuccess={handleAssignmentSuccess}
        />
      )}
    </div>
  );
}
