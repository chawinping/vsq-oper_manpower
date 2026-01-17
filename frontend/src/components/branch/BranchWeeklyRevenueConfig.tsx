'use client';

import { useState, useEffect } from 'react';
import { branchConfigApi, WeeklyRevenue, WeeklyRevenueUpdate } from '@/lib/api/branch-config';

interface BranchWeeklyRevenueConfigProps {
  branchId: string;
  onSave?: () => void;
}

const DAY_NAMES = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

export default function BranchWeeklyRevenueConfig({ branchId, onSave }: BranchWeeklyRevenueConfigProps) {
  const [weeklyRevenue, setWeeklyRevenue] = useState<Map<number, WeeklyRevenue>>(new Map());
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, [branchId]);

  const loadData = async () => {
    setLoading(true);
    setError(null);
    try {
      const revenueData = await branchConfigApi.getWeeklyRevenue(branchId);
      const revenueMap = new Map<number, WeeklyRevenue>();

      // Initialize all days (0-6) with default values
      for (let day = 0; day <= 6; day++) {
        const existing = revenueData.find((r) => r.day_of_week === day);
        revenueMap.set(day, existing || { day_of_week: day, expected_revenue: 0 });
      }

      setWeeklyRevenue(revenueMap);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load weekly revenue');
    } finally {
      setLoading(false);
    }
  };

  const handleRevenueChange = (dayOfWeek: number, value: number) => {
    const revenue = weeklyRevenue.get(dayOfWeek);
    if (!revenue) return;

    const updatedRevenue = { ...revenue, expected_revenue: value >= 0 ? value : 0 };
    setWeeklyRevenue(new Map(weeklyRevenue.set(dayOfWeek, updatedRevenue)));
    setError(null);
  };

  const handleFormattedInputChange = (dayOfWeek: number, inputValue: string) => {
    // Remove all non-digit characters except decimal point
    const cleaned = inputValue.replace(/[^\d.]/g, '');
    
    // Handle empty input
    if (cleaned === '' || cleaned === '.') {
      handleRevenueChange(dayOfWeek, 0);
      return;
    }

    // Parse the numeric value
    const numericValue = parseFloat(cleaned);
    if (!isNaN(numericValue)) {
      handleRevenueChange(dayOfWeek, numericValue);
    }
  };

  const formatNumber = (value: number): string => {
    if (value === 0) return '0';
    // Format with commas, handling decimals
    const parts = value.toString().split('.');
    parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ',');
    return parts.join('.');
  };

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    setSuccess(null);

    try {
      const revenueToUpdate: WeeklyRevenueUpdate[] = Array.from(weeklyRevenue.values()).map((revenue) => ({
        day_of_week: revenue.day_of_week,
        expected_revenue: revenue.expected_revenue,
      }));

      await branchConfigApi.updateWeeklyRevenue(branchId, revenueToUpdate);
      setSuccess('Weekly revenue updated successfully');
      if (onSave) {
        onSave();
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update weekly revenue');
    } finally {
      setSaving(false);
    }
  };

  const handleReset = () => {
    loadData();
    setError(null);
    setSuccess(null);
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-semibold">Expected Revenue per Day of Week</h3>
        <div className="flex gap-2">
          <button
            onClick={handleReset}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
          >
            Reset
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {saving ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </div>

      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-md">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {success && (
        <div className="p-4 bg-green-50 border border-green-200 rounded-md">
          <p className="text-sm text-green-800">{success}</p>
        </div>
      )}

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h4 className="text-sm font-medium text-gray-900">Daily Expected Revenue (THB)</h4>
        </div>
        <div className="px-6 py-4 space-y-3">
          {Array.from({ length: 7 }, (_, i) => i).map((dayOfWeek) => {
            const revenue = weeklyRevenue.get(dayOfWeek);
            if (!revenue) return null;

            return (
              <div key={dayOfWeek} className="flex items-center justify-between">
                <label className="text-sm font-medium text-gray-700 w-32">
                  {DAY_NAMES[dayOfWeek]}:
                </label>
                <div className="flex items-center gap-2 flex-1 max-w-md">
                  <input
                    type="text"
                    value={formatNumber(revenue.expected_revenue)}
                    onChange={(e) => handleFormattedInputChange(dayOfWeek, e.target.value)}
                    placeholder="0"
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <span className="text-sm text-gray-500">THB</span>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      <div className="text-sm text-gray-500">
        <p>Configure expected revenue for each day of the week. This will be used in automatic allocation calculations.</p>
      </div>
    </div>
  );
}
