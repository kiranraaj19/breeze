'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { api, UserDate, SlotAvailability, VenueWithCapacity } from '@/lib/api';

export default function MyDatesPage() {
  const { user, isLoading, logout } = useAuth();
  const router = useRouter();

  const [dates, setDates] = useState<UserDate[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  
  // Modal states
  const [showRescheduleModal, setShowRescheduleModal] = useState(false);
  const [showSwitchModal, setShowSwitchModal] = useState(false);
  const [selectedDate, setSelectedDate] = useState<UserDate | null>(null);
  const [rescheduleSlots, setRescheduleSlots] = useState<SlotAvailability[]>([]);
  const [switchVenues, setSwitchVenues] = useState<VenueWithCapacity[]>([]);
  const [modalLoading, setModalLoading] = useState(false);

  useEffect(() => {
    if (!isLoading && !user) {
      router.push('/login');
      return;
    }

    if (user) {
      loadDates();
    }
  }, [user, isLoading, router]);

  const loadDates = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await api.getMyDates();
      setDates(data.dates);
    } catch (err: any) {
      console.error('Failed to load dates', err);
      setError('Failed to load your dates. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = async (dateId: string) => {
    if (!confirm('Are you sure you want to cancel this date?')) return;
    
    setActionLoading(dateId);
    try {
      await api.cancelDate(dateId);
      await loadDates();
    } catch (err: any) {
      setError(err.message || 'Failed to cancel date');
    } finally {
      setActionLoading(null);
    }
  };

  const openRescheduleModal = async (date: UserDate) => {
    setSelectedDate(date);
    setShowRescheduleModal(true);
    setModalLoading(true);
    try {
      const data = await api.getRescheduleOptions(date.id);
      setRescheduleSlots(data.slots);
    } catch (err: any) {
      setError(err.message || 'Failed to load reschedule options');
    } finally {
      setModalLoading(false);
    }
  };

  const handleReschedule = async (newDate: string, newStartTime: string) => {
    if (!selectedDate) return;
    
    setModalLoading(true);
    try {
      await api.rescheduleDate(selectedDate.id, newDate, newStartTime);
      setShowRescheduleModal(false);
      await loadDates();
    } catch (err: any) {
      setError(err.message || 'Failed to reschedule date');
    } finally {
      setModalLoading(false);
    }
  };

  const openSwitchModal = async (date: UserDate) => {
    setSelectedDate(date);
    setShowSwitchModal(true);
    setModalLoading(true);
    try {
      const data = await api.getSwitchOptions(date.id);
      setSwitchVenues(data.venues);
    } catch (err: any) {
      setError(err.message || 'Failed to load switch options');
    } finally {
      setModalLoading(false);
    }
  };

  const handleSwitchVenue = async (newVenueId: string) => {
    if (!selectedDate) return;
    
    setModalLoading(true);
    try {
      await api.switchVenue(selectedDate.id, newVenueId);
      setShowSwitchModal(false);
      await loadDates();
    } catch (err: any) {
      setError(err.message || 'Failed to switch venue');
    } finally {
      setModalLoading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'confirmed':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">Confirmed</span>;
      case 'pending':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">Pending</span>;
      case 'rescheduling':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">Rescheduling</span>;
      case 'cancelled':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">Cancelled</span>;
      default:
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">{status}</span>;
    }
  };

  const formatDate = (dateStr: string, timeStr: string) => {
    const date = new Date(dateStr + 'T' + timeStr);
    return date.toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const formatTime = (timeStr: string) => {
    const [hours, minutes] = timeStr.split(':');
    const date = new Date();
    date.setHours(parseInt(hours), parseInt(minutes));
    return date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
  };

  const canModify = (status: string) => {
    return status !== 'cancelled';
  };

  if (isLoading || loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-primary-600 rounded-lg flex items-center justify-center">
                <svg className="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                </svg>
              </div>
              <div>
                <h1 className="text-xl font-bold text-gray-900">Breeze</h1>
                <p className="text-sm text-gray-500">Your dates</p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <button
                onClick={() => router.push('/book')}
                className="text-primary-600 hover:text-primary-700 font-medium text-sm"
              >
                Book New Date
              </button>
              <div className="flex items-center gap-3">
                <span className="text-sm text-gray-600 hidden sm:inline">{user?.email}</span>
                <button
                  onClick={logout}
                  className="text-gray-500 hover:text-gray-700 text-sm"
                >
                  Logout
                </button>
              </div>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8 flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">My Dates</h2>
            <p className="text-gray-500 mt-1">View and manage your upcoming dates.</p>
          </div>
          <button
            onClick={() => router.push('/book')}
            className="px-4 py-2 bg-primary-600 text-white font-medium rounded-lg hover:bg-primary-700 transition-colors"
          >
            Book New Date
          </button>
        </div>

        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
            {error}
            <button onClick={() => setError('')} className="ml-4 text-red-800 underline">Dismiss</button>
          </div>
        )}

        {dates.length === 0 ? (
          <div className="bg-white rounded-lg shadow border p-12 text-center">
            <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <h3 className="mt-4 text-lg font-medium text-gray-900">No dates yet</h3>
            <p className="mt-2 text-gray-500">You haven&apos;t booked any dates yet.</p>
            <button
              onClick={() => router.push('/book')}
              className="mt-6 px-4 py-2 bg-primary-600 text-white font-medium rounded-lg hover:bg-primary-700 transition-colors"
            >
              Book Your First Date
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            {dates.map((date) => (
              <div key={date.id} className="bg-white rounded-lg shadow border p-6">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      {getStatusBadge(date.status)}
                      <span className="text-sm text-gray-500">
                        Booked {new Date(date.created_at).toLocaleDateString()}
                      </span>
                    </div>
                    <h3 className="text-lg font-semibold text-gray-900">
                      {date.venue_name}
                    </h3>
                    <p className="text-gray-600">{date.venue_address}</p>
                    <div className="mt-3 flex items-center gap-6 text-sm">
                      <div className="flex items-center gap-2 text-gray-700">
                        <svg className="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                        </svg>
                        {formatDate(date.date, date.start_time)}
                      </div>
                      <div className="flex items-center gap-2 text-gray-700">
                        <svg className="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        {formatTime(date.start_time)}
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-2xl font-bold text-primary-600">
                      #{date.user_pair_id.slice(-4).toUpperCase()}
                    </div>
                    <div className="text-xs text-gray-500">Date ID</div>
                  </div>
                </div>

                {/* Action Buttons */}
                {canModify(date.status) && (
                  <div className="mt-4 pt-4 border-t flex items-center gap-3">
                    <button
                      onClick={() => openRescheduleModal(date)}
                      disabled={actionLoading === date.id}
                      className="px-3 py-1.5 text-sm font-medium text-blue-600 bg-blue-50 rounded-lg hover:bg-blue-100 transition-colors"
                    >
                      Reschedule
                    </button>
                    <button
                      onClick={() => openSwitchModal(date)}
                      disabled={actionLoading === date.id}
                      className="px-3 py-1.5 text-sm font-medium text-purple-600 bg-purple-50 rounded-lg hover:bg-purple-100 transition-colors"
                    >
                      Switch Venue
                    </button>
                    <button
                      onClick={() => handleCancel(date.id)}
                      disabled={actionLoading === date.id}
                      className="px-3 py-1.5 text-sm font-medium text-red-600 bg-red-50 rounded-lg hover:bg-red-100 transition-colors"
                    >
                      {actionLoading === date.id ? 'Cancelling...' : 'Cancel'}
                    </button>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </main>

      {/* Reschedule Modal */}
      {showRescheduleModal && selectedDate && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full max-h-[80vh] overflow-hidden">
            <div className="p-4 border-b">
              <h3 className="text-lg font-semibold text-gray-900">Reschedule Date</h3>
              <p className="text-sm text-gray-500">Select a new date and time</p>
            </div>
            <div className="p-4 overflow-y-auto max-h-[60vh]">
              {modalLoading ? (
                <div className="flex justify-center py-8">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
                </div>
              ) : !rescheduleSlots || rescheduleSlots.length === 0 ? (
                <p className="text-gray-500 text-center py-4">No available slots found</p>
              ) : (
                <div className="space-y-2">
                  {rescheduleSlots?.map((slot) => (
                    <button
                      key={`${slot.date}-${slot.start_time}`}
                      onClick={() => handleReschedule(slot.date, slot.start_time)}
                      disabled={modalLoading}
                      className="w-full p-3 text-left border rounded-lg hover:bg-gray-50 transition-colors"
                    >
                      <div className="flex justify-between items-center">
                        <div>
                          <div className="font-medium text-gray-900">
                            {new Date(slot.date).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })}
                          </div>
                          <div className="text-sm text-gray-500">
                            {slot.start_time} - {slot.end_time}
                          </div>
                        </div>
                        <span className="text-sm text-green-600 font-medium">
                          {slot.available} available
                        </span>
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </div>
            <div className="p-4 border-t bg-gray-50">
              <button
                onClick={() => setShowRescheduleModal(false)}
                className="w-full px-4 py-2 text-gray-700 bg-white border rounded-lg hover:bg-gray-50"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Switch Venue Modal */}
      {showSwitchModal && selectedDate && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full max-h-[80vh] overflow-hidden">
            <div className="p-4 border-b">
              <h3 className="text-lg font-semibold text-gray-900">Switch Venue</h3>
              <p className="text-sm text-gray-500">Select a different venue for the same time</p>
            </div>
            <div className="p-4 overflow-y-auto max-h-[60vh]">
              {modalLoading ? (
                <div className="flex justify-center py-8">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
                </div>
              ) : !switchVenues || switchVenues.length === 0 ? (
                <p className="text-gray-500 text-center py-4">No other venues available at this time</p>
              ) : (
                <div className="space-y-2">
                  {switchVenues?.map((venue) => (
                    <button
                      key={venue.id}
                      onClick={() => handleSwitchVenue(venue.id)}
                      disabled={modalLoading || !venue.can_switch}
                      className={`w-full p-3 text-left border rounded-lg transition-colors ${
                        venue.can_switch 
                          ? 'hover:bg-gray-50' 
                          : 'opacity-50 cursor-not-allowed bg-gray-100'
                      }`}
                    >
                      <div className="flex justify-between items-start">
                        <div>
                          <div className="font-medium text-gray-900">{venue.name}</div>
                          <div className="text-sm text-gray-500">{venue.address}</div>
                        </div>
                        <span className={`text-sm font-medium ${
                          venue.can_switch ? 'text-green-600' : 'text-red-600'
                        }`}>
                          {venue.can_switch ? `${venue.available} spots` : 'Full'}
                        </span>
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </div>
            <div className="p-4 border-t bg-gray-50">
              <button
                onClick={() => setShowSwitchModal(false)}
                className="w-full px-4 py-2 text-gray-700 bg-white border rounded-lg hover:bg-gray-50"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
