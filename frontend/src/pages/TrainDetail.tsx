import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api } from "../lib/api";
import { useAuth } from "../lib/auth";
import { formatDate, formatDuration, formatPrice, formatTime } from "../lib/format";
import type { Booking, Route, Schedule, Train } from "../lib/types";
import a from "../styles/app.module.css";
import t from "../styles/ticket.module.css";

export function TrainDetail() {
    const { id = "" } = useParams();
    const { user } = useAuth();
    const nav = useNavigate();
    const qc = useQueryClient();
    const [seats, setSeats] = useState(1);
    const [createdBooking, setCreatedBooking] = useState<Booking | null>(null);

    const train = useQuery({
        queryKey: ["train", id],
        queryFn: () => api<Train>(`/v1/trains/${id}`, { auth: false }),
    });

    const route = useQuery({
        queryKey: ["route", train.data?.route_id],
        enabled: !!train.data?.route_id,
        queryFn: () => api<Route>(`/v1/routes/${train.data!.route_id}`, { auth: false }),
    });

    const schedule = useQuery({
        queryKey: ["schedule", id],
        queryFn: () => api<Schedule>(`/v1/trains/${id}/schedule`, { auth: false }),
    });

    const book = useMutation({
        mutationFn: () =>
            api<Booking>(`/v1/bookings`, {
                method: "POST",
                body: { train_id: id, seat_count: seats },
            }),
        onSuccess: (b) => {
            setCreatedBooking(b);
            qc.invalidateQueries({ queryKey: ["train", id] });
            qc.invalidateQueries({ queryKey: ["bookings", "me"] });
        },
    });

    if (train.isLoading) return <p className={a.status}>Loading service…</p>;
    if (train.error) return <p className={`${a.status} ${a.statusError}`}>{(train.error as Error).message}</p>;
    if (!train.data) return null;

    const tr = train.data;
    const origin = route.data?.origin ?? "—";
    const destination = route.data?.destination ?? "—";

    return (
        <div className="fade-up">
            <header className={a.pageHeader}>
                <div className={a.meta}>
                    <span className="eyebrow">{tr.code} · {tr.status}</span>
                </div>
                <h1 className={a.pageTitle} style={{ fontSize: "clamp(48px,7vw,96px)" }}>
                    {origin} <em>→</em> {destination}
                </h1>
                <p className={a.pageLede}>
                    {tr.name} · departing {formatDate(tr.departure_time)} at {formatTime(tr.departure_time)} ·
                    journey time {formatDuration(tr.departure_time, tr.arrival_time)}
                </p>
            </header>

            <div style={{ display: "grid", gridTemplateColumns: "1fr 360px", gap: 32 }}>
                <section className={a.card}>
                    <div className={a.cardHeader}>
                        <h2 className={a.cardTitle}>
                            Itinerary <em>· in transit</em>
                        </h2>
                        <span className="eyebrow">Schedule</span>
                    </div>
                    <ol style={{ listStyle: "none", padding: 0, margin: 0, display: "grid", gap: 18 }}>
                        {schedule.data?.stops?.map((stop, i) => (
                            <li key={i} style={{ display: "grid", gridTemplateColumns: "auto 1fr auto", alignItems: "baseline", gap: 18, borderBottom: "1px dashed var(--paper-line-strong)", paddingBottom: 14 }}>
                                <span className="mono" style={{ fontSize: 12, color: "var(--ink-3)" }}>
                                    {String(i + 1).padStart(2, "0")}
                                </span>
                                <div>
                                    <div style={{ fontFamily: "var(--font-display)", fontSize: 24, letterSpacing: "-0.01em" }}>
                                        {stop.station === "transit" ? <em style={{ fontStyle: "italic", color: "var(--burgundy)" }}>transit</em> : stop.station}
                                    </div>
                                    <div className="mono" style={{ fontSize: 11, color: "var(--ink-3)", marginTop: 4, letterSpacing: "0.1em", textTransform: "uppercase" }}>
                                        {stop.arrival && <>arrives {formatTime(stop.arrival)}</>}
                                        {stop.arrival && stop.departure && " · "}
                                        {stop.departure && <>departs {formatTime(stop.departure)}</>}
                                    </div>
                                </div>
                                <span className="mono" style={{ fontSize: 14, color: "var(--ink-2)" }}>
                                    {formatTime(stop.departure ?? stop.arrival)}
                                </span>
                            </li>
                        )) ?? <li className={a.status}>Schedule unavailable.</li>}
                    </ol>

                    <hr className="rule" style={{ marginTop: 28 }} />
                    <div style={{ marginTop: 22, display: "flex", flexWrap: "wrap", gap: 28, fontFamily: "var(--font-mono)", fontSize: 12, color: "var(--ink-2)" }}>
                        <span>seats available · <strong style={{ color: "var(--ink)" }}>{tr.available_seats}</strong> / {tr.total_seats}</span>
                        <span>route · {tr.route_id.slice(0, 8)}…</span>
                        <span>distance · {route.data?.distance_km ?? "—"} km</span>
                    </div>
                </section>

                <aside className={a.card} style={{ position: "sticky", top: 24, alignSelf: "start" }}>
                    <div className={a.cardHeader}>
                        <h2 className={a.cardTitle}>
                            Reserve
                        </h2>
                        <span className="eyebrow">Booking</span>
                    </div>

                    <div style={{ marginTop: 4 }}>
                        <span className="eyebrow">fare per seat</span>
                        <div style={{ fontFamily: "var(--font-display)", fontWeight: 500, fontSize: 48, letterSpacing: "-0.02em", lineHeight: 1, marginTop: 6 }}>
                            <em style={{ fontStyle: "italic", color: "var(--burgundy)" }}>$</em>
                            {formatPrice(tr.price_cents)}
                        </div>
                    </div>

                    <label className={a.field} style={{ marginTop: 28 }}>
                        <span className={a.label}>seats</span>
                        <input className={a.input} type="number" min={1} max={Math.max(1, tr.available_seats)} value={seats} onChange={(e) => setSeats(Math.max(1, Math.min(tr.available_seats, Number(e.target.value || "1"))))} />
                    </label>

                    <hr className="rule" style={{ margin: "20px 0" }} />

                    <div style={{ display: "flex", justifyContent: "space-between", fontFamily: "var(--font-mono)", fontSize: 13, color: "var(--ink-2)" }}>
                        <span>total</span>
                        <strong style={{ color: "var(--ink)" }}>${formatPrice(tr.price_cents * seats)}</strong>
                    </div>

                    <div style={{ marginTop: 20 }}>
                        {user ? (
                            <button
                                className={`${a.btn} ${a.btnPrimary} ${a.btnBlock}`}
                                disabled={book.isPending || tr.available_seats < seats}
                                onClick={() => book.mutate()}
                            >
                                {book.isPending ? "Reserving…" : `Reserve ${seats} seat${seats > 1 ? "s" : ""}`}
                            </button>
                        ) : (
                            <Link to="/login" state={{ from: `/trains/${id}` }} className={`${a.btn} ${a.btnPrimary} ${a.btnBlock} bare`}>
                                Sign in to reserve →
                            </Link>
                        )}
                    </div>

                    {book.error && <div className={a.alert}>{(book.error as Error).message}</div>}
                    {createdBooking && (
                        <div className={`${a.alert} ${a.alertOk}`}>
                            Reserved · booking {createdBooking.id.slice(0, 8)}…{" "}
                            <button onClick={() => nav(`/bookings/${createdBooking.id}`)} className="bare" style={{ background: "transparent", border: 0, color: "inherit", textDecoration: "underline", padding: 0 }}>
                                continue to payment →
                            </button>
                        </div>
                    )}
                </aside>
            </div>
        </div>
    );
}
