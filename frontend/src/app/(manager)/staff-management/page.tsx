'use client';

import { useEffect, useState } from 'react';
import { useUser } from '@/contexts/UserContext';
import { staffApi, Staff, CreateStaffRequest } from '@/lib/api/staff';
import { positionApi, Position } from '@/lib/api/position';
import { branchApi, Branch } from '@/lib/api/branch';
import { effectiveBranchApi, EffectiveBranch } from '@/lib/api/effectiveBranch';

export default function StaffManagementPage() {
  const { user } = useUser();
  const [staff, setStaff] = useState<Staff[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingStaff, setEditingStaff] = useState<Staff | null>(null);
  const [showImportModal, setShowImportModal] = useState(false);
  const [importFile, setImportFile] = useState<File | null>(null);
  const [importing, setImporting] = useState(false);
  const [filterType, setFilterType] = useState<string>('');
  const [filterBranchId, setFilterBranchId] = useState<string>('');

  const [formData, setFormData] = useState<CreateStaffRequest>({
    nickname: '',
    name: '',
    staff_type: 'branch',
    position_id: '',
    branch_id: '',
    coverage_area: '',
    skill_level: 5,
  });
  const [effectiveBranches, setEffectiveBranches] = useState<{ branch_id: string; level: number }[]>([]);
  const [loadingEffectiveBranches, setLoadingEffectiveBranches] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [positionsData, branchesData] = await Promise.all([
          positionApi.list(),
          branchApi.list(),
        ]);
        setPositions(positionsData || []);
        setBranches(branchesData || []);
        
        // Set default branch filter based on user role
        if (user?.role === 'branch_manager' && user?.branch_id) {
          setFilterBranchId(user.branch_id);
        } else {
          setFilterBranchId(''); // All Branches for area/district managers
        }
      } catch (error: any) {
        console.error('Failed to fetch data:', error);
      } finally {
        setLoading(false);
      }
    };

    if (user) {
      fetchData();
    }
  }, [user]);

  // Load staff when filters change (but only after initial data is loaded)
  useEffect(() => {
    if (!loading && branches.length > 0) {
      loadStaff();
    }
  }, [filterType, filterBranchId]);

  const loadStaff = async () => {
    try {
      const filters: any = {};
      if (filterType) filters.staff_type = filterType;
      if (filterBranchId) filters.branch_id = filterBranchId;
      const staffData = await staffApi.list(filters);
      setStaff(staffData || []);
    } catch (error) {
      console.error('Failed to load staff:', error);
      setStaff([]);
    }
  };

  useEffect(() => {
    if (!loading && branches.length > 0) {
      loadStaff();
    }
  }, [filterType, filterBranchId, loading]);


  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const data: CreateStaffRequest = {
        ...formData,
        branch_id: formData.staff_type === 'branch' && formData.branch_id ? formData.branch_id : undefined,
        coverage_area: formData.staff_type === 'rotation' && formData.coverage_area ? formData.coverage_area : undefined,
      };

      let staffId: string;
      if (editingStaff) {
        await staffApi.update(editingStaff.id, data);
        staffId = editingStaff.id;
      } else {
        const createdStaff = await staffApi.create(data);
        staffId = createdStaff.id;
      }

      // Save effective branches if this is rotation staff
      if (formData.staff_type === 'rotation') {
        try {
          // Always use bulk update - it will replace all existing effective branches
          await effectiveBranchApi.bulkUpdate({
            rotation_staff_id: staffId,
            effective_branches: effectiveBranches,
          });
        } catch (error: any) {
          console.error('Failed to save effective branches:', error);
          alert('Staff saved but failed to save effective branches: ' + (error.response?.data?.error || error.message));
        }
      }

      setShowModal(false);
      setEditingStaff(null);
      setFormData({
        nickname: '',
        name: '',
        staff_type: 'branch',
        position_id: '',
        branch_id: '',
        coverage_area: '',
        skill_level: 5,
      });
      setEffectiveBranches([]);
      await loadStaff();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save staff');
    }
  };

  const handleEdit = async (staffMember: Staff) => {
    setEditingStaff(staffMember);
    setFormData({
      nickname: staffMember.nickname || '',
      name: staffMember.name,
      staff_type: staffMember.staff_type,
      position_id: staffMember.position_id,
      branch_id: staffMember.branch_id || '',
      coverage_area: staffMember.coverage_area || '',
      skill_level: staffMember.skill_level || 5,
    });
    
    // Load effective branches if this is rotation staff
    if (staffMember.staff_type === 'rotation') {
      setLoadingEffectiveBranches(true);
      try {
        const ebs = await effectiveBranchApi.getByRotationStaffID(staffMember.id);
        setEffectiveBranches(ebs.map(eb => ({ branch_id: eb.branch_id, level: eb.level })));
      } catch (error) {
        console.error('Failed to load effective branches:', error);
        setEffectiveBranches([]);
      } finally {
        setLoadingEffectiveBranches(false);
      }
    } else {
      setEffectiveBranches([]);
    }
    
    setShowModal(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this staff member?')) {
      return;
    }

    try {
      await staffApi.delete(id);
      await loadStaff();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete staff');
    }
  };

  const handleImport = async () => {
    if (!importFile) {
      alert('Please select a file');
      return;
    }

    setImporting(true);
    try {
      const result = await staffApi.import(importFile);
      alert(`Import completed! ${result.imported || 0} staff members imported.`);
      setShowImportModal(false);
      setImportFile(null);
      await loadStaff();
    } catch (error: any) {
      const errorMsg = error.response?.data?.error || 'Failed to import staff';
      if (error.response?.status === 207) {
        // Partial success
        alert(`${errorMsg}\nImported: ${error.response.data.imported || 0}`);
        await loadStaff();
      } else {
        alert(errorMsg);
      }
    } finally {
      setImporting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-neutral-text-secondary">Loading...</div>
      </div>
    );
  }

  const canManage = user?.role === 'admin' || user?.role === 'area_manager' || user?.role === 'district_manager';
  const isBranchManager = user?.role === 'branch_manager';
  const isAreaManager = user?.role === 'area_manager' || user?.role === 'district_manager';
  const canManageRotation = user?.role === 'admin' || user?.role === 'area_manager' || user?.role === 'district_manager';

  return (
    <>
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">Staff Management</h1>
          {isBranchManager && user?.branch_id ? (
            (() => {
              const userBranch = branches.find(b => b.id === user.branch_id);
              return userBranch ? (
                <div>
                  <h2 className="text-lg font-semibold text-neutral-text-primary mb-1">
                    {userBranch.name} ({userBranch.code})
                  </h2>
                  <p className="text-sm text-neutral-text-secondary">
                    Managing staff for your branch
                  </p>
                </div>
              ) : (
                <p className="text-sm text-neutral-text-secondary">Manage branch and rotation staff</p>
              );
            })()
          ) : (
            <p className="text-sm text-neutral-text-secondary">Manage branch and rotation staff</p>
          )}
        </div>

        <div className="card">
          <div className="p-4 border-b border-neutral-border">
            <div className="flex items-center justify-between flex-wrap gap-4">
              <div className="flex items-center gap-4 flex-wrap">
                <label htmlFor="filter-type" className="text-sm font-medium text-neutral-text-primary">
                  Filter by Type:
                </label>
                <select
                  id="filter-type"
                  value={filterType}
                  onChange={(e) => setFilterType(e.target.value)}
                  className="input-field w-auto min-w-[150px]"
                >
                  <option value="">All</option>
                  <option value="branch">Branch Staff</option>
                  <option value="rotation">Rotation Staff</option>
                </select>
                
                {/* Branch Code Filter - for Area/District Managers */}
                {(isAreaManager || isBranchManager) && (
                  <>
                    <label htmlFor="filter-branch" className="text-sm font-medium text-neutral-text-primary">
                      Branch:
                    </label>
                    <select
                      id="filter-branch"
                      value={filterBranchId}
                      onChange={(e) => setFilterBranchId(e.target.value)}
                      className="input-field w-auto min-w-[200px]"
                      disabled={isBranchManager}
                    >
                      {!isBranchManager && <option value="">All Branches</option>}
                      {(branches || []).map((branch) => (
                        <option key={branch.id} value={branch.id}>
                          {branch.code} - {branch.name}
                        </option>
                      ))}
                    </select>
                    {isBranchManager && (
                      <span className="text-xs text-neutral-text-secondary">
                        (Your branch - cannot change)
                      </span>
                    )}
                  </>
                )}
              </div>
              {(canManage || isBranchManager) && (
                <div className="flex gap-2">
                  {(canManage || isBranchManager) && (
                    <button
                    onClick={() => {
                      setEditingStaff(null);
                      setFormData({
                        nickname: '',
                        name: '',
                        staff_type: 'branch',
                        position_id: '',
                        branch_id: '',
                        coverage_area: '',
                        skill_level: 5,
                      });
                      setEffectiveBranches([]);
                      setShowModal(true);
                    }}
                      className="btn-primary"
                    >
                      Add Staff
                    </button>
                  )}
                  {canManage && (
                    <button
                      onClick={() => setShowImportModal(true)}
                      className="btn-primary bg-green-600 hover:bg-green-700"
                    >
                      Import from Excel
                    </button>
                  )}
                </div>
              )}
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="table-salesforce">
              <thead>
                <tr>
                  <th>Nickname</th>
                  <th>Name</th>
                  <th>Type</th>
                  <th>Position</th>
                  <th>Branch</th>
                  <th>Coverage Area</th>
                  <th>Skill Level</th>
                  {(canManage || isBranchManager) && <th>Actions</th>}
                </tr>
              </thead>
              <tbody>
                {(staff || []).map((staffMember) => {
                  const position = (positions || []).find((p) => p.id === staffMember.position_id);
                  const branch = (branches || []).find((b) => b.id === staffMember.branch_id);
                  return (
                    <tr key={staffMember.id}>
                      <td className="font-medium">{staffMember.nickname || '-'}</td>
                      <td className="font-medium">{staffMember.name}</td>
                      <td>
                        <span className={`badge ${
                          staffMember.staff_type === 'branch'
                            ? 'badge-primary'
                            : 'badge-secondary'
                        }`}>
                          {staffMember.staff_type}
                        </span>
                      </td>
                      <td>{position?.name || '-'}</td>
                      <td>{branch?.name || '-'}</td>
                      <td>{staffMember.coverage_area || '-'}</td>
                      <td>
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                          {staffMember.skill_level || 5}/10
                        </span>
                      </td>
                      {(canManage || isBranchManager) && (
                        <td>
                          <div className="flex gap-3">
                            {/* Branch managers can only edit branch staff */}
                            {(canManage || (isBranchManager && staffMember.staff_type === 'branch')) && (
                              <button
                                onClick={() => handleEdit(staffMember)}
                                className="text-salesforce-blue hover:text-salesforce-blue-hover text-sm"
                              >
                                Edit
                              </button>
                            )}
                            {/* Branch managers can only delete branch staff */}
                            {(user?.role === 'admin' || (isBranchManager && staffMember.staff_type === 'branch')) && (
                              <button
                                onClick={() => handleDelete(staffMember.id)}
                                className="text-red-600 hover:text-red-700 text-sm"
                              >
                                Delete
                              </button>
                            )}
                          </div>
                        </td>
                      )}
                    </tr>
                  );
                })}
              </tbody>
            </table>
            {staff.length === 0 && (
              <div className="text-center py-12 text-neutral-text-secondary">
                No staff members found
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Add/Edit Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="card max-w-md w-full">
            <div className="p-6">
              <h2 className="text-xl font-semibold text-neutral-text-primary mb-6">
                {editingStaff ? 'Edit Staff' : 'Add Staff'}
              </h2>
              <form onSubmit={handleSubmit}>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Nickname
                    </label>
                    <input
                      type="text"
                      value={formData.nickname || ''}
                      onChange={(e) => setFormData({ ...formData, nickname: e.target.value })}
                      className="input-field"
                      placeholder="Optional"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Full Name *
                    </label>
                    <input
                      type="text"
                      required
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      className="input-field"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Skill Level (0-10)
                    </label>
                    <input
                      type="number"
                      min="0"
                      max="10"
                      value={formData.skill_level || 5}
                      onChange={(e) => setFormData({ ...formData, skill_level: parseInt(e.target.value) || 5 })}
                      className="input-field"
                    />
                    <p className="mt-1 text-xs text-neutral-text-secondary">
                      Rate staff skill level from 0 (beginner) to 10 (expert)
                    </p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Staff Type *
                    </label>
                    <select
                      required
                      value={formData.staff_type}
                      onChange={(e) => {
                        const newType = e.target.value as 'branch' | 'rotation';
                        setFormData({ ...formData, staff_type: newType });
                        // Clear effective branches when switching away from rotation
                        if (newType !== 'rotation') {
                          setEffectiveBranches([]);
                        } else if (editingStaff && editingStaff.staff_type === 'rotation') {
                          // Load effective branches when switching to rotation for existing rotation staff
                          setLoadingEffectiveBranches(true);
                          effectiveBranchApi.getByRotationStaffID(editingStaff.id)
                            .then(ebs => {
                              setEffectiveBranches(ebs.map(eb => ({ branch_id: eb.branch_id, level: eb.level })));
                              setLoadingEffectiveBranches(false);
                            })
                            .catch(() => {
                              setEffectiveBranches([]);
                              setLoadingEffectiveBranches(false);
                            });
                        }
                      }}
                      className="input-field"
                      disabled={isBranchManager}
                    >
                      <option value="branch">Branch Staff</option>
                      {canManageRotation && <option value="rotation">Rotation Staff</option>}
                    </select>
                    {isBranchManager && (
                      <p className="mt-1 text-xs text-neutral-text-secondary">
                        Branch managers can only add branch staff
                      </p>
                    )}
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Position *
                    </label>
                    <select
                      required
                      value={formData.position_id}
                      onChange={(e) => setFormData({ ...formData, position_id: e.target.value })}
                      className="input-field"
                    >
                      <option value="">Select Position</option>
                      {(positions || []).map((p) => (
                        <option key={p.id} value={p.id}>
                          {p.name}
                        </option>
                      ))}
                    </select>
                  </div>
                  {formData.staff_type === 'branch' && (
                    <div>
                      <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                        Branch
                      </label>
                      {isBranchManager ? (
                        <input
                          type="text"
                          value={branches.find(b => b.id === user?.branch_id)?.name || 'Your Branch'}
                          className="input-field"
                          disabled
                        />
                      ) : (
                        <select
                          value={formData.branch_id}
                          onChange={(e) => setFormData({ ...formData, branch_id: e.target.value })}
                          className="input-field"
                        >
                          <option value="">Select Branch</option>
                          {(branches || []).map((b) => (
                            <option key={b.id} value={b.id}>
                              {b.name} ({b.code})
                            </option>
                          ))}
                        </select>
                      )}
                    </div>
                  )}
                  {formData.staff_type === 'rotation' && (
                    <>
                      <div>
                        <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                          Coverage Area
                        </label>
                        <input
                          type="text"
                          value={formData.coverage_area}
                          onChange={(e) => setFormData({ ...formData, coverage_area: e.target.value })}
                          className="input-field"
                          placeholder="e.g., Area A"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                          Effective Branches *
                        </label>
                        <p className="text-xs text-neutral-text-secondary mb-2">
                          Select branches this rotation staff can support. Level 1 = Priority, Level 2 = Reserved.
                        </p>
                        {loadingEffectiveBranches ? (
                          <div className="text-sm text-neutral-text-secondary">Loading...</div>
                        ) : (
                          <div className="space-y-2 max-h-64 overflow-y-auto border border-neutral-border rounded-md p-3">
                            {branches.map((branch) => {
                              const existingEB = effectiveBranches.find(eb => eb.branch_id === branch.id);
                              return (
                                <div key={branch.id} className="flex items-center gap-3">
                                  <input
                                    type="checkbox"
                                    id={`branch-${branch.id}`}
                                    checked={!!existingEB}
                                    onChange={(e) => {
                                      if (e.target.checked) {
                                        setEffectiveBranches([...effectiveBranches, { branch_id: branch.id, level: 1 }]);
                                      } else {
                                        setEffectiveBranches(effectiveBranches.filter(eb => eb.branch_id !== branch.id));
                                      }
                                    }}
                                    className="w-4 h-4"
                                  />
                                  <label htmlFor={`branch-${branch.id}`} className="flex-1 cursor-pointer">
                                    <span className="font-medium">{branch.name}</span>
                                    <span className="text-xs text-neutral-text-secondary ml-2">({branch.code})</span>
                                  </label>
                                  {existingEB && (
                                    <select
                                      value={existingEB.level}
                                      onChange={(e) => {
                                        setEffectiveBranches(
                                          effectiveBranches.map(eb =>
                                            eb.branch_id === branch.id
                                              ? { ...eb, level: parseInt(e.target.value) }
                                              : eb
                                          )
                                        );
                                      }}
                                      className="text-xs border border-neutral-border rounded px-2 py-1"
                                      onClick={(e) => e.stopPropagation()}
                                    >
                                      <option value={1}>Level 1 (Priority)</option>
                                      <option value={2}>Level 2 (Reserved)</option>
                                    </select>
                                  )}
                                </div>
                              );
                            })}
                            {effectiveBranches.length === 0 && (
                              <div className="text-sm text-neutral-text-secondary text-center py-2">
                                No branches selected. Select at least one branch.
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    </>
                  )}
                </div>
                <div className="mt-6 flex justify-end gap-2">
                  <button
                    type="button"
                    onClick={() => {
                      setShowModal(false);
                      setEditingStaff(null);
                      setEffectiveBranches([]);
                    }}
                    className="btn-secondary"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="btn-primary"
                  >
                    {editingStaff ? 'Update' : 'Create'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Import Modal */}
      {showImportModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="card max-w-md w-full">
            <div className="p-6">
              <h2 className="text-xl font-semibold text-neutral-text-primary mb-6">Import Staff from Excel</h2>
              <div className="mb-4">
                <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                  Excel File
                </label>
                <input
                  type="file"
                  accept=".xlsx,.xls"
                  onChange={(e) => setImportFile(e.target.files?.[0] || null)}
                  className="input-field"
                />
                <p className="mt-2 text-xs text-neutral-text-secondary">
                  Expected format (columns A-E):<br />
                  Name (required) | Staff Type (branch/rotation, required) | Position Name (required, e.g., "Nurse") | Branch Code (optional, e.g., "TMA") | Nickname (optional)
                </p>
              </div>
              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => {
                    setShowImportModal(false);
                    setImportFile(null);
                  }}
                  className="btn-secondary"
                  disabled={importing}
                >
                  Cancel
                </button>
                <button
                  onClick={handleImport}
                  disabled={!importFile || importing}
                  className="btn-primary bg-green-600 hover:bg-green-700 disabled:opacity-50"
                >
                  {importing ? 'Importing...' : 'Import'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

