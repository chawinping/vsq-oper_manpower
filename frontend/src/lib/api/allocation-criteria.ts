import apiClient from './client';

export interface AllocationCriteria {
  id: string;
  pillar: 'clinic_wide' | 'doctor_specific' | 'branch_specific';
  type: 'bookings' | 'revenue' | 'min_staff_position' | 'min_staff_branch' | 'doctor_count';
  weight: number;
  is_active: boolean;
  description?: string;
  config?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateAllocationCriteriaRequest {
  pillar: 'clinic_wide' | 'doctor_specific' | 'branch_specific';
  type: 'bookings' | 'revenue' | 'min_staff_position' | 'min_staff_branch' | 'doctor_count';
  weight: number;
  is_active?: boolean;
  description?: string;
  config?: string;
}

export const allocationCriteriaApi = {
  list: async (filters?: {
    pillar?: string;
    type?: string;
    is_active?: boolean;
  }): Promise<AllocationCriteria[]> => {
    const params = new URLSearchParams();
    if (filters?.pillar) params.append('pillar', filters.pillar);
    if (filters?.type) params.append('type', filters.type);
    if (filters?.is_active !== undefined) params.append('is_active', filters.is_active.toString());
    
    const response = await apiClient.get(`/allocation-criteria?${params.toString()}`);
    return response.data.criteria || [];
  },

  getById: async (id: string): Promise<AllocationCriteria> => {
    const response = await apiClient.get(`/allocation-criteria/${id}`);
    return response.data.criteria;
  },

  create: async (data: CreateAllocationCriteriaRequest): Promise<AllocationCriteria> => {
    const response = await apiClient.post('/allocation-criteria', data);
    return response.data.criteria;
  },

  update: async (id: string, data: Partial<CreateAllocationCriteriaRequest>): Promise<AllocationCriteria> => {
    const response = await apiClient.put(`/allocation-criteria/${id}`, data);
    return response.data.criteria;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/allocation-criteria/${id}`);
  },
};
