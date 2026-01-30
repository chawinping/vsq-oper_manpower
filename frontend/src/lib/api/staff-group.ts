import apiClient from './client';
import { Position } from './position';

export interface StaffGroup {
  id: string;
  name: string;
  description?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  positions?: StaffGroupPosition[];
}

export interface StaffGroupPosition {
  id: string;
  staff_group_id: string;
  position_id: string;
  position?: Position;
  created_at: string;
}

export interface CreateStaffGroupRequest {
  name: string;
  description?: string;
  is_active?: boolean;
}

export interface UpdateStaffGroupRequest {
  name: string;
  description?: string;
  is_active: boolean;
}

export const staffGroupApi = {
  list: async () => {
    const response = await apiClient.get('/staff-groups');
    return (response.data.staff_groups || []) as StaffGroup[];
  },
  
  getById: async (id: string) => {
    const response = await apiClient.get(`/staff-groups/${id}`);
    return response.data.staff_group as StaffGroup;
  },
  
  create: async (data: CreateStaffGroupRequest) => {
    const response = await apiClient.post('/staff-groups', data);
    return response.data.staff_group as StaffGroup;
  },
  
  update: async (id: string, data: UpdateStaffGroupRequest) => {
    const response = await apiClient.put(`/staff-groups/${id}`, data);
    return response.data.staff_group as StaffGroup;
  },
  
  delete: async (id: string) => {
    const response = await apiClient.delete(`/staff-groups/${id}`);
    return response.data;
  },
  
  addPosition: async (staffGroupId: string, positionId: string) => {
    const response = await apiClient.post(`/staff-groups/${staffGroupId}/positions`, {
      position_id: positionId,
    });
    return response.data.staff_group_position as StaffGroupPosition;
  },
  
  removePosition: async (staffGroupId: string, positionId: string) => {
    const response = await apiClient.delete(`/staff-groups/${staffGroupId}/positions/${positionId}`);
    return response.data;
  },
};
