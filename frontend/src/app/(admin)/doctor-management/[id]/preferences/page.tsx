'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { useUser } from '@/contexts/UserContext';
import { doctorApi, Doctor, DoctorPreference, CreateDoctorPreferenceRequest } from '@/lib/api/doctor';
import { branchApi, Branch } from '@/lib/api/branch';
import { positionApi, Position } from '@/lib/api/position';
import Link from 'next/link';

export default function DoctorPreferencesPage() {
  const router = useRouter();
  const params = useParams();
  const doctorId = params.id as string;
  const { user, loading: userLoading } = useUser();
  const [doctor, setDoctor] = useState<Doctor | null>(null);
  const [preferences, setPreferences] = useState<DoctorPreference[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingPreference, setEditingPreference] = useState<DoctorPreference | null>(null);

  const [formData, setFormData] = useState<CreateDoctorPreferenceRequest>({
    doctor_id: doctorId,
    branch_id: undefined,
    rule_type: 'staff_requirement',
    rule_config: {
      requirements: [],
    },
    is_active: true,
  });

  useEffect(() => {
    if (!userLoading && user && !['admin', 'area_manager'].includes(user.role || '')) {
      router.push('/dashboard');
      return;
    }
  }, [user, userLoading, router]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        if (user && ['admin', 'area_manager'].includes(user.role || '')) {
          await Promise.all([
            loadDoctor(),
            loadPreferences(),
            loadBranches(),
            loadPositions(),
          ]);
        }
      } catch (error: any) {
        console.error('Failed to fetch data:', error);
      } finally {
        setLoading(false);
      }
    };

    if (user && doctorId) {
      fetchData();
    }
  }, [user, doctorId]);

  const loadDoctor = async () => {
    try {
      const doctorData = await doctorApi.getById(doctorId);
      setDoctor(doctorData);
    } catch (error) {
      console.error('Failed to load doctor:', error);
    }
  };

  const loadPreferences = async () => {
    try {
      const preferencesData = await doctorApi.getPreferences(doctorId);
      setPreferences(preferencesData || []);
    } catch (error) {
      console.error('Failed to load preferences:', error);
      setPreferences([]);
    }
  };

  const loadBranches = async () => {
    try {
      const branchesData = await branchApi.list();
      setBranches(branchesData || []);
    } catch (error) {
      console.error('Failed to load branches:', error);
    }
  };

  const loadPositions = async () => {
    try {
      const positionsData = await positionApi.list();
      setPositions(positionsData || []);
    } catch (error) {
      console.error('Failed to load positions:', error);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingPreference) {
        await doctorApi.updatePreference(editingPreference.id, formData);
      } else {
        await doctorApi.createPreference(formData);
      }

      setShowModal(false);
      setEditingPreference(null);
      resetForm();
      await loadPreferences();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save preference');
    }
  };

  const resetForm = () => {
    setFormData({
      doctor_id: doctorId,
      branch_id: undefined,
      rule_type: 'staff_requirement',
      rule_config: {
        requirements: [],
      },
      is_active: true,
    });
  };

  const handleEdit = (preferenceToEdit: DoctorPreference) => {
    setEditingPreference(preferenceToEdit);
    setFormData({
      doctor_id: preferenceToEdit.doctor_id,
      branch_id: preferenceToEdit.branch_id || undefined,
      rule_type: preferenceToEdit.rule_type,
      rule_config: preferenceToEdit.rule_config,
      is_active: preferenceToEdit.is_active,
    });
    setShowModal(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this preference rule?')) {
      return;
    }

    try {
      await doctorApi.deletePreference(id);
      await loadPreferences();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete preference');
    }
  };

  const addRequirement = () => {
    const requirements = (formData.rule_config.requirements || []) as any[];
    setFormData({
      ...formData,
      rule_config: {
        ...formData.rule_config,
        requirements: [...requirements, { position_id: '', min_count: 1 }],
      },
    });
  };

  const removeRequirement = (index: number) => {
    const requirements = (formData.rule_config.requirements || []) as any[];
    setFormData({
      ...formData,
      rule_config: {
        ...formData.rule_config,
        requirements: requirements.filter((_, i) => i !== index),
      },
    });
  };

  const updateRequirement = (index: number, field: string, value: any) => {
    const requirements = (formData.rule_config.requirements || []) as any[];
    const updated = [...requirements];
    updated[index] = { ...updated[index], [field]: value };
    setFormData({
      ...formData,
      rule_config: {
        ...formData.rule_config,
        requirements: updated,
      },
    });
  };

  if (loading || userLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-neutral-text-secondary">Loading...</div>
      </div>
    );
  }

  if (!user || !['admin', 'area_manager'].includes(user.role || '')) {
    return null;
  }

  return (
    <>
      <div className="p-6">
        <div className="mb-6">
          <div className="flex items-center gap-2 mb-2">
            <Link href="/doctor-management" className="text-neutral-text-secondary hover:text-neutral-text-primary">
              ← Back to Doctors
            </Link>
          </div>
          <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">
            Doctor Preferences & Rules
          </h1>
          <p className="text-sm text-neutral-text-secondary">
            {doctor ? `${doctor.name} ${doctor.code ? `(${doctor.code})` : ''}` : 'Configure doctor-specific rules and preferences'}
          </p>
        </div>

        <div className="card">
          <div className="p-4 border-b border-neutral-border">
            <button
              onClick={() => {
                setEditingPreference(null);
                resetForm();
                setShowModal(true);
              }}
              className="btn-primary"
            >
              + Add Preference Rule
            </button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-neutral-border">
                  <th className="text-left p-4 text-sm font-semibold text-neutral-text-primary">Branch</th>
                  <th className="text-left p-4 text-sm font-semibold text-neutral-text-primary">Rule Type</th>
                  <th className="text-left p-4 text-sm font-semibold text-neutral-text-primary">Requirements</th>
                  <th className="text-left p-4 text-sm font-semibold text-neutral-text-primary">Status</th>
                  <th className="text-left p-4 text-sm font-semibold text-neutral-text-primary">Actions</th>
                </tr>
              </thead>
              <tbody>
                {preferences.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="p-8 text-center text-neutral-text-secondary">
                      No preference rules found. Click "Add Preference Rule" to create one.
                    </td>
                  </tr>
                ) : (
                  preferences.map((preference) => {
                    const branch = preference.branch_id ? branches.find(b => b.id === preference.branch_id) : null;
                    const requirements = (preference.rule_config.requirements || []) as any[];
                    
                    return (
                      <tr key={preference.id} className="border-b border-neutral-border hover:bg-neutral-hover">
                        <td className="p-4 text-sm text-neutral-text-primary">
                          {branch ? `${branch.code} - ${branch.name}` : 'All Branches'}
                        </td>
                        <td className="p-4 text-sm text-neutral-text-secondary capitalize">
                          {preference.rule_type.replace('_', ' ')}
                        </td>
                        <td className="p-4 text-sm text-neutral-text-secondary">
                          {requirements.length > 0 ? (
                            <div className="space-y-1">
                              {requirements.map((req: any, idx: number) => {
                                const position = positions.find(p => p.id === req.position_id);
                                return (
                                  <div key={idx} className="text-xs">
                                    {position?.name || 'Unknown'}: Min {req.min_count}
                                  </div>
                                );
                              })}
                            </div>
                          ) : (
                            '-'
                          )}
                        </td>
                        <td className="p-4">
                          <span className={`text-xs px-2 py-1 rounded ${
                            preference.is_active 
                              ? 'bg-green-100 text-green-800' 
                              : 'bg-gray-100 text-gray-800'
                          }`}>
                            {preference.is_active ? 'Active' : 'Inactive'}
                          </span>
                        </td>
                        <td className="p-4">
                          <div className="flex gap-2">
                            <button
                              onClick={() => handleEdit(preference)}
                              className="btn-secondary text-xs"
                            >
                              Edit
                            </button>
                            <button
                              onClick={() => handleDelete(preference.id)}
                              className="btn-secondary text-xs text-red-600 hover:text-red-700"
                            >
                              Delete
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-neutral-bg-secondary rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-semibold text-neutral-text-primary mb-4">
              {editingPreference ? 'Edit Preference Rule' : 'Add Preference Rule'}
            </h2>
            <form onSubmit={handleSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-neutral-text-primary mb-1">
                    Branch (leave empty for all branches)
                  </label>
                  <select
                    value={formData.branch_id || ''}
                    onChange={(e) => setFormData({ ...formData, branch_id: e.target.value || undefined })}
                    className="input-field"
                  >
                    <option value="">All Branches</option>
                    {branches.map((branch) => (
                      <option key={branch.id} value={branch.id}>
                        {branch.code} - {branch.name}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-neutral-text-primary mb-1">
                    Rule Type
                  </label>
                  <select
                    value={formData.rule_type}
                    onChange={(e) => setFormData({ ...formData, rule_type: e.target.value })}
                    className="input-field"
                  >
                    <option value="staff_requirement">Staff Requirement</option>
                    <option value="schedule_preference">Schedule Preference</option>
                  </select>
                </div>

                {formData.rule_type === 'staff_requirement' && (
                  <div>
                    <div className="flex items-center justify-between mb-2">
                      <label className="block text-sm font-medium text-neutral-text-primary">
                        Staff Requirements
                      </label>
                      <button
                        type="button"
                        onClick={addRequirement}
                        className="btn-secondary text-xs"
                      >
                        + Add Requirement
                      </button>
                    </div>
                    <div className="space-y-2">
                      {(formData.rule_config.requirements || []).map((req: any, index: number) => (
                        <div key={index} className="flex gap-2 items-center p-2 bg-neutral-bg-primary rounded">
                          <select
                            value={req.position_id || ''}
                            onChange={(e) => updateRequirement(index, 'position_id', e.target.value)}
                            className="input-field flex-1"
                            required
                          >
                            <option value="">Select Position</option>
                            {positions.map((position) => (
                              <option key={position.id} value={position.id}>
                                {position.name}
                              </option>
                            ))}
                          </select>
                          <input
                            type="number"
                            value={req.min_count || 1}
                            onChange={(e) => updateRequirement(index, 'min_count', parseInt(e.target.value) || 1)}
                            className="input-field w-24"
                            min="1"
                            placeholder="Min"
                            required
                          />
                          <button
                            type="button"
                            onClick={() => removeRequirement(index)}
                            className="btn-secondary text-xs text-red-600"
                          >
                            Remove
                          </button>
                        </div>
                      ))}
                      {(formData.rule_config.requirements || []).length === 0 && (
                        <p className="text-sm text-neutral-text-secondary text-center py-4">
                          No requirements added. Click "Add Requirement" to add one.
                        </p>
                      )}
                    </div>
                  </div>
                )}

                <div>
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={formData.is_active}
                      onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                      className="rounded"
                    />
                    <span className="text-sm font-medium text-neutral-text-primary">Active</span>
                  </label>
                </div>
              </div>
              <div className="flex gap-2 mt-6">
                <button type="submit" className="btn-primary flex-1">
                  {editingPreference ? 'Update' : 'Create'}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowModal(false);
                    setEditingPreference(null);
                    resetForm();
                  }}
                  className="btn-secondary flex-1"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </>
  );
}
