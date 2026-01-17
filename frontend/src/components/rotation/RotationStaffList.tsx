'use client';

import { useState, useEffect } from 'react';
import { staffApi, Staff, CreateStaffRequest } from '@/lib/api/staff';
import { positionApi, Position } from '@/lib/api/position';
import { areaOfOperationApi, AreaOfOperation } from '@/lib/api/areaOfOperation';
import { zoneApi, Zone } from '@/lib/api/zone';
import { branchApi, Branch } from '@/lib/api/branch';
import { effectiveBranchApi } from '@/lib/api/effectiveBranch';

interface RotationStaffListProps {
  onAddToAssignment?: (staff: Staff) => void;
  selectedStaffIds?: string[];
}

export default function RotationStaffList({ onAddToAssignment, selectedStaffIds = [] }: RotationStaffListProps) {
  const [rotationStaff, setRotationStaff] = useState<Staff[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [areasOfOperation, setAreasOfOperation] = useState<AreaOfOperation[]>([]);
  const [zones, setZones] = useState<Zone[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [loading, setLoading] = useState(true);
  const [filterPositionId, setFilterPositionId] = useState<string>('');
  const [filterAreaOfOperationId, setFilterAreaOfOperationId] = useState<string>('');
  const [showEditModal, setShowEditModal] = useState(false);
  const [showParametersModal, setShowParametersModal] = useState(false);
  const [editingStaff, setEditingStaff] = useState<Staff | null>(null);
  const [formData, setFormData] = useState<CreateStaffRequest>({
    zone_id: '',
    branch_ids: [],
  });
  const [selectedBranches, setSelectedBranches] = useState<string[]>([]);
  const [zoneBranchIds, setZoneBranchIds] = useState<string[]>([]);
  const [branchParameters, setBranchParameters] = useState<Record<string, {
    commute_duration_minutes: number;
    transit_count: number;
    travel_cost: number;
  }>>({});
  const [loadingStaffBranches, setLoadingStaffBranches] = useState(false);

  useEffect(() => {
    loadData();
  }, [filterPositionId, filterAreaOfOperationId]);

  const loadData = async () => {
    try {
      setLoading(true);
      const [staffData, positionsData, areasData, zonesData, branchesData] = await Promise.all([
        staffApi.list({
          staff_type: 'rotation',
          position_id: filterPositionId || undefined,
          area_of_operation_id: filterAreaOfOperationId || undefined,
        }),
        positionApi.list(),
        areaOfOperationApi.list(),
        zoneApi.list(true), // Include inactive zones
        branchApi.list(),
      ]);

      setRotationStaff(staffData || []);
      setPositions(positionsData || []);
      setAreasOfOperation(areasData || []);
      setZones(zonesData || []);
      setBranches(branchesData || []);
    } catch (error) {
      console.error('Failed to load data:', error);
      setRotationStaff([]);
      setPositions([]);
      setAreasOfOperation([]);
      setZones([]);
      setBranches([]);
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = async (staffMember: Staff) => {
    setEditingStaff(staffMember);
    setFormData({
      zone_id: (staffMember as any).zone_id || '',
      branch_ids: [],
    });
    
    // Load branches if this is rotation staff
    setLoadingStaffBranches(true);
    try {
      const staffBranches = (staffMember as any).branches || [];
      setSelectedBranches(staffBranches.map((b: any) => b.id));
      
      // Load zone branches if zone is set
      if ((staffMember as any).zone_id) {
        try {
          const zoneBranchList = await zoneApi.getBranches((staffMember as any).zone_id);
          setZoneBranchIds(zoneBranchList.map(b => b.id));
        } catch (error) {
          console.error('Failed to load zone branches:', error);
          setZoneBranchIds([]);
        }
      } else {
        setZoneBranchIds([]);
      }
      
      // Load effective branches to get parameters
      const ebs = await effectiveBranchApi.getByRotationStaffID(staffMember.id);
      const params: Record<string, {
        commute_duration_minutes: number;
        transit_count: number;
        travel_cost: number;
      }> = {};
      
      ebs.forEach(eb => {
        params[eb.branch_id] = {
          commute_duration_minutes: eb.commute_duration_minutes ?? 300,
          transit_count: eb.transit_count ?? 10,
          travel_cost: eb.travel_cost ?? 1000,
        };
      });
      setBranchParameters(params);
    } catch (error) {
      console.error('Failed to load staff branches:', error);
      setSelectedBranches([]);
      setBranchParameters({});
    } finally {
      setLoadingStaffBranches(false);
    }
    
    setShowEditModal(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingStaff) return;

    try {
      const data: CreateStaffRequest = {
        zone_id: formData.zone_id && formData.zone_id !== '' ? formData.zone_id : undefined,
        branch_ids: selectedBranches.length > 0 ? selectedBranches : undefined,
      };

      await staffApi.update(editingStaff.id, data);

      // Save effective branches for zone + additional branches (with default parameters)
      // This determines which branches the rotation staff can work at
      const zoneBranches: string[] = [];
      if (formData.zone_id) {
        try {
          // Fetch zone branches
          const zoneBranchList = await zoneApi.getBranches(formData.zone_id);
          zoneBranches.push(...zoneBranchList.map(b => b.id));
        } catch (error) {
          console.error('Failed to load zone branches:', error);
        }
      }
      
      // All branches = zone branches + additional branches (deduplicated)
      const allBranchIds = [...new Set([...zoneBranches, ...selectedBranches])];
      
      // Create effective branches with default parameters
      // Use existing parameters if available, otherwise use defaults
      const effectiveBranches = allBranchIds.map(branchId => ({
        branch_id: branchId,
        level: 1, // Default level
        commute_duration_minutes: branchParameters[branchId]?.commute_duration_minutes ?? 300,
        transit_count: branchParameters[branchId]?.transit_count ?? 10,
        travel_cost: branchParameters[branchId]?.travel_cost ?? 1000,
      }));

      try {
        await effectiveBranchApi.bulkUpdate({
          rotation_staff_id: editingStaff.id,
          effective_branches: effectiveBranches,
        });
      } catch (error: any) {
        console.error('Failed to save effective branches:', error);
        alert('Staff updated but failed to save branch assignments: ' + (error.response?.data?.error || error.message));
      }

      setShowEditModal(false);
      setEditingStaff(null);
      setFormData({ zone_id: '', branch_ids: [] });
      setSelectedBranches([]);
      setZoneBranchIds([]);
      setBranchParameters({});
      await loadData();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to update staff');
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-neutral-text-secondary">Loading...</div>
      </div>
    );
  }

  return (
    <div className="card">
      <div className="p-3 border-b border-neutral-border">
        <h2 className="text-lg font-semibold text-neutral-text-primary mb-2">
          Rotation Staff
        </h2>
        
        {/* Filters */}
        <div className="flex gap-3 flex-wrap">
          <div className="flex-1 min-w-[200px]">
            <label htmlFor="filter-position" className="block text-xs font-medium text-neutral-text-primary mb-1">
              Filter by Position
            </label>
            <select
              id="filter-position"
              value={filterPositionId}
              onChange={(e) => setFilterPositionId(e.target.value)}
              className="input-field w-full text-sm"
            >
              <option value="">All Positions</option>
              {positions.map((position) => (
                <option key={position.id} value={position.id}>
                  {position.name}
                </option>
              ))}
            </select>
          </div>
          
          <div className="flex-1 min-w-[200px]">
            <label htmlFor="filter-area" className="block text-xs font-medium text-neutral-text-primary mb-1">
              Filter by Area of Operation
            </label>
            <select
              id="filter-area"
              value={filterAreaOfOperationId}
              onChange={(e) => setFilterAreaOfOperationId(e.target.value)}
              className="input-field w-full text-sm"
            >
              <option value="">All Areas</option>
              {areasOfOperation.map((area) => (
                <option key={area.id} value={area.id}>
                  {area.name} ({area.code})
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="table-salesforce">
          <thead>
            <tr>
              <th>Nickname</th>
              <th>Name</th>
              <th>Position</th>
              <th>Area of Operation</th>
              <th>Zone / Branches</th>
              <th>Coverage Area (Legacy)</th>
              <th>Skill Level</th>
              <th>Actions</th>
              {onAddToAssignment && <th>Assignment</th>}
            </tr>
          </thead>
          <tbody>
            {rotationStaff.length === 0 ? (
              <tr>
                <td colSpan={onAddToAssignment ? 9 : 8} className="text-center py-8 text-neutral-text-secondary">
                  No rotation staff found
                </td>
              </tr>
            ) : (
              rotationStaff.map((staff) => {
                const position = positions.find((p) => p.id === staff.position_id);
                const areaOfOp = areasOfOperation.find((a) => a.id === staff.area_of_operation_id);
                const zone = zones.find((z) => z.id === (staff as any).zone_id);
                const staffBranches = (staff as any).branches || [];
                const isSelected = selectedStaffIds.includes(staff.id);
                
                return (
                  <tr key={staff.id}>
                    <td className="font-medium">{staff.nickname || '-'}</td>
                    <td className="font-medium">{staff.name}</td>
                    <td>{position?.name || '-'}</td>
                    <td>
                      {areaOfOp ? (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                          {areaOfOp.name} ({areaOfOp.code})
                        </span>
                      ) : (
                        '-'
                      )}
                    </td>
                    <td>
                      {zone ? (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                          Zone: {zone.name} ({zone.code})
                        </span>
                      ) : staffBranches.length > 0 ? (
                        <span className="text-xs text-neutral-text-secondary">
                          {staffBranches.length} branch(es)
                        </span>
                      ) : (
                        '-'
                      )}
                    </td>
                    <td>{staff.coverage_area || '-'}</td>
                    <td>
                      <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                        {staff.skill_level || 5}/10
                      </span>
                    </td>
                    <td>
                      <div className="flex gap-2">
                        <button
                          onClick={() => handleEdit(staff)}
                          className="text-salesforce-blue hover:text-salesforce-blue-hover text-sm"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => {
                            setEditingStaff(staff);
                            // Load existing parameters
                            effectiveBranchApi.getByRotationStaffID(staff.id)
                              .then(ebs => {
                                const params: Record<string, {
                                  commute_duration_minutes: number;
                                  transit_count: number;
                                  travel_cost: number;
                                }> = {};
                                ebs.forEach(eb => {
                                  params[eb.branch_id] = {
                                    commute_duration_minutes: eb.commute_duration_minutes ?? 300,
                                    transit_count: eb.transit_count ?? 10,
                                    travel_cost: eb.travel_cost ?? 1000,
                                  };
                                });
                                setBranchParameters(params);
                                setShowParametersModal(true);
                              })
                              .catch(error => {
                                console.error('Failed to load parameters:', error);
                                setBranchParameters({});
                                setShowParametersModal(true);
                              });
                          }}
                          className="text-green-600 hover:text-green-700 text-sm"
                        >
                          Parameters
                        </button>
                      </div>
                    </td>
                    {onAddToAssignment && (
                      <td>
                        {isSelected ? (
                          <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                            Added
                          </span>
                        ) : (
                          <button
                            onClick={() => onAddToAssignment(staff)}
                            className="btn-primary text-xs px-3 py-1"
                          >
                            Add to Assignment
                          </button>
                        )}
                      </td>
                    )}
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {/* Edit Zone and Branches Modal */}
      {showEditModal && editingStaff && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4 overflow-y-auto">
          <div className="card max-w-2xl w-full my-8 max-h-[90vh] flex flex-col">
            <div className="p-6 flex-shrink-0 border-b border-neutral-border">
              <h2 className="text-xl font-semibold text-neutral-text-primary">
                Edit Zone and Branches - {editingStaff.name}
              </h2>
            </div>
            <form onSubmit={handleSubmit} className="flex flex-col flex-1 min-h-0">
              <div className="p-6 overflow-y-auto flex-1">
                <div className="space-y-6">
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Zone
                    </label>
                    <select
                      value={formData.zone_id || ''}
                      onChange={async (e) => {
                        const newZoneId = e.target.value || undefined;
                        setFormData({ ...formData, zone_id: newZoneId || '' });
                        
                        // Load zone branches when zone changes
                        if (newZoneId) {
                          try {
                            const zoneBranchList = await zoneApi.getBranches(newZoneId);
                            const zoneIds = zoneBranchList.map(b => b.id);
                            setZoneBranchIds(zoneIds);
                            
                            // Initialize default parameters for zone branches if not already set
                            const newParams = { ...branchParameters };
                            zoneIds.forEach(branchId => {
                              if (!newParams[branchId]) {
                                newParams[branchId] = {
                                  commute_duration_minutes: 300,
                                  transit_count: 10,
                                  travel_cost: 1000,
                                };
                              }
                            });
                            setBranchParameters(newParams);
                          } catch (error) {
                            console.error('Failed to load zone branches:', error);
                            setZoneBranchIds([]);
                          }
                        } else {
                          setZoneBranchIds([]);
                        }
                      }}
                      className="input-field"
                    >
                      <option value="">Select Zone (Optional)</option>
                      {(zones || []).map((zone) => (
                        <option key={zone.id} value={zone.id}>
                          {zone.name} ({zone.code})
                        </option>
                      ))}
                    </select>
                    <p className="mt-1 text-xs text-neutral-text-secondary">
                      Select a zone for this rotation staff member. All branches in the zone will be included.
                    </p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-text-primary mb-1.5">
                      Additional Branches (Outside Zone)
                    </label>
                    <p className="text-xs text-neutral-text-secondary mb-2">
                      {formData.zone_id 
                        ? 'Branches from the selected zone are automatically included. Select additional branches outside the zone for flexibility. Use the "Parameters" button to set travel parameters for all branches.'
                        : 'Select branches this rotation staff can work at. Use the "Parameters" button to set travel parameters for all branches.'}
                    </p>
                    {loadingStaffBranches ? (
                      <div className="text-sm text-neutral-text-secondary">Loading...</div>
                    ) : (
                      <div className="space-y-3 max-h-96 overflow-y-auto border border-neutral-border rounded-md p-3">
                        {branches.map((branch) => {
                          const isZoneBranch = zoneBranchIds.includes(branch.id);
                          const isAdditionalBranch = selectedBranches.includes(branch.id);
                          return (
                            <div key={branch.id} className="flex items-center gap-3">
                              {isZoneBranch ? (
                                <span className="w-4 h-4 flex items-center justify-center text-xs text-purple-600" title="Zone branch (automatically included)">
                                  ✓
                                </span>
                              ) : (
                                <input
                                  type="checkbox"
                                  id={`staff-branch-${branch.id}`}
                                  checked={isAdditionalBranch}
                                  onChange={(e) => {
                                    if (e.target.checked) {
                                      setSelectedBranches([...selectedBranches, branch.id]);
                                    } else {
                                      setSelectedBranches(selectedBranches.filter(id => id !== branch.id));
                                    }
                                  }}
                                  className="w-4 h-4"
                                />
                              )}
                              <label htmlFor={`staff-branch-${branch.id}`} className="flex-1 cursor-pointer">
                                <span className="font-medium">{branch.name}</span>
                                <span className="text-xs text-neutral-text-secondary ml-2">({branch.code})</span>
                                {isZoneBranch && (
                                  <span className="ml-2 text-xs text-purple-600">(Zone)</span>
                                )}
                              </label>
                            </div>
                          );
                        })}
                        {zoneBranchIds.length === 0 && selectedBranches.length === 0 && (
                          <div className="text-sm text-neutral-text-secondary text-center py-2">
                            {formData.zone_id 
                              ? 'Select a zone to see its branches, or select additional branches below'
                              : 'No branches selected. Select a zone or individual branches.'}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              </div>
              <div className="p-6 flex-shrink-0 border-t border-neutral-border flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => {
                    setShowEditModal(false);
                    setEditingStaff(null);
                    setSelectedBranches([]);
                    setZoneBranchIds([]);
                    setBranchParameters({});
                  }}
                  className="btn-secondary"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="btn-primary"
                >
                  Update
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Travel Parameters Modal */}
      {showParametersModal && editingStaff && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4 overflow-y-auto">
          <div className="card max-w-4xl w-full my-8 max-h-[90vh] flex flex-col">
            <div className="p-6 flex-shrink-0 border-b border-neutral-border">
              <h2 className="text-xl font-semibold text-neutral-text-primary">
                Travel Parameters - {editingStaff.name}
              </h2>
              <p className="text-sm text-neutral-text-secondary mt-1">
                Set travel parameters (duration, transits, cost) for all branches. These parameters are used for allocation calculations.
              </p>
            </div>
            <div className="p-6 overflow-y-auto flex-1">
              <div className="space-y-3">
                {branches.map((branch) => {
                  const params = branchParameters[branch.id] || {
                    commute_duration_minutes: 300,
                    transit_count: 10,
                    travel_cost: 1000,
                  };
                  return (
                    <div key={branch.id} className="border border-neutral-border rounded-md p-4 bg-neutral-surface-secondary">
                      <div className="flex items-center justify-between mb-3">
                        <div>
                          <span className="font-medium text-neutral-text-primary">{branch.name}</span>
                          <span className="text-xs text-neutral-text-secondary ml-2">({branch.code})</span>
                        </div>
                      </div>
                      <div className="grid grid-cols-3 gap-4">
                        <div>
                          <label className="block text-xs font-medium text-neutral-text-secondary mb-1.5">
                            Duration (minutes)
                          </label>
                          <input
                            type="number"
                            min="0"
                            value={params.commute_duration_minutes}
                            onChange={(e) => {
                              setBranchParameters({
                                ...branchParameters,
                                [branch.id]: {
                                  ...params,
                                  commute_duration_minutes: parseInt(e.target.value) || 300,
                                },
                              });
                            }}
                            className="w-full border border-neutral-border rounded px-3 py-2 text-sm"
                            placeholder="300"
                          />
                        </div>
                        <div>
                          <label className="block text-xs font-medium text-neutral-text-secondary mb-1.5">
                            Transits
                          </label>
                          <input
                            type="number"
                            min="0"
                            value={params.transit_count}
                            onChange={(e) => {
                              setBranchParameters({
                                ...branchParameters,
                                [branch.id]: {
                                  ...params,
                                  transit_count: parseInt(e.target.value) || 10,
                                },
                              });
                            }}
                            className="w-full border border-neutral-border rounded px-3 py-2 text-sm"
                            placeholder="10"
                          />
                        </div>
                        <div>
                          <label className="block text-xs font-medium text-neutral-text-secondary mb-1.5">
                            Cost
                          </label>
                          <input
                            type="number"
                            min="0"
                            step="0.01"
                            value={params.travel_cost}
                            onChange={(e) => {
                              setBranchParameters({
                                ...branchParameters,
                                [branch.id]: {
                                  ...params,
                                  travel_cost: parseFloat(e.target.value) || 1000,
                                },
                              });
                            }}
                            className="w-full border border-neutral-border rounded px-3 py-2 text-sm"
                            placeholder="1000"
                          />
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
            <div className="p-6 flex-shrink-0 border-t border-neutral-border flex justify-end gap-2">
              <button
                type="button"
                onClick={() => {
                  setShowParametersModal(false);
                  setEditingStaff(null);
                  setBranchParameters({});
                }}
                className="btn-secondary"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={async () => {
                  if (!editingStaff) return;
                  
                  try {
                    // Create effective branches for all branches with their parameters
                    const effectiveBranches = branches.map(branch => {
                      const params = branchParameters[branch.id] || {
                        commute_duration_minutes: 300,
                        transit_count: 10,
                        travel_cost: 1000,
                      };
                      return {
                        branch_id: branch.id,
                        level: 1, // Default level
                        commute_duration_minutes: params.commute_duration_minutes,
                        transit_count: params.transit_count,
                        travel_cost: params.travel_cost,
                      };
                    });

                    await effectiveBranchApi.bulkUpdate({
                      rotation_staff_id: editingStaff.id,
                      effective_branches: effectiveBranches,
                    });

                    setShowParametersModal(false);
                    setEditingStaff(null);
                    setBranchParameters({});
                    await loadData();
                  } catch (error: any) {
                    alert(error.response?.data?.error || 'Failed to save travel parameters');
                  }
                }}
                className="btn-primary"
              >
                Save Parameters
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

