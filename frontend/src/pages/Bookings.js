import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { api } from "../lib/api";
import { formatDate, formatPrice } from "../lib/format";
import a from "../styles/app.module.css";
export function Bookings() {
    const q = useQuery({
        queryKey: ["bookings", "me"],
        queryFn: () => api("/v1/bookings/me"),
    });
    const list = q.data?.bookings ?? [];
    return (_jsxs("div", { className: "fade-up", children: [_jsxs("header", { className: a.pageHeader, children: [_jsx("span", { className: "eyebrow", children: "Reservation ledger" }), _jsxs("h1", { className: a.pageTitle, style: { fontSize: "clamp(48px,7vw,96px)" }, children: ["Your ", _jsx("em", { children: "journeys" }), "."] }), _jsx("p", { className: a.pageLede, children: "Every reservation, from pending to ticketed. Pay, cancel, or refund \u2014 all from one row." })] }), q.isLoading ? (_jsx("p", { className: a.status, children: "Retrieving ledger\u2026" })) : list.length === 0 ? (_jsxs("div", { className: a.empty, children: [_jsxs("h3", { children: ["The platform is ", _jsx("em", { children: "quiet" }), "."] }), _jsxs("p", { children: ["You haven't reserved yet. ", _jsx(Link, { to: "/", className: "bare", style: { color: "var(--burgundy)", borderBottom: "1px solid currentColor" }, children: "browse services \u2192" })] })] })) : (_jsx("div", { style: { borderTop: "1px solid var(--paper-line-strong)" }, children: list.map((b) => (_jsx(BookingRow, { b: b }, b.id))) }))] }));
}
function BookingRow({ b }) {
    const badge = b.status === "CONFIRMED" ? a.badgeOk :
        b.status === "PENDING" ? a.badgeWarn :
            b.status === "CANCELLED" ? a.badgeDanger :
                a.badgeNeutral;
    return (_jsxs(Link, { to: `/bookings/${b.id}`, className: "bare", style: {
            display: "grid",
            gridTemplateColumns: "auto 1fr auto auto auto",
            gap: 24,
            alignItems: "center",
            padding: "20px 4px",
            borderBottom: "1px solid var(--paper-line-strong)",
            transition: "background 0.15s var(--ease)",
        }, children: [_jsx("span", { className: "mono", style: { fontSize: 11, color: "var(--ink-3)", letterSpacing: "0.12em" }, children: b.id.slice(0, 8).toUpperCase() }), _jsxs("div", { children: [_jsxs("div", { style: { fontFamily: "var(--font-display)", fontWeight: 500, fontSize: 20 }, children: ["Train ", b.train_id.slice(0, 8), "\u2026 \u00B7 ", b.seat_count, " seat", b.seat_count > 1 ? "s" : ""] }), _jsxs("div", { className: "mono", style: { fontSize: 11, color: "var(--ink-3)", marginTop: 4, letterSpacing: "0.08em" }, children: ["reserved ", formatDate(b.created_at)] })] }), _jsxs("span", { className: "mono", style: { fontSize: 14, color: "var(--ink)" }, children: ["$", formatPrice(b.amount_cents)] }), _jsx("span", { className: `${a.badge} ${badge}`, children: b.status }), _jsx("span", { style: { color: "var(--burgundy)", fontFamily: "var(--font-mono)", fontSize: 11 }, children: "\u2192" })] }));
}
