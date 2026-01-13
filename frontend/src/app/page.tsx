'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { authApi } from '@/lib/api/auth';

export default function Home() {
  const router = useRouter();

  useEffect(() => {
    const checkAuth = async () => {
      try {
        await authApi.getMe();
        // User is authenticated, redirect to dashboard
        router.replace('/dashboard');
      } catch (error: any) {
        // User is not authenticated or backend is down, redirect to login
        // Network errors and timeouts will also redirect to login
        // Only redirect if not already on login page to prevent loops
        if (window.location.pathname !== '/login') {
          router.replace('/login');
        }
      }
    };

    checkAuth();
  }, [router]);

  // Show loading state while checking authentication
  return (
    <main className="min-h-screen bg-neutral-bg-primary flex items-center justify-center">
      <div className="text-center">
        <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">VSQ Operations Manpower</h1>
        <p className="text-sm text-neutral-text-secondary mb-4">Staff allocation system</p>
        <p className="text-sm text-neutral-text-secondary">Redirecting...</p>
      </div>
    </main>
  );
}

