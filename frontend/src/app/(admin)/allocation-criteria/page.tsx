'use client';

import { useState, useEffect } from 'react';
import { allocationCriteriaApi, AllocationCriteria, CreateAllocationCriteriaRequest } from '@/lib/api/allocation-criteria';

export default function AllocationCriteriaPage() {
  const [criteria, setCriteria] = useState<AllocationCriteria[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [formData, setFormData] = useState<CreateAllocationCriteriaRequest>({
    pillar: 'clinic_wide',
    type: 'revenue',
    weight: 0.5,
    is_active: true,
    description: '',
    config: '',
  });

  useEffect(() => {
    loadCriteria();
  }, []);

  const loadCriteria = async () => {
    try {
      setLoading(true);
      const data = await allocationCriteriaApi.list();
      setCriteria(data);
    } catch (error) {
      console.error('Failed to load criteria:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingId) {
        await allocationCriteriaApi.update(editingId, formData);
      } else {
        await allocationCriteriaApi.create(formData);
      }
      await loadCriteria();
      setShowForm(false);
      setEditingId(null);
      setFormData({
        pillar: 'clinic_wide',
        type: 'revenue',
        weight: 0.5,
        is_active: true,
        description: '',
        config: '',
      });
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save criteria');
    }
  };

  const handleEdit = (item: AllocationCriteria) => {
    setEditingId(item.id);
    setFormData({
      pillar: item.pillar,
      type: item.type,
      weight: item.weight,
      is_active: item.is_active,
      description: item.description || '',
      config: item.config || '',
    });
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this criteria?')) return;
    try {
      await allocationCriteriaApi.delete(id);
      await loadCriteria();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete criteria');
    }
  };

  const pillars = ['clinic_wide', 'doctor_specific', 'branch_specific'];
  const types = ['bookings', 'revenue', 'min_staff_position', 'min_staff_branch', 'doctor_count'];

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  return (
    <div className="w-full p-6">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-3xl font-bold">Allocation Criteria Configuration</h1>
        <button
          onClick={() => {
            setShowForm(true);
            setEditingId(null);
            setFormData({
              pillar: 'clinic_wide',
              type: 'revenue',
              weight: 0.5,
              is_active: true,
              description: '',
              config: '',
            });
          }}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          Add Criteria
        </button>
      </div>

      {showForm && (
        <div className="mb-6 p-4 bg-gray-50 rounded-lg">
          <h2 className="text-xl font-semibold mb-4">
            {editingId ? 'Edit Criteria' : 'New Criteria'}
          </h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">Pillar *</label>
                <select
                  value={formData.pillar}
                  onChange={(e) => setFormData({ ...formData, pillar: e.target.value as any })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  required
                >
                  {pillars.map((p) => (
                    <option key={p} value={p}>
                      {p.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Type *</label>
                <select
                  value={formData.type}
                  onChange={(e) => setFormData({ ...formData, type: e.target.value as any })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  required
                >
                  {types.map((t) => (
                    <option key={t} value={t}>
                      {t.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Weight (0.0 - 1.0) *</label>
                <input
                  type="number"
                  min="0"
                  max="1"
                  step="0.01"
                  value={formData.weight}
                  onChange={(e) => setFormData({ ...formData, weight: parseFloat(e.target.value) })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  required
                />
              </div>

              <div className="flex items-center">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={formData.is_active}
                    onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                    className="mr-2"
                  />
                  <span className="text-sm font-medium">Active</span>
                </label>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Description</label>
              <input
                type="text"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
              />
            </div>

            <div className="flex gap-3">
              <button
                type="submit"
                className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
              >
                {editingId ? 'Update' : 'Create'}
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowForm(false);
                  setEditingId(null);
                }}
                className="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-md"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="space-y-4">
        {pillars.map((pillar) => {
          const pillarCriteria = criteria.filter(c => c.pillar === pillar);
          return (
            <div key={pillar} className="border border-gray-300 rounded-lg p-4">
              <h3 className="text-lg font-semibold mb-3">
                {pillar.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())}
              </h3>
              {pillarCriteria.length === 0 ? (
                <p className="text-gray-500 text-sm">No criteria configured</p>
              ) : (
                <table className="w-full border-collapse">
                  <thead>
                    <tr className="bg-gray-100">
                      <th className="border border-gray-300 p-2 text-left">Type</th>
                      <th className="border border-gray-300 p-2 text-center">Weight</th>
                      <th className="border border-gray-300 p-2 text-center">Status</th>
                      <th className="border border-gray-300 p-2 text-left">Description</th>
                      <th className="border border-gray-300 p-2 text-center">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {pillarCriteria.map((item) => (
                      <tr key={item.id}>
                        <td className="border border-gray-300 p-2">
                          {item.type.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())}
                        </td>
                        <td className="border border-gray-300 p-2 text-center">
                          {item.weight.toFixed(2)}
                        </td>
                        <td className="border border-gray-300 p-2 text-center">
                          <span className={`px-2 py-1 rounded text-xs ${
                            item.is_active ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                          }`}>
                            {item.is_active ? 'Active' : 'Inactive'}
                          </span>
                        </td>
                        <td className="border border-gray-300 p-2 text-sm">
                          {item.description || '-'}
                        </td>
                        <td className="border border-gray-300 p-2 text-center">
                          <button
                            onClick={() => handleEdit(item)}
                            className="px-2 py-1 bg-blue-600 text-white text-xs rounded hover:bg-blue-700 mr-2"
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => handleDelete(item.id)}
                            className="px-2 py-1 bg-red-600 text-white text-xs rounded hover:bg-red-700"
                          >
                            Delete
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
