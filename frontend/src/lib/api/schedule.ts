import apiClient from './client';

export type ScheduleStatus = 'working' | 'off' | 'leave' | 'sick_leave';

export interface StaffSchedule {
  id: string;
  staff_id: string;
  branch_id: string;
  date: string;
  schedule_status: ScheduleStatus;
  is_working_day: boolean; // Deprecated: kept for backward compatibility
  created_by: string;
  created_at: string;
}

export interface CreateScheduleRequest {
  staff_id: string;
  branch_id: string;
  date: string;
  schedule_status?: ScheduleStatus;
  is_working_day?: boolean; // Deprecated: kept for backward compatibility
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


