import apiClient from './client';

export interface RotationAssignment {
  id: string;
  rotation_staff_id: string;
  branch_id: string;
  date: string;
  assignment_level: 1 | 2;
  assigned_by: string;
  created_at: string;
}

export interface AssignRotationRequest {
  rotation_staff_id: string;
  branch_id: string;
  date: string;
  assignment_level: 1 | 2;
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
  
  removeAssignment: async (id: string) => {
    const response = await apiClient.delete(`/rotation/assign/${id}`);
    return response.data;
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

