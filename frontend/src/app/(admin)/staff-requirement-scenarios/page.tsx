'use client';

import { useState, useEffect } from 'react';
import {
  staffRequirementScenarioApi,
  StaffRequirementScenario,
  StaffRequirementScenarioCreate,
  ScenarioPositionRequirementCreate,
} from '@/lib/api/staff-requirement-scenario';
import { revenueLevelTierApi, RevenueLevelTier } from '@/lib/api/revenue-level-tier';
import { positionApi, Position } from '@/lib/api/position';

const DAY_NAMES = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

export default function StaffRequirementScenariosPage() {
  const [scenarios, setScenarios] = useState<StaffRequirementScenario[]>([]);
  const [tiers, setTiers] = useState<RevenueLevelTier[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [formData, setFormData] = useState<StaffRequirementScenarioCreate>({
    scenario_name: '',
    description: '',
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
  });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [scenariosData, tiersData, positionsData] = await Promise.all([
        staffRequirementScenarioApi.list(),
        revenueLevelTierApi.list(),
        positionApi.list(),
      ]);
      setScenarios(scenariosData.sort((a, b) => b.priority - a.priority || a.scenario_name.localeCompare(b.scenario_name)));
      setTiers(tiersData.sort((a, b) => a.level_number - b.level_number));
      setPositions(positionsData);
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingId) {
        await staffRequirementScenarioApi.update(editingId, formData);
      } else {
        await staffRequirementScenarioApi.create(formData);
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
      revenue_level_tier_id: scenario.revenue_level_tier_id,
      min_revenue: scenario.min_revenue,
      max_revenue: scenario.max_revenue,
      use_day_of_week_revenue: scenario.use_day_of_week_revenue,
      use_specific_date_revenue: scenario.use_specific_date_revenue,
      doctor_count: scenario.doctor_count,
      min_doctor_count: scenario.min_doctor_count,
      day_of_week: scenario.day_of_week,
      is_default: scenario.is_default,
      is_active: scenario.is_active,
      priority: scenario.priority,
      position_requirements: (scenario.position_requirements || []).map((req) => ({
        position_id: req.position_id,
        preferred_staff: req.preferred_staff,
        minimum_staff: req.minimum_staff,
        override_base: req.override_base,
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
    });
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
              <h3 className="font-semibold mb-2">Revenue Matching</h3>
              <div className="space-y-2">
                <label className="flex items-center">
                  <input
                    type="radio"
                    checked={formData.revenue_level_tier_id !== null}
                    onChange={() => setFormData({ ...formData, revenue_level_tier_id: tiers[0]?.id || null, min_revenue: null, max_revenue: null })}
                    className="mr-2"
                  />
                  Use Revenue Level Tier
                </label>
                {formData.revenue_level_tier_id !== null && (
                  <select
                    value={formData.revenue_level_tier_id || ''}
                    onChange={(e) => setFormData({ ...formData, revenue_level_tier_id: e.target.value || null })}
                    className="w-full px-3 py-2 border rounded-md"
                  >
                    <option value="">Select Tier</option>
                    {tiers.map((tier) => (
                      <option key={tier.id} value={tier.id}>
                        Level {tier.level_number}: {tier.level_name} ({tier.min_revenue.toLocaleString()} - {tier.max_revenue ? tier.max_revenue.toLocaleString() : '∞'} THB)
                      </option>
                    ))}
                  </select>
                )}

                <label className="flex items-center mt-4">
                  <input
                    type="radio"
                    checked={formData.revenue_level_tier_id === null && (formData.min_revenue !== null || formData.max_revenue !== null)}
                    onChange={() => setFormData({ ...formData, revenue_level_tier_id: null, min_revenue: 0, max_revenue: null })}
                    className="mr-2"
                  />
                  Use Direct Revenue Range
                </label>
                {formData.revenue_level_tier_id === null && (
                  <div className="grid grid-cols-2 gap-2">
                    <input
                      type="number"
                      placeholder="Min Revenue"
                      value={formData.min_revenue || ''}
                      onChange={(e) => setFormData({ ...formData, min_revenue: e.target.value ? parseFloat(e.target.value) : null })}
                      className="px-3 py-2 border rounded-md"
                    />
                    <input
                      type="number"
                      placeholder="Max Revenue"
                      value={formData.max_revenue || ''}
                      onChange={(e) => setFormData({ ...formData, max_revenue: e.target.value ? parseFloat(e.target.value) : null })}
                      className="px-3 py-2 border rounded-md"
                    />
                  </div>
                )}
              </div>

              <div className="mt-4 space-y-2">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={formData.use_day_of_week_revenue}
                    onChange={(e) => setFormData({ ...formData, use_day_of_week_revenue: e.target.checked })}
                    className="mr-2"
                  />
                  Use Day-of-Week Revenue (from branch_weekly_revenue)
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={formData.use_specific_date_revenue}
                    onChange={(e) => setFormData({ ...formData, use_specific_date_revenue: e.target.checked })}
                    className="mr-2"
                  />
                  Use Specific Date Revenue (from revenue_data, overrides DoW)
                </label>
              </div>
            </div>

            <div className="border-t pt-4">
              <h3 className="font-semibold mb-2">Doctor Count Matching</h3>
              <div className="space-y-2">
                <label className="flex items-center">
                  <input
                    type="radio"
                    checked={formData.doctor_count === null && formData.min_doctor_count === null}
                    onChange={() => setFormData({ ...formData, doctor_count: null, min_doctor_count: null })}
                    className="mr-2"
                  />
                  Any doctor count
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    checked={formData.doctor_count !== null}
                    onChange={() => setFormData({ ...formData, doctor_count: 2, min_doctor_count: null })}
                    className="mr-2"
                  />
                  Exact count:
                  <input
                    type="number"
                    min="0"
                    value={formData.doctor_count || ''}
                    onChange={(e) => setFormData({ ...formData, doctor_count: e.target.value ? parseInt(e.target.value) : null, min_doctor_count: null })}
                    className="ml-2 w-20 px-2 py-1 border rounded-md"
                    disabled={formData.doctor_count === null}
                  />
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    checked={formData.min_doctor_count !== null}
                    onChange={() => setFormData({ ...formData, min_doctor_count: 2, doctor_count: null })}
                    className="mr-2"
                  />
                  Minimum count:
                  <input
                    type="number"
                    min="0"
                    value={formData.min_doctor_count || ''}
                    onChange={(e) => setFormData({ ...formData, min_doctor_count: e.target.value ? parseInt(e.target.value) : null, doctor_count: null })}
                    className="ml-2 w-20 px-2 py-1 border rounded-md"
                    disabled={formData.min_doctor_count === null}
                  />
                </label>
              </div>
            </div>

            <div className="border-t pt-4">
              <h3 className="font-semibold mb-2">Day of Week Filter (Optional)</h3>
              <select
                value={formData.day_of_week !== null ? formData.day_of_week.toString() : ''}
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
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={formData.is_default}
                  onChange={(e) => setFormData({ ...formData, is_default: e.target.checked })}
                  className="mr-2"
                />
                Default
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
              if (scenario.revenue_level_tier_id) {
                const tier = tiers.find((t) => t.id === scenario.revenue_level_tier_id);
                if (tier) conditions.push(`Tier ${tier.level_number}`);
              }
              if (scenario.min_revenue !== null || scenario.max_revenue !== null) {
                const minFormatted = scenario.min_revenue ? scenario.min_revenue.toLocaleString('en-US') : '0';
                const maxFormatted = scenario.max_revenue ? scenario.max_revenue.toLocaleString('en-US') : '∞';
                conditions.push(`Rev: ${minFormatted}-${maxFormatted}`);
              }
              if (scenario.doctor_count !== null) conditions.push(`Doctors=${scenario.doctor_count}`);
              if (scenario.min_doctor_count !== null) conditions.push(`Doctors>=${scenario.min_doctor_count}`);
              if (scenario.day_of_week !== null) conditions.push(DAY_NAMES[scenario.day_of_week]);

              return (
                <tr key={scenario.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">{scenario.scenario_name}</td>
                  <td className="px-6 py-4 text-sm">{conditions.join(', ') || 'Default'}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">{scenario.priority}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    {scenario.is_active ? (
                      <span className="text-green-600">Active</span>
                    ) : (
                      <span className="text-gray-400">Inactive</span>
                    )}
                    {scenario.is_default && <span className="ml-2 text-blue-600">(Default)</span>}
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
