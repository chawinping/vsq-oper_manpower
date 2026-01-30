'use client';

import { useState, useEffect } from 'react';
import { branchApi, Branch } from '@/lib/api/branch';
import { rotationApi, AssignRotationRequest } from '@/lib/api/rotation';
import { staffApi, Staff } from '@/lib/api/staff';
import { positionApi, Position } from '@/lib/api/position';

interface BranchDetailDrawerProps {
  isOpen: boolean;
  branchId: string;
  date: string;
  onClose: () => void;
  onSuccess: () => void;
}

export default function BranchDetailDrawer({
  isOpen,
  branchId,
  date,
  onClose,
  onSuccess,
}: BranchDetailDrawerProps) {
  const [branch, setBranch] = useState<Branch | null>(null);
  const [availableStaff, setAvailableStaff] = useState<Staff[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [loading, setLoading] = useState(false);
  const [assigning, setAssigning] = useState(false);
  const [selectedStaffId, setSelectedStaffId] = useState('');
  const [selectedPositionId, setSelectedPositionId] = useState('');
  const [assignmentLevel, setAssignmentLevel] = useState<1 | 2>(1);

  useEffect(() => {
    if (isOpen) {
      loadData();
    }
  }, [isOpen, branchId, date]);

  const loadData = async () => {
    try {
      setLoading(true);
      const [branchesData, staffData, positionsData] = await Promise.all([
        branchApi.list(),
        staffApi.list({ staff_type: 'rotation' }),
        positionApi.list(),
      ]);

      const branchData = branchesData.find(b => b.id === branchId);
      setBranch(branchData || null);
      setAvailableStaff(staffData || []);
      setPositions(positionsData || []);
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleAssign = async () => {
    if (!selectedStaffId || !selectedPositionId) {
      alert('Please select staff and position');
      return;
    }

    try {
      setAssigning(true);
      const request: AssignRotationRequest = {
        rotation_staff_id: selectedStaffId,
        branch_id: branchId,
        date,
        assignment_level: assignmentLevel,
      };
      await rotationApi.assign(request);
      onSuccess();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to assign staff');
    } finally {
      setAssigning(false);
    }
  };

  if (!isOpen) return null;

  const filteredStaff = availableStaff.filter(s => {
    if (selectedPositionId) {
      return s.position_id === selectedPositionId;
    }
    return true;
  });

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-end z-50">
      <div className="bg-white h-full w-full max-w-2xl overflow-y-auto">
        <div className="sticky top-0 bg-white border-b border-gray-200 p-6 flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold">
              {branch?.code} - {branch?.name}
            </h2>
            <p className="text-sm text-gray-600 mt-1">Date: {date}</p>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 text-2xl font-bold"
          >
            ×
          </button>
        </div>

        <div className="p-6">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-lg text-gray-600">Loading...</div>
            </div>
          ) : (
            <>
              {/* Add Rotation Staff Form */}
              <div className="border-t border-gray-200 pt-6">
                <h3 className="text-lg font-semibold mb-4">Add Rotation Staff</h3>
                <div className="space-y-4">
                  {/* Position Selection */}
                  <div>
                    <label className="block text-sm font-medium mb-1">Position *</label>
                    <select
                      value={selectedPositionId}
                      onChange={(e) => {
                        setSelectedPositionId(e.target.value);
                        setSelectedStaffId(''); // Reset staff selection
                      }}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      required
                    >
                      <option value="">Select position</option>
                      {positions.map(position => (
                        <option key={position.id} value={position.id}>
                          {position.name}
                        </option>
                      ))}
                    </select>
                  </div>

                  {/* Staff Selection */}
                  <div>
                    <label className="block text-sm font-medium mb-1">Rotation Staff *</label>
                    <select
                      value={selectedStaffId}
                      onChange={(e) => setSelectedStaffId(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      required
                      disabled={!selectedPositionId}
                    >
                      <option value="">Select staff</option>
                      {filteredStaff.map(staff => (
                        <option key={staff.id} value={staff.id}>
                          {staff.name} {staff.nickname && `(${staff.nickname})`}
                          {staff.position && ` - ${staff.position.name}`}
                        </option>
                      ))}
                    </select>
                  </div>

                  {/* Assignment Level */}
                  <div>
                    <label className="block text-sm font-medium mb-1">Assignment Level *</label>
                    <select
                      value={assignmentLevel}
                      onChange={(e) => setAssignmentLevel(parseInt(e.target.value) as 1 | 2)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      required
                    >
                      <option value="1">Level 1 (Priority)</option>
                      <option value="2">Level 2 (Reserved)</option>
                    </select>
                  </div>

                  {/* Action Buttons */}
                  <div className="flex gap-3 pt-4">
                    <button
                      onClick={onClose}
                      className="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-md"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={handleAssign}
                      disabled={assigning || !selectedStaffId || !selectedPositionId}
                      className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {assigning ? 'Assigning...' : 'Assign Staff'}
                    </button>
                  </div>
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
