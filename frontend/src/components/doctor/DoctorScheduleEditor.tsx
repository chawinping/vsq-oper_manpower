'use client';

import { useState, useEffect } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameDay } from 'date-fns';
import { doctorApi, DoctorOnOffDay } from '@/lib/api/doctor';

interface DoctorScheduleEditorProps {
  branchId: string;
  year?: number;
  month?: number;
}

export default function DoctorScheduleEditor({ branchId, year, month }: DoctorScheduleEditorProps) {
  const [days, setDays] = useState<DoctorOnOffDay[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState<string | null>(null);

  useEffect(() => {
    loadDays();
  }, [branchId, year, month]);

  const loadDays = async () => {
    try {
      setLoading(true);
      const now = new Date();
      const startDate = new Date(year || now.getFullYear(), (month || now.getMonth() + 1) - 1, 1);
      const endDate = endOfMonth(startDate);

      const data = await doctorApi.getDoctorOnOffDays({
        branch_id: branchId,
        start_date: format(startDate, 'yyyy-MM-dd'),
        end_date: format(endDate, 'yyyy-MM-dd'),
      });
      setDays(data || []);
    } catch (error) {
      console.error('Failed to load doctor on/off days:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleToggle = async (date: Date, currentValue: boolean) => {
    const dateStr = format(date, 'yyyy-MM-dd');
    setSaving(dateStr);

    try {
      await doctorApi.createDoctorOnOffDay({
        branch_id: branchId,
        date: dateStr,
        is_doctor_on: !currentValue,
      });
      await loadDays();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to update doctor schedule');
    } finally {
      setSaving(null);
    }
  };

  const getDayStatus = (date: Date): boolean | null => {
    const day = days.find(d => isSameDay(new Date(d.date), date));
    return day ? day.is_doctor_on : null;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  const now = new Date();
  const monthStart = startOfMonth(new Date(year || now.getFullYear(), (month || now.getMonth() + 1) - 1, 1));
  const monthEnd = endOfMonth(monthStart);
  const daysInMonth = eachDayOfInterval({ start: monthStart, end: monthEnd });

  return (
    <div className="w-full">
      <h2 className="text-2xl font-bold mb-4">
        Doctor On/Off Days - {format(monthStart, 'MMMM yyyy')}
      </h2>

      <div className="grid grid-cols-7 gap-2">
        {daysInMonth.map((day) => {
          const status = getDayStatus(day);
          const isSaving = saving === format(day, 'yyyy-MM-dd');

          return (
            <div
              key={day.toISOString()}
              className={`p-3 border rounded-lg cursor-pointer transition-colors ${
                status === true
                  ? 'bg-green-100 border-green-300'
                  : status === false
                  ? 'bg-red-100 border-red-300'
                  : 'bg-gray-50 border-gray-200'
              } ${isSaving ? 'opacity-50' : 'hover:opacity-80'}`}
              onClick={() => !isSaving && handleToggle(day, status === true)}
            >
              <div className="text-xs font-medium mb-1">
                {format(day, 'EEE')}
              </div>
              <div className="text-lg font-bold">
                {format(day, 'd')}
              </div>
              <div className="text-xs mt-1">
                {status === true ? 'Doctor On' : status === false ? 'Doctor Off' : 'Not Set'}
              </div>
            </div>
          );
        })}
      </div>

      <div className="mt-4 flex gap-4 text-sm">
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-green-100 border border-green-300 rounded"></div>
          <span>Doctor On</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-red-100 border border-red-300 rounded"></div>
          <span>Doctor Off</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-gray-50 border border-gray-200 rounded"></div>
          <span>Not Set</span>
        </div>
      </div>
    </div>
  );
}
