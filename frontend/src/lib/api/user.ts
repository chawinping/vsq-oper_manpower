import apiClient from './client';

export interface User {
  id: string;
  username: string;
  email: string;
  role_id: string;
  role_name: string;
  created_at: string;
  updated_at: string;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  role_id: string;
}

export interface UpdateUserRequest {
  username?: string;
  email?: string;
  password?: string;
  role_id?: string;
}

export const userApi = {
  list: async () => {
    const response = await apiClient.get('/users');
    return (response.data.users || []) as User[];
  },
  
  create: async (data: CreateUserRequest) => {
    const response = await apiClient.post('/users', data);
    return response.data.user as User;
  },
  
  update: async (id: string, data: UpdateUserRequest) => {
    const response = await apiClient.put(`/users/${id}`, data);
    return response.data.user as User;
  },
  
  delete: async (id: string) => {
    const response = await apiClient.delete(`/users/${id}`);
    return response.data;
  },
};


