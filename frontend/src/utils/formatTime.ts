export function formatTimeMs(timeMs: number | null | undefined): string {
    if (timeMs == null) {
        return '—';
    }

    const totalMilliseconds = Math.max(0, timeMs);
    const minutes = Math.floor(totalMilliseconds / 60000);
    const seconds = Math.floor((totalMilliseconds % 60000) / 1000);
    const milliseconds = totalMilliseconds % 1000;

    return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}.${String(milliseconds).padStart(3, '0')}`;
}

