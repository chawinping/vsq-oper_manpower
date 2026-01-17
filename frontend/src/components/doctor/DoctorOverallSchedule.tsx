'use client';

import { useState, useEffect } from 'react';
import { format, startOfWeek, endOfWeek, eachDayOfInterval, getDay, parseISO } from 'date-fns';
import { doctorApi, Doctor, DoctorAssignment, DoctorDefaultSchedule } from '@/lib/api/doctor';
import { branchApi, Branch } from '@/lib/api/branch';

interface DoctorOverallScheduleProps {
  doctors: Doctor[];
  branches: Branch[];
}

type ViewMode = 'default' | 'actual' | 'week';

export default function DoctorOverallSchedule({
  doctors,
  branches,
}: DoctorOverallScheduleProps) {
  const [viewMode, setViewMode] = useState<ViewMode>('default');
  const [selectedWeek, setSelectedWeek] = useState<Date>(new Date());
  const [loading, setLoading] = useState(true);
  
  // Data for different views
  const [defaultSchedules, setDefaultSchedules] = useState<DoctorDefaultSchedule[]>([]);
  const [actualAssignments, setActualAssignments] = useState<DoctorAssignment[]>([]);
  const [weekAssignments, setWeekAssignments] = useState<DoctorAssignment[]>([]);

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
      } else if (viewMode === 'actual') {
        // Load all actual assignments (no date filter to get all)
        const assignments = await doctorApi.getAssignments();
        setActualAssignments(assignments);
      } else if (viewMode === 'week') {
        // Load assignments for the selected week
        const weekStart = startOfWeek(selectedWeek, { weekStartsOn: 1 }); // Monday
        const weekEnd = endOfWeek(selectedWeek, { weekStartsOn: 1 }); // Sunday
        const assignments = await doctorApi.getAssignments({
          start_date: format(weekStart, 'yyyy-MM-dd'),
          end_date: format(weekEnd, 'yyyy-MM-dd'),
        });
        setWeekAssignments(assignments);
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

  // Get doctors for a branch and day (actual assignments view)
  const getDoctorsForBranchAndDayActual = (branchId: string, dayOfWeek: number): Doctor[] => {
    const doctorsForDay: Doctor[] = [];
    
    actualAssignments.forEach(assignment => {
      if (assignment.branch_id === branchId) {
        const assignmentDate = parseISO(assignment.date);
        const assignmentDayOfWeek = getDay(assignmentDate); // getDay: 0=Sunday, 1=Monday, ..., 6=Saturday
        // Convert to our format: 1=Monday, 2=Tuesday, ..., 6=Saturday, 7=Sunday
        const displayDayOfWeek = assignmentDayOfWeek === 0 ? 7 : assignmentDayOfWeek;
        
        if (displayDayOfWeek === dayOfWeek) {
          const doctor = doctorMap.get(assignment.doctor_id);
          if (doctor && !doctorsForDay.find(d => d.id === doctor.id)) {
            doctorsForDay.push(doctor);
          }
        }
      }
    });
    
    return doctorsForDay;
  };

  // Get doctors for a branch and day (week view)
  const getDoctorsForBranchAndDayWeek = (branchId: string, dayOfWeek: number): Doctor[] => {
    const doctorsForDay: Doctor[] = [];
    const weekStart = startOfWeek(selectedWeek, { weekStartsOn: 1 });
    const weekDays = eachDayOfInterval({ start: weekStart, end: endOfWeek(selectedWeek, { weekStartsOn: 1 }) });
    // dayOfWeek: 1=Monday, 2=Tuesday, ..., 6=Saturday, 7=Sunday
    // weekDays array: [Monday, Tuesday, ..., Sunday] (index 0-6)
    const targetDate = weekDays[dayOfWeek === 7 ? 6 : dayOfWeek - 1];
    
    weekAssignments.forEach(assignment => {
      if (assignment.branch_id === branchId) {
        const assignmentDate = parseISO(assignment.date);
        if (format(assignmentDate, 'yyyy-MM-dd') === format(targetDate, 'yyyy-MM-dd')) {
          const doctor = doctorMap.get(assignment.doctor_id);
          if (doctor && !doctorsForDay.find(d => d.id === doctor.id)) {
            doctorsForDay.push(doctor);
          }
        }
      }
    });
    
    return doctorsForDay;
  };

  // Get doctors based on view mode
  const getDoctorsForBranchAndDay = (branchId: string, dayOfWeek: number): Doctor[] => {
    if (viewMode === 'default') {
      return getDoctorsForBranchAndDayDefault(branchId, dayOfWeek);
    } else if (viewMode === 'actual') {
      return getDoctorsForBranchAndDayActual(branchId, dayOfWeek);
    } else {
      return getDoctorsForBranchAndDayWeek(branchId, dayOfWeek);
    }
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
                  <option value="actual">Actual Assignments</option>
                  <option value="week">Specific Week</option>
                </select>
              </div>

              {/* Week Selector (only show for week view) */}
              {viewMode === 'week' && (
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
            {viewMode === 'actual' && 'Showing all actual assignments (includes overrides) for all dates'}
            {viewMode === 'week' && `Showing assignments for week of ${format(weekStart, 'MMM d')} - ${format(weekEnd, 'MMM d, yyyy')}`}
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
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
