import apiClient from './client';

export interface AreaOfOperation {
  id: string;
  name: string;
  code: string;
  description?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateAreaOfOperationRequest {
  name: string;
  code: string;
  description?: string;
  is_active?: boolean;
}

export const areaOfOperationApi = {
  list: async (includeInactive?: boolean) => {
    const response = await apiClient.get('/areas-of-operation', {
      params: { include_inactive: includeInactive },
    });
    return (response.data.areas_of_operation || []) as AreaOfOperation[];
  },

  getByID: async (id: string) => {
    const response = await apiClient.get(`/areas-of-operation/${id}`);
    return response.data.area_of_operation as AreaOfOperation;
  },

  create: async (data: CreateAreaOfOperationRequest) => {
    const response = await apiClient.post('/areas-of-operation', data);
    return response.data.area_of_operation as AreaOfOperation;
  },

  update: async (id: string, data: CreateAreaOfOperationRequest) => {
    const response = await apiClient.put(`/areas-of-operation/${id}`, data);
    return response.data.area_of_operation as AreaOfOperation;
  },

  delete: async (id: string) => {
    const response = await apiClient.delete(`/areas-of-operation/${id}`);
    return response.data;
  },

  // Zone management
  addZone: async (areaId: string, zoneId: string) => {
    const response = await apiClient.post(`/areas-of-operation/${areaId}/zones`, {
      zone_id: zoneId,
    });
    return response.data;
  },

  removeZone: async (areaId: string, zoneId: string) => {
    const response = await apiClient.delete(`/areas-of-operation/${areaId}/zones/${zoneId}`);
    return response.data;
  },

  getZones: async (areaId: string) => {
    const response = await apiClient.get(`/areas-of-operation/${areaId}/zones`);
    return (response.data.zones || []) as any[];
  },

  // Branch management
  addBranch: async (areaId: string, branchId: string) => {
    const response = await apiClient.post(`/areas-of-operation/${areaId}/branches`, {
      branch_id: branchId,
    });
    return response.data;
  },

  removeBranch: async (areaId: string, branchId: string) => {
    const response = await apiClient.delete(`/areas-of-operation/${areaId}/branches/${branchId}`);
    return response.data;
  },

  getBranches: async (areaId: string) => {
    const response = await apiClient.get(`/areas-of-operation/${areaId}/branches`);
    return (response.data.branches || []) as any[];
  },

  getAllBranches: async (areaId: string) => {
    const response = await apiClient.get(`/areas-of-operation/${areaId}/all-branches`);
    return (response.data.branches || []) as any[];
  },
};


