'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { api, Venue, SlotAvailability } from '@/lib/api';

export default function BookPage() {
  const { user, isLoading } = useAuth();
  const router = useRouter();

  const [venues, setVenues] = useState<Venue[]>([]);
  const [selectedVenue, setSelectedVenue] = useState<string>('');
  const [selectedDate, setSelectedDate] = useState<string>('');
  const [selectedTime, setSelectedTime] = useState<string>('');
  const [slots, setSlots] = useState<SlotAvailability[]>([]);
  const [loading, setLoading] = useState(true);
  const [booking, setBooking] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!isLoading && !user) {
      router.push('/login');
      return;
    }

    if (user) {
      fetchVenues();
    }
  }, [user, isLoading, router]);

  const fetchVenues = async () => {
    try {
      const data = await api.getVenues();
      setVenues(data);
      if (data.length > 0) {
        setSelectedVenue(data[0].id);
      }
    } catch (err) {
      setError('Failed to load venues');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedVenue) {
      fetchAvailability();
    }
  }, [selectedVenue]);

  const fetchAvailability = async () => {
    try {
      const today = new Date().toISOString().split('T')[0];
      const nextMonth = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
      const data = await api.getAvailability(selectedVenue, today, nextMonth);
      setSlots(data.slots.filter(s => s.available > 0));
      setSelectedDate('');
      setSelectedTime('');
    } catch (err) {
      console.error('Failed to load availability', err);
    }
  };

  const getDatesWithSlots = () => {
    const dates = [...new Set(slots.map(s => s.date))];
    return dates.sort();
  };

  const getTimesForDate = (date: string) => {
    return slots.filter(s => s.date === date);
  };

  const handleBook = async () => {
    if (!selectedVenue || !selectedDate || !selectedTime) {
      setError('Please select a venue, date and time');
      return;
    }

    if (!user) {
      setError('Please log in first');
      return;
    }

    setBooking(true);
    setError('');

    try {
      await api.createDate({
        venue_id: selectedVenue,
        user_pair_id: user.userPairId,
        date: selectedDate,
        start_time: selectedTime,
      });
      setSuccess(true);
      setTimeout(() => {
        router.push('/my-dates');
      }, 2000);
    } catch (err: any) {
      setError(err.message || 'Failed to book date');
    } finally {
      setBooking(false);
    }
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
                <p className="text-sm text-gray-500">Book your perfect date</p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <span className="text-sm text-gray-600">{user?.email}</span>
              <button
                onClick={() => router.push('/my-dates')}
                className="text-primary-600 hover:text-primary-700 font-medium text-sm"
              >
                My Dates
              </button>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {success ? (
          <div className="bg-green-50 border border-green-200 rounded-lg p-8 text-center">
            <svg className="mx-auto h-12 w-12 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <h2 className="mt-4 text-xl font-semibold text-green-900">Date Booked!</h2>
            <p className="mt-2 text-green-700">Your date has been scheduled. Redirecting to your dates...</p>
          </div>
        ) : (
          <>
            <div className="mb-8">
              <h2 className="text-2xl font-bold text-gray-900">Book a Date</h2>
              <p className="text-gray-500 mt-1">Select your preferred venue, date and time.</p>
            </div>

            {error && (
              <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
                {error}
              </div>
            )}

            <div className="bg-white rounded-lg shadow border p-6 space-y-6">
              {/* Venue Selection */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Select Venue</label>
                <select
                  value={selectedVenue}
                  onChange={(e) => setSelectedVenue(e.target.value)}
                  className="block w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
                >
                  {venues.map(venue => (
                    <option key={venue.id} value={venue.id}>
                      {venue.name} - {venue.address}
                    </option>
                  ))}
                </select>
              </div>

              {/* Date Selection */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Select Date</label>
                <div className="grid grid-cols-4 sm:grid-cols-7 gap-2">
                  {getDatesWithSlots().slice(0, 14).map(date => {
                    const dateObj = new Date(date);
                    const isSelected = selectedDate === date;
                    return (
                      <button
                        key={date}
                        onClick={() => {
                          setSelectedDate(date);
                          setSelectedTime('');
                        }}
                        className={`p-3 rounded-lg text-center transition-colors ${
                          isSelected
                            ? 'bg-primary-600 text-white'
                            : 'bg-gray-50 hover:bg-gray-100 text-gray-700'
                        }`}
                      >
                        <div className="text-xs uppercase">{dateObj.toLocaleDateString('en-US', { weekday: 'short' })}</div>
                        <div className="text-lg font-semibold">{dateObj.getDate()}</div>
                      </button>
                    );
                  })}
                </div>
              </div>

              {/* Time Selection */}
              {selectedDate && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Select Time</label>
                  <div className="grid grid-cols-3 sm:grid-cols-4 gap-2">
                    {getTimesForDate(selectedDate).map(slot => {
                      const isSelected = selectedTime === slot.start_time;
                      return (
                        <button
                          key={slot.start_time}
                          onClick={() => setSelectedTime(slot.start_time)}
                          className={`p-3 rounded-lg text-center transition-colors ${
                            isSelected
                              ? 'bg-primary-600 text-white'
                              : 'bg-gray-50 hover:bg-gray-100 text-gray-700'
                          }`}
                        >
                          <div className="font-medium">{slot.start_time}</div>
                          <div className={`text-xs ${isSelected ? 'text-primary-100' : 'text-gray-500'}`}>
                            {slot.available} spots left
                          </div>
                        </button>
                      );
                    })}
                  </div>
                </div>
              )}

              {/* Summary */}
              {selectedVenue && selectedDate && selectedTime && (
                <div className="bg-primary-50 border border-primary-200 rounded-lg p-4">
                  <h3 className="font-medium text-primary-900 mb-2">Booking Summary</h3>
                  <div className="text-sm text-primary-800 space-y-1">
                    <p><strong>Venue:</strong> {venues.find(v => v.id === selectedVenue)?.name}</p>
                    <p><strong>Date:</strong> {new Date(selectedDate).toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' })}</p>
                    <p><strong>Time:</strong> {selectedTime}</p>
                  </div>
                </div>
              )}

              {/* Book Button */}
              <button
                onClick={handleBook}
                disabled={!selectedVenue || !selectedDate || !selectedTime || booking}
                className="w-full py-3 px-4 bg-primary-600 text-white font-medium rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {booking ? 'Booking...' : 'Book Date'}
              </button>
            </div>
          </>
        )}
      </main>
    </div>
  );
}
