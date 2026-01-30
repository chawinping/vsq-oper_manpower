'use client';

import { useState, useEffect } from 'react';
import {
  staffRequirementScenarioApi,
  StaffRequirementScenario,
  StaffRequirementScenarioCreate,
  ScenarioPositionRequirementCreate,
  ScenarioSpecificStaffRequirementCreate,
} from '@/lib/api/staff-requirement-scenario';
import { positionApi, Position } from '@/lib/api/position';
import { doctorApi, Doctor } from '@/lib/api/doctor';
import { branchApi, Branch } from '@/lib/api/branch';
import { staffApi, Staff } from '@/lib/api/staff';

const DAY_NAMES = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

export default function StaffRequirementScenariosPage() {
  const [scenarios, setScenarios] = useState<StaffRequirementScenario[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [doctors, setDoctors] = useState<Doctor[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [staff, setStaff] = useState<Staff[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [formData, setFormData] = useState<StaffRequirementScenarioCreate>({
    scenario_name: '',
    description: '',
    doctor_id: null,
    branch_id: null,
    revenue_level_tier_id: null,
    min_revenue: null,
    max_revenue: null,
    use_day_of_week_revenue: true,
    use_specific_date_revenue: false,
    doctor_count: null,
    min_doctor_count: null,
    day_of_week: null,
    is_default: false,
    is_active: true,
    priority: 0,
    position_requirements: [],
    specific_staff_requirements: [],
  });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [scenariosData, positionsData, doctorsData, branchesData, staffData] = await Promise.all([
        staffRequirementScenarioApi.list(),
        positionApi.list(),
        doctorApi.list(),
        branchApi.list(),
        staffApi.list({}),
      ]);
      setScenarios(scenariosData.sort((a, b) => b.priority - a.priority || a.scenario_name.localeCompare(b.scenario_name)));
      setPositions(positionsData);
      setDoctors(doctorsData.sort((a, b) => a.name.localeCompare(b.name)));
      setBranches(branchesData.sort((a, b) => a.name.localeCompare(b.name)));
      setStaff(staffData.sort((a, b) => a.name.localeCompare(b.name)));
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      // Convert "none" string to null for doctor_id and branch_id
      // Always set is_default to false (no default scenarios allowed)
      const submitData = {
        ...formData,
        doctor_id: formData.doctor_id === 'none' ? null : formData.doctor_id || null,
        branch_id: formData.branch_id === 'none' ? null : formData.branch_id || null,
        is_default: false,
      };
      if (editingId) {
        await staffRequirementScenarioApi.update(editingId, submitData);
      } else {
        await staffRequirementScenarioApi.create(submitData);
      }
      await loadData();
      resetForm();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save scenario');
    }
  };

  const handleEdit = (scenario: StaffRequirementScenario) => {
    setEditingId(scenario.id);
    setFormData({
      scenario_name: scenario.scenario_name,
      description: scenario.description || '',
      doctor_id: scenario.doctor_id,
      branch_id: scenario.branch_id,
      revenue_level_tier_id: scenario.revenue_level_tier_id,
      min_revenue: scenario.min_revenue,
      max_revenue: scenario.max_revenue,
      use_day_of_week_revenue: scenario.use_day_of_week_revenue,
      use_specific_date_revenue: scenario.use_specific_date_revenue,
      doctor_count: scenario.doctor_count,
      min_doctor_count: scenario.min_doctor_count,
      day_of_week: scenario.day_of_week,
      is_default: false,
      is_active: scenario.is_active,
      priority: scenario.priority,
      position_requirements: (scenario.position_requirements || []).map((req) => ({
        position_id: req.position_id,
        preferred_staff: req.preferred_staff,
        minimum_staff: req.minimum_staff,
        override_base: req.override_base,
      })),
      specific_staff_requirements: (scenario.specific_staff_requirements || []).map((req) => ({
        staff_id: req.staff_id,
      })),
    });
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this scenario?')) return;
    try {
      await staffRequirementScenarioApi.delete(id);
      await loadData();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete scenario');
    }
  };

  const addPositionRequirement = () => {
    setFormData({
      ...formData,
      position_requirements: [
        ...(formData.position_requirements || []),
        {
          position_id: positions[0]?.id || '',
          preferred_staff: 0,
          minimum_staff: 0,
          override_base: false,
        },
      ],
    });
  };

  const updatePositionRequirement = (index: number, field: keyof ScenarioPositionRequirementCreate, value: any) => {
    const updated = [...(formData.position_requirements || [])];
    updated[index] = { ...updated[index], [field]: value };
    setFormData({ ...formData, position_requirements: updated });
  };

  const removePositionRequirement = (index: number) => {
    const updated = [...(formData.position_requirements || [])];
    updated.splice(index, 1);
    setFormData({ ...formData, position_requirements: updated });
  };

  const resetForm = () => {
    setShowForm(false);
    setEditingId(null);
    setFormData({
      scenario_name: '',
      description: '',
      doctor_id: null,
      branch_id: null,
      revenue_level_tier_id: null,
      min_revenue: null,
      max_revenue: null,
      use_day_of_week_revenue: true,
      use_specific_date_revenue: false,
      doctor_count: null,
      min_doctor_count: null,
      day_of_week: null,
      is_default: false,
      is_active: true,
      priority: 0,
      position_requirements: [],
      specific_staff_requirements: [],
    });
  };

  const addSpecificStaffRequirement = () => {
    setFormData({
      ...formData,
      specific_staff_requirements: [
        ...(formData.specific_staff_requirements || []),
        {
          staff_id: staff[0]?.id || '',
        },
      ],
    });
  };

  const updateSpecificStaffRequirement = (index: number, field: keyof ScenarioSpecificStaffRequirementCreate, value: any) => {
    const updated = [...(formData.specific_staff_requirements || [])];
    updated[index] = { ...updated[index], [field]: value };
    setFormData({ ...formData, specific_staff_requirements: updated });
  };

  const removeSpecificStaffRequirement = (index: number) => {
    const updated = [...(formData.specific_staff_requirements || [])];
    updated.splice(index, 1);
    setFormData({ ...formData, specific_staff_requirements: updated });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  return (
    <div className="w-full p-6">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-3xl font-bold">Staff Requirement Scenarios</h1>
        <button
          onClick={() => setShowForm(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          + Create Scenario
        </button>
      </div>

      {showForm && (
        <div className="mb-6 p-6 bg-white rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">{editingId ? 'Edit Scenario' : 'Create Scenario'}</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">Scenario Name *</label>
                <input
                  type="text"
                  value={formData.scenario_name}
                  onChange={(e) => setFormData({ ...formData, scenario_name: e.target.value })}
                  className="w-full px-3 py-2 border rounded-md"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Priority</label>
                <input
                  type="number"
                  value={formData.priority}
                  onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) })}
                  className="w-full px-3 py-2 border rounded-md"
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Description</label>
              <textarea
                value={formData.description || ''}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                className="w-full px-3 py-2 border rounded-md"
                rows={2}
              />
            </div>

            <div className="border-t pt-4">
              <h3 className="font-semibold mb-2">Doctor Filter (Optional)</h3>
              <select
                value={formData.doctor_id || ''}
                onChange={(e) => setFormData({ ...formData, doctor_id: e.target.value || null })}
                className="w-full px-3 py-2 border rounded-md"
              >
                <option value="">Any Doctor</option>
                <option value="none">None (No Doctor)</option>
                {doctors.map((doctor) => (
                  <option key={doctor.id} value={doctor.id}>
                    {doctor.name} {doctor.code ? `(${doctor.code})` : ''}
                  </option>
                ))}
              </select>
            </div>

            <div className="border-t pt-4">
              <h3 className="font-semibold mb-2">Branch Filter (Optional)</h3>
              <select
                value={formData.branch_id || ''}
                onChange={(e) => setFormData({ ...formData, branch_id: e.target.value || null })}
                className="w-full px-3 py-2 border rounded-md"
              >
                <option value="">Any Branch</option>
                <option value="none">None (No Branch)</option>
                {branches.map((branch) => (
                  <option key={branch.id} value={branch.id}>
                    {branch.name} ({branch.code})
                  </option>
                ))}
              </select>
            </div>

            <div className="border-t pt-4">
              <h3 className="font-semibold mb-2">Day of Week Filter (Optional)</h3>
              <select
                value={formData.day_of_week != null ? formData.day_of_week.toString() : ''}
                onChange={(e) => setFormData({ ...formData, day_of_week: e.target.value ? parseInt(e.target.value) : null })}
                className="w-full px-3 py-2 border rounded-md"
              >
                <option value="">Any Day</option>
                {DAY_NAMES.map((day, index) => (
                  <option key={index} value={index.toString()}>
                    {day}
                  </option>
                ))}
              </select>
            </div>

            <div className="border-t pt-4">
              <h3 className="font-semibold mb-2">Position Requirements</h3>
              {(formData.position_requirements || []).map((req, index) => (
                <div key={index} className="mb-2 p-3 border rounded-md">
                  <div className="grid grid-cols-4 gap-2">
                    <select
                      value={req.position_id}
                      onChange={(e) => updatePositionRequirement(index, 'position_id', e.target.value)}
                      className="px-2 py-1 border rounded-md"
                    >
                      {positions.map((pos) => (
                        <option key={pos.id} value={pos.id}>
                          {pos.name}
                        </option>
                      ))}
                    </select>
                    <input
                      type="number"
                      placeholder="Preferred"
                      value={req.preferred_staff}
                      onChange={(e) => updatePositionRequirement(index, 'preferred_staff', parseInt(e.target.value) || 0)}
                      className="px-2 py-1 border rounded-md"
                    />
                    <input
                      type="number"
                      placeholder="Minimum"
                      value={req.minimum_staff}
                      onChange={(e) => updatePositionRequirement(index, 'minimum_staff', parseInt(e.target.value) || 0)}
                      className="px-2 py-1 border rounded-md"
                    />
                    <div className="flex items-center gap-2">
                      <label className="flex items-center text-sm">
                        <input
                          type="checkbox"
                          checked={req.override_base}
                          onChange={(e) => updatePositionRequirement(index, 'override_base', e.target.checked)}
                          className="mr-1"
                        />
                        Override
                      </label>
                      <button
                        type="button"
                        onClick={() => removePositionRequirement(index)}
                        className="text-red-600 hover:text-red-800"
                      >
                        Remove
                      </button>
                    </div>
                  </div>
                </div>
              ))}
              <button
                type="button"
                onClick={addPositionRequirement}
                className="mt-2 px-3 py-1 text-sm bg-gray-200 rounded-md hover:bg-gray-300"
              >
                + Add Position Requirement
              </button>
            </div>

            <div className="border-t pt-4">
              <h3 className="font-semibold mb-2">Specific Staff Requirements</h3>
              <p className="text-sm text-gray-600 mb-2">Select specific staff members that must be assigned</p>
              {(formData.specific_staff_requirements || []).map((req, index) => (
                <div key={index} className="mb-2 p-3 border rounded-md">
                  <div className="flex items-center gap-2">
                    <select
                      value={req.staff_id}
                      onChange={(e) => updateSpecificStaffRequirement(index, 'staff_id', e.target.value)}
                      className="flex-1 px-2 py-1 border rounded-md"
                    >
                      <option value="">Select Staff</option>
                      {staff.map((s) => (
                        <option key={s.id} value={s.id}>
                          {s.name} {s.nickname ? `(${s.nickname})` : ''} - {s.position?.name || 'Unknown Position'}
                        </option>
                      ))}
                    </select>
                    <button
                      type="button"
                      onClick={() => removeSpecificStaffRequirement(index)}
                      className="px-3 py-1 text-sm text-red-600 hover:text-red-800 border border-red-300 rounded-md hover:bg-red-50"
                    >
                      Remove
                    </button>
                  </div>
                </div>
              ))}
              <button
                type="button"
                onClick={addSpecificStaffRequirement}
                className="mt-2 px-3 py-1 text-sm bg-gray-200 rounded-md hover:bg-gray-300"
              >
                + Add Specific Staff Requirement
              </button>
            </div>

            <div className="border-t pt-4 flex gap-2">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={formData.is_active}
                  onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                  className="mr-2"
                />
                Active
              </label>
            </div>

            <div className="flex gap-2">
              <button
                type="submit"
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                {editingId ? 'Update' : 'Create'}
              </button>
              <button
                type="button"
                onClick={resetForm}
                className="px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Conditions</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Priority</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {scenarios.map((scenario) => {
              const conditions = [];
              if (scenario.doctor_id) {
                const doctor = doctors.find((d) => d.id === scenario.doctor_id);
                if (doctor) conditions.push(`Doctor: ${doctor.name}`);
              }
              if (scenario.branch_id) {
                const branch = branches.find((b) => b.id === scenario.branch_id);
                if (branch) conditions.push(`Branch: ${branch.name}`);
              }
              if (scenario.day_of_week !== null) conditions.push(DAY_NAMES[scenario.day_of_week]);
              if (scenario.specific_staff_requirements && scenario.specific_staff_requirements.length > 0) {
                conditions.push(`${scenario.specific_staff_requirements.length} specific staff`);
              }

              return (
                <tr key={scenario.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">{scenario.scenario_name}</td>
                  <td className="px-6 py-4 text-sm">{conditions.join(', ') || 'No conditions'}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">{scenario.priority}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    {scenario.is_active ? (
                      <span className="text-green-600">Active</span>
                    ) : (
                      <span className="text-gray-400">Inactive</span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    <button
                      onClick={() => handleEdit(scenario)}
                      className="text-blue-600 hover:text-blue-800 mr-3"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => handleDelete(scenario.id)}
                      className="text-red-600 hover:text-red-800"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
