import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081/api';

export const apiClient = axios.create({
  baseURL: API_URL,
  withCredentials: true,
  timeout: 10000, // 10 second timeout to prevent hanging requests
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor - Add request ID
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // Add request ID header if not already present (backend will generate if missing)
    if (!config.headers['X-Request-ID'] && typeof crypto !== 'undefined' && crypto.randomUUID) {
      config.headers['X-Request-ID'] = crypto.randomUUID();
    }
    
    // Log request in development
    if (process.env.NODE_ENV === 'development') {
      console.log('[API Client] Request:', {
        url: config.url,
        method: config.method?.toUpperCase(),
        baseURL: config.baseURL,
        data: config.data,
        headers: config.headers,
      });
    }
    
    return config;
  },
  (error) => {
    console.error('[API Client] Request error:', error);
    return Promise.reject(error);
  }
);

// Response interceptor
apiClient.interceptors.response.use(
  (response) => {
    // Log successful response in development
    if (process.env.NODE_ENV === 'development') {
      console.log('[API Client] Response:', {
        url: response.config.url,
        method: response.config.method?.toUpperCase(),
        status: response.status,
        data: response.data,
      });
    }
    return response;
  },
  (error: AxiosError) => {
    // Handle 401 Unauthorized - redirect to login
    if (error.response?.status === 401) {
      if (typeof window !== 'undefined' && !window.location.pathname.includes('/login')) {
        window.location.href = '/login';
      }
    }
    
    // Log error (always log, not just in development)
    console.error('[API Client] Response error:', {
      url: error.config?.url,
      method: error.config?.method?.toUpperCase(),
      status: error.response?.status,
      statusText: error.response?.statusText,
      data: error.response?.data,
      message: error.message,
      code: error.code,
      requestId: error.response?.headers['x-request-id'],
    });
    
    return Promise.reject(error);
  }
);

export default apiClient;



