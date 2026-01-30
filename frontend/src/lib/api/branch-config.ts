import apiClient from './client';

export interface PositionQuota {
  id?: string;
  position_id: string;
  position_name: string;
  designated_quota: number;
  minimum_required: number;
  is_active?: boolean;
}

export interface WeeklyRevenue {
  id?: string;
  day_of_week: number; // 0=Sunday, 1=Monday, ..., 6=Saturday
  expected_revenue?: number; // Deprecated: Use skin_revenue instead
  skin_revenue: number; // Skin revenue (THB)
  ls_hm_revenue: number; // LS HM revenue (THB)
  vitamin_cases: number; // Vitamin cases (count)
  slim_pen_cases: number; // Slim Pen cases (count)
}

export interface BranchConfig {
  quotas: PositionQuota[];
  weekly_revenue: WeeklyRevenue[];
}

export interface UpdateQuotasRequest {
  quotas: PositionQuotaUpdate[];
}

export interface PositionQuotaUpdate {
  position_id: string;
  designated_quota: number;
  minimum_required: number;
}

export interface UpdateWeeklyRevenueRequest {
  weekly_revenue: WeeklyRevenueUpdate[];
}

export interface WeeklyRevenueUpdate {
  day_of_week: number;
  expected_revenue?: number; // Deprecated: Use skin_revenue instead
  skin_revenue: number; // Skin revenue (THB)
  ls_hm_revenue: number; // LS HM revenue (THB)
  vitamin_cases: number; // Vitamin cases (count)
  slim_pen_cases: number; // Slim Pen cases (count)
}

export interface StaffGroupRequirement {
  staff_group_id: string;
  minimum_count: number;
}

export interface BranchConstraints {
  id?: string;
  branch_id: string;
  day_of_week: number; // 0=Sunday, 1=Monday, ..., 6=Saturday
  min_front_staff?: number; // DEPRECATED: Use staff_group_requirements instead
  min_managers?: number; // DEPRECATED: Use staff_group_requirements instead
  min_doctor_assistant?: number; // DEPRECATED: Use staff_group_requirements instead
  min_total_staff?: number; // DEPRECATED: Use staff_group_requirements instead
  inherited_from_branch_type_id?: string; // If set, this constraint is inherited from a branch type
  is_overridden: boolean; // If true, this constraint overrides the branch type default
  staff_group_requirements?: StaffGroupRequirement[]; // Staff group requirements
  created_at?: string;
  updated_at?: string;
}

export interface ConstraintsUpdate {
  day_of_week: number;
  staff_group_requirements: StaffGroupRequirement[];
}

const DAY_NAMES = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

export const branchConfigApi = {
  getConfig: async (branchId: string): Promise<BranchConfig> => {
    const response = await apiClient.get(`/branches/${branchId}/config`);
    return response.data as BranchConfig;
  },

  getQuotas: async (branchId: string): Promise<PositionQuota[]> => {
    const response = await apiClient.get(`/branches/${branchId}/config/quotas`);
    return (response.data.quotas || []) as PositionQuota[];
  },

  updateQuotas: async (branchId: string, quotas: PositionQuotaUpdate[]): Promise<void> => {
    await apiClient.put(`/branches/${branchId}/config/quotas`, { quotas });
  },

  getWeeklyRevenue: async (branchId: string): Promise<WeeklyRevenue[]> => {
    const response = await apiClient.get(`/branches/${branchId}/config/weekly-revenue`);
    return (response.data.weekly_revenue || []) as WeeklyRevenue[];
  },

  updateWeeklyRevenue: async (branchId: string, weeklyRevenue: WeeklyRevenueUpdate[]): Promise<void> => {
    await apiClient.put(`/branches/${branchId}/config/weekly-revenue`, { weekly_revenue: weeklyRevenue });
  },

  getDayName: (dayOfWeek: number): string => {
    return DAY_NAMES[dayOfWeek] || `Day ${dayOfWeek}`;
  },

  getConstraints: async (branchId: string): Promise<BranchConstraints[]> => {
    const response = await apiClient.get(`/branches/${branchId}/config/constraints`);
    return (response.data.constraints || []) as BranchConstraints[];
  },

  updateConstraints: async (branchId: string, constraints: ConstraintsUpdate[]): Promise<void> => {
    await apiClient.put(`/branches/${branchId}/config/constraints`, { constraints });
  },
};
