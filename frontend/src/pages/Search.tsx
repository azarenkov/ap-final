import { FormEvent, useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { api } from "../lib/api";
import type { Route, SearchResponse } from "../lib/types";
import { TicketCard } from "../components/TicketCard";
import a from "../styles/app.module.css";

interface SearchState {
    origin: string;
    destination: string;
    after?: string;
    submitted: boolean;
}

export function Search() {
    const [state, setState] = useState<SearchState>({ origin: "", destination: "", submitted: false });
    const [form, setForm] = useState({ origin: "", destination: "", date: "" });

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        setState({
            origin: form.origin.trim(),
            destination: form.destination.trim(),
            after: form.date ? new Date(form.date).toISOString() : undefined,
            submitted: true,
        });
    }

    const query = useQuery({
        queryKey: ["search", state],
        enabled: state.submitted,
        queryFn: async () => {
            const params = new URLSearchParams();
            if (state.origin) params.set("origin", state.origin);
            if (state.destination) params.set("destination", state.destination);
            if (state.after) params.set("after", state.after);
            params.set("page", "1");
            params.set("page_size", "20");
            return api<SearchResponse>(`/v1/trains?${params}`, { auth: false });
        },
    });

    
    const introQuery = useQuery({
        queryKey: ["trains", "intro"],
        enabled: !state.submitted,
        queryFn: () => api<SearchResponse>("/v1/trains?page=1&page_size=6", { auth: false }),
    });

    const list = state.submitted ? query.data?.trains : introQuery.data?.trains;

    return (
        <>
            <header className={a.pageHeader}>
                <div className={a.meta}>
                    <span className={`eyebrow`}>Departure · Continental network</span>
                </div>
                <h1 className={a.pageTitle}>
                    The <em>art</em> of arriving
                    <br /> on time.
                </h1>
                <p className={a.pageLede}>
                    Search the continental network — overnight sleepers, intercity expresses, and the
                    occasional charter. Fares in USD, all prices include the dining-car surcharge.
                </p>
            </header>

            <form onSubmit={onSubmit} className={a.fieldset} style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr 1fr auto",
                gap: 24,
                padding: "28px 0",
                borderTop: "1px solid var(--paper-line-strong)",
                borderBottom: "3px double var(--paper-line-strong)",
            }}>
                <FormField label="from"
                    placeholder="Astana"
                    value={form.origin}
                    onChange={(v) => setForm({ ...form, origin: v })}
                />
                <FormField label="to"
                    placeholder="Almaty"
                    value={form.destination}
                    onChange={(v) => setForm({ ...form, destination: v })}
                />
                <FormField label="on or after"
                    type="datetime-local"
                    value={form.date}
                    onChange={(v) => setForm({ ...form, date: v })}
                />
                <button type="submit" className={`${a.btn} ${a.btnPrimary}`} style={{ alignSelf: "end", padding: "12px 28px" }}>
                    Find →
                </button>
            </form>

            <div style={{ marginTop: 48, display: "grid", gap: 18 }}>
                {(state.submitted && query.isLoading) || (!state.submitted && introQuery.isLoading) ? (
                    <p className={a.status}>Consulting the timetable…</p>
                ) : (state.submitted && query.error) || (!state.submitted && introQuery.error) ? (
                    <ErrorPanel
                        error={(state.submitted ? query.error : introQuery.error) as Error}
                    />
                ) : list && list.length > 0 ? (
                    <Results trains={list} origin={state.origin} destination={state.destination} />
                ) : (
                    <EmptyState submitted={state.submitted} />
                )}
            </div>
        </>
    );
}

function FormField(props: {
    label: string;
    value: string;
    placeholder?: string;
    type?: string;
    onChange: (v: string) => void;
}) {
    return (
        <label className={a.field}>
            <span className={a.label}>{props.label}</span>
            <input
                className={a.input}
                placeholder={props.placeholder}
                value={props.value}
                type={props.type ?? "text"}
                onChange={(e) => props.onChange(e.target.value)}
            />
        </label>
    );
}

function Results({ trains, origin, destination }: { trains: any[]; origin: string; destination: string }) {
    const headerOrigin = origin || "—";
    const headerDest = destination || "—";
    return (
        <>
            <p className={a.status} style={{ marginBottom: 12 }}>
                {trains.length} {trains.length === 1 ? "service" : "services"} {origin && `from ${headerOrigin.toUpperCase()}`}
                {destination && ` to ${headerDest.toUpperCase()}`}
            </p>
            {trains.map((t) => (
                <TicketCard key={t.id} train={t} origin={origin || "—"} destination={destination || "—"} />
            ))}
        </>
    );
}

function EmptyState({ submitted }: { submitted: boolean }) {
    return (
        <div className={a.empty}>
            <h3>
                {submitted ? "No services found." : <>Begin with a <em>destination</em>.</>}
            </h3>
            <p>
                {submitted
                    ? "Try a different origin, or widen the date window."
                    : "Trains may be sparse — try Astana → Almaty, or any pair seeded by the team."}
            </p>
        </div>
    );
}

function ErrorPanel({ error }: { error: Error }) {
    return (
        <div className={a.empty}>
            <h3>The timetable is silent.</h3>
            <p>{error.message}</p>
        </div>
    );
}
