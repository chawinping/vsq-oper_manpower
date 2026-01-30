'use client';

import { useState, useEffect } from 'react';
import { format, startOfWeek, endOfWeek, eachDayOfInterval, getDay, parseISO, isSameDay } from 'date-fns';
import { doctorApi, Doctor, DoctorAssignment, DoctorDefaultSchedule, DoctorScheduleOverride } from '@/lib/api/doctor';
import { branchApi, Branch } from '@/lib/api/branch';

interface DoctorOverallScheduleProps {
  doctors: Doctor[];
  branches: Branch[];
}

type ViewMode = 'default' | 'weekly';

interface DoctorScheduleInfo {
  doctor: Doctor;
  branch: Branch | null;
  isOff: boolean;
  source: 'override' | 'default' | 'none';
  expectedRevenue?: number;
}

export default function DoctorOverallSchedule({
  doctors,
  branches,
}: DoctorOverallScheduleProps) {
  const [viewMode, setViewMode] = useState<ViewMode>('default');
  const [selectedWeek, setSelectedWeek] = useState<Date>(new Date());
  const [loading, setLoading] = useState(true);
  
  // Data for different views
  const [defaultSchedules, setDefaultSchedules] = useState<DoctorDefaultSchedule[]>([]);
  // Weekly view data: maps doctor_id to their schedules
  const [weeklyDefaultSchedules, setWeeklyDefaultSchedules] = useState<Map<string, DoctorDefaultSchedule[]>>(new Map());
  const [weeklyOverrides, setWeeklyOverrides] = useState<Map<string, DoctorScheduleOverride[]>>(new Map());
  const [weeklyAssignments, setWeeklyAssignments] = useState<Map<string, DoctorAssignment[]>>(new Map());

  // Map doctor_id to doctor name
  const doctorMap = new Map(doctors.map(d => [d.id, d]));
  
  // Days of week: Monday = 1, Tuesday = 2, ..., Sunday = 7
  const daysOfWeek = [
    { label: 'Monday', dayOfWeek: 1 },
    { label: 'Tuesday', dayOfWeek: 2 },
    { label: 'Wednesday', dayOfWeek: 3 },
    { label: 'Thursday', dayOfWeek: 4 },
    { label: 'Friday', dayOfWeek: 5 },
    { label: 'Saturday', dayOfWeek: 6 },
    { label: 'Sunday', dayOfWeek: 7 },
  ];

  useEffect(() => {
    loadData();
  }, [viewMode, selectedWeek]);

  const loadData = async () => {
    setLoading(true);
    try {
      if (viewMode === 'default') {
        // Load all default schedules for all doctors
        const allDefaultSchedules: DoctorDefaultSchedule[] = [];
        for (const doctor of doctors) {
          try {
            const schedules = await doctorApi.getDefaultSchedules(doctor.id);
            allDefaultSchedules.push(...schedules);
          } catch (error) {
            console.error(`Failed to load default schedules for doctor ${doctor.id}:`, error);
          }
        }
        setDefaultSchedules(allDefaultSchedules);
      } else if (viewMode === 'weekly') {
        // Load weekly view data: default schedules, overrides, and assignments for all doctors
        const weekStart = startOfWeek(selectedWeek, { weekStartsOn: 1 }); // Monday
        const weekEnd = endOfWeek(selectedWeek, { weekStartsOn: 1 }); // Sunday
        
        const defaultSchedulesMap = new Map<string, DoctorDefaultSchedule[]>();
        const overridesMap = new Map<string, DoctorScheduleOverride[]>();
        const assignmentsMap = new Map<string, DoctorAssignment[]>();
        
        // Load data for all doctors in parallel
        await Promise.all(
          doctors.map(async (doctor) => {
            try {
              const [defaultSchedulesData, overridesData, assignmentsData] = await Promise.all([
                doctorApi.getDefaultSchedules(doctor.id),
                doctorApi.getScheduleOverrides({
                  doctor_id: doctor.id,
                  start_date: format(weekStart, 'yyyy-MM-dd'),
                  end_date: format(weekEnd, 'yyyy-MM-dd'),
                }),
                doctorApi.getAssignments({
                  doctor_id: doctor.id,
                  start_date: format(weekStart, 'yyyy-MM-dd'),
                  end_date: format(weekEnd, 'yyyy-MM-dd'),
                }),
              ]);
              
              defaultSchedulesMap.set(doctor.id, defaultSchedulesData || []);
              overridesMap.set(doctor.id, overridesData || []);
              assignmentsMap.set(doctor.id, assignmentsData || []);
            } catch (error) {
              console.error(`Failed to load weekly data for doctor ${doctor.id}:`, error);
              // Set empty arrays on error
              defaultSchedulesMap.set(doctor.id, []);
              overridesMap.set(doctor.id, []);
              assignmentsMap.set(doctor.id, []);
            }
          })
        );
        
        setWeeklyDefaultSchedules(defaultSchedulesMap);
        setWeeklyOverrides(overridesMap);
        setWeeklyAssignments(assignmentsMap);
      }
    } catch (error) {
      console.error('Failed to load schedule data:', error);
    } finally {
      setLoading(false);
    }
  };

  // Convert day of week: API uses 0=Sunday, 1=Monday, ..., 6=Saturday
  // Our display uses: 1=Monday, 2=Tuesday, ..., 6=Saturday, 7=Sunday
  const convertDayOfWeek = (apiDay: number): number => {
    // API: 0=Sunday, 1=Monday, ..., 6=Saturday
    // Our display: 1=Monday, 2=Tuesday, ..., 6=Saturday, 7=Sunday
    if (apiDay === 0) return 7; // Sunday becomes 7
    return apiDay; // Monday=1, ..., Saturday=6
  };

  // Get doctors for a branch and day (default schedule view)
  const getDoctorsForBranchAndDayDefault = (branchId: string, dayOfWeek: number): Doctor[] => {
    const doctorsForDay: Doctor[] = [];
    
    // Find all default schedules for this branch and day
    const schedules = defaultSchedules.filter(s => {
      const scheduleDay = convertDayOfWeek(s.day_of_week);
      return s.branch_id === branchId && scheduleDay === dayOfWeek;
    });
    
    schedules.forEach(schedule => {
      const doctor = doctorMap.get(schedule.doctor_id);
      if (doctor) {
        doctorsForDay.push(doctor);
      }
    });
    
    return doctorsForDay;
  };

  // Get schedule info for a specific doctor and date (similar to DoctorScheduleCalendar logic)
  const getDoctorScheduleForDate = (doctorId: string, date: Date): DoctorScheduleInfo => {
    const doctor = doctorMap.get(doctorId);
    if (!doctor) {
      return {
        doctor: {} as Doctor,
        branch: null,
        isOff: true,
        source: 'none',
      };
    }
    
    const overrides = weeklyOverrides.get(doctorId) || [];
    const defaultSchedules = weeklyDefaultSchedules.get(doctorId) || [];
    const assignments = weeklyAssignments.get(doctorId) || [];
    
    // Priority 1: Check for override
    const override = overrides.find(o => isSameDay(new Date(o.date), date));
    if (override) {
      if (override.type === 'off') {
        return {
          doctor,
          branch: null,
          isOff: true,
          source: 'override',
        };
      } else {
        // Working override
        const branch = override.branch_id ? branches.find(b => b.id === override.branch_id) || null : null;
        // Check for expected revenue from assignment
        const assignment = assignments.find(a => isSameDay(new Date(a.date), date));
        return {
          doctor,
          branch,
          isOff: false,
          source: 'override',
          expectedRevenue: assignment?.expected_revenue,
        };
      }
    }
    
    // Priority 2: Check default schedule for day of week
    const dayOfWeek = getDay(date); // 0=Sunday, 1=Monday, ..., 6=Saturday
    const defaultSchedule = defaultSchedules.find(s => s.day_of_week === dayOfWeek);
    if (defaultSchedule) {
      const branch = branches.find(b => b.id === defaultSchedule.branch_id) || null;
      // Check for expected revenue from assignment
      const assignment = assignments.find(a => isSameDay(new Date(a.date), date));
      return {
        doctor,
        branch,
        isOff: false,
        source: 'default',
        expectedRevenue: assignment?.expected_revenue,
      };
    }
    
    // Priority 3: No schedule (off day)
    return {
      doctor,
      branch: null,
      isOff: true,
      source: 'none',
    };
  };

  // Get doctors with schedule info for a branch and day (weekly view)
  const getDoctorsForBranchAndDayWeekly = (branchId: string, date: Date): DoctorScheduleInfo[] => {
    const doctorsForDay: DoctorScheduleInfo[] = [];
    
    doctors.forEach(doctor => {
      const scheduleInfo = getDoctorScheduleForDate(doctor.id, date);
      // Only include if doctor is assigned to this branch and not off
      if (!scheduleInfo.isOff && scheduleInfo.branch?.id === branchId) {
        doctorsForDay.push(scheduleInfo);
      }
    });
    
    return doctorsForDay;
  };

  // Get doctors based on view mode
  const getDoctorsForBranchAndDay = (branchId: string, dayOfWeek: number): Doctor[] => {
    if (viewMode === 'default') {
      return getDoctorsForBranchAndDayDefault(branchId, dayOfWeek);
    } else {
      // Weekly view - need to get the actual date for this day of week
      const weekStart = startOfWeek(selectedWeek, { weekStartsOn: 1 });
      const weekDays = eachDayOfInterval({ start: weekStart, end: endOfWeek(selectedWeek, { weekStartsOn: 1 }) });
      // dayOfWeek: 1=Monday, 2=Tuesday, ..., 6=Saturday, 7=Sunday
      // weekDays array: [Monday, Tuesday, ..., Sunday] (index 0-6)
      const targetDate = weekDays[dayOfWeek === 7 ? 6 : dayOfWeek - 1];
      const scheduleInfos = getDoctorsForBranchAndDayWeekly(branchId, targetDate);
      return scheduleInfos.map(info => info.doctor);
    }
  };

  // Get schedule info for a branch and day (weekly view only)
  const getScheduleInfosForBranchAndDay = (branchId: string, dayOfWeek: number): DoctorScheduleInfo[] => {
    if (viewMode !== 'weekly') return [];
    
    const weekStart = startOfWeek(selectedWeek, { weekStartsOn: 1 });
    const weekDays = eachDayOfInterval({ start: weekStart, end: endOfWeek(selectedWeek, { weekStartsOn: 1 }) });
    const targetDate = weekDays[dayOfWeek === 7 ? 6 : dayOfWeek - 1];
    return getDoctorsForBranchAndDayWeekly(branchId, targetDate);
  };

  // Format doctor names with slash separator
  const formatDoctorNames = (doctors: Doctor[]): string => {
    if (doctors.length === 0) return '-';
    return doctors.map(d => d.name).join(' / ');
  };

  const weekStart = startOfWeek(selectedWeek, { weekStartsOn: 1 });
  const weekEnd = endOfWeek(selectedWeek, { weekStartsOn: 1 });

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading schedule data...</div>
      </div>
    );
  }

  return (
    <div className="w-full">
      <div className="card mb-6">
        <div className="p-4">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold text-neutral-text-primary">
              Overall Doctor Schedule
            </h2>
            <div className="flex items-center gap-4">
              {/* View Mode Toggle */}
              <div className="flex items-center gap-2">
                <label className="text-sm font-medium text-neutral-text-primary">View:</label>
                <select
                  value={viewMode}
                  onChange={(e) => setViewMode(e.target.value as ViewMode)}
                  className="input-field text-sm"
                >
                  <option value="default">Default Schedule</option>
                  <option value="weekly">Weekly Schedule</option>
                </select>
              </div>

              {/* Week Selector (only show for weekly view) */}
              {viewMode === 'weekly' && (
                <div className="flex items-center gap-2">
                  <label className="text-sm font-medium text-neutral-text-primary">Week:</label>
                  <input
                    type="date"
                    value={format(weekStart, 'yyyy-MM-dd')}
                    onChange={(e) => {
                      if (e.target.value) {
                        const date = parseISO(e.target.value);
                        setSelectedWeek(date);
                      }
                    }}
                    className="input-field text-sm"
                  />
                  <span className="text-sm text-neutral-text-secondary">
                    {format(weekStart, 'MMM d')} - {format(weekEnd, 'MMM d, yyyy')}
                  </span>
                </div>
              )}
            </div>
          </div>

          <div className="text-sm text-neutral-text-secondary mb-4">
            {viewMode === 'default' && 'Showing weekly default schedule pattern for all doctors'}
            {viewMode === 'weekly' && `Showing combined schedule (default + overrides + assignments) for week of ${format(weekStart, 'MMM d')} - ${format(weekEnd, 'MMM d, yyyy')}`}
          </div>
        </div>
      </div>

      <div className="card">
        <div className="p-6">
          <div className="overflow-x-auto">
            <table className="w-full border-collapse border border-neutral-border">
              <thead>
                <tr>
                  <th className="border border-neutral-border p-3 bg-neutral-hover font-semibold sticky left-0 z-10 text-neutral-text-primary min-w-[200px] text-left">
                    Branch
                  </th>
                  {daysOfWeek.map((day) => (
                    <th
                      key={day.dayOfWeek}
                      className="border border-neutral-border p-3 bg-neutral-hover font-semibold min-w-[150px] text-neutral-text-primary"
                    >
                      {day.label}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {branches.map((branch) => (
                  <tr key={branch.id}>
                    <td className="border border-neutral-border p-3 bg-neutral-bg-secondary sticky left-0 z-10 font-medium text-neutral-text-primary min-w-[200px]">
                      <div>
                        <div className="font-semibold">{branch.code}</div>
                        <div className="text-xs text-neutral-text-secondary">{branch.name}</div>
                      </div>
                    </td>
                    {daysOfWeek.map((day) => {
                      if (viewMode === 'default') {
                        const doctorsForDay = getDoctorsForBranchAndDay(branch.id, day.dayOfWeek);
                        const doctorNames = formatDoctorNames(doctorsForDay);
                        
                        return (
                          <td
                            key={`${branch.id}-${day.dayOfWeek}`}
                            className="border border-neutral-border p-3 bg-neutral-bg-secondary min-h-[60px] text-sm text-neutral-text-primary"
                          >
                            {doctorNames !== '-' ? (
                              <div className="text-xs">
                                {doctorsForDay.map((doctor, idx) => (
                                  <span key={doctor.id}>
                                    {doctor.name}
                                    {idx < doctorsForDay.length - 1 && ' / '}
                                  </span>
                                ))}
                              </div>
                            ) : (
                              <div className="text-xs text-neutral-text-secondary">-</div>
                            )}
                          </td>
                        );
                      } else {
                        // Weekly view - show calendar-style display
                        const scheduleInfos = getScheduleInfosForBranchAndDay(branch.id, day.dayOfWeek);
                        const weekStart = startOfWeek(selectedWeek, { weekStartsOn: 1 });
                        const weekDays = eachDayOfInterval({ start: weekStart, end: endOfWeek(selectedWeek, { weekStartsOn: 1 }) });
                        const targetDate = weekDays[day.dayOfWeek === 7 ? 6 : day.dayOfWeek - 1];
                        const isToday = isSameDay(targetDate, new Date());
                        
                        // Determine if there are any overrides or defaults
                        const hasOverrides = scheduleInfos.some(info => info.source === 'override');
                        const hasDefaults = scheduleInfos.some(info => info.source === 'default');
                        
                        // Set background color based on content
                        let bgColor = 'bg-neutral-bg-secondary';
                        if (hasOverrides) {
                          bgColor = 'bg-orange-50';
                        } else if (hasDefaults) {
                          bgColor = 'bg-blue-50';
                        }
                        
                        return (
                          <td
                            key={`${branch.id}-${day.dayOfWeek}`}
                            className={`
                              border border-neutral-border p-2 min-h-[80px] transition-colors
                              ${bgColor}
                              ${isToday ? 'ring-2 ring-blue-500' : ''}
                            `}
                            title={`${format(targetDate, 'MMM d, yyyy')} - ${scheduleInfos.length} doctor(s)`}
                          >
                            <div className="text-xs mb-1 font-medium text-neutral-text-secondary">
                              {format(targetDate, 'd')}
                            </div>
                            {scheduleInfos.length === 0 ? (
                              <div className="text-xs text-gray-500 text-center">-</div>
                            ) : (
                              <div className="space-y-1">
                                {scheduleInfos.map((info) => (
                                  <div key={info.doctor.id} className="text-xs">
                                    <div className="flex items-center gap-1 flex-wrap">
                                      <span className="font-semibold">{info.doctor.name}</span>
                                      {info.source === 'override' && (
                                        <span className="text-[9px] px-1 py-0.5 rounded bg-orange-200 text-orange-800 font-bold">
                                          O
                                        </span>
                                      )}
                                      {info.source === 'default' && (
                                        <span className="text-[9px] px-1 py-0.5 rounded bg-blue-200 text-blue-800">
                                          D
                                        </span>
                                      )}
                                    </div>
                                    {info.branch && (
                                      <div className="text-[10px] text-neutral-text-secondary truncate" title={info.branch.name}>
                                        {info.branch.code || 'N/A'}
                                      </div>
                                    )}
                                    {info.expectedRevenue && info.expectedRevenue > 0 && (
                                      <div className="text-[10px] mt-0.5 opacity-75">
                                        {info.expectedRevenue.toLocaleString()}
                                      </div>
                                    )}
                                  </div>
                                ))}
                              </div>
                            )}
                          </td>
                        );
                      }
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          
          {/* Legend for weekly view */}
          {viewMode === 'weekly' && (
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
                  <div className="w-4 h-4 bg-neutral-bg-secondary border-2 border-neutral-border rounded"></div>
                  <span>No Assignment</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-[9px] px-1 py-0.5 rounded bg-orange-200 text-orange-800 font-bold">O</span>
                  <span>Override Badge</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-[9px] px-1 py-0.5 rounded bg-blue-200 text-blue-800">D</span>
                  <span>Default Badge</span>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
