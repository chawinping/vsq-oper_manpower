'use client';

import { useState, useEffect, useMemo, useRef } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameDay, addMonths, subMonths } from 'date-fns';
import { rotationApi, RotationAssignment, RotationStaffSchedule } from '@/lib/api/rotation';
import { staffApi, Staff } from '@/lib/api/staff';
import { positionApi, Position } from '@/lib/api/position';
import { areaOfOperationApi, AreaOfOperation } from '@/lib/api/areaOfOperation';
import { branchApi, Branch } from '@/lib/api/branch';
import { effectiveBranchApi } from '@/lib/api/effectiveBranch';
import { ScheduleStatus } from '@/lib/api/schedule';

export default function RotationStaffCalendar() {
  const [currentDate, setCurrentDate] = useState(new Date());
  const [rotationStaff, setRotationStaff] = useState<Staff[]>([]);
  const [assignments, setAssignments] = useState<RotationAssignment[]>([]);
  const [schedules, setSchedules] = useState<RotationStaffSchedule[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [areasOfOperation, setAreasOfOperation] = useState<AreaOfOperation[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [filterPositionId, setFilterPositionId] = useState<string>('');
  const [filterAreaOfOperationId, setFilterAreaOfOperationId] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; staffId: string; date: Date } | null>(null);
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const savedScrollPositionsRef = useRef<{
    horizontal: number;
    pageVertical: number;
  } | null>(null);

  // Memoize month calculations
  const monthStart = useMemo(() => startOfMonth(currentDate), [currentDate]);
  const monthEnd = useMemo(() => endOfMonth(currentDate), [currentDate]);
  const daysInMonth = useMemo(() => eachDayOfInterval({ start: monthStart, end: monthEnd }), [monthStart, monthEnd]);
  const monthKey = useMemo(() => format(currentDate, 'yyyy-MM'), [currentDate]);

  // Close context menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (contextMenu) {
        const target = e.target as HTMLElement;
        if (!target.closest('.context-menu')) {
          setContextMenu(null);
        }
      }
    };
    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  }, [contextMenu]);

  // Restore scroll positions after loading completes
  useEffect(() => {
    if (!loading && savedScrollPositionsRef.current) {
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          const saved = savedScrollPositionsRef.current;
          if (!saved) return;
          
          window.scrollTo(0, saved.pageVertical);
          
          if (scrollContainerRef.current) {
            scrollContainerRef.current.scrollLeft = saved.horizontal;
          }
          
          savedScrollPositionsRef.current = null;
        });
      });
    }
  }, [loading]);

  useEffect(() => {
    loadData();
  }, [monthKey, filterPositionId, filterAreaOfOperationId]);

  const loadData = async () => {
    // Save scroll positions before loading
    savedScrollPositionsRef.current = {
      horizontal: scrollContainerRef.current?.scrollLeft ?? 0,
      pageVertical: window.scrollY ?? 0,
    };

    try {
      setLoading(true);
      const startDateStr = format(monthStart, 'yyyy-MM-dd');
      const endDateStr = format(monthEnd, 'yyyy-MM-dd');

      const [staffData, assignmentsData, schedulesData, positionsData, areasData, branchesData] = await Promise.all([
        staffApi.list({
          staff_type: 'rotation',
          position_id: filterPositionId || undefined,
          area_of_operation_id: filterAreaOfOperationId || undefined,
        }),
        rotationApi.getAssignments({
          start_date: startDateStr,
          end_date: endDateStr,
        }),
        rotationApi.getSchedules({
          start_date: startDateStr,
          end_date: endDateStr,
        }),
        positionApi.list(),
        areaOfOperationApi.list(),
        branchApi.list(),
      ]);

      setRotationStaff(staffData || []);
      setAssignments(assignmentsData || []);
      setSchedules(schedulesData || []);
      setPositions(positionsData || []);
      setAreasOfOperation(areasData || []);
      setBranches(branchesData || []);
    } catch (err: any) {
      console.error('Failed to load data:', err);
      alert(`Failed to load data: ${err.response?.data?.error || err.message || 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  };

  // Normalize date string for comparison (handles both 'yyyy-MM-dd' and ISO format)
  const normalizeDateStr = (dateStr: string): string => {
    // If it's already in 'yyyy-MM-dd' format, return as is
    if (/^\d{4}-\d{2}-\d{2}$/.test(dateStr)) {
      return dateStr;
    }
    // If it's in ISO format (e.g., '2026-01-22T00:00:00Z'), extract the date part
    if (dateStr.includes('T')) {
      return dateStr.split('T')[0];
    }
    return dateStr;
  };

  // Get schedule status for a rotation staff on a specific date
  const getScheduleStatus = (staffId: string, date: Date): ScheduleStatus | null => {
    const dateStr = format(date, 'yyyy-MM-dd');
    const schedule = schedules.find(
      s => s.rotation_staff_id === staffId && normalizeDateStr(s.date) === dateStr
    );
    
    if (!schedule) {
      return null; // No schedule set for this date
    }
    
    return schedule.schedule_status as ScheduleStatus;
  };

  // Get schedule record for a rotation staff on a specific date
  const getSchedule = (staffId: string, date: Date): RotationStaffSchedule | null => {
    const dateStr = format(date, 'yyyy-MM-dd');
    return schedules.find(
      s => s.rotation_staff_id === staffId && normalizeDateStr(s.date) === dateStr
    ) || null;
  };

  // Get assignments for a rotation staff on a specific date (for branch display)
  const getAssignmentsForDate = (staffId: string, date: Date): RotationAssignment[] => {
    const dateStr = format(date, 'yyyy-MM-dd');
    return assignments.filter(
      a => a.rotation_staff_id === staffId && a.date === dateStr
    );
  };

  // Get next schedule status in cycle - Changed to: off -> working -> leave -> sick_leave -> off
  const getNextScheduleStatus = (current: ScheduleStatus | null): ScheduleStatus => {
    if (!current) return 'working'; // Default to working if no current status
    
    const cycle: ScheduleStatus[] = ['off', 'working', 'leave', 'sick_leave'];
    const currentIndex = cycle.indexOf(current);
    return cycle[(currentIndex + 1) % cycle.length];
  };

  const handleCellClick = async (staffId: string, date: Date, e?: React.MouseEvent) => {
    console.log('handleCellClick called', { staffId, date: format(date, 'yyyy-MM-dd') });
    
    // Close context menu if open (but continue with the click action)
    if (contextMenu) {
      setContextMenu(null);
    }

    const isPast = date < new Date(new Date().setHours(0, 0, 0, 0));
    if (isPast) {
      console.log('Past date, ignoring');
      return;
    }

    const dateStr = format(date, 'yyyy-MM-dd');
    const currentStatus = getScheduleStatus(staffId, date);
    const nextStatus = getNextScheduleStatus(currentStatus);
    const existingSchedule = getSchedule(staffId, date);

    try {
      if (existingSchedule) {
        // Update existing schedule
        await rotationApi.updateSchedule(existingSchedule.id, nextStatus);
      } else {
        // Create new schedule
        await rotationApi.setSchedule({
          rotation_staff_id: staffId,
          date: dateStr,
          schedule_status: nextStatus,
        });
      }
      
      await loadData();
    } catch (err: any) {
      console.error('Failed to update schedule:', err);
      alert(`Failed to update schedule: ${err.response?.data?.error || err.message || 'Unknown error'}`);
    }
  };

  const handleCellRightClick = (e: React.MouseEvent, staffId: string, date: Date) => {
    e.preventDefault();
    setContextMenu({ x: e.clientX, y: e.clientY, staffId, date });
  };

  const handleContextMenuSelectSchedule = async (status: ScheduleStatus) => {
    if (!contextMenu) return;
    
    const dateStr = format(contextMenu.date, 'yyyy-MM-dd');
    const isPast = contextMenu.date < new Date(new Date().setHours(0, 0, 0, 0));
    if (isPast) {
      setContextMenu(null);
      return;
    }

    try {
      const existingSchedule = getSchedule(contextMenu.staffId, contextMenu.date);
      
      if (existingSchedule) {
        // Update existing schedule
        await rotationApi.updateSchedule(existingSchedule.id, status);
      } else {
        // Create new schedule
        await rotationApi.setSchedule({
          rotation_staff_id: contextMenu.staffId,
          date: dateStr,
          schedule_status: status,
        });
      }
      
      await loadData();
      setContextMenu(null);
    } catch (err: any) {
      console.error('Failed to update schedule:', err);
      alert(`Failed to update schedule: ${err.response?.data?.error || err.message || 'Unknown error'}`);
      setContextMenu(null);
    }
  };

  const handleContextMenuAssignBranch = async () => {
    if (!contextMenu) return;
    
    const dateStr = format(contextMenu.date, 'yyyy-MM-dd');
    const isPast = contextMenu.date < new Date(new Date().setHours(0, 0, 0, 0));
    if (isPast) {
      setContextMenu(null);
      return;
    }

    // Check if schedule status is "off" - cannot assign branch
    const scheduleStatus = getScheduleStatus(contextMenu.staffId, contextMenu.date);
    if (scheduleStatus === 'off') {
      alert('Cannot assign branch when schedule status is "off". Please set the day to "working", "leave", or "sick_leave" first.');
      setContextMenu(null);
      return;
    }

    try {
      const effectiveBranches = await effectiveBranchApi.getByRotationStaffID(contextMenu.staffId);
      
      if (effectiveBranches.length === 0) {
        alert('This rotation staff has no effective branches assigned. Please assign branches in Rotation Staff Profile first.');
        setContextMenu(null);
        return;
      }
      
      // Use the first effective branch (in future, could show a branch selector)
      const targetBranchId = effectiveBranches[0].branch_id;
      const assignmentLevel = effectiveBranches[0].level || 1;
      
      // Check if assignment already exists for this specific branch
      const dayAssignments = assignments.filter(
        a => a.rotation_staff_id === contextMenu.staffId && a.date === dateStr
      );
      const existingAssignmentForBranch = dayAssignments.find(
        a => a.branch_id === targetBranchId
      );
      
      if (!existingAssignmentForBranch) {
        // Create assignment
        await rotationApi.assign({
          rotation_staff_id: contextMenu.staffId,
          branch_id: targetBranchId,
          date: dateStr,
          assignment_level: assignmentLevel,
        });
      }
      
      await loadData();
      setContextMenu(null);
    } catch (err: any) {
      console.error('Failed to assign branch:', err);
      const errorMsg = err.response?.data?.error || err.message || 'Unknown error';
      alert(`Failed to assign branch: ${errorMsg}`);
      setContextMenu(null);
    }
  };

  const handleContextMenuRemoveBranch = async () => {
    if (!contextMenu) return;
    
    const dateStr = format(contextMenu.date, 'yyyy-MM-dd');
    const isPast = contextMenu.date < new Date(new Date().setHours(0, 0, 0, 0));
    if (isPast) {
      setContextMenu(null);
      return;
    }

    try {
      const dayAssignments = assignments.filter(
        a => a.rotation_staff_id === contextMenu.staffId && a.date === dateStr
      );
      
      if (dayAssignments.length === 0) {
        alert('No branch assignments found for this day.');
        setContextMenu(null);
        return;
      }

      // Remove all assignments for this staff on this date
      await Promise.all(
        dayAssignments.map(assignment => rotationApi.removeAssignment(assignment.id))
      );
      
      await loadData();
      setContextMenu(null);
    } catch (err: any) {
      console.error('Failed to remove branch assignment:', err);
      alert(`Failed to remove branch assignment: ${err.response?.data?.error || err.message || 'Unknown error'}`);
      setContextMenu(null);
    }
  };

  const handleCloseContextMenu = () => {
    setContextMenu(null);
  };

  // Filter rotation staff based on filters
  const filteredRotationStaff = useMemo(() => {
    return rotationStaff.filter(staff => {
      if (filterPositionId && staff.position_id !== filterPositionId) return false;
      if (filterAreaOfOperationId && (staff as any).area_of_operation_id !== filterAreaOfOperationId) return false;
      return true;
    });
  }, [rotationStaff, filterPositionId, filterAreaOfOperationId]);

  const handleBulkTurnOffToWorking = async () => {
    const confirmed = window.confirm(
      `Are you sure you want to turn all "Off" days and "Not Assigned" days to "Working" days for ${format(currentDate, 'MMMM yyyy')}?\n\n` +
      `This will affect all rotation staff members for all days in this month that are currently set to "Off" or are "Not Assigned".`
    );
    
    if (!confirmed) {
      return;
    }

    try {
      setLoading(true);
      const monthStart = startOfMonth(currentDate);
      const monthEnd = endOfMonth(currentDate);
      const daysInMonth = eachDayOfInterval({ start: monthStart, end: monthEnd });
      
      // Get all rotation staff
      const allRotationStaff = filteredRotationStaff;
      
      if (allRotationStaff.length === 0) {
        alert('No rotation staff members found to update.');
        setLoading(false);
        return;
      }
      
      // Count how many updates we'll make
      let updateCount = 0;
      let errorCount = 0;
      const updatePromises: Promise<void>[] = [];
      
      console.log(`Starting bulk update for ${allRotationStaff.length} staff members across ${daysInMonth.length} days`);
      
      // For each day in the month
      for (const day of daysInMonth) {
        const dateStr = format(day, 'yyyy-MM-dd');
        const isPast = day < new Date(new Date().setHours(0, 0, 0, 0));
        if (isPast) continue; // Skip past dates
        
        // For each rotation staff member
        for (const staffMember of allRotationStaff) {
          // Check if schedule exists and get current status
          const currentStatus = getScheduleStatus(staffMember.id, day);
          
          // Update if status is "off" or if no schedule exists (not assigned days)
          if (!currentStatus || currentStatus === 'off') {
            const existingSchedule = getSchedule(staffMember.id, day);
            
            // Create update promise
            const updatePromise = (async () => {
              try {
                if (existingSchedule) {
                  console.log(`Updating existing schedule for staff ${staffMember.id}, date ${dateStr} from ${currentStatus} to working`);
                  const result = await rotationApi.updateSchedule(existingSchedule.id, 'working');
                  console.log(`Successfully updated schedule:`, result);
                  updateCount++;
                } else {
                  console.log(`Creating new schedule for staff ${staffMember.id}, date ${dateStr} as working`);
                  const result = await rotationApi.setSchedule({
                    rotation_staff_id: staffMember.id,
                    date: dateStr,
                    schedule_status: 'working',
                  });
                  console.log(`Successfully created schedule:`, result);
                  updateCount++;
                }
              } catch (err: any) {
                console.error(`Failed to update schedule for staff ${staffMember.id}, date ${dateStr}:`, err);
                console.error('Error details:', {
                  message: err.message,
                  response: err.response?.data,
                  status: err.response?.status,
                });
                errorCount++;
              }
            })();
            
            updatePromises.push(updatePromise);
          }
        }
      }
      
      console.log(`Found ${updatePromises.length} schedules to update`);
      
      if (updatePromises.length === 0) {
        alert('No schedules found that need to be updated. All days are already set to "Working", "Leave", or "Sick Leave".');
        setLoading(false);
        return;
      }
      
      // Execute all updates in parallel (with batching to avoid overwhelming the server)
      const batchSize = 50;
      for (let i = 0; i < updatePromises.length; i += batchSize) {
        const batch = updatePromises.slice(i, i + batchSize);
        await Promise.all(batch);
        console.log(`Completed batch ${Math.floor(i / batchSize) + 1} of ${Math.ceil(updatePromises.length / batchSize)}`);
      }
      
      console.log(`Bulk update completed. Updated: ${updateCount}, Errors: ${errorCount}`);
      
      // Small delay to ensure backend has processed all updates
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // Reload data to reflect changes
      console.log('Reloading data...');
      
      // Force reload by clearing and reloading schedules
      const startDateStr = format(monthStart, 'yyyy-MM-dd');
      const endDateStr = format(monthEnd, 'yyyy-MM-dd');
      
      try {
        const freshSchedules = await rotationApi.getSchedules({
          start_date: startDateStr,
          end_date: endDateStr,
        });
        console.log(`Loaded ${freshSchedules.length} schedules after update`);
        setSchedules(freshSchedules || []);
      } catch (reloadErr: any) {
        console.error('Failed to reload schedules:', reloadErr);
        // Fallback to full reload
        await loadData();
      }
      
      console.log('Data reloaded');
      
      let message = `Successfully updated ${updateCount} schedule(s) to "Working" for ${format(currentDate, 'MMMM yyyy')}.`;
      if (errorCount > 0) {
        message += `\n\nNote: ${errorCount} schedule(s) could not be updated. Please check the console for details.`;
      }
      alert(message);
    } catch (err: any) {
      console.error('Failed to bulk update schedules:', err);
      alert(`Failed to update schedules: ${err.response?.data?.error || err.message || 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  };

  const goToPreviousMonth = () => setCurrentDate(subMonths(currentDate, 1));
  const goToNextMonth = () => setCurrentDate(addMonths(currentDate, 1));
  const goToToday = () => setCurrentDate(new Date());

  // Show loading only on initial load
  if (loading && rotationStaff.length === 0 && assignments.length === 0 && schedules.length === 0) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  // Count unique branches assigned in this month
  const assignedBranchIds = [...new Set(assignments.map(a => a.branch_id))];
  const assignedBranches = branches.filter(b => assignedBranchIds.includes(b.id));

  return (
    <div className="w-full">
      {/* Header with month navigation */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-neutral-text-primary">
            Rotation Staff Scheduling - {format(currentDate, 'MMMM yyyy')}
          </h2>
          {assignedBranches.length > 0 && (
            <p className="text-sm text-neutral-text-secondary mt-1">
              {assignedBranches.length === 1 
                ? `Branch: ${assignedBranches[0].name} (${assignedBranches[0].code})`
                : `Showing ${assignedBranches.length} branches: ${assignedBranches.map(b => `${b.code}`).join(', ')}`}
            </p>
          )}
          {filteredRotationStaff.length > 0 && (
            <p className="text-sm text-neutral-text-secondary mt-1">
              {filteredRotationStaff.length} rotation staff member{filteredRotationStaff.length !== 1 ? 's' : ''}
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
            onClick={goToToday}
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
          <button
            onClick={handleBulkTurnOffToWorking}
            className="btn-primary ml-4"
            disabled={loading || filteredRotationStaff.length === 0}
            title="Turn all 'Off' days and 'Not Assigned' days to 'Working' days for this month"
          >
            Turn All Off Days to Working
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="mb-4 flex gap-3 flex-wrap">
        <div className="min-w-[180px]">
          <label htmlFor="filter-position" className="block text-xs font-medium text-neutral-text-primary mb-1">
            Filter by Position
          </label>
          <select
            id="filter-position"
            value={filterPositionId}
            onChange={(e) => setFilterPositionId(e.target.value)}
            className="input-field w-full text-sm"
          >
            <option value="">All Positions</option>
            {positions.map((position) => (
              <option key={position.id} value={position.id}>
                {position.name}
              </option>
            ))}
          </select>
        </div>
        
        <div className="min-w-[180px]">
          <label htmlFor="filter-area" className="block text-xs font-medium text-neutral-text-primary mb-1">
            Filter by Area
          </label>
          <select
            id="filter-area"
            value={filterAreaOfOperationId}
            onChange={(e) => setFilterAreaOfOperationId(e.target.value)}
            className="input-field w-full text-sm"
          >
            <option value="">All Areas</option>
            {areasOfOperation.map((area) => (
              <option key={area.id} value={area.id}>
                {area.name} ({area.code})
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Calendar Table */}
      <div className="card overflow-x-auto" ref={scrollContainerRef}>
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
            {filteredRotationStaff.length === 0 ? (
              <tr>
                <td colSpan={daysInMonth.length + 3} className="p-8 text-center text-neutral-text-secondary">
                  No rotation staff found
                </td>
              </tr>
            ) : (
              filteredRotationStaff.map((staff) => {
                const position = positions.find((p) => p.id === staff.position_id);
                return (
                  <tr key={staff.id}>
                    <td className="border border-neutral-border p-2 bg-neutral-bg-secondary sticky left-0 z-10 font-medium text-neutral-text-primary min-w-[120px]">
                      <span className="text-sm truncate block" title={staff.name}>{staff.name}</span>
                    </td>
                    <td className="border border-neutral-border p-2 bg-neutral-bg-secondary sticky left-[120px] z-10 text-neutral-text-primary min-w-[80px]">
                      <span className="text-sm font-semibold truncate block" title={staff.nickname || ''}>{staff.nickname || '-'}</span>
                    </td>
                    <td className="border border-neutral-border p-2 bg-neutral-bg-secondary sticky left-[200px] z-10 text-neutral-text-primary min-w-[100px]">
                      <span className="text-sm text-neutral-text-secondary italic truncate block" title={position?.name || ''}>{position?.name || '-'}</span>
                    </td>
                    {daysInMonth.map((day) => {
                      const scheduleStatus = getScheduleStatus(staff.id, day);
                      const dayAssignments = getAssignmentsForDate(staff.id, day);
                      const isPast = day < new Date(new Date().setHours(0, 0, 0, 0));
                      const isToday = isSameDay(day, new Date());
                      
                      // Get branch codes for this date
                      const branchCodes = dayAssignments
                        .map(a => {
                          const branch = branches.find(b => b.id === a.branch_id);
                          return branch?.code;
                        })
                        .filter(Boolean) as string[];
                      
                      // Determine cell styling based on schedule status
                      let bgColor = 'bg-neutral-bg-secondary';
                      let hoverColor = 'hover:bg-neutral-hover';
                      let indicator = '';
                      let statusLabel = 'Not Assigned';
                      
                      if (scheduleStatus) {
                        switch (scheduleStatus) {
                          case 'working':
                            bgColor = 'bg-green-50';
                            hoverColor = 'hover:bg-green-100';
                            indicator = '✓';
                            statusLabel = 'Working';
                            break;
                          case 'leave':
                            bgColor = 'bg-yellow-50';
                            hoverColor = 'hover:bg-yellow-100';
                            indicator = 'L';
                            statusLabel = 'Leave';
                            break;
                          case 'sick_leave':
                            bgColor = 'bg-red-50';
                            hoverColor = 'hover:bg-red-100';
                            indicator = 'S';
                            statusLabel = 'Sick Leave';
                            break;
                          case 'off':
                            bgColor = 'bg-neutral-bg-secondary';
                            hoverColor = 'hover:bg-neutral-hover';
                            indicator = 'X';
                            statusLabel = 'Off';
                            break;
                        }
                      } else {
                        bgColor = 'bg-gray-50';
                        hoverColor = 'hover:bg-gray-100';
                        indicator = '○';
                      }
                      
                      // Build tooltip with branch information
                      let tooltip = `${format(day, 'MMM d, yyyy')} - ${statusLabel}`;
                      if (branchCodes.length > 0) {
                        tooltip += ` | Branches: ${branchCodes.join(', ')}`;
                      }
                      tooltip += isPast ? ' (Past date)' : ' - Left click to cycle schedule status, Right click for menu';
                      
                      return (
                        <td
                          key={day.toISOString()}
                          onClick={(e) => {
                            console.log('Cell clicked directly', { staffId: staff.id, date: format(day, 'yyyy-MM-dd') });
                            if (!isPast) {
                              handleCellClick(staff.id, day, e);
                            }
                          }}
                          onContextMenu={(e) => {
                            if (!isPast) {
                              handleCellRightClick(e, staff.id, day);
                            }
                          }}
                          className={`
                            border border-neutral-border p-1 transition-colors
                            ${isPast ? 'opacity-60 cursor-not-allowed' : 'cursor-pointer'}
                            ${bgColor} ${isPast ? '' : hoverColor}
                            ${isToday ? 'ring-2 ring-blue-500' : ''}
                          `}
                          style={{ userSelect: 'none' }}
                          title={tooltip}
                        >
                          <div 
                            className="w-6 h-6 flex items-center justify-center text-xs font-semibold mx-auto"
                            style={{ pointerEvents: 'none' }}
                          >
                            {indicator}
                          </div>
                          {branchCodes.length > 0 && (
                            <div 
                              className="text-[8px] text-neutral-text-secondary mt-0.5 truncate"
                              style={{ pointerEvents: 'none' }}
                            >
                              {branchCodes.length === 1 ? branchCodes[0] : `${branchCodes.length} br`}
                            </div>
                          )}
                        </td>
                      );
                    })}
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {/* Legend */}
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
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-gray-50 border border-neutral-border flex items-center justify-center text-xs font-semibold">○</div>
          <span>Not Assigned</span>
        </div>
        <div className="ml-auto">Left click to cycle schedule status: Off → Working → Leave → Sick Leave → Off | Right click for menu</div>
      </div>

      {/* Context Menu */}
      {contextMenu && (() => {
        const scheduleStatus = getScheduleStatus(contextMenu.staffId, contextMenu.date);
        const dayAssignments = getAssignmentsForDate(contextMenu.staffId, contextMenu.date);
        const canAssignBranch = scheduleStatus !== 'off' && scheduleStatus !== null;
        
        return (
          <div
            className="context-menu fixed bg-white border border-neutral-border shadow-lg rounded-md py-1 z-50 min-w-[200px]"
            style={{ left: contextMenu.x, top: contextMenu.y }}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="px-3 py-2 text-xs font-semibold text-neutral-text-secondary border-b border-neutral-border">
              Schedule Status
            </div>
            <button
              onClick={() => handleContextMenuSelectSchedule('off')}
              className="w-full text-left px-4 py-2 hover:bg-neutral-hover text-sm text-neutral-text-primary"
            >
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-neutral-bg-secondary border border-neutral-border flex items-center justify-center text-xs font-semibold">X</div>
                <span>Off Day</span>
              </div>
            </button>
            <button
              onClick={() => handleContextMenuSelectSchedule('working')}
              className="w-full text-left px-4 py-2 hover:bg-neutral-hover text-sm text-neutral-text-primary"
            >
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-green-50 border border-neutral-border flex items-center justify-center text-xs font-semibold">✓</div>
                <span>Working Day</span>
              </div>
            </button>
            <button
              onClick={() => handleContextMenuSelectSchedule('leave')}
              className="w-full text-left px-4 py-2 hover:bg-neutral-hover text-sm text-neutral-text-primary"
            >
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-yellow-50 border border-neutral-border flex items-center justify-center text-xs font-semibold">L</div>
                <span>Leave Day</span>
              </div>
            </button>
            <button
              onClick={() => handleContextMenuSelectSchedule('sick_leave')}
              className="w-full text-left px-4 py-2 hover:bg-neutral-hover text-sm text-neutral-text-primary"
            >
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-red-50 border border-neutral-border flex items-center justify-center text-xs font-semibold">S</div>
                <span>Sick Leave</span>
              </div>
            </button>
            
            <div className="px-3 py-2 text-xs font-semibold text-neutral-text-secondary border-t border-neutral-border mt-1">
              Branch Assignment
            </div>
            {canAssignBranch ? (
              <>
                <button
                  onClick={handleContextMenuAssignBranch}
                  className="w-full text-left px-4 py-2 hover:bg-neutral-hover text-sm text-neutral-text-primary"
                  disabled={dayAssignments.length > 0}
                >
                  <span className={dayAssignments.length > 0 ? 'text-neutral-text-secondary' : ''}>
                    {dayAssignments.length > 0 ? 'Already Assigned' : 'Assign Branch'}
                  </span>
                </button>
                {dayAssignments.length > 0 && (
                  <button
                    onClick={handleContextMenuRemoveBranch}
                    className="w-full text-left px-4 py-2 hover:bg-neutral-hover text-sm text-red-600"
                  >
                    Remove Branch Assignment
                  </button>
                )}
              </>
            ) : (
              <div className="px-4 py-2 text-sm text-neutral-text-secondary italic">
                {scheduleStatus === 'off' 
                  ? 'Set schedule status to "working", "leave", or "sick_leave" to assign branch'
                  : 'Set schedule status first'}
              </div>
            )}
          </div>
        );
      })()}
    </div>
  );
}
