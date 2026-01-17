'use client';

import { useState, useEffect } from 'react';
import { doctorApi, Doctor, DoctorDefaultSchedule } from '@/lib/api/doctor';
import { branchApi, Branch } from '@/lib/api/branch';

interface DoctorDefaultScheduleManagerProps {
  doctor: Doctor;
}

const DAYS_OF_WEEK = [
  { value: 0, label: 'Sunday' },
  { value: 1, label: 'Monday' },
  { value: 2, label: 'Tuesday' },
  { value: 3, label: 'Wednesday' },
  { value: 4, label: 'Thursday' },
  { value: 5, label: 'Friday' },
  { value: 6, label: 'Saturday' },
];

export default function DoctorDefaultScheduleManager({ doctor }: DoctorDefaultScheduleManagerProps) {
  const [schedules, setSchedules] = useState<DoctorDefaultSchedule[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState<number | null>(null);

  useEffect(() => {
    loadData();
  }, [doctor.id]);

  const loadData = async () => {
    try {
      setLoading(true);
      const [schedulesData, branchesData] = await Promise.all([
        doctorApi.getDefaultSchedules(doctor.id),
        branchApi.list(),
      ]);
      setSchedules(schedulesData);
      setBranches(branchesData);
    } catch (error) {
      console.error('Failed to load data:', error);
      alert('Failed to load default schedules');
    } finally {
      setLoading(false);
    }
  };

  const getScheduleForDay = (dayOfWeek: number): DoctorDefaultSchedule | undefined => {
    return schedules.find(s => s.day_of_week === dayOfWeek);
  };

  const handleBranchChange = async (dayOfWeek: number, branchId: string) => {
    if (!branchId) {
      // Delete schedule if branch is cleared
      const existing = getScheduleForDay(dayOfWeek);
      if (existing) {
        try {
          setSaving(dayOfWeek);
          await doctorApi.deleteDefaultSchedule(existing.id);
          await loadData();
        } catch (error: any) {
          alert(error.response?.data?.error || 'Failed to delete schedule');
        } finally {
          setSaving(null);
        }
      }
      return;
    }

    try {
      setSaving(dayOfWeek);
      const existing = getScheduleForDay(dayOfWeek);
      
      if (existing) {
        await doctorApi.updateDefaultSchedule(existing.id, { branch_id: branchId });
      } else {
        await doctorApi.createDefaultSchedule({
          doctor_id: doctor.id,
          day_of_week: dayOfWeek,
          branch_id: branchId,
        });
      }
      await loadData();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save schedule');
    } finally {
      setSaving(null);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  return (
    <div className="card">
      <div className="p-6">
        <h3 className="text-lg font-semibold text-neutral-text-primary mb-4">
          Default Weekly Schedule - {doctor.name} {doctor.code ? `(${doctor.code})` : ''}
        </h3>
        <p className="text-sm text-neutral-text-secondary mb-4">
          Set the default branch for each day of the week. Select "Off Day" if the doctor is off on that day. This schedule will be used unless overridden.
        </p>

        <div className="space-y-3">
          {DAYS_OF_WEEK.map((day) => {
            const schedule = getScheduleForDay(day.value);
            const selectedBranchId = schedule?.branch_id || '';
            const isOffDay = !schedule; // No schedule means off day
            const isSaving = saving === day.value;

            return (
              <div key={day.value} className={`flex items-center gap-4 p-3 border rounded-lg ${
                isOffDay ? 'border-red-300 bg-red-50' : 'border-neutral-border'
              }`}>
                <div className="w-32 font-medium text-neutral-text-primary">
                  {day.label}
                </div>
                <div className="flex-1">
                  <select
                    value={isOffDay ? 'OFF_DAY' : selectedBranchId}
                    onChange={(e) => {
                      const value = e.target.value;
                      if (value === 'OFF_DAY') {
                        handleBranchChange(day.value, '');
                      } else {
                        handleBranchChange(day.value, value);
                      }
                    }}
                    disabled={isSaving}
                    className={`input-field ${isSaving ? 'opacity-50' : ''}`}
                  >
                    <option value="OFF_DAY">-- Off Day --</option>
                    {branches.map((branch) => (
                      <option key={branch.id} value={branch.id}>
                        {branch.code} - {branch.name}
                      </option>
                    ))}
                  </select>
                </div>
                {isOffDay && (
                  <div className="text-sm text-red-600 font-medium">
                    Off Day
                  </div>
                )}
                {schedule && !isOffDay && (
                  <div className="text-sm text-neutral-text-secondary">
                    {schedule.branch?.code || 'N/A'}
                  </div>
                )}
                {isSaving && (
                  <div className="text-sm text-neutral-text-secondary">Saving...</div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
