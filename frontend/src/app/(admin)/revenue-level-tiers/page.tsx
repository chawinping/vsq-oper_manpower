'use client';

import { useState, useEffect } from 'react';
import { revenueLevelTierApi, RevenueLevelTier, RevenueLevelTierCreate, RevenueLevelTierUpdate } from '@/lib/api/revenue-level-tier';

export default function RevenueLevelTiersPage() {
  const [tiers, setTiers] = useState<RevenueLevelTier[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [formData, setFormData] = useState<RevenueLevelTierCreate>({
    level_number: 1,
    level_name: '',
    min_revenue: 0,
    max_revenue: null,
    display_order: 0,
    color_code: '#CCCCCC',
    description: '',
  });

  useEffect(() => {
    loadTiers();
  }, []);

  const loadTiers = async () => {
    try {
      setLoading(true);
      const data = await revenueLevelTierApi.list();
      setTiers(data.sort((a, b) => a.display_order - b.display_order || a.level_number - b.level_number));
    } catch (error) {
      console.error('Failed to load tiers:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingId) {
        const updateData: RevenueLevelTierUpdate = {
          level_name: formData.level_name,
          min_revenue: formData.min_revenue,
          max_revenue: formData.max_revenue,
          display_order: formData.display_order,
          color_code: formData.color_code || null,
          description: formData.description || null,
        };
        await revenueLevelTierApi.update(editingId, updateData);
      } else {
        await revenueLevelTierApi.create(formData);
      }
      await loadTiers();
      resetForm();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save tier');
    }
  };

  const handleEdit = (tier: RevenueLevelTier) => {
    setEditingId(tier.id);
    setFormData({
      level_number: tier.level_number,
      level_name: tier.level_name,
      min_revenue: tier.min_revenue,
      max_revenue: tier.max_revenue,
      display_order: tier.display_order,
      color_code: tier.color_code || '#CCCCCC',
      description: tier.description || '',
    });
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this tier?')) return;
    try {
      await revenueLevelTierApi.delete(id);
      await loadTiers();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete tier');
    }
  };

  const resetForm = () => {
    setShowForm(false);
    setEditingId(null);
    setFormData({
      level_number: 1,
      level_name: '',
      min_revenue: 0,
      max_revenue: null,
      display_order: 0,
      color_code: '#CCCCCC',
      description: '',
    });
  };

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
        <h1 className="text-3xl font-bold">Revenue Level Tiers</h1>
        <button
          onClick={() => setShowForm(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          + Add Tier
        </button>
      </div>

      {showForm && (
        <div className="mb-6 p-6 bg-white rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">{editingId ? 'Edit Tier' : 'Create Tier'}</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">Level Number</label>
                <input
                  type="number"
                  min="1"
                  max="10"
                  value={formData.level_number}
                  onChange={(e) => setFormData({ ...formData, level_number: parseInt(e.target.value) })}
                  className="w-full px-3 py-2 border rounded-md"
                  required
                  disabled={!!editingId}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Level Name</label>
                <input
                  type="text"
                  value={formData.level_name}
                  onChange={(e) => setFormData({ ...formData, level_name: e.target.value })}
                  className="w-full px-3 py-2 border rounded-md"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Min Revenue (THB)</label>
                <input
                  type="number"
                  min="0"
                  step="1000"
                  value={formData.min_revenue}
                  onChange={(e) => setFormData({ ...formData, min_revenue: parseFloat(e.target.value) })}
                  className="w-full px-3 py-2 border rounded-md"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Max Revenue (THB) - Leave empty for no limit</label>
                <input
                  type="number"
                  min="0"
                  step="1000"
                  value={formData.max_revenue || ''}
                  onChange={(e) => setFormData({ ...formData, max_revenue: e.target.value ? parseFloat(e.target.value) : null })}
                  className="w-full px-3 py-2 border rounded-md"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Display Order</label>
                <input
                  type="number"
                  value={formData.display_order}
                  onChange={(e) => setFormData({ ...formData, display_order: parseInt(e.target.value) })}
                  className="w-full px-3 py-2 border rounded-md"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Color Code</label>
                <input
                  type="color"
                  value={formData.color_code || '#CCCCCC'}
                  onChange={(e) => setFormData({ ...formData, color_code: e.target.value })}
                  className="w-full h-10 border rounded-md"
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Description</label>
              <textarea
                value={formData.description || ''}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
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

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Level</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Revenue Range (THB)</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Color</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {tiers.map((tier) => (
              <tr key={tier.id}>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">{tier.level_number}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">{tier.level_name}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  {tier.min_revenue.toLocaleString()} - {tier.max_revenue ? tier.max_revenue.toLocaleString() : '∞'}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div
                    className="w-8 h-8 rounded border"
                    style={{ backgroundColor: tier.color_code || '#CCCCCC' }}
                  />
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  <button
                    onClick={() => handleEdit(tier)}
                    className="text-blue-600 hover:text-blue-800 mr-3"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => handleDelete(tier.id)}
                    className="text-red-600 hover:text-red-800"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
