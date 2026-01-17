import apiClient from './client';
import { Branch } from './branch';

export interface Zone {
  id: string;
  name: string;
  code: string;
  description?: string;
  is_active: boolean;
  branches?: Branch[];
  created_at: string;
  updated_at: string;
}

export interface CreateZoneRequest {
  name: string;
  code: string;
  description?: string;
  is_active?: boolean;
}

export interface UpdateZoneBranchesRequest {
  branch_ids: string[];
}

export const zoneApi = {
  list: async (includeInactive?: boolean) => {
    const response = await apiClient.get('/zones', {
      params: { include_inactive: includeInactive },
    });
    return (response.data.zones || []) as Zone[];
  },

  getById: async (id: string) => {
    const response = await apiClient.get(`/zones/${id}`);
    return response.data.zone as Zone;
  },

  create: async (data: CreateZoneRequest) => {
    const response = await apiClient.post('/zones', data);
    return response.data.zone as Zone;
  },

  update: async (id: string, data: CreateZoneRequest) => {
    const response = await apiClient.put(`/zones/${id}`, data);
    return response.data.zone as Zone;
  },

  delete: async (id: string) => {
    const response = await apiClient.delete(`/zones/${id}`);
    return response.data;
  },

  getBranches: async (id: string) => {
    const response = await apiClient.get(`/zones/${id}/branches`);
    return (response.data.branches || []) as Branch[];
  },

  updateBranches: async (id: string, data: UpdateZoneBranchesRequest) => {
    const response = await apiClient.put(`/zones/${id}/branches`, data);
    return response.data;
  },
};
