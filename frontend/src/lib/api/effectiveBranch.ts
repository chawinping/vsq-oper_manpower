import apiClient from './client';
import { Branch } from './branch';

export interface EffectiveBranch {
  id: string;
  rotation_staff_id: string;
  branch_id: string;
  branch?: Branch;
  level: number; // 1 = priority, 2 = reserved
  created_at: string;
}

export interface CreateEffectiveBranchRequest {
  rotation_staff_id: string;
  branch_id: string;
  level: number; // 1 or 2
}

export interface BulkUpdateEffectiveBranchesRequest {
  rotation_staff_id: string;
  effective_branches: {
    branch_id: string;
    level: number;
  }[];
}

export const effectiveBranchApi = {
  getByRotationStaffID: async (rotationStaffId: string) => {
    const response = await apiClient.get(`/effective-branches/rotation-staff/${rotationStaffId}`);
    return (response.data.effective_branches || []) as EffectiveBranch[];
  },

  create: async (data: CreateEffectiveBranchRequest) => {
    const response = await apiClient.post('/effective-branches', data);
    return response.data.effective_branch as EffectiveBranch;
  },

  delete: async (id: string) => {
    const response = await apiClient.delete(`/effective-branches/${id}`);
    return response.data;
  },

  bulkUpdate: async (data: BulkUpdateEffectiveBranchesRequest) => {
    const response = await apiClient.put('/effective-branches/bulk-update', data);
    return response.data.effective_branches as EffectiveBranch[];
  },
};


