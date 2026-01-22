import apiClient from './client';

export interface Position {
  id: string;
  name: string;
  position_code?: string;  // Unique code for position (e.g., "BM", "ABM", "DA")
  position_type: 'branch' | 'rotation';
  manpower_type: 'พนักงานฟร้อนท์' | 'ผู้ช่วยแพทย์' | 'อื่นๆ' | 'ทำความสะอาด';
  display_order: number;
  branch_staff_count?: number;   // Read-only field showing number of branch staff allocated to this position
  rotation_staff_count?: number;  // Read-only field showing number of rotation staff allocated to this position
  created_at: string;
}

export interface PositionQuotaAssociation {
  quota_id: string;
  branch_id: string;
  branch_name: string;
  designated_quota: number;
  minimum_required: number;
}

export interface PositionAssociations {
  staff_count: number;
  quota_count: number;
  quotas: PositionQuotaAssociation[];
  allocation_rule_count: number;
  suggestion_count: number;
  scenario_requirement_count: number;
  total_count: number;
}

export interface UpdatePositionRequest {
  name: string;
  position_code?: string;
  display_order: number;
  position_type: 'branch' | 'rotation';
  manpower_type: 'พนักงานฟร้อนท์' | 'ผู้ช่วยแพทย์' | 'อื่นๆ' | 'ทำความสะอาด';
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
  
  update: async (id: string, data: UpdatePositionRequest) => {
    const response = await apiClient.put(`/positions/${id}`, data);
    return response.data.position as Position;
  },
  
  delete: async (id: string) => {
    const response = await apiClient.delete(`/positions/${id}`);
    return response.data;
  },
  
  getAssociations: async (id: string) => {
    const response = await apiClient.get(`/positions/${id}/associations`);
    return response.data.associations as PositionAssociations;
  },
};


