'use client';

import { useState, useEffect } from 'react';
import { positionApi, PositionAssociations } from '@/lib/api/position';
import { useRouter } from 'next/navigation';

interface PositionAssociationsModalProps {
  positionId: string;
  positionName: string;
  isOpen: boolean;
  onClose: () => void;
  onDeleted: () => void;
}

export default function PositionAssociationsModal({
  positionId,
  positionName,
  isOpen,
  onClose,
  onDeleted,
}: PositionAssociationsModalProps) {
  const router = useRouter();
  const [associations, setAssociations] = useState<PositionAssociations | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen && positionId) {
      loadAssociations();
    }
  }, [isOpen, positionId]);

  const loadAssociations = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await positionApi.getAssociations(positionId);
      setAssociations(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load associations');
    } finally {
      setLoading(false);
    }
  };


  const handleGoToBranchConfig = (branchId: string) => {
    router.push(`/branch-config/${branchId}`);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">Position Associations</h2>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 text-2xl font-bold"
            >
              ×
            </button>
          </div>

          <div className="mb-4">
            <p className="text-sm text-gray-600">
              <span className="font-semibold">Position:</span> {positionName}
            </p>
            <p className="text-xs text-gray-500 mt-1">
              Position quotas will be automatically deleted. Other associations must be removed manually before deletion.
            </p>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {loading ? (
            <div className="flex justify-center items-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            </div>
          ) : associations ? (
            <div className="space-y-4">
              {/* Summary */}
              <div className="p-4 bg-blue-50 border border-blue-200 rounded-md">
                <h3 className="font-semibold text-blue-900 mb-2">Summary</h3>
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div>
                    <span className="font-medium">Staff:</span> {associations.staff_count}
                  </div>
                  <div>
                    <span className="font-medium">Quotas:</span> {associations.quota_count}
                  </div>
                  <div>
                    <span className="font-medium">Allocation Rules:</span> {associations.allocation_rule_count}
                  </div>
                  <div>
                    <span className="font-medium">Suggestions:</span> {associations.suggestion_count}
                  </div>
                  <div>
                    <span className="font-medium">Scenario Requirements:</span> {associations.scenario_requirement_count}
                  </div>
                  <div>
                    <span className="font-medium">Total:</span> {associations.total_count}
                  </div>
                </div>
              </div>

              {/* Position Quotas */}
              {associations.quota_count > 0 && (
                <div>
                  <h3 className="font-semibold mb-2">Position Quotas ({associations.quota_count})</h3>
                  <div className="mb-2 p-2 bg-blue-50 border border-blue-200 rounded text-xs text-blue-800">
                    <span className="font-medium">ℹ️ Note:</span> Position quotas will be automatically deleted when you delete this position.
                  </div>
                  <div className="border border-gray-200 rounded-md overflow-hidden">
                    <table className="min-w-full divide-y divide-gray-200">
                      <thead className="bg-gray-50">
                        <tr>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Branch</th>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Preferred</th>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Minimum</th>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
                        </tr>
                      </thead>
                      <tbody className="bg-white divide-y divide-gray-200">
                        {associations.quotas.map((quota) => (
                          <tr key={quota.quota_id}>
                            <td className="px-4 py-2 text-sm">{quota.branch_name}</td>
                            <td className="px-4 py-2 text-sm">{quota.designated_quota}</td>
                            <td className="px-4 py-2 text-sm">{quota.minimum_required}</td>
                            <td className="px-4 py-2 text-sm">
                              <button
                                onClick={() => handleGoToBranchConfig(quota.branch_id)}
                                className="text-blue-600 hover:text-blue-700 text-xs"
                              >
                                View Config
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              {/* Other Associations */}
              {(associations.staff_count > 0 ||
                associations.allocation_rule_count > 0 ||
                associations.suggestion_count > 0 ||
                associations.scenario_requirement_count > 0) && (
                <div>
                  <h3 className="font-semibold mb-2">Other Associations</h3>
                  <div className="space-y-2 text-sm">
                    {associations.staff_count > 0 && (
                      <div className="p-2 bg-yellow-50 border border-yellow-200 rounded">
                        <span className="font-medium">Staff:</span> {associations.staff_count} staff member(s) are assigned to this position.
                        <span className="text-xs text-gray-600 block mt-1">
                          Please reassign or remove staff before deleting this position.
                        </span>
                      </div>
                    )}
                    {associations.allocation_rule_count > 0 && (
                      <div className="p-2 bg-yellow-50 border border-yellow-200 rounded">
                        <span className="font-medium">Allocation Rules:</span> {associations.allocation_rule_count} rule(s) reference this position.
                        <span className="text-xs text-gray-600 block mt-1">
                          Please remove allocation rules before deleting this position.
                        </span>
                      </div>
                    )}
                    {associations.suggestion_count > 0 && (
                      <div className="p-2 bg-yellow-50 border border-yellow-200 rounded">
                        <span className="font-medium">Suggestions:</span> {associations.suggestion_count} suggestion(s) reference this position.
                        <span className="text-xs text-gray-600 block mt-1">
                          These will be automatically cleaned up when the position is deleted.
                        </span>
                      </div>
                    )}
                    {associations.scenario_requirement_count > 0 && (
                      <div className="p-2 bg-yellow-50 border border-yellow-200 rounded">
                        <span className="font-medium">Scenario Requirements:</span> {associations.scenario_requirement_count} requirement(s) reference this position.
                        <span className="text-xs text-gray-600 block mt-1">
                          Please remove scenario requirements before deleting this position.
                        </span>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* No associations */}
              {associations.total_count === 0 && (
                <div className="p-4 bg-green-50 border border-green-200 rounded-md text-center">
                  <p className="text-green-800 font-medium">No associations found</p>
                  <p className="text-sm text-green-700 mt-1">This position can be safely deleted.</p>
                </div>
              )}
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">No association data available</div>
          )}

          <div className="mt-6 flex justify-end">
            <button
              onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
