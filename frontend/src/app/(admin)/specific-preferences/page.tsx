'use client';

import { useState, useEffect } from 'react';
import { specificPreferenceApi, SpecificPreference, CreateSpecificPreferenceRequest, SpecificPreferenceType } from '@/lib/api/specific-preference';
import { branchApi } from '@/lib/api/branch';
import { doctorApi } from '@/lib/api/doctor';
import { positionApi } from '@/lib/api/position';
import { staffApi } from '@/lib/api/staff';

interface Branch {
  id: string;
  name: string;
  code: string;
}

interface Doctor {
  id: string;
  name: string;
  code: string;
}

interface Position {
  id: string;
  name: string;
}

interface Staff {
  id: string;
  name: string;
  nickname: string;
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

export default function SpecificPreferencesPage() {
  const [preferences, setPreferences] = useState<SpecificPreference[]>([]);
  const [loading, setLoading] = useState(true);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [doctors, setDoctors] = useState<Doctor[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [staffList, setStaffList] = useState<Staff[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [editingPreference, setEditingPreference] = useState<SpecificPreference | null>(null);
  const [formData, setFormData] = useState<CreateSpecificPreferenceRequest>({
    preference_type: 'position_count',
    is_active: true,
  });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [prefsRes, branchesRes, doctorsRes, positionsRes, staffRes] = await Promise.all([
        specificPreferenceApi.list(),
        branchApi.list(),
        doctorApi.list(),
        positionApi.list(),
        staffApi.list({}),
      ]);
      setPreferences(prefsRes.preferences);
      setBranches(branchesRes.branches);
      setDoctors(doctorsRes.doctors);
      setPositions(positionsRes.positions);
      setStaffList(staffRes.staff);
    } catch (error) {
      console.error('Failed to load data:', error);
      alert('Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingPreference(null);
    setFormData({
      preference_type: 'position_count',
      is_active: true,
    });
    setShowModal(true);
  };

  const handleEdit = (pref: SpecificPreference) => {
    setEditingPreference(pref);
    setFormData({
      branch_id: pref.branch_id || null,
      doctor_id: pref.doctor_id || null,
      day_of_week: pref.day_of_week ?? null,
      preference_type: pref.preference_type,
      position_id: pref.position_id || undefined,
      staff_count: pref.staff_count || undefined,
      staff_id: pref.staff_id || undefined,
      is_active: pref.is_active,
    });
    setShowModal(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this preference?')) return;
    try {
      await specificPreferenceApi.delete(id);
      await loadData();
      alert('Preference deleted successfully');
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete preference');
    }
  };

  const handleSubmit = async () => {
    try {
      if (editingPreference) {
        await specificPreferenceApi.update(editingPreference.id, formData);
        alert('Preference updated successfully');
      } else {
        await specificPreferenceApi.create(formData);
        alert('Preference created successfully');
      }
      setShowModal(false);
      await loadData();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save preference');
    }
  };

  const getDisplayText = (pref: SpecificPreference) => {
    const parts: string[] = [];
    
    if (pref.branch) {
      parts.push(`Branch: ${pref.branch.name}`);
    } else {
      parts.push('Branch: Any');
    }
    
    if (pref.doctor) {
      parts.push(`Doctor: ${pref.doctor.name}`);
    } else {
      parts.push('Doctor: Any');
    }
    
    if (pref.day_of_week !== undefined && pref.day_of_week !== null) {
      parts.push(`Day: ${DAYS_OF_WEEK[pref.day_of_week].label}`);
    } else {
      parts.push('Day: Any');
    }
    
    if (pref.preference_type === 'position_count' && pref.position && pref.staff_count !== undefined) {
      parts.push(`→ ${pref.staff_count} ${pref.position.name}`);
    } else if (pref.preference_type === 'staff_name' && pref.staff) {
      parts.push(`→ Staff: ${pref.staff.name}${pref.staff.nickname ? ` (${pref.staff.nickname})` : ''}`);
    }
    
    return parts.join(' | ');
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading specific preferences...</div>
      </div>
    );
  }

  return (
    <div className="w-full p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Specific Preferences</h1>
        <p className="text-gray-600">
          Manage preferences that can override or mix with other filters. Set rules based on combinations of branch, doctor, and day of week.
        </p>
      </div>

      {/* Info Banner */}
      <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
        <div className="flex items-start">
          <div className="text-2xl mr-3">ℹ️</div>
          <div>
            <h3 className="font-semibold text-blue-900 mb-1">How It Works</h3>
            <p className="text-sm text-blue-800 mb-2">
              Create preferences that specify staff requirements based on combinations of:
            </p>
            <ul className="text-sm text-blue-800 list-disc list-inside space-y-1">
              <li><strong>Branch:</strong> Specific branch or any branch</li>
              <li><strong>Doctor:</strong> Specific doctor or any doctor</li>
              <li><strong>Day of Week:</strong> Specific day (Sunday-Saturday) or any day</li>
            </ul>
            <p className="text-sm text-blue-800 mt-2">
              Each preference can either require a <strong>number of staff for a position</strong> or a <strong>specific staff member</strong>.
            </p>
          </div>
        </div>
      </div>

      {/* Action Button */}
      <div className="mb-6">
        <button
          onClick={handleCreate}
          className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          + Create New Preference
        </button>
      </div>

      {/* Preferences List */}
      <div className="space-y-4">
        {preferences.length === 0 ? (
          <div className="text-center py-12 text-gray-500">
            No preferences found. Create your first preference to get started.
          </div>
        ) : (
          preferences.map((pref) => (
            <div
              key={pref.id}
              className={`border rounded-lg p-4 ${pref.is_active ? 'bg-white border-gray-300' : 'bg-gray-50 border-gray-200 opacity-60'}`}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-2">
                    <span className={`px-2 py-1 text-xs rounded ${pref.is_active ? 'bg-green-100 text-green-800' : 'bg-gray-200 text-gray-600'}`}>
                      {pref.is_active ? 'Active' : 'Inactive'}
                    </span>
                    <span className={`px-2 py-1 text-xs rounded ${
                      pref.preference_type === 'position_count' ? 'bg-purple-100 text-purple-800' : 'bg-orange-100 text-orange-800'
                    }`}>
                      {pref.preference_type === 'position_count' ? 'Position Count' : 'Staff Name'}
                    </span>
                  </div>
                  <p className="text-gray-800 font-medium">{getDisplayText(pref)}</p>
                  <p className="text-xs text-gray-500 mt-1">
                    Created: {new Date(pref.created_at).toLocaleString()}
                  </p>
                </div>
                <div className="flex gap-2 ml-4">
                  <button
                    onClick={() => handleEdit(pref)}
                    className="px-3 py-1 text-sm bg-gray-200 text-gray-700 rounded hover:bg-gray-300"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => handleDelete(pref.id)}
                    className="px-3 py-1 text-sm bg-red-200 text-red-700 rounded hover:bg-red-300"
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
            <h2 className="text-2xl font-bold mb-4">
              {editingPreference ? 'Edit Preference' : 'Create New Preference'}
            </h2>

            <div className="space-y-4">
              {/* Branch */}
              <div>
                <label className="block text-sm font-medium mb-1">Branch</label>
                <select
                  value={formData.branch_id || ''}
                  onChange={(e) => setFormData({ ...formData, branch_id: e.target.value || null })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                >
                  <option value="">Any Branch</option>
                  {branches.map((b) => (
                    <option key={b.id} value={b.id}>{b.name} ({b.code})</option>
                  ))}
                </select>
              </div>

              {/* Doctor */}
              <div>
                <label className="block text-sm font-medium mb-1">Doctor</label>
                <select
                  value={formData.doctor_id || ''}
                  onChange={(e) => setFormData({ ...formData, doctor_id: e.target.value || null })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                >
                  <option value="">Any Doctor</option>
                  {doctors.map((d) => (
                    <option key={d.id} value={d.id}>{d.name} {d.code ? `(${d.code})` : ''}</option>
                  ))}
                </select>
              </div>

              {/* Day of Week */}
              <div>
                <label className="block text-sm font-medium mb-1">Day of Week</label>
                <select
                  value={formData.day_of_week ?? ''}
                  onChange={(e) => setFormData({ ...formData, day_of_week: e.target.value === '' ? null : parseInt(e.target.value) })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                >
                  <option value="">Any Day</option>
                  {DAYS_OF_WEEK.map((day) => (
                    <option key={day.value} value={day.value}>{day.label}</option>
                  ))}
                </select>
              </div>

              {/* Preference Type */}
              <div>
                <label className="block text-sm font-medium mb-1">Preference Type</label>
                <select
                  value={formData.preference_type}
                  onChange={(e) => {
                    const newType = e.target.value as SpecificPreferenceType;
                    setFormData({
                      ...formData,
                      preference_type: newType,
                      // Clear fields that don't apply to the new type
                      position_id: newType === 'position_count' ? formData.position_id : undefined,
                      staff_count: newType === 'position_count' ? formData.staff_count : undefined,
                      staff_id: newType === 'staff_name' ? formData.staff_id : undefined,
                    });
                  }}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                >
                  <option value="position_count">Number of Staff for Position</option>
                  <option value="staff_name">Specific Staff Member</option>
                </select>
              </div>

              {/* Position Count Fields */}
              {formData.preference_type === 'position_count' && (
                <>
                  <div>
                    <label className="block text-sm font-medium mb-1">Position *</label>
                    <select
                      value={formData.position_id || ''}
                      onChange={(e) => setFormData({ ...formData, position_id: e.target.value || undefined })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      required
                    >
                      <option value="">Select Position</option>
                      {positions.map((p) => (
                        <option key={p.id} value={p.id}>{p.name}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium mb-1">Staff Count *</label>
                    <input
                      type="number"
                      min="1"
                      value={formData.staff_count || ''}
                      onChange={(e) => setFormData({ ...formData, staff_count: parseInt(e.target.value) || undefined })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      required
                    />
                  </div>
                </>
              )}

              {/* Staff Name Fields */}
              {formData.preference_type === 'staff_name' && (
                <div>
                  <label className="block text-sm font-medium mb-1">Staff Member *</label>
                  <select
                    value={formData.staff_id || ''}
                    onChange={(e) => setFormData({ ...formData, staff_id: e.target.value || undefined })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    required
                  >
                    <option value="">Select Staff</option>
                    {staffList.map((s) => (
                      <option key={s.id} value={s.id}>
                        {s.name} {s.nickname ? `(${s.nickname})` : ''}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              {/* Is Active */}
              <div>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={formData.is_active ?? true}
                    onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                    className="mr-2"
                  />
                  <span className="text-sm font-medium">Active</span>
                </label>
              </div>
            </div>

            {/* Modal Actions */}
            <div className="flex gap-4 justify-end mt-6 pt-4 border-t">
              <button
                onClick={() => setShowModal(false)}
                className="px-6 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300"
              >
                Cancel
              </button>
              <button
                onClick={handleSubmit}
                className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                {editingPreference ? 'Update' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
