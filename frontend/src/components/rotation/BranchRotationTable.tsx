'use client';

import { useState, useEffect, useMemo } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, addMonths, subMonths, isSameDay } from 'date-fns';
import { rotationApi, RotationAssignment, EligibleStaff } from '@/lib/api/rotation';
import { staffApi, Staff } from '@/lib/api/staff';
import { branchApi, Branch } from '@/lib/api/branch';
import { scheduleApi, StaffSchedule, ScheduleStatus } from '@/lib/api/schedule';
import { positionApi, Position } from '@/lib/api/position';

interface BranchRotationTableProps {
  branchId?: string;
  manuallyAddedStaff?: Staff[];
  onRemoveStaff?: (staffId: string) => void;
}

export default function BranchRotationTable({ branchId, manuallyAddedStaff = [], onRemoveStaff }: BranchRotationTableProps) {
  const [selectedBranchId, setSelectedBranchId] = useState<string>(branchId || '');
  const [currentDate, setCurrentDate] = useState(new Date());
  const [assignments, setAssignments] = useState<RotationAssignment[]>([]);
  const [allAssignments, setAllAssignments] = useState<RotationAssignment[]>([]); // All assignments for the month (all branches)
  const [branchStaff, setBranchStaff] = useState<Staff[]>([]);
  const [eligibleRotationStaff, setEligibleRotationStaff] = useState<EligibleStaff[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [branchStaffSchedules, setBranchStaffSchedules] = useState<StaffSchedule[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Memoize month calculations to prevent unnecessary re-renders
  const monthStart = useMemo(() => startOfMonth(currentDate), [currentDate]);
  const monthEnd = useMemo(() => endOfMonth(currentDate), [currentDate]);
  const daysInMonth = useMemo(() => eachDayOfInterval({ start: monthStart, end: monthEnd }), [monthStart, monthEnd]);
  
  // Use a stable string representation for the month in dependency arrays
  const monthKey = useMemo(() => format(currentDate, 'yyyy-MM'), [currentDate]);

  // Load branches on mount
  useEffect(() => {
    loadBranches();
  }, []);

  // Load data when branch or month changes
  useEffect(() => {
    if (selectedBranchId) {
      loadData();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedBranchId, monthKey]);

  const loadBranches = async () => {
    try {
      const branchesData = await branchApi.list();
      setBranches(branchesData);
      if (branchesData.length > 0 && !selectedBranchId) {
        setSelectedBranchId(branchesData[0].id);
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load branches');
    }
  };

  const loadData = async () => {
    if (!selectedBranchId) return;

    setLoading(true);
    setError(null);
    try {
      const startDateStr = format(monthStart, 'yyyy-MM-dd');
      const endDateStr = format(monthEnd, 'yyyy-MM-dd');
      const year = currentDate.getFullYear();
      const month = currentDate.getMonth() + 1;

      // Load branch staff, eligible rotation staff, assignments (current branch), all assignments (all branches), schedules, and positions in parallel
      const [branchStaffData, eligibleStaffData, assignmentsData, allAssignmentsData, schedulesData, positionsData] = await Promise.all([
        staffApi.list({ branch_id: selectedBranchId }),
        rotationApi.getEligibleStaff(selectedBranchId),
        rotationApi.getAssignments({
          branch_id: selectedBranchId,
          start_date: startDateStr,
          end_date: endDateStr,
        }),
        rotationApi.getAssignments({
          start_date: startDateStr,
          end_date: endDateStr,
        }), // Load ALL assignments for the month to check exclusivity
        scheduleApi.getMonthlyView(selectedBranchId, year, month),
        positionApi.list(),
      ]);

      setBranchStaff(branchStaffData || []);
      setEligibleRotationStaff(eligibleStaffData || []);
      setAssignments(assignmentsData || []);
      setAllAssignments(allAssignmentsData || []); // Store all assignments
      setBranchStaffSchedules(schedulesData || []);
      setPositions(positionsData || []);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load data');
      console.error('Failed to load data:', err);
    } finally {
      setLoading(false);
    }
  };

  const selectedBranch = branches.find(b => b.id === selectedBranchId);

  // Combine eligible rotation staff with manually added staff
  const allRotationStaff: (Staff | EligibleStaff)[] = useMemo(() => {
    const eligibleIds = new Set(eligibleRotationStaff.map(s => s.id));
    // Add manually added staff that aren't already in eligible staff
    const additionalStaff = manuallyAddedStaff
      .filter(s => !eligibleIds.has(s.id))
      .map(s => ({
        ...s,
        staff_type: 'rotation' as const,
        assignment_level: 1 as 1 | 2,
      }));
    return [...eligibleRotationStaff, ...additionalStaff];
  }, [eligibleRotationStaff, manuallyAddedStaff]);

  // Combine branch staff and rotation staff - memoize to prevent unnecessary re-renders
  const allStaff: (Staff | EligibleStaff)[] = useMemo(() => [
    ...branchStaff,
    ...allRotationStaff,
  ], [branchStaff, allRotationStaff]);

  const isAssigned = (staffId: string, date: Date): RotationAssignment | undefined => {
    const dateStr = format(date, 'yyyy-MM-dd');
    return assignments.find(
      a => a.rotation_staff_id === staffId && 
           a.branch_id === selectedBranchId && 
           a.date === dateStr
    );
  };

  // Check if rotation staff is assigned to another branch on this date
  const isAssignedToOtherBranch = (staffId: string, date: Date): RotationAssignment | null => {
    const dateStr = format(date, 'yyyy-MM-dd');
    const assignment = allAssignments.find(
      a => a.rotation_staff_id === staffId && a.date === dateStr
    );
    // Return assignment if it exists and is for a different branch
    if (assignment && assignment.branch_id !== selectedBranchId) {
      return assignment;
    }
    return null;
  };

  // Get next schedule status in cycle
  const getNextScheduleStatus = (current: ScheduleStatus): ScheduleStatus => {
    const cycle: ScheduleStatus[] = ['working', 'leave', 'sick_leave', 'off'];
    const currentIndex = cycle.indexOf(current);
    return cycle[(currentIndex + 1) % cycle.length];
  };

  const getScheduleForCell = (staffId: string, date: Date): StaffSchedule | undefined => {
    return branchStaffSchedules.find(
      (s) => s.staff_id === staffId && isSameDay(new Date(s.date), date)
    );
  };

  const handleDateClick = async (staffId: string, date: Date) => {
    const staff = allStaff.find(s => s.id === staffId);
    // Only allow assignment for rotation staff
    if (staff?.staff_type !== 'rotation') return;

    const dateStr = format(date, 'yyyy-MM-dd');
    
    // Check if assigned to another branch - don't allow changes
    const otherBranchAssignment = isAssignedToOtherBranch(staffId, date);
    if (otherBranchAssignment) {
      const otherBranch = branches.find(b => b.id === otherBranchAssignment.branch_id);
      setError(`This rotation staff is already assigned to ${otherBranch?.name || 'another branch'} on this date.`);
      return;
    }

    const existingAssignment = isAssigned(staffId, date);
    const isPast = date < new Date(new Date().setHours(0, 0, 0, 0));
    if (isPast) return;

    try {
      if (existingAssignment) {
        // Cycle through schedule statuses
        const currentStatus: ScheduleStatus = existingAssignment.schedule_status || 'working';
        const nextStatus = getNextScheduleStatus(currentStatus);
        
        // If cycling to 'off', remove assignment instead
        if (nextStatus === 'off') {
          await rotationApi.removeAssignment(existingAssignment.id);
          setAssignments(assignments.filter(a => a.id !== existingAssignment.id));
          setAllAssignments(allAssignments.filter(a => a.id !== existingAssignment.id));
        } else {
          // Update assignment status
          const updatedAssignment = await rotationApi.updateAssignmentStatus(existingAssignment.id, nextStatus);
          setAssignments(assignments.map(a => a.id === existingAssignment.id ? updatedAssignment : a));
          setAllAssignments(allAssignments.map(a => a.id === existingAssignment.id ? updatedAssignment : a));
        }
      } else {
        // Add assignment with default 'working' status
        const eligibleStaff = eligibleRotationStaff.find(s => s.id === staffId);
        const assignmentLevel = eligibleStaff?.assignment_level || 1;

        const newAssignment = await rotationApi.assign({
          rotation_staff_id: staffId,
          branch_id: selectedBranchId,
          date: dateStr,
          assignment_level: assignmentLevel as 1 | 2,
          schedule_status: 'working',
        });
        setAssignments([...assignments, newAssignment]);
        setAllAssignments([...allAssignments, newAssignment]);
      }
      setError(null);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || 'Failed to update assignment';
      setError(errorMessage);
      console.error('Failed to update assignment:', err);
      // Auto-clear error after 5 seconds
      setTimeout(() => setError(null), 5000);
    }
  };


  const goToPreviousMonth = () => setCurrentDate(subMonths(currentDate, 1));
  const goToNextMonth = () => setCurrentDate(addMonths(currentDate, 1));
  const goToToday = () => setCurrentDate(new Date());

  if (loading && !selectedBranchId) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  return (
    <div className="w-full p-4">
      {/* Header */}
      <div className="mb-3 flex items-center justify-between flex-wrap gap-3">
        <div className="flex items-center gap-3">
          <select
            value={selectedBranchId}
            onChange={(e) => setSelectedBranchId(e.target.value)}
            className="px-3 py-1.5 text-sm border border-neutral-border rounded-md bg-white text-neutral-text-primary"
          >
            <option value="">Select Branch</option>
            {branches.map(branch => (
              <option key={branch.id} value={branch.id}>
                {branch.name} ({branch.code})
              </option>
            ))}
          </select>
          {selectedBranch && (
            <span className="text-sm font-medium text-neutral-text-primary">
              {format(currentDate, 'MMMM yyyy')}
            </span>
          )}
        </div>
        <div className="flex gap-1.5">
          <button
            onClick={goToPreviousMonth}
            className="btn-secondary text-sm px-2 py-1"
          >
            ◀ Prev
          </button>
          <button
            onClick={goToToday}
            className="btn-secondary text-sm px-2 py-1"
          >
            Today
          </button>
          <button
            onClick={goToNextMonth}
            className="btn-secondary text-sm px-2 py-1"
          >
            Next ▶
          </button>
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <div className="mb-3 p-2 bg-red-50 border border-red-200 rounded-md">
          <p className="text-xs text-red-800">{error}</p>
        </div>
      )}

      {!selectedBranchId ? (
        <div className="card p-8 text-center">
          <p className="text-neutral-text-secondary">Please select a branch to view rotation staff assignments</p>
        </div>
      ) : loading ? (
        <div className="card p-8 text-center">
          <p className="text-neutral-text-secondary">Loading...</p>
        </div>
      ) : (
        <>
          {/* Branch Staff Table */}
          {branchStaff.length > 0 && (
            <div className="card overflow-x-auto mb-4">
              <h3 className="text-base font-semibold text-neutral-text-primary mb-2 px-3 pt-3">
                Branch Staff Schedule (Read-only)
              </h3>
              <table className="w-full border-collapse border border-neutral-border">
                <thead>
                  <tr>
                    <th className="border border-neutral-border p-1.5 bg-neutral-hover font-semibold sticky left-0 z-10 text-neutral-text-primary min-w-[120px] text-sm">
                      Staff Name
                    </th>
                    <th className="border border-neutral-border p-1.5 bg-neutral-hover font-semibold sticky left-[120px] z-10 text-neutral-text-primary min-w-[80px] text-sm">
                      Nickname
                    </th>
                    <th className="border border-neutral-border p-1.5 bg-neutral-hover font-semibold sticky left-[200px] z-10 text-neutral-text-primary min-w-[100px] text-sm">
                      Position
                    </th>
                    {daysInMonth.map((day) => (
                      <th
                        key={day.toISOString()}
                        className="border border-neutral-border p-1.5 bg-neutral-hover font-semibold min-w-[80px] text-neutral-text-primary"
                      >
                        <div className="text-xs text-neutral-text-secondary">{format(day, 'EEE')}</div>
                        <div className="text-sm">{format(day, 'd')}</div>
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {branchStaff.map((staffMember) => {
                    const position = positions.find((p) => p.id === staffMember.position_id);
                    return (
                      <tr key={staffMember.id}>
                        <td className="border border-neutral-border p-1.5 bg-neutral-bg-secondary sticky left-0 z-10 font-medium text-neutral-text-primary min-w-[120px]">
                          <span className="text-sm truncate block" title={staffMember.name}>{staffMember.name}</span>
                        </td>
                        <td className="border border-neutral-border p-1.5 bg-neutral-bg-secondary sticky left-[120px] z-10 text-neutral-text-primary min-w-[80px]">
                          <span className="text-sm font-semibold truncate block" title={staffMember.nickname || ''}>{staffMember.nickname || '-'}</span>
                        </td>
                        <td className="border border-neutral-border p-1.5 bg-neutral-bg-secondary sticky left-[200px] z-10 text-neutral-text-primary min-w-[100px]">
                          <span className="text-sm text-neutral-text-secondary italic truncate block" title={position?.name || ''}>{position?.name || '-'}</span>
                        </td>
                        {daysInMonth.map((day) => {
                          const schedule = getScheduleForCell(staffMember.id, day);
                          const scheduleStatus: ScheduleStatus = schedule?.schedule_status || 
                            (schedule?.is_working_day ? 'working' : 'off');
                          const isToday = isSameDay(day, new Date());
                          
                          // Determine cell styling based on schedule status
                          let bgColor = 'bg-neutral-bg-secondary';
                          let indicator = '';
                          
                          switch (scheduleStatus) {
                            case 'working':
                              bgColor = 'bg-green-50';
                              indicator = '✓';
                              break;
                            case 'leave':
                              bgColor = 'bg-yellow-50';
                              indicator = 'L';
                              break;
                            case 'sick_leave':
                              bgColor = 'bg-red-50';
                              indicator = 'S';
                              break;
                            case 'off':
                            default:
                              bgColor = 'bg-neutral-bg-secondary';
                              indicator = 'X';
                              break;
                          }
                          
                          const statusLabel = scheduleStatus === 'working' ? 'Working' 
                            : scheduleStatus === 'leave' ? 'Leave' 
                            : scheduleStatus === 'sick_leave' ? 'Sick Leave'
                            : 'Off';
                          
                          return (
                            <td
                              key={day.toISOString()}
                              className={`
                                border border-neutral-border p-1
                                ${bgColor}
                                ${isToday ? 'ring-2 ring-blue-500' : ''}
                              `}
                              title={`${format(day, 'MMM d, yyyy')} - ${statusLabel} (Read-only)`}
                            >
                              <div className="w-6 h-6 flex items-center justify-center text-xs font-semibold">
                                {indicator}
                              </div>
                            </td>
                          );
                        })}
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}

          {/* Rotation Staff Assignment Table */}
          <div className="card overflow-x-auto">
            <table className="w-full border-collapse border border-neutral-border">
              <thead>
                <tr>
                  <th className="border border-neutral-border p-1.5 bg-neutral-hover font-semibold sticky left-0 z-10 text-neutral-text-primary min-w-[200px] text-sm">
                    Staff Name / Position
                  </th>
                  {daysInMonth.map((day) => (
                    <th
                      key={day.toISOString()}
                      className="border border-neutral-border p-1.5 bg-neutral-hover font-semibold min-w-[80px] text-neutral-text-primary"
                    >
                      <div className="text-xs text-neutral-text-secondary">{format(day, 'EEE')}</div>
                      <div className="text-sm">{format(day, 'd')}</div>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {allRotationStaff.map((staff) => {
                  const position = positions.find((p) => p.id === staff.position_id);
                  const isManuallyAdded = manuallyAddedStaff.some(s => s.id === staff.id);
                  return (
                    <tr
                      key={staff.id}
                      className="bg-blue-50 hover:bg-blue-100"
                    >
                      <td className="border border-neutral-border p-1.5 bg-blue-50 sticky left-0 z-10 min-w-[200px]">
                        <div className="flex items-center gap-1.5 flex-wrap">
                          <span className="font-medium text-sm text-neutral-text-primary">{staff.nickname || staff.name}</span>
                          {staff.coverage_area && (
                            <span className="text-xs text-neutral-text-secondary">
                              ({staff.coverage_area})
                            </span>
                          )}
                          {isManuallyAdded && onRemoveStaff && (
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                onRemoveStaff(staff.id);
                              }}
                              className="text-xs text-red-600 hover:text-red-800 px-0.5"
                              title="Remove from assignment table"
                            >
                              ✕
                            </button>
                          )}
                        </div>
                        <div className="text-xs text-neutral-text-secondary italic">
                          {position?.name || 'Unknown Position'}
                        </div>
                      </td>
                      {daysInMonth.map((day) => {
                        const assignment = isAssigned(staff.id, day);
                        const otherBranchAssignment = isAssignedToOtherBranch(staff.id, day);
                        const isPast = day < new Date(new Date().setHours(0, 0, 0, 0));
                        const isToday = isSameDay(day, new Date());

                        // If assigned to another branch, show "off-[Branch Code]" - read-only
                        if (otherBranchAssignment) {
                          const otherBranch = branches.find(b => b.id === otherBranchAssignment.branch_id);
                          return (
                            <td
                              key={day.toISOString()}
                              className={`border border-neutral-border p-1 text-center bg-gray-100 cursor-not-allowed opacity-60 ${
                                isToday ? 'ring-2 ring-blue-500' : ''
                              }`}
                              title={`Assigned to ${otherBranch?.name || 'another branch'} (${otherBranch?.code || '?'}) on this date. Cannot assign to multiple branches.`}
                            >
                              <div className="w-6 h-6 flex items-center justify-center text-xs font-semibold">
                                <span className="text-gray-600 text-[10px]">off-{otherBranch?.code || '?'}</span>
                              </div>
                            </td>
                          );
                        }

                        // If assigned to current branch, show schedule status with cycling
                        if (assignment) {
                          const scheduleStatus: ScheduleStatus = assignment.schedule_status || 'working';
                          let bgColor = 'bg-neutral-bg-secondary';
                          let indicator = '';
                          
                          switch (scheduleStatus) {
                            case 'working':
                              bgColor = 'bg-green-50';
                              indicator = '✓';
                              break;
                            case 'leave':
                              bgColor = 'bg-yellow-50';
                              indicator = 'L';
                              break;
                            case 'sick_leave':
                              bgColor = 'bg-red-50';
                              indicator = 'S';
                              break;
                            case 'off':
                            default:
                              bgColor = 'bg-neutral-bg-secondary';
                              indicator = 'X';
                              break;
                          }
                          
                          const statusLabel = scheduleStatus === 'working' ? 'Working' 
                            : scheduleStatus === 'leave' ? 'Leave' 
                            : scheduleStatus === 'sick_leave' ? 'Sick Leave'
                            : 'Off';
                          
                          return (
                            <td
                              key={day.toISOString()}
                              className={`
                                border border-neutral-border p-1 text-center cursor-pointer
                                ${bgColor}
                                hover:opacity-80
                                ${isPast ? 'opacity-60 cursor-not-allowed' : ''}
                                ${isToday ? 'ring-2 ring-blue-500' : ''}
                              `}
                              onClick={() => !isPast && handleDateClick(staff.id, day)}
                              title={`${format(day, 'MMM d, yyyy')} - ${statusLabel} (Level ${assignment.assignment_level}). Click to cycle status.`}
                            >
                              <div className="w-6 h-6 flex items-center justify-center text-xs font-semibold">
                                {indicator}
                              </div>
                            </td>
                          );
                        }

                        // Not assigned - show empty cell for assignment
                        return (
                          <td
                            key={day.toISOString()}
                            className={`border border-neutral-border p-1 text-center ${
                              isPast
                                ? 'bg-gray-100 cursor-not-allowed'
                                : 'bg-white hover:bg-blue-50 cursor-pointer'
                            } ${isToday ? 'ring-2 ring-blue-500' : ''}`}
                            onClick={() => !isPast && handleDateClick(staff.id, day)}
                            title={`Click to assign for ${format(day, 'MMM d')}`}
                          >
                            <div className="w-6 h-6 flex items-center justify-center text-xs font-semibold">
                              <span className="text-neutral-text-secondary text-xs">○</span>
                            </div>
                          </td>
                        );
                      })}
                    </tr>
                  );
                })}
                {allRotationStaff.length === 0 && (
                  <tr>
                    <td colSpan={daysInMonth.length + 1} className="p-8 text-center text-neutral-text-secondary">
                      No rotation staff added. Use the table above to add rotation staff to the assignment table.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {/* Compact Legend */}
          <div className="mt-2 flex items-center gap-3 text-xs text-neutral-text-secondary flex-wrap">
            <div className="flex items-center gap-1">
              <div className="w-3 h-3 bg-green-50 border border-neutral-border flex items-center justify-center text-[10px] font-semibold">✓</div>
              <span>Working</span>
            </div>
            <div className="flex items-center gap-1">
              <div className="w-3 h-3 bg-yellow-50 border border-neutral-border flex items-center justify-center text-[10px] font-semibold">L</div>
              <span>Leave</span>
            </div>
            <div className="flex items-center gap-1">
              <div className="w-3 h-3 bg-red-50 border border-neutral-border flex items-center justify-center text-[10px] font-semibold">S</div>
              <span>Sick</span>
            </div>
            <div className="flex items-center gap-1">
              <div className="w-3 h-3 bg-neutral-bg-secondary border border-neutral-border flex items-center justify-center text-[10px] font-semibold">X</div>
              <span>Off</span>
            </div>
            <div className="flex items-center gap-1">
              <div className="w-3 h-3 bg-green-200 border border-neutral-border"></div>
              <span>Assigned</span>
            </div>
            <div className="flex items-center gap-1">
              <div className="w-3 h-3 bg-white border border-neutral-border"></div>
              <span>Available</span>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
