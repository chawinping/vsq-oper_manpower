'use client';

import { useState, useEffect } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameDay, addMonths, subMonths } from 'date-fns';
import { rotationApi, RotationAssignment, AssignmentSuggestion } from '@/lib/api/rotation';
import { staffApi, Staff } from '@/lib/api/staff';
import { branchApi, Branch } from '@/lib/api/branch';

type ViewMode = 'by-branch' | 'by-staff' | 'by-area';

interface RotationAssignmentViewProps {
  startDate?: Date;
  endDate?: Date;
}

export default function RotationAssignmentView({ startDate, endDate }: RotationAssignmentViewProps) {
  const [viewMode, setViewMode] = useState<ViewMode>('by-branch');
  const [assignments, setAssignments] = useState<RotationAssignment[]>([]);
  const [rotationStaff, setRotationStaff] = useState<Staff[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentDate, setCurrentDate] = useState(new Date());
  const [selectedBranchId, setSelectedBranchId] = useState<string>('');
  const [selectedStaffId, setSelectedStaffId] = useState<string>('');
  const [selectedArea, setSelectedArea] = useState<string>('');
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [suggestions, setSuggestions] = useState<AssignmentSuggestion[]>([]);
  const [loadingSuggestions, setLoadingSuggestions] = useState(false);

  useEffect(() => {
    loadData();
  }, [viewMode, selectedBranchId, selectedStaffId, selectedArea]);

  const loadData = async () => {
    try {
      setLoading(true);
      const filters: any = {};
      
      if (selectedBranchId) filters.branch_id = selectedBranchId;
      if (selectedStaffId) filters.rotation_staff_id = selectedStaffId;
      if (selectedArea) filters.coverage_area = selectedArea;
      
      const [assignmentsData, staffData, branchesData] = await Promise.all([
        rotationApi.getAssignments(filters),
        staffApi.list({ staff_type: 'rotation' }),
        branchApi.list(),
      ]);
      
      setAssignments(assignmentsData || []);
      setRotationStaff(staffData || []);
      setBranches(branchesData || []);
    } catch (error) {
      console.error('Failed to load data:', error);
      setAssignments([]);
      setRotationStaff([]);
      setBranches([]);
    } finally {
      setLoading(false);
    }
  };

  const monthStart = startOfMonth(currentDate);
  const monthEnd = endOfMonth(currentDate);
  const daysInMonth = eachDayOfInterval({ start: monthStart, end: monthEnd });

  const getAssignmentsForCell = (staffId: string, branchId: string, date: Date): RotationAssignment | undefined => {
    return assignments.find(
      (a) =>
        a.rotation_staff_id === staffId &&
        a.branch_id === branchId &&
        isSameDay(new Date(a.date), date)
    );
  };

  const handleAssign = async (staffId: string, branchId: string, date: Date, level: 1 | 2) => {
    try {
      await rotationApi.assign({
        rotation_staff_id: staffId,
        branch_id: branchId,
        date: format(date, 'yyyy-MM-dd'),
        assignment_level: level,
      });
      await loadData();
    } catch (error) {
      console.error('Failed to assign staff:', error);
    }
  };

  const handleRemoveAssignment = async (assignmentId: string) => {
    try {
      await rotationApi.removeAssignment(assignmentId);
      await loadData();
    } catch (error) {
      console.error('Failed to remove assignment:', error);
    }
  };

  const handleGetSuggestions = async () => {
    setLoadingSuggestions(true);
    try {
      const startDate = format(monthStart, 'yyyy-MM-dd');
      const endDate = format(monthEnd, 'yyyy-MM-dd');
      const filters: any = {
        start_date: startDate,
        end_date: endDate,
      };
      if (selectedBranchId) filters.branch_id = selectedBranchId;
      
      const response = await rotationApi.getSuggestions(filters);
      setSuggestions(response.suggestions || []);
      setShowSuggestions(true);
    } catch (error: any) {
      if (error.response?.status === 501) {
        alert('AI suggestions are not yet implemented. The MCP server needs to be configured.');
      } else {
        alert(error.response?.data?.error || 'Failed to get suggestions');
      }
    } finally {
      setLoadingSuggestions(false);
    }
  };

  const handleRegenerateSuggestions = async () => {
    setLoadingSuggestions(true);
    try {
      const startDate = format(monthStart, 'yyyy-MM-dd');
      const endDate = format(monthEnd, 'yyyy-MM-dd');
      const filters: any = {
        start_date: startDate,
        end_date: endDate,
      };
      if (selectedBranchId) filters.branch_id = selectedBranchId;
      
      const response = await rotationApi.regenerateSuggestions(filters);
      setSuggestions(response.suggestions || []);
    } catch (error: any) {
      if (error.response?.status === 501) {
        alert('AI suggestions are not yet implemented. The MCP server needs to be configured.');
      } else {
        alert(error.response?.data?.error || 'Failed to regenerate suggestions');
      }
    } finally {
      setLoadingSuggestions(false);
    }
  };

  const handleApplySuggestion = async (suggestion: AssignmentSuggestion) => {
    try {
      await rotationApi.assign({
        rotation_staff_id: suggestion.rotation_staff_id,
        branch_id: suggestion.branch_id,
        date: suggestion.date,
        assignment_level: suggestion.assignment_level,
      });
      await loadData();
      // Remove applied suggestion from list
      setSuggestions(suggestions.filter((s) => 
        !(s.rotation_staff_id === suggestion.rotation_staff_id &&
          s.branch_id === suggestion.branch_id &&
          s.date === suggestion.date)
      ));
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to apply suggestion');
    }
  };

  const goToPreviousMonth = () => {
    setCurrentDate(subMonths(currentDate, 1));
  };

  const goToNextMonth = () => {
    setCurrentDate(addMonths(currentDate, 1));
  };

  // Get unique coverage areas
  const coverageAreas = Array.from(new Set(rotationStaff.map((s) => s.coverage_area).filter(Boolean)));

  // Filter data based on view mode
  const filteredBranches = branches.filter((b) => {
    if (viewMode === 'by-area' && selectedArea) {
      // Filter branches by area manager or other criteria
      return true; // Simplified - implement actual area filtering
    }
    return true;
  });

  const filteredStaff = rotationStaff.filter((s) => {
    if (selectedArea && s.coverage_area !== selectedArea) return false;
    return true;
  });

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  return (
    <div className="w-full p-6">
      <div className="mb-6 flex items-center justify-between flex-wrap gap-4">
        <h2 className="text-2xl font-bold">Rotation Staff Assignment - {format(currentDate, 'MMMM yyyy')}</h2>
        <div className="flex gap-2">
          <button
            onClick={goToPreviousMonth}
            className="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-md transition-colors"
          >
            Previous
          </button>
          <button
            onClick={() => setCurrentDate(new Date())}
            className="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-md transition-colors"
          >
            Today
          </button>
          <button
            onClick={goToNextMonth}
            className="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-md transition-colors"
          >
            Next
          </button>
        </div>
      </div>

      {/* AI Suggestions Section */}
      <div className="mb-4 flex items-center justify-between flex-wrap gap-4">
        <div className="flex gap-2">
          <button
            onClick={handleGetSuggestions}
            disabled={loadingSuggestions}
            className="px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700 transition-colors disabled:opacity-50"
          >
            {loadingSuggestions ? 'Loading...' : 'Get AI Suggestions'}
          </button>
          {suggestions.length > 0 && (
            <>
              <button
                onClick={handleRegenerateSuggestions}
                disabled={loadingSuggestions}
                className="px-4 py-2 bg-purple-500 text-white rounded-md hover:bg-purple-600 transition-colors disabled:opacity-50"
              >
                Regenerate
              </button>
              <button
                onClick={() => setShowSuggestions(!showSuggestions)}
                className="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-md transition-colors"
              >
                {showSuggestions ? 'Hide' : 'Show'} Suggestions ({suggestions.length})
              </button>
            </>
          )}
        </div>
      </div>

      {/* Suggestions Panel */}
      {showSuggestions && suggestions.length > 0 && (
        <div className="mb-6 bg-purple-50 border border-purple-200 rounded-lg p-4">
          <h3 className="text-lg font-semibold mb-3 text-purple-900">AI Suggestions</h3>
          <div className="space-y-2 max-h-64 overflow-y-auto">
            {suggestions.map((suggestion, index) => {
              const staff = rotationStaff.find((s) => s.id === suggestion.rotation_staff_id);
              const branch = branches.find((b) => b.id === suggestion.branch_id);
              return (
                <div
                  key={index}
                  className="flex items-center justify-between p-3 bg-white rounded border border-purple-200"
                >
                  <div className="flex-1">
                    <div className="font-medium">
                      {staff?.name || 'Unknown'} → {branch?.name || 'Unknown'}
                    </div>
                    <div className="text-sm text-gray-600">
                      {format(new Date(suggestion.date), 'MMM d, yyyy')} • Level {suggestion.assignment_level} • 
                      Confidence: {(suggestion.confidence * 100).toFixed(0)}%
                    </div>
                    {suggestion.reason && (
                      <div className="text-xs text-gray-500 mt-1">{suggestion.reason}</div>
                    )}
                  </div>
                  <button
                    onClick={() => handleApplySuggestion(suggestion)}
                    className="ml-4 px-3 py-1 bg-green-600 text-white text-sm rounded hover:bg-green-700 transition-colors"
                  >
                    Apply
                  </button>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* View Mode Toggle */}
      <div className="mb-4 flex gap-2">
        <button
          onClick={() => setViewMode('by-branch')}
          className={`px-4 py-2 rounded-md transition-colors ${
            viewMode === 'by-branch'
              ? 'bg-blue-500 text-white'
              : 'bg-gray-200 hover:bg-gray-300'
          }`}
        >
          By Branch
        </button>
        <button
          onClick={() => setViewMode('by-staff')}
          className={`px-4 py-2 rounded-md transition-colors ${
            viewMode === 'by-staff'
              ? 'bg-blue-500 text-white'
              : 'bg-gray-200 hover:bg-gray-300'
          }`}
        >
          By Staff
        </button>
        <button
          onClick={() => setViewMode('by-area')}
          className={`px-4 py-2 rounded-md transition-colors ${
            viewMode === 'by-area'
              ? 'bg-blue-500 text-white'
              : 'bg-gray-200 hover:bg-gray-300'
          }`}
        >
          By Area
        </button>
      </div>

      {/* Filters */}
      <div className="mb-4 flex gap-4 flex-wrap">
        {viewMode === 'by-branch' && (
          <select
            value={selectedBranchId}
            onChange={(e) => setSelectedBranchId(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md"
          >
            <option value="">All Branches</option>
            {branches.map((b) => (
              <option key={b.id} value={b.id}>
                {b.name}
              </option>
            ))}
          </select>
        )}
        {viewMode === 'by-staff' && (
          <select
            value={selectedStaffId}
            onChange={(e) => setSelectedStaffId(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md"
          >
            <option value="">All Staff</option>
            {rotationStaff.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
        )}
        {viewMode === 'by-area' && (
          <select
            value={selectedArea}
            onChange={(e) => setSelectedArea(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md"
          >
            <option value="">All Areas</option>
            {coverageAreas.map((area) => (
              <option key={area} value={area}>
                {area}
              </option>
            ))}
          </select>
        )}
      </div>

      {/* Calendar View */}
      <div className="overflow-x-auto">
        {viewMode === 'by-branch' && (
          <table className="w-full border-collapse border border-gray-300">
            <thead>
              <tr>
                <th className="border border-gray-300 p-2 bg-gray-100 font-semibold sticky left-0 z-10">
                  Branch / Staff
                </th>
                {daysInMonth.map((day) => (
                  <th
                    key={day.toISOString()}
                    className="border border-gray-300 p-2 bg-gray-100 font-semibold min-w-[100px]"
                  >
                    <div className="text-xs">{format(day, 'EEE')}</div>
                    <div className="text-sm">{format(day, 'd')}</div>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {filteredBranches.map((branch) => (
                <tr key={branch.id}>
                  <td className="border border-gray-300 p-2 bg-gray-50 sticky left-0 z-10 font-medium">
                    {branch.name}
                  </td>
                  {daysInMonth.map((day) => {
                    const assignmentsForDay = assignments.filter(
                      (a) => a.branch_id === branch.id && isSameDay(new Date(a.date), day)
                    );
                    return (
                      <td
                        key={day.toISOString()}
                        className="border border-gray-300 p-1 align-top"
                      >
                        <div className="space-y-1">
                          {assignmentsForDay.map((assignment) => {
                            const staff = rotationStaff.find((s) => s.id === assignment.rotation_staff_id);
                            return (
                              <div
                                key={assignment.id}
                                className={`text-xs p-1 rounded ${
                                  assignment.assignment_level === 1
                                    ? 'bg-blue-100'
                                    : 'bg-purple-100'
                                }`}
                              >
                                {staff?.name} (L{assignment.assignment_level})
                                <button
                                  onClick={() => handleRemoveAssignment(assignment.id)}
                                  className="ml-1 text-red-600 hover:text-red-800"
                                >
                                  ×
                                </button>
                              </div>
                            );
                          })}
                        </div>
                      </td>
                    );
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {viewMode === 'by-staff' && (
          <table className="w-full border-collapse border border-gray-300">
            <thead>
              <tr>
                <th className="border border-gray-300 p-2 bg-gray-100 font-semibold sticky left-0 z-10">
                  Staff / Branch
                </th>
                {daysInMonth.map((day) => (
                  <th
                    key={day.toISOString()}
                    className="border border-gray-300 p-2 bg-gray-100 font-semibold min-w-[100px]"
                  >
                    <div className="text-xs">{format(day, 'EEE')}</div>
                    <div className="text-sm">{format(day, 'd')}</div>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {filteredStaff.map((staffMember) => (
                <tr key={staffMember.id}>
                  <td className="border border-gray-300 p-2 bg-gray-50 sticky left-0 z-10 font-medium">
                    {staffMember.name}
                  </td>
                  {daysInMonth.map((day) => {
                    const assignmentsForDay = assignments.filter(
                      (a) => a.rotation_staff_id === staffMember.id && isSameDay(new Date(a.date), day)
                    );
                    return (
                      <td
                        key={day.toISOString()}
                        className="border border-gray-300 p-1 align-top"
                      >
                        <div className="space-y-1">
                          {assignmentsForDay.map((assignment) => {
                            const branch = branches.find((b) => b.id === assignment.branch_id);
                            return (
                              <div
                                key={assignment.id}
                                className={`text-xs p-1 rounded ${
                                  assignment.assignment_level === 1
                                    ? 'bg-blue-100'
                                    : 'bg-purple-100'
                                }`}
                              >
                                {branch?.name} (L{assignment.assignment_level})
                                <button
                                  onClick={() => handleRemoveAssignment(assignment.id)}
                                  className="ml-1 text-red-600 hover:text-red-800"
                                >
                                  ×
                                </button>
                              </div>
                            );
                          })}
                        </div>
                      </td>
                    );
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {viewMode === 'by-area' && (
          <div className="space-y-6">
            {coverageAreas.map((area) => (
              <div key={area} className="border border-gray-300 rounded-lg p-4">
                <h3 className="text-lg font-semibold mb-4">{area}</h3>
                <div className="overflow-x-auto">
                  <table className="w-full border-collapse border border-gray-300">
                    <thead>
                      <tr>
                        <th className="border border-gray-300 p-2 bg-gray-100 font-semibold">
                          Staff / Branch
                        </th>
                        {daysInMonth.map((day) => (
                          <th
                            key={day.toISOString()}
                            className="border border-gray-300 p-2 bg-gray-100 font-semibold min-w-[100px]"
                          >
                            <div className="text-xs">{format(day, 'EEE')}</div>
                            <div className="text-sm">{format(day, 'd')}</div>
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {filteredStaff
                        .filter((s) => s.coverage_area === area)
                        .map((staffMember) => (
                          <tr key={staffMember.id}>
                            <td className="border border-gray-300 p-2 bg-gray-50 font-medium">
                              {staffMember.name}
                            </td>
                            {daysInMonth.map((day) => {
                              const assignmentsForDay = assignments.filter(
                                (a) =>
                                  a.rotation_staff_id === staffMember.id &&
                                  isSameDay(new Date(a.date), day)
                              );
                              return (
                                <td
                                  key={day.toISOString()}
                                  className="border border-gray-300 p-1 align-top"
                                >
                                  <div className="space-y-1">
                                    {assignmentsForDay.map((assignment) => {
                                      const branch = branches.find((b) => b.id === assignment.branch_id);
                                      return (
                                        <div
                                          key={assignment.id}
                                          className={`text-xs p-1 rounded ${
                                            assignment.assignment_level === 1
                                              ? 'bg-blue-100'
                                              : 'bg-purple-100'
                                          }`}
                                        >
                                          {branch?.name} (L{assignment.assignment_level})
                                          <button
                                            onClick={() => handleRemoveAssignment(assignment.id)}
                                            className="ml-1 text-red-600 hover:text-red-800"
                                          >
                                            ×
                                          </button>
                                        </div>
                                      );
                                    })}
                                  </div>
                                </td>
                              );
                            })}
                          </tr>
                        ))}
                    </tbody>
                  </table>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="mt-4 flex items-center gap-4 text-sm">
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-blue-100 border border-gray-300"></div>
          <span>Level 1 Assignment</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-purple-100 border border-gray-300"></div>
          <span>Level 2 Assignment</span>
        </div>
      </div>
    </div>
  );
}

