'use client';

import { useState, useEffect } from 'react';
import { format } from 'date-fns';
import { overviewApi, DayOverview } from '@/lib/api/overview';
import { BranchQuotaStatus } from '@/lib/api/quota';

interface AllocationOverviewTableProps {
  date?: Date;
  onBranchClick?: (branchId: string, date: Date) => void;
}

export default function AllocationOverviewTable({ date, onBranchClick }: AllocationOverviewTableProps) {
  const [overview, setOverview] = useState<DayOverview | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedBranchId, setSelectedBranchId] = useState<string | null>(null);
  const [selectedDate, setSelectedDate] = useState<Date | null>(null);

  useEffect(() => {
    loadOverview();
  }, [date]);

  const loadOverview = async () => {
    try {
      setLoading(true);
      const dateStr = date ? format(date, 'yyyy-MM-dd') : undefined;
      const data = await overviewApi.getDayOverview(dateStr);
      setOverview(data);
    } catch (error) {
      console.error('Failed to load overview:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleBranchClick = (branchId: string, branchDate: Date) => {
    setSelectedBranchId(branchId);
    setSelectedDate(branchDate);
    if (onBranchClick) {
      onBranchClick(branchId, branchDate);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  if (!overview) {
    return (
      <div className="p-8 text-center text-gray-500">
        No overview data available
      </div>
    );
  }

  return (
    <div className="w-full">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-2xl font-bold">
          All Branches Overview - {format(new Date(overview.date), 'MMMM d, yyyy')}
        </h2>
        <div className="text-sm text-gray-600">
          {overview.branches_with_shortage} of {overview.total_branches} branches need staff
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full border-collapse border border-gray-300">
          <thead>
            <tr className="bg-gray-100">
              <th className="border border-gray-300 p-2 text-left sticky left-0 z-10 bg-gray-100">
                Branch
              </th>
              <th className="border border-gray-300 p-2 text-center">Designated</th>
              <th className="border border-gray-300 p-2 text-center">Available</th>
              <th className="border border-gray-300 p-2 text-center">Assigned</th>
              <th className="border border-gray-300 p-2 text-center">Required</th>
              <th className="border border-gray-300 p-2 text-center">Status</th>
            </tr>
          </thead>
          <tbody>
            {overview.branch_statuses.map((status) => {
              const hasShortage = status.total_required > 0;
              const fulfillmentRate = status.total_designated > 0 
                ? (status.total_assigned / status.total_designated) * 100 
                : 0;

              return (
                <tr
                  key={status.branch_id}
                  className={`hover:bg-gray-50 cursor-pointer ${
                    selectedBranchId === status.branch_id ? 'bg-blue-50' : ''
                  }`}
                  onClick={() => handleBranchClick(status.branch_id, new Date(status.date))}
                >
                  <td className="border border-gray-300 p-2 sticky left-0 z-10 bg-white font-medium">
                    {status.branch_name} ({status.branch_code})
                  </td>
                  <td className="border border-gray-300 p-2 text-center">
                    {status.total_designated}
                  </td>
                  <td className="border border-gray-300 p-2 text-center">
                    {status.total_available}
                  </td>
                  <td className="border border-gray-300 p-2 text-center">
                    {status.total_assigned}
                  </td>
                  <td className={`border border-gray-300 p-2 text-center font-semibold ${
                    hasShortage ? 'text-red-600' : 'text-green-600'
                  }`}>
                    {status.total_required}
                  </td>
                  <td className="border border-gray-300 p-2 text-center">
                    <span className={`px-2 py-1 rounded text-xs ${
                      hasShortage 
                        ? 'bg-red-100 text-red-800' 
                        : fulfillmentRate >= 100
                        ? 'bg-green-100 text-green-800'
                        : 'bg-yellow-100 text-yellow-800'
                    }`}>
                      {hasShortage ? 'Shortage' : fulfillmentRate >= 100 ? 'Full' : 'Partial'}
                    </span>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {selectedBranchId && selectedDate && (
        <div className="mt-6 p-4 bg-gray-50 rounded-lg">
          <h3 className="text-lg font-semibold mb-2">Position Details</h3>
          {overview.branch_statuses
            .find(s => s.branch_id === selectedBranchId)
            ?.position_statuses.map((posStatus) => (
              <div key={posStatus.position_id} className="mb-2 text-sm">
                <span className="font-medium">{posStatus.position_name}:</span>{' '}
                Designated: {posStatus.designated_quota}, 
                Available: {posStatus.available_local}, 
                Rotation: {posStatus.assigned_rotation}, 
                Required: <span className={posStatus.still_required > 0 ? 'text-red-600 font-semibold' : 'text-green-600'}>
                  {posStatus.still_required}
                </span>
              </div>
            ))}
        </div>
      )}
    </div>
  );
}
