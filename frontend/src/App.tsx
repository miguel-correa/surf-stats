import { useMaps, type MapFilters, type SortColumn } from "./hooks/useMaps"
import { MapTable } from "./components/MapTable";
import { Pagination } from "./components/Pagination";
import { useEffect, useState } from "react";
import { MapFilters as MapFiltersComponent } from "./components/MapFilters";

function App() {
  const [filters, setFilters] = useState<MapFilters>({
    tiers: [],
    search: '',
    sort: 'difficulty',
    order: 'desc'
  })
  const [page, setPage] = useState(1)
  const { paginatedData, isLoading, error } = useMaps(filters, { page });

  useEffect(() => {
    setPage(1);
  }, [filters.tiers, filters.search, filters.sort, filters.order]);

  const handleSort = (column: SortColumn) => {
    const newOrder = filters.sort === column && filters.order === 'desc' ? 'asc' : 'desc';
    setFilters({ ...filters, sort: column, order: newOrder });
  };


  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-800 to-gray-700 p-8">
      <div className="fixed inset-0 pointer-events-none bg-[radial-gradient(circle,transparent_0%,rgba(0,0,0,0.5)_100%)]" />
      <div className="max-w-7xl mx-auto relative">
        <header className="mb-10">
          <h1 className="text-5xl font-bold text-white mb-2">
            Billy's Surf Stats Extravaganza
          </h1>
          <p className="text-lg text-gray-300">
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
              <MapTable maps={paginatedData?.maps || []} sortCol={filters.sort} sortOrder={filters.order} onSort={handleSort} />
              <Pagination
                currentPage={page}
                perPage={10}
                totalItems={paginatedData?.total || 0}
                label="maps"
                onPageChange={setPage}
              />
            </div>
          )}
        </main>
      </div>
    </div>
  )
}

export default App
