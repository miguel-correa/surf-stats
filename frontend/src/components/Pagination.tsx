import { ChevronLeftIcon, ChevronRightIcon } from '@heroicons/react/20/solid';

interface PaginationProps {
  currentPage: number;
  perPage: number;
  totalItems: number;
  onPageChange: (page: number) => void;
  label?: string;
}

type PageItem = number | 'ellipsis';

function getVisiblePages(
  currentPage: number,
  totalPages: number,
  maxVisible = 5
): PageItem[] {
  // If everything fits, just show all pages
  if (totalPages <= maxVisible) {
    return Array.from({ length: totalPages }, (_, i) => i + 1);
  }

  const pages: PageItem[] = [];

  const half = Math.floor(maxVisible / 2);

  let windowStart = currentPage - half;
  let windowEnd = currentPage + half;

  // Clamp window to bounds
  if (windowStart < 1) {
    windowStart = 1;
    windowEnd = maxVisible;
  }

  if (windowEnd > totalPages) {
    windowEnd = totalPages;
    windowStart = totalPages - maxVisible + 1;
  }

  // First page + left ellipsis
  if (windowStart > 1) {
    pages.push(1);
    if (windowStart > 2) {
      pages.push('ellipsis');
    }
  }

  // Window pages (exactly maxVisible)
  for (let i = windowStart; i <= windowEnd; i++) {
    pages.push(i);
  }

  // Right ellipsis + last page
  if (windowEnd < totalPages) {
    if (windowEnd < totalPages - 1) {
      pages.push('ellipsis');
    }
    pages.push(totalPages);
  }

  return pages;
}


export function Pagination({
  currentPage,
  perPage,
  totalItems,
  onPageChange,
  label = 'results',
}: PaginationProps) {
  if (totalItems <= perPage) return null;

  const totalPages = Math.ceil(totalItems / perPage);

  const isFirstPage = currentPage === 1;
  const isLastPage = currentPage === totalPages;

  const start = (currentPage - 1) * perPage + 1;
  const end = Math.min(currentPage * perPage, totalItems);

  const pages = getVisiblePages(currentPage, totalPages, 3);

  return (
    <div className="w-full flex items-center justify-between  px-4 py-3 sm:px-6">
      {/* Mobile */}
      <div className="flex flex-1 justify-between sm:hidden">
        <button
          onClick={() => onPageChange(Math.max(1, currentPage - 1))}
          disabled={isFirstPage}
          className="relative inline-flex items-center rounded-md border border-white/10
                     bg-white/5 px-4 py-2 text-sm font-medium text-gray-200
                     hover:bg-white/10 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Previous
        </button>

        <button
          onClick={() => onPageChange(Math.min(totalPages, currentPage + 1))}
          disabled={isLastPage}
          className="relative ml-3 inline-flex items-center rounded-md border border-white/10
                     bg-white/5 px-4 py-2 text-sm font-medium text-gray-200
                     hover:bg-white/10 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Next
        </button>
      </div>

      {/* Desktop */}
      <div className="hidden sm:flex w-full items-center justify-between">
        <p className="text-sm text-gray-300">
          Showing <span className="font-medium">{start}</span> to{' '}
          <span className="font-medium">{end}</span> of{' '}
          <span className="font-medium">{totalItems}</span> {label}
        </p>

        <nav aria-label="Pagination" className="isolate inline-flex -space-x-px rounded-md">
          {/* Previous */}
          <button
            onClick={() => onPageChange(Math.max(1, currentPage - 1))}
            disabled={isFirstPage}
            className="relative inline-flex items-center rounded-l-md px-2 py-2
                       text-gray-400 inset-ring inset-ring-gray-700
                       hover:bg-white/5 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span className="sr-only">Previous</span>
            <ChevronLeftIcon className="size-5" />
          </button>

          {/* Page numbers */}
          {pages.map((p, idx) =>
            p === 'ellipsis' ? (
              <span
                key={`ellipsis-${idx}`}
                className="relative inline-flex items-center px-4 py-2
                           text-sm font-semibold text-gray-400
                           inset-ring inset-ring-gray-700"
              >
                …
              </span>
            ) : (
              <button
                key={p}
                onClick={() => onPageChange(p)}
                aria-current={p === currentPage ? 'page' : undefined}
                className={
                  p === currentPage
                    ? 'relative z-10 inline-flex items-center bg-indigo-500 px-4 py-2 text-sm font-semibold text-white'
                    : 'relative inline-flex items-center px-4 py-2 text-sm font-semibold text-gray-200 inset-ring inset-ring-gray-700 hover:bg-white/5'
                }
              >
                {p}
              </button>
            )
          )}

          {/* Next */}
          <button
            onClick={() => onPageChange(Math.min(totalPages, currentPage + 1))}
            disabled={isLastPage}
            className="relative inline-flex items-center rounded-r-md px-2 py-2
                       text-gray-400 inset-ring inset-ring-gray-700
                       hover:bg-white/5 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span className="sr-only">Next</span>
            <ChevronRightIcon className="size-5" />
          </button>
        </nav>
      </div>
    </div>
  );
}
