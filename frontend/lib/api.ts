const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface Venue {
  id: string;
  name: string;
  address: string;
  city: string;
  timezone: string;
  status: string;
}

export interface SlotAvailability {
  date: string;
  start_time: string;
  end_time: string;
  capacity: number;
  booked: number;
  available: number;
}

export interface DateItem {
  id: string;
  user_pair_id: string;
  date: string;
  start_time: string;
  status: string;
  external_reservation_id: string | null;
  reservation_status: string;
  created_at: string;
}

export interface SyncLog {
  id: string;
  venue_id: string | null;
  sync_type: string;
  status: string;
  started_at: string;
  completed_at: string | null;
  records_processed: number;
  error_message: string | null;
  created_at: string;
}

async function fetchApi<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`API error: ${response.status} ${response.statusText}`);
  }

  return response.json();
}

export const api = {
  // Venues
  getVenues: () => fetchApi<Venue[]>('/api/v1/venues'),
  getVenue: (id: string) => fetchApi<Venue>(`/api/v1/venues/${id}`),

  // Availability
  getAvailability: (venueId: string, from?: string, to?: string) => {
    const params = new URLSearchParams();
    if (from) params.append('from', from);
    if (to) params.append('to', to);
    return fetchApi<{ venue_id: string; from: string; to: string; slots: SlotAvailability[] }>(
      `/api/v1/venues/${venueId}/availability?${params}`
    );
  },

  // Dates
  getDates: (venueId: string) =>
    fetchApi<{ venue_id: string; dates: DateItem[] }>(`/api/v1/venues/${venueId}/dates`),

  // Sync
  getSyncStatus: (venueId: string) =>
    fetchApi<{ venue_id: string; last_sync: SyncLog | null }>(`/api/v1/venues/${venueId}/sync-status`),

  triggerSync: (venueId: string) =>
    fetchApi<{ venue_id: string; status: string }>(`/api/v1/venues/${venueId}/sync`, {
      method: 'POST',
    }),
};
