import { apiClient } from './client';

export interface VersionInfo {
  version: string;
  buildDate: string;
  buildTime: string;
}

export interface VersionResponse {
  frontend: VersionInfo;
  backend: VersionInfo;
  database: VersionInfo;
}

export const versionApi = {
  /**
   * Get backend and database versions from API
   */
  getVersions: async (): Promise<VersionResponse> => {
    try {
      const response = await apiClient.get<VersionResponse>('/v1/version', {
        timeout: 5000, // 5 second timeout
      });
      return response.data;
    } catch (error: any) {
      console.error('Error fetching versions from API:', error);
      throw error;
    }
  },

  /**
   * Get frontend version from public VERSION.json
   */
  getFrontendVersion: async (): Promise<VersionInfo> => {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000); // 5 second timeout
      
      const response = await fetch('/VERSION.json', {
        signal: controller.signal,
      });
      
      clearTimeout(timeoutId);
      
      if (!response.ok) {
        throw new Error(`Failed to load frontend version: ${response.status} ${response.statusText}`);
      }
      
      return response.json();
    } catch (error: any) {
      console.error('Error fetching frontend version:', error);
      throw error;
    }
  },
};

