import apiClient from './client';

export interface PositionQuota {
  id: string;
  branch_id: string;
  branch?: { id: string; name: string; code: string };
  position_id: string;
  position?: { id: string; name: string };
  designated_quota: number;
  minimum_required: number;
  is_active: boolean;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface PositionQuotaStatus {
  position_id: string;
  position_name: string;
  designated_quota: number;
  minimum_required: number;
  available_local: number;
  assigned_rotation: number;
  total_assigned: number;
  still_required: number;
}

export interface BranchQuotaStatus {
  branch_id: string;
  branch_name: string;
  branch_code: string;
  date: string;
  position_statuses: PositionQuotaStatus[];
  total_designated: number;
  total_available: number;
  total_assigned: number;
  total_required: number;
}

export interface CreatePositionQuotaRequest {
  branch_id: string;
  position_id: string;
  designated_quota: number;
  minimum_required: number;
}

export const quotaApi = {
  list: async (filters?: {
    branch_id?: string;
    position_id?: string;
  }): Promise<PositionQuota[]> => {
    const params = new URLSearchParams();
    if (filters?.branch_id) params.append('branch_id', filters.branch_id);
    if (filters?.position_id) params.append('position_id', filters.position_id);
    
    const response = await apiClient.get(`/quotas?${params.toString()}`);
    return response.data.quotas || [];
  },

  getById: async (id: string): Promise<PositionQuota> => {
    const response = await apiClient.get(`/quotas/${id}`);
    return response.data.quota;
  },

  create: async (data: CreatePositionQuotaRequest): Promise<PositionQuota> => {
    const response = await apiClient.post('/quotas', data);
    return response.data.quota;
  },

  update: async (id: string, data: Partial<CreatePositionQuotaRequest & { is_active?: boolean }>): Promise<PositionQuota> => {
    const response = await apiClient.put(`/quotas/${id}`, data);
    return response.data.quota;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/quotas/${id}`);
  },

  getBranchQuotaStatus: async (branchId: string, date?: string): Promise<BranchQuotaStatus> => {
    const params = new URLSearchParams();
    if (date) params.append('date', date);
    
    const response = await apiClient.get(`/quotas/branch/${branchId}/status?${params.toString()}`);
    return response.data.status;
  },
};
