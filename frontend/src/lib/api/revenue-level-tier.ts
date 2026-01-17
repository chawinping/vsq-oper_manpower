import apiClient from './client';

export interface RevenueLevelTier {
  id: string;
  level_number: number;
  level_name: string;
  min_revenue: number;
  max_revenue: number | null;
  display_order: number;
  color_code: string | null;
  description: string | null;
  created_at: string;
  updated_at: string;
}

export interface RevenueLevelTierCreate {
  level_number: number;
  level_name: string;
  min_revenue: number;
  max_revenue?: number | null;
  display_order?: number;
  color_code?: string | null;
  description?: string | null;
}

export interface RevenueLevelTierUpdate {
  level_name?: string;
  min_revenue?: number;
  max_revenue?: number | null;
  display_order?: number;
  color_code?: string | null;
  description?: string | null;
}

export const revenueLevelTierApi = {
  list: async () => {
    const response = await apiClient.get('/revenue-level-tiers');
    return response.data as RevenueLevelTier[];
  },

  getById: async (id: string) => {
    const response = await apiClient.get(`/revenue-level-tiers/${id}`);
    return response.data as RevenueLevelTier;
  },

  create: async (data: RevenueLevelTierCreate) => {
    const response = await apiClient.post('/revenue-level-tiers', data);
    return response.data as RevenueLevelTier;
  },

  update: async (id: string, data: RevenueLevelTierUpdate) => {
    const response = await apiClient.put(`/revenue-level-tiers/${id}`, data);
    return response.data as RevenueLevelTier;
  },

  delete: async (id: string) => {
    const response = await apiClient.delete(`/revenue-level-tiers/${id}`);
    return response.data;
  },

  getTierForRevenue: async (revenue: number) => {
    const response = await apiClient.post('/revenue-level-tiers/match', { revenue });
    return response.data as RevenueLevelTier;
  },
};
