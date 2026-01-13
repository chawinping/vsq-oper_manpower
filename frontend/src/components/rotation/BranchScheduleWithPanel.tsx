'use client';

import { useState, useEffect } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, addMonths, subMonths } from 'date-fns';

// Placeholder data types
interface Staff {
  id: string;
  name: string;
  position: string;
  staffType: 'branch' | 'rotation';
  coverageArea?: string;
  effectiveBranches?: { branchId: string; level: number }[];
}

interface Assignment {
  id: string;
  staffId: string;
  branchId: string;
  date: string; // YYYY-MM-DD
  assignmentLevel?: number;
}

interface Branch {
  id: string;
  name: string;
  code: string;
}

// Placeholder data
const PLACEHOLDER_BRANCHES: Branch[] = [
  { id: '1', name: 'Central Park', code: 'CPN' },
  { id: '2', name: 'Central World', code: 'CTR' },
  { id: '3', name: 'Paragon', code: 'PNK' },
];

const PLACEHOLDER_BRANCH_STAFF: Staff[] = [
  { id: 'bs1', name: 'John Doe', position: 'Branch Manager', staffType: 'branch' },
  { id: 'bs2', name: 'Jane Smith', position: 'Nurse', staffType: 'branch' },
  { id: 'bs3', name: 'Bob Johnson', position: 'Service Consultant', staffType: 'branch' },
];

const PLACEHOLDER_ROTATION_STAFF: Staff[] = [
  { 
    id: 'rs1', 
    name: 'Alice Rotation', 
    position: 'Nurse', 
    staffType: 'rotation', 
    coverageArea: 'Area A',
    effectiveBranches: [{ branchId: '1', level: 1 }, { branchId: '2', level: 2 }]
  },
  { 
    id: 'rs2', 
    name: 'Bob Rotation', 
    position: 'Doctor', 
    staffType: 'rotation', 
    coverageArea: 'Area B',
    effectiveBranches: [{ branchId: '1', level: 1 }]
  },
  { 
    id: 'rs3', 
    name: 'Charlie Rotation', 
    position: 'Service Consultant', 
    staffType: 'rotation', 
    coverageArea: 'Area A',
    effectiveBranches: [{ branchId: '2', level: 1 }, { branchId: '3', level: 2 }]
  },
  { 
    id: 'rs4', 
    name: 'Diana Rotation', 
    position: 'Nurse', 
    staffType: 'rotation', 
    coverageArea: 'Area C',
    effectiveBranches: [{ branchId: '3', level: 1 }]
  },
  { 
    id: 'rs5', 
    name: 'Eve Rotation', 
    position: 'Doctor', 
    staffType: 'rotation', 
    coverageArea: 'Area B',
    effectiveBranches: [{ branchId: '1', level: 2 }, { branchId: '2', level: 1 }]
  },
];

interface BranchScheduleWithPanelProps {
  branchId?: string;
}

