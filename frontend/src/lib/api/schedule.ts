import apiClient from './client';

export interface StaffSchedule {
  id: string;
  staff_id: string;
  branch_id: string;
  date: string;
  is_working_day: boolean;
  created_by: string;
  created_at: string;
}

export interface CreateScheduleRequest {
  staff_id: string;
  branch_id: string;
  date: string;
  is_working_day: boolean;
}

export const scheduleApi = {
  getBranchSchedule: async (branchId: string, startDate?: string, endDate?: string) => {
    const response = await apiClient.get(`/schedules/branch/${branchId}`, {
      params: { start_date: startDate, end_date: endDate },
    });
    return (response.data.schedules || []) as StaffSchedule[];
  },
  
  create: async (data: CreateScheduleRequest) => {
    const response = await apiClient.post('/schedules', data);
    return response.data.schedule as StaffSchedule;
  },
  
  getMonthlyView: async (branchId: string, year: number, month: number) => {
    const response = await apiClient.get('/schedules/monthly', {
      params: { branch_id: branchId, year, month },
    });
    return (response.data.schedules || []) as StaffSchedule[];
  },
};


