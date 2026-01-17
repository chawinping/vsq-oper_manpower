import apiClient from './client';
import { BranchQuotaStatus } from './quota';

export interface DayOverview {
  date: string;
  branch_statuses: BranchQuotaStatus[];
  total_branches: number;
  branches_with_shortage: number;
}

export interface MonthlyOverview {
  branch_id: string;
  branch_name: string;
  branch_code: string;
  year: number;
  month: number;
  day_statuses: BranchQuotaStatus[];
  average_fulfillment: number;
}

export const overviewApi = {
  getDayOverview: async (date?: string): Promise<DayOverview> => {
    const params = new URLSearchParams();
    if (date) params.append('date', date);
    
    const response = await apiClient.get(`/overview/day?${params.toString()}`);
    return response.data.overview;
  },

  getMonthlyOverview: async (filters: {
    branch_id: string;
    year?: number;
    month?: number;
  }): Promise<MonthlyOverview> => {
    const params = new URLSearchParams();
    params.append('branch_id', filters.branch_id);
    if (filters.year) params.append('year', filters.year.toString());
    if (filters.month) params.append('month', filters.month.toString());
    
    const response = await apiClient.get(`/overview/monthly?${params.toString()}`);
    return response.data.overview;
  },
};
