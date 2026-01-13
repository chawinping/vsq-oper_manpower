'use client';

import { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import { authApi, User } from '@/lib/api/auth';

interface UserContextType {
  user: User | null;
  loading: boolean;
  refetch: () => Promise<void>;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

export function UserProvider({ children }: { children: ReactNode }) {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchUser = async () => {
    try {
      const userData = await authApi.getMe();
      setUser(userData);
    } catch (error: any) {
      // Suppress 401 errors on login page - this is expected when not authenticated
      const isLoginPage = typeof window !== 'undefined' && window.location.pathname.includes('/login');
      const isUnauthorized = error.response?.status === 401;
      const isNetworkError = !error.response && (error.code === 'ECONNABORTED' || error.message?.includes('timeout') || error.message?.includes('Network Error'));
      
      // Log network errors and other non-401 errors (unless on login page)
      if (!isLoginPage || (!isUnauthorized && !isNetworkError)) {
        if (isNetworkError) {
          console.warn('Backend connection failed - is the server running?', error.message);
        } else {
          console.error('Failed to fetch user:', error);
        }
      }
      
      setUser(null);
      // Only redirect if not already on login page
      if (typeof window !== 'undefined' && !isLoginPage) {
        router.push('/login');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUser();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Only run once on mount

  return (
    <UserContext.Provider value={{ user, loading, refetch: fetchUser }}>
      {children}
    </UserContext.Provider>
  );
}

export function useUser() {
  const context = useContext(UserContext);
  if (context === undefined) {
    throw new Error('useUser must be used within a UserProvider');
  }
  return context;
}


