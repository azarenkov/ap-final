import { jsx as _jsx, jsxs as _jsxs, Fragment as _Fragment } from "react/jsx-runtime";
import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "../lib/api";
import { useAuth } from "../lib/auth";
import { formatDate, formatDuration, formatPrice, formatTime } from "../lib/format";
import a from "../styles/app.module.css";
export function TrainDetail() {
    const { id = "" } = useParams();
    const { user } = useAuth();
    const nav = useNavigate();
    const qc = useQueryClient();
    const [seats, setSeats] = useState(1);
    const [createdBooking, setCreatedBooking] = useState(null);
    const train = useQuery({
        queryKey: ["train", id],
        queryFn: () => api(`/v1/trains/${id}`, { auth: false }),
    });
    const route = useQuery({
        queryKey: ["route", train.data?.route_id],
        enabled: !!train.data?.route_id,
        queryFn: () => api(`/v1/routes/${train.data.route_id}`, { auth: false }),
    });
    const schedule = useQuery({
        queryKey: ["schedule", id],
        queryFn: () => api(`/v1/trains/${id}/schedule`, { auth: false }),
    });
    const book = useMutation({
        mutationFn: () => api(`/v1/bookings`, {
            method: "POST",
            body: { train_id: id, seat_count: seats },
        }),
        onSuccess: (b) => {
            setCreatedBooking(b);
            qc.invalidateQueries({ queryKey: ["train", id] });
            qc.invalidateQueries({ queryKey: ["bookings", "me"] });
        },
    });
    if (train.isLoading)
        return _jsx("p", { className: a.status, children: "Loading service\u2026" });
    if (train.error)
        return _jsx("p", { className: `${a.status} ${a.statusError}`, children: train.error.message });
    if (!train.data)
        return null;
    const tr = train.data;
    const origin = route.data?.origin ?? "—";
    const destination = route.data?.destination ?? "—";
    return (_jsxs("div", { className: "fade-up", children: [_jsxs("header", { className: a.pageHeader, children: [_jsx("div", { className: a.meta, children: _jsxs("span", { className: "eyebrow", children: [tr.code, " \u00B7 ", tr.status] }) }), _jsxs("h1", { className: a.pageTitle, style: { fontSize: "clamp(48px,7vw,96px)" }, children: [origin, " ", _jsx("em", { children: "\u2192" }), " ", destination] }), _jsxs("p", { className: a.pageLede, children: [tr.name, " \u00B7 departing ", formatDate(tr.departure_time), " at ", formatTime(tr.departure_time), " \u00B7 journey time ", formatDuration(tr.departure_time, tr.arrival_time)] })] }), _jsxs("div", { style: { display: "grid", gridTemplateColumns: "1fr 360px", gap: 32 }, children: [_jsxs("section", { className: a.card, children: [_jsxs("div", { className: a.cardHeader, children: [_jsxs("h2", { className: a.cardTitle, children: ["Itinerary ", _jsx("em", { children: "\u00B7 in transit" })] }), _jsx("span", { className: "eyebrow", children: "Schedule" })] }), _jsx("ol", { style: { listStyle: "none", padding: 0, margin: 0, display: "grid", gap: 18 }, children: schedule.data?.stops?.map((stop, i) => (_jsxs("li", { style: { display: "grid", gridTemplateColumns: "auto 1fr auto", alignItems: "baseline", gap: 18, borderBottom: "1px dashed var(--paper-line-strong)", paddingBottom: 14 }, children: [_jsx("span", { className: "mono", style: { fontSize: 12, color: "var(--ink-3)" }, children: String(i + 1).padStart(2, "0") }), _jsxs("div", { children: [_jsx("div", { style: { fontFamily: "var(--font-display)", fontSize: 24, letterSpacing: "-0.01em" }, children: stop.station === "transit" ? _jsx("em", { style: { fontStyle: "italic", color: "var(--burgundy)" }, children: "transit" }) : stop.station }), _jsxs("div", { className: "mono", style: { fontSize: 11, color: "var(--ink-3)", marginTop: 4, letterSpacing: "0.1em", textTransform: "uppercase" }, children: [stop.arrival && _jsxs(_Fragment, { children: ["arrives ", formatTime(stop.arrival)] }), stop.arrival && stop.departure && " · ", stop.departure && _jsxs(_Fragment, { children: ["departs ", formatTime(stop.departure)] })] })] }), _jsx("span", { className: "mono", style: { fontSize: 14, color: "var(--ink-2)" }, children: formatTime(stop.departure ?? stop.arrival) })] }, i))) ?? _jsx("li", { className: a.status, children: "Schedule unavailable." }) }), _jsx("hr", { className: "rule", style: { marginTop: 28 } }), _jsxs("div", { style: { marginTop: 22, display: "flex", flexWrap: "wrap", gap: 28, fontFamily: "var(--font-mono)", fontSize: 12, color: "var(--ink-2)" }, children: [_jsxs("span", { children: ["seats available \u00B7 ", _jsx("strong", { style: { color: "var(--ink)" }, children: tr.available_seats }), " / ", tr.total_seats] }), _jsxs("span", { children: ["route \u00B7 ", tr.route_id.slice(0, 8), "\u2026"] }), _jsxs("span", { children: ["distance \u00B7 ", route.data?.distance_km ?? "—", " km"] })] })] }), _jsxs("aside", { className: a.card, style: { position: "sticky", top: 24, alignSelf: "start" }, children: [_jsxs("div", { className: a.cardHeader, children: [_jsx("h2", { className: a.cardTitle, children: "Reserve" }), _jsx("span", { className: "eyebrow", children: "Booking" })] }), _jsxs("div", { style: { marginTop: 4 }, children: [_jsx("span", { className: "eyebrow", children: "fare per seat" }), _jsxs("div", { style: { fontFamily: "var(--font-display)", fontWeight: 500, fontSize: 48, letterSpacing: "-0.02em", lineHeight: 1, marginTop: 6 }, children: [_jsx("em", { style: { fontStyle: "italic", color: "var(--burgundy)" }, children: "$" }), formatPrice(tr.price_cents)] })] }), _jsxs("label", { className: a.field, style: { marginTop: 28 }, children: [_jsx("span", { className: a.label, children: "seats" }), _jsx("input", { className: a.input, type: "number", min: 1, max: Math.max(1, tr.available_seats), value: seats, onChange: (e) => setSeats(Math.max(1, Math.min(tr.available_seats, Number(e.target.value || "1")))) })] }), _jsx("hr", { className: "rule", style: { margin: "20px 0" } }), _jsxs("div", { style: { display: "flex", justifyContent: "space-between", fontFamily: "var(--font-mono)", fontSize: 13, color: "var(--ink-2)" }, children: [_jsx("span", { children: "total" }), _jsxs("strong", { style: { color: "var(--ink)" }, children: ["$", formatPrice(tr.price_cents * seats)] })] }), _jsx("div", { style: { marginTop: 20 }, children: user ? (_jsx("button", { className: `${a.btn} ${a.btnPrimary} ${a.btnBlock}`, disabled: book.isPending || tr.available_seats < seats, onClick: () => book.mutate(), children: book.isPending ? "Reserving…" : `Reserve ${seats} seat${seats > 1 ? "s" : ""}` })) : (_jsx(Link, { to: "/login", state: { from: `/trains/${id}` }, className: `${a.btn} ${a.btnPrimary} ${a.btnBlock} bare`, children: "Sign in to reserve \u2192" })) }), book.error && _jsx("div", { className: a.alert, children: book.error.message }), createdBooking && (_jsxs("div", { className: `${a.alert} ${a.alertOk}`, children: ["Reserved \u00B7 booking ", createdBooking.id.slice(0, 8), "\u2026", " ", _jsx("button", { onClick: () => nav(`/bookings/${createdBooking.id}`), className: "bare", style: { background: "transparent", border: 0, color: "inherit", textDecoration: "underline", padding: 0 }, children: "continue to payment \u2192" })] }))] })] })] }));
}
