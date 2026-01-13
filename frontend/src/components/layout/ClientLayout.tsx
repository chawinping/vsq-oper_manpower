'use client';

import { UserProvider } from '@/contexts/UserContext';
import AppLayout from './AppLayout';

export default function ClientLayout({ children }: { children: React.ReactNode }) {
  return (
    <UserProvider>
      <AppLayout>{children}</AppLayout>
    </UserProvider>
  );
}
