'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/contexts/UserContext';
import { doctorApi, Doctor, CreateDoctorRequest } from '@/lib/api/doctor';
import Link from 'next/link';

export default function DoctorManagementPage() {
  const router = useRouter();
  const { user, loading: userLoading } = useUser();
  const [doctors, setDoctors] = useState<Doctor[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingDoctor, setEditingDoctor] = useState<Doctor | null>(null);
  const [showImportModal, setShowImportModal] = useState(false);
  const [importFile, setImportFile] = useState<File | null>(null);
  const [importing, setImporting] = useState(false);

  const [formData, setFormData] = useState<CreateDoctorRequest>({
    name: '',
    code: '',
    preferences: '',
  });

  useEffect(() => {
    // Check if user has permission
    if (!userLoading && user && !['admin', 'area_manager'].includes(user.role || '')) {
      router.push('/dashboard');
      return;
    }
  }, [user, userLoading, router]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        if (user && ['admin', 'area_manager'].includes(user.role || '')) {
          await loadDoctors();
        }
      } catch (error: any) {
        console.error('Failed to fetch data:', error);
      } finally {
        setLoading(false);
      }
    };

    if (user && ['admin', 'area_manager'].includes(user.role || '')) {
      fetchData();
    }
  }, [user]);

  const loadDoctors = async () => {
    try {
      const doctorsData = await doctorApi.list();
      setDoctors(doctorsData || []);
    } catch (error) {
      console.error('Failed to load doctors:', error);
      setDoctors([]);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingDoctor) {
        await doctorApi.update(editingDoctor.id, formData);
      } else {
        await doctorApi.create(formData);
      }

      setShowModal(false);
      setEditingDoctor(null);
      setFormData({
        name: '',
        code: '',
        preferences: '',
      });
      await loadDoctors();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save doctor');
    }
  };

  const handleEdit = (doctorToEdit: Doctor) => {
    setEditingDoctor(doctorToEdit);
    setFormData({
      name: doctorToEdit.name,
      code: doctorToEdit.code || '',
      preferences: doctorToEdit.preferences || '',
    });
    setShowModal(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this doctor? This will also delete all associated assignments and preferences.')) {
      return;
    }

    try {
      await doctorApi.delete(id);
      await loadDoctors();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete doctor');
    }
  };

  const handleImport = async () => {
    if (!importFile) {
      alert('Please select a file');
      return;
    }

    setImporting(true);
    try {
      const result = await doctorApi.import(importFile);
      let message = `Import completed! ${result.imported || 0} doctor(s) imported.`;
      
      if (result.parse_warnings) {
        message += `\n\nParse warnings: ${result.parse_warnings}`;
      }
      if (result.save_warnings && result.save_warnings.length > 0) {
        message += `\n\nSave warnings:\n${result.save_warnings.slice(0, 5).join('\n')}`;
        if (result.save_warnings.length > 5) {
          message += `\n... and ${result.save_warnings.length - 5} more`;
        }
      }
      
      alert(message);
      setShowImportModal(false);
      setImportFile(null);
      await loadDoctors();
    } catch (error: any) {
      const errorMsg = error.response?.data?.error || 'Failed to import doctors';
      if (error.response?.status === 207) {
        // Partial success
        const imported = error.response.data.imported || 0;
        let message = `${errorMsg}\nImported: ${imported}`;
        if (error.response.data.parse_warnings) {
          message += `\n\nParse warnings: ${error.response.data.parse_warnings}`;
        }
        if (error.response.data.save_warnings) {
          message += `\n\nSave warnings:\n${error.response.data.save_warnings.slice(0, 5).join('\n')}`;
        }
        alert(message);
        await loadDoctors();
      } else {
        alert(errorMsg);
      }
    } finally {
      setImporting(false);
    }
  };

  if (loading || userLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-neutral-text-secondary">Loading...</div>
      </div>
    );
  }

  if (!user || !['admin', 'area_manager'].includes(user.role || '')) {
    return null;
  }

  return (
    <>
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">Doctor Profile Management</h1>
          <p className="text-sm text-neutral-text-secondary">Manage doctor profiles, schedules, and preferences</p>
        </div>

        <div className="card">
          <div className="p-4 border-b border-neutral-border flex items-center justify-between">
            <button
              onClick={() => {
                setEditingDoctor(null);
                setFormData({
                  name: '',
                  code: '',
                  preferences: '',
                });
                setShowModal(true);
              }}
              className="btn-primary"
            >
              + Add Doctor
            </button>
            <button
              onClick={() => setShowImportModal(true)}
              className="btn-primary bg-green-600 hover:bg-green-700"
            >
              Import from Excel
            </button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-neutral-border">
                  <th className="text-left p-4 text-sm font-semibold text-neutral-text-primary">Name</th>
                  <th className="text-left p-4 text-sm font-semibold text-neutral-text-primary">Code</th>
                  <th className="text-left p-4 text-sm font-semibold text-neutral-text-primary">Preferences</th>
                  <th className="text-left p-4 text-sm font-semibold text-neutral-text-primary">Actions</th>
                </tr>
              </thead>
              <tbody>
                {doctors.length === 0 ? (
                  <tr>
                    <td colSpan={4} className="p-8 text-center text-neutral-text-secondary">
                      No doctors found. Click "Add Doctor" to create one.
                    </td>
                  </tr>
                ) : (
                  doctors.map((doctor) => (
                    <tr key={doctor.id} className="border-b border-neutral-border hover:bg-neutral-hover">
                      <td className="p-4 text-sm text-neutral-text-primary">{doctor.name}</td>
                      <td className="p-4 text-sm text-neutral-text-secondary">{doctor.code || '-'}</td>
                      <td className="p-4 text-sm text-neutral-text-secondary">{doctor.preferences || '-'}</td>
                      <td className="p-4">
                        <div className="flex gap-2">
                          <Link
                            href={`/doctor-schedule?doctor_id=${doctor.id}`}
                            className="btn-secondary text-xs"
                            title="View Schedule"
                          >
                            Schedule
                          </Link>
                          <Link
                            href={`/doctor-management/${doctor.id}/preferences`}
                            className="btn-secondary text-xs"
                            title="Manage Preferences"
                          >
                            Preferences
                          </Link>
                          <button
                            onClick={() => handleEdit(doctor)}
                            className="btn-secondary text-xs"
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => handleDelete(doctor.id)}
                            className="btn-secondary text-xs text-red-600 hover:text-red-700"
                          >
                            Delete
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-neutral-bg-secondary rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-semibold text-neutral-text-primary mb-4">
              {editingDoctor ? 'Edit Doctor' : 'Add Doctor'}
            </h2>
            <form onSubmit={handleSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-neutral-text-primary mb-1">
                    Name <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="text"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="input-field"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-neutral-text-primary mb-1">
                    Code
                  </label>
                  <input
                    type="text"
                    value={formData.code}
                    onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                    className="input-field"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-neutral-text-primary mb-1">
                    Preferences
                  </label>
                  <textarea
                    value={formData.preferences}
                    onChange={(e) => setFormData({ ...formData, preferences: e.target.value })}
                    className="input-field"
                    rows={3}
                    placeholder="Noted remark/preferences"
                  />
                </div>
              </div>
              <div className="flex gap-2 mt-6">
                <button type="submit" className="btn-primary flex-1">
                  {editingDoctor ? 'Update' : 'Create'}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowModal(false);
                    setEditingDoctor(null);
                  }}
                  className="btn-secondary flex-1"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Import Modal */}
      {showImportModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="card max-w-md w-full">
            <div className="p-6">
              <h2 className="text-xl font-semibold text-neutral-text-primary mb-6">Import Doctors from Excel</h2>
              <div className="mb-4">
                <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                  Excel File
                </label>
                <input
                  type="file"
                  accept=".xlsx,.xls"
                  onChange={(e) => setImportFile(e.target.files?.[0] || null)}
                  className="input-field"
                />
                <p className="mt-2 text-xs text-neutral-text-secondary">
                  Expected format (columns A-C):<br />
                  <strong>Column A:</strong> Name (required)<br />
                  <strong>Column B:</strong> Code (optional) - doctor code/nickname<br />
                  <strong>Column C:</strong> Preferences (optional) - noted remark/preferences<br />
                  <br />
                  <strong>Note:</strong> Doctor IDs are automatically generated. Header row is optional and will be auto-detected.
                </p>
              </div>
              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => {
                    setShowImportModal(false);
                    setImportFile(null);
                  }}
                  className="btn-secondary"
                  disabled={importing}
                >
                  Cancel
                </button>
                <button
                  onClick={handleImport}
                  disabled={!importFile || importing}
                  className="btn-primary bg-green-600 hover:bg-green-700 disabled:opacity-50"
                >
                  {importing ? 'Importing...' : 'Import'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
