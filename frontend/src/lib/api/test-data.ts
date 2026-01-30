import apiClient from './client';

export interface ScheduleRules {
  min_working_days_per_week: number;
  max_working_days_per_week: number;
  leave_probability: number; // 0.0-1.0
  consecutive_leave_max: number;
  weekend_working_ratio: number; // 0.0-1.0
  exclude_holidays: boolean;
  min_off_days_per_month: number;
  max_off_days_per_month: number;
  enforce_min_staff_per_group: boolean;
  branch_specific_rules?: Record<string, BranchRules>;
}

export interface BranchRules {
  min_working_days_per_week?: number;
  max_working_days_per_week?: number;
  leave_probability?: number;
  weekend_working_ratio?: number;
  min_off_days_per_month?: number;
  max_off_days_per_month?: number;
}

export interface GenerateScheduleRequest {
  start_date: string; // YYYY-MM-DD
  end_date: string; // YYYY-MM-DD
  rules: ScheduleRules;
  overwrite_existing: boolean;
  branch_ids?: string[]; // Optional: filter by specific branch IDs. If empty/undefined, generates for all branches.
}

export interface GenerateScheduleResult {
  total_staff: number;
  total_schedules: number;
  working_days: number;
  leave_days: number;
  off_days: number;
  errors?: string[];
}

export interface GenerateScheduleResponse {
  message: string;
  result: GenerateScheduleResult;
}

export const testDataApi = {
  generateSchedules: async (data: GenerateScheduleRequest): Promise<GenerateScheduleResponse> => {
    const response = await apiClient.post('/admin/test-data/generate-schedules', data);
    return response.data as GenerateScheduleResponse;
  },
};
