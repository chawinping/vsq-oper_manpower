import { AxiosError } from 'axios';
import toast from 'react-hot-toast';

export interface ApiErrorResponse {
  error: string;
  code?: string;
  details?: Record<string, string>;
  request_id?: string;
  debug?: {
    message: string;
    type: string;
  };
}

/**
 * Extracts error message from various error types
 */
export function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    // Check if it's an Axios error
    if ('isAxiosError' in error && error.isAxiosError) {
      const axiosError = error as AxiosError<ApiErrorResponse>;
      
      // Network error (no response)
      if (!axiosError.response) {
        if (axiosError.code === 'ECONNABORTED' || axiosError.message.includes('timeout')) {
          return 'Request timed out. Please check your connection and try again.';
        }
        if (axiosError.message.includes('Network Error') || axiosError.code === 'ERR_NETWORK') {
          return 'Cannot connect to server. Please check your network connection.';
        }
        return 'Network error. Please check your connection and try again.';
      }
      
      // API error response
      const response = axiosError.response;
      const errorData = response.data as ApiErrorResponse;
      
      // Use error message from API if available
      if (errorData?.error) {
        return errorData.error;
      }
      
      // Fallback to status-based messages
      switch (response.status) {
        case 400:
          return 'Invalid request. Please check your input and try again.';
        case 401:
          return 'You are not authorized. Please log in again.';
        case 403:
          return "You don't have permission to perform this action.";
        case 404:
          return 'The requested resource was not found.';
        case 409:
          return 'This action conflicts with existing data. Please refresh and try again.';
        case 422:
          return 'Validation error. Please check your input.';
        case 500:
          return 'An unexpected error occurred. Please try again or contact support.';
        case 503:
          return 'Service temporarily unavailable. Please try again later.';
        default:
          return `Error ${response.status}: ${response.statusText || 'An error occurred'}`;
      }
    }
    
    // Regular Error object
    return error.message;
  }
  
  // String error
  if (typeof error === 'string') {
    return error;
  }
  
  // Unknown error type
  return 'An unexpected error occurred. Please try again.';
}

/**
 * Gets field-level validation errors from API response
 */
export function getValidationErrors(error: unknown): Record<string, string> | null {
  if ('isAxiosError' in error && error instanceof Error) {
    const axiosError = error as AxiosError<ApiErrorResponse>;
    if (axiosError.response?.data?.details) {
      return axiosError.response.data.details;
    }
  }
  return null;
}

/**
 * Gets error code from API response
 */
export function getErrorCode(error: unknown): string | null {
  if ('isAxiosError' in error && error instanceof Error) {
    const axiosError = error as AxiosError<ApiErrorResponse>;
    return axiosError.response?.data?.code || null;
  }
  return null;
}

/**
 * Gets request ID from API response (for support/debugging)
 */
export function getRequestId(error: unknown): string | null {
  if ('isAxiosError' in error && error instanceof Error) {
    const axiosError = error as AxiosError<ApiErrorResponse>;
    return axiosError.response?.data?.request_id || null;
  }
  return null;
}

/**
 * Shows error toast notification
 */
export function showError(error: unknown, customMessage?: string): void {
  const message = customMessage || getErrorMessage(error);
  const requestId = getRequestId(error);
  
  // Include request ID in development
  const devMessage = process.env.NODE_ENV === 'development' && requestId
    ? `${message} (Request ID: ${requestId})`
    : message;
  
  toast.error(devMessage, {
    duration: 5000,
    position: 'top-right',
  });
  
  // Log full error details in development
  if (process.env.NODE_ENV === 'development') {
    console.error('Error details:', error);
  }
}

/**
 * Shows success toast notification
 */
export function showSuccess(message: string): void {
  toast.success(message, {
    duration: 3000,
    position: 'top-right',
  });
}

/**
 * Shows info toast notification
 */
export function showInfo(message: string): void {
  toast(message, {
    duration: 3000,
    position: 'top-right',
    icon: 'ℹ️',
  });
}

/**
 * Handles API errors with appropriate user feedback
 */
export function handleApiError(error: unknown, customMessage?: string): void {
  showError(error, customMessage);
  
  // Log validation errors separately
  const validationErrors = getValidationErrors(error);
  if (validationErrors) {
    console.warn('Validation errors:', validationErrors);
  }
}
