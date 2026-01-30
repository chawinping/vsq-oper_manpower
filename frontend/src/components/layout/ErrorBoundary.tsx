'use client';

import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  // Normalize error to Error object - React can throw anything
  // This function MUST always return an Error object, never null/undefined
  private static normalizeError(error: unknown): Error {
    try {
      if (error instanceof Error) {
        return error;
      }
      if (error === null || error === undefined) {
        return new Error('An unknown error occurred');
      }
      if (typeof error === 'string') {
        return new Error(error || 'An unknown error occurred');
      }
      if (typeof error === 'number') {
        return new Error(`Error code: ${error}`);
      }
      if (typeof error === 'boolean') {
        return new Error(`Error: ${error}`);
      }
      // For objects, try to stringify safely
      try {
        const errorString = String(error);
        return new Error(errorString || 'An unknown error occurred');
      } catch {
        return new Error('An unknown error occurred');
      }
    } catch (normalizationError) {
      // If normalization itself fails, return a safe error
      // This should never happen, but we need to be defensive
      return new Error('An error occurred while processing another error');
    }
  }

  static getDerivedStateFromError(error: unknown): State {
    return {
      hasError: true,
      error: ErrorBoundary.normalizeError(error),
      errorInfo: null,
    };
  }

  componentDidCatch(error: unknown, errorInfo: ErrorInfo) {
    const normalizedError = ErrorBoundary.normalizeError(error);
    
    // Log error to console
    console.error('ErrorBoundary caught an error:', normalizedError, errorInfo);
    
    // Update state with error info
    this.setState({
      error: normalizedError,
      errorInfo,
    });

    // TODO: Send error to error tracking service (e.g., Sentry)
    // Example:
    // if (process.env.NODE_ENV === 'production') {
    //   Sentry.captureException(normalizedError, { contexts: { react: { componentStack: errorInfo.componentStack } } });
    // }
  }

  handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  render() {
    if (this.state.hasError) {
      // Ensure error is always defined and safe to display
      const error = this.state.error || new Error('An unknown error occurred');
      const errorMessage = error?.message || (error instanceof Error ? error.toString() : String(error)) || 'Unknown error';
      
      return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
          <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6">
            <div className="flex items-center justify-center w-12 h-12 mx-auto bg-red-100 rounded-full mb-4">
              <svg
                className="w-6 h-6 text-red-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                />
              </svg>
            </div>
            
            <h2 className="text-xl font-semibold text-gray-900 text-center mb-2">
              Something went wrong
            </h2>
            
            <p className="text-gray-600 text-center mb-6">
              We're sorry, but something unexpected happened. Please try refreshing the page.
            </p>

            {process.env.NODE_ENV === 'development' && error && (
              <div className="mb-6 p-4 bg-gray-100 rounded-md overflow-auto max-h-48">
                <p className="text-sm font-mono text-red-600 mb-2">
                  {errorMessage}
                </p>
                {this.state.errorInfo && this.state.errorInfo.componentStack && (
                  <pre className="text-xs text-gray-600 overflow-auto">
                    {this.state.errorInfo.componentStack}
                  </pre>
                )}
              </div>
            )}

            <div className="flex gap-3">
              <button
                onClick={this.handleReset}
                className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
              >
                Try Again
              </button>
              <button
                onClick={() => window.location.reload()}
                className="flex-1 px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 transition-colors"
              >
                Refresh Page
              </button>
            </div>

            {process.env.NODE_ENV === 'production' && (
              <p className="mt-4 text-xs text-gray-500 text-center">
                If this problem persists, please contact support.
              </p>
            )}
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
