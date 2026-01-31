'use client';

import { useState } from 'react';

interface ScoringGroup {
  id: string;
  name: string;
  priority: number;
  description: string;
  scoringFormula: string;
  examples: string[];
  color: string;
  icon: string;
}

const scoringGroups: ScoringGroup[] = [
  {
    id: 'group1',
    name: 'Position Quota - Minimum Shortage',
    priority: 1,
    description: 'Identifies critical shortages where branches are below minimum staffing requirements.',
    scoringFormula: '-1 point per staff below minimum',
    examples: [
      'Position needs 5 minimum, has 3 → -2 points',
      'Position needs 3 minimum, has 3 → 0 points',
      'Position needs 4 minimum, has 1 → -3 points',
    ],
    color: 'red',
    icon: '🔴',
  },
  {
    id: 'group2',
    name: 'Daily Staff Constraints - Minimum Shortage',
    priority: 2,
    description: 'Identifies shortages in staff group-based constraints (Daily Staff Constraints).',
    scoringFormula: '-1 point per staff group below minimum',
    examples: [
      'Staff Group "Nurses": Needs 3, has 1 → -2 points',
      'Staff Group "Managers": Needs 2, has 2 → 0 points',
      'Staff Group "Assistants": Needs 4, has 2 → -2 points',
    ],
    color: 'orange',
    icon: '🟠',
  },
  {
    id: 'group3',
    name: 'Position Quota - Preferred Excess',
    priority: 3,
    description: 'Tracks positions that are overstaffed relative to preferred quotas (informational only - shows how much above the limit for each position).',
    scoringFormula: '+1 point per staff above preferred quota',
    examples: [
      'Position needs 5 preferred, has 7 → +2 points (2 above preferred)',
      'Position needs 4 preferred, has 4 → 0 points (at preferred)',
      'Position needs 3 preferred, has 2 → 0 points (below preferred, not counted)',
    ],
    color: 'green',
    icon: '🟢',
  },
];

interface CalculatorInputs {
  positionMinRequired: number;
  positionCurrent: number;
  positionPreferred: number;
  staffGroupMinRequired: number;
  staffGroupActual: number;
}

const exampleRankings = [
  { branch: 'A01', position: 'Nurse', g1: -5, g2: -2, g3: 0, rank: 1, explanation: 'Highest G1' },
  { branch: 'B02', position: 'Nurse', g1: -3, g2: -1, g3: 1, rank: 2, explanation: 'Lower G1' },
  { branch: 'C03', position: 'Nurse', g1: -3, g2: 0, g3: 2, rank: 3, explanation: 'Same G1, lower G2' },
  { branch: 'D04', position: 'Nurse', g1: -2, g2: -3, g3: 1, rank: 4, explanation: 'Lower G1' },
];

