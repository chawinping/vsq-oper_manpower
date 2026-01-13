'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/contexts/UserContext';
import { positionApi, Position, UpdatePositionRequest } from '@/lib/api/position';

export default function PositionsPage() {
  const router = useRouter();
  const { user, loading: userLoading } = useUser();
  const [positions, setPositions] = useState<Position[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editData, setEditData] = useState<UpdatePositionRequest>({
    name: '',
    min_staff_per_branch: 1,
    revenue_multiplier: 0,
    display_order: 999,
  });
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    // Check if user has permission
    if (!userLoading && user && user.role !== 'admin') {
      router.push('/dashboard');
      return;
    }
  }, [user, userLoading, router]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        if (user?.role === 'admin') {
          await loadPositions();
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

  const loadPositions = async () => {
    try {
      const positionsData = await positionApi.list();
      setPositions(positionsData || []);
    } catch (error) {
      console.error('Failed to load positions:', error);
      setPositions([]);
    }
  };

  const handleEdit = (position: Position) => {
    setEditingId(position.id);
    setEditData({
      name: position.name,
      min_staff_per_branch: position.min_staff_per_branch,
      revenue_multiplier: position.revenue_multiplier,
      display_order: position.display_order,
    });
  };

  const handleSave = async () => {
    if (!editingId) return;

    setSaving(true);
    try {
      await positionApi.update(editingId, editData);
      setEditingId(null);
      await loadPositions();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to update position');
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    setEditingId(null);
    setEditData({
      name: '',
      min_staff_per_branch: 1,
      revenue_multiplier: 0,
      display_order: 999,
    });
  };

  if (userLoading || loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-neutral-text-secondary">Loading...</div>
      </div>
    );
  }

  if (user?.role !== 'admin') {
    return null;
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">Position Management</h1>
        <p className="text-sm text-neutral-text-secondary">
          Manage positions and their display order. Lower numbers appear first in staff tables.
        </p>
      </div>

      <div className="card">
        <div className="overflow-x-auto">
          <table className="table-salesforce">
            <thead>
              <tr>
                <th>Display Order</th>
                <th>Name</th>
                <th>Min Staff/Branch</th>
                <th>Revenue Multiplier</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {(positions || []).map((position) => (
                <tr key={position.id}>
                  <td>
                    {editingId === position.id ? (
                      <input
                        type="number"
                        value={editData.display_order}
                        onChange={(e) => setEditData({ ...editData, display_order: parseInt(e.target.value) || 999 })}
                        className="input-field w-24"
                        min="1"
                      />
                    ) : (
                      <span className="font-medium">{position.display_order}</span>
                    )}
                  </td>
                  <td className="font-medium">
                    {editingId === position.id ? (
                      <input
                        type="text"
                        value={editData.name}
                        onChange={(e) => setEditData({ ...editData, name: e.target.value })}
                        className="input-field"
                      />
                    ) : (
                      position.name
                    )}
                  </td>
                  <td>
                    {editingId === position.id ? (
                      <input
                        type="number"
                        value={editData.min_staff_per_branch}
                        onChange={(e) => setEditData({ ...editData, min_staff_per_branch: parseInt(e.target.value) || 0 })}
                        className="input-field w-24"
                        min="0"
                      />
                    ) : (
                      position.min_staff_per_branch
                    )}
                  </td>
                  <td>
                    {editingId === position.id ? (
                      <input
                        type="number"
                        step="0.1"
                        value={editData.revenue_multiplier}
                        onChange={(e) => setEditData({ ...editData, revenue_multiplier: parseFloat(e.target.value) || 0 })}
                        className="input-field w-24"
                        min="0"
                      />
                    ) : (
                      position.revenue_multiplier.toFixed(2)
                    )}
                  </td>
                  <td>
                    {editingId === position.id ? (
                      <div className="flex gap-3">
                        <button
                          onClick={handleSave}
                          disabled={saving}
                          className="text-green-600 hover:text-green-700 text-sm disabled:opacity-50"
                        >
                          Save
                        </button>
                        <button
                          onClick={handleCancel}
                          className="text-neutral-text-secondary hover:text-neutral-text-primary text-sm"
                        >
                          Cancel
                        </button>
                      </div>
                    ) : (
                      <button
                        onClick={() => handleEdit(position)}
                        className="text-salesforce-blue hover:text-salesforce-blue-hover text-sm"
                      >
                        Edit
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {positions.length === 0 && (
            <div className="text-center py-12 text-neutral-text-secondary">
              No positions found
            </div>
          )}
        </div>
      </div>

      <div className="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-md">
        <h3 className="text-sm font-semibold text-blue-900 mb-2">About Display Order</h3>
        <ul className="text-xs text-blue-800 space-y-1">
          <li>• Lower numbers appear first in staff tables</li>
          <li>• Branch Manager should typically have display_order = 1</li>
          <li>• Staff are sorted by position display_order, then by name</li>
          <li>• You can use any positive integer for display_order</li>
        </ul>
      </div>
    </div>
  );
}


