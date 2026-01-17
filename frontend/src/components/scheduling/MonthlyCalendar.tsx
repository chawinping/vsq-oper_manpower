'use client';

import { useState, useEffect } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameDay, addMonths, subMonths, compareAsc } from 'date-fns';
import { scheduleApi, StaffSchedule, ScheduleStatus } from '@/lib/api/schedule';
import { staffApi, Staff } from '@/lib/api/staff';
import { positionApi } from '@/lib/api/position';
import { rotationApi, RotationAssignment } from '@/lib/api/rotation';
import { Branch } from '@/lib/api/branch';
import { useUser } from '@/contexts/UserContext';

interface MonthlyCalendarProps {
  branchIds: string[];
  branches?: Branch[];
}

export default function MonthlyCalendar({ branchIds, branches = [] }: MonthlyCalendarProps) {
  const { user } = useUser();
  const [currentDate, setCurrentDate] = useState(new Date());
  const [schedules, setSchedules] = useState<StaffSchedule[]>([]);
  const [staff, setStaff] = useState<Staff[]>([]);
  const [rotationAssignments, setRotationAssignments] = useState<RotationAssignment[]>([]);
  const [rotationStaffMap, setRotationStaffMap] = useState<Record<string, Staff>>({});
  const [positions, setPositions] = useState<Record<string, { id: string; name: string }>>({});
  const [loading, setLoading] = useState(true);
  const [selectedCell, setSelectedCell] = useState<{ staffId: string; date: Date } | null>(null);
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; staffId: string; date: Date } | null>(null);
  const isBranchManager = user?.role === 'branch_manager';
  const showRotationStaff = user?.role === 'admin' || user?.role === 'area_manager';
  const canEditSchedules = user?.role === 'branch_manager' || user?.role === 'admin' || user?.role === 'area_manager';

  const year = currentDate.getFullYear();
  const month = currentDate.getMonth() + 1;

  useEffect(() => {
    if (branchIds.length > 0) {
      loadData();
    }
  }, [branchIds, year, month]);

  // Close context menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (contextMenu) {
        // Don't close if clicking on the context menu itself
        const target = e.target as HTMLElement;
        if (!target.closest('.context-menu')) {
          setContextMenu(null);
        }
      }
    };
    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  }, [contextMenu]);

  const loadData = async () => {
    try {
      setLoading(true);
      const monthStart = new Date(year, month - 1, 1);
      const monthEnd = new Date(year, month, 0);
      const startDateStr = format(monthStart, 'yyyy-MM-dd');
      const endDateStr = format(monthEnd, 'yyyy-MM-dd');
      
      // Load data for all selected branches
      const [positionsData] = await Promise.all([
        positionApi.list(),
      ]);
      
      // Create positions lookup map
      const positionsMap: Record<string, { id: string; name: string }> = {};
      (positionsData || []).forEach(pos => {
        positionsMap[pos.id] = { id: pos.id, name: pos.name };
      });
      setPositions(positionsMap);
      
      // Load schedules, staff, and rotation assignments for all branches
      const allSchedules: StaffSchedule[] = [];
      const allStaff: Staff[] = [];
      const allRotationAssignments: RotationAssignment[] = [];
      
      await Promise.all(
        branchIds.map(async (branchId) => {
          const [schedulesData, staffData, rotationAssignmentsData] = await Promise.all([
            scheduleApi.getMonthlyView(branchId, year, month),
            staffApi.list({ staff_type: 'branch', branch_id: branchId }),
            rotationApi.getAssignments({ branch_id: branchId, start_date: startDateStr, end_date: endDateStr }),
          ]);
          
          allSchedules.push(...(schedulesData || []));
          allStaff.push(...(staffData || []));
          allRotationAssignments.push(...(rotationAssignmentsData || []));
        })
      );
      
      setSchedules(allSchedules);
      setRotationAssignments(allRotationAssignments);
      
      // Enrich staff with position data and branch info
      const enrichedStaff = allStaff.map(s => {
        const branch = branches.find(b => b.id === s.branch_id);
        return {
          ...s,
          position: s.position_id ? positionsMap[s.position_id] : undefined,
          branchName: branch?.name,
          branchCode: branch?.code,
        };
      });
      setStaff(enrichedStaff);
      
      // Load rotation staff details for assigned rotation staff
      const rotationStaffIds = [...new Set(allRotationAssignments.map(a => a.rotation_staff_id))];
      if (rotationStaffIds.length > 0) {
        // Fetch all rotation staff and filter by IDs we need
        const allRotationStaff = await staffApi.list({ staff_type: 'rotation' });
        const rotationStaffMapData: Record<string, Staff> = {};
        allRotationStaff.forEach(s => {
          if (rotationStaffIds.includes(s.id)) {
            rotationStaffMapData[s.id] = {
              ...s,
              position: s.position_id ? positionsMap[s.position_id] : undefined,
            };
          }
        });
        setRotationStaffMap(rotationStaffMapData);
      }
    } catch (error) {
      console.error('Failed to load data:', error);
      setSchedules([]);
      setStaff([]);
      setRotationAssignments([]);
    } finally {
      setLoading(false);
    }
  };

  const monthStart = startOfMonth(currentDate);
  const monthEnd = endOfMonth(currentDate);
  const daysInMonth = eachDayOfInterval({ start: monthStart, end: monthEnd }).sort((a, b) => compareAsc(a, b));

  const handleCellClick = async (staffId: string, date: Date, e?: React.MouseEvent) => {
    // Close context menu if open (but continue with the click action)
    if (contextMenu) {
      setContextMenu(null);
    }
    
    const dateStr = format(date, 'yyyy-MM-dd');
    const existingSchedule = schedules.find(
      (s) => s.staff_id === staffId && isSameDay(new Date(s.date), date)
    );

    try {
      // Cycle through: off -> working -> leave -> sick_leave -> off
      let nextStatus: ScheduleStatus = 'off';
      if (existingSchedule) {
        const currentStatus = existingSchedule.schedule_status || (existingSchedule.is_working_day ? 'working' : 'off');
        switch (currentStatus) {
          case 'off':
            nextStatus = 'working';
            break;
          case 'working':
            nextStatus = 'leave';
            break;
          case 'leave':
            nextStatus = 'sick_leave';
            break;
          case 'sick_leave':
            nextStatus = 'off';
            break;
          default:
            nextStatus = 'working';
            break;
        }
      } else {
        nextStatus = 'working'; // Default to working for new schedules
      }

      // Find the branch ID for this staff member
      const staffMember = staff.find(s => s.id === staffId);
      const staffBranchId = staffMember?.branch_id || branchIds[0];
      
      await scheduleApi.create({
        staff_id: staffId,
        branch_id: staffBranchId,
        date: dateStr,
        schedule_status: nextStatus,
      });
      await loadData();
    } catch (error: any) {
      console.error('Failed to update schedule:', error);
      alert(`Failed to update schedule: ${error?.response?.data?.error || error?.message || 'Unknown error'}`);
    }
  };

  const handleCellRightClick = (e: React.MouseEvent, staffId: string, date: Date) => {
    e.preventDefault();
    setContextMenu({ x: e.clientX, y: e.clientY, staffId, date });
  };

  const handleContextMenuSelect = async (status: ScheduleStatus) => {
    if (!contextMenu) return;
    
    const dateStr = format(contextMenu.date, 'yyyy-MM-dd');
    try {
      // Find the branch ID for this staff member
      const staffMember = staff.find(s => s.id === contextMenu.staffId);
      const staffBranchId = staffMember?.branch_id || branchIds[0];
      
      await scheduleApi.create({
        staff_id: contextMenu.staffId,
        branch_id: staffBranchId,
        date: dateStr,
        schedule_status: status,
      });
      await loadData();
      setContextMenu(null);
    } catch (error) {
      console.error('Failed to update schedule:', error);
    }
  };

  const handleCloseContextMenu = () => {
    setContextMenu(null);
  };

  const getScheduleForCell = (staffId: string, date: Date): StaffSchedule | undefined => {
    return schedules.find(
      (s) => s.staff_id === staffId && isSameDay(new Date(s.date), date)
    );
  };

  const goToPreviousMonth = () => {
    setCurrentDate(subMonths(currentDate, 1));
  };

  const goToNextMonth = () => {
    setCurrentDate(addMonths(currentDate, 1));
  };

  const handleBulkTurnOffToWorking = async () => {
    // Confirmation dialog
    const confirmed = window.confirm(
      `Are you sure you want to turn all "Off" days to "Working" days for ${format(currentDate, 'MMMM yyyy')}?\n\n` +
      `This will affect all staff members in the selected branches for all days in this month that are currently set to "Off".`
    );
    
    if (!confirmed) {
      return;
    }

    try {
      setLoading(true);
      const monthStart = startOfMonth(currentDate);
      const monthEnd = endOfMonth(currentDate);
      const daysInMonth = eachDayOfInterval({ start: monthStart, end: monthEnd });
      
      // Get all staff members for selected branches
      const branchStaff = staff.filter(s => branchIds.includes(s.branch_id));
      
      // Count how many updates we'll make
      let updateCount = 0;
      const updatePromises: Promise<any>[] = [];
      
      // For each day in the month
      for (const day of daysInMonth) {
        const dateStr = format(day, 'yyyy-MM-dd');
        
        // For each staff member
        for (const staffMember of branchStaff) {
          // Check if schedule exists and is "off"
          const existingSchedule = schedules.find(
            (s) => s.staff_id === staffMember.id && isSameDay(new Date(s.date), day)
          );
          
          const currentStatus = existingSchedule?.schedule_status || 
            (existingSchedule?.is_working_day ? 'working' : 'off');
          
          // Only update if status is "off" or doesn't exist (defaults to off)
          if (!existingSchedule || currentStatus === 'off') {
            updatePromises.push(
              scheduleApi.create({
                staff_id: staffMember.id,
                branch_id: staffMember.branch_id,
                date: dateStr,
                schedule_status: 'working',
              })
            );
            updateCount++;
          }
        }
      }
      
      // Execute all updates in parallel (with some batching to avoid overwhelming the server)
      const batchSize = 50;
      for (let i = 0; i < updatePromises.length; i += batchSize) {
        const batch = updatePromises.slice(i, i + batchSize);
        await Promise.all(batch);
      }
      
      // Reload data to reflect changes
      await loadData();
      
      alert(`Successfully updated ${updateCount} schedule(s) to "Working" for ${format(currentDate, 'MMMM yyyy')}.`);
    } catch (error: any) {
      console.error('Failed to bulk update schedules:', error);
      alert(`Failed to update schedules: ${error?.response?.data?.error || error?.message || 'Unknown error'}`);
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

  // Get unique rotation staff for this month (for display)
  // Show rotation staff that have assignments in the current month
  const rotationStaffForMonth = Object.values(rotationStaffMap).filter(rs => {
    const staffAssignments = rotationAssignments.filter(ra => ra.rotation_staff_id === rs.id);
    return staffAssignments.some(ra => {
      const assignmentDate = new Date(ra.date);
      return assignmentDate.getMonth() === currentDate.getMonth() &&
             assignmentDate.getFullYear() === currentDate.getFullYear();
    });
  });

  return (
    <div className="w-full p-6">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-neutral-text-primary">
            Staff Scheduling - {format(currentDate, 'MMMM yyyy')}
          </h2>
          {branches.length > 0 && (
            <p className="text-sm text-neutral-text-secondary mt-1">
              {branches.length === 1 
                ? `Branch: ${branches[0].name} (${branches[0].code})`
                : `Showing ${branches.length} branches: ${branches.map(b => `${b.code}`).join(', ')}`}
            </p>
          )}
        </div>
        <div className="flex gap-2 items-center">
          <button
            onClick={goToPreviousMonth}
            className="btn-secondary"
          >
            Previous
          </button>
          <button
            onClick={() => setCurrentDate(new Date())}
            className="btn-secondary"
          >
            Today
          </button>
          <button
            onClick={goToNextMonth}
            className="btn-secondary"
          >
            Next
          </button>
          {canEditSchedules && (
            <button
              onClick={handleBulkTurnOffToWorking}
              className="btn-primary ml-4"
              disabled={loading || staff.length === 0}
              title="Turn all 'Off' days to 'Working' days for this month"
            >
              Turn All Off Days to Working
            </button>
          )}
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full border-collapse border border-neutral-border">
          <thead>
            <tr>
              <th className="border border-neutral-border p-2 bg-neutral-hover font-semibold sticky left-0 z-10 text-neutral-text-primary min-w-[120px]">
                Staff Name
              </th>
              <th className="border border-neutral-border p-2 bg-neutral-hover font-semibold sticky left-[120px] z-10 text-neutral-text-primary min-w-[80px]">
                Nickname
              </th>
              <th className="border border-neutral-border p-2 bg-neutral-hover font-semibold sticky left-[200px] z-10 text-neutral-text-primary min-w-[100px]">
                Position
              </th>
              {branches.length > 1 && (
                <th className="border border-neutral-border p-2 bg-neutral-hover font-semibold sticky left-[300px] z-10 text-neutral-text-primary min-w-[70px]">
                  Branch Code
                </th>
              )}
              {daysInMonth.map((day) => (
                <th
                  key={day.toISOString()}
                  className="border border-neutral-border p-2 bg-neutral-hover font-semibold min-w-[80px] text-neutral-text-primary"
                >
                  <div className="text-xs text-neutral-text-secondary">{format(day, 'EEE')}</div>
                  <div className="text-sm">{format(day, 'd')}</div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {/* Branch Staff */}
            {(staff || []).map((staffMember) => (
              <tr key={staffMember.id}>
                <td className="border border-neutral-border p-2 bg-neutral-bg-secondary sticky left-0 z-10 font-medium text-neutral-text-primary min-w-[120px]">
                  <span className="text-sm truncate block" title={staffMember.name}>{staffMember.name}</span>
                </td>
                <td className="border border-neutral-border p-2 bg-neutral-bg-secondary sticky left-[120px] z-10 text-neutral-text-primary min-w-[80px]">
                  <span className="text-sm font-semibold truncate block" title={staffMember.nickname || ''}>{staffMember.nickname || '-'}</span>
                </td>
                <td className="border border-neutral-border p-2 bg-neutral-bg-secondary sticky left-[200px] z-10 text-neutral-text-primary min-w-[100px]">
                  <span className="text-sm text-neutral-text-secondary italic truncate block" title={staffMember.position?.name || ''}>{staffMember.position?.name || '-'}</span>
                </td>
                {branches.length > 1 && (
                  <td className="border border-neutral-border p-2 bg-neutral-bg-secondary sticky left-[300px] z-10 text-neutral-text-primary min-w-[70px]">
                    <span className="text-sm font-mono font-semibold">{staffMember.branchCode || '-'}</span>
                  </td>
                )}
                {daysInMonth.map((day) => {
                  const schedule = getScheduleForCell(staffMember.id, day);
                  const scheduleStatus: ScheduleStatus = schedule?.schedule_status || 
                    (schedule?.is_working_day ? 'working' : 'off');
                  const isToday = isSameDay(day, new Date());
                  
                  // Determine cell styling based on schedule status
                  let bgColor = 'bg-neutral-bg-secondary';
                  let hoverColor = 'hover:bg-neutral-hover';
                  let indicator = '';
                  
                  switch (scheduleStatus) {
                    case 'working':
                      bgColor = 'bg-green-50';
                      hoverColor = 'hover:bg-green-100';
                      indicator = '✓';
                      break;
                    case 'leave':
                      bgColor = 'bg-yellow-50';
                      hoverColor = 'hover:bg-yellow-100';
                      indicator = 'L';
                      break;
                    case 'sick_leave':
                      bgColor = 'bg-red-50';
                      hoverColor = 'hover:bg-red-100';
                      indicator = 'S';
                      break;
                    case 'off':
                    default:
                      bgColor = 'bg-neutral-bg-secondary';
                      hoverColor = 'hover:bg-neutral-hover';
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
                      onClick={(e) => handleCellClick(staffMember.id, day, e)}
                      onContextMenu={(e) => handleCellRightClick(e, staffMember.id, day)}
                      className={`
                        border border-neutral-border p-1 cursor-pointer transition-colors
                        ${bgColor} ${hoverColor}
                        ${isToday ? 'ring-2 ring-blue-500' : ''}
                      `}
                      title={`${format(day, 'MMM d, yyyy')} - ${statusLabel} - Left click to cycle, Right click for menu`}
                    >
                      <div className="w-6 h-6 flex items-center justify-center text-xs font-semibold">
                        {indicator}
                      </div>
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Rotation Staff Table - Separate table below branch staff (Admin/Area Manager only) */}
      {showRotationStaff && rotationStaffForMonth.length > 0 && (
        <div className="mt-8">
          <h3 className="text-lg font-semibold text-neutral-text-primary mb-4">
            Rotation Staff Scheduling - {format(currentDate, 'MMMM yyyy')}
          </h3>
          <div className="overflow-x-auto">
            <table className="w-full border-collapse border border-neutral-border">
              <thead>
                <tr>
                  <th className="border border-neutral-border p-2 bg-blue-100 font-semibold sticky left-0 z-10 text-neutral-text-primary min-w-[120px]">
                    Rotation Staff Name
                  </th>
                  <th className="border border-neutral-border p-2 bg-blue-100 font-semibold sticky left-[120px] z-10 text-neutral-text-primary min-w-[80px]">
                    Nickname
                  </th>
                  <th className="border border-neutral-border p-2 bg-blue-100 font-semibold sticky left-[200px] z-10 text-neutral-text-primary min-w-[100px]">
                    Position
                  </th>
                  <th className="border border-neutral-border p-2 bg-blue-100 font-semibold sticky left-[300px] z-10 text-neutral-text-primary min-w-[120px]">
                    Coverage Area
                  </th>
                  {branches.length > 1 && (
                    <th className="border border-neutral-border p-2 bg-blue-100 font-semibold sticky left-[420px] z-10 text-neutral-text-primary min-w-[100px]">
                      Assigned Branch
                    </th>
                  )}
                  {daysInMonth.map((day) => (
                    <th
                      key={day.toISOString()}
                      className="border border-neutral-border p-2 bg-blue-100 font-semibold min-w-[80px] text-neutral-text-primary"
                    >
                      <div className="text-xs text-neutral-text-secondary">{format(day, 'EEE')}</div>
                      <div className="text-sm">{format(day, 'd')}</div>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {rotationStaffForMonth.map((rotationStaff) => {
                  const assignmentsForStaff = rotationAssignments.filter(
                    ra => ra.rotation_staff_id === rotationStaff.id &&
                    new Date(ra.date).getMonth() === currentDate.getMonth() &&
                    new Date(ra.date).getFullYear() === currentDate.getFullYear()
                  );
                  
                  return (
                    <tr key={`rotation-${rotationStaff.id}`} className="bg-blue-50/30">
                      <td className="border border-neutral-border p-2 bg-blue-50/50 sticky left-0 z-10 font-medium text-neutral-text-primary min-w-[120px]">
                        <div className="flex flex-col">
                          <span className="text-sm truncate" title={rotationStaff.name}>{rotationStaff.name}</span>
                          <span className="text-xs text-blue-600 font-semibold mt-1">[Rotation]</span>
                        </div>
                      </td>
                      <td className="border border-neutral-border p-2 bg-blue-50/50 sticky left-[120px] z-10 text-neutral-text-primary min-w-[80px]">
                        <span className="text-sm font-semibold truncate block" title={rotationStaff.nickname || ''}>{rotationStaff.nickname || '-'}</span>
                      </td>
                      <td className="border border-neutral-border p-2 bg-blue-50/50 sticky left-[200px] z-10 text-neutral-text-primary min-w-[100px]">
                        <span className="text-sm text-neutral-text-secondary italic truncate block" title={rotationStaff.position?.name || ''}>{rotationStaff.position?.name || '-'}</span>
                      </td>
                      <td className="border border-neutral-border p-2 bg-blue-50/50 sticky left-[300px] z-10 text-neutral-text-primary min-w-[120px]">
                        <span className="text-sm text-neutral-text-secondary truncate block" title={rotationStaff.coverage_area || ''}>{rotationStaff.coverage_area || '-'}</span>
                      </td>
                      {branches.length > 1 && (
                        <td className="border border-neutral-border p-2 bg-blue-50/50 sticky left-[420px] z-10 text-neutral-text-primary min-w-[100px]">
                          <div className="text-xs text-neutral-text-secondary">
                            {assignmentsForStaff.length > 0 ? (
                              <div className="space-y-1">
                                {[...new Set(assignmentsForStaff.map(a => {
                                  const branch = branches.find(b => b.id === a.branch_id);
                                  return branch ? `${branch.code}` : '';
                                }))].filter(Boolean).map((code, idx) => (
                                  <span key={idx} className="block">{code}</span>
                                ))}
                              </div>
                            ) : (
                              <span>-</span>
                            )}
                          </div>
                        </td>
                      )}
                      {daysInMonth.map((day) => {
                        const assignment = assignmentsForStaff.find(ra => 
                          isSameDay(new Date(ra.date), day) && branchIds.includes(ra.branch_id)
                        );
                        const isToday = isSameDay(day, new Date());
                        const hasAssignment = !!assignment;
                        const assignmentBranch = assignment ? branches.find(b => b.id === assignment.branch_id) : null;
                        
                        return (
                          <td
                            key={day.toISOString()}
                            className={`
                              border border-neutral-border p-1
                              ${hasAssignment 
                                ? assignment.assignment_level === 1 
                                  ? 'bg-blue-200' 
                                  : 'bg-blue-100'
                                : 'bg-neutral-bg-secondary'}
                              ${isToday ? 'ring-2 ring-blue-500' : ''}
                            `}
                            title={hasAssignment 
                              ? `${format(day, 'MMM d, yyyy')} - Assigned to ${assignmentBranch?.name || 'branch'} (Level ${assignment.assignment_level === 1 ? '1 - Priority' : '2 - Reserved'})`
                              : `${format(day, 'MMM d, yyyy')} - No Assignment`
                            }
                          >
                            <div className="w-6 h-6 flex items-center justify-center text-xs font-semibold">
                              {hasAssignment ? (
                                <span className={assignment.assignment_level === 1 ? 'text-blue-800 font-bold' : 'text-blue-600'}>
                                  {assignment.assignment_level === 1 ? 'P' : 'R'}
                                </span>
                              ) : ''}
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
          <div className="mt-4 flex items-center gap-4 text-sm text-neutral-text-secondary flex-wrap">
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-blue-200 border border-neutral-border flex items-center justify-center text-xs font-semibold text-blue-800 font-bold">P</div>
              <span>Level 1 (Priority)</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-blue-100 border border-neutral-border flex items-center justify-center text-xs font-semibold text-blue-600">R</div>
              <span>Level 2 (Reserved)</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-neutral-bg-secondary border border-neutral-border"></div>
              <span>No Assignment</span>
            </div>
          </div>
        </div>
      )}

      <div className="mt-4 flex items-center gap-4 text-sm text-neutral-text-secondary flex-wrap">
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-green-50 border border-neutral-border flex items-center justify-center text-xs font-semibold">✓</div>
          <span>Working Day</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-yellow-50 border border-neutral-border flex items-center justify-center text-xs font-semibold">L</div>
          <span>Leave Day</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-red-50 border border-neutral-border flex items-center justify-center text-xs font-semibold">S</div>
          <span>Sick Leave</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-neutral-bg-secondary border border-neutral-border flex items-center justify-center text-xs font-semibold">X</div>
          <span>Off Day</span>
        </div>
        <div className="ml-auto">Left click to cycle: Off → Working → Leave → Sick Leave → Off | Right click for menu</div>
      </div>

      {/* Context Menu */}
      {contextMenu && (
        <div
          className="context-menu fixed bg-white border border-neutral-border shadow-lg rounded-md py-1 z-50 min-w-[160px]"
          style={{ left: contextMenu.x, top: contextMenu.y }}
          onClick={(e) => e.stopPropagation()}
        >
          <button
            onClick={() => handleContextMenuSelect('off')}
            className="w-full text-left px-4 py-2 hover:bg-neutral-hover text-sm text-neutral-text-primary"
          >
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-neutral-bg-secondary border border-neutral-border flex items-center justify-center text-xs font-semibold">X</div>
              <span>Off Day</span>
            </div>
          </button>
          <button
            onClick={() => handleContextMenuSelect('working')}
            className="w-full text-left px-4 py-2 hover:bg-neutral-hover text-sm text-neutral-text-primary"
          >
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-green-50 border border-neutral-border flex items-center justify-center text-xs font-semibold">✓</div>
              <span>Working Day</span>
            </div>
          </button>
          <button
            onClick={() => handleContextMenuSelect('leave')}
            className="w-full text-left px-4 py-2 hover:bg-neutral-hover text-sm text-neutral-text-primary"
          >
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-yellow-50 border border-neutral-border flex items-center justify-center text-xs font-semibold">L</div>
              <span>Leave Day</span>
            </div>
          </button>
          <button
            onClick={() => handleContextMenuSelect('sick_leave')}
            className="w-full text-left px-4 py-2 hover:bg-neutral-hover text-sm text-neutral-text-primary"
          >
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-red-50 border border-neutral-border flex items-center justify-center text-xs font-semibold">S</div>
              <span>Sick Leave</span>
            </div>
          </button>
        </div>
      )}
    </div>
  );
}

