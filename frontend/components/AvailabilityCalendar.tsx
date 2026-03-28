'use client';

import { useState, useEffect } from 'react';
import { api, SlotAvailability } from '@/lib/api';

interface AvailabilityCalendarProps {
  venueId: string;
}

interface DaySlots {
  date: Date;
  slots: SlotAvailability[];
}

export default function AvailabilityCalendar({ venueId }: AvailabilityCalendarProps) {
  const [weekOffset, setWeekOffset] = useState(0);
  const [slots, setSlots] = useState<SlotAvailability[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Calculate week dates
  const getWeekDates = (offset: number) => {
    const today = new Date();
    const startOfWeek = new Date(today);
    startOfWeek.setDate(today.getDate() - today.getDay() + offset * 7);
    
    const dates: Date[] = [];
    for (let i = 0; i < 7; i++) {
      const date = new Date(startOfWeek);
      date.setDate(startOfWeek.getDate() + i);
      dates.push(date);
    }
    return dates;
  };

  const weekDates = getWeekDates(weekOffset);

  useEffect(() => {
    const fetchAvailability = async () => {
      setLoading(true);
      try {
        const from = weekDates[0].toISOString().split('T')[0];
        const to = weekDates[6].toISOString().split('T')[0];
        const data = await api.getAvailability(venueId, from, to);
        setSlots(data.slots);
        setError(null);
      } catch (err) {
        setError('Failed to load availability');
      } finally {
        setLoading(false);
      }
    };

    fetchAvailability();
  }, [venueId, weekOffset]);

  // Group slots by date
  const slotsByDay = weekDates.map((date) => {
    const dateStr = date.toISOString().split('T')[0];
    return {
      date,
      slots: slots.filter((s) => s.date === dateStr),
    };
  });

  const formatDate = (date: Date) => {
    return date.toLocaleDateString('en-US', {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
    });
  };

  const isToday = (date: Date) => {
    const today = new Date();
    return date.toDateString() === today.toDateString();
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64 text-red-600">
        {error}
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow border overflow-hidden">
      {/* Week navigation */}
      <div className="flex items-center justify-between px-6 py-4 border-b bg-gray-50">
        <button
          onClick={() => setWeekOffset(weekOffset - 1)}
          className="p-2 hover:bg-gray-200 rounded-full transition-colors"
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h3 className="text-lg font-semibold text-gray-900">
          Week of {weekDates[0].toLocaleDateString('en-US', { month: 'long', day: 'numeric' })}
        </h3>
        <button
          onClick={() => setWeekOffset(weekOffset + 1)}
          className="p-2 hover:bg-gray-200 rounded-full transition-colors"
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>

      {/* Calendar grid */}
      <div className="grid grid-cols-7 divide-x divide-gray-200">
        {slotsByDay.map(({ date, slots }) => (
          <div key={date.toISOString()} className="min-h-[300px]">
            <div
              className={`px-3 py-2 text-center border-b ${
                isToday(date) ? 'bg-primary-50' : 'bg-gray-50'
              }`}
            >
              <div
                className={`text-sm font-medium ${
                  isToday(date) ? 'text-primary-700' : 'text-gray-900'
                }`}
              >
                {date.toLocaleDateString('en-US', { weekday: 'short' })}
              </div>
              <div
                className={`text-lg ${
                  isToday(date) ? 'text-primary-900 font-semibold' : 'text-gray-600'
                }`}
              >
                {date.getDate()}
              </div>
            </div>

            <div className="p-2 space-y-2">
              {slots.length === 0 ? (
                <div className="text-center text-gray-400 text-sm py-4">No slots</div>
              ) : (
                slots.map((slot, idx) => (
                  <div
                    key={idx}
                    className={`p-2 rounded-lg text-sm ${
                      slot.available === 0
                        ? 'bg-gray-100 text-gray-500'
                        : slot.available <= 2
                        ? 'bg-yellow-50 border border-yellow-200 text-yellow-800'
                        : 'bg-green-50 border border-green-200 text-green-800'
                    }`}
                  >
                    <div className="font-medium">
                      {slot.start_time} - {slot.end_time}
                    </div>
                    <div className="text-xs mt-1">
                      {slot.available === 0 ? (
                        <span className="font-semibold">FULL</span>
                      ) : (
                        <>
                          <span className="font-semibold">{slot.available}</span>/{slot.capacity} available
                        </>
                      )}
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Legend */}
      <div className="px-6 py-3 border-t bg-gray-50 flex items-center gap-6 text-sm">
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 bg-green-50 border border-green-200 rounded"></div>
          <span className="text-gray-600">Available</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 bg-yellow-50 border border-yellow-200 rounded"></div>
          <span className="text-gray-600">Low availability</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 bg-gray-100 rounded"></div>
          <span className="text-gray-600">Fully booked</span>
        </div>
      </div>
    </div>
  );
}
