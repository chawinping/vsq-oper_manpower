'use client';

import { useState } from 'react';
import { format } from 'date-fns';

export interface AllocationSuggestion {
  id: string;
  rotation_staff_id: string;
  rotation_staff?: { name: string };
  branch_id: string;
  branch?: { name: string; code: string };
  date: string;
  position_id: string;
  position?: { name: string };
  status: 'pending' | 'approved' | 'rejected' | 'modified';
  confidence: number;
  reason: string;
}

interface SuggestionPanelProps {
  suggestions: AllocationSuggestion[];
  onApprove: (suggestionId: string) => Promise<void>;
  onReject: (suggestionId: string) => Promise<void>;
  onModify?: (suggestionId: string) => void;
}

export default function SuggestionPanel({ suggestions, onApprove, onReject, onModify }: SuggestionPanelProps) {
  const [processing, setProcessing] = useState<string | null>(null);

  const handleApprove = async (suggestionId: string) => {
    setProcessing(suggestionId);
    try {
      await onApprove(suggestionId);
    } finally {
      setProcessing(null);
    }
  };

  const handleReject = async (suggestionId: string) => {
    setProcessing(suggestionId);
    try {
      await onReject(suggestionId);
    } finally {
      setProcessing(null);
    }
  };

  const pendingSuggestions = suggestions.filter(s => s.status === 'pending');

  if (pendingSuggestions.length === 0) {
    return (
      <div className="p-4 bg-gray-50 rounded-lg text-center text-gray-500">
        No pending suggestions
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Pending Suggestions ({pendingSuggestions.length})</h3>
      {pendingSuggestions.map((suggestion) => (
        <div
          key={suggestion.id}
          className="p-4 bg-white border border-gray-200 rounded-lg shadow-sm"
        >
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <div className="font-medium">
                {suggestion.rotation_staff?.name || 'Unknown'} → {suggestion.branch?.name || 'Unknown'}
              </div>
              <div className="text-sm text-gray-600 mt-1">
                {format(new Date(suggestion.date), 'MMM d, yyyy')} • {suggestion.position?.name || 'Unknown Position'}
              </div>
              <div className="text-xs text-gray-500 mt-1">
                Confidence: {(suggestion.confidence * 100).toFixed(0)}%
              </div>
              {suggestion.reason && (
                <div className="text-xs text-gray-600 mt-2 italic">
                  {suggestion.reason}
                </div>
              )}
            </div>
            <div className="flex gap-2 ml-4">
              <button
                onClick={() => handleApprove(suggestion.id)}
                disabled={processing === suggestion.id}
                className="px-3 py-1 bg-green-600 text-white text-sm rounded hover:bg-green-700 disabled:opacity-50"
              >
                {processing === suggestion.id ? 'Processing...' : 'Approve'}
              </button>
              {onModify && (
                <button
                  onClick={() => onModify(suggestion.id)}
                  className="px-3 py-1 bg-blue-600 text-white text-sm rounded hover:bg-blue-700"
                >
                  Modify
                </button>
              )}
              <button
                onClick={() => handleReject(suggestion.id)}
                disabled={processing === suggestion.id}
                className="px-3 py-1 bg-red-600 text-white text-sm rounded hover:bg-red-700 disabled:opacity-50"
              >
                Reject
              </button>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
