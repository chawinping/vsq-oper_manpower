'use client';

interface FilterBarProps {
  filter: {
    status: 'all' | 'needs_attention' | 'critical' | 'ok';
    priority: 'all' | 'high' | 'medium' | 'low';
    search: string;
  };
  onFilterChange: (filter: {
    status: 'all' | 'needs_attention' | 'critical' | 'ok';
    priority: 'all' | 'high' | 'medium' | 'low';
    search: string;
  }) => void;
}

export default function FilterBar({ filter, onFilterChange }: FilterBarProps) {
  const handleStatusChange = (status: typeof filter.status) => {
    onFilterChange({ ...filter, status });
  };

  const handlePriorityChange = (priority: typeof filter.priority) => {
    onFilterChange({ ...filter, priority });
  };

  const handleSearchChange = (search: string) => {
    onFilterChange({ ...filter, search });
  };

  return (
    <div className="bg-white border border-gray-200 rounded-lg p-4 mb-4">
      <div className="flex flex-wrap gap-4 items-center">
        {/* Status Filter */}
        <div className="flex items-center gap-2">
          <label className="text-sm font-medium text-gray-700">Status:</label>
          <select
            value={filter.status}
            onChange={(e) => handleStatusChange(e.target.value as typeof filter.status)}
            className="px-3 py-2 border border-gray-300 rounded-md text-sm"
          >
            <option value="all">All</option>
            <option value="needs_attention">Needs Attention</option>
            <option value="critical">Critical</option>
            <option value="ok">OK</option>
          </select>
        </div>

        {/* Priority Filter */}
        <div className="flex items-center gap-2">
          <label className="text-sm font-medium text-gray-700">Priority:</label>
          <select
            value={filter.priority}
            onChange={(e) => handlePriorityChange(e.target.value as typeof filter.priority)}
            className="px-3 py-2 border border-gray-300 rounded-md text-sm"
          >
            <option value="all">All</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>
        </div>

        {/* Search */}
        <div className="flex items-center gap-2 flex-1 min-w-[200px]">
          <label className="text-sm font-medium text-gray-700">Search:</label>
          <input
            type="text"
            value={filter.search}
            onChange={(e) => handleSearchChange(e.target.value)}
            placeholder="Branch name or code..."
            className="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm"
          />
        </div>
      </div>
    </div>
  );
}
