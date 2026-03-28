'use client';

import { useState, useEffect } from 'react';
import { api, SyncLog } from '@/lib/api';

interface SyncStatusProps {
  venueId: string;
  onSync?: () => void; // Callback after sync completes
}

export default function SyncStatus({ venueId, onSync }: SyncStatusProps) {
  const [syncStatus, setSyncStatus] = useState<SyncLog | null>(null);
  const [isSyncing, setIsSyncing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchSyncStatus = async () => {
    try {
      const data = await api.getSyncStatus(venueId);
      setSyncStatus(data.last_sync);
      setError(null);
    } catch (err) {
      setError('Failed to fetch sync status');
    }
  };

  const handleSync = async () => {
    setIsSyncing(true);
    try {
      await api.triggerSync(venueId);
      // Poll for completion
      let attempts = 0;
      const pollInterval = setInterval(async () => {
        attempts++;
        await fetchSyncStatus();
        
        // Check if sync completed (not running) or max attempts reached
        const status = syncStatus?.status;
        if (status !== 'running' || attempts >= 10) {
          clearInterval(pollInterval);
          setIsSyncing(false);
          // Call the callback to refresh parent data
          if (onSync) {
            onSync();
          }
        }
      }, 1000);
      
      // Safety timeout
      setTimeout(() => {
        clearInterval(pollInterval);
        setIsSyncing(false);
        if (onSync) {
          onSync();
        }
      }, 15000);
      
    } catch (err) {
      setError('Sync failed');
      setIsSyncing(false);
    }
  };

  useEffect(() => {
    fetchSyncStatus();
    // Poll every 30 seconds
    const interval = setInterval(fetchSyncStatus, 30000);
    return () => clearInterval(interval);
  }, [venueId]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success':
        return 'bg-green-500';
      case 'failed':
        return 'bg-red-500';
      case 'running':
        return 'bg-yellow-500';
      default:
        return 'bg-gray-400';
    }
  };

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleString();
  };

  const getTimeAgo = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.round(diffMs / 60000);
    const diffHours = Math.round(diffMs / 3600000);
    const diffDays = Math.round(diffMs / 86400000);

    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return `${diffDays}d ago`;
  };

  return (
    <div className="flex items-center gap-4 bg-white px-4 py-3 rounded-lg shadow-sm border">
      <div className="flex items-center gap-2">
        <div
          className={`w-3 h-3 rounded-full ${
            syncStatus ? getStatusColor(syncStatus.status) : 'bg-gray-400'
          }`}
        />
        <span className="text-sm text-gray-600">
          {syncStatus ? (
            <>
              Last sync:{' '}
              <span className="font-medium text-gray-900">
                {getTimeAgo(syncStatus.started_at)}
              </span>
              {syncStatus.status === 'success' && (
                <span className="text-green-600 ml-1">
                  ({syncStatus.records_processed} records)
                </span>
              )}
              {syncStatus.status === 'failed' && (
                <span className="text-red-600 ml-1">(failed)</span>
              )}
            </>
          ) : (
            'Never synced'
          )}
        </span>
      </div>

      <button
        onClick={handleSync}
        disabled={isSyncing}
        className="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-md hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        {isSyncing ? (
          <>
            <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
                fill="none"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
            Syncing...
          </>
        ) : (
          <>
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            Sync Now
          </>
        )}
      </button>

      {error && <span className="text-sm text-red-600">{error}</span>}
    </div>
  );
}
