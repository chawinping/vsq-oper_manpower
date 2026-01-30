'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/contexts/UserContext';
import { testDataApi, GenerateScheduleRequest, GenerateScheduleResponse } from '@/lib/api/test-data';
import { branchApi, Branch } from '@/lib/api/branch';

export default function TestDataPage() {
  const router = useRouter();
  const { user, loading: userLoading } = useUser();
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<GenerateScheduleResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Form state
  const [startDate, setStartDate] = useState(() => {
    const today = new Date();
    return today.toISOString().split('T')[0];
  });
  const [endDate, setEndDate] = useState(() => {
    const today = new Date();
    const nextMonth = new Date(today);
    nextMonth.setMonth(nextMonth.getMonth() + 1);
    return nextMonth.toISOString().split('T')[0];
  });
  const [minWorkingDays, setMinWorkingDays] = useState(4);
  const [maxWorkingDays, setMaxWorkingDays] = useState(6);
  const [leaveProbability, setLeaveProbability] = useState(0.15);
  const [consecutiveLeaveMax, setConsecutiveLeaveMax] = useState(3);
  const [weekendWorkingRatio, setWeekendWorkingRatio] = useState(0.3);
  const [excludeHolidays, setExcludeHolidays] = useState(true);
  const [minOffDaysPerMonth, setMinOffDaysPerMonth] = useState(4);
  const [maxOffDaysPerMonth, setMaxOffDaysPerMonth] = useState(8);
  const [enforceMinStaffPerGroup, setEnforceMinStaffPerGroup] = useState(true);
  const [selectedMonth, setSelectedMonth] = useState(() => {
    const today = new Date();
    return `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}`;
  });
  const [overwriteExisting, setOverwriteExisting] = useState(false);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [selectedBranchIds, setSelectedBranchIds] = useState<string[]>([]);
  const [branchesLoading, setBranchesLoading] = useState(true);

  useEffect(() => {
    // Check if user has permission
    if (!userLoading && user && user.role !== 'admin') {
      router.push('/dashboard');
      return;
    }
  }, [user, userLoading, router]);

  // Load branches on mount
  useEffect(() => {
    const loadBranches = async () => {
      try {
        setBranchesLoading(true);
        const branchesList = await branchApi.list();
        setBranches(branchesList);
      } catch (err) {
        console.error('Error loading branches:', err);
        setError('Failed to load branches');
      } finally {
        setBranchesLoading(false);
      }
    };

    if (user?.role === 'admin') {
      loadBranches();
    }
  }, [user]);

  // Update dates when month changes
  useEffect(() => {
    if (selectedMonth) {
      const [year, month] = selectedMonth.split('-').map(Number);
      const firstDay = new Date(year, month - 1, 1);
      const lastDay = new Date(year, month, 0);
      setStartDate(firstDay.toISOString().split('T')[0]);
      setEndDate(lastDay.toISOString().split('T')[0]);
    }
  }, [selectedMonth]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const request: GenerateScheduleRequest = {
        start_date: startDate,
        end_date: endDate,
        rules: {
          min_working_days_per_week: minWorkingDays,
          max_working_days_per_week: maxWorkingDays,
          leave_probability: leaveProbability,
          consecutive_leave_max: consecutiveLeaveMax,
          weekend_working_ratio: weekendWorkingRatio,
          exclude_holidays: excludeHolidays,
          min_off_days_per_month: minOffDaysPerMonth,
          max_off_days_per_month: maxOffDaysPerMonth,
          enforce_min_staff_per_group: enforceMinStaffPerGroup,
        },
        overwrite_existing: overwriteExisting,
        branch_ids: selectedBranchIds.length > 0 ? selectedBranchIds : undefined,
      };

      const response = await testDataApi.generateSchedules(request);
      setResult(response);
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Failed to generate schedules');
      console.error('Error generating schedules:', err);
    } finally {
      setLoading(false);
    }
  };

  if (userLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  if (user?.role !== 'admin') {
    return null;
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-4xl">
      <h1 className="text-3xl font-bold mb-6">Test Data Generation</h1>
      <p className="text-gray-600 mb-8">
        Generate staff working days and leave days for all branches with configurable rules.
        This tool is for development and testing purposes only.
      </p>

      <div className="bg-white rounded-lg shadow-md p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4">Schedule Generation Rules</h2>
        
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Branch Selection */}
          <div>
            <label htmlFor="branchSelection" className="block text-sm font-medium text-gray-700 mb-1">
              Select Branches (Optional)
            </label>
            <p className="text-xs text-gray-500 mb-2">
              Leave empty to generate for all branches, or select specific branches
            </p>
            {branchesLoading ? (
              <div className="text-sm text-gray-500">Loading branches...</div>
            ) : (
              <div className="border border-gray-300 rounded-md p-2 max-h-48 overflow-y-auto">
                {branches.length === 0 ? (
                  <div className="text-sm text-gray-500">No branches available</div>
                ) : (
                  <>
                    <div className="mb-2">
                      <button
                        type="button"
                        onClick={() => setSelectedBranchIds([])}
                        className="text-xs text-blue-600 hover:text-blue-800"
                      >
                        Clear Selection
                      </button>
                      <button
                        type="button"
                        onClick={() => setSelectedBranchIds(branches.map(b => b.id))}
                        className="text-xs text-blue-600 hover:text-blue-800 ml-4"
                      >
                        Select All
                      </button>
                    </div>
                    <div className="space-y-2">
                      {branches.map((branch) => (
                        <label key={branch.id} className="flex items-center cursor-pointer">
                          <input
                            type="checkbox"
                            checked={selectedBranchIds.includes(branch.id)}
                            onChange={(e) => {
                              if (e.target.checked) {
                                setSelectedBranchIds([...selectedBranchIds, branch.id]);
                              } else {
                                setSelectedBranchIds(selectedBranchIds.filter(id => id !== branch.id));
                              }
                            }}
                            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                          />
                          <span className="ml-2 text-sm text-gray-700">
                            {branch.name} ({branch.code})
                          </span>
                        </label>
                      ))}
                    </div>
                  </>
                )}
              </div>
            )}
            {selectedBranchIds.length > 0 && (
              <p className="text-xs text-gray-500 mt-1">
                {selectedBranchIds.length} branch{selectedBranchIds.length !== 1 ? 'es' : ''} selected
              </p>
            )}
          </div>

          {/* Month Selection */}
          <div>
            <label htmlFor="selectedMonth" className="block text-sm font-medium text-gray-700 mb-1">
              Select Month
            </label>
            <input
              type="month"
              id="selectedMonth"
              value={selectedMonth}
              onChange={(e) => setSelectedMonth(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
            />
            <p className="text-xs text-gray-500 mt-1">
              Schedules will be generated for the entire selected month
            </p>
          </div>

          {/* Date Range (read-only, auto-filled from month) */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label htmlFor="startDate" className="block text-sm font-medium text-gray-700 mb-1">
                Start Date (auto-filled)
              </label>
              <input
                type="date"
                id="startDate"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-gray-50"
                required
              />
            </div>
            <div>
              <label htmlFor="endDate" className="block text-sm font-medium text-gray-700 mb-1">
                End Date (auto-filled)
              </label>
              <input
                type="date"
                id="endDate"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-gray-50"
                required
              />
            </div>
          </div>

          {/* Working Days */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label htmlFor="minWorkingDays" className="block text-sm font-medium text-gray-700 mb-1">
                Min Working Days Per Week
              </label>
              <input
                type="number"
                id="minWorkingDays"
                min="0"
                max="7"
                value={minWorkingDays}
                onChange={(e) => setMinWorkingDays(parseInt(e.target.value) || 0)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
            </div>
            <div>
              <label htmlFor="maxWorkingDays" className="block text-sm font-medium text-gray-700 mb-1">
                Max Working Days Per Week
              </label>
              <input
                type="number"
                id="maxWorkingDays"
                min="0"
                max="7"
                value={maxWorkingDays}
                onChange={(e) => setMaxWorkingDays(parseInt(e.target.value) || 0)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
            </div>
          </div>

          {/* Leave Settings */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label htmlFor="leaveProbability" className="block text-sm font-medium text-gray-700 mb-1">
                Leave Probability (0.0 - 1.0)
              </label>
              <input
                type="number"
                id="leaveProbability"
                min="0"
                max="1"
                step="0.01"
                value={leaveProbability}
                onChange={(e) => setLeaveProbability(parseFloat(e.target.value) || 0)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
              <p className="text-xs text-gray-500 mt-1">
                {Math.round(leaveProbability * 100)}% chance of leave on any given day
              </p>
            </div>
            <div>
              <label htmlFor="consecutiveLeaveMax" className="block text-sm font-medium text-gray-700 mb-1">
                Max Consecutive Leave Days
              </label>
              <input
                type="number"
                id="consecutiveLeaveMax"
                min="0"
                value={consecutiveLeaveMax}
                onChange={(e) => setConsecutiveLeaveMax(parseInt(e.target.value) || 0)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
            </div>
          </div>

          {/* Weekend Settings */}
          <div>
            <label htmlFor="weekendWorkingRatio" className="block text-sm font-medium text-gray-700 mb-1">
              Weekend Working Ratio (0.0 - 1.0)
            </label>
            <input
              type="number"
              id="weekendWorkingRatio"
              min="0"
              max="1"
              step="0.01"
              value={weekendWorkingRatio}
              onChange={(e) => setWeekendWorkingRatio(parseFloat(e.target.value) || 0)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
            />
            <p className="text-xs text-gray-500 mt-1">
              {Math.round(weekendWorkingRatio * 100)}% of weekends will be working days
            </p>
          </div>

          {/* Off Days Per Month */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label htmlFor="minOffDaysPerMonth" className="block text-sm font-medium text-gray-700 mb-1">
                Min Off Days Per Month
              </label>
              <input
                type="number"
                id="minOffDaysPerMonth"
                min="0"
                value={minOffDaysPerMonth}
                onChange={(e) => setMinOffDaysPerMonth(parseInt(e.target.value) || 0)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
            </div>
            <div>
              <label htmlFor="maxOffDaysPerMonth" className="block text-sm font-medium text-gray-700 mb-1">
                Max Off Days Per Month
              </label>
              <input
                type="number"
                id="maxOffDaysPerMonth"
                min="0"
                value={maxOffDaysPerMonth}
                onChange={(e) => setMaxOffDaysPerMonth(parseInt(e.target.value) || 0)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
            </div>
          </div>

          {/* Options */}
          <div className="space-y-3">
            <div className="flex items-center">
              <input
                type="checkbox"
                id="excludeHolidays"
                checked={excludeHolidays}
                onChange={(e) => setExcludeHolidays(e.target.checked)}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label htmlFor="excludeHolidays" className="ml-2 block text-sm text-gray-700">
                Exclude Public Holidays (set as off days)
              </label>
            </div>
            <div className="flex items-center">
              <input
                type="checkbox"
                id="enforceMinStaffPerGroup"
                checked={enforceMinStaffPerGroup}
                onChange={(e) => setEnforceMinStaffPerGroup(e.target.checked)}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label htmlFor="enforceMinStaffPerGroup" className="ml-2 block text-sm text-gray-700">
                Enforce Minimum Staff Per Group (ensures branch constraints are met)
              </label>
            </div>
            <div className="flex items-center">
              <input
                type="checkbox"
                id="overwriteExisting"
                checked={overwriteExisting}
                onChange={(e) => setOverwriteExisting(e.target.checked)}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label htmlFor="overwriteExisting" className="ml-2 block text-sm text-gray-700">
                Overwrite Existing Schedules
              </label>
            </div>
          </div>

          {/* Submit Button */}
          <div className="pt-4">
            <button
              type="submit"
              disabled={loading}
              className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Generating Schedules...' : 'Generate Schedules'}
            </button>
          </div>
        </form>
      </div>

      {/* Error Display */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <h3 className="text-red-800 font-semibold mb-2">Error</h3>
          <p className="text-red-700">{error}</p>
        </div>
      )}

      {/* Result Display */}
      {result && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-6">
          <h3 className="text-green-800 font-semibold mb-4 text-lg">Generation Complete!</h3>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            <div>
              <p className="text-sm text-gray-600">Total Staff</p>
              <p className="text-2xl font-bold text-gray-900">{result.result.total_staff}</p>
            </div>
            <div>
              <p className="text-sm text-gray-600">Total Schedules</p>
              <p className="text-2xl font-bold text-gray-900">{result.result.total_schedules}</p>
            </div>
            <div>
              <p className="text-sm text-gray-600">Working Days</p>
              <p className="text-2xl font-bold text-green-600">{result.result.working_days}</p>
            </div>
            <div>
              <p className="text-sm text-gray-600">Leave Days</p>
              <p className="text-2xl font-bold text-yellow-600">{result.result.leave_days}</p>
            </div>
            <div>
              <p className="text-sm text-gray-600">Off Days</p>
              <p className="text-2xl font-bold text-gray-600">{result.result.off_days}</p>
            </div>
          </div>
          {result.result.errors && result.result.errors.length > 0 && (
            <div className="mt-4">
              <p className="text-sm font-semibold text-red-700 mb-2">Errors ({result.result.errors.length}):</p>
              <ul className="list-disc list-inside text-sm text-red-600 space-y-1">
                {result.result.errors.slice(0, 10).map((err, idx) => (
                  <li key={idx}>{err}</li>
                ))}
                {result.result.errors.length > 10 && (
                  <li className="text-gray-500">... and {result.result.errors.length - 10} more errors</li>
                )}
              </ul>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
