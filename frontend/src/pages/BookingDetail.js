import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { api } from "../lib/api";
import { formatDate, formatPrice, formatTime } from "../lib/format";
import a from "../styles/app.module.css";
export function BookingDetail() {
    const { id = "" } = useParams();
    const qc = useQueryClient();
    const [ticket, setTicket] = useState(null);
    const [payment, setPayment] = useState(null);
    const b = useQuery({
        queryKey: ["booking", id],
        queryFn: () => api(`/v1/bookings/${id}`),
    });
    const pay = useMutation({
        mutationFn: () => api(`/v1/bookings/${id}/pay`, { method: "POST" }),
        onSuccess: (res) => {
            setPayment(res);
            qc.invalidateQueries({ queryKey: ["booking", id] });
        },
    });
    const cancel = useMutation({
        mutationFn: () => api(`/v1/bookings/${id}/cancel`, { method: "POST" }),
        onSuccess: () => qc.invalidateQueries({ queryKey: ["booking", id] }),
    });
    const confirm = useMutation({
        mutationFn: () => api(`/v1/bookings/${id}/confirm`, { method: "POST" }),
        onSuccess: () => qc.invalidateQueries({ queryKey: ["booking", id] }),
    });
    const ticketMut = useMutation({
        mutationFn: () => api(`/v1/bookings/${id}/ticket`, { method: "POST" }),
        onSuccess: (t) => setTicket(t),
    });
    const refund = useMutation({
        mutationFn: () => api(`/v1/bookings/${id}/refund`, { method: "POST" }),
        onSuccess: () => qc.invalidateQueries({ queryKey: ["booking", id] }),
    });
    if (b.isLoading)
        return _jsx("p", { className: a.status, children: "Loading reservation\u2026" });
    if (b.error)
        return _jsxs("div", { className: a.empty, children: [_jsx("h3", { children: "Not found." }), _jsx("p", { children: b.error.message })] });
    if (!b.data)
        return null;
    const data = b.data;
    const badge = data.status === "CONFIRMED" ? a.badgeOk :
        data.status === "PENDING" ? a.badgeWarn :
            data.status === "CANCELLED" ? a.badgeDanger :
                a.badgeNeutral;
    return (_jsxs("div", { className: "fade-up", children: [_jsxs("header", { className: a.pageHeader, children: [_jsxs("div", { className: a.meta, children: [_jsxs("span", { className: "eyebrow", children: ["Booking \u00B7 ", data.id.slice(0, 12).toUpperCase()] }), _jsx("span", { className: `${a.badge} ${badge}`, children: data.status })] }), _jsxs("h1", { className: a.pageTitle, style: { fontSize: "clamp(48px,7vw,96px)" }, children: [data.seat_count, " ", _jsxs("em", { children: ["seat", data.seat_count > 1 ? "s" : ""] })] }), _jsxs("p", { className: a.pageLede, children: ["Reserved ", formatDate(data.created_at), " at ", formatTime(data.created_at), " \u00B7 train ", _jsxs("span", { className: "mono", children: [data.train_id.slice(0, 8), "\u2026"] })] })] }), _jsxs("div", { style: { display: "grid", gridTemplateColumns: "1fr 360px", gap: 32 }, children: [_jsxs("section", { className: a.card, children: [_jsxs("div", { className: a.cardHeader, children: [_jsxs("h2", { className: a.cardTitle, children: ["Actions ", _jsx("em", { children: "\u00B7 next steps" })] }), _jsx("span", { className: "eyebrow", children: "Workflow" })] }), _jsxs("ol", { style: { listStyle: "none", padding: 0, margin: 0, display: "grid", gap: 20 }, children: [_jsxs(Step, { n: "01", title: "Process payment", done: data.status !== "PENDING", disabled: data.status !== "PENDING", children: [_jsxs("p", { style: { color: "var(--ink-2)", margin: "0 0 12px" }, children: ["Mock processor \u2014 80% success rate. The booking moves to ", _jsx("em", { children: "CONFIRMED" }), " on a positive outcome."] }), _jsx("button", { className: `${a.btn} ${a.btnPrimary}`, disabled: pay.isPending || data.status !== "PENDING", onClick: () => pay.mutate(), children: pay.isPending ? "Processing…" : "Pay now" }), payment && (_jsxs("div", { className: `${a.alert} ${payment.success ? a.alertOk : ""}`, children: [payment.success ? "✓" : "✗", " ", payment.message] }))] }), _jsxs(Step, { n: "02", title: "Generate ticket", done: !!ticket, disabled: data.status !== "CONFIRMED", children: [_jsx("p", { style: { color: "var(--ink-2)", margin: "0 0 12px" }, children: "Once payment is confirmed, the box office prints the boarding pass." }), _jsx("button", { className: a.btn, disabled: ticketMut.isPending || data.status !== "CONFIRMED", onClick: () => ticketMut.mutate(), children: ticketMut.isPending ? "Printing…" : "Generate ticket" }), ticket && (_jsxs("div", { style: { marginTop: 14, padding: 18, background: "var(--paper-3)", border: "1px dashed var(--paper-line-strong)" }, children: [_jsx("div", { className: "eyebrow", children: "Boarding pass" }), _jsx("div", { style: { fontFamily: "var(--font-display)", fontSize: 30, marginTop: 4 }, children: ticket.code }), _jsxs("div", { className: "mono", style: { fontSize: 11, color: "var(--ink-3)", marginTop: 6 }, children: ["issued ", formatTime(ticket.issued_at), " \u00B7 ticket ", ticket.id.slice(0, 8), "\u2026"] })] }))] }), _jsxs(Step, { n: "03", title: "Manage", done: false, disabled: data.status === "REFUNDED", children: [_jsx("p", { style: { color: "var(--ink-2)", margin: "0 0 12px" }, children: "Cancel a pending reservation, or request a refund on a confirmed one." }), _jsxs("div", { className: a.btnGroup, children: [_jsx("button", { className: a.btn, disabled: cancel.isPending || (data.status !== "PENDING" && data.status !== "CONFIRMED"), onClick: () => cancel.mutate(), children: cancel.isPending ? "Cancelling…" : "Cancel" }), _jsx("button", { className: a.btn, disabled: confirm.isPending || data.status !== "PENDING", onClick: () => confirm.mutate(), children: confirm.isPending ? "Confirming…" : "Confirm manually" }), _jsx("button", { className: a.btn, disabled: refund.isPending || (data.status !== "CONFIRMED" && data.status !== "CANCELLED"), onClick: () => refund.mutate(), children: refund.isPending ? "Refunding…" : "Refund" })] }), (cancel.error || confirm.error || refund.error) && (_jsx("div", { className: a.alert, children: (cancel.error || confirm.error || refund.error) instanceof Error
                                                    ? (cancel.error || confirm.error || refund.error).message
                                                    : "Action failed" }))] })] })] }), _jsxs("aside", { className: a.card, style: { alignSelf: "start" }, children: [_jsxs("div", { className: a.cardHeader, children: [_jsx("h2", { className: a.cardTitle, children: "Receipt" }), _jsx("span", { className: "eyebrow", children: "Stub" })] }), _jsx(SummaryRow, { k: "Booking", v: data.id.slice(0, 12).toUpperCase() }), _jsx(SummaryRow, { k: "Train", v: data.train_id.slice(0, 12).toUpperCase() }), _jsx(SummaryRow, { k: "Seats", v: String(data.seat_count) }), _jsx(SummaryRow, { k: "Status", v: data.status }), _jsx("hr", { className: "rule", style: { margin: "16px 0" } }), _jsx(SummaryRow, { k: "Total", v: `$${formatPrice(data.amount_cents)}`, emphasis: true })] })] })] }));
}
function Step({ n, title, children, done, disabled }) {
    return (_jsxs("li", { style: {
            display: "grid",
            gridTemplateColumns: "48px 1fr",
            gap: 22,
            paddingBottom: 22,
            borderBottom: "1px dashed var(--paper-line-strong)",
            opacity: disabled ? 0.55 : 1,
        }, children: [_jsx("span", { className: "mono", style: {
                    fontSize: 11,
                    color: done ? "var(--sage)" : "var(--ink-3)",
                    letterSpacing: "0.16em",
                    paddingTop: 6,
                }, children: done ? "DONE" : n }), _jsxs("div", { children: [_jsx("div", { style: { fontFamily: "var(--font-display)", fontWeight: 500, fontSize: 22, marginBottom: 6 }, children: title }), children] })] }));
}
function SummaryRow({ k, v, emphasis }) {
    return (_jsxs("div", { style: {
            display: "flex",
            justifyContent: "space-between",
            padding: "8px 0",
            fontFamily: "var(--font-mono)",
            fontSize: emphasis ? 16 : 12,
            color: emphasis ? "var(--ink)" : "var(--ink-2)",
        }, children: [_jsx("span", { style: { letterSpacing: "0.12em", textTransform: "uppercase", fontSize: 10.5 }, children: k }), _jsx("span", { style: { color: emphasis ? "var(--burgundy)" : "var(--ink)" }, children: v })] }));
}
