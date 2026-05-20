import { useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import { api } from "../lib/api";
import { formatDate, formatPrice, formatTime } from "../lib/format";
import type { Booking, PaymentResult, Ticket } from "../lib/types";
import a from "../styles/app.module.css";

export function BookingDetail() {
    const { id = "" } = useParams();
    const qc = useQueryClient();
    const [ticket, setTicket] = useState<Ticket | null>(null);
    const [payment, setPayment] = useState<PaymentResult | null>(null);

    const b = useQuery({
        queryKey: ["booking", id],
        queryFn: () => api<Booking>(`/v1/bookings/${id}`),
    });

    const pay = useMutation({
        mutationFn: () => api<PaymentResult>(`/v1/bookings/${id}/pay`, { method: "POST" }),
        onSuccess: (res) => {
            setPayment(res);
            qc.invalidateQueries({ queryKey: ["booking", id] });
        },
    });

    const cancel = useMutation({
        mutationFn: () => api<Booking>(`/v1/bookings/${id}/cancel`, { method: "POST" }),
        onSuccess: () => qc.invalidateQueries({ queryKey: ["booking", id] }),
    });

    const confirm = useMutation({
        mutationFn: () => api<Booking>(`/v1/bookings/${id}/confirm`, { method: "POST" }),
        onSuccess: () => qc.invalidateQueries({ queryKey: ["booking", id] }),
    });

    const ticketMut = useMutation({
        mutationFn: () => api<Ticket>(`/v1/bookings/${id}/ticket`, { method: "POST" }),
        onSuccess: (t) => setTicket(t),
    });

    const refund = useMutation({
        mutationFn: () => api<{ booking_id: string; amount_cents: number }>(`/v1/bookings/${id}/refund`, { method: "POST" }),
        onSuccess: () => qc.invalidateQueries({ queryKey: ["booking", id] }),
    });

    if (b.isLoading) return <p className={a.status}>Loading reservation…</p>;
    if (b.error) return <div className={a.empty}><h3>Not found.</h3><p>{(b.error as Error).message}</p></div>;
    if (!b.data) return null;
    const data = b.data;

    const badge =
        data.status === "CONFIRMED" ? a.badgeOk :
        data.status === "PENDING"   ? a.badgeWarn :
        data.status === "CANCELLED" ? a.badgeDanger :
                                       a.badgeNeutral;

    return (
        <div className="fade-up">
            <header className={a.pageHeader}>
                <div className={a.meta}>
                    <span className="eyebrow">Booking · {data.id.slice(0, 12).toUpperCase()}</span>
                    <span className={`${a.badge} ${badge}`}>{data.status}</span>
                </div>
                <h1 className={a.pageTitle} style={{ fontSize: "clamp(48px,7vw,96px)" }}>
                    {data.seat_count} <em>seat{data.seat_count > 1 ? "s" : ""}</em>
                </h1>
                <p className={a.pageLede}>
                    Reserved {formatDate(data.created_at)} at {formatTime(data.created_at)} ·
                    train <span className="mono">{data.train_id.slice(0, 8)}…</span>
                </p>
            </header>

            <div style={{ display: "grid", gridTemplateColumns: "1fr 360px", gap: 32 }}>
                <section className={a.card}>
                    <div className={a.cardHeader}>
                        <h2 className={a.cardTitle}>
                            Actions <em>· next steps</em>
                        </h2>
                        <span className="eyebrow">Workflow</span>
                    </div>

                    <ol style={{ listStyle: "none", padding: 0, margin: 0, display: "grid", gap: 20 }}>
                        <Step n="01" title="Process payment" done={data.status !== "PENDING"} disabled={data.status !== "PENDING"}>
                            <p style={{ color: "var(--ink-2)", margin: "0 0 12px" }}>
                                Mock processor — 80% success rate. The booking moves to <em>CONFIRMED</em> on a positive outcome.
                            </p>
                            <button className={`${a.btn} ${a.btnPrimary}`} disabled={pay.isPending || data.status !== "PENDING"} onClick={() => pay.mutate()}>
                                {pay.isPending ? "Processing…" : "Pay now"}
                            </button>
                            {payment && (
                                <div className={`${a.alert} ${payment.success ? a.alertOk : ""}`}>
                                    {payment.success ? "✓" : "✗"} {payment.message}
                                </div>
                            )}
                        </Step>

                        <Step n="02" title="Generate ticket" done={!!ticket} disabled={data.status !== "CONFIRMED"}>
                            <p style={{ color: "var(--ink-2)", margin: "0 0 12px" }}>
                                Once payment is confirmed, the box office prints the boarding pass.
                            </p>
                            <button className={a.btn} disabled={ticketMut.isPending || data.status !== "CONFIRMED"} onClick={() => ticketMut.mutate()}>
                                {ticketMut.isPending ? "Printing…" : "Generate ticket"}
                            </button>
                            {ticket && (
                                <div style={{ marginTop: 14, padding: 18, background: "var(--paper-3)", border: "1px dashed var(--paper-line-strong)" }}>
                                    <div className="eyebrow">Boarding pass</div>
                                    <div style={{ fontFamily: "var(--font-display)", fontSize: 30, marginTop: 4 }}>
                                        {ticket.code}
                                    </div>
                                    <div className="mono" style={{ fontSize: 11, color: "var(--ink-3)", marginTop: 6 }}>
                                        issued {formatTime(ticket.issued_at)} · ticket {ticket.id.slice(0, 8)}…
                                    </div>
                                </div>
                            )}
                        </Step>

                        <Step n="03" title="Manage" done={false} disabled={data.status === "REFUNDED"}>
                            <p style={{ color: "var(--ink-2)", margin: "0 0 12px" }}>
                                Cancel a pending reservation, or request a refund on a confirmed one.
                            </p>
                            <div className={a.btnGroup}>
                                <button className={a.btn} disabled={cancel.isPending || (data.status !== "PENDING" && data.status !== "CONFIRMED")} onClick={() => cancel.mutate()}>
                                    {cancel.isPending ? "Cancelling…" : "Cancel"}
                                </button>
                                <button className={a.btn} disabled={confirm.isPending || data.status !== "PENDING"} onClick={() => confirm.mutate()}>
                                    {confirm.isPending ? "Confirming…" : "Confirm manually"}
                                </button>
                                <button className={a.btn} disabled={refund.isPending || (data.status !== "CONFIRMED" && data.status !== "CANCELLED")} onClick={() => refund.mutate()}>
                                    {refund.isPending ? "Refunding…" : "Refund"}
                                </button>
                            </div>
                            {(cancel.error || confirm.error || refund.error) && (
                                <div className={a.alert}>
                                    {(cancel.error || confirm.error || refund.error) instanceof Error
                                        ? ((cancel.error || confirm.error || refund.error) as Error).message
                                        : "Action failed"}
                                </div>
                            )}
                        </Step>
                    </ol>
                </section>

                <aside className={a.card} style={{ alignSelf: "start" }}>
                    <div className={a.cardHeader}>
                        <h2 className={a.cardTitle}>
                            Receipt
                        </h2>
                        <span className="eyebrow">Stub</span>
                    </div>
                    <SummaryRow k="Booking" v={data.id.slice(0, 12).toUpperCase()} />
                    <SummaryRow k="Train" v={data.train_id.slice(0, 12).toUpperCase()} />
                    <SummaryRow k="Seats" v={String(data.seat_count)} />
                    <SummaryRow k="Status" v={data.status} />
                    <hr className="rule" style={{ margin: "16px 0" }} />
                    <SummaryRow k="Total" v={`$${formatPrice(data.amount_cents)}`} emphasis />
                </aside>
            </div>
        </div>
    );
}

function Step({ n, title, children, done, disabled }: { n: string; title: string; children: React.ReactNode; done: boolean; disabled?: boolean }) {
    return (
        <li style={{
            display: "grid",
            gridTemplateColumns: "48px 1fr",
            gap: 22,
            paddingBottom: 22,
            borderBottom: "1px dashed var(--paper-line-strong)",
            opacity: disabled ? 0.55 : 1,
        }}>
            <span className="mono" style={{
                fontSize: 11,
                color: done ? "var(--sage)" : "var(--ink-3)",
                letterSpacing: "0.16em",
                paddingTop: 6,
            }}>
                {done ? "DONE" : n}
            </span>
            <div>
                <div style={{ fontFamily: "var(--font-display)", fontWeight: 500, fontSize: 22, marginBottom: 6 }}>
                    {title}
                </div>
                {children}
            </div>
        </li>
    );
}

function SummaryRow({ k, v, emphasis }: { k: string; v: string; emphasis?: boolean }) {
    return (
        <div style={{
            display: "flex",
            justifyContent: "space-between",
            padding: "8px 0",
            fontFamily: "var(--font-mono)",
            fontSize: emphasis ? 16 : 12,
            color: emphasis ? "var(--ink)" : "var(--ink-2)",
        }}>
            <span style={{ letterSpacing: "0.12em", textTransform: "uppercase", fontSize: 10.5 }}>{k}</span>
            <span style={{ color: emphasis ? "var(--burgundy)" : "var(--ink)" }}>{v}</span>
        </div>
    );
}
