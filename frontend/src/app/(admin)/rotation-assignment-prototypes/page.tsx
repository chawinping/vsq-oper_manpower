'use client';

import { useState } from 'react';
import BranchRotationTable from '@/components/rotation/BranchRotationTable';
import BranchScheduleWithPanel from '@/components/rotation/BranchScheduleWithPanel';

type PrototypeView = 'table' | 'panel';

export default function RotationAssignmentPrototypesPage() {
  const [activeView, setActiveView] = useState<PrototypeView>('table');

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-neutral-text-primary mb-2">
          Rotation Staff Assignment - UI Prototypes
        </h1>
        <p className="text-sm text-neutral-text-secondary mb-4">
          Two alternative UI approaches for assigning rotation staff to branches. Both use placeholder data.
        </p>

        {/* View Toggle */}
        <div className="flex gap-2 mb-6">
          <button
            onClick={() => setActiveView('table')}
            className={`px-4 py-2 rounded-md transition-colors ${
              activeView === 'table'
                ? 'bg-salesforce-blue text-white'
                : 'bg-neutral-bg-secondary text-neutral-text-primary hover:bg-neutral-hover'
            }`}
          >
            Alternative 1: Branch-Centric Table View
          </button>
          <button
            onClick={() => setActiveView('panel')}
            className={`px-4 py-2 rounded-md transition-colors ${
              activeView === 'panel'
                ? 'bg-salesforce-blue text-white'
                : 'bg-neutral-bg-secondary text-neutral-text-primary hover:bg-neutral-hover'
            }`}
          >
            Alternative 2: Schedule View with Panel
          </button>
        </div>

        {/* Info Box */}
        <div className="mb-4 p-4 bg-yellow-50 border border-yellow-200 rounded-md">
          <p className="text-sm text-yellow-800">
            <strong>Note:</strong> These are prototypes with placeholder data. 
            Click on date cells for rotation staff to see assignment functionality.
          </p>
        </div>
      </div>

      {/* Render Active View */}
      {activeView === 'table' ? (
        <div>
          <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-md">
            <h3 className="font-semibold text-blue-900 mb-1">Alternative 1: Branch-Centric Table View</h3>
            <p className="text-sm text-blue-800">
              All eligible rotation staff are automatically shown as rows in the table. 
              Click date cells directly to assign/unassign rotation staff.
            </p>
          </div>
          <BranchRotationTable />
        </div>
      ) : (
        <div>
          <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-md">
            <h3 className="font-semibold text-blue-900 mb-1">Alternative 2: Schedule View with Panel</h3>
            <p className="text-sm text-blue-800">
              Rotation staff are shown in a side panel. Click "Add to Schedule" to add them to the schedule view, 
              then click date cells to assign dates.
            </p>
          </div>
          <BranchScheduleWithPanel />
        </div>
      )}

      {/* Comparison Notes */}
      <div className="mt-8 p-4 bg-neutral-bg-secondary border border-neutral-border rounded-md">
        <h3 className="font-semibold mb-2">Comparison Notes</h3>
        <div className="grid md:grid-cols-2 gap-4 text-sm">
          <div>
            <h4 className="font-semibold mb-1">Alternative 1 (Table View)</h4>
            <ul className="list-disc list-inside space-y-1 text-neutral-text-secondary">
              <li>All staff visible at once</li>
              <li>Direct click-to-assign</li>
              <li>Better for bulk operations</li>
              <li>Requires horizontal scroll for many dates</li>
            </ul>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Alternative 2 (Panel View)</h4>
            <ul className="list-disc list-inside space-y-1 text-neutral-text-secondary">
              <li>Two-step process (add then assign)</li>
              <li>Panel can be filtered/searched</li>
              <li>Less cluttered schedule view</li>
              <li>More flexible for many rotation staff</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}


