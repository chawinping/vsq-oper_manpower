import apiClient from './client';

// Criterion ID constants
export const CRITERION_ZEROTH = 'zeroth_criteria';  // Doctor preferences
export const CRITERION_FIRST = 'first_criteria';    // Branch-level variables
export const CRITERION_SECOND = 'second_criteria';  // Preferred staff shortage
export const CRITERION_THIRD = 'third_criteria';    // Minimum staff shortage
export const CRITERION_FOURTH = 'fourth_criteria';  // Branch type staff groups

export type CriterionID = 
  | typeof CRITERION_ZEROTH
  | typeof CRITERION_FIRST
  | typeof CRITERION_SECOND
  | typeof CRITERION_THIRD
  | typeof CRITERION_FOURTH;

export interface AllocationCriteriaConfig {
  priority_order: CriterionID[];  // Array of criterion IDs in priority order (highest to lowest)
  enable_doctor_preferences: boolean;
}

export const allocationCriteriaApi = {
  getPriorityOrder: async (): Promise<AllocationCriteriaConfig> => {
    const response = await apiClient.get('/allocation-criteria/priority-order');
    return response.data;
  },

  updatePriorityOrder: async (config: AllocationCriteriaConfig): Promise<AllocationCriteriaConfig> => {
    const response = await apiClient.put('/allocation-criteria/priority-order', config);
    return response.data;
  },

  resetPriorityOrder: async (): Promise<AllocationCriteriaConfig> => {
    const response = await apiClient.post('/allocation-criteria/priority-order/reset');
    return response.data;
  },
};
