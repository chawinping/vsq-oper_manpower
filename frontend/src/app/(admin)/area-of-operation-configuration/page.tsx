'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/contexts/UserContext';
import { areaOfOperationApi, AreaOfOperation, CreateAreaOfOperationRequest } from '@/lib/api/areaOfOperation';
import { zoneApi, Zone } from '@/lib/api/zone';
import { branchApi, Branch } from '@/lib/api/branch';

export default function AreaOfOperationConfigurationPage() {
  const router = useRouter();
  const { user, loading: userLoading } = useUser();
  const [areas, setAreas] = useState<AreaOfOperation[]>([]);
  const [zones, setZones] = useState<Zone[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingArea, setEditingArea] = useState<AreaOfOperation | null>(null);
  const [selectedArea, setSelectedArea] = useState<AreaOfOperation | null>(null);
  const [areaZones, setAreaZones] = useState<Zone[]>([]);
  const [areaBranches, setAreaBranches] = useState<Branch[]>([]);
  const [allAreaBranches, setAllAreaBranches] = useState<Branch[]>([]);
  const [loadingData, setLoadingData] = useState(false);
  const [activeTab, setActiveTab] = useState<'zones' | 'branches' | 'all'>('zones');

  const [formData, setFormData] = useState<CreateAreaOfOperationRequest>({
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
          await Promise.all([loadAreas(), loadZones(), loadBranches()]);
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

  const loadAreas = async () => {
    try {
      const areasData = await areaOfOperationApi.list(true); // Include inactive
      setAreas(areasData || []);
    } catch (error) {
      console.error('Failed to load areas:', error);
      setAreas([]);
    }
  };

  const loadZones = async () => {
    try {
      const zonesData = await zoneApi.list(true); // Include inactive
      setZones(zonesData || []);
    } catch (error) {
      console.error('Failed to load zones:', error);
      setZones([]);
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

  const loadAreaData = async (areaId: string) => {
    setLoadingData(true);
    try {
      // Load zones, individual branches, and all branches
      const [zonesData, branchesData, allBranchesData] = await Promise.all([
        areaOfOperationApi.getZones(areaId),
        areaOfOperationApi.getBranches(areaId),
        areaOfOperationApi.getAllBranches(areaId),
      ]);
      setAreaZones(zonesData as Zone[]);
      setAreaBranches(branchesData as Branch[]);
      setAllAreaBranches(allBranchesData as Branch[]);
    } catch (error) {
      console.error('Failed to load area data:', error);
      setAreaZones([]);
      setAreaBranches([]);
      setAllAreaBranches([]);
    } finally {
      setLoadingData(false);
    }
  };

  const handleCreate = () => {
    setEditingArea(null);
    setFormData({
      name: '',
      code: '',
      description: '',
      is_active: true,
    });
    setShowModal(true);
  };

  const handleEdit = (area: AreaOfOperation) => {
    setEditingArea(area);
    setFormData({
      name: area.name,
      code: area.code,
      description: area.description || '',
      is_active: area.is_active,
    });
    setShowModal(true);
  };

  const handleManage = async (area: AreaOfOperation) => {
    setSelectedArea(area);
    setActiveTab('zones');
    await loadAreaData(area.id);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingArea) {
        await areaOfOperationApi.update(editingArea.id, formData);
      } else {
        await areaOfOperationApi.create(formData);
      }
      setShowModal(false);
      setEditingArea(null);
      await loadAreas();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save area of operation');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this area of operation? This will remove all zone and branch associations.')) {
      return;
    }

    try {
      await areaOfOperationApi.delete(id);
      await loadAreas();
      if (selectedArea?.id === id) {
        setSelectedArea(null);
        setAreaZones([]);
        setAreaBranches([]);
        setAllAreaBranches([]);
      }
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete area of operation');
    }
  };

  const handleAddZone = async (zoneId: string) => {
    if (!selectedArea) return;
    try {
      await areaOfOperationApi.addZone(selectedArea.id, zoneId);
      await loadAreaData(selectedArea.id);
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to add zone');
    }
  };

  const handleRemoveZone = async (zoneId: string) => {
    if (!selectedArea) return;
    try {
      await areaOfOperationApi.removeZone(selectedArea.id, zoneId);
      await loadAreaData(selectedArea.id);
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to remove zone');
    }
  };

  const handleAddBranch = async (branchId: string) => {
    if (!selectedArea) return;
    try {
      await areaOfOperationApi.addBranch(selectedArea.id, branchId);
      await loadAreaData(selectedArea.id);
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to add branch');
    }
  };

  const handleRemoveBranch = async (branchId: string) => {
    if (!selectedArea) return;
    try {
      await areaOfOperationApi.removeBranch(selectedArea.id, branchId);
      await loadAreaData(selectedArea.id);
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to remove branch');
    }
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
          <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">Area of Operation Configuration</h1>
          <p className="text-sm text-neutral-text-secondary">
            Manage areas of operation, their zones, and individual branches
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Areas List */}
          <div className="card">
            <div className="p-4 border-b border-neutral-border">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-neutral-text-primary">Areas of Operation</h2>
                <button onClick={handleCreate} className="btn-primary">
                  Add Area
                </button>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="table-salesforce">
                <thead>
                  <tr>
                    <th>Code</th>
                    <th>Name</th>
                    <th>Status</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {areas.length === 0 ? (
                    <tr>
                      <td colSpan={4} className="text-center py-8 text-neutral-text-secondary">
                        No areas found
                      </td>
                    </tr>
                  ) : (
                    areas.map((area) => (
                      <tr key={area.id}>
                        <td className="font-medium">{area.code}</td>
                        <td>{area.name}</td>
                        <td>
                          <span className={`badge ${area.is_active ? 'badge-primary' : 'badge-secondary'}`}>
                            {area.is_active ? 'Active' : 'Inactive'}
                          </span>
                        </td>
                        <td>
                          <div className="flex gap-2">
                            <button
                              onClick={() => handleEdit(area)}
                              className="text-salesforce-blue hover:text-salesforce-blue-hover text-sm"
                            >
                              Edit
                            </button>
                            <button
                              onClick={() => handleManage(area)}
                              className="text-salesforce-blue hover:text-salesforce-blue-hover text-sm"
                            >
                              Configure
                            </button>
                            <button
                              onClick={() => handleDelete(area.id)}
                              className="text-red-600 hover:text-red-700 text-sm"
                            >
                              Delete
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>

          {/* Configuration Panel */}
          <div className="card">
            <div className="p-4 border-b border-neutral-border">
              <h2 className="text-lg font-semibold text-neutral-text-primary">
                {selectedArea ? `${selectedArea.name} (${selectedArea.code})` : 'Select an Area'}
              </h2>
            </div>

            {selectedArea ? (
              <div className="p-4">
                {loadingData ? (
                  <div className="text-center py-8 text-neutral-text-secondary">Loading...</div>
                ) : (
                  <>
                    {/* Tabs */}
                    <div className="flex gap-2 mb-4 border-b border-neutral-border">
                      <button
                        onClick={() => setActiveTab('zones')}
                        className={`px-4 py-2 text-sm font-medium ${
                          activeTab === 'zones'
                            ? 'border-b-2 border-salesforce-blue text-salesforce-blue'
                            : 'text-neutral-text-secondary hover:text-neutral-text-primary'
                        }`}
                      >
                        Zones ({areaZones.length})
                      </button>
                      <button
                        onClick={() => setActiveTab('branches')}
                        className={`px-4 py-2 text-sm font-medium ${
                          activeTab === 'branches'
                            ? 'border-b-2 border-salesforce-blue text-salesforce-blue'
                            : 'text-neutral-text-secondary hover:text-neutral-text-primary'
                        }`}
                      >
                        Individual Branches ({areaBranches.length})
                      </button>
                      <button
                        onClick={() => setActiveTab('all')}
                        className={`px-4 py-2 text-sm font-medium ${
                          activeTab === 'all'
                            ? 'border-b-2 border-salesforce-blue text-salesforce-blue'
                            : 'text-neutral-text-secondary hover:text-neutral-text-primary'
                        }`}
                      >
                        All Branches ({allAreaBranches.length})
                      </button>
                    </div>

                    {/* Zones Tab */}
                    {activeTab === 'zones' && (
                      <div>
                        <div className="mb-4">
                          <label className="block text-sm font-medium text-neutral-text-primary mb-2">
                            Add Zone
                          </label>
                          <select
                            onChange={(e) => {
                              if (e.target.value) {
                                handleAddZone(e.target.value);
                                e.target.value = '';
                              }
                            }}
                            className="input-field"
                          >
                            <option value="">Select a zone to add...</option>
                            {zones
                              .filter(z => !areaZones.some(az => az.id === z.id))
                              .map((zone) => (
                                <option key={zone.id} value={zone.id}>
                                  {zone.name} ({zone.code})
                                </option>
                              ))}
                          </select>
                        </div>
                        <div className="space-y-2 max-h-64 overflow-y-auto">
                          {areaZones.length === 0 ? (
                            <div className="text-sm text-neutral-text-secondary text-center py-4">
                              No zones assigned
                            </div>
                          ) : (
                            areaZones.map((zone) => (
                              <div key={zone.id} className="flex items-center justify-between p-2 border border-neutral-border rounded">
                                <div>
                                  <span className="font-medium">{zone.name}</span>
                                  <span className="text-xs text-neutral-text-secondary ml-2">({zone.code})</span>
                                </div>
                                <button
                                  onClick={() => handleRemoveZone(zone.id)}
                                  className="text-red-600 hover:text-red-700 text-sm"
                                >
                                  Remove
                                </button>
                              </div>
                            ))
                          )}
                        </div>
                      </div>
                    )}

                    {/* Individual Branches Tab */}
                    {activeTab === 'branches' && (
                      <div>
                        <div className="mb-4">
                          <label className="block text-sm font-medium text-neutral-text-primary mb-2">
                            Add Individual Branch
                          </label>
                          <select
                            onChange={(e) => {
                              if (e.target.value) {
                                handleAddBranch(e.target.value);
                                e.target.value = '';
                              }
                            }}
                            className="input-field"
                          >
                            <option value="">Select a branch to add...</option>
                            {branches
                              .filter(b => !areaBranches.some(ab => ab.id === b.id))
                              .map((branch) => (
                                <option key={branch.id} value={branch.id}>
                                  {branch.name} ({branch.code})
                                </option>
                              ))}
                          </select>
                        </div>
                        <div className="space-y-2 max-h-64 overflow-y-auto">
                          {areaBranches.length === 0 ? (
                            <div className="text-sm text-neutral-text-secondary text-center py-4">
                              No individual branches assigned
                            </div>
                          ) : (
                            areaBranches.map((branch) => (
                              <div key={branch.id} className="flex items-center justify-between p-2 border border-neutral-border rounded">
                                <div>
                                  <span className="font-medium">{branch.name}</span>
                                  <span className="text-xs text-neutral-text-secondary ml-2">({branch.code})</span>
                                </div>
                                <button
                                  onClick={() => handleRemoveBranch(branch.id)}
                                  className="text-red-600 hover:text-red-700 text-sm"
                                >
                                  Remove
                                </button>
                              </div>
                            ))
                          )}
                        </div>
                      </div>
                    )}

                    {/* All Branches Tab */}
                    {activeTab === 'all' && (
                      <div className="space-y-2 max-h-64 overflow-y-auto">
                        {allAreaBranches.length === 0 ? (
                          <div className="text-sm text-neutral-text-secondary text-center py-4">
                            No branches assigned
                          </div>
                        ) : (
                          allAreaBranches.map((branch) => (
                            <div key={branch.id} className="p-2 border border-neutral-border rounded">
                              <span className="font-medium">{branch.name}</span>
                              <span className="text-xs text-neutral-text-secondary ml-2">({branch.code})</span>
                            </div>
                          ))
                        )}
                      </div>
                    )}

                    <div className="mt-4 pt-4 border-t border-neutral-border">
                      <button
                        onClick={() => {
                          setSelectedArea(null);
                          setAreaZones([]);
                          setAreaBranches([]);
                          setAllAreaBranches([]);
                        }}
                        className="btn-secondary w-full"
                      >
                        Close
                      </button>
                    </div>
                  </>
                )}
              </div>
            ) : (
              <div className="p-8 text-center text-neutral-text-secondary">
                Select an area from the list to configure zones and branches
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
                {editingArea ? 'Edit Area of Operation' : 'Create Area of Operation'}
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
                      placeholder="e.g., AREA1"
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
                      placeholder="e.g., North Region"
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
                      setEditingArea(null);
                    }}
                    className="btn-secondary"
                  >
                    Cancel
                  </button>
                  <button type="submit" className="btn-primary">
                    {editingArea ? 'Update' : 'Create'}
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
