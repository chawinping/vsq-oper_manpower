import apiClient from './client';

export interface SystemSetting {
  id: string;
  key: string;
  value: string;
  description?: string;
  updated_at: string;
}

export interface UpdateSettingRequest {
  value: string;
  description?: string;
}

export const settingsApi = {
  getAll: async () => {
    const response = await apiClient.get('/settings');
    return (response.data.settings || []) as SystemSetting[];
  },
  
  update: async (key: string, data: UpdateSettingRequest) => {
    const response = await apiClient.put(`/settings/${key}`, data);
    return response.data.setting as SystemSetting;
  },
};


