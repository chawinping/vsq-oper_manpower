import apiClient from './client';
import { Position } from './position';

export type ClinicPreferenceCriteriaType =
  | 'skin_revenue'
  | 'laser_yag_revenue'
  | 'iv_cases'
  | 'slim_pen_cases'
  | 'doctor_count';

export interface PreferencePositionRequirement {
  id: string;
  preference_id: string;
  position_id: string;
  position?: Position;
  minimum_staff: number;
  preferred_staff: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ClinicWidePreference {
  id: string;
  criteria_type: ClinicPreferenceCriteriaType;
  criteria_name: string;
  min_value: number;
  max_value: number | null;
  is_active: boolean;
  display_order: number;
  description: string | null;
  position_requirements?: PreferencePositionRequirement[];
  created_at: string;
  updated_at: string;
}

export interface ClinicWidePreferenceCreate {
  criteria_type: ClinicPreferenceCriteriaType;
  criteria_name: string;
  min_value: number;
  max_value?: number | null;
  is_active?: boolean;
  display_order?: number;
  description?: string | null;
  position_requirements?: PreferencePositionRequirementCreate[];
}

export interface PreferencePositionRequirementCreate {
  position_id: string;
  minimum_staff: number;
  preferred_staff: number;
  is_active?: boolean;
}

export interface ClinicWidePreferenceUpdate {
  criteria_name?: string;
  min_value?: number;
  max_value?: number | null;
  is_active?: boolean;
  display_order?: number;
  description?: string | null;
}

export interface PreferencePositionRequirementUpdate {
  minimum_staff?: number;
  preferred_staff?: number;
  is_active?: boolean;
}

export interface ClinicPreferenceFilters {
  criteria_type?: ClinicPreferenceCriteriaType;
  is_active?: boolean;
}

export const clinicPreferenceApi = {
  list: async (filters?: ClinicPreferenceFilters) => {
    const params = new URLSearchParams();
    if (filters?.criteria_type) {
      params.append('criteria_type', filters.criteria_type);
    }
    if (filters?.is_active !== undefined) {
      params.append('is_active', filters.is_active.toString());
    }
    const queryString = params.toString();
    const url = queryString ? `/clinic-preferences?${queryString}` : '/clinic-preferences';
    const response = await apiClient.get(url);
    return response.data as ClinicWidePreference[];
  },

  getById: async (id: string) => {
    const response = await apiClient.get(`/clinic-preferences/${id}`);
    return response.data as ClinicWidePreference;
  },

  create: async (data: ClinicWidePreferenceCreate) => {
    const response = await apiClient.post('/clinic-preferences', data);
    return response.data as ClinicWidePreference;
  },

  update: async (id: string, data: ClinicWidePreferenceUpdate) => {
    const response = await apiClient.put(`/clinic-preferences/${id}`, data);
    return response.data as ClinicWidePreference;
  },

  delete: async (id: string) => {
    const response = await apiClient.delete(`/clinic-preferences/${id}`);
    return response.data;
  },

  addPositionRequirement: async (preferenceId: string, data: PreferencePositionRequirementCreate) => {
    const response = await apiClient.post(`/clinic-preferences/${preferenceId}/positions`, data);
    return response.data as PreferencePositionRequirement;
  },

  updatePositionRequirement: async (
    preferenceId: string,
    positionId: string,
    data: PreferencePositionRequirementUpdate
  ) => {
    const response = await apiClient.put(`/clinic-preferences/${preferenceId}/positions/${positionId}`, data);
    return response.data as PreferencePositionRequirement;
  },

  deletePositionRequirement: async (preferenceId: string, positionId: string) => {
    const response = await apiClient.delete(`/clinic-preferences/${preferenceId}/positions/${positionId}`);
    return response.data;
  },

  getByCriteriaAndValue: async (criteriaType: ClinicPreferenceCriteriaType, value: number) => {
    const response = await apiClient.get(`/clinic-preferences/${criteriaType}/match?value=${value}`);
    return response.data as ClinicWidePreference[];
  },
};
