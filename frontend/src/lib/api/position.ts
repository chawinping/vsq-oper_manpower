import apiClient from './client';

export interface Position {
  id: string;
  name: string;
  min_staff_per_branch: number;
  revenue_multiplier: number;
  created_at: string;
}

export const positionApi = {
  list: async () => {
    const response = await apiClient.get('/positions');
    return (response.data.positions || []) as Position[];
  },
  
  getById: async (id: string) => {
    const response = await apiClient.get(`/positions/${id}`);
    return response.data.position as Position;
  },
};


