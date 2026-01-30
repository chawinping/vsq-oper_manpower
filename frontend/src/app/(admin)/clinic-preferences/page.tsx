'use client';

import { useState, useEffect } from 'react';
import {
  clinicPreferenceApi,
  ClinicWidePreference,
  ClinicWidePreferenceCreate,
  ClinicWidePreferenceUpdate,
  ClinicPreferenceCriteriaType,
  PreferencePositionRequirementCreate,
  PreferencePositionRequirementUpdate,
} from '@/lib/api/clinic-preference';
import { positionApi, Position } from '@/lib/api/position';

// Extended type for form position requirements that includes optional id for tracking
type FormPositionRequirement = PreferencePositionRequirementCreate & {
  id?: string;
};

const CRITERIA_TYPES: { value: ClinicPreferenceCriteriaType; label: string; unit: string }[] = [
  { value: 'skin_revenue', label: 'Skin Revenue', unit: 'THB' },
  { value: 'laser_yag_revenue', label: 'Laser YAG Revenue', unit: 'THB' },
  { value: 'iv_cases', label: 'IV Cases', unit: 'cases' },
  { value: 'slim_pen_cases', label: 'Slim Pen Cases', unit: 'cases' },
  { value: 'doctor_count', label: 'Doctor Count', unit: 'doctors' },
];

