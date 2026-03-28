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

export interface UserDate {
  id: string;
  venue_id: string;
  venue_name: string;
  venue_address: string;
  user_pair_id: string;
  date: string;
  start_time: string;
  status: string;
  external_reservation_id: string | null;
  reservation_status: string;
  created_at: string;
}

export interface VenueWithCapacity extends Venue {
  capacity: number;
  available: number;
  can_switch: boolean;
}

export interface AuthResponse {
  token: string;
  id: string;
  email: string;
  user_pair_id: string;
  full_name: string;
}

export interface CreateDateRequest {
  venue_id: string;
  user_pair_id: string;
  date: string;
  start_time: string;
}

// Get token from localStorage
function getToken(): string | null {
  if (typeof window !== 'undefined') {
    const user = localStorage.getItem('breeze_user');
    if (user) {
      return JSON.parse(user).token;
    }
  }
  return null;
}

async function fetchApi<T>(path: string, options?: RequestInit, auth: boolean = false): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...options?.headers as Record<string, string>,
  };

  // Add auth token if required
  if (auth) {
    const token = getToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(`API error: ${response.status} ${error}`);
  }

  return response.json();
}

export const api = {
  // Auth
  login: (email: string, password: string) =>
    fetchApi<AuthResponse>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  register: (email: string, password: string, fullName: string) =>
    fetchApi<AuthResponse>('/api/v1/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password, full_name: fullName }),
    }),

  getMe: () =>
    fetchApi<{ id: string; email: string; user_pair_id: string; full_name: string }>('/api/v1/users/me', {}, true),

  // Dates
  getMyDates: () =>
    fetchApi<{ dates: UserDate[] }>('/api/v1/users/me/dates', {}, true),

  createDate: (data: CreateDateRequest) =>
    fetchApi<UserDate>('/api/v1/dates', {
      method: 'POST',
      body: JSON.stringify(data),
    }, true),

  // Date modification
  cancelDate: (dateId: string) =>
    fetchApi<{ message: string; status: string }>(`/api/v1/dates/${dateId}/cancel`, {
      method: 'POST',
    }, true),

  rescheduleDate: (dateId: string, newDate: string, newStartTime: string) =>
    fetchApi<{ message: string; date: UserDate }>(`/api/v1/dates/${dateId}/reschedule`, {
      method: 'POST',
      body: JSON.stringify({ new_date: newDate, new_start_time: newStartTime }),
    }, true),

  getRescheduleOptions: (dateId: string) =>
    fetchApi<{ slots: SlotAvailability[] }>(`/api/v1/dates/${dateId}/reschedule-options`, {}, true),

  switchVenue: (dateId: string, newVenueId: string) =>
    fetchApi<{ message: string; date: UserDate; venue_name: string }>(`/api/v1/dates/${dateId}/switch-venue`, {
      method: 'POST',
      body: JSON.stringify({ new_venue_id: newVenueId }),
    }, true),

  getSwitchOptions: (dateId: string) =>
    fetchApi<{ current_venue_id: string; date: string; start_time: string; venues: VenueWithCapacity[] }>(
      `/api/v1/dates/${dateId}/switch-options`, {}, true
    ),

  // Venues (public)
  getVenues: () => fetchApi<Venue[]>('/api/v1/venues'),
  getVenue: (id: string) => fetchApi<Venue>(`/api/v1/venues/${id}`),

  // Availability (public)
  getAvailability: (venueId: string, from?: string, to?: string) => {
    const params = new URLSearchParams();
    if (from) params.append('from', from);
    if (to) params.append('to', to);
    return fetchApi<{ venue_id: string; from: string; to: string; slots: SlotAvailability[] }>(
      `/api/v1/venues/${venueId}/availability?${params}`
    );
  },
};
