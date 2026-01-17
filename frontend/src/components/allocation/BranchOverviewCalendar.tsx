'use client';

import { useState, useEffect } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameDay } from 'date-fns';
import { overviewApi, MonthlyOverview } from '@/lib/api/overview';
import { BranchQuotaStatus } from '@/lib/api/quota';

interface BranchOverviewCalendarProps {
  branchId: string;
  year?: number;
  month?: number;
}

export default function BranchOverviewCalendar({ branchId, year, month }: BranchOverviewCalendarProps) {
  const [overview, setOverview] = useState<MonthlyOverview | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedDate, setSelectedDate] = useState<Date | null>(null);

  useEffect(() => {
    loadOverview();
  }, [branchId, year, month]);

  const loadOverview = async () => {
    try {
      setLoading(true);
      const now = new Date();
      const data = await overviewApi.getMonthlyOverview({
        branch_id: branchId,
        year: year || now.getFullYear(),
        month: month || now.getMonth() + 1,
      });
      setOverview(data);
    } catch (error) {
      console.error('Failed to load monthly overview:', error);
    } finally {
      setLoading(false);
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

  const monthStart = startOfMonth(new Date(overview.year, overview.month - 1, 1));
  const monthEnd = endOfMonth(monthStart);
  const daysInMonth = eachDayOfInterval({ start: monthStart, end: monthEnd });

  const getStatusForDate = (date: Date): BranchQuotaStatus | undefined => {
    return overview.day_statuses.find(status => 
      isSameDay(new Date(status.date), date)
    );
  };

  return (
    <div className="w-full">
      <div className="mb-4">
        <h2 className="text-2xl font-bold">
          {overview.branch_name} ({overview.branch_code}) - {format(monthStart, 'MMMM yyyy')}
        </h2>
        <div className="text-sm text-gray-600 mt-1">
          Average Fulfillment: {(overview.average_fulfillment * 100).toFixed(1)}%
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full border-collapse border border-gray-300">
          <thead>
            <tr className="bg-gray-100">
              <th className="border border-gray-300 p-2 text-left sticky left-0 z-10 bg-gray-100">
                Date
              </th>
              <th className="border border-gray-300 p-2 text-center">Designated</th>
              <th className="border border-gray-300 p-2 text-center">Available</th>
              <th className="border border-gray-300 p-2 text-center">Assigned</th>
              <th className="border border-gray-300 p-2 text-center">Required</th>
              <th className="border border-gray-300 p-2 text-center">Status</th>
            </tr>
          </thead>
          <tbody>
            {daysInMonth.map((day) => {
              const status = getStatusForDate(day);
              if (!status) {
                return (
                  <tr key={day.toISOString()}>
                    <td className="border border-gray-300 p-2 sticky left-0 z-10 bg-white">
                      {format(day, 'MMM d, EEE')}
                    </td>
                    <td colSpan={5} className="border border-gray-300 p-2 text-center text-gray-400">
                      No data
                    </td>
                  </tr>
                );
              }

              const hasShortage = status.total_required > 0;
              const fulfillmentRate = status.total_designated > 0 
                ? (status.total_assigned / status.total_designated) * 100 
                : 0;

              return (
                <tr
                  key={day.toISOString()}
                  className={`hover:bg-gray-50 cursor-pointer ${
                    selectedDate && isSameDay(selectedDate, day) ? 'bg-blue-50' : ''
                  }`}
                  onClick={() => setSelectedDate(day)}
                >
                  <td className="border border-gray-300 p-2 sticky left-0 z-10 bg-white font-medium">
                    {format(day, 'MMM d, EEE')}
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

      {selectedDate && (
        <div className="mt-6 p-4 bg-gray-50 rounded-lg">
          <h3 className="text-lg font-semibold mb-2">
            Position Details - {format(selectedDate, 'MMMM d, yyyy')}
          </h3>
          {getStatusForDate(selectedDate)?.position_statuses.map((posStatus) => (
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
