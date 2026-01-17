'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { branchApi, Branch } from '@/lib/api/branch';
import BranchPositionQuotaConfig from '@/components/branch/BranchPositionQuotaConfig';
import BranchWeeklyRevenueConfig from '@/components/branch/BranchWeeklyRevenueConfig';

export default function BranchConfigPage() {
  const params = useParams();
  const router = useRouter();
  const branchId = params.id as string;

  const [branch, setBranch] = useState<Branch | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (branchId) {
      loadBranch();
    }
  }, [branchId]);

  const loadBranch = async () => {
    try {
      setLoading(true);
      const branches = await branchApi.list();
      const foundBranch = branches.find((b) => b.id === branchId);
      if (foundBranch) {
        setBranch(foundBranch);
      } else {
        setError('Branch not found');
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load branch');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = () => {
    // Optionally reload data or show success message
    console.log('Configuration saved');
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error || !branch) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <p className="text-sm text-red-800">{error || 'Branch not found'}</p>
          <button
            onClick={() => router.back()}
            className="mt-4 px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"
          >
            Go Back
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-6">
        <button
          onClick={() => router.back()}
          className="mb-4 text-sm text-blue-600 hover:text-blue-800"
        >
          ← Back
        </button>
        <h1 className="text-2xl font-bold text-gray-900">Branch Configuration</h1>
        <p className="mt-1 text-sm text-gray-500">
          Configure staff quotas and expected revenue for {branch.name} ({branch.code})
        </p>
      </div>

      <div className="space-y-8">
        {/* Position Quota Configuration */}
        <div className="bg-white shadow rounded-lg p-6">
          <BranchPositionQuotaConfig branchId={branchId} onSave={handleSave} />
        </div>

        {/* Weekly Revenue Configuration */}
        <div className="bg-white shadow rounded-lg p-6">
          <BranchWeeklyRevenueConfig branchId={branchId} onSave={handleSave} />
        </div>
      </div>
    </div>
  );
}
