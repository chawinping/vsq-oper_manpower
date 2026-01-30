import apiClient from './client';

export interface BranchType {
  id: string;
  name: string;
  description?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateBranchTypeRequest {
  name: string;
  description?: string;
  is_active?: boolean;
}

export interface UpdateBranchTypeRequest {
  name: string;
  description?: string;
  is_active: boolean;
}

export interface BranchTypeRequirement {
  id: string;
  branch_type_id: string;
  staff_group_id: string;
  day_of_week: number; // 0=Sunday, 1=Monday, ..., 6=Saturday
  minimum_staff_count: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  staff_group?: {
    id: string;
    name: string;
    description?: string;
  };
}

export interface StaffGroupRequirement {
  staff_group_id: string;
  minimum_count: number;
}

export interface BranchTypeConstraintStaffGroup {
  id: string;
  branch_type_constraint_id: string;
  staff_group_id: string;
  minimum_count: number;
  created_at: string;
  updated_at: string;
  staff_group?: {
    id: string;
    name: string;
    description?: string;
  };
}

export interface BranchTypeConstraints {
  id: string;
  branch_type_id: string;
  day_of_week: number; // 0=Sunday, 1=Monday, ..., 6=Saturday
  min_front_staff?: number; // Deprecated
  min_managers?: number; // Deprecated
  min_doctor_assistant?: number; // Deprecated
  min_total_staff?: number; // Deprecated
  created_at: string;
  updated_at: string;
  staff_group_requirements?: BranchTypeConstraintStaffGroup[];
}

export interface ConstraintsUpdate {
  day_of_week: number;
  staff_group_requirements: StaffGroupRequirement[];
  // Deprecated fields (kept for backward compatibility)
  min_front_staff?: number;
  min_managers?: number;
  min_doctor_assistant?: number;
  min_total_staff?: number;
}

export const branchTypeApi = {
  list: async () => {
    const response = await apiClient.get('/branch-types');
    return (response.data.branch_types || []) as BranchType[];
  },
  
  getById: async (id: string) => {
    const response = await apiClient.get(`/branch-types/${id}`);
    return response.data.branch_type as BranchType;
  },
  
  create: async (data: CreateBranchTypeRequest) => {
    const response = await apiClient.post('/branch-types', data);
    return response.data.branch_type as BranchType;
  },
  
  update: async (id: string, data: UpdateBranchTypeRequest) => {
    const response = await apiClient.put(`/branch-types/${id}`, data);
    return response.data.branch_type as BranchType;
  },
  
  delete: async (id: string) => {
    const response = await apiClient.delete(`/branch-types/${id}`);
    return response.data;
  },
  
  getRequirements: async (branchTypeId: string) => {
    const response = await apiClient.get(`/branch-types/${branchTypeId}/requirements`);
    return (response.data.requirements || []) as BranchTypeRequirement[];
  },
  
  createRequirement: async (branchTypeId: string, data: {
    staff_group_id: string;
    day_of_week: number;
    minimum_staff_count: number;
    is_active?: boolean;
  }) => {
    const response = await apiClient.post(`/branch-types/${branchTypeId}/requirements`, data);
    return response.data.requirement as BranchTypeRequirement;
  },
  
  bulkUpsertRequirements: async (branchTypeId: string, requirements: {
    staff_group_id: string;
    day_of_week: number;
    minimum_staff_count: number;
    is_active: boolean;
  }[]) => {
    const response = await apiClient.put(`/branch-types/${branchTypeId}/requirements/bulk`, {
      requirements,
    });
    return (response.data.requirements || []) as BranchTypeRequirement[];
  },
  
  updateRequirement: async (requirementId: string, data: {
    minimum_staff_count: number;
    is_active: boolean;
  }) => {
    const response = await apiClient.put(`/branch-type-requirements/${requirementId}`, data);
    return response.data.requirement as BranchTypeRequirement;
  },
  
  deleteRequirement: async (requirementId: string) => {
    const response = await apiClient.delete(`/branch-type-requirements/${requirementId}`);
    return response.data;
  },
  
  getConstraints: async (branchTypeId: string) => {
    const response = await apiClient.get(`/branch-types/${branchTypeId}/constraints`);
    return (response.data.constraints || []) as BranchTypeConstraints[];
  },
  
  updateConstraints: async (branchTypeId: string, constraints: ConstraintsUpdate[]) => {
    console.log('[API] updateConstraints called:', { branchTypeId, constraintsCount: constraints.length, constraints });
    try {
      const url = `/branch-types/${branchTypeId}/constraints`;
      const payload = { constraints };
      console.log('[API] Making PUT request to:', url);
      console.log('[API] Request payload:', JSON.stringify(payload, null, 2));
      
      const response = await apiClient.put(url, payload);
      
      console.log('[API] Response received:', { status: response.status, data: response.data });
      return (response.data.constraints || []) as BranchTypeConstraints[];
    } catch (error: any) {
      console.error('[API] updateConstraints error:', error);
      console.error('[API] Error response:', error.response);
      console.error('[API] Error response data:', error.response?.data);
      console.error('[API] Error response status:', error.response?.status);
      throw error;
    }
  },
};
