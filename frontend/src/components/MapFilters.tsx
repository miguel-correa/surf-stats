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
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
                <div className="lg:col-span-2">
                    <label className="block text-sm font-semibold text-gray-300 mb-3">
                        Filter by Tier
                    </label>
                    <div className="flex flex-wrap gap-2">
                        {[1, 2, 3, 4, 5, 6, 7, 8].map((tier) => (
                            <label
                                key={tier}
                                className={`
                                    flex items-center gap-2 px-4 py-2 rounded-lg border-2 cursor-pointer transition-all
                                    ${filters.tiers.includes(tier)
                                        ? 'bg-cyan-500 border-cyan-400 text-gray-900 font-semibold shadow-lg shadow-cyan-500/20 hover:bg-cyan-400 hover:border-cyan-300 active:scale-95'
      : 'bg-gray-700/50 border-gray-600 text-gray-300 hover:border-gray-500 hover:bg-gray-600 active:scale-95'                                    }
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
                        className="w-full px-4 py-2.5 border-2 border-gray-600 rounded-lg bg-gray-700/50 text-white placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition"
                    />
                </div>

            </div>
        </div>
    )
}