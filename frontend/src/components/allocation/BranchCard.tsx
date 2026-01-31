'use client';

import { useMemo, memo } from 'react';
import { Branch } from '@/lib/api/branch';

interface BranchSummary {
  branch: Branch;
  currentStaffCount: number;
  preferredStaffCount: number;
  minimumStaffCount: number;
  // Doctors assigned to this branch on this date
  doctors: Array<{ id: string; name: string; code: string }>;
  // Scoring group points and missing staff
  group1Score: number;
  group2Score: number;
  group3Score: number;
  group1MissingStaff: string[];
  group2MissingStaff: string[];
  group3MissingStaff: string[];
}

interface BranchCardProps {
  branch: Branch;
  summary?: BranchSummary;
  onClick: () => void;
}

type PriorityLevel = 'high' | 'medium' | 'low';

// Memoized component for rendering missing staff list
const MissingStaffList = memo(({ staff, groupKey }: { staff: string[]; groupKey: string }) => {
  const displayStaff = useMemo(() => staff.slice(0, 8), [staff]);
  const remainingCount = useMemo(() => Math.max(0, staff.length - 8), [staff.length]);

  if (staff.length === 0) {
    return <div className="text-gray-400 text-xs">-</div>;
  }

  return (
    <>
      {Array.from({ length: Math.ceil(displayStaff.length / 2) }).map((_, rowIdx) => {
        const startIdx = rowIdx * 2;
        const endIdx = Math.min(startIdx + 2, displayStaff.length);
        return (
          <div key={`${groupKey}-row-${rowIdx}`} className="grid grid-cols-2 gap-x-1">
            {displayStaff.slice(startIdx, endIdx).map((nickname, idx) => (
              <span key={`${groupKey}-${startIdx + idx}`} className="truncate">
                {nickname}
              </span>
            ))}
          </div>
        );
      })}
      {remainingCount > 0 && (
        <div className="text-gray-400 text-xs mt-1">
          +{remainingCount} more
        </div>
      )}
    </>
  );
});

MissingStaffList.displayName = 'MissingStaffList';

const BranchCard = memo(function BranchCard({ branch, summary, onClick }: BranchCardProps) {
  const getStaffRatio = () => {
    if (!summary) return 'N/A';
    return `${summary.currentStaffCount}/${summary.preferredStaffCount}`;
  };

  const getPriorityLevel = (): PriorityLevel | null => {
    if (!summary) return null;
    
    const { currentStaffCount, preferredStaffCount, minimumStaffCount } = summary;
    
    // High priority: Below minimum staff (critical)
    if (currentStaffCount < minimumStaffCount) {
      return 'high';
    }
    
    // Medium priority: Below preferred but above minimum
    if (currentStaffCount < preferredStaffCount) {
      return 'medium';
    }
    
    // Low priority: Meets or exceeds preferred
    return 'low';
  };

  const getPriorityBadge = () => {
    const priority = getPriorityLevel();
    if (!priority) return null;

    const badgeStyles = {
      high: 'bg-red-100 text-red-800 border-red-300',
      medium: 'bg-yellow-100 text-yellow-800 border-yellow-300',
      low: 'bg-green-100 text-green-800 border-green-300',
    };

    const badgeIcons = {
      high: '🔴',
      medium: '🟡',
      low: '🟢',
    };

    const badgeLabels = {
      high: 'High',
      medium: 'Medium',
      low: 'Low',
    };

    return (
      <div className={`inline-flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium border ${badgeStyles[priority]}`}>
        <span>{badgeIcons[priority]}</span>
        <span>{badgeLabels[priority]}</span>
      </div>
    );
  };

  const hasDoctors = summary && summary.doctors && summary.doctors.length > 0;
  const isBranchOff = summary && (!summary.doctors || summary.doctors.length === 0);

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

      {/* Doctors Section */}
      {summary && (
        <div className="mb-3">
          {hasDoctors ? (
            <div>
              <div className="text-xs font-semibold text-gray-800 mb-1">
                👨‍⚕️ Doctors:
              </div>
              <div className="space-y-1">
                {summary.doctors.map((doctor) => (
                  <div key={doctor.id} className="text-xs text-gray-700">
                    {doctor.name} {doctor.code && `(${doctor.code})`}
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <div className="text-sm text-gray-500 italic mb-2">
              ⚠️ Branch is off
            </div>
          )}
        </div>
      )}

      {/* Priority Badge - Only show if branch has doctors */}
      {summary && hasDoctors && (
        <div className="flex items-center justify-end mb-3">
          {getPriorityBadge()}
        </div>
      )}

      {/* Three Column Layout: Group 1 | Group 2 | Group 3 - Only show if branch has doctors */}
      {summary && hasDoctors && (
        <div className="grid grid-cols-3 gap-2 mb-3 border-t border-gray-200 pt-2">
          {/* Column 1: Group 1 */}
          <div className="flex flex-col">
            <div className="text-xs font-semibold text-gray-800 mb-1">
              Group 1
            </div>
            <div className="text-sm font-bold text-gray-900 mb-1">
              {summary.group1Score}
            </div>
            <div className="space-y-0.5 text-xs text-gray-600">
              <MissingStaffList staff={summary.group1MissingStaff} groupKey="g1" />
            </div>
          </div>

          {/* Column 2: Group 2 */}
          <div className="flex flex-col border-l border-r border-gray-200 px-2">
            <div className="text-xs font-semibold text-gray-800 mb-1">
              Group 2
            </div>
            <div className="text-sm font-bold text-gray-900 mb-1">
              {summary.group2Score}
            </div>
            <div className="space-y-0.5 text-xs text-gray-600">
              <MissingStaffList staff={summary.group2MissingStaff} groupKey="g2" />
            </div>
          </div>

          {/* Column 3: Group 3 */}
          <div className="flex flex-col">
            <div className="text-xs font-semibold text-gray-800 mb-1">
              Group 3
            </div>
            <div className="text-sm font-bold text-gray-900 mb-1">
              {summary.group3Score}
            </div>
            <div className="space-y-0.5 text-xs text-gray-600">
              <MissingStaffList staff={summary.group3MissingStaff} groupKey="g3" />
            </div>
          </div>
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
});

BranchCard.displayName = 'BranchCard';

export default BranchCard;
