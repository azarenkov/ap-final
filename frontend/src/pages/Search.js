import { jsx as _jsx, jsxs as _jsxs, Fragment as _Fragment } from "react/jsx-runtime";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "../lib/api";
import { TicketCard } from "../components/TicketCard";
import a from "../styles/app.module.css";
export function Search() {
    const [state, setState] = useState({ origin: "", destination: "", submitted: false });
    const [form, setForm] = useState({ origin: "", destination: "", date: "" });
    function onSubmit(e) {
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
            if (state.origin)
                params.set("origin", state.origin);
            if (state.destination)
                params.set("destination", state.destination);
            if (state.after)
                params.set("after", state.after);
            params.set("page", "1");
            params.set("page_size", "20");
            return api(`/v1/trains?${params}`, { auth: false });
        },
    });
    
    const introQuery = useQuery({
        queryKey: ["trains", "intro"],
        enabled: !state.submitted,
        queryFn: () => api("/v1/trains?page=1&page_size=6", { auth: false }),
    });
    const list = state.submitted ? query.data?.trains : introQuery.data?.trains;
    return (_jsxs(_Fragment, { children: [_jsxs("header", { className: a.pageHeader, children: [_jsx("div", { className: a.meta, children: _jsx("span", { className: `eyebrow`, children: "Departure \u00B7 Continental network" }) }), _jsxs("h1", { className: a.pageTitle, children: ["The ", _jsx("em", { children: "art" }), " of arriving", _jsx("br", {}), " on time."] }), _jsx("p", { className: a.pageLede, children: "Search the continental network \u2014 overnight sleepers, intercity expresses, and the occasional charter. Fares in USD, all prices include the dining-car surcharge." })] }), _jsxs("form", { onSubmit: onSubmit, className: a.fieldset, style: {
                    display: "grid",
                    gridTemplateColumns: "1fr 1fr 1fr auto",
                    gap: 24,
                    padding: "28px 0",
                    borderTop: "1px solid var(--paper-line-strong)",
                    borderBottom: "3px double var(--paper-line-strong)",
                }, children: [_jsx(FormField, { label: "from", placeholder: "Astana", value: form.origin, onChange: (v) => setForm({ ...form, origin: v }) }), _jsx(FormField, { label: "to", placeholder: "Almaty", value: form.destination, onChange: (v) => setForm({ ...form, destination: v }) }), _jsx(FormField, { label: "on or after", type: "datetime-local", value: form.date, onChange: (v) => setForm({ ...form, date: v }) }), _jsx("button", { type: "submit", className: `${a.btn} ${a.btnPrimary}`, style: { alignSelf: "end", padding: "12px 28px" }, children: "Find \u2192" })] }), _jsx("div", { style: { marginTop: 48, display: "grid", gap: 18 }, children: (state.submitted && query.isLoading) || (!state.submitted && introQuery.isLoading) ? (_jsx("p", { className: a.status, children: "Consulting the timetable\u2026" })) : (state.submitted && query.error) || (!state.submitted && introQuery.error) ? (_jsx(ErrorPanel, { error: (state.submitted ? query.error : introQuery.error) })) : list && list.length > 0 ? (_jsx(Results, { trains: list, origin: state.origin, destination: state.destination })) : (_jsx(EmptyState, { submitted: state.submitted })) })] }));
}
function FormField(props) {
    return (_jsxs("label", { className: a.field, children: [_jsx("span", { className: a.label, children: props.label }), _jsx("input", { className: a.input, placeholder: props.placeholder, value: props.value, type: props.type ?? "text", onChange: (e) => props.onChange(e.target.value) })] }));
}
function Results({ trains, origin, destination }) {
    const headerOrigin = origin || "—";
    const headerDest = destination || "—";
    return (_jsxs(_Fragment, { children: [_jsxs("p", { className: a.status, style: { marginBottom: 12 }, children: [trains.length, " ", trains.length === 1 ? "service" : "services", " ", origin && `from ${headerOrigin.toUpperCase()}`, destination && ` to ${headerDest.toUpperCase()}`] }), trains.map((t) => (_jsx(TicketCard, { train: t, origin: origin || "—", destination: destination || "—" }, t.id)))] }));
}
function EmptyState({ submitted }) {
    return (_jsxs("div", { className: a.empty, children: [_jsx("h3", { children: submitted ? "No services found." : _jsxs(_Fragment, { children: ["Begin with a ", _jsx("em", { children: "destination" }), "."] }) }), _jsx("p", { children: submitted
                    ? "Try a different origin, or widen the date window."
                    : "Trains may be sparse — try Astana → Almaty, or any pair seeded by the team." })] }));
}
function ErrorPanel({ error }) {
    return (_jsxs("div", { className: a.empty, children: [_jsx("h3", { children: "The timetable is silent." }), _jsx("p", { children: error.message })] }));
}
