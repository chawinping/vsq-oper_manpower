'use client';

import { Branch } from '@/lib/api/branch';

interface BranchSummary {
  branch: Branch;
  currentStaffCount: number;
  preferredStaffCount: number;
  minimumStaffCount: number;
}

interface BranchCardProps {
  branch: Branch;
  summary?: BranchSummary;
  onClick: () => void;
}

export default function BranchCard({ branch, summary, onClick }: BranchCardProps) {
  const getStaffRatio = () => {
    if (!summary) return 'N/A';
    return `${summary.currentStaffCount}/${summary.preferredStaffCount}`;
  };

  return (
    <div
      onClick={onClick}
      className="bg-white border border-gray-200 rounded-lg p-4 cursor-pointer hover:shadow-md hover:border-blue-300 transition-all duration-200"
    >
      {/* Branch Code - Large */}
      <div className="text-xl font-bold text-gray-900 mb-1">
        {branch.code}
      </div>
      
      {/* Branch Name - Small */}
      <div className="text-sm text-gray-600 mb-3 line-clamp-2">
        {branch.name}
      </div>

      {/* Staff Ratio */}
      {summary && (
        <div className="text-sm font-medium text-gray-700 mb-2">
          Staff: {getStaffRatio()}
        </div>
      )}

      {/* Add Staff Button */}
      <button
        onClick={(e) => {
          e.stopPropagation();
          onClick();
        }}
        className="w-full px-3 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 transition-colors"
      >
        + Add Staff
      </button>
    </div>
  );
}
