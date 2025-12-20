'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { authApi, User as AuthUser } from '@/lib/api/auth';
import { userApi, User, CreateUserRequest } from '@/lib/api/user';
import { roleApi, Role } from '@/lib/api/role';
import AppLayout from '@/components/layout/AppLayout';

export default function UsersManagementPage() {
  const router = useRouter();
  const [user, setUser] = useState<AuthUser | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);

  const [formData, setFormData] = useState<CreateUserRequest>({
    username: '',
    email: '',
    password: '',
    role_id: '',
  });

  useEffect(() => {
    const fetchData = async () => {
      try {
        const userData = await authApi.getMe();
        if (!userData) {
          throw new Error('User data not available');
        }
        setUser(userData);
        
        if (!userData.role || userData.role !== 'admin') {
          router.push('/dashboard');
          return;
        }

        await loadUsers();
        const rolesData = await roleApi.list();
        setRoles(rolesData || []);
      } catch (error: any) {
        console.error('Failed to fetch data:', error);
        if (typeof window !== 'undefined' && !window.location.pathname.includes('/login')) {
          router.push('/login');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [router]);

  const loadUsers = async () => {
    try {
      const usersData = await userApi.list();
      setUsers(usersData || []);
    } catch (error) {
      console.error('Failed to load users:', error);
      setUsers([]);
    }
  };


  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingUser) {
        const updateData: any = {
          username: formData.username,
          email: formData.email,
          role_id: formData.role_id,
        };
        if (formData.password) {
          updateData.password = formData.password;
        }
        await userApi.update(editingUser.id, updateData);
      } else {
        await userApi.create(formData);
      }

      setShowModal(false);
      setEditingUser(null);
      setFormData({
        username: '',
        email: '',
        password: '',
        role_id: '',
      });
      await loadUsers();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save user');
    }
  };

  const handleEdit = (userToEdit: User) => {
    setEditingUser(userToEdit);
    setFormData({
      username: userToEdit.username,
      email: userToEdit.email,
      password: '', // Don't pre-fill password
      role_id: userToEdit.role_id,
    });
    setShowModal(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this user?')) {
      return;
    }

    try {
      await userApi.delete(id);
      await loadUsers();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete user');
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-neutral-text-secondary">Loading...</div>
      </div>
    );
  }

  return (
    <AppLayout>
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">User Management</h1>
          <p className="text-sm text-neutral-text-secondary">Manage system users and their roles</p>
        </div>

        <div className="card">
          <div className="p-4 border-b border-neutral-border">
            <button
              onClick={() => {
                setEditingUser(null);
                setFormData({
                  username: '',
                  email: '',
                  password: '',
                  role_id: '',
                });
                setShowModal(true);
              }}
              className="btn-primary"
            >
              Add User
            </button>
          </div>

          <div className="overflow-x-auto">
            <table className="table-salesforce">
              <thead>
                <tr>
                  <th>Username</th>
                  <th>Email</th>
                  <th>Role</th>
                  <th>Created At</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {(users || []).map((userItem) => (
                  <tr key={userItem.id}>
                    <td className="font-medium">{userItem.username}</td>
                    <td>{userItem.email}</td>
                    <td>
                      <span className="badge badge-primary">
                        {userItem.role_name}
                      </span>
                    </td>
                    <td>{new Date(userItem.created_at).toLocaleDateString()}</td>
                    <td>
                      <div className="flex gap-3">
                        <button
                          onClick={() => handleEdit(userItem)}
                          className="text-salesforce-blue hover:text-salesforce-blue-hover text-sm"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => handleDelete(userItem.id)}
                          className="text-red-600 hover:text-red-700 text-sm"
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {users.length === 0 && (
              <div className="text-center py-12 text-neutral-text-secondary">
                No users found
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Add/Edit Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="card max-w-md w-full">
            <div className="p-6">
              <h2 className="text-xl font-semibold text-neutral-text-primary mb-6">
                {editingUser ? 'Edit User' : 'Add User'}
              </h2>
              <form onSubmit={handleSubmit}>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Username *
                    </label>
                    <input
                      type="text"
                      required
                      value={formData.username}
                      onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                      className="input-field"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Email *
                    </label>
                    <input
                      type="email"
                      required
                      value={formData.email}
                      onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                      className="input-field"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Password {editingUser ? '(leave blank to keep current)' : '*'}
                    </label>
                    <input
                      type="password"
                      required={!editingUser}
                      value={formData.password}
                      onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                      className="input-field"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Role *
                    </label>
                    <select
                      required
                      value={formData.role_id}
                      onChange={(e) => setFormData({ ...formData, role_id: e.target.value })}
                      className="input-field"
                    >
                      <option value="">Select Role</option>
                      {(roles || []).map((role) => (
                        <option key={role.id} value={role.id}>
                          {role.name}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>
                <div className="mt-6 flex justify-end gap-2">
                  <button
                    type="button"
                    onClick={() => {
                      setShowModal(false);
                      setEditingUser(null);
                    }}
                    className="btn-secondary"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="btn-primary"
                  >
                    {editingUser ? 'Update' : 'Create'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}
    </AppLayout>
  );
}

