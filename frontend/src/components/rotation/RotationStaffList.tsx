'use client';

import { useState, useEffect } from 'react';
import { staffApi, Staff } from '@/lib/api/staff';
import { positionApi, Position } from '@/lib/api/position';
import { areaOfOperationApi, AreaOfOperation } from '@/lib/api/areaOfOperation';

interface RotationStaffListProps {
  onAddToAssignment?: (staff: Staff) => void;
  selectedStaffIds?: string[];
}

export default function RotationStaffList({ onAddToAssignment, selectedStaffIds = [] }: RotationStaffListProps) {
  const [rotationStaff, setRotationStaff] = useState<Staff[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [areasOfOperation, setAreasOfOperation] = useState<AreaOfOperation[]>([]);
  const [loading, setLoading] = useState(true);
  const [filterPositionId, setFilterPositionId] = useState<string>('');
  const [filterAreaOfOperationId, setFilterAreaOfOperationId] = useState<string>('');

  useEffect(() => {
    loadData();
  }, [filterPositionId, filterAreaOfOperationId]);

  const loadData = async () => {
    try {
      setLoading(true);
      const [staffData, positionsData, areasData] = await Promise.all([
        staffApi.list({
          staff_type: 'rotation',
          position_id: filterPositionId || undefined,
          area_of_operation_id: filterAreaOfOperationId || undefined,
        }),
        positionApi.list(),
        areaOfOperationApi.list(),
      ]);

      setRotationStaff(staffData || []);
      setPositions(positionsData || []);
      setAreasOfOperation(areasData || []);
    } catch (error) {
      console.error('Failed to load data:', error);
      setRotationStaff([]);
      setPositions([]);
      setAreasOfOperation([]);
    } finally {
      setLoading(false);
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
              <th>Coverage Area (Legacy)</th>
              <th>Skill Level</th>
              {onAddToAssignment && <th>Action</th>}
            </tr>
          </thead>
          <tbody>
            {rotationStaff.length === 0 ? (
              <tr>
                <td colSpan={onAddToAssignment ? 7 : 6} className="text-center py-8 text-neutral-text-secondary">
                  No rotation staff found
                </td>
              </tr>
            ) : (
              rotationStaff.map((staff) => {
                const position = positions.find((p) => p.id === staff.position_id);
                const areaOfOp = areasOfOperation.find((a) => a.id === staff.area_of_operation_id);
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
                    <td>{staff.coverage_area || '-'}</td>
                    <td>
                      <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                        {staff.skill_level || 5}/10
                      </span>
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
    </div>
  );
}

