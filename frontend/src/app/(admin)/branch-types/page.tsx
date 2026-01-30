'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/contexts/UserContext';
import { branchTypeApi, BranchType, CreateBranchTypeRequest, UpdateBranchTypeRequest, BranchTypeConstraints, ConstraintsUpdate, StaffGroupRequirement } from '@/lib/api/branch-type';
import { staffGroupApi, StaffGroup } from '@/lib/api/staff-group';
import { handleApiError, showSuccess } from '@/lib/errors/errorHandler';

const DAY_NAMES = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

export default function BranchTypesPage() {
  const router = useRouter();
  const { user, loading: userLoading } = useUser();
  const [branchTypes, setBranchTypes] = useState<BranchType[]>([]);
  const [staffGroups, setStaffGroups] = useState<StaffGroup[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [formData, setFormData] = useState<CreateBranchTypeRequest>({
    name: '',
    description: '',
    is_active: true,
  });
  const [showConstraintsModal, setShowConstraintsModal] = useState(false);
  const [selectedBranchType, setSelectedBranchType] = useState<BranchType | null>(null);
  const [constraints, setConstraints] = useState<Map<number, BranchTypeConstraints>>(new Map()); // key: dayOfWeek
  const [savingConstraints, setSavingConstraints] = useState(false);

  useEffect(() => {
    if (!userLoading && user && user.role !== 'admin') {
      router.push('/dashboard');
      return;
    }
  }, [user, userLoading, router]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        if (user?.role === 'admin') {
          await Promise.all([loadBranchTypes(), loadStaffGroups()]);
        }
      } catch (error: any) {
        console.error('Failed to fetch data:', error);
      } finally {
        setLoading(false);
      }
    };

    if (user && user.role === 'admin') {
      fetchData();
    }
  }, [user]);

  const loadBranchTypes = async () => {
    try {
      const data = await branchTypeApi.list();
      setBranchTypes(data || []);
    } catch (error) {
      console.error('Failed to load branch types:', error);
      setBranchTypes([]);
    }
  };

  const loadStaffGroups = async () => {
    try {
      const data = await staffGroupApi.list();
      // Filter to only active staff groups
      setStaffGroups((data || []).filter(group => group.is_active));
    } catch (error) {
      console.error('Failed to load staff groups:', error);
      setStaffGroups([]);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingId) {
        const updateData: UpdateBranchTypeRequest = {
          name: formData.name,
          description: formData.description,
          is_active: formData.is_active ?? true,
        };
        await branchTypeApi.update(editingId, updateData);
      } else {
        await branchTypeApi.create(formData);
      }
      await loadBranchTypes();
      resetForm();
      showSuccess(editingId ? 'Branch type updated successfully' : 'Branch type created successfully');
    } catch (error: any) {
      handleApiError(error, editingId ? 'Failed to update branch type' : 'Failed to create branch type');
    }
  };

  const handleEdit = (branchType: BranchType) => {
    setEditingId(branchType.id);
    setFormData({
      name: branchType.name,
      description: branchType.description || '',
      is_active: branchType.is_active,
    });
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this branch type? This action cannot be undone.')) return;
    try {
      await branchTypeApi.delete(id);
      await loadBranchTypes();
      showSuccess('Branch type deleted successfully');
    } catch (error: any) {
      handleApiError(error, 'Failed to delete branch type');
    }
  };

  const resetForm = () => {
    setShowForm(false);
    setEditingId(null);
    setFormData({
      name: '',
      description: '',
      is_active: true,
    });
  };

  const handleManageConstraints = async (branchType: BranchType) => {
    setSelectedBranchType(branchType);
    setShowConstraintsModal(true);
    
    try {
      // Load existing constraints
      const existingConstraints = await branchTypeApi.getConstraints(branchType.id);
      const constraintsMap = new Map<number, BranchTypeConstraints>();
      
      // Create map of existing constraints: key = dayOfWeek
      existingConstraints.forEach((constraint) => {
        constraintsMap.set(constraint.day_of_week, constraint);
      });
      
      // Initialize constraints for all days (0-6) if they don't exist
      for (let day = 0; day < 7; day++) {
        if (!constraintsMap.has(day)) {
          constraintsMap.set(day, {
            id: '',
            branch_type_id: branchType.id,
            day_of_week: day,
            created_at: '',
            updated_at: '',
            staff_group_requirements: [],
          });
        } else {
          // Ensure staff_group_requirements is initialized
          const constraint = constraintsMap.get(day);
          if (constraint && !constraint.staff_group_requirements) {
            constraint.staff_group_requirements = [];
          }
        }
      }
      
      setConstraints(constraintsMap);
    } catch (error) {
      handleApiError(error, 'Failed to load constraints');
      setShowConstraintsModal(false);
    }
  };

  const handleConstraintChange = (dayOfWeek: number, staffGroupId: string, value: number) => {
    const constraint = constraints.get(dayOfWeek);
    if (!constraint) return;
    
    const updatedConstraint = { ...constraint };
    if (!updatedConstraint.staff_group_requirements) {
      updatedConstraint.staff_group_requirements = [];
    }
    
    // Find existing requirement for this staff group
    const existingIndex = updatedConstraint.staff_group_requirements.findIndex(
      req => req.staff_group_id === staffGroupId
    );
    
    if (existingIndex >= 0) {
      // Update existing requirement
      updatedConstraint.staff_group_requirements[existingIndex] = {
        ...updatedConstraint.staff_group_requirements[existingIndex],
        minimum_count: value,
      };
    } else {
      // Add new requirement
      updatedConstraint.staff_group_requirements.push({
        id: '',
        branch_type_constraint_id: constraint.id,
        staff_group_id: staffGroupId,
        minimum_count: value,
        created_at: '',
        updated_at: '',
      });
    }
    
    const updatedConstraints = new Map(constraints);
    updatedConstraints.set(dayOfWeek, updatedConstraint);
    setConstraints(updatedConstraints);
  };

  const getConstraintValue = (dayOfWeek: number, staffGroupId: string): number => {
    const constraint = constraints.get(dayOfWeek);
    if (!constraint || !constraint.staff_group_requirements) {
      return 0;
    }
    
    const requirement = constraint.staff_group_requirements.find(
      req => req.staff_group_id === staffGroupId
    );
    
    return requirement?.minimum_count || 0;
  };

  const handleSaveConstraints = async () => {
    console.log('[handleSaveConstraints] Called');
    
    if (!selectedBranchType) {
      console.error('[handleSaveConstraints] No selectedBranchType');
      return;
    }
    
    console.log('[handleSaveConstraints] Constraints map size:', constraints.size);
    console.log('[handleSaveConstraints] Constraints map contents:', Array.from(constraints.entries()));
    
    if (constraints.size === 0) {
      console.error('[handleSaveConstraints] Constraints map is empty');
      handleApiError(new Error('No constraints to save. Please wait for constraints to load.'), 'No constraints available');
      return;
    }
    
    setSavingConstraints(true);
    try {
      const constraintsToUpdate: ConstraintsUpdate[] = Array.from(constraints.values())
        .filter((constraint) => constraint !== undefined && constraint !== null)
        .map((constraint) => {
          const staffGroupRequirements: StaffGroupRequirement[] = 
            (constraint.staff_group_requirements || []).map(req => ({
              staff_group_id: req.staff_group_id,
              minimum_count: req.minimum_count,
            }));
          
          return {
            day_of_week: constraint.day_of_week,
            staff_group_requirements: staffGroupRequirements,
          };
        });
      
      console.log('[handleSaveConstraints] Constraints to update:', constraintsToUpdate);
      
      if (constraintsToUpdate.length === 0) {
        console.error('[handleSaveConstraints] No valid constraints after filtering');
        handleApiError(new Error('No valid constraints to save.'), 'No valid constraints');
        setSavingConstraints(false);
        return;
      }
      
      console.log('[handleSaveConstraints] Calling API with:', { 
        branchTypeId: selectedBranchType.id, 
        constraintsCount: constraintsToUpdate.length,
        constraints: constraintsToUpdate 
      });
      
      const result = await branchTypeApi.updateConstraints(selectedBranchType.id, constraintsToUpdate);
      
      console.log('[handleSaveConstraints] API call successful, result:', result);
      
      showSuccess('Constraints saved successfully!');
      setShowConstraintsModal(false);
      setSelectedBranchType(null);
      setConstraints(new Map());
    } catch (error: any) {
      console.error('[handleSaveConstraints] Error caught:', error);
      console.error('[handleSaveConstraints] Error type:', typeof error);
      console.error('[handleSaveConstraints] Error keys:', Object.keys(error || {}));
      handleApiError(error, 'Failed to save constraints');
    } finally {
      setSavingConstraints(false);
    }
  };

  if (loading || userLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  return (
    <div className="w-full p-6">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-3xl font-bold">Branch Types</h1>
        <button
          onClick={() => setShowForm(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          + Add Branch Type
        </button>
      </div>

      {showForm && (
        <div className="mb-6 p-6 bg-white rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">{editingId ? 'Edit Branch Type' : 'Create Branch Type'}</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Name *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="w-full px-3 py-2 border rounded-md"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Description</label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                className="w-full px-3 py-2 border rounded-md"
                rows={3}
              />
            </div>
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
            <div className="flex gap-2">
              <button
                type="submit"
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                {editingId ? 'Update' : 'Create'}
              </button>
              <button
                type="button"
                onClick={resetForm}
                className="px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Name
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Description
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {branchTypes.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-6 py-4 text-center text-gray-500">
                  No branch types found
                </td>
              </tr>
            ) : (
              branchTypes.map((branchType) => (
                <tr key={branchType.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    {branchType.name}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500">
                    {branchType.description || '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                        branchType.is_active
                          ? 'bg-green-100 text-green-800'
                          : 'bg-red-100 text-red-800'
                      }`}
                    >
                      {branchType.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <button
                      onClick={() => handleManageConstraints(branchType)}
                      className="text-green-600 hover:text-green-900 mr-4"
                    >
                      Manage Constraints
                    </button>
                    <button
                      onClick={() => handleEdit(branchType)}
                      className="text-blue-600 hover:text-blue-900 mr-4"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => handleDelete(branchType.id)}
                      className="text-red-600 hover:text-red-900"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Constraints Modal */}
      {showConstraintsModal && selectedBranchType && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-4xl w-full max-h-[90vh] overflow-y-auto">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-semibold">Daily Staff Constraints: {selectedBranchType.name}</h2>
              <button
                onClick={() => {
                  setShowConstraintsModal(false);
                  setSelectedBranchType(null);
                  setConstraints(new Map());
                }}
                className="text-gray-500 hover:text-gray-700"
              >
                ✕
              </button>
            </div>

            <div className="mb-4">
              <p className="text-sm text-gray-600">
                Set minimum staff constraints per day. These constraints will be inherited by all branches assigned to this branch type.
                Branches can override these constraints if needed.
              </p>
            </div>

            {staffGroups.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <p>No active staff groups found. Please create staff groups in Staff Groups settings first.</p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50 z-10">
                        Day
                      </th>
                      {staffGroups.map((staffGroup) => (
                        <th
                          key={staffGroup.id}
                          className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider min-w-[120px]"
                        >
                          Min {staffGroup.name}
                          {staffGroup.description && (
                            <span className="block text-xs font-normal text-gray-400 mt-1">
                              ({staffGroup.description})
                            </span>
                          )}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {Array.from({ length: 7 }, (_, i) => i).map((dayOfWeek) => {
                      const constraint = constraints.get(dayOfWeek);
                      if (!constraint) return null;

                      return (
                        <tr key={dayOfWeek}>
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 sticky left-0 bg-white z-10">
                            {DAY_NAMES[dayOfWeek]}
                          </td>
                          {staffGroups.map((staffGroup) => (
                            <td key={staffGroup.id} className="px-6 py-4 whitespace-nowrap">
                              <input
                                type="number"
                                min="0"
                                step="1"
                                value={getConstraintValue(dayOfWeek, staffGroup.id)}
                                onChange={(e) =>
                                  handleConstraintChange(dayOfWeek, staffGroup.id, parseInt(e.target.value) || 0)
                                }
                                className="w-24 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                              />
                            </td>
                          ))}
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}

            <div className="mt-6 flex gap-2">
              <button
                onClick={handleSaveConstraints}
                disabled={savingConstraints}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                {savingConstraints ? 'Saving...' : 'Save Constraints'}
              </button>
              <button
                onClick={() => {
                  setShowConstraintsModal(false);
                  setSelectedBranchType(null);
                  setConstraints(new Map());
                }}
                className="px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400"
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
