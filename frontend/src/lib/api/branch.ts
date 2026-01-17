import apiClient from './client';

export interface Branch {
  id: string;
  name: string;
  code: string;
  area_manager_id?: string;
  priority: number;
  created_at: string;
  updated_at: string;
}

export interface CreateBranchRequest {
  name: string;
  code: string;
  area_manager_id?: string;
  priority?: number;
}

export const branchApi = {
  list: async () => {
    const response = await apiClient.get('/branches');
    return (response.data.branches || []) as Branch[];
  },
  
  create: async (data: CreateBranchRequest) => {
    const response = await apiClient.post('/branches', data);
    return response.data.branch as Branch;
  },
  
  update: async (id: string, data: CreateBranchRequest) => {
    const response = await apiClient.put(`/branches/${id}`, data);
    return response.data.branch as Branch;
  },
  
  getRevenue: async (id: string, startDate?: string, endDate?: string) => {
    const response = await apiClient.get(`/branches/${id}/revenue`, {
      params: { start_date: startDate, end_date: endDate },
    });
    return response.data.revenue_data || [];
  },
};


