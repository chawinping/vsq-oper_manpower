'use client';

import { useState } from 'react';
import { Branch } from '@/lib/api/branch';

interface BranchSelectorProps {
  branches: Branch[];
  allBranchesSelected: boolean;
  selectedBranchIds: string[];
  onAllBranchesToggle: (checked: boolean) => void;
  onBranchIdsChange: (ids: string[]) => void;
}

export default function BranchSelector({
  branches,
  allBranchesSelected,
  selectedBranchIds,
  onAllBranchesToggle,
  onBranchIdsChange,
}: BranchSelectorProps) {
  const [showDialog, setShowDialog] = useState(false);
  const [tempSelectedIds, setTempSelectedIds] = useState<string[]>(selectedBranchIds);

  const handleDialogOpen = () => {
    setTempSelectedIds(selectedBranchIds);
    setShowDialog(true);
  };

  const handleDialogClose = () => {
    setShowDialog(false);
  };

  const handleApply = () => {
    onBranchIdsChange(tempSelectedIds);
    setShowDialog(false);
  };

  const handleToggleBranch = (branchId: string) => {
    if (tempSelectedIds.includes(branchId)) {
      setTempSelectedIds(tempSelectedIds.filter(id => id !== branchId));
    } else {
      setTempSelectedIds([...tempSelectedIds, branchId]);
    }
  };

  const handleSelectAll = () => {
    if (tempSelectedIds.length === branches.length) {
      setTempSelectedIds([]);
    } else {
      setTempSelectedIds(branches.map(b => b.id));
    }
  };

  return (
    <>
      <div className="flex items-center gap-2">
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={allBranchesSelected}
            onChange={(e) => onAllBranchesToggle(e.target.checked)}
            className="w-4 h-4"
          />
          <span className="text-sm font-medium">All ({branches.length})</span>
        </label>
        {!allBranchesSelected && (
          <button
            onClick={handleDialogOpen}
            className="px-3 py-2 border border-gray-300 rounded-md hover:bg-gray-50 text-sm"
          >
            Select Branches ({selectedBranchIds.length})
          </button>
        )}
      </div>

      {/* Branch Selection Dialog */}
      {showDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md max-h-[80vh] flex flex-col">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">Select Branches</h3>
              <button
                onClick={handleDialogClose}
                className="text-gray-400 hover:text-gray-600 text-2xl font-bold"
              >
                ×
              </button>
            </div>
            
            <div className="mb-4">
              <input
                type="text"
                placeholder="Search branches..."
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                onChange={(e) => {
                  // Simple search filter - can be enhanced
                  const search = e.target.value.toLowerCase();
                  // This is a simple implementation - you might want to filter the list
                }}
              />
            </div>

            <div className="flex-1 overflow-y-auto mb-4 border border-gray-200 rounded-md p-2">
              <div className="mb-2">
                <button
                  onClick={handleSelectAll}
                  className="text-sm text-blue-600 hover:text-blue-800"
                >
                  {tempSelectedIds.length === branches.length ? 'Deselect All' : 'Select All'}
                </button>
              </div>
              <div className="space-y-1">
                {branches.map(branch => (
                  <label
                    key={branch.id}
                    className="flex items-center gap-2 p-2 hover:bg-gray-50 cursor-pointer rounded"
                  >
                    <input
                      type="checkbox"
                      checked={tempSelectedIds.includes(branch.id)}
                      onChange={() => handleToggleBranch(branch.id)}
                      className="w-4 h-4"
                    />
                    <span className="text-sm">
                      {branch.code} - {branch.name}
                    </span>
                  </label>
                ))}
              </div>
            </div>

            <div className="flex gap-2 justify-end">
              <button
                onClick={handleDialogClose}
                className="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-md text-sm"
              >
                Cancel
              </button>
              <button
                onClick={handleApply}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
              >
                Apply ({tempSelectedIds.length})
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
