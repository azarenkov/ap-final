import { FormEvent, useMemo, useState } from "react";
import { useMutation, useQueries, useQuery, useQueryClient } from "@tanstack/react-query";

import { api } from "../lib/api";
import { formatDate, formatPrice, formatTime } from "../lib/format";
import type { Route, SearchResponse, Train } from "../lib/types";
import a from "../styles/app.module.css";

export function Admin() {
    
    
    const [localRoutes, setLocalRoutes] = useState<Route[]>([]);

    const trains = useQuery({
        queryKey: ["admin", "trains"],
        queryFn: () => api<SearchResponse>("/v1/trains?page=1&page_size=100", { auth: false }),
        refetchInterval: 15_000,
    });

    const uniqueRouteIds = useMemo(() => {
        const set = new Set<string>();
        for (const t of trains.data?.trains ?? []) set.add(t.route_id);
        for (const r of localRoutes) set.add(r.id);
        return Array.from(set);
    }, [trains.data, localRoutes]);

    const routeQueries = useQueries({
        queries: uniqueRouteIds.map((id) => ({
            queryKey: ["route", id],
            queryFn: () => api<Route>(`/v1/routes/${id}`, { auth: false }),
            staleTime: 5 * 60_000,
        })),
    });

    const routes = useMemo<Route[]>(() => {
        const byId = new Map<string, Route>();
        for (const r of localRoutes) byId.set(r.id, r);
        for (const q of routeQueries) {
            if (q.data) byId.set(q.data.id, q.data);
        }
        return Array.from(byId.values()).sort((a, b) =>
            (a.origin + a.destination).localeCompare(b.origin + b.destination),
        );
    }, [localRoutes, routeQueries]);

    return (
        <div className="fade-up">
            <header className={a.pageHeader}>
                <span className="eyebrow">Stationmaster · administration</span>
                <h1 className={a.pageTitle} style={{ fontSize: "clamp(48px,7vw,96px)" }}>
                    The <em>dispatcher's</em> desk.
                </h1>
                <p className={a.pageLede}>
                    Manage routes, post new services, mark delays, cancel runs. Every change here is a
                    privileged write through the API gateway — the audience sees it on the next refresh.
                </p>
            </header>

            <div style={{ display: "grid", gridTemplateColumns: "1fr 1.2fr", gap: 32 }}>
                <RoutesPanel routes={routes} onCreated={(r) => setLocalRoutes((prev) => [r, ...prev])} />
                <TrainsPanel routes={routes} trains={trains.data?.trains ?? []} loading={trains.isLoading} />
            </div>
        </div>
    );
}

function RoutesPanel({ routes, onCreated }: { routes: Route[]; onCreated: (r: Route) => void }) {
    const [form, setForm] = useState({ origin: "", destination: "", distance_km: 500, estimated_minutes: 480 });

    const create = useMutation({
        mutationFn: () =>
            api<Route>("/v1/routes", {
                method: "POST",
                body: {
                    origin: form.origin,
                    destination: form.destination,
                    distance_km: Number(form.distance_km),
                    estimated_minutes: Number(form.estimated_minutes),
                },
            }),
        onSuccess: (r) => {
            onCreated(r);
            setForm({ origin: "", destination: "", distance_km: 500, estimated_minutes: 480 });
        },
    });

    const onSubmit = (e: FormEvent) => {
        e.preventDefault();
        if (form.origin && form.destination) create.mutate();
    };

    return (
        <section className={a.card}>
            <div className={a.cardHeader}>
                <h2 className={a.cardTitle}>
                    Routes <em>· network</em>
                </h2>
                <span className="eyebrow">{routes.length} known</span>
            </div>

            <form onSubmit={onSubmit} className={a.fieldset}>
                <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 }}>
                    <label className={a.field}>
                        <span className={a.label}>from</span>
                        <input className={a.input} value={form.origin} onChange={(e) => setForm({ ...form, origin: e.target.value })} required />
                    </label>
                    <label className={a.field}>
                        <span className={a.label}>to</span>
                        <input className={a.input} value={form.destination} onChange={(e) => setForm({ ...form, destination: e.target.value })} required />
                    </label>
                </div>
                <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 }}>
                    <label className={a.field}>
                        <span className={a.label}>distance (km)</span>
                        <input className={a.input} type="number" min={1} value={form.distance_km} onChange={(e) => setForm({ ...form, distance_km: Number(e.target.value) })} />
                    </label>
                    <label className={a.field}>
                        <span className={a.label}>est. minutes</span>
                        <input className={a.input} type="number" min={1} value={form.estimated_minutes} onChange={(e) => setForm({ ...form, estimated_minutes: Number(e.target.value) })} />
                    </label>
                </div>
                <button type="submit" className={`${a.btn} ${a.btnPrimary}`} disabled={create.isPending}>
                    {create.isPending ? "Plotting…" : "Add route"}
                </button>
                {create.error && <div className={a.alert}>{(create.error as Error).message}</div>}
            </form>

            <hr className="rule" style={{ margin: "24px 0 16px" }} />
            <span className="eyebrow">Network</span>
            {routes.length === 0 ? (
                <p className={a.status} style={{ marginTop: 12 }}>No routes yet. Create one above.</p>
            ) : (
                <ul style={{ listStyle: "none", padding: 0, marginTop: 12 }}>
                    {routes.map((r) => (
                        <li key={r.id} style={{ display: "flex", justifyContent: "space-between", padding: "10px 0", borderBottom: "1px dashed var(--paper-line-strong)" }}>
                            <div style={{ fontFamily: "var(--font-display)", fontSize: 18 }}>
                                {r.origin} <em style={{ fontStyle: "italic", color: "var(--burgundy)" }}>→</em> {r.destination}
                            </div>
                            <div className="mono" style={{ fontSize: 11, color: "var(--ink-3)" }}>{r.distance_km} km · {r.estimated_minutes}m</div>
                        </li>
                    ))}
                </ul>
            )}
        </section>
    );
}

