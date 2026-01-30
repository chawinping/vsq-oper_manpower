'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/contexts/UserContext';
import { rotationStaffBranchPositionApi, RotationStaffBranchPosition, CreateRotationStaffBranchPositionRequest, UpdateRotationStaffBranchPositionRequest } from '@/lib/api/rotation-staff-branch-position';
import { staffApi, Staff } from '@/lib/api/staff';
import { positionApi, Position } from '@/lib/api/position';
import { handleApiError, showSuccess } from '@/lib/errors/errorHandler';

type ViewMode = 'staff' | 'position';

export default function RotationStaffPositionMappingPage() {
  const router = useRouter();
  const { user, loading: userLoading } = useUser();
  const [mappings, setMappings] = useState<RotationStaffBranchPosition[]>([]);
  const [rotationStaff, setRotationStaff] = useState<Staff[]>([]);
  const [branchPositions, setBranchPositions] = useState<Position[]>([]);
  const [loading, setLoading] = useState(true);
  const [viewMode, setViewMode] = useState<ViewMode>('staff');
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [formData, setFormData] = useState<CreateRotationStaffBranchPositionRequest>({
    rotation_staff_id: '',
    branch_position_id: '',
    substitution_level: 2,
    is_active: true,
  });
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedStaffFilter, setSelectedStaffFilter] = useState<string>('');

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
          await Promise.all([loadMappings(), loadRotationStaff(), loadBranchPositions()]);
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

  const loadMappings = async () => {
    try {
      // Load all mappings (no filter) so we can show all staff with or without mappings
      const data = await rotationStaffBranchPositionApi.list();
      setMappings(data || []);
    } catch (error) {
      console.error('Failed to load mappings:', error);
      setMappings([]);
    }
  };

  const loadRotationStaff = async () => {
    try {
      const allStaff = await staffApi.list();
      const rotationOnly = allStaff.filter(s => s.staff_type === 'rotation');
      setRotationStaff(rotationOnly || []);
    } catch (error) {
      console.error('Failed to load rotation staff:', error);
      setRotationStaff([]);
    }
  };

  const loadBranchPositions = async () => {
    try {
      const allPositions = await positionApi.list();
      const branchOnly = allPositions.filter(p => p.position_type === 'branch');
      setBranchPositions(branchOnly || []);
    } catch (error) {
      console.error('Failed to load branch positions:', error);
      setBranchPositions([]);
    }
  };

  // Remove the useEffect that reloads mappings on filter change
  // We want to show all mappings and filter in the UI instead

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingId) {
        const updateData: UpdateRotationStaffBranchPositionRequest = {
          substitution_level: formData.substitution_level,
          is_active: formData.is_active ?? true,
          notes: formData.notes || '',
        };
        await rotationStaffBranchPositionApi.update(editingId, updateData);
        showSuccess('Mapping updated successfully');
      } else {
        await rotationStaffBranchPositionApi.create(formData);
        showSuccess('Mapping created successfully');
      }
      await loadMappings();
      resetForm();
    } catch (error: any) {
      handleApiError(error, editingId ? 'Failed to update mapping' : 'Failed to create mapping');
    }
  };

  const handleEdit = (mapping: RotationStaffBranchPosition) => {
    setEditingId(mapping.id);
    setFormData({
      rotation_staff_id: mapping.rotation_staff_id,
      branch_position_id: mapping.branch_position_id,
      substitution_level: mapping.substitution_level,
      is_active: mapping.is_active,
      notes: mapping.notes || '',
    });
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this mapping? This action cannot be undone.')) return;
    try {
      await rotationStaffBranchPositionApi.delete(id);
      await loadMappings();
      showSuccess('Mapping deleted successfully');
    } catch (error: any) {
      handleApiError(error, 'Failed to delete mapping');
    }
  };

  const resetForm = () => {
    setShowForm(false);
    setEditingId(null);
    setFormData({
      rotation_staff_id: '',
      branch_position_id: '',
      substitution_level: 2,
      is_active: true,
      notes: '',
    });
  };

  const getSubstitutionLevelLabel = (level: number) => {
    switch (level) {
      case 1:
        return 'Preferred';
      case 2:
        return 'Acceptable';
      case 3:
        return 'Emergency Only';
      default:
        return 'Unknown';
    }
  };

  const getSubstitutionLevelColor = (level: number) => {
    switch (level) {
      case 1:
        return 'bg-green-100 text-green-800';
      case 2:
        return 'bg-yellow-100 text-yellow-800';
      case 3:
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  // Group mappings by staff for staff-centric view
  const staffMappingsMap = new Map<string, RotationStaffBranchPosition[]>();
  mappings.forEach(mapping => {
    const staffId = mapping.rotation_staff_id;
    if (!staffMappingsMap.has(staffId)) {
      staffMappingsMap.set(staffId, []);
    }
    staffMappingsMap.get(staffId)!.push(mapping);
  });

  // Ensure all rotation staff are in the map (even if they have no mappings)
  rotationStaff.forEach(staff => {
    if (!staffMappingsMap.has(staff.id)) {
      staffMappingsMap.set(staff.id, []);
    }
  });

  // Group mappings by position for position-centric view
  const positionMappingsMap = new Map<string, RotationStaffBranchPosition[]>();
  mappings.forEach(mapping => {
    const positionId = mapping.branch_position_id;
    if (!positionMappingsMap.has(positionId)) {
      positionMappingsMap.set(positionId, []);
    }
    positionMappingsMap.get(positionId)!.push(mapping);
  });

  // Filter staff based on search
  const filteredRotationStaff = rotationStaff.filter(staff =>
    staff.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    staff.nickname?.toLowerCase().includes(searchTerm.toLowerCase())
  );

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
        <h1 className="text-3xl font-bold">Rotation Staff Position Mapping</h1>
        <button
          onClick={() => setShowForm(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          + Add Mapping
        </button>
      </div>

      {/* View Mode Toggle */}
      <div className="mb-4 flex items-center gap-4">
        <span className="text-sm font-medium">View:</span>
        <button
          onClick={() => setViewMode('staff')}
          className={`px-4 py-2 rounded-md ${
            viewMode === 'staff'
              ? 'bg-blue-600 text-white'
              : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
          }`}
        >
          Staff View
        </button>
        <button
          onClick={() => setViewMode('position')}
          className={`px-4 py-2 rounded-md ${
            viewMode === 'position'
              ? 'bg-blue-600 text-white'
              : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
          }`}
        >
          Position View
        </button>
      </div>

      {/* Filters */}
      <div className="mb-4 flex gap-4">
        <input
          type="text"
          placeholder="Search staff..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="px-3 py-2 border rounded-md"
        />
        <select
          value={selectedStaffFilter}
          onChange={(e) => setSelectedStaffFilter(e.target.value)}
          className="px-3 py-2 border rounded-md"
        >
          <option value="">All Rotation Staff</option>
          {rotationStaff.map((staff) => (
            <option key={staff.id} value={staff.id}>
              {staff.name} {staff.nickname && `(${staff.nickname})`}
            </option>
          ))}
        </select>
      </div>

      {showForm && (
        <div className="mb-6 p-6 bg-white rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">
            {editingId ? 'Edit Mapping' : 'Create Mapping'}
          </h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Rotation Staff *</label>
              <select
                value={formData.rotation_staff_id}
                onChange={(e) => setFormData({ ...formData, rotation_staff_id: e.target.value })}
                className="w-full px-3 py-2 border rounded-md"
                required
                disabled={!!editingId}
              >
                <option value="">Select rotation staff...</option>
                {rotationStaff.map((staff) => (
                  <option key={staff.id} value={staff.id}>
                    {staff.name} {staff.nickname && `(${staff.nickname})`} - {staff.position?.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Branch Position *</label>
              <select
                value={formData.branch_position_id}
                onChange={(e) => setFormData({ ...formData, branch_position_id: e.target.value })}
                className="w-full px-3 py-2 border rounded-md"
                required
                disabled={!!editingId}
              >
                <option value="">Select branch position...</option>
                {branchPositions.map((position) => (
                  <option key={position.id} value={position.id}>
                    {position.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Substitution Level *</label>
              <select
                value={formData.substitution_level}
                onChange={(e) => setFormData({ ...formData, substitution_level: parseInt(e.target.value) })}
                className="w-full px-3 py-2 border rounded-md"
                required
              >
                <option value="1">Preferred</option>
                <option value="2">Acceptable</option>
                <option value="3">Emergency Only</option>
              </select>
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
            <div>
              <label className="block text-sm font-medium mb-1">Notes</label>
              <textarea
                value={formData.notes || ''}
                onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                className="w-full px-3 py-2 border rounded-md"
                rows={3}
              />
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

      {/* Staff-Centric View */}
      {viewMode === 'staff' && (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Rotation Staff
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Can Fill Positions
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {rotationStaff.length === 0 ? (
                <tr>
                  <td colSpan={3} className="px-6 py-4 text-center text-gray-500">
                    No rotation staff found
                  </td>
                </tr>
              ) : (
                rotationStaff
                  .filter(staff => {
                    // Filter by search term
                    if (searchTerm && !staff.name.toLowerCase().includes(searchTerm.toLowerCase()) && 
                        !staff.nickname?.toLowerCase().includes(searchTerm.toLowerCase())) {
                      return false;
                    }
                    
                    // Filter by selected staff
                    if (selectedStaffFilter && staff.id !== selectedStaffFilter) {
                      return false;
                    }
                    
                    return true;
                  })
                  .map((staff) => {
                    const staffMappings = staffMappingsMap.get(staff.id) || [];
                    
                    return (
                      <tr key={staff.id}>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="text-sm font-medium text-gray-900">{staff.name}</div>
                          <div className="text-sm text-gray-500">
                            {staff.nickname && `(${staff.nickname})`} - {staff.position?.name}
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          {staffMappings.length === 0 ? (
                            <span className="text-sm text-gray-400 italic">No positions mapped</span>
                          ) : (
                            <div className="space-y-1">
                              {staffMappings.map((mapping) => {
                                const position = branchPositions.find(p => p.id === mapping.branch_position_id);
                                return (
                                  <div key={mapping.id} className="flex items-center gap-2">
                                    <span className="text-sm">• {position?.name || mapping.branch_position_id}</span>
                                    <span className={`px-2 py-1 text-xs rounded-full ${getSubstitutionLevelColor(mapping.substitution_level)}`}>
                                      {getSubstitutionLevelLabel(mapping.substitution_level)}
                                    </span>
                                    {!mapping.is_active && (
                                      <span className="text-xs text-red-600">(Inactive)</span>
                                    )}
                                    <button
                                      onClick={() => handleEdit(mapping)}
                                      className="text-blue-600 hover:text-blue-900 text-xs"
                                    >
                                      Edit
                                    </button>
                                    <button
                                      onClick={() => handleDelete(mapping.id)}
                                      className="text-red-600 hover:text-red-900 text-xs"
                                    >
                                      Delete
                                    </button>
                                  </div>
                                );
                              })}
                            </div>
                          )}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                          <button
                            onClick={() => {
                              setFormData({
                                rotation_staff_id: staff.id,
                                branch_position_id: '',
                                substitution_level: 2,
                                is_active: true,
                                notes: '',
                              });
                              setShowForm(true);
                            }}
                            className="text-blue-600 hover:text-blue-900"
                          >
                            + Add Position
                          </button>
                        </td>
                      </tr>
                    );
                  })
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Position-Centric View */}
      {viewMode === 'position' && (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Branch Position
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Eligible Rotation Staff
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {positionMappingsMap.size === 0 ? (
                <tr>
                  <td colSpan={3} className="px-6 py-4 text-center text-gray-500">
                    No mappings found
                  </td>
                </tr>
              ) : (
                Array.from(positionMappingsMap.entries()).map(([positionId, positionMappings]) => {
                  const position = branchPositions.find(p => p.id === positionId);
                  if (!position) return null;

                  return (
                    <tr key={positionId}>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">{position.name}</div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="space-y-1">
                          {positionMappings.map((mapping) => {
                            const staff = rotationStaff.find(s => s.id === mapping.rotation_staff_id);
                            return (
                              <div key={mapping.id} className="flex items-center gap-2">
                                <span className="text-sm">
                                  • {staff?.name || mapping.rotation_staff_id}
                                  {staff?.nickname && ` (${staff.nickname})`}
                                </span>
                                <span className={`px-2 py-1 text-xs rounded-full ${getSubstitutionLevelColor(mapping.substitution_level)}`}>
                                  {getSubstitutionLevelLabel(mapping.substitution_level)}
                                </span>
                                {!mapping.is_active && (
                                  <span className="text-xs text-red-600">(Inactive)</span>
                                )}
                                <button
                                  onClick={() => handleEdit(mapping)}
                                  className="text-blue-600 hover:text-blue-900 text-xs"
                                >
                                  Edit
                                </button>
                                <button
                                  onClick={() => handleDelete(mapping.id)}
                                  className="text-red-600 hover:text-red-900 text-xs"
                                >
                                  Delete
                                </button>
                              </div>
                            );
                          })}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                        <button
                          onClick={() => {
                            setFormData({
                              rotation_staff_id: '',
                              branch_position_id: positionId,
                              substitution_level: 2,
                              is_active: true,
                              notes: '',
                            });
                            setShowForm(true);
                          }}
                          className="text-blue-600 hover:text-blue-900"
                        >
                          + Add Staff
                        </button>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
