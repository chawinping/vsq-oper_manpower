'use client';

import { useState, useEffect } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameDay, addMonths, subMonths } from 'date-fns';
import { doctorApi, Doctor, DoctorScheduleOverride } from '@/lib/api/doctor';
import { branchApi, Branch } from '@/lib/api/branch';

interface DoctorScheduleOverridesManagerProps {
  doctor: Doctor;
  currentDate: Date;
  onDateChange: (date: Date) => void;
}

export default function DoctorScheduleOverridesManager({
  doctor,
  currentDate,
  onDateChange,
}: DoctorScheduleOverridesManagerProps) {
  const [overrides, setOverrides] = useState<DoctorScheduleOverride[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState<string | null>(null);
  const [showOverrideDialog, setShowOverrideDialog] = useState<{ date: Date; override: DoctorScheduleOverride | null } | null>(null);
  const [overrideType, setOverrideType] = useState<'working' | 'off'>('working');
  const [selectedBranchId, setSelectedBranchId] = useState<string>('');

  const year = currentDate.getFullYear();
  const month = currentDate.getMonth() + 1;

  useEffect(() => {
    loadData();
  }, [doctor.id, year, month]);

  const loadData = async () => {
    try {
      setLoading(true);
      const monthStart = startOfMonth(currentDate);
      const monthEnd = endOfMonth(currentDate);
      
      const [overridesData, branchesData] = await Promise.all([
        doctorApi.getScheduleOverrides({
          doctor_id: doctor.id,
          start_date: format(monthStart, 'yyyy-MM-dd'),
          end_date: format(monthEnd, 'yyyy-MM-dd'),
        }),
        branchApi.list(),
      ]);
      setOverrides(overridesData);
      setBranches(branchesData);
    } catch (error) {
      console.error('Failed to load overrides:', error);
      alert('Failed to load schedule overrides');
    } finally {
      setLoading(false);
    }
  };

  const getOverrideForDate = (date: Date): DoctorScheduleOverride | undefined => {
    return overrides.find(o => isSameDay(new Date(o.date), date));
  };

  const handleCellClick = (date: Date) => {
    const existingOverride = getOverrideForDate(date);
    setShowOverrideDialog({ date, override: existingOverride || null });
    
    if (existingOverride) {
      setOverrideType(existingOverride.type);
      setSelectedBranchId(existingOverride.branch_id || '');
    } else {
      setOverrideType('working');
      setSelectedBranchId('');
    }
  };

  const handleSaveOverride = async () => {
    if (!showOverrideDialog) return;

    if (overrideType === 'working' && !selectedBranchId) {
      alert('Please select a branch for working day override');
      return;
    }

    const dateStr = format(showOverrideDialog.date, 'yyyy-MM-dd');
    setSaving(dateStr);

    try {
      const overrideData = {
        doctor_id: doctor.id,
        date: dateStr,
        type: overrideType,
        branch_id: overrideType === 'working' ? selectedBranchId : undefined,
      };

      if (showOverrideDialog.override) {
        await doctorApi.updateScheduleOverride(showOverrideDialog.override.id, overrideData);
      } else {
        await doctorApi.createScheduleOverride(overrideData);
      }

      setShowOverrideDialog(null);
      await loadData();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save override');
    } finally {
      setSaving(null);
    }
  };

  const handleDeleteOverride = async () => {
    if (!showOverrideDialog?.override) return;

    if (!confirm('Are you sure you want to remove this override?')) {
      return;
    }

    try {
      await doctorApi.deleteScheduleOverride(showOverrideDialog.override.id);
      setShowOverrideDialog(null);
      await loadData();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete override');
    }
  };

  const goToPreviousMonth = () => {
    onDateChange(subMonths(currentDate, 1));
  };

  const goToNextMonth = () => {
    onDateChange(addMonths(currentDate, 1));
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  const monthStart = startOfMonth(currentDate);
  const monthEnd = endOfMonth(currentDate);
  const daysInMonth = eachDayOfInterval({ start: monthStart, end: monthEnd });

  return (
    <div className="w-full">
      <div className="card">
        <div className="p-6">
          <div className="mb-6 flex items-center justify-between">
            <div>
              <h3 className="text-lg font-semibold text-neutral-text-primary">
                Schedule Overrides - {doctor.name} {doctor.code ? `(${doctor.code})` : ''}
              </h3>
              <p className="text-sm text-neutral-text-secondary mt-1">
                {format(monthStart, 'MMMM yyyy')}
              </p>
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

          <p className="text-sm text-neutral-text-secondary mb-4">
            Override the default schedule for specific dates. Click on a date to set an override.
          </p>

          <div className="overflow-x-auto">
            <div className="grid grid-cols-7 gap-2">
              {daysInMonth.map((day) => {
                const override = getOverrideForDate(day);
                const isToday = isSameDay(day, new Date());
                const isSaving = saving === format(day, 'yyyy-MM-dd');

                let bgColor = 'bg-neutral-bg-secondary';
                let borderColor = 'border-neutral-border';
                let textColor = 'text-neutral-text-primary';
                
                if (override) {
                  if (override.type === 'off') {
                    bgColor = 'bg-red-50';
                    borderColor = 'border-red-300';
                    textColor = 'text-red-800';
                  } else {
                    bgColor = 'bg-yellow-50';
                    borderColor = 'border-yellow-300';
                    textColor = 'text-yellow-800';
                  }
                }

                return (
                  <div
                    key={day.toISOString()}
                    onClick={() => !isSaving && handleCellClick(day)}
                    className={`
                      p-3 border-2 rounded-lg cursor-pointer transition-colors min-h-[80px]
                      ${bgColor} ${borderColor} ${textColor}
                      ${isToday ? 'ring-2 ring-blue-500' : ''}
                      ${isSaving ? 'opacity-50' : 'hover:opacity-80'}
                    `}
                    title={override 
                      ? `${format(day, 'MMM d, yyyy')} - ${override.type === 'off' ? 'Off Day' : `Working at ${override.branch?.code || 'N/A'}`}` 
                      : `${format(day, 'MMM d, yyyy')} - Click to set override`
                    }
                  >
                    <div className="text-xs font-medium mb-1">
                      {format(day, 'EEE')}
                    </div>
                    <div className="text-lg font-bold">
                      {format(day, 'd')}
                    </div>
                    {override && (
                      <div className="text-xs mt-1">
                        {override.type === 'off' ? (
                          <span className="font-semibold">OFF</span>
                        ) : (
                          <span>{override.branch?.code || 'N/A'}</span>
                        )}
                      </div>
                    )}
                    {isSaving && (
                      <div className="text-xs mt-1 text-neutral-text-secondary">Saving...</div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>

          <div className="mt-4 flex gap-4 text-sm text-neutral-text-secondary">
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-yellow-50 border-2 border-yellow-300 rounded"></div>
              <span>Working Day Override</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-red-50 border-2 border-red-300 rounded"></div>
              <span>Off Day Override</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-neutral-bg-secondary border-2 border-neutral-border rounded"></div>
              <span>No Override (Uses Default)</span>
            </div>
          </div>
        </div>
      </div>

      {/* Override Dialog */}
      {showOverrideDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-neutral-bg-secondary rounded-lg p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold text-neutral-text-primary mb-4">
              {showOverrideDialog.override ? 'Edit Override' : 'Create Override'}
            </h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-neutral-text-primary mb-1">
                  Date
                </label>
                <input
                  type="text"
                  value={format(showOverrideDialog.date, 'yyyy-MM-dd')}
                  disabled
                  className="input-field bg-neutral-bg-primary"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-neutral-text-primary mb-1">
                  Override Type <span className="text-red-500">*</span>
                </label>
                <select
                  value={overrideType}
                  onChange={(e) => {
                    setOverrideType(e.target.value as 'working' | 'off');
                    if (e.target.value === 'off') {
                      setSelectedBranchId('');
                    }
                  }}
                  className="input-field"
                >
                  <option value="working">Working Day</option>
                  <option value="off">Off Day</option>
                </select>
              </div>
              {overrideType === 'working' && (
                <div>
                  <label className="block text-sm font-medium text-neutral-text-primary mb-1">
                    Branch <span className="text-red-500">*</span>
                  </label>
                  <select
                    value={selectedBranchId}
                    onChange={(e) => setSelectedBranchId(e.target.value)}
                    className="input-field"
                    required
                  >
                    <option value="">-- Select branch --</option>
                    {branches.map((branch) => (
                      <option key={branch.id} value={branch.id}>
                        {branch.code} - {branch.name}
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </div>
            <div className="flex gap-2 mt-6">
              <button onClick={handleSaveOverride} className="btn-primary flex-1">
                {showOverrideDialog.override ? 'Update' : 'Create'}
              </button>
              {showOverrideDialog.override && (
                <button onClick={handleDeleteOverride} className="btn-secondary text-red-600 hover:text-red-700">
                  Remove
                </button>
              )}
              <button
                onClick={() => setShowOverrideDialog(null)}
                className="btn-secondary flex-1"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
