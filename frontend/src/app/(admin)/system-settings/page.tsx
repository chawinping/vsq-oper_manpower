'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/contexts/UserContext';
import { settingsApi, SystemSetting, UpdateSettingRequest } from '@/lib/api/settings';

export default function SystemSettingsPage() {
  const router = useRouter();
  const { user, loading: userLoading } = useUser();
  const [settings, setSettings] = useState<SystemSetting[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [editValue, setEditValue] = useState<string>('');
  const [editDescription, setEditDescription] = useState<string>('');
  const [saving, setSaving] = useState(false);
  const [showAddModal, setShowAddModal] = useState(false);
  const [newKey, setNewKey] = useState('');
  const [newValue, setNewValue] = useState('');
  const [newDescription, setNewDescription] = useState('');

  useEffect(() => {
    // Check if user has permission
    if (!userLoading && user && user.role !== 'admin') {
      router.push('/dashboard');
      return;
    }
  }, [user, userLoading, router]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        if (user?.role === 'admin') {
          await loadSettings();
        }
      } catch (error: any) {
        console.error('Failed to fetch data:', error);
      } finally {
        setLoading(false);
      }
    };

    if (user && user.role === 'admin') {
      fetchData();
    }
  }, [user]);

  const loadSettings = async () => {
    try {
      const settingsData = await settingsApi.getAll();
      setSettings(settingsData || []);
    } catch (error) {
      console.error('Failed to load settings:', error);
      setSettings([]);
    }
  };


  const handleEdit = (setting: SystemSetting) => {
    setEditingKey(setting.key);
    setEditValue(setting.value);
    setEditDescription(setting.description || '');
  };

  const handleSave = async () => {
    if (!editingKey) return;

    setSaving(true);
    try {
      await settingsApi.update(editingKey, {
        value: editValue,
        description: editDescription,
      });
      setEditingKey(null);
      await loadSettings();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to update setting');
    } finally {
      setSaving(false);
    }
  };

  const handleAdd = async () => {
    if (!newKey || !newValue) {
      alert('Key and value are required');
      return;
    }

    setSaving(true);
    try {
      await settingsApi.update(newKey, {
        value: newValue,
        description: newDescription,
      });
      setShowAddModal(false);
      setNewKey('');
      setNewValue('');
      setNewDescription('');
      await loadSettings();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to create setting');
    } finally {
      setSaving(false);
    }
  };

  if (userLoading || loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-neutral-text-secondary">Loading...</div>
      </div>
    );
  }

  if (user?.role !== 'admin') {
    return null;
  }

  return (
    <>
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">System Settings</h1>
          <p className="text-sm text-neutral-text-secondary">Manage system configuration settings</p>
        </div>

        <div className="card">
          <div className="p-4 border-b border-neutral-border">
            <button
              onClick={() => setShowAddModal(true)}
              className="btn-primary"
            >
              Add Setting
            </button>
          </div>

          <div className="overflow-x-auto">
            <table className="table-salesforce">
              <thead>
                <tr>
                  <th>Key</th>
                  <th>Value</th>
                  <th>Description</th>
                  <th>Updated At</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {(settings || []).map((setting) => (
                  <tr key={setting.id}>
                    <td className="font-medium">{setting.key}</td>
                    <td>
                      {editingKey === setting.key ? (
                        <input
                          type="text"
                          value={editValue}
                          onChange={(e) => setEditValue(e.target.value)}
                          className="input-field"
                        />
                      ) : (
                        <span className="break-all">{setting.value}</span>
                      )}
                    </td>
                    <td>
                      {editingKey === setting.key ? (
                        <input
                          type="text"
                          value={editDescription}
                          onChange={(e) => setEditDescription(e.target.value)}
                          className="input-field"
                        />
                      ) : (
                        setting.description || '-'
                      )}
                    </td>
                    <td>{new Date(setting.updated_at).toLocaleString()}</td>
                    <td>
                      {editingKey === setting.key ? (
                        <div className="flex gap-3">
                          <button
                            onClick={handleSave}
                            disabled={saving}
                            className="text-green-600 hover:text-green-700 text-sm disabled:opacity-50"
                          >
                            Save
                          </button>
                          <button
                            onClick={() => {
                              setEditingKey(null);
                              setEditValue('');
                              setEditDescription('');
                            }}
                            className="text-neutral-text-secondary hover:text-neutral-text-primary text-sm"
                          >
                            Cancel
                          </button>
                        </div>
                      ) : (
                        <button
                          onClick={() => handleEdit(setting)}
                          className="text-salesforce-blue hover:text-salesforce-blue-hover text-sm"
                        >
                          Edit
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {settings.length === 0 && (
              <div className="text-center py-12 text-neutral-text-secondary">
                No settings found
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Add Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="card max-w-md w-full">
            <div className="p-6">
              <h2 className="text-xl font-semibold text-neutral-text-primary mb-6">Add System Setting</h2>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                    Key *
                  </label>
                  <input
                    type="text"
                    required
                    value={newKey}
                    onChange={(e) => setNewKey(e.target.value)}
                    className="input-field"
                    placeholder="e.g., max_staff_per_branch"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                    Value *
                  </label>
                  <input
                    type="text"
                    required
                    value={newValue}
                    onChange={(e) => setNewValue(e.target.value)}
                    className="input-field"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                    Description
                  </label>
                  <textarea
                    value={newDescription}
                    onChange={(e) => setNewDescription(e.target.value)}
                    rows={3}
                    className="input-field"
                  />
                </div>
              </div>
              <div className="mt-6 flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => {
                    setShowAddModal(false);
                    setNewKey('');
                    setNewValue('');
                    setNewDescription('');
                  }}
                  className="btn-secondary"
                >
                  Cancel
                </button>
                <button
                  onClick={handleAdd}
                  disabled={saving}
                  className="btn-primary disabled:opacity-50"
                >
                  {saving ? 'Saving...' : 'Create'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

