'use client';

import { useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import Link from 'next/link';
import { authApi } from '@/lib/api/auth';
import { useUser } from '@/contexts/UserContext';

interface AppLayoutProps {
  children: React.ReactNode;
}

interface NavItem {
  label: string;
  href: string;
  roles?: string[];
}

interface NavCategory {
  label: string;
  items: NavItem[];
  roles?: string[];
}

const navCategories: NavCategory[] = [
  {
    label: 'Personnel',
    roles: ['admin', 'area_manager', 'district_manager', 'branch_manager'],
    items: [
      { label: 'Branch Staff Profile', href: '/staff-management', roles: ['admin', 'area_manager', 'district_manager', 'branch_manager'] },
      { label: 'Rotation Staff Profile', href: '/rotation-staff-profile', roles: ['admin', 'area_manager', 'district_manager'] },
      { label: 'Doctor Profile', href: '/doctor-management', roles: ['admin', 'area_manager', 'district_manager', 'branch_manager'] },
      { label: 'Branch Management', href: '/branch-management', roles: ['admin', 'area_manager', 'district_manager'] },
      { label: 'Positions', href: '/positions', roles: ['admin'] },
    ],
  },
  {
    label: 'Scheduling',
    roles: ['admin', 'area_manager', 'district_manager', 'branch_manager'],
    items: [
      { label: 'Staff Scheduling', href: '/staff-scheduling', roles: ['admin', 'branch_manager'] },
      { label: 'Doctor Scheduling', href: '/doctor-schedule', roles: ['admin', 'branch_manager'] },
      { label: 'Rotation Staff Scheduling', href: '/rotation-scheduling', roles: ['admin', 'area_manager', 'district_manager'] },
    ],
  },
  {
    label: 'Allocation Logic',
    roles: ['admin'],
    items: [
      { label: 'Allocation Criteria', href: '/allocation-criteria', roles: ['admin'] },
      { label: 'Revenue Level Tiers', href: '/revenue-level-tiers', roles: ['admin'] },
      { label: 'Staff Requirement Scenarios', href: '/staff-requirement-scenarios', roles: ['admin'] },
    ],
  },
  {
    label: 'System and Administration',
    roles: ['admin'],
    items: [
      { label: 'Users', href: '/users', roles: ['admin'] },
      { label: 'Zone Configuration', href: '/zone-configuration', roles: ['admin'] },
      { label: 'System Settings', href: '/system-settings', roles: ['admin'] },
    ],
  },
];

export default function AppLayout({ children }: AppLayoutProps) {
  const router = useRouter();
  const pathname = usePathname();
  const { user, loading } = useUser();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [expandedCategories, setExpandedCategories] = useState<Set<string>>(new Set(['Personnel', 'Scheduling', 'Allocation Logic', 'System and Administration']));

  // Skip layout for login page
  if (pathname === '/login') {
    return <>{children}</>;
  }

  // Skip loading check for root page - it handles its own redirect
  if (pathname === '/') {
    return <>{children}</>;
  }

  const handleLogout = async () => {
    try {
      await authApi.logout();
      router.push('/login');
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const toggleCategory = (categoryLabel: string) => {
    setExpandedCategories((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(categoryLabel)) {
        newSet.delete(categoryLabel);
      } else {
        newSet.add(categoryLabel);
      }
      return newSet;
    });
  };

  const isCategoryVisible = (category: NavCategory) => {
    if (!category.roles || !user || !user.role) return true;
    return category.roles.includes(user.role);
  };

  const isItemVisible = (item: NavItem) => {
    if (!item.roles || !user || !user.role) return true;
    return item.roles.includes(user.role);
  };

  const filteredCategories = navCategories.filter(isCategoryVisible).map((category) => ({
    ...category,
    items: category.items.filter(isItemVisible),
  })).filter((category) => category.items.length > 0);

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
        className={`bg-neutral-bg-secondary border-r border-neutral-border transition-all duration-300 flex flex-col ${
          sidebarOpen ? 'w-64' : 'w-16'
        }`}
      >
        <div className="h-16 flex items-center justify-between px-4 border-b border-neutral-border flex-shrink-0">
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

        <nav className="flex-1 py-4 overflow-y-auto">
          {/* Dashboard - Standalone */}
          <Link
            href="/dashboard"
            className={`flex items-center px-4 py-2 mx-2 mb-2 rounded transition-colors ${
              pathname === '/dashboard'
                ? 'bg-salesforce-blue-light text-salesforce-blue-dark font-medium'
                : 'text-neutral-text-primary hover:bg-neutral-hover'
            }`}
            title={sidebarOpen ? undefined : 'Dashboard'}
          >
            {sidebarOpen ? (
              <span>Dashboard</span>
            ) : (
              <span className="text-xs font-medium truncate">D</span>
            )}
          </Link>

          {/* Categories */}
          {filteredCategories.map((category) => {
            const isExpanded = expandedCategories.has(category.label);
            const hasActiveItem = category.items.some((item) => pathname === item.href);

            return (
              <div key={category.label} className="mb-1">
                {/* Category Header */}
                {sidebarOpen ? (
                  <button
                    onClick={() => toggleCategory(category.label)}
                    className={`w-full flex items-center justify-between px-4 py-2 mx-2 rounded transition-colors text-sm font-semibold ${
                      hasActiveItem
                        ? 'bg-salesforce-blue-light text-salesforce-blue-dark'
                        : 'text-neutral-text-secondary hover:bg-neutral-hover'
                    }`}
                  >
                    <span>{category.label}</span>
                    <svg
                      className={`w-4 h-4 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                    </svg>
                  </button>
                ) : (
                  <div className="px-4 py-2 mx-2 text-xs font-semibold text-neutral-text-secondary text-center">
                    {category.label.charAt(0)}
                  </div>
                )}

                {/* Category Items */}
                {isExpanded && (
                  <div className="ml-2 mt-1">
                    {category.items.map((item) => {
                      const isActive = pathname === item.href;
                      return (
                        <Link
                          key={item.href}
                          href={item.href}
                          className={`flex items-center px-4 py-2 mx-2 rounded transition-colors text-sm ${
                            isActive
                              ? 'bg-salesforce-blue-light text-salesforce-blue-dark font-medium'
                              : 'text-neutral-text-primary hover:bg-neutral-hover'
                          }`}
                          title={sidebarOpen ? undefined : item.label}
                        >
                          {sidebarOpen ? (
                            <span className="pl-4">{item.label}</span>
                          ) : (
                            <span className="text-xs font-medium truncate">{item.label.charAt(0)}</span>
                          )}
                        </Link>
                      );
                    })}
                  </div>
                )}
              </div>
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
            {user && (
              <span className="text-sm text-neutral-text-secondary">
                {user.role === 'branch_manager' && user.branch_code && user.branch_name ? (
                  <>
                    <span className="text-neutral-text-primary font-medium">
                      {user.branch_code} - {user.branch_name}
                    </span>
                    <span className="mx-2">|</span>
                    <span>{user.username}</span>
                    <span className="text-neutral-text-primary"> ({user.role})</span>
                  </>
                ) : (
                  <>
                    {user.username} <span className="text-neutral-text-primary">({user.role})</span>
                  </>
                )}
              </span>
            )}
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

