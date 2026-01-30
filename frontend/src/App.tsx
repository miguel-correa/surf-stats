import { useMaps, type MapFilters, type SortColumn } from "./hooks/useMaps"
import { MapTable } from "./components/MapTable";
import { useState } from "react";
import { MapFilters as MapFiltersComponent } from "./components/MapFilters";

function App() {
  const [filters, setFilters] = useState<MapFilters>({
    tiers: [],
    search: '',
    sort: 'difficulty',
    order: 'desc'
  })
  const { maps, isLoading, error } = useMaps(filters);

  const handleSort = (column: SortColumn) => {
    const newOrder = filters.sort === column && filters.order === 'desc' ? 'asc' : 'desc';
    setFilters({ ...filters, sort: column, order: newOrder});
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 p-8">
      <div className="max-w-7xl mx-auto">
        <header className="mb-10">
          <h1 className="text-5xl font-bold text-gray-900 mb-2">
            CS:S Surf Stats
          </h1>
          <p className="text-lg text-gray-600">
            Map statistics from KSF servers
          </p>
        </header>

        <main>
          <MapFiltersComponent filters={filters} onFiltersChange={setFilters} />

          {isLoading && (
            <div className="text-center py-12">
              <p className="text-gray-500 text-lg">Loading maps...</p>
            </div>
          )}

          {error && (
            <div className="bg-red-50 border-l-4 border-red-500 p-4 rounded-lg">
              <p className="text-red-700 font-medium">Error loading maps</p>
            </div>
          )}

          {!isLoading && !error && (
            <div>
              <div className="flex items-center justify-between mb-4">
                <p className="text-gray-700 font-medium">
                  Showing <span className="text-blue-600 font-bold">{maps.length}</span> maps
                </p>
              </div>

              <MapTable maps={maps} sortCol={filters.sort} sortOrder={filters.order} onSort={handleSort}/>
            </div>
          )}
        </main>
      </div>
    </div>
  )
}

export default App
