'use client';

import { useEffect, useState, Fragment } from 'react';
import { useUser } from '@/contexts/UserContext';
import { branchApi, Branch, CreateBranchRequest } from '@/lib/api/branch';
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

  const [formData, setFormData] = useState<CreateBranchRequest>({
    name: '',
    code: '',
    address: '',
    area_manager_id: '',
    expected_revenue: 0,
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
        expected_revenue: formData.expected_revenue || 0,
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
        address: '',
        area_manager_id: '',
        expected_revenue: 0,
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
      address: branch.address,
      area_manager_id: branch.area_manager_id || '',
      expected_revenue: branch.expected_revenue,
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
          <div className="p-4 border-b border-neutral-border">
            {canManage && (
              <button
                onClick={() => {
                  setEditingBranch(null);
                  setFormData({
                    name: '',
                    code: '',
                    address: '',
                    area_manager_id: '',
                    expected_revenue: 0,
                    priority: 1,
                  });
                  setShowModal(true);
                }}
                className="btn-primary"
              >
                Add Branch
              </button>
            )}
          </div>

          <div className="overflow-x-auto">
            <table className="table-salesforce">
              <thead>
                <tr>
                  <th>Code</th>
                  <th>Name</th>
                  <th>Address</th>
                  <th>Expected Revenue</th>
                  <th>Priority</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {(branches || []).map((branch) => (
                  <tr key={branch.id}>
                    <td className="font-medium">{branch.code}</td>
                    <td>{branch.name}</td>
                    <td>{branch.address || '-'}</td>
                    <td>
                      {branch.expected_revenue.toLocaleString('en-US', {
                        style: 'currency',
                        currency: 'THB',
                        minimumFractionDigits: 0,
                      })}
                    </td>
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
                          <button
                            onClick={() => handleEdit(branch)}
                            className="text-salesforce-blue hover:text-salesforce-blue-hover text-sm"
                          >
                            Edit
                          </button>
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
                      Address
                    </label>
                    <textarea
                      value={formData.address}
                      onChange={(e) => setFormData({ ...formData, address: e.target.value })}
                      rows={3}
                      className="input-field"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Expected Revenue (Daily)
                    </label>
                    <input
                      type="number"
                      step="0.01"
                      min="0"
                      value={formData.expected_revenue}
                      onChange={(e) =>
                        setFormData({ ...formData, expected_revenue: parseFloat(e.target.value) || 0 })
                      }
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
    </Fragment>
  );
}

