'use client';

import { useState, useEffect } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameDay, addMonths, subMonths } from 'date-fns';
import { scheduleApi, StaffSchedule } from '@/lib/api/schedule';
import { staffApi, Staff } from '@/lib/api/staff';

interface MonthlyCalendarProps {
  branchId: string;
}

export default function MonthlyCalendar({ branchId }: MonthlyCalendarProps) {
  const [currentDate, setCurrentDate] = useState(new Date());
  const [schedules, setSchedules] = useState<StaffSchedule[]>([]);
  const [staff, setStaff] = useState<Staff[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedCell, setSelectedCell] = useState<{ staffId: string; date: Date } | null>(null);

  const year = currentDate.getFullYear();
  const month = currentDate.getMonth() + 1;

  useEffect(() => {
    loadData();
  }, [branchId, year, month]);

  const loadData = async () => {
    try {
      setLoading(true);
      const [schedulesData, staffData] = await Promise.all([
        scheduleApi.getMonthlyView(branchId, year, month),
        staffApi.list({ staff_type: 'branch', branch_id: branchId }),
      ]);
      setSchedules(schedulesData || []);
      setStaff(staffData || []);
    } catch (error) {
      console.error('Failed to load data:', error);
      setSchedules([]);
      setStaff([]);
    } finally {
      setLoading(false);
    }
  };

  const monthStart = startOfMonth(currentDate);
  const monthEnd = endOfMonth(currentDate);
  const daysInMonth = eachDayOfInterval({ start: monthStart, end: monthEnd });

  const handleCellClick = async (staffId: string, date: Date) => {
    const dateStr = format(date, 'yyyy-MM-dd');
    const existingSchedule = schedules.find(
      (s) => s.staff_id === staffId && isSameDay(new Date(s.date), date)
    );

    try {
      if (existingSchedule) {
        // Toggle working day status
        await scheduleApi.create({
          staff_id: staffId,
          branch_id: branchId,
          date: dateStr,
          is_working_day: !existingSchedule.is_working_day,
        });
      } else {
        // Create new schedule (default to working day)
        await scheduleApi.create({
          staff_id: staffId,
          branch_id: branchId,
          date: dateStr,
          is_working_day: true,
        });
      }
      await loadData();
    } catch (error) {
      console.error('Failed to update schedule:', error);
    }
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

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  return (
    <div className="w-full p-6">
      <div className="mb-6 flex items-center justify-between">
        <h2 className="text-xl font-semibold text-neutral-text-primary">
          Staff Scheduling - {format(currentDate, 'MMMM yyyy')}
        </h2>
        <div className="flex gap-2">
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
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full border-collapse border border-neutral-border">
          <thead>
            <tr>
              <th className="border border-neutral-border p-2 bg-neutral-hover font-semibold sticky left-0 z-10 text-neutral-text-primary">
                Staff Name
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
            {staff.map((staffMember) => (
              <tr key={staffMember.id}>
                <td className="border border-neutral-border p-2 bg-neutral-bg-secondary sticky left-0 z-10 font-medium text-neutral-text-primary">
                  {staffMember.name}
                </td>
                {daysInMonth.map((day) => {
                  const schedule = getScheduleForCell(staffMember.id, day);
                  const isWorkingDay = schedule?.is_working_day ?? false;
                  const isToday = isSameDay(day, new Date());
                  
                  return (
                    <td
                      key={day.toISOString()}
                      onClick={() => handleCellClick(staffMember.id, day)}
                      className={`
                        border border-neutral-border p-1 cursor-pointer transition-colors
                        ${isWorkingDay ? 'bg-green-50 hover:bg-green-100' : 'bg-neutral-bg-secondary hover:bg-neutral-hover'}
                        ${isToday ? 'ring-2 ring-salesforce-blue' : ''}
                      `}
                      title={`${format(day, 'MMM d, yyyy')} - Click to toggle`}
                    >
                      <div className="w-6 h-6 flex items-center justify-center text-xs">
                        {isWorkingDay ? '✓' : ''}
                      </div>
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="mt-4 flex items-center gap-4 text-sm text-neutral-text-secondary">
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-green-50 border border-neutral-border"></div>
          <span>Working Day</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-neutral-bg-secondary border border-neutral-border"></div>
          <span>Off Day</span>
        </div>
        <div>Click on a cell to toggle working/off day</div>
      </div>
    </div>
  );
}

