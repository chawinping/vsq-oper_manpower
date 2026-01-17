'use client';

import { useState, useEffect } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameDay, addMonths, subMonths } from 'date-fns';
import { doctorApi, Doctor, DoctorAssignment } from '@/lib/api/doctor';
import { branchApi, Branch } from '@/lib/api/branch';

interface DoctorScheduleCalendarProps {
  doctor: Doctor;
  branches: Branch[];
  currentDate: Date;
  onDateChange: (date: Date) => void;
}

export default function DoctorScheduleCalendar({
  doctor,
  branches,
  currentDate,
  onDateChange,
}: DoctorScheduleCalendarProps) {
  const [assignments, setAssignments] = useState<DoctorAssignment[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState<string | null>(null);
  const [showBranchDialog, setShowBranchDialog] = useState<{ date: Date; assignment: DoctorAssignment | null } | null>(null);
  const [selectedBranchId, setSelectedBranchId] = useState<string>('');
  const [expectedRevenue, setExpectedRevenue] = useState<number>(0);

  const year = currentDate.getFullYear();
  const month = currentDate.getMonth() + 1;

  useEffect(() => {
    loadSchedule();
  }, [doctor.id, year, month]);

  const loadSchedule = async () => {
    try {
      setLoading(true);
      const data = await doctorApi.getMonthlySchedule(doctor.id, year, month);
      setAssignments(data.assignments || []);
    } catch (error) {
      console.error('Failed to load schedule:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCellClick = (date: Date) => {
    const existingAssignment = assignments.find(a => isSameDay(new Date(a.date), date));
    setShowBranchDialog({ date, assignment: existingAssignment || null });
    if (existingAssignment) {
      setSelectedBranchId(existingAssignment.branch_id);
      setExpectedRevenue(existingAssignment.expected_revenue);
    } else {
      setSelectedBranchId('');
      setExpectedRevenue(0);
    }
  };

  const handleSaveAssignment = async () => {
    if (!selectedBranchId) {
      alert('Please select a branch');
      return;
    }

    if (!showBranchDialog) return;

    const dateStr = format(showBranchDialog.date, 'yyyy-MM-dd');
    setSaving(dateStr);

    try {
      // Check if assignment already exists
      if (showBranchDialog.assignment) {
        // Delete existing assignment
        await doctorApi.deleteAssignment(showBranchDialog.assignment.id);
      }

      // Create new assignment
      await doctorApi.createAssignment({
        doctor_id: doctor.id,
        branch_id: selectedBranchId,
        date: dateStr,
        expected_revenue: expectedRevenue,
      });

      setShowBranchDialog(null);
      await loadSchedule();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save assignment');
    } finally {
      setSaving(null);
    }
  };

  const handleDeleteAssignment = async () => {
    if (!showBranchDialog?.assignment) return;

    if (!confirm('Are you sure you want to remove this assignment?')) {
      return;
    }

    try {
      await doctorApi.deleteAssignment(showBranchDialog.assignment.id);
      setShowBranchDialog(null);
      await loadSchedule();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete assignment');
    }
  };

  const getAssignmentForDate = (date: Date): DoctorAssignment | undefined => {
    return assignments.find(a => isSameDay(new Date(a.date), date));
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
              <h2 className="text-xl font-semibold text-neutral-text-primary">
                {doctor.name} {doctor.code ? `(${doctor.code})` : ''} - Schedule
              </h2>
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

          <div className="overflow-x-auto">
            <table className="w-full border-collapse border border-neutral-border">
              <thead>
                <tr>
                  <th className="border border-neutral-border p-2 bg-neutral-hover font-semibold sticky left-0 z-10 text-neutral-text-primary min-w-[200px]">
                    Doctor
                  </th>
                  {daysInMonth.map((day) => (
                    <th
                      key={day.toISOString()}
                      className="border border-neutral-border p-2 bg-neutral-hover font-semibold min-w-[100px] text-neutral-text-primary"
                    >
                      <div className="text-xs text-neutral-text-secondary">{format(day, 'EEE')}</div>
                      <div className="text-sm">{format(day, 'd')}</div>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td className="border border-neutral-border p-2 bg-neutral-bg-secondary sticky left-0 z-10 font-medium text-neutral-text-primary min-w-[200px]">
                    <div>
                      <div className="font-semibold">{doctor.name}</div>
                      {doctor.code && <div className="text-xs text-neutral-text-secondary">{doctor.code}</div>}
                      {doctor.preferences && <div className="text-xs text-neutral-text-secondary italic mt-1">{doctor.preferences}</div>}
                    </div>
                  </td>
                  {daysInMonth.map((day) => {
                    const assignment = getAssignmentForDate(day);
                    const branch = assignment ? branches.find(b => b.id === assignment.branch_id) : null;
                    const isToday = isSameDay(day, new Date());
                    const isSaving = saving === format(day, 'yyyy-MM-dd');

                    return (
                      <td
                        key={day.toISOString()}
                        onClick={() => !isSaving && handleCellClick(day)}
                        className={`
                          border border-neutral-border p-2 cursor-pointer transition-colors min-h-[60px]
                          ${assignment ? 'bg-blue-50 hover:bg-blue-100' : 'bg-neutral-bg-secondary hover:bg-neutral-hover'}
                          ${isToday ? 'ring-2 ring-blue-500' : ''}
                          ${isSaving ? 'opacity-50' : ''}
                        `}
                        title={assignment ? `${format(day, 'MMM d, yyyy')} - ${branch?.name || 'Unknown Branch'}` : `${format(day, 'MMM d, yyyy')} - Click to assign`}
                      >
                        {assignment ? (
                          <div className="text-xs">
                            <div className="font-semibold text-blue-800">{branch?.code || 'N/A'}</div>
                            <div className="text-blue-600 truncate" title={branch?.name}>{branch?.name || 'Unknown'}</div>
                            {assignment.expected_revenue > 0 && (
                              <div className="text-blue-500 text-[10px] mt-1">
                                {assignment.expected_revenue.toLocaleString()}
                              </div>
                            )}
                          </div>
                        ) : (
                          <div className="text-xs text-neutral-text-secondary">-</div>
                        )}
                      </td>
                    );
                  })}
                </tr>
              </tbody>
            </table>
          </div>

          <div className="mt-4 flex items-center gap-4 text-sm text-neutral-text-secondary">
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-blue-50 border border-blue-300 rounded"></div>
              <span>Assigned</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-neutral-bg-secondary border border-neutral-border rounded"></div>
              <span>Not Assigned</span>
            </div>
          </div>
        </div>
      </div>

      {/* Branch Selection Dialog */}
      {showBranchDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-neutral-bg-secondary rounded-lg p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold text-neutral-text-primary mb-4">
              {showBranchDialog.assignment ? 'Edit Assignment' : 'Assign Doctor to Branch'}
            </h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-neutral-text-primary mb-1">
                  Date
                </label>
                <input
                  type="text"
                  value={format(showBranchDialog.date, 'yyyy-MM-dd')}
                  disabled
                  className="input-field bg-neutral-bg-primary"
                />
              </div>
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
              <div>
                <label className="block text-sm font-medium text-neutral-text-primary mb-1">
                  Expected Revenue
                </label>
                <input
                  type="number"
                  value={expectedRevenue}
                  onChange={(e) => setExpectedRevenue(parseFloat(e.target.value) || 0)}
                  className="input-field"
                  min="0"
                  step="0.01"
                />
              </div>
            </div>
            <div className="flex gap-2 mt-6">
              <button onClick={handleSaveAssignment} className="btn-primary flex-1">
                {showBranchDialog.assignment ? 'Update' : 'Assign'}
              </button>
              {showBranchDialog.assignment && (
                <button onClick={handleDeleteAssignment} className="btn-secondary text-red-600 hover:text-red-700">
                  Remove
                </button>
              )}
              <button
                onClick={() => setShowBranchDialog(null)}
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
