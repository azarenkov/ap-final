import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";

import { api } from "../lib/api";
import { formatDate, formatPrice } from "../lib/format";
import type { Booking } from "../lib/types";
import a from "../styles/app.module.css";

interface BookingsResponse {
    bookings?: Booking[];
}

export function Bookings() {
    const q = useQuery({
        queryKey: ["bookings", "me"],
        queryFn: () => api<BookingsResponse>("/v1/bookings/me"),
    });

    const list = q.data?.bookings ?? [];

    return (
        <div className="fade-up">
            <header className={a.pageHeader}>
                <span className="eyebrow">Reservation ledger</span>
                <h1 className={a.pageTitle} style={{ fontSize: "clamp(48px,7vw,96px)" }}>
                    Your <em>journeys</em>.
                </h1>
                <p className={a.pageLede}>Every reservation, from pending to ticketed. Pay, cancel, or refund — all from one row.</p>
            </header>

            {q.isLoading ? (
                <p className={a.status}>Retrieving ledger…</p>
            ) : list.length === 0 ? (
                <div className={a.empty}>
                    <h3>The platform is <em>quiet</em>.</h3>
                    <p>You haven't reserved yet. <Link to="/" className="bare" style={{ color: "var(--burgundy)", borderBottom: "1px solid currentColor" }}>browse services →</Link></p>
                </div>
            ) : (
                <div style={{ borderTop: "1px solid var(--paper-line-strong)" }}>
                    {list.map((b) => (
                        <BookingRow key={b.id} b={b} />
                    ))}
                </div>
            )}
        </div>
    );
}

function BookingRow({ b }: { b: Booking }) {
    const badge =
        b.status === "CONFIRMED" ? a.badgeOk :
        b.status === "PENDING"   ? a.badgeWarn :
        b.status === "CANCELLED" ? a.badgeDanger :
                                   a.badgeNeutral;
    return (
        <Link to={`/bookings/${b.id}`} className="bare" style={{
            display: "grid",
            gridTemplateColumns: "auto 1fr auto auto auto",
            gap: 24,
            alignItems: "center",
            padding: "20px 4px",
            borderBottom: "1px solid var(--paper-line-strong)",
            transition: "background 0.15s var(--ease)",
        }}>
            <span className="mono" style={{ fontSize: 11, color: "var(--ink-3)", letterSpacing: "0.12em" }}>
                {b.id.slice(0, 8).toUpperCase()}
            </span>
            <div>
                <div style={{ fontFamily: "var(--font-display)", fontWeight: 500, fontSize: 20 }}>
                    Train {b.train_id.slice(0, 8)}… · {b.seat_count} seat{b.seat_count > 1 ? "s" : ""}
                </div>
                <div className="mono" style={{ fontSize: 11, color: "var(--ink-3)", marginTop: 4, letterSpacing: "0.08em" }}>
                    reserved {formatDate(b.created_at)}
                </div>
            </div>
            <span className="mono" style={{ fontSize: 14, color: "var(--ink)" }}>
                ${formatPrice(b.amount_cents)}
            </span>
            <span className={`${a.badge} ${badge}`}>{b.status}</span>
            <span style={{ color: "var(--burgundy)", fontFamily: "var(--font-mono)", fontSize: 11 }}>→</span>
        </Link>
    );
}
