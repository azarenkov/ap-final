import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "../lib/api";
import { formatDate, formatTime } from "../lib/format";
import a from "../styles/app.module.css";
export function Notifications() {
    const qc = useQueryClient();
    const q = useQuery({
        queryKey: ["notifications"],
        queryFn: () => api("/v1/notifications"),
        refetchInterval: 8_000,
    });
    const markRead = useMutation({
        mutationFn: (id) => api(`/v1/notifications/${id}/read`, { method: "POST" }),
        onSuccess: () => qc.invalidateQueries({ queryKey: ["notifications"] }),
    });
    const items = q.data?.items ?? [];
    return (_jsxs("div", { className: "fade-up", children: [_jsxs("header", { className: a.pageHeader, children: [_jsx("span", { className: "eyebrow", children: "Bulletin \u00B7 live updates" }), _jsxs("h1", { className: a.pageTitle, style: { fontSize: "clamp(48px,7vw,96px)" }, children: ["The ", _jsx("em", { children: "bulletin" }), " board."] }), _jsx("p", { className: a.pageLede, children: "Booking confirmations, payment receipts, train delays \u2014 every NATS event the protocol publishes lands here, indexed by the notification service." })] }), q.isLoading ? (_jsx("p", { className: a.status, children: "Refreshing the board\u2026" })) : items.length === 0 ? (_jsxs("div", { className: a.empty, children: [_jsxs("h3", { children: ["The board is ", _jsx("em", { children: "quiet" }), "."] }), _jsx("p", { children: "Make a booking, process a payment, or wait for a train-status event." })] })) : (_jsxs("table", { style: {
                    width: "100%",
                    borderCollapse: "collapse",
                    fontFamily: "var(--font-mono)",
                    fontSize: 13,
                }, children: [_jsx("thead", { children: _jsxs("tr", { style: { borderBottom: "1px solid var(--paper-line-strong)" }, children: [_jsx(Th, { children: "Time" }), _jsx(Th, { children: "Kind" }), _jsx(Th, { children: "Subject / body" }), _jsx(Th, { children: "Status" })] }) }), _jsx("tbody", { children: items.map((n) => (_jsxs("tr", { style: {
                                borderBottom: "1px solid var(--paper-line-strong)",
                                opacity: n.read ? 0.55 : 1,
                            }, children: [_jsxs(Td, { children: [_jsx("span", { style: { color: "var(--ink-2)" }, children: formatDate(n.created_at) }), " ", _jsx("span", { style: { color: "var(--ink)" }, children: formatTime(n.created_at) })] }), _jsx(Td, { children: _jsx("span", { style: { letterSpacing: "0.1em", color: "var(--burgundy)" }, children: n.kind.replaceAll("_", " ").toLowerCase() }) }), _jsxs(Td, { children: [_jsx("div", { style: { fontFamily: "var(--font-display)", fontSize: 17, color: "var(--ink)" }, children: n.subject }), _jsx("div", { style: { fontFamily: "var(--font-body)", fontSize: 13, color: "var(--ink-2)", marginTop: 2 }, children: n.body })] }), _jsx(Td, { children: n.read ? (_jsx("span", { style: { color: "var(--ink-3)" }, children: "read" })) : (_jsx("button", { onClick: () => markRead.mutate(n.id), className: "bare", style: {
                                            background: "transparent",
                                            border: "1px solid var(--burgundy)",
                                            color: "var(--burgundy)",
                                            padding: "4px 10px",
                                            cursor: "pointer",
                                            fontFamily: "var(--font-mono)",
                                            fontSize: 10,
                                            letterSpacing: "0.12em",
                                            textTransform: "uppercase",
                                        }, children: "mark read" })) })] }, n.id))) })] }))] }));
}
const Th = ({ children }) => (_jsx("th", { style: {
        textAlign: "left",
        padding: "12px 16px 12px 0",
        fontFamily: "var(--font-mono)",
        fontSize: 10.5,
        fontWeight: 500,
        letterSpacing: "0.16em",
        textTransform: "uppercase",
        color: "var(--ink-3)",
    }, children: children }));
const Td = ({ children }) => (_jsx("td", { style: { padding: "16px 16px 16px 0", verticalAlign: "top" }, children: children }));
