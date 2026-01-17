'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/contexts/UserContext';
import { zoneApi, Zone, CreateZoneRequest } from '@/lib/api/zone';
import { branchApi, Branch } from '@/lib/api/branch';

export default function ZoneConfigurationPage() {
  const router = useRouter();
  const { user, loading: userLoading } = useUser();
  const [zones, setZones] = useState<Zone[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [zoneBranchesMap, setZoneBranchesMap] = useState<Record<string, Branch[]>>({});
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingZone, setEditingZone] = useState<Zone | null>(null);
  const [selectedZone, setSelectedZone] = useState<Zone | null>(null);
  const [zoneBranches, setZoneBranches] = useState<Branch[]>([]);
  const [selectedBranchIds, setSelectedBranchIds] = useState<string[]>([]);
  const [loadingBranches, setLoadingBranches] = useState(false);
  const [savingBranches, setSavingBranches] = useState(false);

  const [formData, setFormData] = useState<CreateZoneRequest>({
    name: '',
    code: '',
    description: '',
    is_active: true,
  });

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
          await Promise.all([loadZones(), loadBranches()]);
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

  const loadZones = async () => {
    try {
      const zonesData = await zoneApi.list(true); // Include inactive
      setZones(zonesData || []);
      
      // Load branches for all zones in parallel
      const branchesMap: Record<string, Branch[]> = {};
      const branchPromises = zonesData.map(async (zone) => {
        try {
          const zoneBranchesData = await zoneApi.getBranches(zone.id);
          return { zoneId: zone.id, branches: zoneBranchesData || [] };
        } catch (error) {
          console.error(`Failed to load branches for zone ${zone.id}:`, error);
          return { zoneId: zone.id, branches: [] };
        }
      });
      
      const branchResults = await Promise.all(branchPromises);
      branchResults.forEach(({ zoneId, branches }) => {
        branchesMap[zoneId] = branches;
      });
      setZoneBranchesMap(branchesMap);
    } catch (error) {
      console.error('Failed to load zones:', error);
      setZones([]);
      setZoneBranchesMap({});
    }
  };

  const loadBranches = async () => {
    try {
      const branchesData = await branchApi.list();
      setBranches(branchesData || []);
    } catch (error) {
      console.error('Failed to load branches:', error);
      setBranches([]);
    }
  };

  const loadZoneBranches = async (zoneId: string) => {
    setLoadingBranches(true);
    try {
      const branchesData = await zoneApi.getBranches(zoneId);
      setZoneBranches(branchesData || []);
      setSelectedBranchIds(branchesData.map(b => b.id));
    } catch (error) {
      console.error('Failed to load zone branches:', error);
      setZoneBranches([]);
      setSelectedBranchIds([]);
    } finally {
      setLoadingBranches(false);
    }
  };

  const handleCreate = () => {
    setEditingZone(null);
    setFormData({
      name: '',
      code: '',
      description: '',
      is_active: true,
    });
    setShowModal(true);
  };

  const handleEdit = (zone: Zone) => {
    setEditingZone(zone);
    setFormData({
      name: zone.name,
      code: zone.code,
      description: zone.description || '',
      is_active: zone.is_active,
    });
    setShowModal(true);
  };

  const handleManageBranches = async (zone: Zone) => {
    setSelectedZone(zone);
    setSelectedBranchIds([]);
    await loadZoneBranches(zone.id);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingZone) {
        await zoneApi.update(editingZone.id, formData);
      } else {
        await zoneApi.create(formData);
      }
      setShowModal(false);
      setEditingZone(null);
      await loadZones();
      // Reload branches for selected zone if one is selected
      if (selectedZone) {
        await loadZoneBranches(selectedZone.id);
      }
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save zone');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this zone? This will remove all branch associations.')) {
      return;
    }

    try {
      await zoneApi.delete(id);
      await loadZones();
      if (selectedZone?.id === id) {
        setSelectedZone(null);
        setZoneBranches([]);
        setSelectedBranchIds([]);
      }
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete zone');
    }
  };

  const handleUpdateBranches = async () => {
    if (!selectedZone) return;

    setSavingBranches(true);
    try {
      await zoneApi.updateBranches(selectedZone.id, { branch_ids: selectedBranchIds });
      await loadZoneBranches(selectedZone.id);
      // Reload zones to update the branches display in the table
      await loadZones();
      alert('Zone branches updated successfully');
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to update zone branches');
    } finally {
      setSavingBranches(false);
    }
  };

  const handleBranchToggle = (branchId: string) => {
    setSelectedBranchIds(prev => {
      if (prev.includes(branchId)) {
        return prev.filter(id => id !== branchId);
      } else {
        return [...prev, branchId];
      }
    });
  };

  if (userLoading || loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-neutral-text-secondary">Loading...</div>
      </div>
    );
  }

  return (
    <>
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">Zone Configuration</h1>
          <p className="text-sm text-neutral-text-secondary">
            Manage zones in Bangkok and their associated branches
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Zones List */}
          <div className="card">
            <div className="p-4 border-b border-neutral-border">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-neutral-text-primary">Zones</h2>
                <button onClick={handleCreate} className="btn-primary">
                  Add Zone
                </button>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="table-salesforce">
                <thead>
                  <tr>
                    <th>Code</th>
                    <th>Name</th>
                    <th>Branches</th>
                    <th>Status</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {zones.length === 0 ? (
                    <tr>
                      <td colSpan={5} className="text-center py-8 text-neutral-text-secondary">
                        No zones found
                      </td>
                    </tr>
                  ) : (
                    zones.map((zone) => {
                      const zoneBranchesList = zoneBranchesMap[zone.id] || [];
                      return (
                        <tr key={zone.id}>
                          <td className="font-medium">{zone.code}</td>
                          <td>{zone.name}</td>
                          <td>
                            {zoneBranchesList.length > 0 ? (
                              <div className="flex flex-wrap gap-1">
                                {zoneBranchesList.slice(0, 3).map((branch) => (
                                  <span
                                    key={branch.id}
                                    className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800"
                                  >
                                    {branch.code}
                                  </span>
                                ))}
                                {zoneBranchesList.length > 3 && (
                                  <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-600">
                                    +{zoneBranchesList.length - 3} more
                                  </span>
                                )}
                              </div>
                            ) : (
                              <span className="text-xs text-neutral-text-secondary">No branches</span>
                            )}
                          </td>
                          <td>
                            <span className={`badge ${zone.is_active ? 'badge-primary' : 'badge-secondary'}`}>
                              {zone.is_active ? 'Active' : 'Inactive'}
                            </span>
                          </td>
                          <td>
                            <div className="flex gap-2">
                              <button
                                onClick={() => handleEdit(zone)}
                                className="text-salesforce-blue hover:text-salesforce-blue-hover text-sm"
                              >
                                Edit
                              </button>
                              <button
                                onClick={() => handleManageBranches(zone)}
                                className="text-salesforce-blue hover:text-salesforce-blue-hover text-sm"
                              >
                                Branches
                              </button>
                              <button
                                onClick={() => handleDelete(zone.id)}
                                className="text-red-600 hover:text-red-700 text-sm"
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

          {/* Branch Management */}
          <div className="card">
            <div className="p-4 border-b border-neutral-border">
              <h2 className="text-lg font-semibold text-neutral-text-primary">
                {selectedZone ? `Branches: ${selectedZone.name} (${selectedZone.code})` : 'Select a Zone'}
              </h2>
            </div>

            {selectedZone ? (
              <div className="p-4">
                {loadingBranches ? (
                  <div className="text-center py-8 text-neutral-text-secondary">Loading branches...</div>
                ) : (
                  <>
                    <div className="mb-4">
                      <p className="text-sm text-neutral-text-secondary mb-3">
                        Select branches to include in this zone. Changes are saved when you click "Update Branches".
                      </p>
                    </div>
                    <div className="space-y-2 max-h-96 overflow-y-auto border border-neutral-border rounded-md p-3 mb-4">
                      {branches.length === 0 ? (
                        <div className="text-sm text-neutral-text-secondary text-center py-4">
                          No branches available
                        </div>
                      ) : (
                        branches.map((branch) => {
                          const isSelected = selectedBranchIds.includes(branch.id);
                          return (
                            <div key={branch.id} className="flex items-center gap-3">
                              <input
                                type="checkbox"
                                id={`branch-${branch.id}`}
                                checked={isSelected}
                                onChange={() => handleBranchToggle(branch.id)}
                                className="w-4 h-4"
                              />
                              <label htmlFor={`branch-${branch.id}`} className="flex-1 cursor-pointer">
                                <span className="font-medium">{branch.name}</span>
                                <span className="text-xs text-neutral-text-secondary ml-2">({branch.code})</span>
                              </label>
                            </div>
                          );
                        })
                      )}
                    </div>
                    <div className="mt-4 pt-4 border-t border-neutral-border flex gap-2">
                      <button
                        onClick={() => {
                          setSelectedZone(null);
                          setZoneBranches([]);
                          setSelectedBranchIds([]);
                        }}
                        className="btn-secondary flex-1"
                      >
                        Close
                      </button>
                      <button
                        onClick={handleUpdateBranches}
                        disabled={savingBranches || selectedBranchIds.length === 0}
                        className="btn-primary flex-1"
                      >
                        {savingBranches ? 'Saving...' : 'Update Branches'}
                      </button>
                    </div>
                  </>
                )}
              </div>
            ) : (
              <div className="p-8 text-center text-neutral-text-secondary">
                Select a zone from the list to manage its branches
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Create/Edit Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="card max-w-md w-full">
            <div className="p-6">
              <h2 className="text-xl font-semibold text-neutral-text-primary mb-6">
                {editingZone ? 'Edit Zone' : 'Create Zone'}
              </h2>
              <form onSubmit={handleSubmit}>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Code *
                    </label>
                    <input
                      type="text"
                      required
                      value={formData.code}
                      onChange={(e) => setFormData({ ...formData, code: e.target.value.toUpperCase() })}
                      className="input-field"
                      placeholder="e.g., ZONE1"
                      maxLength={50}
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Name *
                    </label>
                    <input
                      type="text"
                      required
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      className="input-field"
                      placeholder="e.g., Central Zone"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Description
                    </label>
                    <textarea
                      value={formData.description}
                      onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                      className="input-field"
                      rows={3}
                      placeholder="Optional description"
                    />
                  </div>
                  <div>
                    <label className="flex items-center gap-2">
                      <input
                        type="checkbox"
                        checked={formData.is_active}
                        onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                        className="w-4 h-4"
                      />
                      <span className="text-sm font-medium text-neutral-text-primary">Active</span>
                    </label>
                  </div>
                </div>
                <div className="mt-6 flex justify-end gap-2">
                  <button
                    type="button"
                    onClick={() => {
                      setShowModal(false);
                      setEditingZone(null);
                    }}
                    className="btn-secondary"
                  >
                    Cancel
                  </button>
                  <button type="submit" className="btn-primary">
                    {editingZone ? 'Update' : 'Create'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
