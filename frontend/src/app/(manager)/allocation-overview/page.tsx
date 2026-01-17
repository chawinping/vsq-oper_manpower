'use client';

import { useState } from 'react';
import { format, addDays, subDays } from 'date-fns';
import AllocationOverviewTable from '@/components/allocation/AllocationOverviewTable';

export default function AllocationOverviewPage() {
  const [currentDate, setCurrentDate] = useState(new Date());

  const handlePreviousDay = () => {
    setCurrentDate(subDays(currentDate, 1));
  };

  const handleNextDay = () => {
    setCurrentDate(addDays(currentDate, 1));
  };

  const handleToday = () => {
    setCurrentDate(new Date());
  };

  return (
    <div className="w-full p-6">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-3xl font-bold">Allocation Overview</h1>
        <div className="flex gap-2">
          <button
            onClick={handlePreviousDay}
            className="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-md"
          >
            Previous Day
          </button>
          <button
            onClick={handleToday}
            className="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-md"
          >
            Today
          </button>
          <button
            onClick={handleNextDay}
            className="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-md"
          >
            Next Day
          </button>
        </div>
      </div>

      <AllocationOverviewTable date={currentDate} />
    </div>
  );
}