export default function AllocationCriteriaPage() {
  const [showCalculator, setShowCalculator] = useState(false);
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>({});
  const [calculatorInputs, setCalculatorInputs] = useState<CalculatorInputs>({
    positionMinRequired: 5,
    positionCurrent: 3,
    positionPreferred: 7,
    staffGroupMinRequired: 3,
    staffGroupActual: 1,
  });

  const calculateScores = () => {
    const { positionMinRequired, positionCurrent, positionPreferred, staffGroupMinRequired, staffGroupActual } = calculatorInputs;
    
    // Group 1: Position Quota - Minimum
    const positionMinShortage = Math.max(0, positionMinRequired - positionCurrent);
    const group1Score = -1 * positionMinShortage;
    
    // Group 2: Daily Staff Constraints - Minimum
    const staffGroupShortage = Math.max(0, staffGroupMinRequired - staffGroupActual);
    const group2Score = -1 * staffGroupShortage;
    
    // Group 3: Position Quota - Preferred Excess (only positions above preferred quota)
    const preferredExcess = Math.max(0, positionCurrent - positionPreferred);
    const group3Score = preferredExcess > 0 ? preferredExcess : 0;
    
    return {
      group1: group1Score,
      group2: group2Score,
      group3: group3Score,
      positionMinShortage,
      staffGroupShortage,
      preferredExcess: preferredExcess,
    };
  };

  const scores = calculateScores();

  const getPriorityLevel = (group1Score: number, group2Score: number) => {
    if (group1Score < -2) return { level: 'Critical', color: 'bg-red-100 text-red-800 border-red-300' };
    if (group1Score < 0) return { level: 'High', color: 'bg-orange-100 text-orange-800 border-orange-300' };
    if (group2Score < 0) return { level: 'Medium', color: 'bg-yellow-100 text-yellow-800 border-yellow-300' };
    return { level: 'Low', color: 'bg-green-100 text-green-800 border-green-300' };
  };

  const priorityInfo = getPriorityLevel(scores.group1, scores.group2);

  return (
    <div className="w-full p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Allocation Scoring System</h1>
        <p className="text-gray-600">
          The allocation system uses a point-based scoring system with three fixed priority groups.
          Priorities cannot be changed as they ensure critical staffing needs are always addressed first.
        </p>
      </div>

      {/* Info Banner */}
      <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
        <div className="flex items-start">
          <div className="text-2xl mr-3">ℹ️</div>
          <div>
            <h3 className="font-semibold text-blue-900 mb-1">How the Scoring System Works</h3>
            <p className="text-sm text-blue-800">
              The system uses lexicographic ordering: Group 1 (highest priority) is checked first, then Group 2, then Group 3.
              More negative scores = higher priority (more urgent). Positive scores in Group 3 indicate less urgent needs.
            </p>
          </div>
        </div>
      </div>

      {/* Scoring Groups - Dashboard Style */}
      <div className="mb-8">
        <h2 className="text-xl font-semibold mb-4">Scoring Groups (Fixed Priority Order)</h2>
        <div className="space-y-4">
          {scoringGroups.map((group) => {
            const colorClasses = {
              red: 'border-red-300 bg-red-50',
              orange: 'border-orange-300 bg-orange-50',
              green: 'border-green-300 bg-green-50',
            };

            return (
              <div
                key={group.id}
                className={`border-2 rounded-lg p-6 ${colorClasses[group.color as keyof typeof colorClasses]}`}
              >
                <div className="flex items-start justify-between mb-4">
                  <div className="flex items-start flex-1">
                    <span className="text-4xl mr-4">{group.icon}</span>
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <h3 className="text-xl font-bold">{group.name}</h3>
                        <span className="px-2 py-1 bg-white rounded text-sm font-semibold">
                          Priority {group.priority}
                          {group.priority === 1 && ' (Highest)'}
                          {group.priority === 3 && ' (Lowest)'}
                        </span>
                      </div>
                      <p className="text-gray-700 mb-2">{group.description}</p>
                      <div className="text-sm font-medium text-gray-800 mb-3">
                        Scoring: <span className="font-mono">{group.scoringFormula}</span>
                      </div>
                      
                      <button
                        onClick={() => setExpandedGroups({ ...expandedGroups, [group.id]: !expandedGroups[group.id] })}
                        className="text-sm text-blue-600 hover:text-blue-800 font-medium"
                      >
                        {expandedGroups[group.id] ? '▼ Hide Examples' : '▶ Show Examples'}
                      </button>
                      
                      {expandedGroups[group.id] && (
                        <div className="mt-3 p-3 bg-white rounded border border-gray-200">
                          <span className="text-xs font-semibold text-gray-500 uppercase mb-2 block">Examples:</span>
                          <ul className="list-disc list-inside text-sm text-gray-700 space-y-1">
                            {group.examples.map((example, idx) => (
                              <li key={idx}>{example}</li>
                            ))}
                          </ul>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Ranking Logic */}
      <div className="mb-8">
        <h2 className="text-xl font-semibold mb-4">Ranking Logic</h2>
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-6">
          <p className="text-gray-700 mb-4">
            The system ranks branch-position combinations using strict lexicographic ordering:
          </p>
          <ol className="list-decimal list-inside space-y-2 text-gray-700 mb-4">
            <li>Sort by <strong>Group 1 Score</strong> (ascending - more negative = higher priority)</li>
            <li>If tied, sort by <strong>Group 2 Score</strong> (ascending - more negative = higher priority)</li>
            <li>If tied, sort by <strong>Group 3 Score</strong> (descending - more positive = lower priority)</li>
            <li>If still tied, sort by <strong>Branch Code</strong> (alphabetical)</li>
          </ol>
          
          {/* Example Ranking Table */}
          <div className="mt-6">
            <h3 className="font-semibold text-gray-800 mb-3">Example Ranking:</h3>
            <div className="overflow-x-auto">
              <table className="w-full border-collapse border border-gray-300 text-sm">
                <thead>
                  <tr className="bg-gray-100">
                    <th className="border border-gray-300 px-3 py-2 text-left">Rank</th>
                    <th className="border border-gray-300 px-3 py-2 text-left">Branch</th>
                    <th className="border border-gray-300 px-3 py-2 text-left">Position</th>
                    <th className="border border-gray-300 px-3 py-2 text-center">Group 1</th>
                    <th className="border border-gray-300 px-3 py-2 text-center">Group 2</th>
                    <th className="border border-gray-300 px-3 py-2 text-center">Group 3</th>
                    <th className="border border-gray-300 px-3 py-2 text-left">Explanation</th>
                  </tr>
                </thead>
                <tbody>
                  {exampleRankings.map((row) => (
                    <tr key={row.branch} className="hover:bg-gray-50">
                      <td className="border border-gray-300 px-3 py-2 font-semibold">{row.rank}</td>
                      <td className="border border-gray-300 px-3 py-2">{row.branch}</td>
                      <td className="border border-gray-300 px-3 py-2">{row.position}</td>
                      <td className={`border border-gray-300 px-3 py-2 text-center font-semibold ${row.g1 < 0 ? 'text-red-600' : 'text-gray-600'}`}>
                        {row.g1}
                      </td>
                      <td className={`border border-gray-300 px-3 py-2 text-center font-semibold ${row.g2 < 0 ? 'text-orange-600' : 'text-gray-600'}`}>
                        {row.g2}
                      </td>
                      <td className={`border border-gray-300 px-3 py-2 text-center font-semibold ${row.g3 > 0 ? 'text-blue-600' : 'text-gray-600'}`}>
                        {row.g3 > 0 ? `+${row.g3}` : row.g3}
                      </td>
                      <td className="border border-gray-300 px-3 py-2 text-gray-600">{row.explanation}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <p className="text-xs text-gray-500 mt-2">Key: G1 = Group 1, G2 = Group 2, G3 = Group 3</p>
          </div>
        </div>
      </div>

      {/* Interactive Calculator - Expandable */}
      <div className="mb-8">
        <button
          onClick={() => setShowCalculator(!showCalculator)}
          className="w-full flex items-center justify-between p-4 bg-blue-50 border border-blue-200 rounded-lg hover:bg-blue-100 transition-colors"
        >
          <div className="flex items-center gap-3">
            <span className="text-2xl">🧮</span>
            <div className="text-left">
              <h2 className="text-xl font-semibold text-blue-900">Try It: Interactive Scoring Calculator</h2>
              <p className="text-sm text-blue-700">Experiment with different scenarios to see how they score</p>
            </div>
          </div>
          <span className="text-2xl text-blue-600">{showCalculator ? '▼' : '▶'}</span>
        </button>

        {showCalculator && (
          <div className="mt-4 border-2 border-blue-200 rounded-lg p-6 bg-white">
            <div className="mb-6">
              <h3 className="text-lg font-semibold mb-4">Scenario: Branch A - Nurse Position</h3>
              
              {/* Input Fields */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
                <div>
                  <h4 className="font-semibold text-gray-700 mb-3">Position Quota Configuration</h4>
                  <div className="space-y-3">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Minimum Required
                      </label>
                      <input
                        type="number"
                        value={calculatorInputs.positionMinRequired}
                        onChange={(e) => setCalculatorInputs({ ...calculatorInputs, positionMinRequired: parseInt(e.target.value) || 0 })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        min="0"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Current Count
                      </label>
                      <input
                        type="number"
                        value={calculatorInputs.positionCurrent}
                        onChange={(e) => setCalculatorInputs({ ...calculatorInputs, positionCurrent: parseInt(e.target.value) || 0 })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        min="0"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Preferred Quota
                      </label>
                      <input
                        type="number"
                        value={calculatorInputs.positionPreferred}
                        onChange={(e) => setCalculatorInputs({ ...calculatorInputs, positionPreferred: parseInt(e.target.value) || 0 })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        min="0"
                      />
                    </div>
                  </div>
                </div>

                <div>
                  <h4 className="font-semibold text-gray-700 mb-3">Daily Staff Constraints</h4>
                  <div className="space-y-3">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Staff Group "Nurses" - Minimum Required
                      </label>
                      <input
                        type="number"
                        value={calculatorInputs.staffGroupMinRequired}
                        onChange={(e) => setCalculatorInputs({ ...calculatorInputs, staffGroupMinRequired: parseInt(e.target.value) || 0 })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        min="0"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Staff Group "Nurses" - Actual Count
                      </label>
                      <input
                        type="number"
                        value={calculatorInputs.staffGroupActual}
                        onChange={(e) => setCalculatorInputs({ ...calculatorInputs, staffGroupActual: parseInt(e.target.value) || 0 })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        min="0"
                      />
                    </div>
                  </div>
                </div>
              </div>

              {/* Score Calculation Display */}
              <div className="border-t border-gray-200 pt-6">
                <h4 className="font-semibold text-gray-700 mb-4">SCORES:</h4>
                
                <div className="space-y-4 mb-6">
                  {/* Group 1 */}
                  <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-semibold text-red-900">Group 1 (Position Quota - Min):</span>
                      <span className={`text-lg font-bold ${scores.group1 < 0 ? 'text-red-600' : 'text-gray-600'}`}>
                        {scores.group1} points
                      </span>
                    </div>
                    <div className="text-sm text-gray-700">
                      Position "Nurse": {scores.positionMinShortage} below minimum → {scores.group1} points
                    </div>
                  </div>

                  {/* Group 2 */}
                  <div className="p-4 bg-orange-50 border border-orange-200 rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-semibold text-orange-900">Group 2 (Daily Constraints - Min):</span>
                      <span className={`text-lg font-bold ${scores.group2 < 0 ? 'text-orange-600' : 'text-gray-600'}`}>
                        {scores.group2} points
                      </span>
                    </div>
                    <div className="text-sm text-gray-700">
                      Staff Group "Nurses": {scores.staffGroupShortage} below minimum → {scores.group2} points
                    </div>
                  </div>

                  {/* Group 3 */}
                  <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-semibold text-green-900">Group 3 (Position Quota - Preferred):</span>
                      <span className={`text-lg font-bold ${scores.group3 > 0 ? 'text-blue-600' : 'text-gray-600'}`}>
                        {scores.group3 > 0 ? `+${scores.group3}` : scores.group3} points
                      </span>
                    </div>
                    <div className="text-sm text-gray-700">
                      {scores.group3 > 0 
                        ? `Position "Nurse": ${scores.preferredExcess} above preferred quota → +${scores.group3} points (informational only)`
                        : `Position "Nurse": At or below preferred quota → 0 points`
                      }
                    </div>
                  </div>
                </div>

                {/* Final Score Summary */}
                <div className="border-t border-gray-300 pt-4">
                  <div className="flex items-center justify-between mb-3">
                    <span className="font-bold text-lg text-gray-900">FINAL SCORE:</span>
                    <div className="flex items-center gap-4">
                      <span className="text-sm text-gray-600">
                        Group 1: <span className={`font-semibold ${scores.group1 < 0 ? 'text-red-600' : 'text-gray-600'}`}>{scores.group1}</span>
                      </span>
                      <span className="text-sm text-gray-600">
                        Group 2: <span className={`font-semibold ${scores.group2 < 0 ? 'text-orange-600' : 'text-gray-600'}`}>{scores.group2}</span>
                      </span>
                      <span className="text-sm text-gray-600">
                        Group 3: <span className={`font-semibold ${scores.group3 > 0 ? 'text-blue-600' : 'text-gray-600'}`}>
                          {scores.group3 > 0 ? `+${scores.group3}` : scores.group3}
                        </span>
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="font-medium text-gray-700">Ranking Priority:</span>
                    <span className={`px-3 py-1 rounded-md text-sm font-semibold border ${priorityInfo.color}`}>
                      {priorityInfo.level}
                    </span>
                    <span className="text-sm text-gray-600">
                      (Group 1 = {scores.group1}, Group 2 = {scores.group2})
                    </span>
                  </div>
                </div>
              </div>

              <div className="mt-6 flex gap-3">
                <button
                  onClick={() => setCalculatorInputs({
                    positionMinRequired: 5,
                    positionCurrent: 3,
                    positionPreferred: 7,
                    staffGroupMinRequired: 3,
                    staffGroupActual: 1,
                  })}
                  className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 transition-colors"
                >
                  Reset Calculator
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Key Features */}
      <div className="mb-8">
        <h2 className="text-xl font-semibold mb-4">Key Features</h2>
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-6">
          <ul className="space-y-2 text-gray-700">
            <li className="flex items-start">
              <span className="text-green-600 mr-2">✓</span>
              <span><strong>Magnitude matters:</strong> 2 staff below minimum = -2 points (not -1)</span>
            </li>
            <li className="flex items-start">
              <span className="text-green-600 mr-2">✓</span>
              <span><strong>Separate display:</strong> All three groups shown independently</span>
            </li>
            <li className="flex items-start">
              <span className="text-green-600 mr-2">✓</span>
              <span><strong>Fixed priorities:</strong> Cannot be changed (ensures critical needs are always addressed first)</span>
            </li>
            <li className="flex items-start">
              <span className="text-green-600 mr-2">✓</span>
              <span><strong>Deterministic:</strong> Always produces consistent ranking</span>
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
}
