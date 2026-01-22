'use client';

import { useEffect, useState, Fragment } from 'react';
import { useUser } from '@/contexts/UserContext';
import { branchApi, Branch, CreateBranchRequest } from '@/lib/api/branch';
import { quotaApi } from '@/lib/api/quota';
import { format, subDays } from 'date-fns';

interface RevenueData {
  id: string;
  branch_id: string;
  date: string;
  expected_revenue: number;
  actual_revenue?: number;
}

export default function BranchManagementPage() {
  const { user } = useUser();
  const [branches, setBranches] = useState<Branch[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingBranch, setEditingBranch] = useState<Branch | null>(null);
  const [showRevenueModal, setShowRevenueModal] = useState(false);
  const [selectedBranch, setSelectedBranch] = useState<Branch | null>(null);
  const [revenueData, setRevenueData] = useState<RevenueData[]>([]);
  const [loadingRevenue, setLoadingRevenue] = useState(false);
  
  // Import state
  const [showImportModal, setShowImportModal] = useState(false);
  const [importFile, setImportFile] = useState<File | null>(null);
  const [importing, setImporting] = useState(false);
  const [importResult, setImportResult] = useState<{ created: number; updated: number; errors?: string[] } | null>(null);

  const [formData, setFormData] = useState<CreateBranchRequest>({
    name: '',
    code: '',
    area_manager_id: '',
    priority: 1,
  });

  useEffect(() => {
    const fetchData = async () => {
      try {
        await loadBranches();
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

  const loadBranches = async () => {
    try {
      const branchesData = await branchApi.list();
      setBranches(branchesData || []);
    } catch (error) {
      console.error('Failed to load branches:', error);
      setBranches([]);
    }
  };


  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const data: CreateBranchRequest = {
        ...formData,
        area_manager_id: formData.area_manager_id || undefined,
        priority: formData.priority || 1,
      };

      if (editingBranch) {
        await branchApi.update(editingBranch.id, data);
      } else {
        await branchApi.create(data);
      }

      setShowModal(false);
      setEditingBranch(null);
      setFormData({
        name: '',
        code: '',
        area_manager_id: '',
        priority: 1,
      });
      await loadBranches();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save branch');
    }
  };

  const handleEdit = (branch: Branch) => {
    setEditingBranch(branch);
    setFormData({
      name: branch.name,
      code: branch.code,
      area_manager_id: branch.area_manager_id || '',
      priority: branch.priority,
    });
    setShowModal(true);
  };

  const handleViewRevenue = async (branch: Branch) => {
    setSelectedBranch(branch);
    setLoadingRevenue(true);
    try {
      const endDate = format(new Date(), 'yyyy-MM-dd');
      const startDate = format(subDays(new Date(), 30), 'yyyy-MM-dd');
      const revenue = await branchApi.getRevenue(branch.id, startDate, endDate);
      setRevenueData(revenue || []);
      setShowRevenueModal(true);
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to load revenue data');
    } finally {
      setLoadingRevenue(false);
    }
  };

  const handleImport = async () => {
    if (!importFile) return;

    setImporting(true);
    setImportResult(null);

    try {
      const result = await quotaApi.import(importFile);
      setImportResult(result);
      
      // Clear file input
      setImportFile(null);
      
      // Show success message
      alert(`Import completed successfully!\nCreated: ${result.created}\nUpdated: ${result.updated}${result.errors && result.errors.length > 0 ? `\n\nErrors: ${result.errors.length}` : ''}`);
      
      // Close modal if no errors or if user wants to proceed
      if (!result.errors || result.errors.length === 0) {
        setShowImportModal(false);
      }
    } catch (err: any) {
      const errorMsg = err.response?.data?.error || 'Failed to import quotas';
      const errors = err.response?.data?.errors || [];
      const created = err.response?.data?.created || 0;
      const updated = err.response?.data?.updated || 0;
      
      setImportResult({
        created,
        updated,
        errors: errors.length > 0 ? errors : [errorMsg],
      });
      
      alert(`Import completed with errors.\nCreated: ${created}\nUpdated: ${updated}\nErrors: ${errors.length}`);
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
    <Fragment>
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">Branch Management</h1>
          <p className="text-sm text-neutral-text-secondary">Manage branch information and revenue tracking</p>
        </div>

        <div className="card">
          <div className="p-4 border-b border-neutral-border flex justify-between items-center">
            <div className="flex gap-2">
              {canManage && (
                <button
                  onClick={() => {
                    setEditingBranch(null);
                    setFormData({
                      name: '',
                      code: '',
                      area_manager_id: '',
                      priority: 1,
                    });
                    setShowModal(true);
                  }}
                  className="btn-primary"
                >
                  Add Branch
                </button>
              )}
              {canManage && (
                <button
                  onClick={() => {
                    setShowImportModal(true);
                    setImportFile(null);
                    setImportResult(null);
                  }}
                  className="btn-primary bg-green-600 hover:bg-green-700"
                >
                  Import Position Quotas
                </button>
              )}
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="table-salesforce">
              <thead>
                <tr>
                  <th>Code</th>
                  <th>Name</th>
                  <th>Priority</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {(branches || []).map((branch) => (
                  <tr key={branch.id}>
                    <td className="font-medium">{branch.code}</td>
                    <td>{branch.name}</td>
                    <td>
                      <span className={`badge ${
                        branch.priority === 1
                          ? 'bg-red-50 text-red-700'
                          : branch.priority === 2
                          ? 'bg-yellow-50 text-yellow-700'
                          : 'bg-green-50 text-green-700'
                      }`}>
                        {branch.priority}
                      </span>
                    </td>
                    <td>
                      <div className="flex gap-3">
                        <button
                          onClick={() => handleViewRevenue(branch)}
                          className="text-green-600 hover:text-green-700 text-sm"
                        >
                          Revenue
                        </button>
                        {canManage && (
                          <>
                            <button
                              onClick={() => window.location.href = `/branch-config/${branch.id}`}
                              className="text-purple-600 hover:text-purple-700 text-sm"
                            >
                              Configure
                            </button>
                            <button
                              onClick={() => handleEdit(branch)}
                              className="text-salesforce-blue hover:text-salesforce-blue-hover text-sm"
                            >
                              Edit
                            </button>
                          </>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {branches.length === 0 && (
              <div className="text-center py-12 text-neutral-text-secondary">
                No branches found
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Add/Edit Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="card max-w-md w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <h2 className="text-xl font-semibold text-neutral-text-primary mb-6">
                {editingBranch ? 'Edit Branch' : 'Add Branch'}
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
                      Code *
                    </label>
                    <input
                      type="text"
                      required
                      value={formData.code}
                      onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                      className="input-field"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Priority (1=High, 2=Medium, 3=Low)
                    </label>
                    <select
                      value={formData.priority}
                      onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) })}
                      className="input-field"
                    >
                      <option value={1}>1 - High</option>
                      <option value={2}>2 - Medium</option>
                      <option value={3}>3 - Low</option>
                    </select>
                  </div>
                </div>
                <div className="mt-6 flex justify-end gap-2">
                  <button
                    type="button"
                    onClick={() => {
                      setShowModal(false);
                      setEditingBranch(null);
                    }}
                    className="btn-secondary"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="btn-primary"
                  >
                    {editingBranch ? 'Update' : 'Create'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Revenue Modal */}
      {showRevenueModal && selectedBranch && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="card max-w-4xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-semibold text-neutral-text-primary">
                  Revenue History - {selectedBranch.name}
                </h2>
                <button
                  onClick={() => {
                    setShowRevenueModal(false);
                    setSelectedBranch(null);
                    setRevenueData([]);
                  }}
                  className="text-neutral-text-secondary hover:text-neutral-text-primary"
                >
                  ✕
                </button>
              </div>
              {loadingRevenue ? (
                <div className="text-center py-8 text-neutral-text-secondary">Loading revenue data...</div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="table-salesforce">
                    <thead>
                      <tr>
                        <th>Date</th>
                        <th>Expected Revenue</th>
                        <th>Actual Revenue</th>
                        <th>Difference</th>
                      </tr>
                    </thead>
                    <tbody>
                      {(revenueData || []).map((revenue) => {
                        const diff = (revenue.actual_revenue || 0) - revenue.expected_revenue;
                        return (
                          <tr key={revenue.id}>
                            <td>{format(new Date(revenue.date), 'MMM d, yyyy')}</td>
                            <td>
                              {revenue.expected_revenue.toLocaleString('en-US', {
                                style: 'currency',
                                currency: 'THB',
                                minimumFractionDigits: 0,
                              })}
                            </td>
                            <td>
                              {revenue.actual_revenue
                                ? revenue.actual_revenue.toLocaleString('en-US', {
                                    style: 'currency',
                                    currency: 'THB',
                                    minimumFractionDigits: 0,
                                  })
                                : '-'}
                            </td>
                            <td className={`font-medium ${
                              diff >= 0 ? 'text-green-600' : 'text-red-600'
                            }`}>
                              {revenue.actual_revenue
                                ? `${diff >= 0 ? '+' : ''}${diff.toLocaleString('en-US', {
                                    style: 'currency',
                                    currency: 'THB',
                                    minimumFractionDigits: 0,
                                  })}`
                                : '-'}
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                  {revenueData.length === 0 && (
                    <div className="text-center py-8 text-neutral-text-secondary">
                      No revenue data available
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Import Position Quotas Modal */}
      {showImportModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="card max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-semibold text-neutral-text-primary">Import Position Quotas from Excel</h2>
                <button
                  onClick={() => {
                    setShowImportModal(false);
                    setImportFile(null);
                    setImportResult(null);
                  }}
                  className="text-neutral-text-secondary hover:text-neutral-text-primary"
                >
                  ✕
                </button>
              </div>
              
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
                  Expected format (columns A-D):<br />
                  <strong>Column A:</strong> Branch Code (required) - e.g., "TMA", "CPN"<br />
                  <strong>Column B:</strong> Position Code (required) - e.g., "BM", "ABM", "DA"<br />
                  <strong>Column C:</strong> Preferred No. (required) - designated quota<br />
                  <strong>Column D:</strong> Minimum No. (required) - minimum required<br />
                  <br />
                  <strong>Note:</strong> Header row is optional and will be auto-detected. The import will create new quotas or update existing ones for all branches specified in the file.
                </p>
              </div>

              {importResult && (
                <div className={`mb-4 p-3 rounded-md ${importResult.errors && importResult.errors.length > 0 ? 'bg-yellow-50 border border-yellow-200' : 'bg-green-50 border border-green-200'}`}>
                  <p className={`text-sm font-medium ${importResult.errors && importResult.errors.length > 0 ? 'text-yellow-800' : 'text-green-800'}`}>
                    Import completed: {importResult.created} created, {importResult.updated} updated
                  </p>
                  {importResult.errors && importResult.errors.length > 0 && (
                    <div className="mt-2">
                      <p className="text-xs font-semibold text-yellow-800 mb-1">Errors ({importResult.errors.length}):</p>
                      <ul className="text-xs text-yellow-700 list-disc list-inside max-h-40 overflow-y-auto">
                        {importResult.errors.slice(0, 20).map((err, idx) => (
                          <li key={idx}>{err}</li>
                        ))}
                        {importResult.errors.length > 20 && (
                          <li className="text-yellow-600 italic">... and {importResult.errors.length - 20} more errors</li>
                        )}
                      </ul>
                    </div>
                  )}
                </div>
              )}

              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => {
                    setShowImportModal(false);
                    setImportFile(null);
                    setImportResult(null);
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
    </Fragment>
  );
}

