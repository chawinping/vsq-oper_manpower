import apiClient from './client';

// TODO: Implement report API functionality
// Related: FR-RP-04

export interface AllocationReport {
  id: string;
  iteration_id: string;
  start_date: string;
  end_date: string;
  branches_covered: number;
  total_assignments: number;
  total_positions_filled: number;
  total_positions_needed: number;
  overall_fulfillment_rate: number;
  average_confidence_score: number;
  created_at: string;
  created_by: string;
  assignment_details?: AllocationReportAssignment[];
  gap_analysis?: AllocationReportGap[];
}

export interface AllocationReportAssignment {
  id: string;
  rotation_staff_id: string;
  rotation_staff_name: string;
  branch_id: string;
  branch_name: string;
  branch_code: string;
  date: string;
  position_id: string;
  position_name: string;
  reason: string;
  criteria_used: string;
  confidence_score: number;
  is_overridden: boolean;
  override_reason?: string;
  override_by?: string;
  override_at?: string;
  status: 'approved' | 'rejected' | 'overridden';
}

export interface AllocationReportGap {
  branch_id: string;
  branch_name: string;
  branch_code: string;
  date: string;
  position_id: string;
  position_name: string;
  required_staff_count: number;
  available_local_staff: number;
  assigned_rotation_staff: number;
  still_required_staff: number;
}

export interface GenerateReportRequest {
  start_date: string;
  end_date: string;
  branch_ids?: string[];
  iteration_id: string;
}

export interface ReportFilters {
  start_date?: string;
  end_date?: string;
  branch_id?: string;
  position_id?: string;
  rotation_staff_id?: string;
  status?: string;
}

export const reportApi = {
  /**
   * Get all allocation reports with optional filtering
   * GET /api/reports
   */
  list: async (filters?: ReportFilters): Promise<AllocationReport[]> => {
    const params = new URLSearchParams();
    if (filters?.start_date) params.append('start_date', filters.start_date);
    if (filters?.end_date) params.append('end_date', filters.end_date);
    if (filters?.branch_id) params.append('branch_id', filters.branch_id);
    if (filters?.position_id) params.append('position_id', filters.position_id);
    if (filters?.rotation_staff_id) params.append('rotation_staff_id', filters.rotation_staff_id);
    if (filters?.status) params.append('status', filters.status);

    const queryString = params.toString();
    const url = queryString ? `/api/reports?${queryString}` : '/api/reports';
    
    const response = await apiClient.get(url);
    return response.data;
  },

  /**
   * Get a specific allocation report by ID
   * GET /api/reports/:id
   */
  get: async (id: string): Promise<AllocationReport> => {
    const response = await apiClient.get(`/api/reports/${id}`);
    return response.data;
  },

  /**
   * Generate a new allocation report for a specific iteration
   * POST /api/reports/generate
   */
  generate: async (request: GenerateReportRequest): Promise<AllocationReport> => {
    const response = await apiClient.post('/api/reports/generate', request);
    return response.data;
  },

  /**
   * Export a report to PDF or Excel
   * GET /api/reports/:id/export?format=pdf|excel
   */
  export: async (id: string, format: 'pdf' | 'excel'): Promise<Blob> => {
    const response = await apiClient.get(`/api/reports/${id}/export?format=${format}`, {
      responseType: 'blob',
    });
    return response.data;
  },
};
