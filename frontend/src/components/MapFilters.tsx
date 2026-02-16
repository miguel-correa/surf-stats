import type React from "react";
import type { MapFilters } from "../hooks/useMaps";

interface MapFiltersProps {
    filters: MapFilters;
    onFiltersChange: (filters: MapFilters) => void;
}

export function MapFilters({ filters, onFiltersChange }: MapFiltersProps) {
    const handleTierToggle = (tier: number) => {
        const newTiers = filters.tiers.includes(tier)
            ? filters.tiers.filter(t => t !== tier)
            : [...filters.tiers, tier];

        onFiltersChange({ ...filters, tiers: newTiers });
    };

    const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        onFiltersChange({ ...filters, search: e.target.value });
    }

    return (
        <div className="bg-gray-800/50 rounded-xl shadow-lg border border-gray-700 p-6 mb-8">
            <div className="grid grid-cols-1 lg:grid-cols-4 gap-6 items-start">
                <div className="lg:col-span-2">
                    <label className="block text-sm font-semibold text-gray-300 mb-3">
                        Filter by Tier
                    </label>
                    <div className="flex flex-nowrap gap-2 overflow-x-auto pb-1 pr-1">
                        {[1, 2, 3, 4, 5, 6, 7, 8].map((tier) => (
                            <label
                                key={tier}
                                className={`
                                    shrink-0 flex h-10 items-center gap-2 px-3 rounded-lg border-2 cursor-pointer transition-all
                                    ${filters.tiers.includes(tier)
                                        ? 'bg-indigo-500 border-indigo-400 text-white font-semibold shadow-lg shadow-indigo-500/20 hover:bg-indigo-400 hover:border-indigo-300 active:scale-95'
                                        : 'bg-gray-700/50 border-gray-600 text-gray-300 hover:border-gray-500 hover:bg-gray-600 active:scale-95'}
                                `}
                            >
                                <input
                                    type="checkbox"
                                    checked={filters.tiers.includes(tier)}
                                    onChange={() => handleTierToggle(tier)}
                                    className="hidden"
                                />
                                <span className="text-sm font-medium">Tier {tier}</span>
                            </label>
                        ))}
                    </div>
                </div>

                <div>
                    <label htmlFor="search" className="block text-sm font-semibold text-gray-300 mb-3">
                        Search Maps
                    </label>
                    <input
                        id="search"
                        type="text"
                        value={filters.search}
                        onChange={handleSearchChange}
                        placeholder="Type map name..."
                        className="h-10 w-full px-4 border-2 border-gray-600 rounded-lg bg-gray-700/50 text-white placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition"
                    />
                </div>

                <div>
                    <label className="block text-sm font-semibold text-gray-300 mb-3">
                        Map Type
                    </label>
                    <div className="inline-flex h-10 w-full rounded-lg border-2 border-gray-600 bg-gray-700/50 p-1">
                        {[
                            { value: 'all', label: 'All' },
                            { value: 'linear', label: 'Linear' },
                            { value: 'staged', label: 'Staged' },
                        ].map((option) => (
                            <button
                                key={option.value}
                                type="button"
                                onClick={() =>
                                    onFiltersChange({
                                        ...filters,
                                        linear: option.value as MapFilters['linear'],
                                    })
                                }
                                className={`h-full flex-1 rounded-md px-3 text-sm font-medium transition ${filters.linear === option.value
                                    ? 'bg-indigo-500 text-white shadow'
                                    : 'text-gray-300 hover:bg-gray-600/60'
                                    }`}
                            >
                                {option.label}
                            </button>
                        ))}
                    </div>
                </div>

            </div>
        </div>
    )
}
