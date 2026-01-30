import apiClient from './client';
import { Position } from './position';
import { Staff } from './staff';

export interface RotationStaffBranchPosition {
  id: string;
  rotation_staff_id: string;
  branch_position_id: string;
  rotation_staff?: Staff;
  branch_position?: Position;
  substitution_level: number; // 1 = preferred, 2 = acceptable, 3 = emergency only
  is_active: boolean;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateRotationStaffBranchPositionRequest {
  rotation_staff_id: string;
  branch_position_id: string;
  substitution_level: number;
  is_active?: boolean;
  notes?: string;
}

export interface UpdateRotationStaffBranchPositionRequest {
  substitution_level: number;
  is_active: boolean;
  notes?: string;
}

export const rotationStaffBranchPositionApi = {
  list: async (params?: { staff_id?: string; position_id?: string }) => {
    const response = await apiClient.get('/rotation-staff-branch-positions', { params });
    return (response.data.mappings || []) as RotationStaffBranchPosition[];
  },
  
  getById: async (id: string) => {
    const response = await apiClient.get(`/rotation-staff-branch-positions/${id}`);
    return response.data.mapping as RotationStaffBranchPosition;
  },
  
  create: async (data: CreateRotationStaffBranchPositionRequest) => {
    const response = await apiClient.post('/rotation-staff-branch-positions', data);
    return response.data.mapping as RotationStaffBranchPosition;
  },
  
  update: async (id: string, data: UpdateRotationStaffBranchPositionRequest) => {
    const response = await apiClient.put(`/rotation-staff-branch-positions/${id}`, data);
    return response.data.mapping as RotationStaffBranchPosition;
  },
  
  delete: async (id: string) => {
    const response = await apiClient.delete(`/rotation-staff-branch-positions/${id}`);
    return response.data;
  },
};
