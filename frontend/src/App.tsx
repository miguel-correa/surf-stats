import { useMaps, type MapFilters, type SortColumn } from "./hooks/useMaps"
import { MapTable } from "./components/MapTable";
import { Pagination } from "./components/Pagination";
import { useEffect, useMemo, useState } from "react";
import { MapFilters as MapFiltersComponent } from "./components/MapFilters";
import { usePlayers } from "./hooks/usePlayers";

const PLAYER_IDS_STORAGE_KEY = 'surfstats.playerIds';
const PRIMARY_PLAYER_STORAGE_KEY = 'surfstats.primaryPlayerId';

function readStoredPlayerIds(): string[] {
  if (typeof window === 'undefined') {
    return [];
  }

  try {
    const raw = window.localStorage.getItem(PLAYER_IDS_STORAGE_KEY);
    if (raw == null) {
      return [];
    }

    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) {
      return [];
    }

    return parsed.filter((value): value is string => typeof value === 'string');
  } catch {
    return [];
  }
}

function readStoredPrimaryPlayerId(): string {
  if (typeof window === 'undefined') {
    return '';
  }

  return window.localStorage.getItem(PRIMARY_PLAYER_STORAGE_KEY) ?? '';
}

function normalizePlayerSelection(playerIds: string[], primaryPlayerId: string, validPlayerIds: Set<string>) {
  const nextPlayerIds: string[] = [];
  const seen = new Set<string>();

  for (const playerId of playerIds) {
    if (!validPlayerIds.has(playerId) || seen.has(playerId)) {
      continue;
    }

    seen.add(playerId);
    nextPlayerIds.push(playerId);
  }

  return {
    playerIds: nextPlayerIds,
    primaryPlayerId: nextPlayerIds.includes(primaryPlayerId) ? primaryPlayerId : (nextPlayerIds[0] ?? ''),
  };
}

function getInitialFilters(): MapFilters {
  return {
    tiers: [],
    search: '',
    linear: 'all',
    playerIds: readStoredPlayerIds(),
    primaryPlayerId: readStoredPrimaryPlayerId(),
    primaryGroups: [],
    completion: 'all',
    sort: 'difficulty',
    order: 'desc'
  };
}

function App() {
  const { players, isLoading: playersLoading, error: playersError } = usePlayers();
  const playerLabels = Object.fromEntries(players.map((player) => [player.steam_id, player.name || player.steam_id]));

  const [storedFilters, setStoredFilters] = useState<MapFilters>(getInitialFilters)
  const [page, setPage] = useState(1)
  const [searchText, setSearchText] = useState(storedFilters.search)
  const [mapsRefreshKey, setMapsRefreshKey] = useState(0)
  const [expandedMapIds, setExpandedMapIds] = useState<number[]>([])
  const validPlayerIds = useMemo(() => new Set(players.map((player) => player.steam_id)), [players]);
  const filters = useMemo(() => {
    if (playersLoading || playersError) {
      return storedFilters;
    }

    const nextSelection = normalizePlayerSelection(storedFilters.playerIds, storedFilters.primaryPlayerId, validPlayerIds);
    if (
      nextSelection.playerIds.length === storedFilters.playerIds.length &&
      nextSelection.playerIds.every((playerId, index) => playerId === storedFilters.playerIds[index]) &&
      nextSelection.primaryPlayerId === storedFilters.primaryPlayerId
    ) {
      return storedFilters;
    }

    return {
      ...storedFilters,
      playerIds: nextSelection.playerIds,
      primaryPlayerId: nextSelection.primaryPlayerId,
    };
  }, [playersError, playersLoading, storedFilters, validPlayerIds]);
  const { paginatedData, isLoading, error } = useMaps(filters, { page, refreshKey: mapsRefreshKey });
  const hasActiveRefresh = (paginatedData?.maps ?? []).some((map) =>
    map.record_refresh_status === 'queued' || map.record_refresh_status === 'running'
  );

  useEffect(() => {
    const id = setTimeout(() => {
      setStoredFilters(prev => prev.search === searchText ? prev : { ...prev, search: searchText });
      setPage(1);
    }, 300);
    return () => clearTimeout(id);
  }, [searchText]);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }

    window.localStorage.setItem(PLAYER_IDS_STORAGE_KEY, JSON.stringify(filters.playerIds));
    if (filters.primaryPlayerId) {
      window.localStorage.setItem(PRIMARY_PLAYER_STORAGE_KEY, filters.primaryPlayerId);
      return;
    }

    window.localStorage.removeItem(PRIMARY_PLAYER_STORAGE_KEY);
  }, [filters.playerIds, filters.primaryPlayerId]);

  useEffect(() => {
    if (!hasActiveRefresh) {
      return;
    }

    const id = window.setInterval(() => {
      setMapsRefreshKey((value) => value + 1);
    }, 5000);
    return () => window.clearInterval(id);
  }, [hasActiveRefresh]);

  const handleFiltersChange = (nextFilters: MapFilters) => {
    if (nextFilters.search !== searchText) {
      setSearchText(nextFilters.search);
    }
    setStoredFilters(nextFilters);
    setPage(1);
  };

  const handleSort = (column: SortColumn) => {
    const newOrder = filters.sort === column
      ? (filters.order === 'desc' ? 'asc' : 'desc')
      : (column === 'tier' || column === 'primary_group' ? 'asc' : 'desc');
    handleFiltersChange({ ...filters, sort: column, order: newOrder });
  };

  const handleToggleExpandedMap = (mapID: number) => {
    setExpandedMapIds((current) =>
      current.includes(mapID)
        ? current.filter((id) => id !== mapID)
        : [...current, mapID]
    );
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
          <MapFiltersComponent
            filters={filters}
            onFiltersChange={handleFiltersChange}
            searchText={searchText}
            onSearchChange={setSearchText}
            players={players}
            playersLoading={playersLoading}
            playersError={playersError}
          />

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
              <MapTable
                maps={paginatedData?.maps || []}
                sortCol={filters.sort}
                sortOrder={filters.order}
                onSort={handleSort}
                primaryPlayerId={filters.primaryPlayerId}
                playerLabels={playerLabels}
                expandedMapIds={expandedMapIds}
                onToggleExpandedMap={handleToggleExpandedMap}
                onRefreshRecords={() => setMapsRefreshKey((value) => value + 1)}
              />
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
