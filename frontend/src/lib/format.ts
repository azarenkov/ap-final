

type TimestampLike = string | { seconds?: number | string; nanos?: number } | null | undefined;

function toDate(t: TimestampLike): Date | null {
    if (t == null) return null;
    if (typeof t === "string") {
        const d = new Date(t);
        return isNaN(d.getTime()) ? null : d;
    }
    if (typeof t === "object" && t.seconds != null) {
        const secs = typeof t.seconds === "string" ? parseInt(t.seconds, 10) : t.seconds;
        const ms = secs * 1000 + Math.floor((t.nanos ?? 0) / 1_000_000);
        const d = new Date(ms);
        return isNaN(d.getTime()) ? null : d;
    }
    return null;
}

export function formatTime(t?: TimestampLike): string {
    const d = toDate(t);
    if (!d) return "—";
    return d.toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit", hour12: false });
}

export function formatDate(t?: TimestampLike): string {
    const d = toDate(t);
    if (!d) return "—";
    return d.toLocaleDateString("en-GB", { day: "2-digit", month: "short", year: "numeric" });
}

export function formatDuration(from: TimestampLike, to: TimestampLike): string {
    const a = toDate(from);
    const b = toDate(to);
    if (!a || !b) return "—";
    const ms = b.getTime() - a.getTime();
    if (!Number.isFinite(ms) || ms <= 0) return "—";
    const totalMin = Math.round(ms / 60000);
    const h = Math.floor(totalMin / 60);
    const m = totalMin % 60;
    return `${h}h ${String(m).padStart(2, "0")}m`;
}

export function formatPrice(cents?: number): string {
    if (cents == null) return "—";
    const value = cents / 100;
    return value.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

export function classnames(...parts: Array<string | false | null | undefined>): string {
    return parts.filter(Boolean).join(" ");
}
