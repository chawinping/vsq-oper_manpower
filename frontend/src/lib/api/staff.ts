import apiClient from './client';

export interface Staff {
  id: string;
  nickname?: string;
  name: string; // Full name
  staff_type: 'branch' | 'rotation';
  position_id: string;
  position?: {
    id: string;
    name: string;
  };
  branch_id?: string;
  coverage_area?: string;
  area_of_operation_id?: string; // Legacy field
  area_of_operation?: {
    id: string;
    name: string;
    code: string;
  };
  zone_id?: string; // Zone assignment for rotation staff
  zone?: {
    id: string;
    name: string;
    code: string;
  };
  branches?: Array<{
    id: string;
    name: string;
    code: string;
  }>; // Individual branches for rotation staff
  skill_level: number; // Rating 0-10
  created_at: string;
  updated_at: string;
}

export interface CreateStaffRequest {
  nickname?: string;
  name: string;
  staff_type: 'branch' | 'rotation';
  position_id: string;
  branch_id?: string;
  coverage_area?: string;
  area_of_operation_id?: string; // Legacy field
  zone_id?: string; // Zone assignment for rotation staff
  branch_ids?: string[]; // Individual branches for rotation staff
  skill_level?: number; // Rating 0-10
}

export const staffApi = {
  list: async (filters?: { staff_type?: string; branch_id?: string; position_id?: string; area_of_operation_id?: string }) => {
    const response = await apiClient.get('/staff', { params: filters });
    return (response.data.staff || []) as Staff[];
  },
  
  create: async (data: CreateStaffRequest) => {
    const response = await apiClient.post('/staff', data);
    return response.data.staff as Staff;
  },
  
  update: async (id: string, data: CreateStaffRequest) => {
    const response = await apiClient.put(`/staff/${id}`, data);
    return response.data.staff as Staff;
  },
  
  delete: async (id: string) => {
    const response = await apiClient.delete(`/staff/${id}`);
    return response.data;
  },
  
  import: async (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    const response = await apiClient.post('/staff/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    return response.data;
  },
};


