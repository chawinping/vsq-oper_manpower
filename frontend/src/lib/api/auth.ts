import apiClient from './client';

export interface LoginRequest {
  username: string;
  password: string;
}

export interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  branch_id?: string;
  branch_name?: string;
  branch_code?: string;
}

export const authApi = {
  login: async (data: LoginRequest) => {
    const response = await apiClient.post('/auth/login', data);
    return response.data;
  },
  
  logout: async () => {
    const response = await apiClient.post('/auth/logout');
    return response.data;
  },
  
  getMe: async () => {
    const response = await apiClient.get('/auth/me');
    if (!response.data || !response.data.user) {
      throw new Error('Invalid response format');
    }
    return response.data.user as User;
  },
};



