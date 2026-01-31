'use client';

import { useState, useEffect, useMemo } from 'react';
import { branchApi, Branch } from '@/lib/api/branch';
import { rotationApi, AssignRotationRequest, AllocationSuggestion, RotationStaffSchedule, RotationAssignment } from '@/lib/api/rotation';
import { staffApi, Staff } from '@/lib/api/staff';
import { positionApi, Position } from '@/lib/api/position';
import { overviewApi, BranchQuotaStatus } from '@/lib/api/overview';
import { branchTypeApi, BranchType, BranchTypeRequirement, BranchTypeConstraints } from '@/lib/api/branch-type';
import { staffGroupApi, StaffGroup } from '@/lib/api/staff-group';
import { effectiveBranchApi, EffectiveBranch } from '@/lib/api/effectiveBranch';

interface BranchDetailDrawerProps {
  isOpen: boolean;
  branchId: string;
  date: string;
  onClose: () => void;
  onSuccess: () => void;
}

export default function BranchDetailDrawer({
  isOpen,
  branchId,
  date,
  onClose,
  onSuccess,
}: BranchDetailDrawerProps) {
  const [branch, setBranch] = useState<Branch | null>(null);
  const [availableStaff, setAvailableStaff] = useState<Staff[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [branchQuotaStatus, setBranchQuotaStatus] = useState<BranchQuotaStatus | null>(null);
  const [suggestions, setSuggestions] = useState<AllocationSuggestion[]>([]);
  const [branchType, setBranchType] = useState<BranchType | null>(null);
  const [branchTypeRequirements, setBranchTypeRequirements] = useState<BranchTypeRequirement[]>([]);
  const [branchTypeConstraints, setBranchTypeConstraints] = useState<BranchTypeConstraints[]>([]);
  const [staffGroups, setStaffGroups] = useState<StaffGroup[]>([]);
  const [schedules, setSchedules] = useState<RotationStaffSchedule[]>([]);
  const [assignments, setAssignments] = useState<RotationAssignment[]>([]);
  const [effectiveBranchesMap, setEffectiveBranchesMap] = useState<Map<string, EffectiveBranch[]>>(new Map());
  const [loading, setLoading] = useState(false);
  const [assigning, setAssigning] = useState(false);
  const [selectedStaffId, setSelectedStaffId] = useState('');
  const [selectedPositionId, setSelectedPositionId] = useState('');
  const [assignmentLevel, setAssignmentLevel] = useState<1 | 2>(1);

  useEffect(() => {
    if (isOpen) {
      loadData();
    }
  }, [isOpen, branchId, date]);

  const loadData = async () => {
    try {
      setLoading(true);
      const [branchesData, staffData, positionsData, quotaStatus, suggestionsData, staffGroupsData, schedulesData, assignmentsData] = await Promise.all([
        branchApi.list(),
        staffApi.list({ staff_type: 'rotation' }),
        positionApi.list(),
        overviewApi.getBranchQuotaStatus(branchId, date).catch(() => null),
        rotationApi.getSuggestions({ branch_id: branchId, start_date: date, end_date: date }).catch(() => ({ suggestions: [] })),
        staffGroupApi.list().catch(() => []),
        rotationApi.getSchedules({ date }).catch(() => []),
        rotationApi.getAssignments({ branch_id: branchId, date }).catch(() => []),
      ]);

      // Load positions for each staff group
      const staffGroupsWithPositions = await Promise.all(
        (staffGroupsData || []).map(async (sg) => {
          try {
            const fullGroup = await staffGroupApi.getById(sg.id);
            return fullGroup;
          } catch {
            return sg;
          }
        })
      );

      // Load effective branches for all rotation staff
      const effectiveBranchesMapData = new Map<string, EffectiveBranch[]>();
      await Promise.all(
        (staffData || []).map(async (staff) => {
          try {
            const effectiveBranches = await effectiveBranchApi.getByRotationStaffID(staff.id);
            effectiveBranchesMapData.set(staff.id, effectiveBranches || []);
          } catch {
            effectiveBranchesMapData.set(staff.id, []);
          }
        })
      );

      const branchData = branchesData.find(b => b.id === branchId);
      setBranch(branchData || null);
      setAvailableStaff(staffData || []);
      setPositions(positionsData || []);
      setBranchQuotaStatus(quotaStatus);
      setSuggestions((suggestionsData.suggestions || []) as AllocationSuggestion[]);
      setStaffGroups(staffGroupsWithPositions || []);
      setSchedules(schedulesData || []);
      setAssignments(assignmentsData || []);
      setEffectiveBranchesMap(effectiveBranchesMapData);

      // Load branch type information if branch has a branch_type_id
      if (branchData?.branch_type_id) {
        try {
          const [branchTypeData, requirementsData, constraintsData] = await Promise.all([
            branchTypeApi.getById(branchData.branch_type_id).catch(() => null),
            branchTypeApi.getRequirements(branchData.branch_type_id).catch(() => []),
            branchTypeApi.getConstraints(branchData.branch_type_id).catch(() => []),
          ]);
          
          setBranchType(branchTypeData);
          setBranchTypeRequirements(requirementsData || []);
          setBranchTypeConstraints(constraintsData || []);
        } catch (error) {
          console.error('Failed to load branch type data:', error);
        }
      }
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  // Helper function to get staff group name by ID
  const getStaffGroupName = (staffGroupId: string): string => {
    const staffGroup = staffGroups.find(sg => sg.id === staffGroupId);
    return staffGroup?.name || 'Unknown';
  };

  // Helper function to calculate actual staff count for a staff group
  const getActualStaffCountForGroup = (staffGroupId: string): number => {
    if (!branchQuotaStatus || !staffGroups.length || !positions.length) {
      return 0;
    }

    // Find the staff group
    const staffGroup = staffGroups.find(sg => sg.id === staffGroupId);
    if (!staffGroup || !staffGroup.positions || staffGroup.positions.length === 0) {
      return 0;
    }

    // Get all position IDs for this staff group
    const positionIds = staffGroup.positions.map(p => p.position_id);

    // Sum up total_assigned for all positions in this staff group
    const actualCount = branchQuotaStatus.position_statuses
      .filter(pos => positionIds.includes(pos.position_id))
      .reduce((sum, pos) => sum + pos.total_assigned, 0);

    return actualCount;
  };

  // Get day of week (0 = Sunday, 1 = Monday, ..., 6 = Saturday)
  const getDayOfWeek = (dateString: string): number => {
    const date = new Date(dateString);
    return date.getDay();
  };

  const getDayName = (dayOfWeek: number): string => {
    const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
    return days[dayOfWeek] || '';
  };

  const handleAssign = async () => {
    if (!selectedStaffId || !selectedPositionId) {
      alert('Please select staff and position');
      return;
    }

    try {
      setAssigning(true);
      const request: AssignRotationRequest = {
        rotation_staff_id: selectedStaffId,
        branch_id: branchId,
        date,
        assignment_level: assignmentLevel,
      };
      await rotationApi.assign(request);
      onSuccess();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to assign staff');
    } finally {
      setAssigning(false);
    }
  };

  if (!isOpen) return null;

  // Filter rotation staff: must have branch in effective branches, be working on selected date, and match position if selected
  const filteredStaff = useMemo(() => {
    return availableStaff.filter(staff => {
      // Filter by position if selected
      if (selectedPositionId && staff.position_id !== selectedPositionId) {
        return false;
      }

      // Check if staff has this branch in their effective branches
      const effectiveBranches = effectiveBranchesMap.get(staff.id) || [];
      const hasBranchAccess = effectiveBranches.some(eb => eb.branch_id === branchId);
      if (!hasBranchAccess) {
        return false;
      }

      // Check if staff is working on the selected date
      // Only include staff who have an explicit 'working' schedule entry
      const schedule = schedules.find(s => {
        // Handle potential date format differences (compare normalized dates)
        const scheduleDate = s.date.split('T')[0]; // Remove time if present
        const targetDate = date.split('T')[0];
        return s.rotation_staff_id === staff.id && scheduleDate === targetDate;
      });
      
      // Only include if schedule exists and status is 'working'
      if (!schedule || schedule.schedule_status !== 'working') {
        return false;
      }

      return true;
    });
  }, [availableStaff, selectedPositionId, branchId, date, schedules, effectiveBranchesMap]);

  // Check if staff is already assigned to this branch on this date
  const isStaffAssigned = (staffId: string): boolean => {
    return assignments.some(a => a.rotation_staff_id === staffId && a.branch_id === branchId && a.date === date);
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-end z-50">
      <div className="bg-white h-full w-full max-w-2xl overflow-y-auto">
        <div className="sticky top-0 bg-white border-b border-gray-200 p-6 flex items-center justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-3 mb-1">
              <h2 className="text-2xl font-bold">
                {branch?.code} - {branch?.name}
              </h2>
              {branchQuotaStatus && (() => {
                const totalAssigned = branchQuotaStatus.position_statuses.reduce((sum, pos) => sum + pos.total_assigned, 0);
                const totalPreferred = branchQuotaStatus.position_statuses.reduce((sum, pos) => sum + pos.designated_quota, 0);
                const totalMinimum = branchQuotaStatus.position_statuses.reduce((sum, pos) => sum + pos.minimum_required, 0);
                
                let priority: 'high' | 'medium' | 'low' = 'low';
                if (totalAssigned < totalMinimum) {
                  priority = 'high';
                } else if (totalAssigned < totalPreferred) {
                  priority = 'medium';
                }
                
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
              })()}
            </div>
            <p className="text-sm text-gray-600 mt-1">Date: {date}</p>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 text-2xl font-bold ml-4"
          >
            ×
          </button>
        </div>

        <div className="p-6">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-lg text-gray-600">Loading...</div>
            </div>
          ) : (
            <>
              {/* Branch Type Information */}
              {branchType && (
                <div className="mb-6 pb-6 border-b border-gray-200">
                  <h3 className="text-lg font-semibold mb-3">Branch Type Information</h3>
                  <div className="space-y-3">
                    <div>
                      <span className="text-sm font-medium text-gray-700">Type:</span>
                      <span className="ml-2 text-sm text-gray-900">{branchType.name}</span>
                    </div>
                    {branchType.description && (
                      <div>
                        <span className="text-sm font-medium text-gray-700">Description:</span>
                        <p className="ml-2 text-sm text-gray-600 mt-1">{branchType.description}</p>
                      </div>
                    )}
                    
                    {/* Branch Type Requirements for Selected Date */}
                    {branchTypeRequirements.length > 0 && (() => {
                      const dayOfWeek = getDayOfWeek(date);
                      const dayRequirements = branchTypeRequirements.filter(req => req.day_of_week === dayOfWeek && req.is_active);
                      
                      if (dayRequirements.length > 0) {
                        return (
                          <div className="mt-4">
                            <h4 className="text-sm font-semibold text-gray-700 mb-2">
                              Requirements for {getDayName(dayOfWeek)}:
                            </h4>
                            <div className="bg-gray-50 rounded-lg p-3">
                              <table className="w-full text-sm">
                                <thead>
                                  <tr className="border-b border-gray-200">
                                    <th className="text-left py-2 px-2 font-medium text-gray-700">Staff Group</th>
                                    <th className="text-right py-2 px-2 font-medium text-gray-700">Actual/Minimum</th>
                                  </tr>
                                </thead>
                                <tbody>
                                  {dayRequirements.map((req) => {
                                    const actualCount = getActualStaffCountForGroup(req.staff_group_id);
                                    return (
                                      <tr key={req.id} className="border-b border-gray-100">
                                        <td className="py-2 px-2 text-gray-900">
                                          {getStaffGroupName(req.staff_group_id)}
                                        </td>
                                        <td className="py-2 px-2 text-right text-gray-700 font-medium">
                                          {actualCount}/{req.minimum_staff_count}
                                        </td>
                                      </tr>
                                    );
                                  })}
                                </tbody>
                              </table>
                            </div>
                          </div>
                        );
                      }
                      return null;
                    })()}
                    
                    {/* Branch Type Constraints for Selected Date */}
                    {branchTypeConstraints.length > 0 && (() => {
                      const dayOfWeek = getDayOfWeek(date);
                      const dayConstraints = branchTypeConstraints.filter(con => con.day_of_week === dayOfWeek);
                      
                      if (dayConstraints.length > 0) {
                        const constraint = dayConstraints[0];
                        if (constraint.staff_group_requirements && constraint.staff_group_requirements.length > 0) {
                          return (
                            <div className="mt-4">
                              <h4 className="text-sm font-semibold text-gray-700 mb-2">
                                Constraints for {getDayName(dayOfWeek)}:
                              </h4>
                              <div className="bg-blue-50 rounded-lg p-3">
                                <table className="w-full text-sm">
                                  <thead>
                                    <tr className="border-b border-blue-200">
                                      <th className="text-left py-2 px-2 font-medium text-gray-700">Staff Group</th>
                                      <th className="text-right py-2 px-2 font-medium text-gray-700">Actual/Minimum</th>
                                    </tr>
                                  </thead>
                                  <tbody>
                                    {constraint.staff_group_requirements.map((sgReq) => {
                                      const actualCount = getActualStaffCountForGroup(sgReq.staff_group_id);
                                      return (
                                        <tr key={sgReq.id} className="border-b border-blue-100">
                                          <td className="py-2 px-2 text-gray-900">
                                            {getStaffGroupName(sgReq.staff_group_id)}
                                          </td>
                                          <td className="py-2 px-2 text-right text-gray-700 font-medium">
                                            {actualCount}/{sgReq.minimum_count}
                                          </td>
                                        </tr>
                                      );
                                    })}
                                  </tbody>
                                </table>
                              </div>
                            </div>
                          );
                        }
                      }
                      return null;
                    })()}
                  </div>
                </div>
              )}

              {/* Current Staff Summary */}
              {branchQuotaStatus && branchQuotaStatus.position_statuses.length > 0 && (
                <div className="mb-6">
                  <h3 className="text-lg font-semibold mb-4">Current Staff Summary</h3>
                  <div className="overflow-x-auto">
                    <table className="w-full border-collapse border border-gray-300">
                      <thead>
                        <tr className="bg-gray-50">
                          <th className="border border-gray-300 px-4 py-2 text-left text-sm font-medium">Position</th>
                          <th className="border border-gray-300 px-4 py-2 text-center text-sm font-medium">Current</th>
                          <th className="border border-gray-300 px-4 py-2 text-center text-sm font-medium">Preferred</th>
                          <th className="border border-gray-300 px-4 py-2 text-center text-sm font-medium">Minimum</th>
                          <th className="border border-gray-300 px-4 py-2 text-center text-sm font-medium">Status</th>
                        </tr>
                      </thead>
                      <tbody>
                        {branchQuotaStatus.position_statuses.map((status) => {
                          const meetsMinimum = status.total_assigned >= status.minimum_required;
                          const meetsPreferred = status.total_assigned >= status.designated_quota;
                          const isBelowMinimum = status.total_assigned < status.minimum_required;
                          
                          return (
                            <tr key={status.position_id} className="hover:bg-gray-50">
                              <td className="border border-gray-300 px-4 py-2 text-sm">{status.position_name}</td>
                              <td className="border border-gray-300 px-4 py-2 text-center text-sm">{status.total_assigned}</td>
                              <td className="border border-gray-300 px-4 py-2 text-center text-sm">{status.designated_quota}</td>
                              <td className="border border-gray-300 px-4 py-2 text-center text-sm">{status.minimum_required}</td>
                              <td className="border border-gray-300 px-4 py-2 text-center text-sm">
                                {meetsPreferred ? (
                                  <span className="text-green-600 font-medium">✓</span>
                                ) : isBelowMinimum ? (
                                  <span className="text-red-600 font-medium">⚠️</span>
                                ) : (
                                  <span className="text-yellow-600 font-medium">⚠️</span>
                                )}
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              {/* Allocation Suggestions */}
              {suggestions.length > 0 && (
                <div className="mb-6">
                  <h3 className="text-lg font-semibold mb-4">Allocation Suggestions</h3>
                  <div className="space-y-3">
                    {suggestions
                      .sort((a, b) => {
                        // Lexicographic ordering: Group 1 → Group 2 → Group 3 → Branch Code
                        // More negative = higher priority
                        
                        // Primary: Group 1 Score (ascending - more negative = higher priority)
                        if (a.group1_score !== b.group1_score) {
                          return a.group1_score - b.group1_score;
                        }
                        
                        // Secondary: Group 2 Score (ascending - more negative = higher priority)
                        if (a.group2_score !== b.group2_score) {
                          return a.group2_score - b.group2_score;
                        }
                        
                        // Tertiary: Group 3 Score (descending - more positive = lower priority)
                        if (a.group3_score !== b.group3_score) {
                          return b.group3_score - a.group3_score;
                        }
                        
                        // Deterministic tie-breaker: Branch Code (alphabetical)
                        return (a.branch_code || '').localeCompare(b.branch_code || '');
                      })
                      .slice(0, 5) // Show top 5 suggestions
                      .map((suggestion, index) => {
                        const group1Score = suggestion.group1_score ?? 0;
                        const group2Score = suggestion.group2_score ?? 0;
                        const group3Score = suggestion.group3_score ?? 0;
                        
                        // Determine overall priority based on Group 1 (highest priority)
                        const getPriorityBadge = () => {
                          if (group1Score < -2) {
                            return <span className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium bg-red-100 text-red-800 border border-red-300">🔴 Critical</span>;
                          } else if (group1Score < 0) {
                            return <span className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium bg-orange-100 text-orange-800 border border-orange-300">🟠 High</span>;
                          } else if (group2Score < 0) {
                            return <span className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium bg-yellow-100 text-yellow-800 border border-yellow-300">🟡 Medium</span>;
                          } else {
                            return <span className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium bg-green-100 text-green-800 border border-green-300">🟢 Low</span>;
                          }
                        };

                        return (
                          <div key={index} className="border border-gray-200 rounded-lg p-4 bg-gray-50">
                            <div className="flex items-start justify-between mb-3">
                              <div className="flex-1">
                                <div className="flex items-center gap-2 mb-2">
                                  <span className="font-medium text-gray-900">
                                    {index + 1}. Assign {suggestion.position_name || 'Staff'}
                                  </span>
                                  {getPriorityBadge()}
                                </div>
                                
                                {/* Scoring Groups Display */}
                                <div className="space-y-2 mb-3">
                                  {/* Group 1: Position Quota - Minimum */}
                                  <div className="flex items-center gap-2 text-sm">
                                    <span className="font-medium text-gray-700 min-w-[200px]">
                                      Group 1 (Position Quota - Min):
                                    </span>
                                    <span className={`font-semibold ${group1Score < 0 ? 'text-red-600' : 'text-gray-600'}`}>
                                      {group1Score} points
                                    </span>
                                    {group1Score < 0 && (
                                      <span className="text-xs text-red-600">
                                        ({Math.abs(group1Score)} staff below minimum)
                                      </span>
                                    )}
                                  </div>
                                  
                                  {/* Group 2: Daily Staff Constraints - Minimum */}
                                  <div className="flex items-center gap-2 text-sm">
                                    <span className="font-medium text-gray-700 min-w-[200px]">
                                      Group 2 (Daily Constraints - Min):
                                    </span>
                                    <span className={`font-semibold ${group2Score < 0 ? 'text-orange-600' : 'text-gray-600'}`}>
                                      {group2Score} points
                                    </span>
                                    {group2Score < 0 && (
                                      <span className="text-xs text-orange-600">
                                        ({Math.abs(group2Score)} groups below minimum)
                                      </span>
                                    )}
                                  </div>
                                  
                                  {/* Group 3: Position Quota - Preferred */}
                                  <div className="flex items-center gap-2 text-sm">
                                    <span className="font-medium text-gray-700 min-w-[200px]">
                                      Group 3 (Position Quota - Preferred):
                                    </span>
                                    <span className={`font-semibold ${group3Score > 0 ? 'text-blue-600' : 'text-gray-600'}`}>
                                      {group3Score > 0 ? '+' : ''}{group3Score} points
                                    </span>
                                    {group3Score > 0 && (
                                      <span className="text-xs text-blue-600">
                                        ({group3Score} staff above preferred - informational only)
                                      </span>
                                    )}
                                  </div>
                                </div>
                                
                                {/* Score Breakdown Details (if available) */}
                                {suggestion.score_breakdown && (
                                  <details className="mb-2 text-sm">
                                    <summary className="cursor-pointer text-gray-600 hover:text-gray-800 font-medium">
                                      View Score Breakdown
                                    </summary>
                                    <div className="mt-2 ml-4 space-y-2 text-xs">
                                      {suggestion.score_breakdown.position_quota_minimum.length > 0 && (
                                        <div>
                                          <div className="font-medium text-red-700 mb-1">Position Quota - Minimum:</div>
                                          {suggestion.score_breakdown.position_quota_minimum.map((item, idx) => (
                                            <div key={idx} className="ml-2 text-gray-600">
                                              {item.position_name}: Needs {item.minimum_required}, Has {item.current_count} → {item.points} points
                                            </div>
                                          ))}
                                        </div>
                                      )}
                                      {suggestion.score_breakdown.daily_constraints_minimum.length > 0 && (
                                        <div>
                                          <div className="font-medium text-orange-700 mb-1">Daily Constraints - Minimum:</div>
                                          {suggestion.score_breakdown.daily_constraints_minimum.map((item, idx) => (
                                            <div key={idx} className="ml-2 text-gray-600">
                                              {item.staff_group_name}: Needs {item.minimum_count}, Has {item.actual_count} → {item.points} points
                                            </div>
                                          ))}
                                        </div>
                                      )}
                                      {suggestion.score_breakdown.position_quota_preferred.length > 0 && (
                                        <div>
                                          <div className="font-medium text-blue-700 mb-1">Position Quota - Preferred:</div>
                                          {suggestion.score_breakdown.position_quota_preferred.map((item, idx) => (
                                            <div key={idx} className="ml-2 text-gray-600">
                                              {item.position_name}: Preferred {item.preferred_quota}, Has {item.current_count} → +{item.points} points ({item.shortage} above preferred)
                                            </div>
                                          ))}
                                        </div>
                                      )}
                                    </div>
                                  </details>
                                )}
                                
                                <div className="text-sm text-gray-700 mt-2">
                                  <strong>Reason:</strong> {suggestion.reason}
                                </div>
                                {suggestion.suggested_staff_name && (
                                  <div className="text-sm text-gray-600 mt-2">
                                    <strong>Suggested:</strong> {suggestion.suggested_staff_name}
                                  </div>
                                )}
                                
                                {/* Legacy priority score (if available, for backward compatibility) */}
                                {suggestion.priority_score !== undefined && (
                                  <div className="text-xs text-gray-500 mt-1">
                                    Legacy Priority Score: {(suggestion.priority_score * 100).toFixed(0)}%
                                  </div>
                                )}
                              </div>
                            </div>
                            {suggestion.suggested_staff_id && (
                              <button
                                onClick={() => {
                                  setSelectedPositionId(suggestion.position_id);
                                  setSelectedStaffId(suggestion.suggested_staff_id || '');
                                }}
                                className="mt-2 px-3 py-1 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                              >
                                Quick Assign →
                              </button>
                            )}
                          </div>
                        );
                      })}
                  </div>
                </div>
              )}

              {/* Add Rotation Staff Form */}
              <div className="border-t border-gray-200 pt-6">
                <h3 className="text-lg font-semibold mb-4">Add Rotation Staff</h3>
                <div className="space-y-4">
                  {/* Position Selection (Optional Filter) */}
                  <div>
                    <label className="block text-sm font-medium mb-1">Filter by Position (Optional)</label>
                    <select
                      value={selectedPositionId}
                      onChange={(e) => {
                        setSelectedPositionId(e.target.value);
                        setSelectedStaffId(''); // Reset staff selection
                      }}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    >
                      <option value="">All Positions</option>
                      {positions.map(position => (
                        <option key={position.id} value={position.id}>
                          {position.name}
                        </option>
                      ))}
                    </select>
                  </div>

                  {/* Staff Selection - Cards */}
                  <div>
                    <label className="block text-sm font-medium mb-2">
                      Available Rotation Staff *
                      {!loading && (
                        <span className="ml-2 text-xs font-normal text-gray-500">
                          ({filteredStaff.length} available)
                        </span>
                      )}
                    </label>
                    {loading ? (
                      <div className="text-sm text-gray-500 py-4 text-center">
                        Loading available rotation staff...
                      </div>
                    ) : filteredStaff.length === 0 ? (
                      <div className="text-sm text-gray-500 py-4 text-center border border-gray-200 rounded-lg p-4 bg-gray-50">
                        <p className="font-medium mb-1">No rotation staff available</p>
                        <p className="text-xs">
                          {selectedPositionId 
                            ? `No rotation staff found for this position on ${date} who have this branch in their effective branches and are available to work.`
                            : `No rotation staff found for this branch on ${date} who are available to work.`}
                        </p>
                        <p className="text-xs mt-2 text-gray-400">
                          Make sure rotation staff have this branch in their effective branches and have an explicit 'working' schedule entry for this date.
                        </p>
                        {/* Debug info - remove in production */}
                        {process.env.NODE_ENV === 'development' && (
                          <details className="mt-3 text-left">
                            <summary className="text-xs cursor-pointer text-gray-600">Debug Info</summary>
                            <div className="mt-2 text-xs space-y-1">
                              <p>Total rotation staff: {availableStaff.length}</p>
                              <p>Staff with effective branches loaded: {effectiveBranchesMap.size}</p>
                              <p>Schedules loaded for {date}: {schedules.length}</p>
                              <p>Branch ID: {branchId}</p>
                            </div>
                          </details>
                        )}
                      </div>
                    ) : (
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 max-h-96 overflow-y-auto">
                        {filteredStaff.map(staff => {
                          const isAssigned = isStaffAssigned(staff.id);
                          const effectiveBranches = effectiveBranchesMap.get(staff.id) || [];
                          const schedule = schedules.find(s => s.rotation_staff_id === staff.id && s.date === date);
                          const assignment = assignments.find(a => a.rotation_staff_id === staff.id && a.branch_id === branchId && a.date === date);
                          
                          return (
                            <div
                              key={staff.id}
                              onClick={() => {
                                setSelectedStaffId(staff.id);
                                setSelectedPositionId(staff.position_id);
                              }}
                              className={`p-4 border-2 rounded-lg cursor-pointer transition-all ${
                                selectedStaffId === staff.id
                                  ? 'border-blue-500 bg-blue-50'
                                  : isAssigned
                                  ? 'border-yellow-400 bg-yellow-50'
                                  : 'border-gray-200 bg-white hover:border-gray-300 hover:shadow-md'
                              }`}
                            >
                              <div className="flex items-start justify-between mb-2">
                                <div className="flex-1">
                                  <div className="font-semibold text-gray-900">
                                    {staff.name}
                                    {staff.nickname && (
                                      <span className="text-gray-600 font-normal ml-1">({staff.nickname})</span>
                                    )}
                                  </div>
                                  {staff.position && (
                                    <div className="text-sm text-gray-600 mt-1">
                                      {staff.position.name}
                                    </div>
                                  )}
                                </div>
                                {isAssigned && (
                                  <span className="px-2 py-1 text-xs font-medium bg-yellow-200 text-yellow-800 rounded">
                                    Assigned
                                  </span>
                                )}
                                {selectedStaffId === staff.id && (
                                  <span className="px-2 py-1 text-xs font-medium bg-blue-200 text-blue-800 rounded">
                                    Selected
                                  </span>
                                )}
                              </div>
                              
                              <div className="mt-2 space-y-1 text-xs text-gray-600">
                                {schedule && (
                                  <div className="flex items-center gap-1">
                                    <span className="font-medium">Schedule:</span>
                                    <span className={`px-1.5 py-0.5 rounded ${
                                      schedule.schedule_status === 'working' 
                                        ? 'bg-green-100 text-green-700' 
                                        : schedule.schedule_status === 'off'
                                        ? 'bg-gray-100 text-gray-700'
                                        : 'bg-red-100 text-red-700'
                                    }`}>
                                      {schedule.schedule_status === 'working' ? 'Working' : 
                                       schedule.schedule_status === 'off' ? 'Off' :
                                       schedule.schedule_status === 'leave' ? 'Leave' : 'Sick Leave'}
                                    </span>
                                  </div>
                                )}
                                {assignment && (
                                  <div className="flex items-center gap-1">
                                    <span className="font-medium">Current Assignment:</span>
                                    <span className="px-1.5 py-0.5 rounded bg-yellow-200 text-yellow-800 text-xs font-medium">
                                      Level {assignment.assignment_level} {assignment.assignment_level === 1 ? '(Priority)' : '(Reserved)'}
                                    </span>
                                  </div>
                                )}
                                {effectiveBranches.length > 0 && (
                                  <div className="flex items-start gap-1">
                                    <span className="font-medium">Effective Branches:</span>
                                    <span className="flex-1">
                                      {effectiveBranches.slice(0, 3).map(eb => eb.branch?.code || eb.branch_id).join(', ')}
                                      {effectiveBranches.length > 3 && ` +${effectiveBranches.length - 3} more`}
                                    </span>
                                  </div>
                                )}
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </div>

                  {/* Assignment Level */}
                  <div>
                    <label className="block text-sm font-medium mb-1">Assignment Level *</label>
                    <select
                      value={assignmentLevel}
                      onChange={(e) => setAssignmentLevel(parseInt(e.target.value) as 1 | 2)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      required
                    >
                      <option value="1">Level 1 (Priority)</option>
                      <option value="2">Level 2 (Reserved)</option>
                    </select>
                  </div>

                  {/* Action Buttons */}
                  <div className="flex gap-3 pt-4">
                    <button
                      onClick={onClose}
                      className="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-md"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={handleAssign}
                      disabled={assigning || !selectedStaffId || !selectedPositionId}
                      className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {assigning ? 'Assigning...' : 'Assign Staff'}
                    </button>
                  </div>
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
