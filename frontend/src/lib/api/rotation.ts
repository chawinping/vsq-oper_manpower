import apiClient from './client';

export interface RotationAssignment {
  id: string;
  rotation_staff_id: string;
  branch_id: string;
  date: string;
  assignment_level: 1 | 2;
  schedule_status?: 'working' | 'off' | 'leave' | 'sick_leave';
  assigned_by: string;
  created_at: string;
}

export interface AssignRotationRequest {
  rotation_staff_id: string;
  branch_id: string;
  date: string;
  assignment_level: 1 | 2;
  schedule_status?: 'working' | 'off' | 'leave' | 'sick_leave';
}

export interface EligibleStaff {
  id: string;
  nickname?: string;
  name: string;
  staff_type: 'branch' | 'rotation';
  position_id: string;
  position?: {
    id: string;
    name: string;
  };
  branch_id?: string;
  coverage_area?: string;
  skill_level: number;
  assignment_level: 1 | 2;
  created_at: string;
  updated_at: string;
}

export interface BulkAssignRequest {
  branch_id: string;
  assignments: {
    rotation_staff_id: string;
    dates: string[];
    assignment_level: 1 | 2;
  }[];
}

export interface BulkAssignResponse {
  created: number;
  assignments: RotationAssignment[];
  errors: string[];
}

export interface AssignmentSuggestion {
  rotation_staff_id: string;
  branch_id: string;
  date: string;
  assignment_level: 1 | 2;
  confidence: number;
  reason: string;
}

export interface SuggestionsResponse {
  suggestions: AssignmentSuggestion[];
}

export const rotationApi = {
  getAssignments: async (filters?: {
    branch_id?: string;
    rotation_staff_id?: string;
    start_date?: string;
    end_date?: string;
    coverage_area?: string;
  }) => {
    const response = await apiClient.get('/rotation/assignments', { params: filters });
    return (response.data.assignments || []) as RotationAssignment[];
  },
  
  assign: async (data: AssignRotationRequest) => {
    const response = await apiClient.post('/rotation/assign', data);
    return response.data.assignment as RotationAssignment;
  },

  bulkAssign: async (data: BulkAssignRequest) => {
    const response = await apiClient.post('/rotation/bulk-assign', data);
    return response.data as BulkAssignResponse;
  },
  
  updateAssignmentStatus: async (assignmentId: string, scheduleStatus: 'working' | 'off' | 'leave' | 'sick_leave') => {
    const response = await apiClient.patch(`/rotation/assign/${assignmentId}/status`, {
      schedule_status: scheduleStatus,
    });
    return response.data.assignment as RotationAssignment;
  },
  
  removeAssignment: async (id: string) => {
    const response = await apiClient.delete(`/rotation/assign/${id}`);
    return response.data;
  },

  getEligibleStaff: async (branchId: string) => {
    const response = await apiClient.get(`/rotation/eligible-staff/${branchId}`);
    return (response.data.eligible_staff || []) as EligibleStaff[];
  },
  
  getSuggestions: async (filters?: {
    branch_id?: string;
    start_date?: string;
    end_date?: string;
  }) => {
    const response = await apiClient.get('/rotation/suggestions', { params: filters });
    const data = response.data as SuggestionsResponse;
    return {
      suggestions: data?.suggestions || []
    } as SuggestionsResponse;
  },
  
  regenerateSuggestions: async (filters?: {
    branch_id?: string;
    start_date?: string;
    end_date?: string;
  }) => {
    const response = await apiClient.post('/rotation/regenerate-suggestions', filters);
    const data = response.data as SuggestionsResponse;
    return {
      suggestions: data?.suggestions || []
    } as SuggestionsResponse;
  },
};
