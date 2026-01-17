'use client';

import { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useUser } from '@/contexts/UserContext';
import { doctorApi, Doctor, DoctorAssignment } from '@/lib/api/doctor';
import { branchApi, Branch } from '@/lib/api/branch';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameDay, addMonths, subMonths } from 'date-fns';
import DoctorScheduleCalendar from '@/components/doctor/DoctorScheduleCalendar';
import DoctorScheduleEditor from '@/components/doctor/DoctorScheduleEditor';
import DoctorDefaultScheduleManager from '@/components/doctor/DoctorDefaultScheduleManager';
import DoctorScheduleOverridesManager from '@/components/doctor/DoctorScheduleOverridesManager';
import DoctorOverallSchedule from '@/components/doctor/DoctorOverallSchedule';

export default function DoctorSchedulePage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { user, loading: userLoading } = useUser();
  const [doctors, setDoctors] = useState<Doctor[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [selectedDoctorId, setSelectedDoctorId] = useState<string | null>(null);
  const [currentDate, setCurrentDate] = useState(new Date());
  const [branch, setBranch] = useState<Branch | null>(null);
  const [year, setYear] = useState(new Date().getFullYear());
  const [month, setMonth] = useState(new Date().getMonth() + 1);
  const [loading, setLoading] = useState(true);
  const [mainView, setMainView] = useState<'overall' | 'individual'>('overall');
  const [activeTab, setActiveTab] = useState<'schedule' | 'defaults' | 'overrides'>('schedule');
  const [showImportModal, setShowImportModal] = useState(false);
  const [importFile, setImportFile] = useState<File | null>(null);
  const [importing, setImporting] = useState(false);

  useEffect(() => {
    if (!userLoading && user && !['admin', 'area_manager', 'branch_manager'].includes(user.role || '')) {
      router.push('/dashboard');
      return;
    }
  }, [user, userLoading, router]);

  useEffect(() => {
    const doctorIdParam = searchParams.get('doctor_id');
    if (doctorIdParam) {
      setSelectedDoctorId(doctorIdParam);
    }
  }, [searchParams]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        if (user) {
          if (['admin', 'area_manager'].includes(user.role || '')) {
            // Admin/Area Manager: Load doctors and branches
            const [doctorsData, branchesData] = await Promise.all([
              doctorApi.list(),
              branchApi.list(),
            ]);
            setDoctors(doctorsData || []);
            setBranches(branchesData || []);
          } else if (user.role === 'branch_manager' && user.branch_id) {
            // Branch Manager: Load their branch
            const branchesData = await branchApi.list();
            const userBranch = branchesData.find(b => b.id === user.branch_id);
            if (userBranch) {
              setBranch(userBranch);
            }
          }
        }
      } catch (error: any) {
        console.error('Failed to fetch data:', error);
      } finally {
        setLoading(false);
      }
    };

    if (user) {
      fetchData();
    }
  }, [user]);

  // Switch to individual view when doctor is selected
  useEffect(() => {
    if (selectedDoctorId) {
      setMainView('individual');
    }
  }, [selectedDoctorId]);

  if (loading || userLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-neutral-text-secondary">Loading...</div>
      </div>
    );
  }

  if (!user || !['admin', 'area_manager', 'branch_manager'].includes(user.role || '')) {
    return null;
  }

  // Branch Manager View: Show branch-specific doctor schedule editor
  if (user.role === 'branch_manager') {
    if (!branch) {
      return (
        <div className="flex items-center justify-center p-8">
          <div className="text-lg">Loading branch information...</div>
        </div>
      );
    }

    return (
      <div className="w-full p-6">
        <div className="mb-6 flex items-center justify-between">
          <h1 className="text-3xl font-bold">Doctor Schedule</h1>
          <div className="flex gap-2">
            <input
              type="month"
              value={`${year}-${String(month).padStart(2, '0')}`}
              onChange={(e) => {
                const [y, m] = e.target.value.split('-').map(Number);
                setYear(y);
                setMonth(m);
              }}
              className="px-3 py-2 border border-gray-300 rounded-md"
            />
          </div>
        </div>

        <DoctorScheduleEditor branchId={branch.id} year={year} month={month} />
      </div>
    );
  }

  // Admin/Area Manager View: Show doctor selection and calendar
  const selectedDoctor = doctors.find(d => d.id === selectedDoctorId);

  const handleImportDefaultSchedules = async () => {
    if (!importFile) {
      alert('Please select a file');
      return;
    }

    setImporting(true);
    try {
      const result = await doctorApi.importDefaultSchedules(importFile);
      let message = `Import completed! ${result.imported || 0} default schedule(s) imported`;
      if (result.off_days_set && result.off_days_set > 0) {
        message += ` and ${result.off_days_set} off day(s) set`;
      }
      message += '.';
      
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
      // Optionally refresh the selected doctor's schedules if one is selected
      if (selectedDoctorId) {
        // The component will refresh when the tab is viewed
      }
    } catch (error: any) {
      const errorMsg = error.response?.data?.error || 'Failed to import default schedules';
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
      } else {
        alert(errorMsg);
      }
    } finally {
      setImporting(false);
    }
  };

  return (
    <div className="p-6">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">Doctor Schedule Management</h1>
          <p className="text-sm text-neutral-text-secondary">Manage doctor schedules - assign doctors to branches by date</p>
        </div>
        <button
          onClick={() => setShowImportModal(true)}
          className="btn-primary bg-green-600 hover:bg-green-700"
        >
          Import Default Schedules
        </button>
      </div>

      {/* Main View Tabs */}
      <div className="card mb-6">
        <div className="border-b border-neutral-border">
          <nav className="flex -mb-px">
            <button
              onClick={() => {
                setMainView('overall');
                setSelectedDoctorId(null);
              }}
              className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                mainView === 'overall'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-neutral-text-secondary hover:text-neutral-text-primary hover:border-neutral-border'
              }`}
            >
              Overall Schedule
            </button>
            <button
              onClick={() => setMainView('individual')}
              className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                mainView === 'individual'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-neutral-text-secondary hover:text-neutral-text-primary hover:border-neutral-border'
              }`}
            >
              Individual Doctor
            </button>
          </nav>
        </div>
      </div>

      {mainView === 'overall' ? (
        <DoctorOverallSchedule doctors={doctors} branches={branches} />
      ) : (
        <>
          <div className="card mb-6">
            <div className="p-4">
              <label className="block text-sm font-medium text-neutral-text-primary mb-2">
                Select Doctor
              </label>
              <select
                value={selectedDoctorId || ''}
                onChange={(e) => {
                  setSelectedDoctorId(e.target.value || null);
                  setActiveTab('schedule'); // Reset to schedule tab when changing doctor
                }}
                className="input-field"
              >
                <option value="">-- Select a doctor --</option>
                {doctors.map((doctor) => (
                  <option key={doctor.id} value={doctor.id}>
                    {doctor.name} {doctor.code ? `(${doctor.code})` : ''}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {selectedDoctorId && selectedDoctor && (
            <>
              {/* Tabs */}
              <div className="card mb-6">
                <div className="border-b border-neutral-border">
                  <nav className="flex -mb-px">
                    <button
                      onClick={() => setActiveTab('schedule')}
                      className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                        activeTab === 'schedule'
                          ? 'border-blue-500 text-blue-600'
                          : 'border-transparent text-neutral-text-secondary hover:text-neutral-text-primary hover:border-neutral-border'
                      }`}
                    >
                      Schedule Calendar
                    </button>
                    <button
                      onClick={() => setActiveTab('defaults')}
                      className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                        activeTab === 'defaults'
                          ? 'border-blue-500 text-blue-600'
                          : 'border-transparent text-neutral-text-secondary hover:text-neutral-text-primary hover:border-neutral-border'
                      }`}
                    >
                      Default Schedule
                    </button>
                    <button
                      onClick={() => setActiveTab('overrides')}
                      className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                        activeTab === 'overrides'
                          ? 'border-blue-500 text-blue-600'
                          : 'border-transparent text-neutral-text-secondary hover:text-neutral-text-primary hover:border-neutral-border'
                      }`}
                    >
                      Schedule Overrides
                    </button>
                  </nav>
                </div>
              </div>

              {/* Tab Content */}
              {activeTab === 'schedule' && (
                <DoctorScheduleCalendar
                  doctor={selectedDoctor}
                  branches={branches}
                  currentDate={currentDate}
                  onDateChange={setCurrentDate}
                />
              )}

              {activeTab === 'defaults' && (
                <DoctorDefaultScheduleManager doctor={selectedDoctor} />
              )}

              {activeTab === 'overrides' && (
                <DoctorScheduleOverridesManager
                  doctor={selectedDoctor}
                  currentDate={currentDate}
                  onDateChange={setCurrentDate}
                />
              )}
            </>
          )}
        </>
      )}

      {/* Import Default Schedules Modal */}
      {showImportModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="card max-w-md w-full">
            <div className="p-6">
              <h2 className="text-xl font-semibold text-neutral-text-primary mb-6">Import Default Schedules from Excel</h2>
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
                  <strong>Column A:</strong> Doctor Code (required)<br />
                  <strong>Column B:</strong> Day of Week (required) - 1=Monday, 2=Tuesday, ..., 7=Sunday<br />
                  <strong>Column C:</strong> Branch Code or Branch Name (optional - leave empty or use "OFF"/"Off Day" for off days)<br />
                  <br />
                  <strong>Note:</strong> This will import default schedules for all doctors in the file. If a doctor has duplicate branches on the same workday, the last imported entry will be used. Empty branch or "OFF" means the doctor is off on that day. Header row is optional and will be auto-detected.
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
                  onClick={handleImportDefaultSchedules}
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
    </div>
  );
}
