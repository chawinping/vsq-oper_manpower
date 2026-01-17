import apiClient from './client';
import { Branch } from './branch';

export interface EffectiveBranch {
  id: string;
  rotation_staff_id: string;
  branch_id: string;
  branch?: Branch;
  level: number; // 1 = priority, 2 = reserved
  commute_duration_minutes?: number; // Travel time in minutes (default: 300)
  transit_count?: number; // Number of transits (default: 10)
  travel_cost?: number; // Cost of traveling (default: 1000)
  created_at: string;
}

export interface CreateEffectiveBranchRequest {
  rotation_staff_id: string;
  branch_id: string;
  level: number; // 1 or 2
  commute_duration_minutes?: number;
  transit_count?: number;
  travel_cost?: number;
}

export interface BulkUpdateEffectiveBranchesRequest {
  rotation_staff_id: string;
  effective_branches: {
    branch_id: string;
    level: number;
    commute_duration_minutes?: number;
    transit_count?: number;
    travel_cost?: number;
  }[];
}

export interface UpdateEffectiveBranchRequest {
  branch_id: string;
  level: number;
  commute_duration_minutes?: number;
  transit_count?: number;
  travel_cost?: number;
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

  update: async (id: string, data: UpdateEffectiveBranchRequest) => {
    const response = await apiClient.put(`/effective-branches/${id}`, data);
    return response.data.effective_branch as EffectiveBranch;
  },
};


