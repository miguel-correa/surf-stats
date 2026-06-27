import { ArrowPathIcon, ChevronDownIcon, ChevronRightIcon } from "@heroicons/react/20/solid";
import { Fragment, useState } from "react";
import type { SurfMap } from "../types/Map";
import type { SortColumn } from "../hooks/useMaps";
import { formatTimeMs } from "../utils/formatTime";

interface MapTableProps {
    maps: SurfMap[];
    sortCol: SortColumn;
    sortOrder: 'asc' | 'desc';
    onSort: (column: SortColumn) => void;
    primaryPlayerId: string;
    playerLabels: Record<string, string>;
    expandedMapIds: number[];
    onToggleExpandedMap: (mapID: number) => void;
    onRefreshRecords: () => void;
}

interface SortableHeaderProps {
    column: SortColumn;
    label: string;
    sortCol: SortColumn;
    sortOrder: 'asc' | 'desc';
    onSort: (column: SortColumn) => void;
    width?: string;
}

function SortableHeader({ column, label, sortCol, sortOrder, onSort, width }: SortableHeaderProps) {
    return (
        <th
            onClick={() => onSort(column)}
            className={`${width || ''} px-6 py-4 text-left text-xs font-bold text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-600/50 transition`}
        >
            {label}
            {sortCol === column && (
                <span className="ml-2">{sortOrder === 'asc' ? '↑' : '↓'}</span>
            )}
        </th>
    );
}

function MapTypeBadge({ linear }: { linear: number }) {
    const isLinear = linear === 1;
    return (
        <span
            className="ml-2 inline-flex items-center rounded-full bg-white/5 px-2 py-0.5 text-xs font-medium text-gray-200 ring-1 ring-white/15"
        >
            {isLinear ? 'Linear' : 'Staged'}
        </span>
    );
}

interface RefreshCooldown {
    availableAt: string;
}

