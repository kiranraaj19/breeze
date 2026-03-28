import Link from 'next/link';
import { api } from '@/lib/api';
import SyncStatus from '@/components/SyncStatus';
import DatesList from '@/components/DatesList';

export const dynamic = 'force-dynamic';

interface DatesPageProps {
  params: { id: string };
}

export default async function DatesPage({ params }: DatesPageProps) {
  const venue = await api.getVenue(params.id);

  if (!venue) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Venue not found</h1>
          <Link href="/" className="text-primary-600 hover:text-primary-700">
            ← Back to venues
          </Link>
        </div>
      </div>
    );
  }

  return (
    <main className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link
                href="/"
                className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              >
                <svg className="w-5 h-5 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                </svg>
              </Link>
              <div>
                <h1 className="text-xl font-bold text-gray-900">{venue.name}</h1>
                <p className="text-sm text-gray-500">{venue.address}, {venue.city}</p>
              </div>
            </div>
            <SyncStatus venueId={params.id} />
          </div>
        </div>
      </header>

      {/* Navigation tabs */}
      <div className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <nav className="flex space-x-8">
            <Link
              href={`/venues/${params.id}`}
              className="border-b-2 border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 py-4 px-1 text-sm font-medium"
            >
              Availability
            </Link>
            <Link
              href={`/venues/${params.id}/dates`}
              className="border-b-2 border-primary-500 text-primary-600 py-4 px-1 text-sm font-medium"
            >
              Upcoming Dates
            </Link>
          </nav>
        </div>
      </div>

      {/* Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-6">
          <h2 className="text-2xl font-bold text-gray-900">Upcoming Dates</h2>
          <p className="text-gray-500 mt-1">
            View all confirmed and pending dates at your venue.
          </p>
        </div>

        <DatesList venueId={params.id} />
      </div>
    </main>
  );
}
