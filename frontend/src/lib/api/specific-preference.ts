import apiClient from './client';

export type SpecificPreferenceType = 'position_count' | 'staff_name';

export interface SpecificPreference {
  id: string;
  branch_id?: string;
  branch?: {
    id: string;
    name: string;
    code: string;
  };
  doctor_id?: string;
  doctor?: {
    id: string;
    name: string;
    code: string;
  };
  day_of_week?: number; // 0-6 (Sunday-Saturday), null for any day
  preference_type: SpecificPreferenceType;
  position_id?: string;
  position?: {
    id: string;
    name: string;
  };
  staff_count?: number;
  staff_id?: string;
  staff?: {
    id: string;
    name: string;
    nickname: string;
  };
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateSpecificPreferenceRequest {
  branch_id?: string | null; // null for "any branch"
  doctor_id?: string | null; // null for "any doctor"
  day_of_week?: number | null; // 0-6 or null for "any day"
  preference_type: SpecificPreferenceType;
  position_id?: string; // Required for position_count
  staff_count?: number; // Required for position_count
  staff_id?: string; // Required for staff_name
  is_active?: boolean;
}

export interface UpdateSpecificPreferenceRequest {
  branch_id?: string | null;
  doctor_id?: string | null;
  day_of_week?: number | null;
  preference_type?: SpecificPreferenceType;
  position_id?: string | null;
  staff_count?: number | null;
  staff_id?: string | null;
  is_active?: boolean;
}

export interface SpecificPreferenceFilters {
  branch_id?: string;
  doctor_id?: string;
  day_of_week?: number;
  is_active?: boolean;
}

export const specificPreferenceApi = {
  list: async (filters?: SpecificPreferenceFilters): Promise<{ preferences: SpecificPreference[] }> => {
    const params = new URLSearchParams();
    if (filters?.branch_id) params.append('branch_id', filters.branch_id);
    if (filters?.doctor_id) params.append('doctor_id', filters.doctor_id);
    if (filters?.day_of_week !== undefined) params.append('day_of_week', filters.day_of_week.toString());
    if (filters?.is_active !== undefined) params.append('is_active', filters.is_active.toString());
    
    const queryString = params.toString();
    const url = `/specific-preferences${queryString ? `?${queryString}` : ''}`;
    const response = await apiClient.get(url);
    return response.data;
  },

  getById: async (id: string): Promise<{ preference: SpecificPreference }> => {
    const response = await apiClient.get(`/specific-preferences/${id}`);
    return response.data;
  },

  create: async (data: CreateSpecificPreferenceRequest): Promise<{ preference: SpecificPreference }> => {
    const response = await apiClient.post('/specific-preferences', data);
    return response.data;
  },

  update: async (id: string, data: UpdateSpecificPreferenceRequest): Promise<{ preference: SpecificPreference }> => {
    const response = await apiClient.put(`/specific-preferences/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/specific-preferences/${id}`);
  },
};
