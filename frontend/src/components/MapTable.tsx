import type { SurfMap } from "../types/Map";
import type { SortColumn } from "../hooks/useMaps";

interface MapTableProps {
    maps: SurfMap[];
    sortCol: SortColumn;
    sortOrder: 'asc' | 'desc';
    onSort: (column: SortColumn) => void;
}

export function MapTable({ maps, sortCol, sortOrder, onSort }: MapTableProps) {
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

    const SortableHeader = ({ column, label, width }: { column: SortColumn, label: string, width?: string }) => (
        <th
            onClick={() => onSort(column)}
            className={`${width || ''} px-6 py-4 text-left text-xs font-bold text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-600/50 transition"`}
        >
            {label}
            {sortCol === column && (
                <span className="ml-2">{sortOrder === 'asc' ? '↑' : '↓'}</span>
            )}
        </th>
    );


    return (
        <div className="overflow-x-auto bg-gray-800/50 rounded-xl shadow-lg border border-gray-700">
            <table className="min-w-full divide-y divide-gray-700 table-fixed">
                <thead className="bg-gradient-to-r from-gray-800 to-gray-700">
                    <tr>
                        <th className="w-[40%] px-6 py-4 text-left text-xs font-bold text-gray-300 uppercase tracking-wider">Map Name</th>
                        <th className="w-[10%] px-6 py-4 text-left text-xs font-bold text-gray-300 uppercase tracking-wider">Tier</th>
                        <th className="w-[10%] px-6 py-4 text-left text-xs font-bold text-gray-300 uppercase tracking-wider">Year</th>
                        <SortableHeader column="completions" label="Completions" width="w-[15%]" />
                        <th className="w-[12%] px-6 py-4 text-left text-xs font-bold text-gray-300 uppercase tracking-wider">Hours Played</th>
                        <SortableHeader column="difficulty" label="Comp/Hour" width="w-[13%]" />
                    </tr>
                </thead>
                <tbody className="bg-gray-900/50 divide-y divide-gray-700">
                    {maps.map((map, index) => (
                        <tr
                            key={map.id}
                            className={`
                                transition-colors
                                ${index % 2 === 0 ? 'bg-transparent' : 'bg-transparent'}
                                hover:bg-gray-700/50
                            `}
                        >
                            <td className="px-6 py-4 whitespace-nowrap text-sm font-semibold text-white">
                                {map.name}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap">
                                <span className={`inline-flex items-center px-2.5 py-1 rounded text-xs font-medium ${getTierColor(map.tier)}`}>
                                    {map.tier}
                                </span>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-white">
                                <td>{new Date(map.year * 1000).getUTCFullYear()}</td>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-white font-medium">
                                {map.completions.toLocaleString()}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-white">
                                  <td>{(map.playtime_seconds / 3600).toFixed(1)}h</td>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm font-semibold text-white">
                                {map.comp_per_hour.toFixed(2)}
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    )
}