export function MapTable({
    maps,
    sortCol,
    sortOrder,
    onSort,
    primaryPlayerId,
    playerLabels,
    expandedMapIds,
    onToggleExpandedMap,
    onRefreshRecords,
}: MapTableProps) {
    const [refreshingRows, setRefreshingRows] = useState<number[]>([]);
    const [refreshCooldowns, setRefreshCooldowns] = useState<Record<number, RefreshCooldown>>({});
    const primaryPlayerLabel = primaryPlayerId ? (playerLabels[primaryPlayerId] ?? 'Player') : 'Player';

    const getTierColor = (tier: number) => {
        const colors: Record<number, string> = {
            1: 'bg-green-500/20 text-gray-300',
            2: 'bg-lime-500/20 text-gray-300',
            3: 'bg-yellow-500/20 text-gray-300',
            4: 'bg-orange-500/20 text-gray-300',
            5: 'bg-red-400/20 text-gray-300',
            6: 'bg-red-500/20 text-gray-300',
            7: 'bg-red-600/20 text-gray-300',
            8: 'bg-red-700/20 text-gray-300',
        };
        return colors[tier] || 'bg-gray-500/20 text-gray-300';
    };

    const formatGroupTier = (groupTier?: number | null) => {
        if (groupTier == null) {
            return '-';
        }

        return `G${groupTier}`;
    };

    const formatAvailableAt = (value?: string | null) => {
        if (!value) {
            return '';
        }
        return new Date(value).toLocaleString();
    };

    const getAvailableAt = (map: SurfMap) => {
        return refreshCooldowns[map.ksf_map_id]?.availableAt ?? map.record_refresh_available_at ?? null;
    };

    const isRefreshCoolingDown = (map: SurfMap) => {
        const availableAt = getAvailableAt(map);
        const status = map.record_refresh_status;
        return status !== 'queued' && status !== 'running' && availableAt ? new Date(availableAt).getTime() > Date.now() : false;
    };

    const refreshButtonClass = (state: 'available' | 'posting' | 'queued' | 'running' | 'cooldown') => {
        const base = 'inline-flex h-8 w-8 items-center justify-center rounded-full border transition disabled:cursor-not-allowed';
        if (state === 'posting' || state === 'running') {
            return `${base} border-cyan-300 bg-cyan-400/20 text-cyan-100`;
        }
        if (state === 'queued') {
            return `${base} border-amber-300/70 bg-amber-400/15 text-amber-100`;
        }
        if (state === 'cooldown') {
            return `${base} border-gray-700 bg-gray-800/70 text-gray-600`;
        }
        return `${base} border-emerald-300/70 bg-emerald-400/15 text-emerald-100 shadow-sm shadow-emerald-500/20 hover:border-emerald-200 hover:bg-emerald-400/25`;
    };

    const refreshButtonState = (map: SurfMap, posting: boolean): 'available' | 'posting' | 'queued' | 'running' | 'cooldown' => {
        if (posting) return 'posting';
        if (map.record_refresh_status === 'queued') return 'queued';
        if (map.record_refresh_status === 'running') return 'running';
        if (isRefreshCoolingDown(map)) return 'cooldown';
        return 'available';
    };

    const refreshButtonTitle = (map: SurfMap, state: 'available' | 'posting' | 'queued' | 'running' | 'cooldown') => {
        if (state === 'posting') return 'Queueing refresh';
        if (state === 'queued') return 'Refresh queued';
        if (state === 'running') return 'Refresh running';
        if (state === 'cooldown') return `Available ${formatAvailableAt(getAvailableAt(map))}`;
        return 'Refresh player records';
    };

    const handleRefreshRecords = async (map: SurfMap) => {
        if (refreshingRows.includes(map.ksf_map_id) || refreshButtonState(map, false) !== 'available') {
            return;
        }

        setRefreshingRows((current) => [...current, map.ksf_map_id]);
        try {
            const response = await fetch(`/api/maps/${map.ksf_map_id}/refresh-records`, {
                method: 'POST',
            });
            if (response.status === 429) {
                const body = await response.json() as { available_at?: string };
                if (body.available_at) {
                    setRefreshCooldowns((current) => ({
                        ...current,
                        [map.ksf_map_id]: { availableAt: body.available_at! },
                    }));
                }
                return;
            }
            if (!response.ok) {
                throw new Error('Failed to refresh records');
            }
            onRefreshRecords();
        } finally {
            setRefreshingRows((current) => current.filter((id) => id !== map.ksf_map_id));
        }
    };

    return (
        <div className="overflow-x-auto bg-gray-800/50 rounded-xl shadow-lg border border-gray-700">
            <table className="min-w-full divide-y divide-gray-700 table-fixed">
                <thead className="bg-gradient-to-r from-gray-800 to-gray-700">
                    <tr>
                        <th className="w-[4%] px-4 py-4 text-left text-xs font-bold text-gray-300 uppercase tracking-wider"> </th>
                        <th className="w-[30%] px-6 py-4 text-left text-xs font-bold text-gray-300 uppercase tracking-wider">Map Name</th>
                        <th className="w-[8%] px-6 py-4 text-left text-xs font-bold text-gray-300 uppercase tracking-wider">Tier</th>
                        <SortableHeader column="completions" label="Completions" sortCol={sortCol} sortOrder={sortOrder} onSort={onSort} width="w-[14%]" />
                        <SortableHeader column="difficulty" label="Comp/Hour" sortCol={sortCol} sortOrder={sortOrder} onSort={onSort} width="w-[12%]" />
                        <th className="w-[16%] px-6 py-4 text-left text-xs font-bold text-gray-300 uppercase tracking-wider">{primaryPlayerLabel} PB</th>
                        <th className="w-[10%] px-6 py-4 text-left text-xs font-bold text-gray-300 uppercase tracking-wider">{primaryPlayerLabel} Group</th>
                        <th className="w-[6%] px-4 py-4 text-left text-xs font-bold text-gray-300 uppercase tracking-wider"> </th>
                    </tr>
                </thead>
                <tbody className="bg-gray-900/50 divide-y divide-gray-700">
                    {maps.map((map, index) => {
                        const isExpanded = expandedMapIds.includes(map.ksf_map_id);
                        const primaryRecord = (map.player_records ?? []).find((record) => record.steam_id === primaryPlayerId);
                        const refreshing = refreshingRows.includes(map.ksf_map_id);
                        const refreshState = refreshButtonState(map, refreshing);

                        return (
                            <Fragment key={map.id}>
                                <tr
                                    className={`
                                        transition-colors
                                        ${index % 2 === 0 ? 'bg-transparent' : 'bg-transparent'}
                                        hover:bg-gray-700/50
                                    `}
                                >
                                    <td className="px-4 py-4">
                                        <button
                                            type="button"
                                            onClick={() => onToggleExpandedMap(map.ksf_map_id)}
                                            className="inline-flex h-8 w-8 items-center justify-center rounded-full border border-white/10 bg-white/5 text-gray-300 hover:bg-white/10"
                                        >
                                            {isExpanded ? <ChevronDownIcon className="size-4" /> : <ChevronRightIcon className="size-4" />}
                                        </button>
                                    </td>
                                    <td className="px-6 py-4 text-sm font-semibold text-white">
                                        <div className="flex items-center justify-between gap-3">
                                            <a href={`https://ksf.surf/maps/${map.name}`}
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                className="truncate text-white/90 underline decoration-white/30 underline-offset-2 hover:text-white hover:decoration-white/70 transition">
                                                {map.name}
                                            </a>
                                            <MapTypeBadge linear={map.linear} />
                                        </div>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <span className={`inline-flex items-center px-2.5 py-1 rounded text-xs font-medium ${getTierColor(map.tier)}`}>
                                            {map.tier}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-white font-medium">
                                        {map.completions.toLocaleString()}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm font-semibold text-white">
                                        {map.comp_per_hour.toFixed(2)}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-white">
                                        <div className="font-semibold text-cyan-100">{formatTimeMs(map.player_best_time_ms ?? null)}</div>
                                        <div className="text-xs text-gray-400">{primaryPlayerLabel}</div>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-white">
                                        <span className="inline-flex rounded-full bg-white/10 px-2.5 py-1 text-xs font-semibold text-gray-100">
                                            {formatGroupTier(primaryRecord?.group_tier)}
                                        </span>
                                    </td>
                                    <td className="px-4 py-4">
                                        <button
                                            type="button"
                                            onClick={() => handleRefreshRecords(map)}
                                            disabled={refreshState !== 'available'}
                                            title={refreshButtonTitle(map, refreshState)}
                                            className={refreshButtonClass(refreshState)}
                                        >
                                            <ArrowPathIcon className={`size-4 ${refreshState === 'posting' || refreshState === 'running' ? 'animate-spin' : ''}`} />
                                        </button>
                                    </td>
                                </tr>
                                {isExpanded && (
                                    <tr className="bg-gray-950/40">
                                        <td colSpan={8} className="px-6 py-4">
                                            <table className="w-full border-separate border-spacing-0 text-sm">
                                                <thead>
                                                    <tr className="text-xs font-semibold uppercase text-gray-400">
                                                        <th className="py-2 pr-4 text-left">Player</th>
                                                        <th className="py-2 pr-4 text-left">Time</th>
                                                        <th className="py-2 pr-4 text-left">Rank</th>
                                                        <th className="py-2 text-left">Group</th>
                                                    </tr>
                                                </thead>
                                                <tbody className="divide-y divide-white/5">
                                                    {[...(map.player_records ?? [])].sort((a, b) => (a.rank ?? Infinity) - (b.rank ?? Infinity)).map((record) => {
                                                        const isPrimary = record.steam_id === primaryPlayerId;
                                                        return (
                                                            <tr key={record.steam_id} className={isPrimary ? 'bg-cyan-500/10' : ''}>
                                                                <td className={`py-2 pr-4 font-semibold text-white ${isPrimary ? 'border-l-4 border-cyan-300 pl-3' : 'pl-4'}`}>
                                                                    {playerLabels[record.steam_id] ?? record.steam_id}
                                                                    {isPrimary && <span className="ml-2 text-xs text-cyan-300">★</span>}
                                                                </td>
                                                                <td className="py-2 pr-4 font-semibold text-cyan-100">{formatTimeMs(record.best_time_ms ?? null)}</td>
                                                                <td className="py-2 pr-4 text-gray-300">{record.rank != null ? `#${record.rank}` : '—'}</td>
                                                                <td className="py-2 text-gray-300">{formatGroupTier(record.group_tier)}</td>
                                                            </tr>
                                                        );
                                                    })}
                                                </tbody>
                                            </table>
                                        </td>
                                    </tr>
                                )}
                            </Fragment>
                        );
                    })}
                </tbody>
            </table>
        </div>
    )
}
