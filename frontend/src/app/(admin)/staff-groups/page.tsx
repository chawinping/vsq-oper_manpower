'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/contexts/UserContext';
import { staffGroupApi, StaffGroup, CreateStaffGroupRequest, UpdateStaffGroupRequest } from '@/lib/api/staff-group';
import { positionApi, Position } from '@/lib/api/position';
import { handleApiError, showSuccess } from '@/lib/errors/errorHandler';

export default function StaffGroupsPage() {
  const router = useRouter();
  const { user, loading: userLoading } = useUser();
  const [staffGroups, setStaffGroups] = useState<StaffGroup[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [formData, setFormData] = useState<CreateStaffGroupRequest>({
    name: '',
    description: '',
    is_active: true,
  });
  const [selectedGroup, setSelectedGroup] = useState<StaffGroup | null>(null);
  const [showPositionModal, setShowPositionModal] = useState(false);

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
          await Promise.all([loadStaffGroups(), loadPositions()]);
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

  const loadStaffGroups = async () => {
    try {
      const data = await staffGroupApi.list();
      setStaffGroups(data || []);
    } catch (error) {
      console.error('Failed to load staff groups:', error);
      setStaffGroups([]);
    }
  };

  const loadPositions = async () => {
    try {
      const data = await positionApi.list();
      setPositions(data || []);
    } catch (error) {
      console.error('Failed to load positions:', error);
      setPositions([]);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingId) {
        const updateData: UpdateStaffGroupRequest = {
          name: formData.name,
          description: formData.description,
          is_active: formData.is_active ?? true,
        };
        await staffGroupApi.update(editingId, updateData);
      } else {
        await staffGroupApi.create(formData);
      }
      await loadStaffGroups();
      resetForm();
      showSuccess(editingId ? 'Staff group updated successfully' : 'Staff group created successfully');
    } catch (error: any) {
      handleApiError(error, editingId ? 'Failed to update staff group' : 'Failed to create staff group');
    }
  };

  const handleEdit = (staffGroup: StaffGroup) => {
    setEditingId(staffGroup.id);
    setFormData({
      name: staffGroup.name,
      description: staffGroup.description || '',
      is_active: staffGroup.is_active,
    });
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this staff group? This action cannot be undone.')) return;
    try {
      await staffGroupApi.delete(id);
      await loadStaffGroups();
      showSuccess('Staff group deleted successfully');
    } catch (error: any) {
      handleApiError(error, 'Failed to delete staff group');
    }
  };

  const handleManagePositions = (staffGroup: StaffGroup) => {
    setSelectedGroup(staffGroup);
    setShowPositionModal(true);
  };

  const handleAddPosition = async (positionId: string) => {
    if (!selectedGroup) return;
    try {
      await staffGroupApi.addPosition(selectedGroup.id, positionId);
      await loadStaffGroups();
      // Refresh selected group
      const updated = await staffGroupApi.getById(selectedGroup.id);
      setSelectedGroup(updated);
      showSuccess('Position added successfully');
    } catch (error: any) {
      handleApiError(error, 'Failed to add position');
    }
  };

  const handleRemovePosition = async (positionId: string) => {
    if (!selectedGroup) return;
    try {
      await staffGroupApi.removePosition(selectedGroup.id, positionId);
      await loadStaffGroups();
      // Refresh selected group
      const updated = await staffGroupApi.getById(selectedGroup.id);
      setSelectedGroup(updated);
      showSuccess('Position removed successfully');
    } catch (error: any) {
      handleApiError(error, 'Failed to remove position');
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

  const getAvailablePositions = () => {
    if (!selectedGroup) return [];
    const groupPositionIds = selectedGroup.positions?.map(p => p.position_id) || [];
    return positions.filter(p => !groupPositionIds.includes(p.id));
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
        <h1 className="text-3xl font-bold">Staff Groups</h1>
        <button
          onClick={() => setShowForm(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          + Add Staff Group
        </button>
      </div>

      {showForm && (
        <div className="mb-6 p-6 bg-white rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">{editingId ? 'Edit Staff Group' : 'Create Staff Group'}</h2>
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
                Positions
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
            {staffGroups.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-6 py-4 text-center text-gray-500">
                  No staff groups found
                </td>
              </tr>
            ) : (
              staffGroups.map((group) => (
                <tr key={group.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    {group.name}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500">
                    {group.description || '-'}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500">
                    {group.positions?.length || 0} position(s)
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                        group.is_active
                          ? 'bg-green-100 text-green-800'
                          : 'bg-red-100 text-red-800'
                      }`}
                    >
                      {group.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <button
                      onClick={() => handleManagePositions(group)}
                      className="text-green-600 hover:text-green-900 mr-4"
                    >
                      Manage Positions
                    </button>
                    <button
                      onClick={() => handleEdit(group)}
                      className="text-blue-600 hover:text-blue-900 mr-4"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => handleDelete(group.id)}
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

      {showPositionModal && selectedGroup && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-2xl w-full max-h-[80vh] overflow-y-auto">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-semibold">Manage Positions: {selectedGroup.name}</h2>
              <button
                onClick={() => {
                  setShowPositionModal(false);
                  setSelectedGroup(null);
                }}
                className="text-gray-500 hover:text-gray-700"
              >
                ✕
              </button>
            </div>

            <div className="mb-4">
              <h3 className="font-medium mb-2">Current Positions:</h3>
              {selectedGroup.positions && selectedGroup.positions.length > 0 ? (
                <div className="space-y-2">
                  {selectedGroup.positions.map((sgp) => {
                    const position = positions.find(p => p.id === sgp.position_id);
                    return (
                      <div key={sgp.id} className="flex items-center justify-between p-2 bg-gray-50 rounded">
                        <span>{position?.name || sgp.position_id}</span>
                        <button
                          onClick={() => handleRemovePosition(sgp.position_id)}
                          className="text-red-600 hover:text-red-800 text-sm"
                        >
                          Remove
                        </button>
                      </div>
                    );
                  })}
                </div>
              ) : (
                <p className="text-gray-500 text-sm">No positions assigned</p>
              )}
            </div>

            <div>
              <h3 className="font-medium mb-2">Add Position:</h3>
              <select
                onChange={(e) => {
                  if (e.target.value) {
                    handleAddPosition(e.target.value);
                    e.target.value = '';
                  }
                }}
                className="w-full px-3 py-2 border rounded-md"
              >
                <option value="">Select a position...</option>
                {getAvailablePositions().map((position) => (
                  <option key={position.id} value={position.id}>
                    {position.name}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
