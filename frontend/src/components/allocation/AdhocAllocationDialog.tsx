'use client';

import { useState, useEffect } from 'react';
import { format } from 'date-fns';
import { rotationApi, AssignRotationRequest } from '@/lib/api/rotation';
import { staffApi, Staff } from '@/lib/api/staff';
import { branchApi, Branch } from '@/lib/api/branch';

interface AdhocAllocationDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
  branchId?: string;
  date?: Date;
}

export default function AdhocAllocationDialog({
  isOpen,
  onClose,
  onSuccess,
  branchId: initialBranchId,
  date: initialDate,
}: AdhocAllocationDialogProps) {
  const [branchId, setBranchId] = useState(initialBranchId || '');
  const [date, setDate] = useState(initialDate ? format(initialDate, 'yyyy-MM-dd') : '');
  const [rotationStaffId, setRotationStaffId] = useState('');
  const [assignmentLevel, setAssignmentLevel] = useState<1 | 2>(1);
  const [reason, setReason] = useState('');
  const [loading, setLoading] = useState(false);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [rotationStaff, setRotationStaff] = useState<Staff[]>([]);

  useEffect(() => {
    if (isOpen) {
      loadData();
    }
  }, [isOpen]);

  const loadData = async () => {
    try {
      const [branchesData, staffData] = await Promise.all([
        branchApi.list(),
        staffApi.list({ staff_type: 'rotation' }),
      ]);
      setBranches(branchesData || []);
      setRotationStaff(staffData || []);
    } catch (error) {
      console.error('Failed to load data:', error);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!branchId || !date || !rotationStaffId || !reason.trim()) {
      alert('Please fill in all required fields');
      return;
    }

    try {
      setLoading(true);
      const request: AssignRotationRequest = {
        rotation_staff_id: rotationStaffId,
        branch_id: branchId,
        date,
        assignment_level: assignmentLevel,
        is_adhoc: true,
        adhoc_reason: reason,
      };
      await rotationApi.assign(request);
      onSuccess();
      onClose();
      // Reset form
      setBranchId(initialBranchId || '');
      setDate(initialDate ? format(initialDate, 'yyyy-MM-dd') : '');
      setRotationStaffId('');
      setReason('');
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to create adhoc allocation');
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-full max-w-md">
        <h2 className="text-xl font-bold mb-4">Create Adhoc Allocation</h2>
        <form onSubmit={handleSubmit}>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Branch *</label>
              <select
                value={branchId}
                onChange={(e) => setBranchId(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                required
              >
                <option value="">Select branch</option>
                {branches.map((b) => (
                  <option key={b.id} value={b.id}>
                    {b.name} ({b.code})
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Date *</label>
              <input
                type="date"
                value={date}
                onChange={(e) => setDate(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Rotation Staff *</label>
              <select
                value={rotationStaffId}
                onChange={(e) => setRotationStaffId(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                required
              >
                <option value="">Select staff</option>
                {rotationStaff.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name} {s.nickname && `(${s.nickname})`}
                  </option>
                ))}
              </select>
            </div>

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

            <div>
              <label className="block text-sm font-medium mb-1">Reason for Adhoc Allocation *</label>
              <textarea
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                rows={3}
                placeholder="e.g., Unplanned sick leave, emergency coverage needed..."
                required
              />
            </div>
          </div>

          <div className="mt-6 flex gap-3 justify-end">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-md"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? 'Creating...' : 'Create Allocation'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
