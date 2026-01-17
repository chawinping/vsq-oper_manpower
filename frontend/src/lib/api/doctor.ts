import apiClient from './client';

export interface Doctor {
  id: string;
  name: string;
  code?: string;
  preferences?: string; // Noted remark/preferences
  created_at: string;
  updated_at: string;
}

export interface DoctorPreference {
  id: string;
  doctor_id: string;
  doctor?: Doctor;
  branch_id?: string;
  branch?: { id: string; name: string; code: string };
  rule_type: string;
  rule_config: Record<string, any>;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface DoctorAssignment {
  id: string;
  doctor_id: string;
  doctor?: Doctor;
  doctor_name: string;
  doctor_code?: string;
  branch_id: string;
  branch?: { id: string; name: string; code: string };
  date: string;
  expected_revenue: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface DoctorOnOffDay {
  id: string;
  branch_id: string;
  branch?: { id: string; name: string; code: string };
  date: string;
  is_doctor_on: boolean;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateDoctorRequest {
  name: string;
  code?: string;
  preferences?: string; // Noted remark/preferences
}

export interface UpdateDoctorRequest {
  name?: string;
  code?: string;
  preferences?: string; // Noted remark/preferences
}

export interface CreateDoctorAssignmentRequest {
  doctor_id: string;
  branch_id: string;
  date: string;
  expected_revenue?: number;
}

export interface CreateDoctorPreferenceRequest {
  doctor_id: string;
  branch_id?: string;
  rule_type: string;
  rule_config: Record<string, any>;
  is_active?: boolean;
}

export interface CreateDoctorOnOffDayRequest {
  branch_id: string;
  date: string;
  is_doctor_on: boolean;
}

export interface DoctorDefaultSchedule {
  id: string;
  doctor_id: string;
  doctor?: Doctor;
  day_of_week: number; // 0=Sunday, 6=Saturday
  branch_id: string;
  branch?: { id: string; name: string; code: string };
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface DoctorWeeklyOffDay {
  id: string;
  doctor_id: string;
  doctor?: Doctor;
  day_of_week: number; // 0=Sunday, 6=Saturday
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface DoctorScheduleOverride {
  id: string;
  doctor_id: string;
  doctor?: Doctor;
  date: string;
  type: 'working' | 'off';
  branch_id?: string;
  branch?: { id: string; name: string; code: string };
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateDoctorDefaultScheduleRequest {
  doctor_id: string;
  day_of_week: number; // 0=Sunday, 6=Saturday
  branch_id: string;
}

export interface UpdateDoctorDefaultScheduleRequest {
  branch_id: string;
}

export interface CreateDoctorWeeklyOffDayRequest {
  doctor_id: string;
  day_of_week: number; // 0=Sunday, 6=Saturday
}

export interface CreateDoctorScheduleOverrideRequest {
  doctor_id: string;
  date: string;
  type: 'working' | 'off';
  branch_id?: string; // Required if type is 'working'
}

export const doctorApi = {
  // Doctor CRUD
  list: async (): Promise<Doctor[]> => {
    const response = await apiClient.get('/doctors');
    return response.data.doctors || [];
  },

  getById: async (id: string): Promise<Doctor> => {
    const response = await apiClient.get(`/doctors/${id}`);
    return response.data.doctor;
  },

  create: async (data: CreateDoctorRequest): Promise<Doctor> => {
    const response = await apiClient.post('/doctors', data);
    return response.data.doctor;
  },

  update: async (id: string, data: UpdateDoctorRequest): Promise<Doctor> => {
    const response = await apiClient.put(`/doctors/${id}`, data);
    return response.data.doctor;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/doctors/${id}`);
  },

  // Doctor Schedule
  getMonthlySchedule: async (doctorId: string, year?: number, month?: number): Promise<{ assignments: DoctorAssignment[]; year: number; month: number }> => {
    const params = new URLSearchParams();
    if (year) params.append('year', year.toString());
    if (month) params.append('month', month.toString());
    const response = await apiClient.get(`/doctors/${doctorId}/schedule?${params.toString()}`);
    return response.data;
  },

  getAssignments: async (filters?: {
    branch_id?: string;
    doctor_id?: string;
    start_date?: string;
    end_date?: string;
  }): Promise<DoctorAssignment[]> => {
    const params = new URLSearchParams();
    if (filters?.branch_id) params.append('branch_id', filters.branch_id);
    if (filters?.doctor_id) params.append('doctor_id', filters.doctor_id);
    if (filters?.start_date) params.append('start_date', filters.start_date);
    if (filters?.end_date) params.append('end_date', filters.end_date);
    
    const response = await apiClient.get(`/doctors/assignments?${params.toString()}`);
    return response.data.assignments || [];
  },

  createAssignment: async (data: CreateDoctorAssignmentRequest): Promise<DoctorAssignment> => {
    const response = await apiClient.post('/doctors/assignments', data);
    return response.data.assignment;
  },

  deleteAssignment: async (id: string): Promise<void> => {
    await apiClient.delete(`/doctors/assignments/${id}`);
  },

  // Doctor Preferences
  getPreferences: async (doctorId: string, branchId?: string): Promise<DoctorPreference[]> => {
    const params = new URLSearchParams();
    params.append('doctor_id', doctorId);
    if (branchId) params.append('branch_id', branchId);
    const response = await apiClient.get(`/doctors/preferences?${params.toString()}`);
    return response.data.preferences || [];
  },

  createPreference: async (data: CreateDoctorPreferenceRequest): Promise<DoctorPreference> => {
    const response = await apiClient.post('/doctors/preferences', data);
    return response.data.preference;
  },

  updatePreference: async (id: string, data: CreateDoctorPreferenceRequest): Promise<DoctorPreference> => {
    const response = await apiClient.put(`/doctors/preferences/${id}`, data);
    return response.data.preference;
  },

  deletePreference: async (id: string): Promise<void> => {
    await apiClient.delete(`/doctors/preferences/${id}`);
  },

  getDoctorOnOffDays: async (filters?: {
    branch_id: string;
    start_date?: string;
    end_date?: string;
  }): Promise<DoctorOnOffDay[]> => {
    const params = new URLSearchParams();
    if (filters?.branch_id) params.append('branch_id', filters.branch_id);
    if (filters?.start_date) params.append('start_date', filters.start_date);
    if (filters?.end_date) params.append('end_date', filters.end_date);
    
    const response = await apiClient.get(`/doctors/on-off-days?${params.toString()}`);
    return response.data.days || [];
  },

  createDoctorOnOffDay: async (data: CreateDoctorOnOffDayRequest): Promise<DoctorOnOffDay> => {
    const response = await apiClient.post('/doctors/on-off-days', data);
    return response.data.day;
  },

  deleteDoctorOnOffDay: async (id: string): Promise<void> => {
    await apiClient.delete(`/doctors/on-off-days/${id}`);
  },

  // Doctor Default Schedules
  getDefaultSchedules: async (doctorId: string): Promise<DoctorDefaultSchedule[]> => {
    const params = new URLSearchParams();
    params.append('doctor_id', doctorId);
    const response = await apiClient.get(`/doctors/default-schedules?${params.toString()}`);
    return response.data.schedules || [];
  },

  createDefaultSchedule: async (data: CreateDoctorDefaultScheduleRequest): Promise<DoctorDefaultSchedule> => {
    const response = await apiClient.post('/doctors/default-schedules', data);
    return response.data.schedule;
  },

  updateDefaultSchedule: async (id: string, data: UpdateDoctorDefaultScheduleRequest): Promise<DoctorDefaultSchedule> => {
    const response = await apiClient.put(`/doctors/default-schedules/${id}`, data);
    return response.data.schedule;
  },

  deleteDefaultSchedule: async (id: string): Promise<void> => {
    await apiClient.delete(`/doctors/default-schedules/${id}`);
  },

  importDefaultSchedules: async (file: File): Promise<{
    message: string;
    imported: number;
    off_days_set?: number;
    total_processed?: number;
    schedules: DoctorDefaultSchedule[];
    parse_warnings?: string;
    save_warnings?: string[];
  }> => {
    const formData = new FormData();
    formData.append('file', file);
    
    const response = await apiClient.post('/doctors/default-schedules/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  // Doctor Weekly Off Days
  getWeeklyOffDays: async (doctorId: string): Promise<DoctorWeeklyOffDay[]> => {
    const params = new URLSearchParams();
    params.append('doctor_id', doctorId);
    const response = await apiClient.get(`/doctors/weekly-off-days?${params.toString()}`);
    return response.data.off_days || [];
  },

  createWeeklyOffDay: async (data: CreateDoctorWeeklyOffDayRequest): Promise<DoctorWeeklyOffDay> => {
    const response = await apiClient.post('/doctors/weekly-off-days', data);
    return response.data.off_day;
  },

  deleteWeeklyOffDay: async (id: string): Promise<void> => {
    await apiClient.delete(`/doctors/weekly-off-days/${id}`);
  },

  // Doctor Schedule Overrides
  getScheduleOverrides: async (filters?: {
    doctor_id: string;
    start_date?: string;
    end_date?: string;
  }): Promise<DoctorScheduleOverride[]> => {
    const params = new URLSearchParams();
    if (filters?.doctor_id) params.append('doctor_id', filters.doctor_id);
    if (filters?.start_date) params.append('start_date', filters.start_date);
    if (filters?.end_date) params.append('end_date', filters.end_date);
    
    const response = await apiClient.get(`/doctors/schedule-overrides?${params.toString()}`);
    return response.data.overrides || [];
  },

  createScheduleOverride: async (data: CreateDoctorScheduleOverrideRequest): Promise<DoctorScheduleOverride> => {
    const response = await apiClient.post('/doctors/schedule-overrides', data);
    return response.data.override;
  },

  updateScheduleOverride: async (id: string, data: CreateDoctorScheduleOverrideRequest): Promise<DoctorScheduleOverride> => {
    const response = await apiClient.put(`/doctors/schedule-overrides/${id}`, data);
    return response.data.override;
  },

  deleteScheduleOverride: async (id: string): Promise<void> => {
    await apiClient.delete(`/doctors/schedule-overrides/${id}`);
  },

  // Import doctors from Excel
  import: async (file: File): Promise<{
    message: string;
    imported: number;
    total_rows: number;
    doctors: Doctor[];
    parse_warnings?: string;
    save_warnings?: string[];
  }> => {
    const formData = new FormData();
    formData.append('file', file);
    
    const response = await apiClient.post('/doctors/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },
};
