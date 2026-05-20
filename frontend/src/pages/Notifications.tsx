import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "../lib/api";
import { formatDate, formatTime } from "../lib/format";
import type { Notification } from "../lib/types";
import a from "../styles/app.module.css";

interface ListResponse {
    items?: Notification[];
}

export function Notifications() {
    const qc = useQueryClient();
    const q = useQuery({
        queryKey: ["notifications"],
        queryFn: () => api<ListResponse>("/v1/notifications"),
        refetchInterval: 8_000,
    });
    const markRead = useMutation({
        mutationFn: (id: string) => api<Notification>(`/v1/notifications/${id}/read`, { method: "POST" }),
        onSuccess: () => qc.invalidateQueries({ queryKey: ["notifications"] }),
    });

    const items = q.data?.items ?? [];

    return (
        <div className="fade-up">
            <header className={a.pageHeader}>
                <span className="eyebrow">Bulletin · live updates</span>
                <h1 className={a.pageTitle} style={{ fontSize: "clamp(48px,7vw,96px)" }}>
                    The <em>bulletin</em> board.
                </h1>
                <p className={a.pageLede}>
                    Booking confirmations, payment receipts, train delays — every NATS event the protocol
                    publishes lands here, indexed by the notification service.
                </p>
            </header>

            {q.isLoading ? (
                <p className={a.status}>Refreshing the board…</p>
            ) : items.length === 0 ? (
                <div className={a.empty}>
                    <h3>The board is <em>quiet</em>.</h3>
                    <p>Make a booking, process a payment, or wait for a train-status event.</p>
                </div>
            ) : (
                <table style={{
                    width: "100%",
                    borderCollapse: "collapse",
                    fontFamily: "var(--font-mono)",
                    fontSize: 13,
                }}>
                    <thead>
                        <tr style={{ borderBottom: "1px solid var(--paper-line-strong)" }}>
                            <Th>Time</Th>
                            <Th>Kind</Th>
                            <Th>Subject / body</Th>
                            <Th>Status</Th>
                        </tr>
                    </thead>
                    <tbody>
                        {items.map((n) => (
                            <tr key={n.id} style={{
                                borderBottom: "1px solid var(--paper-line-strong)",
                                opacity: n.read ? 0.55 : 1,
                            }}>
                                <Td>
                                    <span style={{ color: "var(--ink-2)" }}>{formatDate(n.created_at)}</span>{" "}
                                    <span style={{ color: "var(--ink)" }}>{formatTime(n.created_at)}</span>
                                </Td>
                                <Td>
                                    <span style={{ letterSpacing: "0.1em", color: "var(--burgundy)" }}>
                                        {n.kind.replaceAll("_", " ").toLowerCase()}
                                    </span>
                                </Td>
                                <Td>
                                    <div style={{ fontFamily: "var(--font-display)", fontSize: 17, color: "var(--ink)" }}>{n.subject}</div>
                                    <div style={{ fontFamily: "var(--font-body)", fontSize: 13, color: "var(--ink-2)", marginTop: 2 }}>{n.body}</div>
                                </Td>
                                <Td>
                                    {n.read ? (
                                        <span style={{ color: "var(--ink-3)" }}>read</span>
                                    ) : (
                                        <button
                                            onClick={() => markRead.mutate(n.id)}
                                            className="bare"
                                            style={{
                                                background: "transparent",
                                                border: "1px solid var(--burgundy)",
                                                color: "var(--burgundy)",
                                                padding: "4px 10px",
                                                cursor: "pointer",
                                                fontFamily: "var(--font-mono)",
                                                fontSize: 10,
                                                letterSpacing: "0.12em",
                                                textTransform: "uppercase",
                                            }}
                                        >
                                            mark read
                                        </button>
                                    )}
                                </Td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            )}
        </div>
    );
}

const Th: React.FC<{ children: React.ReactNode }> = ({ children }) => (
    <th style={{
        textAlign: "left",
        padding: "12px 16px 12px 0",
        fontFamily: "var(--font-mono)",
        fontSize: 10.5,
        fontWeight: 500,
        letterSpacing: "0.16em",
        textTransform: "uppercase",
        color: "var(--ink-3)",
    }}>
        {children}
    </th>
);

const Td: React.FC<{ children: React.ReactNode }> = ({ children }) => (
    <td style={{ padding: "16px 16px 16px 0", verticalAlign: "top" }}>{children}</td>
);
