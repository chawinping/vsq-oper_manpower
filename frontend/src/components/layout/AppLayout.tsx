'use client';

import { useEffect, useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import Link from 'next/link';
import { authApi, User } from '@/lib/api/auth';

interface AppLayoutProps {
  children: React.ReactNode;
}

interface NavItem {
  label: string;
  href: string;
  roles?: string[];
}

const navItems: NavItem[] = [
  { label: 'Dashboard', href: '/dashboard' },
  { label: 'Staff Management', href: '/staff-management', roles: ['admin', 'area_manager', 'district_manager', 'branch_manager'] },
  { label: 'Staff Scheduling', href: '/staff-scheduling', roles: ['admin', 'branch_manager'] },
  { label: 'Rotation Scheduling', href: '/rotation-scheduling', roles: ['admin', 'area_manager', 'district_manager'] },
  { label: 'Branch Management', href: '/branch-management', roles: ['admin', 'area_manager', 'district_manager'] },
  { label: 'Users', href: '/users', roles: ['admin'] },
  { label: 'System Settings', href: '/system-settings', roles: ['admin'] },
];

export default function AppLayout({ children }: AppLayoutProps) {
  const router = useRouter();
  const pathname = usePathname();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [sidebarOpen, setSidebarOpen] = useState(true);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const userData = await authApi.getMe();
        setUser(userData);
      } catch (error) {
        router.push('/login');
      } finally {
        setLoading(false);
      }
    };

    fetchUser();
  }, [router]);

  const handleLogout = async () => {
    try {
      await authApi.logout();
      router.push('/login');
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const filteredNavItems = navItems.filter((item) => {
    if (!item.roles || !user) return true;
    return item.roles.includes(user.role);
  });

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-neutral-text-secondary">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-neutral-bg-primary flex">
      {/* Sidebar */}
      <aside
        className={`bg-neutral-bg-secondary border-r border-neutral-border transition-all duration-300 ${
          sidebarOpen ? 'w-64' : 'w-16'
        }`}
      >
        <div className="h-16 flex items-center justify-between px-4 border-b border-neutral-border">
          {sidebarOpen && (
            <h1 className="text-lg font-semibold text-salesforce-blue">VSQ Manpower</h1>
          )}
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="p-2 hover:bg-neutral-hover rounded transition-colors"
            aria-label="Toggle sidebar"
          >
            <svg
              className="w-5 h-5 text-neutral-text-secondary"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              {sidebarOpen ? (
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              ) : (
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              )}
            </svg>
          </button>
        </div>

        <nav className="py-4">
          {filteredNavItems.map((item) => {
            const isActive = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`flex items-center px-4 py-2 mx-2 rounded transition-colors ${
                  isActive
                    ? 'bg-salesforce-blue-light text-salesforce-blue-dark font-medium'
                    : 'text-neutral-text-primary hover:bg-neutral-hover'
                }`}
                title={sidebarOpen ? undefined : item.label}
              >
                {sidebarOpen ? (
                  <span>{item.label}</span>
                ) : (
                  <span className="text-xs font-medium truncate">{item.label.charAt(0)}</span>
                )}
              </Link>
            );
          })}
        </nav>
      </aside>

      {/* Main Content */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Header */}
        <header className="h-16 bg-neutral-bg-secondary border-b border-neutral-border flex items-center justify-between px-6">
          <div className="flex-1"></div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-neutral-text-secondary">
              {user?.username} <span className="text-neutral-text-primary">({user?.role})</span>
            </span>
            <button
              onClick={handleLogout}
              className="btn-secondary text-sm"
            >
              Logout
            </button>
          </div>
        </header>

        {/* Page Content */}
        <main className="flex-1 overflow-auto">
          {children}
        </main>
      </div>
    </div>
  );
}

