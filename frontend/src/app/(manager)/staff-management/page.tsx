'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { authApi, User } from '@/lib/api/auth';
import { staffApi, Staff, CreateStaffRequest } from '@/lib/api/staff';
import { positionApi, Position } from '@/lib/api/position';
import { branchApi, Branch } from '@/lib/api/branch';
import AppLayout from '@/components/layout/AppLayout';

export default function StaffManagementPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
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

  const [formData, setFormData] = useState<CreateStaffRequest>({
    name: '',
    staff_type: 'branch',
    position_id: '',
    branch_id: '',
    coverage_area: '',
  });

  useEffect(() => {
    const fetchData = async () => {
      try {
        const userData = await authApi.getMe();
        setUser(userData);
        
        await loadStaff();
        const [positionsData, branchesData] = await Promise.all([
          positionApi.list(),
          branchApi.list(),
        ]);
        setPositions(positionsData || []);
        setBranches(branchesData || []);
      } catch (error: any) {
        console.error('Failed to fetch data:', error);
        // Only redirect if not already on login page
        if (typeof window !== 'undefined' && !window.location.pathname.includes('/login')) {
          router.push('/login');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [router]);

  const loadStaff = async () => {
    try {
      const filters: any = {};
      if (filterType) filters.staff_type = filterType;
      const staffData = await staffApi.list(filters);
      setStaff(staffData || []);
    } catch (error) {
      console.error('Failed to load staff:', error);
      setStaff([]);
    }
  };

  useEffect(() => {
    loadStaff();
  }, [filterType]);


  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const data: CreateStaffRequest = {
        ...formData,
        branch_id: formData.staff_type === 'branch' && formData.branch_id ? formData.branch_id : undefined,
        coverage_area: formData.staff_type === 'rotation' && formData.coverage_area ? formData.coverage_area : undefined,
      };

      if (editingStaff) {
        await staffApi.update(editingStaff.id, data);
      } else {
        await staffApi.create(data);
      }

      setShowModal(false);
      setEditingStaff(null);
      setFormData({
        name: '',
        staff_type: 'branch',
        position_id: '',
        branch_id: '',
        coverage_area: '',
      });
      await loadStaff();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save staff');
    }
  };

  const handleEdit = (staffMember: Staff) => {
    setEditingStaff(staffMember);
    setFormData({
      name: staffMember.name,
      staff_type: staffMember.staff_type,
      position_id: staffMember.position_id,
      branch_id: staffMember.branch_id || '',
      coverage_area: staffMember.coverage_area || '',
    });
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

  return (
    <AppLayout>
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">Staff Management</h1>
          <p className="text-sm text-neutral-text-secondary">Manage branch and rotation staff</p>
        </div>

        <div className="card">
          <div className="p-4 border-b border-neutral-border">
            <div className="flex items-center justify-between flex-wrap gap-4">
              <div className="flex items-center gap-4">
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
              </div>
              {canManage && (
                <div className="flex gap-2">
                  <button
                    onClick={() => {
                      setEditingStaff(null);
                      setFormData({
                        name: '',
                        staff_type: 'branch',
                        position_id: '',
                        branch_id: '',
                        coverage_area: '',
                      });
                      setShowModal(true);
                    }}
                    className="btn-primary"
                  >
                    Add Staff
                  </button>
                  <button
                    onClick={() => setShowImportModal(true)}
                    className="btn-primary bg-green-600 hover:bg-green-700"
                  >
                    Import from Excel
                  </button>
                </div>
              )}
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="table-salesforce">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Type</th>
                  <th>Position</th>
                  <th>Branch</th>
                  <th>Coverage Area</th>
                  {canManage && <th>Actions</th>}
                </tr>
              </thead>
              <tbody>
                {(staff || []).map((staffMember) => {
                  const position = (positions || []).find((p) => p.id === staffMember.position_id);
                  const branch = (branches || []).find((b) => b.id === staffMember.branch_id);
                  return (
                    <tr key={staffMember.id}>
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
                      {canManage && (
                        <td>
                          <div className="flex gap-3">
                            <button
                              onClick={() => handleEdit(staffMember)}
                              className="text-salesforce-blue hover:text-salesforce-blue-hover text-sm"
                            >
                              Edit
                            </button>
                            {user?.role === 'admin' && (
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
                      Name *
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
                      Staff Type *
                    </label>
                    <select
                      required
                      value={formData.staff_type}
                      onChange={(e) =>
                        setFormData({ ...formData, staff_type: e.target.value as 'branch' | 'rotation' })
                      }
                      className="input-field"
                    >
                      <option value="branch">Branch Staff</option>
                      <option value="rotation">Rotation Staff</option>
                    </select>
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
                    </div>
                  )}
                  {formData.staff_type === 'rotation' && (
                    <div>
                      <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                        Coverage Area
                      </label>
                      <input
                        type="text"
                        value={formData.coverage_area}
                        onChange={(e) => setFormData({ ...formData, coverage_area: e.target.value })}
                        className="input-field"
                      />
                    </div>
                  )}
                </div>
                <div className="mt-6 flex justify-end gap-2">
                  <button
                    type="button"
                    onClick={() => {
                      setShowModal(false);
                      setEditingStaff(null);
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
                  Expected format: Name | Staff Type | Position ID | Branch ID | Coverage Area
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
    </AppLayout>
  );
}

