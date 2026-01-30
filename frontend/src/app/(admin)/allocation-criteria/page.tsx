'use client';

import { useState, useEffect } from 'react';
import { 
  allocationCriteriaApi, 
  AllocationCriteriaConfig,
  CriterionID,
  CRITERION_ZEROTH,
  CRITERION_FIRST,
  CRITERION_SECOND,
  CRITERION_THIRD,
  CRITERION_FOURTH,
} from '@/lib/api/allocation-criteria';

interface CriteriaInfo {
  id: CriterionID;
  name: string;
  description: string;
  details: string[];
  icon: string;
  color: string;
  example: string;
}

const criteriaInfoMap: Record<CriterionID, CriteriaInfo> = {
  [CRITERION_ZEROTH]: {
    id: CRITERION_ZEROTH,
    name: 'Doctor Preferences',
    description: 'Prioritizes branches based on doctor-specific staff requirements and preferences',
    details: [
      'Considers doctor preferences for specific staff positions',
      'Evaluates rotation staff requirements from doctor profiles',
      'Can be enabled/disabled as a filter or scoring factor',
      'When enabled, branches meeting doctor preferences get higher priority'
    ],
    icon: '👨‍⚕️',
    color: 'purple',
    example: 'If Doctor A requires 3 doctor assistants, branches with this requirement get higher priority'
  },
  [CRITERION_FIRST]: {
    id: CRITERION_FIRST,
    name: 'Branch-Level Variables',
    description: 'Evaluates universal branch metrics that indicate activity level and demand',
    details: [
      'Skin revenue: Total revenue from skin-related services',
      'Laser YAG revenue: Revenue from laser treatments',
      'IV Drip cases: Number of IV drip procedures',
      'Slim Pen cases: Number of slim pen procedures',
      'Doctor count: Number of doctors working at the branch'
    ],
    icon: '📊',
    color: 'blue',
    example: 'A branch with high skin revenue (800K THB) and 4 doctors will score higher than a branch with low revenue (200K THB) and 2 doctors'
  },
  [CRITERION_SECOND]: {
    id: CRITERION_SECOND,
    name: 'Preferred Staff Shortage',
    description: 'Measures how far below the preferred (designated) staff quota each position is',
    details: [
      'Compares current staff count to designated quota',
      'Higher shortage = higher priority for allocation',
      'Helps maintain optimal staffing levels',
      'Ensures branches reach their preferred staffing targets'
    ],
    icon: '⭐',
    color: 'yellow',
    example: 'If a branch needs 5 nurses (preferred) but only has 3, it gets a higher priority score than a branch with 4 nurses'
  },
  [CRITERION_THIRD]: {
    id: CRITERION_THIRD,
    name: 'Minimum Staff Shortage',
    description: 'Critical priority - identifies branches that are below minimum required staffing levels',
    details: [
      'Highest priority criteria - ensures operational safety',
      'Compares current staff to minimum required threshold',
      'Branches below minimum get maximum priority score',
      'Must be addressed before other allocation needs'
    ],
    icon: '🚨',
    color: 'red',
    example: 'A branch that needs minimum 3 nurses but only has 1 will get the highest priority, regardless of other factors'
  },
  [CRITERION_FOURTH]: {
    id: CRITERION_FOURTH,
    name: 'Branch Type Staff Groups',
    description: 'Evaluates staffing needs based on branch type classifications and staff group requirements',
    details: [
      'Considers branch type (e.g., Premium, Standard, Express)',
      'Evaluates staff group requirements for each branch type',
      'Ensures branches meet their type-specific staffing standards',
      'Helps maintain consistency across branch types'
    ],
    icon: '🏢',
    color: 'green',
    example: 'A Premium branch type requires specific staff groups (e.g., Senior Nurses, Specialists). If these groups are understaffed, priority increases'
  },
};

const defaultPriorityOrder: CriterionID[] = [
  CRITERION_THIRD,  // Priority 1: Minimum staff shortage (highest)
  CRITERION_SECOND, // Priority 2: Preferred staff shortage
  CRITERION_FIRST,  // Priority 3: Branch-level variables
  CRITERION_FOURTH, // Priority 4: Branch type staff groups
  CRITERION_ZEROTH, // Priority 5: Doctor preferences (lowest)
];

