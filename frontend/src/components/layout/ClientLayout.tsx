'use client';

import { Toaster } from 'react-hot-toast';
import { UserProvider } from '@/contexts/UserContext';
import AppLayout from './AppLayout';
import ErrorBoundary from './ErrorBoundary';

export default function ClientLayout({ children }: { children: React.ReactNode }) {
  return (
    <ErrorBoundary>
      <UserProvider>
        <AppLayout>{children}</AppLayout>
        <Toaster
          position="top-right"
          toastOptions={{
            duration: 4000,
            style: {
              background: '#363636',
              color: '#fff',
            },
            success: {
              duration: 3000,
              iconTheme: {
                primary: '#10b981',
                secondary: '#fff',
              },
            },
            error: {
              duration: 5000,
              iconTheme: {
                primary: '#ef4444',
                secondary: '#fff',
              },
            },
          }}
        />
      </UserProvider>
    </ErrorBoundary>
  );
}
