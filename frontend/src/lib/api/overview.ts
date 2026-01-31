import apiClient from './client';

export interface PositionQuotaStatus {
  position_id: string;
  position_name: string;
  designated_quota: number; // Preferred staff count
  minimum_required: number; // Minimum staff count
  available_local: number; // Local branch staff available
  assigned_rotation: number; // Rotation staff assigned
  total_assigned: number; // Total staff (local + rotation)
  still_required: number; // Staff still needed
}

export interface DoctorInfo {
  id: string;
  name: string;
  code: string;
}

export interface BranchQuotaStatus {
  branch_id: string;
  branch_name: string;
  branch_code: string;
  date: string;
  position_statuses: PositionQuotaStatus[];
  total_designated: number;
  total_available: number;
  total_assigned: number;
  total_required: number;
  // Doctors assigned to this branch on this date
  doctors: DoctorInfo[];
  // Scoring group points and missing staff
  group1_score: number; // Position Quota - Minimum Shortage
  group2_score: number; // Daily Staff Constraints - Minimum Shortage
  group3_score: number; // Position Quota - Preferred Excess
  group1_missing_staff: string[]; // Staff nicknames who don't work (Group 1)
  group2_missing_staff: string[]; // Staff nicknames who don't work (Group 2)
  group3_missing_staff: string[]; // Staff nicknames who don't work (Group 3)
}

export interface DayOverview {
  date: string;
  branch_statuses: BranchQuotaStatus[];
  total_branches: number;
  branches_with_shortage: number;
}

export const overviewApi = {
  getDayOverview: async (date: string, branchIds?: string[]) => {
    const params: { date: string; branch_ids?: string } = { date };
    if (branchIds && branchIds.length > 0) {
      params.branch_ids = branchIds.join(',');
    }
    const response = await apiClient.get('/overview/day', { params });
    return response.data.overview as DayOverview;
  },
  
  getBranchQuotaStatus: async (branchId: string, date: string) => {
    const overview = await overviewApi.getDayOverview(date, [branchId]);
    return overview.branch_statuses.find(bs => bs.branch_id === branchId);
  },
};