export default function AllocationCriteriaPage() {
  const [config, setConfig] = useState<AllocationCriteriaConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [priorityOrder, setPriorityOrder] = useState<CriterionID[]>(defaultPriorityOrder);
  const [enableDoctorPreferences, setEnableDoctorPreferences] = useState(false);
  const [draggedItem, setDraggedItem] = useState<CriterionID | null>(null);

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      setLoading(true);
      const data = await allocationCriteriaApi.getPriorityOrder();
      setConfig(data);
      setPriorityOrder(data.priority_order.length > 0 ? data.priority_order : defaultPriorityOrder);
      setEnableDoctorPreferences(data.enable_doctor_preferences);
    } catch (error) {
      console.error('Failed to load criteria config:', error);
      // Use defaults on error
      setPriorityOrder(defaultPriorityOrder);
      setEnableDoctorPreferences(false);
    } finally {
      setLoading(false);
    }
  };

  const handleDragStart = (e: React.DragEvent, criterionId: CriterionID) => {
    setDraggedItem(criterionId);
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/html', criterionId);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
  };

  const handleDrop = (e: React.DragEvent, targetIndex: number) => {
    e.preventDefault();
    if (draggedItem === null) return;

    const newOrder = [...priorityOrder];
    const draggedIndex = newOrder.indexOf(draggedItem);

    if (draggedIndex === -1) return;

    // Remove dragged item from its current position
    newOrder.splice(draggedIndex, 1);
    // Insert at new position
    newOrder.splice(targetIndex, 0, draggedItem);

    setPriorityOrder(newOrder);
    setDraggedItem(null);
  };

  const handleDragEnd = () => {
    setDraggedItem(null);
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      await allocationCriteriaApi.updatePriorityOrder({
        priority_order: priorityOrder,
        enable_doctor_preferences: enableDoctorPreferences,
      });
      await loadConfig();
      alert('Criteria priority order saved successfully!');
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to save criteria priority order');
    } finally {
      setSaving(false);
    }
  };

  const handleReset = async () => {
    if (!confirm('Are you sure you want to reset to default priority order?')) return;
    try {
      setSaving(true);
      const data = await allocationCriteriaApi.resetPriorityOrder();
      setPriorityOrder(data.priority_order);
      setEnableDoctorPreferences(data.enable_doctor_preferences);
      alert('Criteria priority order reset to defaults!');
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to reset criteria priority order');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading allocation criteria configuration...</div>
      </div>
    );
  }

  return (
    <div className="w-full p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Allocation Criteria Priority Configuration</h1>
        <p className="text-gray-600">
          Configure the priority order for allocation criteria. Criteria are evaluated in strict priority order:
          Priority 1 (highest) is evaluated first, and only if candidates are equal on Priority 1 do we consider Priority 2, and so on.
        </p>
      </div>

      {/* Info Banner */}
      <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
        <div className="flex items-start">
          <div className="text-2xl mr-3">ℹ️</div>
          <div>
            <h3 className="font-semibold text-blue-900 mb-1">How Priority Ranking Works</h3>
            <p className="text-sm text-blue-800">
              The system uses strict priority ordering (lexicographic sorting). The first criterion that differentiates 
              between branch-position combinations determines the ranking. Lower priority criteria are only considered 
              when higher priority criteria scores are equal. Drag and drop items to reorder priorities.
            </p>
          </div>
        </div>
      </div>

      {/* Priority Order List */}
      <div className="mb-8">
        <h2 className="text-xl font-semibold mb-4">Priority Order (Drag to Reorder)</h2>
        <div className="space-y-2">
          {priorityOrder.map((criterionId, index) => {
            const info = criteriaInfoMap[criterionId];
            const isDragging = draggedItem === criterionId;
            const isZeroth = criterionId === CRITERION_ZEROTH;
            const isDisabled = isZeroth && !enableDoctorPreferences;

            const colorClasses = {
              purple: 'border-purple-300 bg-purple-50',
              blue: 'border-blue-300 bg-blue-50',
              yellow: 'border-yellow-300 bg-yellow-50',
              red: 'border-red-300 bg-red-50',
              green: 'border-green-300 bg-green-50',
            };

            return (
              <div
                key={criterionId}
                draggable={!isDisabled}
                onDragStart={(e) => !isDisabled && handleDragStart(e, criterionId)}
                onDragOver={handleDragOver}
                onDrop={(e) => handleDrop(e, index)}
                onDragEnd={handleDragEnd}
                className={`
                  border-2 rounded-lg p-4 cursor-move
                  ${colorClasses[info.color as keyof typeof colorClasses]}
                  ${isDragging ? 'opacity-50' : ''}
                  ${isDisabled ? 'opacity-60 cursor-not-allowed' : 'hover:shadow-md transition-shadow'}
                `}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-4 flex-1">
                    <div className="text-2xl">⋮⋮</div>
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <span className="text-3xl">{info.icon}</span>
                        <div>
                          <h3 className="text-lg font-bold">{info.name}</h3>
                          <p className="text-sm text-gray-600">{info.description}</p>
                        </div>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="text-2xl font-bold text-gray-700">Priority {index + 1}</div>
                      {index === 0 && (
                        <div className="text-xs text-gray-500 mt-1">(Highest)</div>
                      )}
                      {index === priorityOrder.length - 1 && (
                        <div className="text-xs text-gray-500 mt-1">(Lowest)</div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Criteria Details */}
      <div className="mb-8">
        <h2 className="text-xl font-semibold mb-4">Criteria Details</h2>
        <div className="space-y-4">
          {priorityOrder.map((criterionId, index) => {
            const info = criteriaInfoMap[criterionId];
            const isZeroth = criterionId === CRITERION_ZEROTH;
            const isDisabled = isZeroth && !enableDoctorPreferences;

            const colorClasses = {
              purple: 'border-purple-300 bg-purple-50',
              blue: 'border-blue-300 bg-blue-50',
              yellow: 'border-yellow-300 bg-yellow-50',
              red: 'border-red-300 bg-red-50',
              green: 'border-green-300 bg-green-50',
            };

            return (
              <div
                key={criterionId}
                className={`border-2 rounded-lg p-6 ${colorClasses[info.color as keyof typeof colorClasses]} ${
                  isDisabled ? 'opacity-60' : ''
                }`}
              >
                <div className="flex items-start justify-between mb-4">
                  <div className="flex items-start">
                    <span className="text-4xl mr-4">{info.icon}</span>
                    <div>
                      <div className="flex items-center gap-2 mb-2">
                        <h3 className="text-xl font-bold">{info.name}</h3>
                        <span className="px-2 py-1 bg-white rounded text-sm font-semibold">
                          Priority {index + 1}
                        </span>
                      </div>
                      <p className="text-gray-700 mb-3">{info.description}</p>
                      
                      {/* Details List */}
                      <ul className="list-disc list-inside text-sm text-gray-600 mb-3 space-y-1">
                        {info.details.map((detail, idx) => (
                          <li key={idx}>{detail}</li>
                        ))}
                      </ul>

                      {/* Example */}
                      <div className="mt-3 p-3 bg-white rounded border border-gray-200">
                        <span className="text-xs font-semibold text-gray-500 uppercase">Example:</span>
                        <p className="text-sm text-gray-700 mt-1">{info.example}</p>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Enable Doctor Preferences Toggle */}
                {isZeroth && (
                  <div className="mt-4 pt-4 border-t border-gray-300">
                    <label className="flex items-center">
                      <input
                        type="checkbox"
                        checked={enableDoctorPreferences}
                        onChange={(e) => setEnableDoctorPreferences(e.target.checked)}
                        className="mr-2 w-4 h-4"
                      />
                      <span className="font-medium">Enable Doctor Preferences</span>
                    </label>
                    <p className="text-xs text-gray-600 ml-6 mt-1">
                      When enabled, doctor preferences are used as a filter and scoring factor
                    </p>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* Action Buttons */}
      <div className="flex gap-4 justify-end pt-6 border-t border-gray-200">
        <button
          onClick={handleReset}
          disabled={saving}
          className="px-6 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 disabled:opacity-50"
        >
          Reset to Defaults
        </button>
        <button
          onClick={handleSave}
          disabled={saving}
          className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {saving ? 'Saving...' : 'Save Configuration'}
        </button>
      </div>
    </div>
  );
}
