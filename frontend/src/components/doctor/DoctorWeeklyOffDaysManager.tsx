'use client';

import { useState, useEffect } from 'react';
import { doctorApi, Doctor, DoctorWeeklyOffDay } from '@/lib/api/doctor';

interface DoctorWeeklyOffDaysManagerProps {
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

export default function DoctorWeeklyOffDaysManager({ doctor }: DoctorWeeklyOffDaysManagerProps) {
  const [offDays, setOffDays] = useState<DoctorWeeklyOffDay[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState<number | null>(null);

  useEffect(() => {
    loadOffDays();
  }, [doctor.id]);

  const loadOffDays = async () => {
    try {
      setLoading(true);
      const data = await doctorApi.getWeeklyOffDays(doctor.id);
      setOffDays(data);
    } catch (error) {
      console.error('Failed to load weekly off days:', error);
      alert('Failed to load weekly off days');
    } finally {
      setLoading(false);
    }
  };

  const isOffDay = (dayOfWeek: number): boolean => {
    return offDays.some(od => od.day_of_week === dayOfWeek);
  };

  const handleToggleOffDay = async (dayOfWeek: number) => {
    const isCurrentlyOff = isOffDay(dayOfWeek);
    
    try {
      setSaving(dayOfWeek);
      
      if (isCurrentlyOff) {
        // Remove off day
        const offDay = offDays.find(od => od.day_of_week === dayOfWeek);
        if (offDay) {
          await doctorApi.deleteWeeklyOffDay(offDay.id);
        }
      } else {
        // Add off day
        await doctorApi.createWeeklyOffDay({
          doctor_id: doctor.id,
          day_of_week: dayOfWeek,
        });
      }
      
      await loadOffDays();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to update weekly off day');
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
          Default Weekly Off Days - {doctor.name} {doctor.code ? `(${doctor.code})` : ''}
        </h3>
        <p className="text-sm text-neutral-text-secondary mb-4">
          Select the days of the week when this doctor is typically off. These days will be used as defaults unless overridden.
        </p>

        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {DAYS_OF_WEEK.map((day) => {
            const isOff = isOffDay(day.value);
            const isSaving = saving === day.value;

            return (
              <button
                key={day.value}
                onClick={() => !isSaving && handleToggleOffDay(day.value)}
                disabled={isSaving}
                className={`
                  p-4 border-2 rounded-lg transition-colors text-center
                  ${isOff 
                    ? 'bg-red-50 border-red-300 text-red-800 hover:bg-red-100' 
                    : 'bg-neutral-bg-secondary border-neutral-border text-neutral-text-primary hover:bg-neutral-hover'
                  }
                  ${isSaving ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
                `}
              >
                <div className="font-medium">{day.label}</div>
                {isOff && (
                  <div className="text-xs mt-1 text-red-600">Off Day</div>
                )}
                {isSaving && (
                  <div className="text-xs mt-1 text-neutral-text-secondary">Saving...</div>
                )}
              </button>
            );
          })}
        </div>

        <div className="mt-4 flex items-center gap-4 text-sm text-neutral-text-secondary">
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 bg-red-50 border-2 border-red-300 rounded"></div>
            <span>Off Day</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 bg-neutral-bg-secondary border-2 border-neutral-border rounded"></div>
            <span>Working Day</span>
          </div>
        </div>
      </div>
    </div>
  );
}