function TrainsPanel({ routes, trains, loading }: { routes: Route[]; trains: Train[]; loading: boolean }) {
    const qc = useQueryClient();

    return (
        <section className={a.card}>
            <div className={a.cardHeader}>
                <h2 className={a.cardTitle}>
                    Trains <em>· timetable</em>
                </h2>
                <span className="eyebrow">{trains.length} on the board</span>
            </div>

            <CreateTrainForm routes={routes} onCreated={() => qc.invalidateQueries({ queryKey: ["admin", "trains"] })} />

            <hr className="rule" style={{ margin: "24px 0 12px" }} />

            {loading ? (
                <p className={a.status}>Loading services…</p>
            ) : trains.length === 0 ? (
                <p className={a.status}>No services. Add a route, then post a service.</p>
            ) : (
                <ul style={{ listStyle: "none", padding: 0, margin: 0 }}>
                    {trains.map((t) => (
                        <TrainRow
                            key={t.id}
                            train={t}
                            route={routes.find((r) => r.id === t.route_id)}
                            onChange={() => qc.invalidateQueries({ queryKey: ["admin", "trains"] })}
                        />
                    ))}
                </ul>
            )}
        </section>
    );
}

function CreateTrainForm({ routes, onCreated }: { routes: Route[]; onCreated: () => void }) {
    const [open, setOpen] = useState(false);
    const [form, setForm] = useState({
        code: "",
        name: "",
        route_id: "",
        departure_time: "",
        arrival_time: "",
        total_seats: 200,
        price_cents: 1_000_000,
    });

    const create = useMutation({
        mutationFn: () =>
            api<Train>("/v1/trains", {
                method: "POST",
                body: {
                    code: form.code,
                    name: form.name,
                    route_id: form.route_id,
                    departure_time: new Date(form.departure_time).toISOString(),
                    arrival_time: new Date(form.arrival_time).toISOString(),
                    total_seats: Number(form.total_seats),
                    price_cents: Number(form.price_cents),
                },
            }),
        onSuccess: () => {
            onCreated();
            setOpen(false);
            setForm({ code: "", name: "", route_id: "", departure_time: "", arrival_time: "", total_seats: 200, price_cents: 1_000_000 });
        },
    });

    if (!open) {
        return (
            <button className={`${a.btn} ${a.btnGhost}`} onClick={() => setOpen(true)} disabled={routes.length === 0}>
                {routes.length === 0 ? "+ new service (add a route first)" : "+ new service"}
            </button>
        );
    }

    return (
        <form onSubmit={(e) => { e.preventDefault(); create.mutate(); }} className={a.fieldset}>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 2fr", gap: 14 }}>
                <label className={a.field}>
                    <span className={a.label}>code</span>
                    <input className={a.input} value={form.code} onChange={(e) => setForm({ ...form, code: e.target.value })} placeholder="IC-999" required />
                </label>
                <label className={a.field}>
                    <span className={a.label}>name</span>
                    <input className={a.input} value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="Talgo Express" required />
                </label>
            </div>
            <label className={a.field}>
                <span className={a.label}>route</span>
                <select
                    className={a.select}
                    value={form.route_id}
                    onChange={(e) => setForm({ ...form, route_id: e.target.value })}
                    required
                >
                    <option value="" disabled>— pick a route —</option>
                    {routes.map((r) => (
                        <option key={r.id} value={r.id}>
                            {r.origin} → {r.destination}  ({r.distance_km} km)
                        </option>
                    ))}
                </select>
            </label>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14 }}>
                <label className={a.field}>
                    <span className={a.label}>departure</span>
                    <input className={a.input} type="datetime-local" value={form.departure_time} onChange={(e) => setForm({ ...form, departure_time: e.target.value })} required />
                </label>
                <label className={a.field}>
                    <span className={a.label}>arrival</span>
                    <input className={a.input} type="datetime-local" value={form.arrival_time} onChange={(e) => setForm({ ...form, arrival_time: e.target.value })} required />
                </label>
            </div>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14 }}>
                <label className={a.field}>
                    <span className={a.label}>seats</span>
                    <input className={a.input} type="number" min={1} value={form.total_seats} onChange={(e) => setForm({ ...form, total_seats: Number(e.target.value) })} />
                </label>
                <label className={a.field}>
                    <span className={a.label}>price (cents)</span>
                    <input className={a.input} type="number" min={1} value={form.price_cents} onChange={(e) => setForm({ ...form, price_cents: Number(e.target.value) })} />
                </label>
            </div>
            <div className={a.btnGroup}>
                <button type="submit" className={`${a.btn} ${a.btnPrimary}`} disabled={create.isPending}>
                    {create.isPending ? "Posting…" : "Post service"}
                </button>
                <button type="button" className={`${a.btn} ${a.btnGhost}`} onClick={() => setOpen(false)}>cancel</button>
            </div>
            {create.error && <div className={a.alert}>{(create.error as Error).message}</div>}
        </form>
    );
}

