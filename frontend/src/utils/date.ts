// <input type="datetime-local"> gives "2026-10-19T12:00" (no timezone).
// The API expects RFC3339, so convert to a full ISO string. Empty -> undefined.
export function toISO(value?: string): string | undefined {
    return value ? new Date(value).toISOString() : undefined
}

// "2027-12-11T00:00:00Z" -> "2027-12-11"
export function formatDate(iso: string): string {
    return iso.slice(0, 10)
}