export default function ClinicPreferencesPage() {
  const [activeTab, setActiveTab] = useState<ClinicPreferenceCriteriaType>('skin_revenue');
  const [preferences, setPreferences] = useState<ClinicWidePreference[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingPreference, setEditingPreference] = useState<ClinicWidePreference | null>(null);
  const [viewingId, setViewingId] = useState<string | null>(null);
  const [formData, setFormData] = useState<ClinicWidePreferenceCreate>({
    criteria_type: 'skin_revenue',
    criteria_name: '',
    min_value: 0,
    max_value: null,
    is_active: true,
    display_order: 0,
    description: '',
    position_requirements: [],
  });

  useEffect(() => {
    loadData();
  }, [activeTab]);

  // Ensure min_value is always a valid number in formData
  useEffect(() => {
    if (formData.min_value === undefined || formData.min_value === null || (typeof formData.min_value === 'number' && isNaN(formData.min_value))) {
      setFormData((prev) => {
        if (prev.min_value === undefined || prev.min_value === null || (typeof prev.min_value === 'number' && isNaN(prev.min_value))) {
          return { ...prev, min_value: 0 };
        }
        return prev;
      });
    }
  }, [formData.min_value]);

  const loadData = async () => {
    try {
      setLoading(true);
      const [prefsData, positionsData] = await Promise.all([
        clinicPreferenceApi.list({ criteria_type: activeTab }),
        positionApi.list(),
      ]);
      setPreferences(prefsData.sort((a, b) => a.display_order - b.display_order || a.min_value - b.min_value));
      setPositions(positionsData);
    } catch (error) {
      console.error('Failed to load data:', error);
      alert('Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      // Ensure min_value is always a valid number (explicitly handle all edge cases)
      let minValue: number;
      if (formData.min_value === undefined || formData.min_value === null) {
        minValue = 0;
      } else if (typeof formData.min_value === 'string') {
        const parsed = parseFloat(formData.min_value);
        minValue = isNaN(parsed) ? 0 : parsed;
      } else if (typeof formData.min_value === 'number') {
        minValue = isNaN(formData.min_value) ? 0 : formData.min_value;
      } else {
        minValue = 0;
      }
      
      // Ensure min_value is non-negative
      minValue = Math.max(0, minValue);
      
      // Validate max_value >= min_value if max_value is set
      // For doctor_count, allow equality (max_value == min_value)
      // For other criteria types, require max_value > min_value
      if (formData.max_value != null) {
        if (activeTab === 'doctor_count') {
          // Allow equality for doctor_count
          if (formData.max_value < minValue) {
            alert('Max value must be greater than or equal to min value');
            return;
          }
        } else {
          // Require strict inequality for other criteria types
          if (formData.max_value <= minValue) {
            alert('Max value must be greater than min value');
            return;
          }
        }
      }
      
      if (editingId) {
        const updateData: ClinicWidePreferenceUpdate = {
          criteria_name: formData.criteria_name,
          min_value: minValue,
          max_value: formData.max_value,
          is_active: formData.is_active,
          display_order: formData.display_order,
          description: formData.description || null,
        };
        await clinicPreferenceApi.update(editingId, updateData);
        
        // Reload the preference to get current state of requirements before syncing
        const currentPreference = await clinicPreferenceApi.getById(editingId);
        
        // Sync position requirements
        const originalReqs = currentPreference?.position_requirements || [];
        const formReqs = (formData.position_requirements || []) as FormPositionRequirement[];
        
        // Create maps for tracking by both ID and position_id
        const originalReqByIdMap = new Map(
          originalReqs.map(req => [req.id, req])
        );
        const originalReqByPositionIdMap = new Map(
          originalReqs.map(req => [req.position_id, req])
        );
        
        // Create a set of form requirement IDs (those that have IDs are existing)
        const formReqIds = new Set(
          formReqs.map(req => req.id).filter((id): id is string => !!id)
        );
        // Track which position_ids are in the form
        const formPositionIds = new Set(
          formReqs.map(req => req.position_id).filter((id): id is string => !!id)
        );
        
        // Process each form requirement
        for (const formReq of formReqs) {
          if (!formReq.position_id) {
            // Skip requirements without a position_id
            continue;
          }
          
          const reqId = formReq.id;
          const existingByPositionId = originalReqByPositionIdMap.get(formReq.position_id);
          
          if (!reqId) {
            // This requirement has no ID - check if it's truly new or if position_id already exists
            if (existingByPositionId) {
              // Position already exists - update it instead of adding
              try {
                await clinicPreferenceApi.updatePositionRequirement(editingId, formReq.position_id, {
                  minimum_staff: formReq.minimum_staff,
                  preferred_staff: formReq.preferred_staff,
                  is_active: formReq.is_active ?? true,
                });
              } catch (error: any) {
                // If update fails (e.g., requirement was deleted), try adding it
                if (error.response?.status === 404) {
                  await clinicPreferenceApi.addPositionRequirement(editingId, {
                    position_id: formReq.position_id,
                    minimum_staff: formReq.minimum_staff,
                    preferred_staff: formReq.preferred_staff,
                    is_active: formReq.is_active ?? true,
                  });
                } else {
                  throw error;
                }
              }
            } else {
              // Truly new requirement - add it
              await clinicPreferenceApi.addPositionRequirement(editingId, {
                position_id: formReq.position_id,
                minimum_staff: formReq.minimum_staff,
                preferred_staff: formReq.preferred_staff,
                is_active: formReq.is_active ?? true,
              });
            }
          } else {
            // This is an existing requirement (has ID) - check if it needs updating
            const originalReq = originalReqByIdMap.get(reqId);
            if (originalReq) {
              // If position_id changed, delete old and create new
              if (originalReq.position_id !== formReq.position_id) {
                // Delete the old requirement
                await clinicPreferenceApi.deletePositionRequirement(editingId, originalReq.position_id);
                // Check if new position_id already exists (shouldn't happen, but be safe)
                const existingAtNewPosition = originalReqByPositionIdMap.get(formReq.position_id);
                if (existingAtNewPosition && existingAtNewPosition.id !== reqId) {
                  // Position already exists at new location - update it instead
                  await clinicPreferenceApi.updatePositionRequirement(editingId, formReq.position_id, {
                    minimum_staff: formReq.minimum_staff,
                    preferred_staff: formReq.preferred_staff,
                    is_active: formReq.is_active ?? true,
                  });
                } else {
                  // Create new requirement with new position_id
                  await clinicPreferenceApi.addPositionRequirement(editingId, {
                    position_id: formReq.position_id,
                    minimum_staff: formReq.minimum_staff,
                    preferred_staff: formReq.preferred_staff,
                    is_active: formReq.is_active ?? true,
                  });
                }
              } else {
                // Position ID unchanged, update other fields if changed
                if (
                  originalReq.minimum_staff !== formReq.minimum_staff ||
                  originalReq.preferred_staff !== formReq.preferred_staff ||
                  originalReq.is_active !== formReq.is_active
                ) {
                  try {
                    await clinicPreferenceApi.updatePositionRequirement(editingId, formReq.position_id, {
                      minimum_staff: formReq.minimum_staff,
                      preferred_staff: formReq.preferred_staff,
                      is_active: formReq.is_active ?? true,
                    });
                  } catch (error: any) {
                    // If update fails (e.g., requirement was deleted), try adding it
                    if (error.response?.status === 404) {
                      await clinicPreferenceApi.addPositionRequirement(editingId, {
                        position_id: formReq.position_id,
                        minimum_staff: formReq.minimum_staff,
                        preferred_staff: formReq.preferred_staff,
                        is_active: formReq.is_active ?? true,
                      });
                    } else {
                      throw error;
                    }
                  }
                }
              }
            }
          }
        }
        
        // Delete removed requirements (in original but not in form by position_id)
        for (const originalReq of originalReqs) {
          if (!formPositionIds.has(originalReq.position_id)) {
            await clinicPreferenceApi.deletePositionRequirement(editingId, originalReq.position_id);
          }
        }
      } else {
        const createData: ClinicWidePreferenceCreate = {
          criteria_type: activeTab,
          criteria_name: formData.criteria_name,
          min_value: minValue,
          max_value: formData.max_value ?? null,
          is_active: formData.is_active ?? true,
          display_order: formData.display_order ?? 0,
          description: formData.description || null,
          position_requirements: formData.position_requirements || [],
        };
        
        // Debug log in development
        if (process.env.NODE_ENV === 'development') {
          console.log('Creating preference with data:', createData);
        }
        
        await clinicPreferenceApi.create(createData);
      }
      await loadData();
      resetForm();
    } catch (error: any) {
      console.error('Submit error:', error);
      alert(error.response?.data?.error || 'Failed to save preference');
    }
  };

  const handleEdit = (pref: ClinicWidePreference) => {
    setEditingId(pref.id);
    setEditingPreference(pref);
    setFormData({
      criteria_type: pref.criteria_type,
      criteria_name: pref.criteria_name,
      min_value: pref.min_value,
      max_value: pref.max_value,
      is_active: pref.is_active,
      display_order: pref.display_order,
      description: pref.description || '',
      position_requirements: pref.position_requirements?.map((req) => ({
        id: req.id, // Preserve ID to track existing requirements
        position_id: req.position_id,
        minimum_staff: req.minimum_staff,
        preferred_staff: req.preferred_staff,
        is_active: req.is_active,
      })) || [],
    });
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this preference? This will also delete all position requirements.')) return;
    try {
      await clinicPreferenceApi.delete(id);
      await loadData();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete preference');
    }
  };

  const handleViewRequirements = async (id: string) => {
    if (viewingId === id) {
      setViewingId(null);
      return;
    }
    setViewingId(id);
    // Requirements are already loaded in the preference object
  };

  const handleAddPositionRequirement = async (preferenceId: string, data: PreferencePositionRequirementCreate) => {
    try {
      await clinicPreferenceApi.addPositionRequirement(preferenceId, data);
      await loadData();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to add position requirement');
    }
  };

  const handleUpdatePositionRequirement = async (
    preferenceId: string,
    positionId: string,
    data: PreferencePositionRequirementUpdate
  ) => {
    try {
      await clinicPreferenceApi.updatePositionRequirement(preferenceId, positionId, data);
      await loadData();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to update position requirement');
    }
  };

  const handleDeletePositionRequirement = async (preferenceId: string, positionId: string) => {
    if (!confirm('Are you sure you want to delete this position requirement?')) return;
    try {
      await clinicPreferenceApi.deletePositionRequirement(preferenceId, positionId);
      await loadData();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to delete position requirement');
    }
  };

  const resetForm = () => {
    setShowForm(false);
    setEditingId(null);
    setEditingPreference(null);
    setFormData({
      criteria_type: activeTab,
      criteria_name: '',
      min_value: 0,
      max_value: null,
      is_active: true,
      display_order: 0,
      description: '',
      position_requirements: [],
    });
  };

  const getCurrentCriteria = () => CRITERIA_TYPES.find((ct) => ct.value === activeTab)!;

  const formatValue = (value: number) => {
    const criteria = getCurrentCriteria();
    if (criteria.unit === 'THB') {
      return value.toLocaleString();
    }
    return value.toString();
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
        <h1 className="text-3xl font-bold">Clinic-Wide Preferences</h1>
        <button
          onClick={() => setShowForm(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          + Add Preference
        </button>
      </div>

      {/* Tabs */}
      <div className="mb-6 border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          {CRITERIA_TYPES.map((criteria) => (
            <button
              key={criteria.value}
              onClick={() => {
                setActiveTab(criteria.value);
                resetForm();
              }}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === criteria.value
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              {criteria.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Form */}
      {showForm && (
        <div className="mb-6 p-6 bg-white rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">
            {editingId ? 'Edit Preference' : 'Create Preference'} - {getCurrentCriteria().label}
          </h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">Preference Name</label>
                <input
                  type="text"
                  value={formData.criteria_name}
                  onChange={(e) => setFormData({ ...formData, criteria_name: e.target.value })}
                  className="w-full px-3 py-2 border rounded-md"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Display Order</label>
                <input
                  type="number"
                  value={formData.display_order}
                  onChange={(e) => setFormData({ ...formData, display_order: parseInt(e.target.value) || 0 })}
                  className="w-full px-3 py-2 border rounded-md"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">
                  Min {getCurrentCriteria().label} ({getCurrentCriteria().unit})
                </label>
                <input
                  type="number"
                  min="0"
                  step={getCurrentCriteria().unit === 'THB' ? '1000' : '1'}
                  value={formData.min_value ?? 0}
                  onChange={(e) => {
                    const inputValue = e.target.value;
                    if (inputValue === '' || inputValue === null || inputValue === undefined) {
                      setFormData({ ...formData, min_value: 0 });
                    } else {
                      const numValue = parseFloat(inputValue);
                      if (!isNaN(numValue) && numValue >= 0) {
                        setFormData({ ...formData, min_value: numValue });
                      } else {
                        setFormData({ ...formData, min_value: 0 });
                      }
                    }
                  }}
                  onBlur={(e) => {
                    // Ensure value is always set when field loses focus
                    const inputValue = e.target.value;
                    const numValue = inputValue === '' ? 0 : parseFloat(inputValue);
                    const finalValue = isNaN(numValue) || numValue < 0 ? 0 : numValue;
                    if (formData.min_value !== finalValue) {
                      setFormData({ ...formData, min_value: finalValue });
                    }
                  }}
                  className="w-full px-3 py-2 border rounded-md"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">
                  Max {getCurrentCriteria().label} ({getCurrentCriteria().unit}) - Leave empty for no limit
                </label>
                <input
                  type="number"
                  min="0"
                  step={getCurrentCriteria().unit === 'THB' ? '1000' : '1'}
                  value={formData.max_value || ''}
                  onChange={(e) =>
                    setFormData({ ...formData, max_value: e.target.value ? parseFloat(e.target.value) : null })
                  }
                  className="w-full px-3 py-2 border rounded-md"
                />
              </div>
              <div className="flex items-center">
                <input
                  type="checkbox"
                  id="is_active"
                  checked={formData.is_active}
                  onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                  className="mr-2"
                />
                <label htmlFor="is_active" className="text-sm font-medium">
                  Active
                </label>
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Description</label>
              <textarea
                value={formData.description || ''}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                className="w-full px-3 py-2 border rounded-md"
                rows={3}
              />
            </div>

            {/* Position Requirements Section */}
            <div className="mt-6">
              <h3 className="text-lg font-semibold mb-3">Position Requirements</h3>
              {formData.position_requirements?.map((req, index) => {
                const position = positions.find((p) => p.id === req.position_id);
                return (
                  <div key={index} className="mb-4 p-4 border rounded-md">
                    <div className="grid grid-cols-4 gap-4">
                      <div>
                        <label className="block text-sm font-medium mb-1">Position</label>
                        <select
                          value={req.position_id}
                          onChange={(e) => {
                            const newReqs = [...(formData.position_requirements || [])];
                            newReqs[index].position_id = e.target.value;
                            setFormData({ ...formData, position_requirements: newReqs });
                          }}
                          className="w-full px-3 py-2 border rounded-md"
                          required
                        >
                          <option value="">Select Position</option>
                          {positions.map((pos) => (
                            <option key={pos.id} value={pos.id}>
                              {pos.name}
                            </option>
                          ))}
                        </select>
                      </div>
                      <div>
                        <label className="block text-sm font-medium mb-1">Minimum Staff</label>
                        <input
                          type="number"
                          min="0"
                          value={req.minimum_staff}
                          onChange={(e) => {
                            const newReqs = [...(formData.position_requirements || [])];
                            newReqs[index].minimum_staff = parseInt(e.target.value) || 0;
                            setFormData({ ...formData, position_requirements: newReqs });
                          }}
                          className="w-full px-3 py-2 border rounded-md"
                          required
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium mb-1">Preferred Staff</label>
                        <input
                          type="number"
                          min={req.minimum_staff}
                          value={req.preferred_staff}
                          onChange={(e) => {
                            const newReqs = [...(formData.position_requirements || [])];
                            newReqs[index].preferred_staff = parseInt(e.target.value) || 0;
                            setFormData({ ...formData, position_requirements: newReqs });
                          }}
                          className="w-full px-3 py-2 border rounded-md"
                          required
                        />
                      </div>
                      <div className="flex items-end">
                        <button
                          type="button"
                          onClick={() => {
                            const newReqs = formData.position_requirements?.filter((_, i) => i !== index) || [];
                            setFormData({ ...formData, position_requirements: newReqs });
                          }}
                          className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700"
                        >
                          Remove
                        </button>
                      </div>
                    </div>
                  </div>
                );
              })}
              <button
                type="button"
                onClick={() => {
                  const newReqs = [
                    ...(formData.position_requirements || []),
                    { position_id: '', minimum_staff: 0, preferred_staff: 0, is_active: true },
                  ];
                  setFormData({ ...formData, position_requirements: newReqs });
                }}
                className="mt-2 px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700"
              >
                + Add Position Requirement
              </button>
            </div>

            <div className="flex gap-2">
              <button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">
                {editingId ? 'Update' : 'Create'}
              </button>
              <button type="button" onClick={resetForm} className="px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400">
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Preferences Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Range ({getCurrentCriteria().unit})
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Positions</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {preferences.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-6 py-4 text-center text-sm text-gray-500">
                  No preferences configured for {getCurrentCriteria().label}
                </td>
              </tr>
            ) : (
              preferences.map((pref) => (
                <>
                  <tr key={pref.id}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">{pref.criteria_name}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      {formatValue(pref.min_value)} - {pref.max_value ? formatValue(pref.max_value) : '∞'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`px-2 py-1 text-xs rounded-full ${
                          pref.is_active ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                        }`}
                      >
                        {pref.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      {pref.position_requirements?.length || 0} position(s)
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      <button
                        onClick={() => handleViewRequirements(pref.id)}
                        className="text-blue-600 hover:text-blue-800 mr-3"
                      >
                        {viewingId === pref.id ? 'Hide' : 'View'} Requirements
                      </button>
                      <button onClick={() => handleEdit(pref)} className="text-blue-600 hover:text-blue-800 mr-3">
                        Edit
                      </button>
                      <button onClick={() => handleDelete(pref.id)} className="text-red-600 hover:text-red-800">
                        Delete
                      </button>
                    </td>
                  </tr>
                  {viewingId === pref.id && pref.position_requirements && pref.position_requirements.length > 0 && (
                    <tr>
                      <td colSpan={5} className="px-6 py-4 bg-gray-50">
                        <div className="space-y-2">
                          <h4 className="font-semibold mb-2">Position Requirements:</h4>
                          <table className="min-w-full">
                            <thead>
                              <tr>
                                <th className="text-left text-xs font-medium text-gray-500 uppercase">Position</th>
                                <th className="text-left text-xs font-medium text-gray-500 uppercase">Minimum</th>
                                <th className="text-left text-xs font-medium text-gray-500 uppercase">Preferred</th>
                                <th className="text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
                              </tr>
                            </thead>
                            <tbody>
                              {pref.position_requirements.map((req) => (
                                <tr key={req.id}>
                                  <td className="text-sm">{req.position?.name || req.position_id}</td>
                                  <td className="text-sm">{req.minimum_staff}</td>
                                  <td className="text-sm">{req.preferred_staff}</td>
                                  <td>
                                    <button
                                      onClick={() => handleDeletePositionRequirement(pref.id, req.position_id)}
                                      className="text-red-600 hover:text-red-800 text-sm"
                                    >
                                      Remove
                                    </button>
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      </td>
                    </tr>
                  )}
                </>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
