import apiClient from './client';

export interface Role {
  id: string;
  name: string;
  created_at: string;
}

export const roleApi = {
  list: async () => {
    const response = await apiClient.get('/roles');
    return (response.data.roles || []) as Role[];
  },
};

