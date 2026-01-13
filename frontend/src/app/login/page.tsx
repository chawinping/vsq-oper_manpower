'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { authApi } from '@/lib/api/auth';
import { versionApi, VersionInfo } from '@/lib/api/version';

export default function LoginPage() {
  useEffect(() => {
    // Global error handler
    const errorHandler = (e: ErrorEvent) => {
      console.error('Global error:', e.error, e.message, e.filename, e.lineno);
    };
    
    const rejectionHandler = (e: PromiseRejectionEvent) => {
      console.error('Unhandled promise rejection:', e.reason);
    };
    
    window.addEventListener('error', errorHandler);
    window.addEventListener('unhandledrejection', rejectionHandler);
    
    return () => {
      window.removeEventListener('error', errorHandler);
      window.removeEventListener('unhandledrejection', rejectionHandler);
    };
  }, []);

  // Load versions
  const [frontendVersion, setFrontendVersion] = useState<VersionInfo | null>(null);
  const [backendVersion, setBackendVersion] = useState<VersionInfo | null>(null);
  const [databaseVersion, setDatabaseVersion] = useState<VersionInfo | null>(null);
  const [versionsLoading, setVersionsLoading] = useState(true);
  const [versionsError, setVersionsError] = useState<string | null>(null);

  useEffect(() => {
    const loadVersions = async () => {
      setVersionsLoading(true);
      setVersionsError(null);
      
      try {
        // Load frontend version
        try {
          const frontend = await versionApi.getFrontendVersion();
          console.log('Frontend version loaded:', frontend);
          setFrontendVersion(frontend);
        } catch (err: any) {
          console.error('Failed to load frontend version:', err);
          setFrontendVersion({ version: 'error', buildDate: 'N/A', buildTime: 'N/A' });
        }

        // Load backend and database versions
        try {
          const versions = await versionApi.getVersions();
          console.log('Backend/Database versions loaded:', versions);
          setBackendVersion(versions.backend);
          setDatabaseVersion(versions.database);
        } catch (err: any) {
          console.error('Failed to load backend/database versions:', err);
          setBackendVersion({ version: 'error', buildDate: 'N/A', buildTime: 'N/A' });
          setDatabaseVersion({ version: 'error', buildDate: 'N/A', buildTime: 'N/A' });
          setVersionsError('Unable to load server versions');
        }
      } catch (err: any) {
        console.error('Failed to load versions:', err);
        setVersionsError('Failed to load version information');
      } finally {
        setVersionsLoading(false);
      }
    };

    loadVersions();
  }, []);
  
  const router = useRouter();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e?: React.FormEvent) => {
    if (e) {
      e.preventDefault();
      e.stopPropagation();
    }
    
    setError('');
    setLoading(true);

    try {
      await authApi.login({ username, password });
      
      // Use window.location for a full page reload to ensure session is recognized
      window.location.href = '/dashboard';
    } catch (err: any) {
      console.error('Login error:', err);
      
      // Handle different error structures
      let errorMessage = 'Login failed';
      if (err.response?.data?.error) {
        errorMessage = err.response.data.error;
      } else if (err.response?.data?.message) {
        errorMessage = err.response.data.message;
      } else if (err.message) {
        errorMessage = err.message;
      } else if (err.response?.status === 401) {
        errorMessage = 'Invalid username or password';
      } else if (err.response?.status === 500) {
        errorMessage = 'Server error. Please try again later.';
      } else if (!err.response) {
        errorMessage = 'Network error. Please check your connection.';
      }
      
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-neutral-bg-primary flex items-center justify-center px-4 py-8">
      <div className="w-full max-w-md space-y-4">
        <div className="card p-8">
          <div className="text-center mb-8">
            <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">VSQ Operations Manpower</h1>
            <p className="text-sm text-neutral-text-secondary">Sign in to your account</p>
          </div>

          <form 
            onSubmit={(e) => {
              e.preventDefault();
              e.stopPropagation();
              handleSubmit(e);
              return false;
            }} 
            className="space-y-5"
            noValidate
          >
            {error && (
              <div className="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
                {error}
              </div>
            )}

            <div>
              <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                Username
              </label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                className="input-field"
                placeholder="Enter your username"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                Password
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                className="input-field"
                placeholder="Enter your password"
              />
            </div>

            <button
              type="button"
              disabled={loading}
              className="btn-primary w-full"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                handleSubmit();
              }}
            >
              {loading ? 'Signing in...' : 'Sign In'}
            </button>
          </form>
        </div>

        {/* Version Information - Always visible */}
        <div 
          className="card p-4" 
          style={{ 
            backgroundColor: '#f9fafb',
            border: '1px solid #d1d5db'
          }}
        >
          <div className="text-xs space-y-1.5" style={{ fontSize: '11px' }}>
            <div 
              className="font-semibold mb-3 text-center" 
              style={{ 
                fontSize: '11px', 
                fontWeight: 600,
                color: '#374151'
              }}
            >
              Version Information
            </div>
            {versionsError && (
              <div 
                className="mb-2 text-center p-2 bg-red-50 rounded" 
                style={{ 
                  fontSize: '10px', 
                  color: '#ef4444',
                  backgroundColor: '#fef2f2'
                }}
              >
                ⚠️ {versionsError}
              </div>
            )}
            {!versionsLoading && !frontendVersion && !backendVersion && !databaseVersion && !versionsError && (
              <div 
                className="mb-2 text-center p-2 bg-yellow-50 rounded" 
                style={{ 
                  fontSize: '10px', 
                  color: '#92400e',
                  backgroundColor: '#fef3c7'
                }}
              >
                ⚠️ Version information not available
              </div>
            )}
            {/* Frontend Version */}
            <div className="space-y-0.5 mb-2">
              <div 
                className="flex justify-between items-center py-1" 
                style={{ fontSize: '11px', paddingTop: '2px', paddingBottom: '2px' }}
              >
                <span className="font-medium" style={{ fontWeight: 500, color: '#6b7280' }}>Frontend:</span>
                <span 
                  className="font-mono px-2 py-0.5 bg-gray-100 rounded" 
                  style={{ 
                    fontFamily: 'monospace', 
                    fontWeight: 600,
                    color: '#111827',
                    backgroundColor: '#f3f4f6'
                  }}
                >
                  {versionsLoading ? 'Loading...' : (frontendVersion?.version || 'N/A')}
                </span>
              </div>
              {!versionsLoading && frontendVersion && frontendVersion.buildDate && frontendVersion.buildDate !== 'N/A' && frontendVersion.buildDate !== 'unknown' && (
                <div 
                  className="flex justify-between items-center py-0.5 pl-1" 
                  style={{ fontSize: '10px', paddingTop: '1px', paddingBottom: '1px', color: '#9ca3af' }}
                >
                  <span>Build Date:</span>
                  <span className="font-mono">{frontendVersion.buildDate}</span>
                </div>
              )}
              {!versionsLoading && frontendVersion && frontendVersion.buildTime && frontendVersion.buildTime !== 'N/A' && frontendVersion.buildTime !== 'unknown' && (
                <div 
                  className="flex justify-between items-center py-0.5 pl-1" 
                  style={{ fontSize: '10px', paddingTop: '1px', paddingBottom: '1px', color: '#9ca3af' }}
                >
                  <span>Build Time:</span>
                  <span className="font-mono">{frontendVersion.buildTime}</span>
                </div>
              )}
            </div>

            {/* Backend Version */}
            <div className="space-y-0.5 mb-2">
              <div 
                className="flex justify-between items-center py-1" 
                style={{ fontSize: '11px', paddingTop: '2px', paddingBottom: '2px' }}
              >
                <span className="font-medium" style={{ fontWeight: 500, color: '#6b7280' }}>Backend:</span>
                <span 
                  className="font-mono px-2 py-0.5 bg-gray-100 rounded" 
                  style={{ 
                    fontFamily: 'monospace', 
                    fontWeight: 600,
                    color: '#111827',
                    backgroundColor: '#f3f4f6'
                  }}
                >
                  {versionsLoading ? 'Loading...' : (backendVersion?.version || 'N/A')}
                </span>
              </div>
              {!versionsLoading && backendVersion && backendVersion.buildDate && backendVersion.buildDate !== 'N/A' && backendVersion.buildDate !== 'unknown' && (
                <div 
                  className="flex justify-between items-center py-0.5 pl-1" 
                  style={{ fontSize: '10px', paddingTop: '1px', paddingBottom: '1px', color: '#9ca3af' }}
                >
                  <span>Build Date:</span>
                  <span className="font-mono">{backendVersion.buildDate}</span>
                </div>
              )}
              {!versionsLoading && backendVersion && backendVersion.buildTime && backendVersion.buildTime !== 'N/A' && backendVersion.buildTime !== 'unknown' && (
                <div 
                  className="flex justify-between items-center py-0.5 pl-1" 
                  style={{ fontSize: '10px', paddingTop: '1px', paddingBottom: '1px', color: '#9ca3af' }}
                >
                  <span>Build Time:</span>
                  <span className="font-mono">{backendVersion.buildTime}</span>
                </div>
              )}
            </div>

            {/* Database Version */}
            <div className="space-y-0.5">
              <div 
                className="flex justify-between items-center py-1" 
                style={{ fontSize: '11px', paddingTop: '2px', paddingBottom: '2px' }}
              >
                <span className="font-medium" style={{ fontWeight: 500, color: '#6b7280' }}>Database:</span>
                <span 
                  className="font-mono px-2 py-0.5 bg-gray-100 rounded" 
                  style={{ 
                    fontFamily: 'monospace', 
                    fontWeight: 600,
                    color: '#111827',
                    backgroundColor: '#f3f4f6'
                  }}
                >
                  {versionsLoading ? 'Loading...' : (databaseVersion?.version || 'N/A')}
                </span>
              </div>
              {!versionsLoading && databaseVersion && databaseVersion.buildDate && databaseVersion.buildDate !== 'N/A' && databaseVersion.buildDate !== 'unknown' && (
                <div 
                  className="flex justify-between items-center py-0.5 pl-1" 
                  style={{ fontSize: '10px', paddingTop: '1px', paddingBottom: '1px', color: '#9ca3af' }}
                >
                  <span>Build Date:</span>
                  <span className="font-mono">{databaseVersion.buildDate}</span>
                </div>
              )}
              {!versionsLoading && databaseVersion && databaseVersion.buildTime && databaseVersion.buildTime !== 'N/A' && databaseVersion.buildTime !== 'unknown' && (
                <div 
                  className="flex justify-between items-center py-0.5 pl-1" 
                  style={{ fontSize: '10px', paddingTop: '1px', paddingBottom: '1px', color: '#9ca3af' }}
                >
                  <span>Build Time:</span>
                  <span className="font-mono">{databaseVersion.buildTime}</span>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

