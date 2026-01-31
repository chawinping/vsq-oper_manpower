'use client';

import { useState, useEffect } from 'react';
import { branchConfigApi, PositionQuota, PositionQuotaUpdate, BranchConstraints, ConstraintsUpdate, StaffGroupRequirement } from '@/lib/api/branch-config';
import { positionApi, Position } from '@/lib/api/position';
import { staffRequirementScenarioApi, ScenarioMatch, CalculatedRequirement } from '@/lib/api/staff-requirement-scenario';
import { revenueLevelTierApi, RevenueLevelTier } from '@/lib/api/revenue-level-tier';
import { branchApi, Branch } from '@/lib/api/branch';
import { branchTypeApi, BranchType } from '@/lib/api/branch-type';
import { staffGroupApi, StaffGroup } from '@/lib/api/staff-group';

interface BranchPositionQuotaConfigProps {
  branchId: string;
  onSave?: () => void;
}

export default function BranchPositionQuotaConfig({ branchId, onSave }: BranchPositionQuotaConfigProps) {
  const [positions, setPositions] = useState<Position[]>([]);
  const [quotas, setQuotas] = useState<Map<string, PositionQuota>>(new Map());
  const [constraints, setConstraints] = useState<Map<number, BranchConstraints>>(new Map());
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [savingConstraints, setSavingConstraints] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  
  // Scenario preview state
  const [previewDate, setPreviewDate] = useState<string>(new Date().toISOString().split('T')[0]);
  const [scenarioMatches, setScenarioMatches] = useState<ScenarioMatch[]>([]);
  const [calculatedRequirements, setCalculatedRequirements] = useState<Map<string, CalculatedRequirement>>(new Map());
  const [dayOfWeekRevenue, setDayOfWeekRevenue] = useState<number | null>(null);
  const [revenueTier, setRevenueTier] = useState<RevenueLevelTier | null>(null);
  const [doctorCount, setDoctorCount] = useState<number>(0);
  const [loadingPreview, setLoadingPreview] = useState(false);
  const [branch, setBranch] = useState<Branch | null>(null);
  const [branchType, setBranchType] = useState<BranchType | null>(null);
  const [branchTypes, setBranchTypes] = useState<BranchType[]>([]);
  const [savingBranchType, setSavingBranchType] = useState(false);
  const [staffGroups, setStaffGroups] = useState<StaffGroup[]>([]);

  // Target positions to configure (using Thai position names)
  const targetPositions = [
    'ผู้จัดการสาขา',           // Manager
    'รองผู้จัดการสาขา',         // Assistant Manager
    'ผู้ช่วยผู้จัดการสาขา',      // Assistant Manager (alternative)
    'ฟร้อนท์วนสาขา',            // Front Rotation (replaces Front 3)
    'ผู้ประสานงานคลินิก',        // Coordinator
    'ผู้ช่วยแพทย์',              // Doctor Assistant
    'พยาบาล',                   // Nurse
    'พนักงานต้อนรับ',            // Receptionist (replaces Front Laser)
    'ผู้ช่วย Laser Specialist',  // Laser Assistant
  ];

  useEffect(() => {
    loadData();
  }, [branchId]);

  const loadData = async () => {
    setLoading(true);
    setError(null);
    try {
      const [positionsData, quotasData, constraintsData, branchesData, branchTypesData, staffGroupsData] = await Promise.all([
        positionApi.list(),
        branchConfigApi.getQuotas(branchId),
        branchConfigApi.getConstraints(branchId),
        branchApi.list(),
        branchTypeApi.list(),
        staffGroupApi.list(),
      ]);

      // Filter to only active staff groups
      setStaffGroups((staffGroupsData || []).filter(group => group.is_active));

      // Set branch types list
      setBranchTypes(branchTypesData.filter(bt => bt.is_active));

      // Find branch and load branch type if exists
      const foundBranch = branchesData.find(b => b.id === branchId);
      if (foundBranch) {
        setBranch(foundBranch);
        // Load branch type if branch has one
        if (foundBranch.branch_type_id) {
          try {
            const bt = await branchTypeApi.getById(foundBranch.branch_type_id);
            setBranchType(bt);
          } catch (err) {
            console.error('Failed to load branch type:', err);
          }
        } else {
          setBranchType(null);
        }
      }

      setPositions(positionsData);

      // Filter positions to match target positions AND only show branch-type positions
      const filteredPositions = positionsData.filter((pos) => {
        // First check if it matches target positions
        const matchesTarget = targetPositions.some((target) =>
          pos.name.toLowerCase().includes(target.toLowerCase()) ||
          target.toLowerCase().includes(pos.name.toLowerCase())
        );
        // Then check if it's a branch-type position (exclude rotation positions)
        return matchesTarget && pos.position_type === 'branch';
      });

      // Create a map of quotas by position_id
      const quotasMap = new Map<string, PositionQuota>();
      quotasData.forEach((quota) => {
        quotasMap.set(quota.position_id, quota);
      });

      // Initialize quotas for positions that don't have them yet
      filteredPositions.forEach((position) => {
        if (!quotasMap.has(position.id)) {
          quotasMap.set(position.id, {
            position_id: position.id,
            position_name: position.name,
            designated_quota: 0, // Default to 0, user can set appropriate quota
            minimum_required: 0,  // Default to 0, user can set appropriate minimum
          });
        }
      });

      setQuotas(quotasMap);

      // #region agent log
      fetch('http://127.0.0.1:7242/ingest/0ee72595-fb2a-4cfb-9b7a-6463c0da4d1f',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({location:'BranchPositionQuotaConfig.tsx:127',message:'loadData: received constraints from API',data:{branchId,constraintsCount:constraintsData.length,constraints:constraintsData.map(c=>({day:c.day_of_week,is_overridden:c.is_overridden,reqCount:c.staff_group_requirements?.length||0}))},timestamp:Date.now(),sessionId:'debug-session',runId:'run1',hypothesisId:'F'})}).catch(()=>{});
      // #endregion

      // Create a map of constraints by day_of_week
      const constraintsMap = new Map<number, BranchConstraints>();
      constraintsData.forEach((constraint) => {
        constraintsMap.set(constraint.day_of_week, constraint);
      });

      // Initialize constraints for all days of week (0-6) if they don't exist
      for (let day = 0; day <= 6; day++) {
        if (!constraintsMap.has(day)) {
          constraintsMap.set(day, {
            branch_id: branchId,
            day_of_week: day,
            is_overridden: false,
            inherited_from_branch_type_id: branchType?.id,
            staff_group_requirements: [],
          });
        } else {
          // Ensure existing constraints have inheritance fields and staff group requirements
          const constraint = constraintsMap.get(day)!;
          if (constraint.is_overridden === undefined) {
            constraint.is_overridden = false;
          }
          if (!constraint.inherited_from_branch_type_id && branchType) {
            constraint.inherited_from_branch_type_id = branchType.id;
          }
          // Ensure staff_group_requirements exists
          if (!constraint.staff_group_requirements) {
            constraint.staff_group_requirements = [];
          }
        }
      }

      // #region agent log
      fetch('http://127.0.0.1:7242/ingest/0ee72595-fb2a-4cfb-9b7a-6463c0da4d1f',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({location:'BranchPositionQuotaConfig.tsx:159',message:'loadData: final constraints map before setting state',data:{branchId,constraints:Array.from(constraintsMap.entries()).map(([day,c])=>({day,is_overridden:c.is_overridden,reqCount:c.staff_group_requirements?.length||0}))},timestamp:Date.now(),sessionId:'debug-session',runId:'run1',hypothesisId:'F'})}).catch(()=>{});
      // #endregion

      setConstraints(constraintsMap);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load configuration');
    } finally {
      setLoading(false);
    }
  };

  const handleQuotaChange = (positionId: string, field: 'designated_quota' | 'minimum_required', value: number) => {
    const quota = quotas.get(positionId);
    if (!quota) return;

    const updatedQuota = { ...quota };
    updatedQuota[field] = value;

    // Validate: minimum_required <= designated_quota
    if (field === 'designated_quota' && updatedQuota.minimum_required > value) {
      setError(`Minimum required cannot be greater than designated quota for ${quota.position_name}`);
      return;
    }
    if (field === 'minimum_required' && value > updatedQuota.designated_quota) {
      setError(`Minimum required cannot be greater than designated quota for ${quota.position_name}`);
      return;
    }

    setError(null);
    setQuotas(new Map(quotas.set(positionId, updatedQuota)));
  };

  const handleDisplayOrderChange = async (positionId: string, displayOrder: number) => {
    const position = positions.find((p) => p.id === positionId);
    if (!position) return;

    try {
      // Update position display_order via position API
      await positionApi.update(positionId, {
        name: position.name,
        display_order: displayOrder,
        position_type: position.position_type,
        manpower_type: position.manpower_type,
      });

      // Update local positions state
      setPositions(
        positions.map((p) => (p.id === positionId ? { ...p, display_order: displayOrder } : p))
      );

      setError(null);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update display order');
    }
  };

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    setSuccess(null);

    try {
      const quotasToUpdate: PositionQuotaUpdate[] = Array.from(quotas.values()).map((quota) => ({
        position_id: quota.position_id,
        designated_quota: quota.designated_quota,
        minimum_required: quota.minimum_required,
      }));

      await branchConfigApi.updateQuotas(branchId, quotasToUpdate);
      setSuccess('Quotas updated successfully');
      if (onSave) {
        onSave();
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update quotas');
    } finally {
      setSaving(false);
    }
  };

  const handleConstraintChange = (dayOfWeek: number, staffGroupId: string, value: number) => {
    const constraint = constraints.get(dayOfWeek);
    if (!constraint) return;

    // Get or create staff group requirements array
    let staffGroupRequirements = constraint.staff_group_requirements || [];
    
    // If this constraint was inherited (not overridden), we need to ensure ALL staff groups
    // are represented in the requirements array before making changes.
    // This preserves all inherited values when converting to an override.
    if (!constraint.is_overridden && staffGroups.length > 0) {
      // Create a map of existing requirements by staff group ID
      const existingReqMap = new Map<string, number>();
      staffGroupRequirements.forEach(req => {
        existingReqMap.set(req.staff_group_id, req.minimum_count);
      });
      
      // Initialize requirements for ALL staff groups with their current values from the constraint
      // If a staff group isn't in the existing requirements, it defaults to 0
      // This ensures we preserve all inherited values when converting to override
      staffGroupRequirements = staffGroups.map(sg => {
        const existingValue = existingReqMap.get(sg.id);
        return {
          staff_group_id: sg.id,
          minimum_count: existingValue !== undefined ? existingValue : 0,
        };
      });
    }
    
    // Find existing requirement for this staff group
    const existingIndex = staffGroupRequirements.findIndex(
      req => req.staff_group_id === staffGroupId
    );

    let updatedRequirements: StaffGroupRequirement[];
    if (existingIndex >= 0) {
      // Update existing requirement
      updatedRequirements = [...staffGroupRequirements];
      updatedRequirements[existingIndex] = {
        staff_group_id: staffGroupId,
        minimum_count: value,
      };
    } else {
      // Add new requirement (shouldn't happen if we initialized all staff groups above, but handle it)
      updatedRequirements = [
        ...staffGroupRequirements,
        {
          staff_group_id: staffGroupId,
          minimum_count: value,
        },
      ];
    }

    // IMPORTANT: When converting from inherited to override, we should NOT filter out zeros yet
    // because we want to preserve the fact that some staff groups have zero requirements.
    // Only filter zeros if this was already an override (to allow users to clear values).
    // However, when saving, the backend expects only non-zero requirements for overrides.
    // So we'll keep zeros in the local state for display, but filter them when saving.
    // For now, keep all requirements (including zeros) in the local state.
    // The save function will filter them appropriately.

    const updatedConstraint = { 
      ...constraint,
      staff_group_requirements: updatedRequirements,
      // Mark as overridden when user changes it (even if some values are zero)
      is_overridden: true,
    };
    setConstraints(new Map(constraints.set(dayOfWeek, updatedConstraint)));
    setError(null);
  };

  const getConstraintValue = (dayOfWeek: number, staffGroupId: string): number => {
    const constraint = constraints.get(dayOfWeek);
    if (!constraint || !constraint.staff_group_requirements) {
      return 0;
    }
    
    const requirement = constraint.staff_group_requirements.find(
      req => req.staff_group_id === staffGroupId
    );
    
    return requirement?.minimum_count || 0;
  };

  const handleResetToDefaults = async () => {
    if (!branchType) return;
    
    if (!confirm('Are you sure you want to reset all constraints to branch type defaults? This will remove all overrides.')) {
      return;
    }

    try {
      // Get branch type constraints
      const branchTypeConstraints = await branchTypeApi.getConstraints(branchType.id);
      const constraintsMap = new Map<number, BranchConstraints>();

      // Create constraints from branch type defaults
      for (let day = 0; day < 7; day++) {
        const branchTypeConstraint = branchTypeConstraints.find(c => c.day_of_week === day);
        constraintsMap.set(day, {
          branch_id: branchId,
          day_of_week: day,
          is_overridden: false,
          inherited_from_branch_type_id: branchType.id,
          staff_group_requirements: branchTypeConstraint?.staff_group_requirements 
            ? branchTypeConstraint.staff_group_requirements.map(req => ({
                staff_group_id: req.staff_group_id,
                minimum_count: req.minimum_count,
              }))
            : [],
        });
      }

      setConstraints(constraintsMap);
      
      // Save the reset constraints (they will be marked as not overridden)
      // When is_overridden is false, the backend will delete the constraint records
      // so they inherit from branch type defaults
      const constraintsToUpdate: ConstraintsUpdate[] = Array.from(constraintsMap.values()).map((constraint) => ({
        day_of_week: constraint.day_of_week,
        staff_group_requirements: constraint.staff_group_requirements || [],
        is_overridden: false, // Signal to backend to delete and inherit from branch type
      }));

      // #region agent log
      fetch('http://127.0.0.1:7242/ingest/0ee72595-fb2a-4cfb-9b7a-6463c0da4d1f',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({location:'BranchPositionQuotaConfig.tsx:332',message:'handleResetToDefaults: sending reset request',data:{branchId,constraintsCount:constraintsToUpdate.length,constraints:constraintsToUpdate.map(c=>({day:c.day_of_week,is_overridden:c.is_overridden,reqCount:c.staff_group_requirements.length}))},timestamp:Date.now(),sessionId:'debug-session',runId:'run1',hypothesisId:'A'})}).catch(()=>{});
      // #endregion

      await branchConfigApi.updateConstraints(branchId, constraintsToUpdate);

      // #region agent log
      fetch('http://127.0.0.1:7242/ingest/0ee72595-fb2a-4cfb-9b7a-6463c0da4d1f',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({location:'BranchPositionQuotaConfig.tsx:336',message:'handleResetToDefaults: reset completed, reloading constraints',data:{branchId},timestamp:Date.now(),sessionId:'debug-session',runId:'run1',hypothesisId:'E'})}).catch(()=>{});
      // #endregion

      // Reload constraints after reset to verify they're deleted
      const reloadedConstraints = await branchConfigApi.getConstraints(branchId);
      
      // #region agent log
      fetch('http://127.0.0.1:7242/ingest/0ee72595-fb2a-4cfb-9b7a-6463c0da4d1f',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({location:'BranchPositionQuotaConfig.tsx:342',message:'handleResetToDefaults: reloaded constraints after reset',data:{branchId,constraintsCount:reloadedConstraints.length,constraints:reloadedConstraints.map(c=>({day:c.day_of_week,is_overridden:c.is_overridden,reqCount:c.staff_group_requirements?.length||0}))},timestamp:Date.now(),sessionId:'debug-session',runId:'run1',hypothesisId:'E'})}).catch(()=>{});
      // #endregion
      setSuccess('Constraints reset to branch type defaults');
    } catch (error: any) {
      setError(error.response?.data?.error || 'Failed to reset constraints');
    }
  };

  const handleSaveConstraints = async () => {
    setSavingConstraints(true);
    setError(null);
    setSuccess(null);

    try {
      // Convert constraints to update format with staff group requirements
      // IMPORTANT: Only send constraints that are actually overridden (have non-empty requirements)
      // OR constraints that should be deleted (is_overridden: false)
      // Don't send inherited constraints with empty requirements - they should inherit from branch type
      const constraintsToUpdate: ConstraintsUpdate[] = Array.from(constraints.values())
        .filter((constraint) => {
          // Include if explicitly marked as not overridden (for deletion)
          if (constraint.is_overridden === false) {
            return true;
          }
          // Include if overridden AND has non-empty staff group requirements
          const hasRequirements = constraint.staff_group_requirements && constraint.staff_group_requirements.length > 0;
          return constraint.is_overridden === true && hasRequirements;
        })
        .map((constraint) => ({
          day_of_week: constraint.day_of_week,
          // Filter out zero requirements when saving - backend expects only non-zero requirements for overrides
          // This ensures we only save the staff groups that actually have requirements
          staff_group_requirements: (constraint.staff_group_requirements || [])
            .filter(req => req.minimum_count > 0)
            .map(req => ({
              staff_group_id: req.staff_group_id,
              minimum_count: req.minimum_count,
            })),
          // Preserve the is_overridden flag: false for deletion, true for override
          // Since we've already filtered, this will be either false (for deletion) or true (for override with requirements)
          is_overridden: constraint.is_overridden,
        }));

      // #region agent log
      fetch('http://127.0.0.1:7242/ingest/0ee72595-fb2a-4cfb-9b7a-6463c0da4d1f',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({location:'BranchPositionQuotaConfig.tsx:377',message:'handleSaveConstraints: filtered constraints to send',data:{branchId,totalConstraints:constraints.size,filteredCount:constraintsToUpdate.length,constraints:constraintsToUpdate.map(c=>({day:c.day_of_week,is_overridden:c.is_overridden,reqCount:c.staff_group_requirements.length}))},timestamp:Date.now(),sessionId:'debug-session',runId:'run1',hypothesisId:'D'})}).catch(()=>{});
      // #endregion

      await branchConfigApi.updateConstraints(branchId, constraintsToUpdate);
      setSuccess('Constraints updated successfully');
      if (onSave) {
        onSave();
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update constraints');
    } finally {
      setSavingConstraints(false);
    }
  };

  const handleReset = () => {
    loadData();
    setError(null);
    setSuccess(null);
  };

  const handleBranchTypeChange = async (branchTypeId: string | null) => {
    if (!branch) return;
    
    setSavingBranchType(true);
    setError(null);
    setSuccess(null);

    try {
      // Update branch with new branch type
      const updatedBranch = await branchApi.update(branch.id, {
        name: branch.name,
        code: branch.code,
        area_manager_id: branch.area_manager_id || undefined,
        branch_type_id: branchTypeId || undefined,
        priority: branch.priority,
      });

      // Update local state
      setBranch(updatedBranch);
      
      // Load branch type if assigned
      if (updatedBranch.branch_type_id) {
        try {
          const bt = await branchTypeApi.getById(updatedBranch.branch_type_id);
          setBranchType(bt);
        } catch (err) {
          console.error('Failed to load branch type:', err);
          setBranchType(null);
        }
      } else {
        setBranchType(null);
      }

      // Reload constraints since they may have changed due to branch type change
      const constraintsData = await branchConfigApi.getConstraints(branchId);
      const constraintsMap = new Map<number, BranchConstraints>();
      constraintsData.forEach((constraint) => {
        constraintsMap.set(constraint.day_of_week, constraint);
      });
      setConstraints(constraintsMap);

      setSuccess('Branch type updated successfully');
      if (onSave) {
        onSave();
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update branch type');
    } finally {
      setSavingBranchType(false);
    }
  };

  const loadScenarioPreview = async () => {
    if (!previewDate) return;
    
    setLoadingPreview(true);
    try {
      // Get matching scenarios
      const matches = await staffRequirementScenarioApi.getMatchingScenarios(branchId, previewDate);
      setScenarioMatches(matches);

      // Get day-of-week revenue
      const dateObj = new Date(previewDate);
      const dayOfWeek = dateObj.getDay();
      const weeklyRevenue = await branchConfigApi.getWeeklyRevenue(branchId);
      const dayRevenue = weeklyRevenue.find((r) => r.day_of_week === dayOfWeek);
      if (dayRevenue) {
        // Calculate total revenue from all 4 types
        // Using multipliers: Vitamin Cases * 1000, Slim Pen Cases * 1500
        const totalRevenue = (dayRevenue.skin_revenue || 0) + 
                            (dayRevenue.ls_hm_revenue || 0) + 
                            ((dayRevenue.vitamin_cases || 0) * 1000) + 
                            ((dayRevenue.slim_pen_cases || 0) * 1500);
        // Fallback to expected_revenue for backward compatibility
        const revenueValue = totalRevenue > 0 ? totalRevenue : (dayRevenue.expected_revenue || 0);
        setDayOfWeekRevenue(revenueValue);
        // Get revenue tier
        try {
          const tier = await revenueLevelTierApi.getTierForRevenue(revenueValue);
          setRevenueTier(tier);
        } catch (err) {
          console.error('Failed to get revenue tier:', err);
        }
      }

      // Filter positions to show only target positions AND only branch-type positions
      const currentFilteredPositions = positions.filter((pos) => {
        const matchesTarget = targetPositions.some((target) =>
          pos.name.toLowerCase().includes(target.toLowerCase()) ||
          target.toLowerCase().includes(pos.name.toLowerCase())
        );
        return matchesTarget && pos.position_type === 'branch';
      });

      // Calculate requirements for each position
      const requirementsMap = new Map<string, CalculatedRequirement>();
      for (const position of currentFilteredPositions) {
        const quota = quotas.get(position.id);
        if (quota) {
          try {
            const calculated = await staffRequirementScenarioApi.calculateRequirements({
              branch_id: branchId,
              date: previewDate,
              position_id: position.id,
              base_preferred: quota.designated_quota,
              base_minimum: quota.minimum_required,
            });
            requirementsMap.set(position.id, calculated);
          } catch (err) {
            console.error(`Failed to calculate for position ${position.id}:`, err);
          }
        }
      }
      setCalculatedRequirements(requirementsMap);

      // Get doctor count (simplified - you may need to add an API endpoint for this)
      // For now, we'll set it to 0 or fetch from doctor assignments if available
      setDoctorCount(0);
    } catch (err: any) {
      console.error('Failed to load scenario preview:', err);
    } finally {
      setLoadingPreview(false);
    }
  };

  useEffect(() => {
    if (quotas.size > 0 && previewDate) {
      loadScenarioPreview();
    }
  }, [previewDate, branchId, quotas]);

  const DAY_NAMES = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

  // Filter positions to show only target positions AND only branch-type positions
  const filteredPositions = positions.filter((pos) => {
    // First check if it matches target positions
    const matchesTarget = targetPositions.some((target) =>
      pos.name.toLowerCase().includes(target.toLowerCase()) ||
      target.toLowerCase().includes(pos.name.toLowerCase())
    );
    // Then check if it's a branch-type position (exclude rotation positions)
    return matchesTarget && pos.position_type === 'branch';
  });

  // Sort by position display_order
  filteredPositions.sort((a, b) => a.display_order - b.display_order);

  if (loading) {
    return (
      <div className="flex justify-center items-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Branch Type Assignment Section */}
      <div className="bg-white p-4 rounded-lg border border-gray-200">
        <div className="flex justify-between items-center mb-4">
          <div>
            <h3 className="text-lg font-semibold">Branch Type Assignment</h3>
            <p className="text-sm text-gray-600 mt-1">
              Branch type is used as one of the 5 filter criteria (FourthCriteria) for staff allocation
            </p>
          </div>
        </div>
        <div className="flex items-center gap-4">
          <label htmlFor="branch-type-select" className="text-sm font-medium text-gray-700">
            Branch Type:
          </label>
          <select
            id="branch-type-select"
            value={branchType?.id || ''}
            onChange={(e) => handleBranchTypeChange(e.target.value || null)}
            disabled={savingBranchType}
            className="px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <option value="">None (No branch type assigned)</option>
            {branchTypes.map((bt) => (
              <option key={bt.id} value={bt.id}>
                {bt.name}
              </option>
            ))}
          </select>
          {savingBranchType && (
            <span className="text-sm text-gray-500">Saving...</span>
          )}
          {branchType && (
            <div className="text-sm text-gray-600">
              <span className="font-medium">Current:</span> {branchType.name}
              {branchType.description && (
                <span className="ml-2 text-gray-500">({branchType.description})</span>
              )}
            </div>
          )}
        </div>
      </div>

      <div className="flex justify-between items-center">
        <h3 className="text-lg font-semibold">Position Quota Configuration</h3>
        <div className="flex gap-2">
          <button
            onClick={handleReset}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
          >
            Reset
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {saving ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </div>

      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-md">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {success && (
        <div className="p-4 bg-green-50 border border-green-200 rounded-md">
          <p className="text-sm text-green-800">{success}</p>
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Display Order
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Position
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Preferred
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Minimum
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {filteredPositions.map((position) => {
              const quota = quotas.get(position.id);
              if (!quota) return null;

              return (
                <tr key={position.id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <input
                      type="number"
                      min="0"
                      value={position.display_order}
                      onChange={(e) => {
                        const value = parseInt(e.target.value) || 0;
                        handleDisplayOrderChange(position.id, value);
                      }}
                      className="w-20 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      title="Display order (lower numbers appear first)"
                    />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    {quota.position_name}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <input
                      type="number"
                      min="0"
                      value={quota.designated_quota}
                      onChange={(e) =>
                        handleQuotaChange(quota.position_id, 'designated_quota', parseInt(e.target.value) || 0)
                      }
                      className="w-24 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <input
                      type="number"
                      min="0"
                      value={quota.minimum_required}
                      onChange={(e) =>
                        handleQuotaChange(quota.position_id, 'minimum_required', parseInt(e.target.value) || 0)
                      }
                      className="w-24 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {/* Scenario Preview Section */}
      <div className="mt-8 p-6 bg-blue-50 rounded-lg border border-blue-200">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-semibold text-blue-900">Dynamic Staff Requirements Preview</h3>
          <div className="flex items-center gap-2">
            <label className="text-sm text-blue-800">Preview Date:</label>
            <input
              type="date"
              value={previewDate}
              onChange={(e) => setPreviewDate(e.target.value)}
              className="px-3 py-1 border border-blue-300 rounded-md text-sm"
            />
            <button
              onClick={loadScenarioPreview}
              disabled={loadingPreview}
              className="px-3 py-1 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {loadingPreview ? 'Loading...' : 'Refresh'}
            </button>
          </div>
        </div>

        {loadingPreview ? (
          <div className="text-center py-4">
            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600 mx-auto"></div>
          </div>
        ) : (
          <>
            {/* Current Conditions */}
            <div className="mb-4 p-4 bg-white rounded-md">
              <h4 className="font-semibold mb-2 text-blue-900">Current Conditions:</h4>
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div>
                  <span className="font-medium">Date:</span> {new Date(previewDate).toLocaleDateString()} ({DAY_NAMES[new Date(previewDate).getDay()]})
                </div>
                <div>
                  <span className="font-medium">Day-of-Week Revenue:</span>{' '}
                  {dayOfWeekRevenue !== null ? `${dayOfWeekRevenue.toLocaleString()} THB` : 'Not set'}
                </div>
                {revenueTier && (
                  <div>
                    <span className="font-medium">Revenue Level:</span> Level {revenueTier.level_number} - {revenueTier.level_name} ({revenueTier.min_revenue.toLocaleString()} - {revenueTier.max_revenue ? revenueTier.max_revenue.toLocaleString() : '∞'} THB)
                  </div>
                )}
                <div>
                  <span className="font-medium">Doctors Scheduled:</span> {doctorCount}
                </div>
              </div>
            </div>

            {/* Matching Scenarios */}
            {scenarioMatches.length > 0 && (
              <div className="mb-4 p-4 bg-white rounded-md">
                <h4 className="font-semibold mb-2 text-blue-900">Matching Scenarios:</h4>
                <div className="space-y-1 text-sm">
                  {scenarioMatches
                    .filter((m) => m.matches)
                    .sort((a, b) => b.priority - a.priority)
                    .map((match) => (
                      <div key={match.scenario_id} className="flex items-start gap-2">
                        <span className="text-green-600">✅</span>
                        <div>
                          <span className="font-medium">{match.scenario_name}</span> (Priority: {match.priority})
                          {match.match_reason && (
                            <div className="text-xs text-gray-600 ml-4">{match.match_reason}</div>
                          )}
                        </div>
                      </div>
                    ))}
                  {scenarioMatches.filter((m) => m.matches).length === 0 && (
                    <div className="text-gray-500 text-sm">No matching scenarios (using default)</div>
                  )}
                </div>
              </div>
            )}

            {/* Calculated Requirements */}
            {calculatedRequirements.size > 0 && (
              <div className="p-4 bg-white rounded-md">
                <h4 className="font-semibold mb-3 text-blue-900">Calculated Requirements:</h4>
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200 text-sm">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Position</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Base</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Calculated</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Scenario</th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {filteredPositions.map((position) => {
                        const calculated = calculatedRequirements.get(position.id);
                        const quota = quotas.get(position.id);
                        if (!calculated && !quota) return null;

                        return (
                          <tr key={position.id}>
                            <td className="px-4 py-2 font-medium">{position.name}</td>
                            <td className="px-4 py-2">
                              {quota ? (
                                <div className="text-xs">
                                  Pref: {quota.designated_quota} | Min: {quota.minimum_required}
                                </div>
                              ) : (
                                <span className="text-gray-400">-</span>
                              )}
                            </td>
                            <td className="px-4 py-2">
                              {calculated ? (
                                <div className="text-xs">
                                  <div className="font-medium text-blue-600">
                                    Pref: {calculated.calculated_preferred} | Min: {calculated.calculated_minimum}
                                  </div>
                                  {calculated.factors_applied.length > 0 && (
                                    <div className="text-xs text-gray-500 mt-1">
                                      {calculated.factors_applied.join(', ')}
                                    </div>
                                  )}
                                </div>
                              ) : (
                                <span className="text-gray-400">-</span>
                              )}
                            </td>
                            <td className="px-4 py-2 text-xs">
                              {calculated?.matched_scenario_name ? (
                                <span className="text-blue-600">{calculated.matched_scenario_name}</span>
                              ) : (
                                <span className="text-gray-400">Default</span>
                              )}
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </>
        )}
      </div>

      {/* Constraints Configuration Section */}
      <div className="mt-8">
        <div className="flex justify-between items-center mb-4">
          <div>
            <h3 className="text-lg font-semibold">Daily Staff Constraints</h3>
            {branchType && (
              <p className="text-sm text-gray-600 mt-1">
                Inherited from branch type: <span className="font-medium">{branchType.name}</span>
                {constraints.size > 0 && Array.from(constraints.values()).some(c => c.is_overridden) && (
                  <span className="ml-2 text-blue-600">(Some constraints are overridden)</span>
                )}
              </p>
            )}
          </div>
          <div className="flex gap-2">
            {branchType && Array.from(constraints.values()).some(c => c.is_overridden) && (
              <button
                onClick={handleResetToDefaults}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-200 rounded-md hover:bg-gray-300"
              >
                Reset to Defaults
              </button>
            )}
            <button
              onClick={handleSaveConstraints}
              disabled={savingConstraints}
              className="px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-md hover:bg-green-700 disabled:opacity-50"
            >
              {savingConstraints ? 'Saving...' : 'Save Constraints'}
            </button>
          </div>
        </div>

        {staffGroups.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            <p>No active staff groups found. Please create staff groups in Staff Groups settings first.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50 z-10">
                    Day
                  </th>
                  {branchType && (
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-[120px] bg-gray-50 z-10">
                      Status
                    </th>
                  )}
                  {staffGroups.map((staffGroup) => (
                    <th
                      key={staffGroup.id}
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider min-w-[120px]"
                    >
                      Min {staffGroup.name}
                      {staffGroup.description && (
                        <span className="block text-xs font-normal text-gray-400 mt-1">
                          ({staffGroup.description})
                        </span>
                      )}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {Array.from({ length: 7 }, (_, i) => i).map((dayOfWeek) => {
                  const constraint = constraints.get(dayOfWeek);
                  if (!constraint) return null;

                  const isOverridden = constraint.is_overridden || false;

                  return (
                    <tr key={dayOfWeek} className={isOverridden ? 'bg-yellow-50' : ''}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 sticky left-0 bg-white z-10">
                        {DAY_NAMES[dayOfWeek]}
                      </td>
                      {branchType && (
                        <td className="px-6 py-4 whitespace-nowrap sticky left-[120px] bg-white z-10">
                          {isOverridden ? (
                            <span className="px-2 py-1 text-xs font-semibold rounded-full bg-yellow-100 text-yellow-800">
                              Overridden
                            </span>
                          ) : (
                            <span className="px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800">
                              Inherited
                            </span>
                          )}
                        </td>
                      )}
                      {staffGroups.map((staffGroup) => (
                        <td key={staffGroup.id} className="px-6 py-4 whitespace-nowrap">
                          <input
                            type="number"
                            min="0"
                            step="1"
                            value={getConstraintValue(dayOfWeek, staffGroup.id)}
                            onChange={(e) =>
                              handleConstraintChange(dayOfWeek, staffGroup.id, parseInt(e.target.value) || 0)
                            }
                            className="w-24 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
                          />
                        </td>
                      ))}
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