export default function BranchScheduleWithPanel({ branchId }: BranchScheduleWithPanelProps) {
  const [selectedBranchId, setSelectedBranchId] = useState<string>(branchId || PLACEHOLDER_BRANCHES[0].id);
  const [currentDate, setCurrentDate] = useState(new Date());
  const [assignments, setAssignments] = useState<Assignment[]>([]);
  const [rotationStaffInSchedule, setRotationStaffInSchedule] = useState<string[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [showBulkDialog, setShowBulkDialog] = useState(false);

  const monthStart = startOfMonth(currentDate);
  const monthEnd = endOfMonth(currentDate);
  const daysInMonth = eachDayOfInterval({ start: monthStart, end: monthEnd });

  const selectedBranch = PLACEHOLDER_BRANCHES.find(b => b.id === selectedBranchId);
  const branchStaff = PLACEHOLDER_BRANCH_STAFF;
  
  // Filter eligible rotation staff based on effective branches
  const eligibleRotationStaff = PLACEHOLDER_ROTATION_STAFF.filter(staff =>
    staff.effectiveBranches?.some(eb => eb.branchId === selectedBranchId)
  );

  // Filter rotation staff by search query
  const filteredRotationStaff = eligibleRotationStaff.filter(staff =>
    staff.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    staff.position.toLowerCase().includes(searchQuery.toLowerCase()) ||
    staff.coverageArea?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Get rotation staff that are in the schedule
  const rotationStaffRows = PLACEHOLDER_ROTATION_STAFF.filter(staff =>
    rotationStaffInSchedule.includes(staff.id)
  );

  // Combine branch staff and rotation staff in schedule
  const scheduleStaff: Staff[] = [
    ...branchStaff,
    ...rotationStaffRows,
  ];

  const isAssigned = (staffId: string, date: Date): boolean => {
    const dateStr = format(date, 'yyyy-MM-dd');
    return assignments.some(
      a => a.staffId === staffId && a.date === dateStr && a.branchId === selectedBranchId
    );
  };

  const handleAddRotationStaff = (staffId: string) => {
    if (!rotationStaffInSchedule.includes(staffId)) {
      setRotationStaffInSchedule([...rotationStaffInSchedule, staffId]);
    }
  };

  const handleRemoveRotationStaff = (staffId: string) => {
    // Remove from schedule
    setRotationStaffInSchedule(rotationStaffInSchedule.filter(id => id !== staffId));
    // Remove all assignments for this staff
    setAssignments(assignments.filter(a => a.staffId !== staffId));
  };

  const handleDateClick = (staffId: string, date: Date) => {
    const staff = scheduleStaff.find(s => s.id === staffId);
    // Only allow assignment for rotation staff
    if (staff?.staffType !== 'rotation') return;

    const dateStr = format(date, 'yyyy-MM-dd');
    const existingAssignment = assignments.find(
      a => a.staffId === staffId && a.date === dateStr && a.branchId === selectedBranchId
    );

    if (existingAssignment) {
      // Remove assignment
      setAssignments(assignments.filter(a => a.id !== existingAssignment.id));
    } else {
      // Add assignment
      const staff = PLACEHOLDER_ROTATION_STAFF.find(s => s.id === staffId);
      const effectiveBranch = staff?.effectiveBranches?.find(eb => eb.branchId === selectedBranchId);
      const newAssignment: Assignment = {
        id: `assign-${staffId}-${dateStr}`,
        staffId,
        branchId: selectedBranchId,
        date: dateStr,
        assignmentLevel: effectiveBranch?.level || 1,
      };
      setAssignments([...assignments, newAssignment]);
    }
  };

  const handleSave = () => {
    setLoading(true);
    // Simulate API call
    setTimeout(() => {
      alert(`Saved ${assignments.length} assignments for ${selectedBranch?.name}`);
      setLoading(false);
    }, 500);
  };

  const goToPreviousMonth = () => setCurrentDate(subMonths(currentDate, 1));
  const goToNextMonth = () => setCurrentDate(addMonths(currentDate, 1));
  const goToToday = () => setCurrentDate(new Date());

  return (
    <div className="w-full p-6">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between flex-wrap gap-4">
        <div className="flex items-center gap-4">
          <h2 className="text-2xl font-semibold">Rotation Staff Assignment</h2>
          <select
            value={selectedBranchId}
            onChange={(e) => {
              setSelectedBranchId(e.target.value);
              // Clear rotation staff from schedule when branch changes
              setRotationStaffInSchedule([]);
              setAssignments([]);
            }}
            className="px-4 py-2 border border-neutral-border rounded-md bg-white text-neutral-text-primary"
          >
            {PLACEHOLDER_BRANCHES.map(branch => (
              <option key={branch.id} value={branch.id}>
                {branch.name} ({branch.code})
              </option>
            ))}
          </select>
        </div>
        <div className="flex gap-2">
          <button
            onClick={goToPreviousMonth}
            className="btn-secondary"
          >
            ◀ Previous
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
            Next ▶
          </button>
        </div>
      </div>

      {/* Month Header */}
      <div className="mb-4">
        <h3 className="text-lg font-semibold text-neutral-text-primary">
          {format(currentDate, 'MMMM yyyy')}
        </h3>
      </div>

      <div className="flex gap-4">
        {/* Main Schedule View */}
        <div className="flex-1">
          <div className="card overflow-x-auto">
            <table className="w-full border-collapse">
              <thead>
                <tr className="bg-neutral-bg-secondary">
                  <th className="border border-neutral-border p-3 text-left font-semibold sticky left-0 z-10 bg-neutral-bg-secondary min-w-[200px]">
                    Staff Name / Position
                  </th>
                  {daysInMonth.map((day) => (
                    <th
                      key={day.toISOString()}
                      className="border border-neutral-border p-2 text-center font-semibold bg-neutral-bg-secondary min-w-[80px]"
                    >
                      <div className="text-xs text-neutral-text-secondary">{format(day, 'EEE')}</div>
                      <div className="text-sm font-semibold">{format(day, 'd')}</div>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {scheduleStaff.map((staff, index) => {
                  const isRotationStaff = staff.staffType === 'rotation';
                  const isBranchStaffRow = !isRotationStaff && index === branchStaff.length - 1;
                  
                  return (
                    <tr
                      key={staff.id}
                      className={`${
                        isRotationStaff
                          ? 'bg-blue-50 hover:bg-blue-100'
                          : 'bg-white hover:bg-neutral-hover'
                      } ${isBranchStaffRow ? 'border-b-2 border-neutral-border' : ''}`}
                    >
                      <td className="border border-neutral-border p-3 sticky left-0 z-10 bg-inherit">
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{staff.name}</span>
                          {isRotationStaff && (
                            <>
                              <span className="px-2 py-0.5 text-xs bg-blue-200 text-blue-800 rounded">
                                Rotation
                              </span>
                              <button
                                onClick={() => handleRemoveRotationStaff(staff.id)}
                                className="ml-auto text-red-600 hover:text-red-800 text-xs"
                                title="Remove from schedule"
                              >
                                × Remove
                              </button>
                            </>
                          )}
                        </div>
                        <div className="text-sm text-neutral-text-secondary">{staff.position}</div>
                      </td>
                      {daysInMonth.map((day) => {
                        const assigned = isAssigned(staff.id, day);
                        const isClickable = isRotationStaff;
                        const isPast = day < new Date(new Date().setHours(0, 0, 0, 0));

                        return (
                          <td
                            key={day.toISOString()}
                            className={`border border-neutral-border p-1 text-center ${
                              isClickable
                                ? assigned
                                  ? 'bg-green-200 hover:bg-green-300 cursor-pointer'
                                  : isPast
                                  ? 'bg-gray-100 cursor-not-allowed'
                                  : 'bg-white hover:bg-blue-50 cursor-pointer'
                                : 'bg-gray-50'
                            }`}
                            onClick={() => isClickable && !isPast && handleDateClick(staff.id, day)}
                            title={
                              isClickable
                                ? assigned
                                  ? `Click to remove assignment for ${format(day, 'MMM d')}`
                                  : `Click to assign for ${format(day, 'MMM d')}`
                                : 'Branch staff schedule (read-only)'
                            }
                          >
                            {isClickable && (
                              <div className="flex items-center justify-center h-8">
                                {assigned ? (
                                  <span className="text-green-700 font-bold">✓</span>
                                ) : (
                                  <span className="text-neutral-text-secondary text-xs">○</span>
                                )}
                              </div>
                            )}
                            {!isClickable && (
                              <div className="flex items-center justify-center h-8">
                                <span className="text-neutral-text-secondary text-xs">—</span>
                              </div>
                            )}
                          </td>
                        );
                      })}
                    </tr>
                  );
                })}
                {/* Add Rotation Staff Row */}
                <tr className="bg-gray-100 hover:bg-gray-200">
                  <td className="border border-neutral-border p-3 sticky left-0 z-10 bg-inherit">
                    <div className="flex items-center gap-2 text-neutral-text-secondary italic">
                      <span>+ Add Rotation Staff</span>
                    </div>
                  </td>
                  {daysInMonth.map((day) => (
                    <td key={day.toISOString()} className="border border-neutral-border p-1 bg-gray-50"></td>
                  ))}
                </tr>
              </tbody>
            </table>
          </div>

          {/* Legend and Actions */}
          <div className="mt-4 flex items-center justify-between flex-wrap gap-4">
            <div className="flex items-center gap-4 text-sm">
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-green-200 border border-neutral-border"></div>
                <span>Assigned</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-white border border-neutral-border"></div>
                <span>Available</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-gray-50 border border-neutral-border"></div>
                <span>Branch Staff (Read-only)</span>
              </div>
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => {
                  setRotationStaffInSchedule([]);
                  setAssignments([]);
                }}
                className="btn-secondary"
                disabled={assignments.length === 0 && rotationStaffInSchedule.length === 0}
              >
                Clear All
              </button>
              <button
                onClick={handleSave}
                className="btn-primary"
                disabled={loading}
              >
                {loading ? 'Saving...' : `Save Changes (${assignments.length})`}
              </button>
            </div>
          </div>
        </div>

        {/* Eligible Rotation Staff Panel */}
        <div className="w-80 card">
          <div className="mb-4">
            <h3 className="text-lg font-semibold mb-2">Eligible Rotation Staff</h3>
            <div className="relative">
              <input
                type="text"
                placeholder="Search rotation staff..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full px-3 py-2 border border-neutral-border rounded-md text-sm"
              />
              <span className="absolute right-3 top-2.5 text-neutral-text-secondary">🔍</span>
            </div>
            <div className="mt-2 text-xs text-neutral-text-secondary">
              {filteredRotationStaff.length} staff available
            </div>
          </div>

          <div className="space-y-3 max-h-[600px] overflow-y-auto">
            {filteredRotationStaff.length === 0 ? (
              <div className="text-center py-8 text-neutral-text-secondary text-sm">
                No rotation staff found
              </div>
            ) : (
              filteredRotationStaff.map((staff) => {
                const isInSchedule = rotationStaffInSchedule.includes(staff.id);
                const effectiveBranch = staff.effectiveBranches?.find(eb => eb.branchId === selectedBranchId);
                
                return (
                  <div
                    key={staff.id}
                    className={`p-3 border rounded-md ${
                      isInSchedule
                        ? 'border-green-300 bg-green-50'
                        : 'border-neutral-border bg-white hover:bg-neutral-hover'
                    }`}
                  >
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex-1">
                        <div className="font-medium text-sm">{staff.name}</div>
                        <div className="text-xs text-neutral-text-secondary">{staff.position}</div>
                      </div>
                      {isInSchedule && (
                        <span className="px-2 py-0.5 text-xs bg-green-200 text-green-800 rounded">
                          Added
                        </span>
                      )}
                    </div>
                    {staff.coverageArea && (
                      <div className="text-xs text-neutral-text-secondary mb-2">
                        Coverage: {staff.coverageArea}
                      </div>
                    )}
                    {effectiveBranch && (
                      <div className="text-xs text-neutral-text-secondary mb-2">
                        Level {effectiveBranch.level} branch
                      </div>
                    )}
                    <button
                      onClick={() => isInSchedule ? handleRemoveRotationStaff(staff.id) : handleAddRotationStaff(staff.id)}
                      className={`w-full mt-2 px-3 py-1.5 text-xs rounded transition-colors ${
                        isInSchedule
                          ? 'bg-red-100 text-red-700 hover:bg-red-200'
                          : 'bg-blue-100 text-blue-700 hover:bg-blue-200'
                      }`}
                    >
                      {isInSchedule ? 'Remove from Schedule' : 'Add to Schedule'}
                    </button>
                  </div>
                );
              })
            )}
          </div>
        </div>
      </div>

      {/* Info Box */}
      <div className="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-md">
        <p className="text-sm text-blue-800">
          <strong>Instructions:</strong> Click "Add to Schedule" in the panel to add rotation staff to the schedule. 
          Then click on date cells in the schedule to assign/unassign them. Branch staff schedules are shown for reference.
        </p>
      </div>
    </div>
  );
}

