import apiClient from './client';
import { Position } from './position';
import { RevenueLevelTier } from './revenue-level-tier';

export interface ScenarioPositionRequirement {
  id: string;
  scenario_id: string;
  position_id: string;
  position?: Position;
  preferred_staff: number;
  minimum_staff: number;
  override_base: boolean;
  created_at: string;
  updated_at: string;
}

export interface ScenarioPositionRequirementCreate {
  position_id: string;
  preferred_staff: number;
  minimum_staff: number;
  override_base: boolean;
}

export interface ScenarioSpecificStaffRequirement {
  id: string;
  scenario_id: string;
  staff_id: string;
  staff?: {
    id: string;
    name: string;
    nickname: string;
  };
  created_at: string;
  updated_at: string;
}

export interface ScenarioSpecificStaffRequirementCreate {
  staff_id: string;
}

export interface StaffRequirementScenario {
  id: string;
  scenario_name: string;
  description: string | null;
  doctor_id: string | null;
  doctor?: {
    id: string;
    name: string;
    code: string | null;
  };
  branch_id: string | null;
  branch?: {
    id: string;
    name: string;
    code: string;
  };
  revenue_level_tier_id: string | null;
  revenue_level_tier?: RevenueLevelTier;
  min_revenue: number | null;
  max_revenue: number | null;
  use_day_of_week_revenue: boolean;
  use_specific_date_revenue: boolean;
  doctor_count: number | null;
  min_doctor_count: number | null;
  day_of_week: number | null;
  is_default: boolean;
  is_active: boolean;
  priority: number;
  created_at: string;
  updated_at: string;
  position_requirements?: ScenarioPositionRequirement[];
  specific_staff_requirements?: ScenarioSpecificStaffRequirement[];
}

export interface StaffRequirementScenarioCreate {
  scenario_name: string;
  description?: string | null;
  doctor_id?: string | null;
  branch_id?: string | null;
  revenue_level_tier_id?: string | null;
  min_revenue?: number | null;
  max_revenue?: number | null;
  use_day_of_week_revenue?: boolean;
  use_specific_date_revenue?: boolean;
  doctor_count?: number | null;
  min_doctor_count?: number | null;
  day_of_week?: number | null;
  is_default?: boolean;
  is_active?: boolean;
  priority?: number;
  position_requirements?: ScenarioPositionRequirementCreate[];
  specific_staff_requirements?: ScenarioSpecificStaffRequirementCreate[];
}

export interface StaffRequirementScenarioUpdate {
  scenario_name?: string;
  description?: string | null;
  doctor_id?: string | null;
  branch_id?: string | null;
  revenue_level_tier_id?: string | null;
  min_revenue?: number | null;
  max_revenue?: number | null;
  use_day_of_week_revenue?: boolean;
  use_specific_date_revenue?: boolean;
  doctor_count?: number | null;
  min_doctor_count?: number | null;
  day_of_week?: number | null;
  is_default?: boolean;
  is_active?: boolean;
  priority?: number;
}

export interface CalculatedRequirement {
  position_id: string;
  position_name: string;
  base_preferred: number;
  base_minimum: number;
  calculated_preferred: number;
  calculated_minimum: number;
  matched_scenario_id?: string | null;
  matched_scenario_name?: string | null;
  factors_applied: string[];
}

export interface ScenarioMatch {
  scenario_id: string;
  scenario_name: string;
  matches: boolean;
  match_reason: string;
  priority: number;
}

export const staffRequirementScenarioApi = {
  list: async (includeInactive?: boolean) => {
    const params = includeInactive ? { include_inactive: 'true' } : {};
    const response = await apiClient.get('/staff-requirement-scenarios', { params });
    return response.data as StaffRequirementScenario[];
  },

  getById: async (id: string) => {
    const response = await apiClient.get(`/staff-requirement-scenarios/${id}`);
    return response.data as StaffRequirementScenario;
  },

  create: async (data: StaffRequirementScenarioCreate) => {
    const response = await apiClient.post('/staff-requirement-scenarios', data);
    return response.data as StaffRequirementScenario;
  },

  update: async (id: string, data: StaffRequirementScenarioUpdate) => {
    const response = await apiClient.put(`/staff-requirement-scenarios/${id}`, data);
    return response.data as StaffRequirementScenario;
  },

  delete: async (id: string) => {
    const response = await apiClient.delete(`/staff-requirement-scenarios/${id}`);
    return response.data;
  },

  updatePositionRequirements: async (id: string, requirements: ScenarioPositionRequirementCreate[]) => {
    const response = await apiClient.put(`/staff-requirement-scenarios/${id}/position-requirements`, {
      requirements,
    });
    return response.data;
  },

  updateSpecificStaffRequirements: async (id: string, requirements: ScenarioSpecificStaffRequirementCreate[]) => {
    const response = await apiClient.put(`/staff-requirement-scenarios/${id}/specific-staff-requirements`, {
      requirements,
    });
    return response.data;
  },

  calculateRequirements: async (data: {
    branch_id: string;
    date: string;
    position_id: string;
    base_preferred?: number;
    base_minimum?: number;
  }) => {
    const response = await apiClient.post('/staff-requirement-scenarios/calculate', data);
    return response.data as CalculatedRequirement;
  },

  getMatchingScenarios: async (branchId: string, date: string) => {
    const response = await apiClient.post('/staff-requirement-scenarios/match', {
      branch_id: branchId,
      date,
    });
    return response.data as ScenarioMatch[];
  },
};