function TrainRow({ train, route, onChange }: { train: Train; route?: Route; onChange: () => void }) {
    const status = train.status;
    const badge = useMemo(() => {
        if (status === "SCHEDULED") return a.badgeOk;
        if (status === "DELAYED") return a.badgeWarn;
        if (status === "CANCELLED") return a.badgeDanger;
        return a.badgeNeutral;
    }, [status]);

    const setStatus = useMutation({
        mutationFn: (newStatus: string) =>
            api<Train>(`/v1/trains/${train.id}`, {
                method: "PATCH",
                body: { status: newStatus },
            }),
        onSuccess: onChange,
    });

    const del = useMutation({
        mutationFn: () => api<unknown>(`/v1/trains/${train.id}`, { method: "DELETE" }),
        onSuccess: onChange,
    });

    return (
        <li style={{
            display: "grid",
            gridTemplateColumns: "auto 1fr auto auto",
            gap: 16,
            alignItems: "center",
            padding: "14px 0",
            borderBottom: "1px dashed var(--paper-line-strong)",
        }}>
            <span className="mono" style={{ fontSize: 11, color: "var(--ink-3)", letterSpacing: "0.12em" }}>
                {train.code}
            </span>
            <div>
                <div style={{ fontFamily: "var(--font-display)", fontSize: 17 }}>{train.name}</div>
                <div className="mono" style={{ fontSize: 10.5, color: "var(--ink-3)", marginTop: 2, letterSpacing: "0.08em" }}>
                    {route ? `${route.origin} → ${route.destination} · ` : ""}
                    {formatDate(train.departure_time)} · {formatTime(train.departure_time)} → {formatTime(train.arrival_time)} ·
                    ${formatPrice(train.price_cents)} · {train.available_seats}/{train.total_seats}
                </div>
            </div>
            <span className={`${a.badge} ${badge}`}>{status}</span>
            <div style={{ display: "flex", gap: 6 }}>
                <button
                    className={`${a.btn} ${a.btnGhost}`}
                    style={{ padding: "5px 10px", fontSize: 9 }}
                    disabled={setStatus.isPending || status === "DELAYED"}
                    onClick={() => setStatus.mutate("DELAYED")}
                    title="Mark delayed (publishes train.delayed)"
                >
                    Delay
                </button>
                <button
                    className={`${a.btn} ${a.btnGhost}`}
                    style={{ padding: "5px 10px", fontSize: 9, color: "var(--danger)", borderColor: "var(--danger)" }}
                    disabled={setStatus.isPending || status === "CANCELLED"}
                    onClick={() => setStatus.mutate("CANCELLED")}
                    title="Cancel (publishes train.cancelled)"
                >
                    Cancel
                </button>
                <button
                    className={`${a.btn} ${a.btnGhost}`}
                    style={{ padding: "5px 10px", fontSize: 9 }}
                    disabled={del.isPending}
                    onClick={() => {
                        if (confirm(`Delete train ${train.code}?`)) del.mutate();
                    }}
                    title="Hard delete"
                >
                    ✕
                </button>
            </div>
        </li>
    );
}
