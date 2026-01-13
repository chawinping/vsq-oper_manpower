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
};


