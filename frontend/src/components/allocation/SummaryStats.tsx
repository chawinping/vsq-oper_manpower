'use client';

interface SummaryStatsProps {
  stats: {
    total: number;
    needsAttention: number;
    critical: number;
    ok: number;
  };
}

export default function SummaryStats({ stats }: SummaryStatsProps) {
  return (
    <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
      <div className="flex flex-wrap gap-6">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-gray-700">📊 Summary:</span>
          <span className="text-sm text-gray-600">{stats.total} branches</span>
        </div>
        {stats.needsAttention > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-yellow-700">⚠️</span>
            <span className="text-sm text-gray-600">{stats.needsAttention} need attention</span>
          </div>
        )}
        {stats.critical > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-red-700">🚨</span>
            <span className="text-sm text-gray-600">{stats.critical} critical</span>
          </div>
        )}
        {stats.ok > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-green-700">✓</span>
            <span className="text-sm text-gray-600">{stats.ok} OK</span>
          </div>
        )}
      </div>
    </div>
  );
}
