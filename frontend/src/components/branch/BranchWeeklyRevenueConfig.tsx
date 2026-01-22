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
        revenueMap.set(day, existing || {
          day_of_week: day,
          skin_revenue: 0,
          ls_hm_revenue: 0,
          vitamin_cases: 0,
          slim_pen_cases: 0,
        });
      }

      setWeeklyRevenue(revenueMap);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load weekly revenue');
    } finally {
      setLoading(false);
    }
  };

  const handleRevenueChange = (
    dayOfWeek: number,
    field: 'skin_revenue' | 'ls_hm_revenue' | 'vitamin_cases' | 'slim_pen_cases',
    value: number
  ) => {
    const revenue = weeklyRevenue.get(dayOfWeek);
    if (!revenue) return;

    const updatedRevenue = { ...revenue, [field]: value >= 0 ? value : 0 };
    setWeeklyRevenue(new Map(weeklyRevenue.set(dayOfWeek, updatedRevenue)));
    setError(null);
  };

  const handleFormattedInputChange = (
    dayOfWeek: number,
    field: 'skin_revenue' | 'ls_hm_revenue' | 'vitamin_cases' | 'slim_pen_cases',
    inputValue: string,
    isInteger: boolean = false
  ) => {
    // Remove all non-digit characters except decimal point (if not integer)
    const cleaned = isInteger
      ? inputValue.replace(/[^\d]/g, '')
      : inputValue.replace(/[^\d.]/g, '');
    
    // Handle empty input
    if (cleaned === '' || cleaned === '.') {
      handleRevenueChange(dayOfWeek, field, 0);
      return;
    }

    // Parse the numeric value
    const numericValue = isInteger ? parseInt(cleaned, 10) : parseFloat(cleaned);
    if (!isNaN(numericValue)) {
      handleRevenueChange(dayOfWeek, field, numericValue);
    }
  };

  const formatNumber = (value: number, isInteger: boolean = false): string => {
    if (value === 0) return '0';
    if (isInteger) {
      return value.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
    }
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
        skin_revenue: revenue.skin_revenue || 0,
        ls_hm_revenue: revenue.ls_hm_revenue || 0,
        vitamin_cases: revenue.vitamin_cases || 0,
        slim_pen_cases: revenue.slim_pen_cases || 0,
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
        <h3 className="text-lg font-semibold">Expected Revenue & Cases per Day of Week</h3>
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
          <h4 className="text-sm font-medium text-gray-900">Daily Expected Revenue & Cases</h4>
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Day
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Skin Revenue (THB)
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  LS HM Revenue (THB)
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Vitamin Cases
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Slim Pen Cases
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {Array.from({ length: 7 }, (_, i) => i).map((dayOfWeek) => {
                const revenue = weeklyRevenue.get(dayOfWeek);
                if (!revenue) return null;

                return (
                  <tr key={dayOfWeek}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {DAY_NAMES[dayOfWeek]}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <input
                          type="text"
                          value={formatNumber(revenue.skin_revenue || 0)}
                          onChange={(e) => handleFormattedInputChange(dayOfWeek, 'skin_revenue', e.target.value)}
                          placeholder="0"
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                        <span className="text-sm text-gray-500">THB</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <input
                          type="text"
                          value={formatNumber(revenue.ls_hm_revenue || 0)}
                          onChange={(e) => handleFormattedInputChange(dayOfWeek, 'ls_hm_revenue', e.target.value)}
                          placeholder="0"
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                        <span className="text-sm text-gray-500">THB</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <input
                          type="text"
                          value={formatNumber(revenue.vitamin_cases || 0, true)}
                          onChange={(e) => handleFormattedInputChange(dayOfWeek, 'vitamin_cases', e.target.value, true)}
                          placeholder="0"
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                        <span className="text-sm text-gray-500">cases</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <input
                          type="text"
                          value={formatNumber(revenue.slim_pen_cases || 0, true)}
                          onChange={(e) => handleFormattedInputChange(dayOfWeek, 'slim_pen_cases', e.target.value, true)}
                          placeholder="0"
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                        <span className="text-sm text-gray-500">cases</span>
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>

      <div className="text-sm text-gray-500">
        <p>Configure expected revenue and cases for each day of the week. These values will be used in automatic staff allocation calculations.</p>
      </div>
    </div>
  );
}
