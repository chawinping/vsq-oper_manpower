'use client';

import { useState, useEffect, useMemo } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameDay, addMonths, subMonths, getDay, startOfWeek, endOfWeek, eachWeekOfInterval, addDays } from 'date-fns';
import { doctorApi, Doctor, DoctorAssignment, DoctorDefaultSchedule, DoctorScheduleOverride } from '@/lib/api/doctor';
import { branchApi, Branch } from '@/lib/api/branch';

interface DoctorScheduleCalendarProps {
  doctor: Doctor;
  branches: Branch[];
  currentDate: Date;
  onDateChange: (date: Date) => void;
}

interface ScheduleInfo {
  branch: Branch | null;
  isOff: boolean;
  source: 'override' | 'default' | 'none';
  sourceType?: 'working' | 'off';
  expectedRevenue?: number;
}

export default function DoctorScheduleCalendar({
  doctor,
  branches,
  currentDate,
  onDateChange,
}: DoctorScheduleCalendarProps) {
  const [defaultSchedules, setDefaultSchedules] = useState<DoctorDefaultSchedule[]>([]);
  const [overrides, setOverrides] = useState<DoctorScheduleOverride[]>([]);
  const [assignments, setAssignments] = useState<DoctorAssignment[]>([]);
  const [loading, setLoading] = useState(true);

  const year = currentDate.getFullYear();
  const month = currentDate.getMonth() + 1;

  useEffect(() => {
    loadSchedule();
  }, [doctor.id, year, month]);

  const loadSchedule = async () => {
    try {
      setLoading(true);
      const monthStart = startOfMonth(currentDate);
      const monthEnd = endOfMonth(currentDate);
      
      // Load all three data sources in parallel
      const [defaultSchedulesData, overridesData, assignmentsData] = await Promise.all([
        doctorApi.getDefaultSchedules(doctor.id),
        doctorApi.getScheduleOverrides({
          doctor_id: doctor.id,
          start_date: format(monthStart, 'yyyy-MM-dd'),
          end_date: format(monthEnd, 'yyyy-MM-dd'),
        }),
        doctorApi.getMonthlySchedule(doctor.id, year, month).then(data => data.assignments || []),
      ]);

      setDefaultSchedules(defaultSchedulesData || []);
      setOverrides(overridesData || []);
      setAssignments(assignmentsData || []);
    } catch (error) {
      console.error('Failed to load schedule:', error);
    } finally {
      setLoading(false);
    }
  };

  // Get final schedule for a specific date (merged default + override)
  const getScheduleForDate = (date: Date): ScheduleInfo => {
    const dateStr = format(date, 'yyyy-MM-dd');
    
    // Priority 1: Check for override
    const override = overrides.find(o => isSameDay(new Date(o.date), date));
    if (override) {
      if (override.type === 'off') {
        return {
          branch: null,
          isOff: true,
          source: 'override',
          sourceType: 'off',
        };
      } else {
        // Working override
        const branch = override.branch_id ? branches.find(b => b.id === override.branch_id) : null;
        // Check for expected revenue from assignment
        const assignment = assignments.find(a => isSameDay(new Date(a.date), date));
        return {
          branch: branch || null,
          isOff: false,
          source: 'override',
          sourceType: 'working',
          expectedRevenue: assignment?.expected_revenue,
        };
      }
    }

    // Priority 2: Check default schedule for day of week
    const dayOfWeek = getDay(date); // 0=Sunday, 1=Monday, ..., 6=Saturday
    const defaultSchedule = defaultSchedules.find(s => s.day_of_week === dayOfWeek);
    if (defaultSchedule) {
      const branch = branches.find(b => b.id === defaultSchedule.branch_id);
      // Check for expected revenue from assignment
      const assignment = assignments.find(a => isSameDay(new Date(a.date), date));
      return {
        branch: branch || null,
        isOff: false,
        source: 'default',
        expectedRevenue: assignment?.expected_revenue,
      };
    }

    // Priority 3: No schedule (off day)
    return {
      branch: null,
      isOff: true,
      source: 'none',
    };
  };

  const goToPreviousMonth = () => {
    onDateChange(subMonths(currentDate, 1));
  };

  const goToNextMonth = () => {
    onDateChange(addMonths(currentDate, 1));
  };

  // Calculate summary statistics
  const summaryStats = useMemo(() => {
    const monthStart = startOfMonth(currentDate);
    const monthEnd = endOfMonth(currentDate);
    const daysInMonth = eachDayOfInterval({ start: monthStart, end: monthEnd });
    
    let workingDays = 0;
    let overrideDays = 0;
    let defaultDays = 0;
    let offDays = 0;

    daysInMonth.forEach(day => {
      const schedule = getScheduleForDate(day);
      if (schedule.isOff) {
        offDays++;
      } else {
        workingDays++;
        if (schedule.source === 'override') {
          overrideDays++;
        } else if (schedule.source === 'default') {
          defaultDays++;
        }
      }
    });

    return { workingDays, overrideDays, defaultDays, offDays };
  }, [currentDate, defaultSchedules, overrides, assignments, branches]);

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  const monthStart = startOfMonth(currentDate);
  const monthEnd = endOfMonth(currentDate);
  
  // Get all weeks that contain days from this month
  const weeks = eachWeekOfInterval(
    { start: monthStart, end: monthEnd },
    { weekStartsOn: 0 } // Sunday = 0
  );

  // Generate days for each week
  const weekDays = weeks.map(weekStart => {
    const weekEnd = endOfWeek(weekStart, { weekStartsOn: 0 });
    const days = eachDayOfInterval({ start: weekStart, end: weekEnd });
    return days;
  });

  // Day names for header
  const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

  return (
    <div className="w-full">
      <div className="card">
        <div className="p-6">
          <div className="mb-6 flex items-center justify-between">
            <div>
              <h2 className="text-xl font-semibold text-neutral-text-primary">
                {doctor.name} {doctor.code ? `(${doctor.code})` : ''} - Schedule
              </h2>
              <p className="text-sm text-neutral-text-secondary mt-1">
                {format(monthStart, 'MMMM yyyy')} • Read-only view
              </p>
              {doctor.preferences && (
                <p className="text-xs text-neutral-text-secondary italic mt-1">{doctor.preferences}</p>
              )}
            </div>
            <div className="flex gap-2">
              <button onClick={goToPreviousMonth} className="btn-secondary">
                Previous
              </button>
              <button onClick={() => onDateChange(new Date())} className="btn-secondary">
                Today
              </button>
              <button onClick={goToNextMonth} className="btn-secondary">
                Next
              </button>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full border-collapse border border-neutral-border">
              <thead>
                <tr>
                  <th className="border border-neutral-border p-2 bg-neutral-hover font-semibold sticky left-0 z-10 text-neutral-text-primary min-w-[120px]">
                    Week
                  </th>
                  {dayNames.map((dayName, index) => (
                    <th
                      key={dayName}
                      className="border border-neutral-border p-2 bg-neutral-hover font-semibold min-w-[100px] text-neutral-text-primary"
                    >
                      {dayName}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {weekDays.map((week, weekIndex) => {
                  const weekStartDate = week[0];
                  const weekEndDate = week[week.length - 1];
                  const isCurrentMonth = (date: Date) => {
                    return date.getMonth() === currentDate.getMonth() && 
                           date.getFullYear() === currentDate.getFullYear();
                  };

                  return (
                    <tr key={weekIndex}>
                      <td className="border border-neutral-border p-2 bg-neutral-bg-secondary sticky left-0 z-10 font-medium text-neutral-text-primary min-w-[120px]">
                        <div className="text-xs">
                          <div className="font-semibold">
                            Week {weekIndex + 1}
                          </div>
                          <div className="text-neutral-text-secondary mt-1">
                            {format(weekStartDate, 'MMM d')} - {format(weekEndDate, 'MMM d')}
                          </div>
                        </div>
                      </td>
                      {week.map((day, dayIndex) => {
                        const isInCurrentMonth = isCurrentMonth(day);
                        const schedule = getScheduleForDate(day);
                        const isToday = isSameDay(day, new Date());
                        
                        // Determine styling based on source and status
                        let bgColor = 'bg-neutral-bg-secondary';
                        let borderColor = 'border-neutral-border';
                        let textColor = 'text-neutral-text-primary';
                        let sourceBadge = '';
                        
                        if (!isInCurrentMonth) {
                          bgColor = 'bg-gray-50';
                          borderColor = 'border-gray-200';
                          textColor = 'text-gray-400';
                        } else if (schedule.isOff) {
                          bgColor = 'bg-gray-100';
                          borderColor = 'border-gray-300';
                          textColor = 'text-gray-600';
                        } else if (schedule.source === 'override') {
                          // Override days - more visually distinct
                          bgColor = 'bg-orange-50';
                          borderColor = 'border-orange-300';
                          textColor = 'text-orange-900';
                          sourceBadge = 'O';
                        } else if (schedule.source === 'default') {
                          // Default schedule days
                          bgColor = 'bg-blue-50';
                          borderColor = 'border-blue-300';
                          textColor = 'text-blue-800';
                          sourceBadge = 'D';
                        }

                        // Build tooltip content
                        const tooltipParts = [
                          format(day, 'MMM d, yyyy'),
                          schedule.isOff ? 'Off Day' : schedule.branch?.name || 'Unknown Branch',
                          `Source: ${schedule.source === 'override' ? 'Override' : schedule.source === 'default' ? 'Default Schedule' : 'None'}`,
                        ];
                        if (schedule.expectedRevenue && schedule.expectedRevenue > 0) {
                          tooltipParts.push(`Expected Revenue: ${schedule.expectedRevenue.toLocaleString()}`);
                        }
                        const tooltip = tooltipParts.join(' • ');

                        return (
                          <td
                            key={day.toISOString()}
                            className={`
                              border ${borderColor} p-2 min-h-[80px] transition-colors
                              ${bgColor} ${textColor}
                              ${isToday ? 'ring-2 ring-blue-500' : ''}
                            `}
                            title={tooltip}
                          >
                            <div className="text-xs mb-1 font-medium text-neutral-text-secondary">
                              {format(day, 'd')}
                            </div>
                            {!isInCurrentMonth ? (
                              <div className="text-xs text-gray-400 text-center">-</div>
                            ) : schedule.isOff ? (
                              <div className="text-xs text-gray-500 text-center">Off</div>
                            ) : schedule.branch ? (
                              <div className="text-xs">
                                <div className="flex items-center gap-1 flex-wrap">
                                  <span className="font-semibold">{schedule.branch.code || 'N/A'}</span>
                                  {sourceBadge && (
                                    <span className={`text-[9px] px-1 py-0.5 rounded ${
                                      schedule.source === 'override' 
                                        ? 'bg-orange-200 text-orange-800 font-bold' 
                                        : 'bg-blue-200 text-blue-800'
                                    }`}>
                                      {sourceBadge}
                                    </span>
                                  )}
                                </div>
                                <div className="truncate mt-0.5" title={schedule.branch.name}>
                                  {schedule.branch.name || 'Unknown'}
                                </div>
                                {schedule.expectedRevenue && schedule.expectedRevenue > 0 && (
                                  <div className="text-[10px] mt-1 opacity-75">
                                    {schedule.expectedRevenue.toLocaleString()}
                                  </div>
                                )}
                              </div>
                            ) : (
                              <div className="text-xs text-gray-500 text-center">-</div>
                            )}
                          </td>
                        );
                      })}
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {/* Legend and Summary */}
          <div className="mt-6 space-y-3">
            <div className="flex items-center gap-4 text-sm text-neutral-text-secondary flex-wrap">
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-orange-50 border-2 border-orange-300 rounded"></div>
                <span>Override - Working</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-blue-50 border-2 border-blue-300 rounded"></div>
                <span>Default Schedule</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-gray-100 border-2 border-gray-300 rounded"></div>
                <span>Off Day</span>
              </div>
            </div>
            <div className="text-xs text-neutral-text-secondary">
              <strong>Summary:</strong> {summaryStats.workingDays} working day(s) ({summaryStats.defaultDays} default, {summaryStats.overrideDays} override), {summaryStats.offDays} off day(s)
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
