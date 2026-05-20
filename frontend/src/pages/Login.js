import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../lib/auth";
import a from "../styles/app.module.css";
export function Login() {
    const { login, loading } = useAuth();
    const nav = useNavigate();
    const loc = useLocation();
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState(null);
    async function onSubmit(e) {
        e.preventDefault();
        setError(null);
        try {
            await login(email, password);
            const from = loc.state?.from ?? "/";
            nav(from, { replace: true });
        }
        catch (err) {
            setError(err instanceof Error ? err.message : "Sign-in failed");
        }
    }
    return (_jsxs("div", { style: { maxWidth: 480, margin: "48px auto" }, children: [_jsxs("header", { className: a.pageHeader, style: { marginBottom: 40 }, children: [_jsx("span", { className: "eyebrow", children: "Wagon-Lits \u00B7 Authorised travellers only" }), _jsxs("h1", { className: a.pageTitle, style: { fontSize: "clamp(48px,6vw,80px)" }, children: ["Welcome ", _jsx("em", { children: "aboard" }), "."] }), _jsx("p", { className: a.pageLede, children: "Sign in to retrieve your reservations and bulletin." })] }), _jsxs("form", { onSubmit: onSubmit, className: a.fieldset, children: [_jsxs("label", { className: a.field, children: [_jsx("span", { className: a.label, children: "email" }), _jsx("input", { className: a.input, type: "email", value: email, onChange: (e) => setEmail(e.target.value), required: true, autoFocus: true })] }), _jsxs("label", { className: a.field, children: [_jsx("span", { className: a.label, children: "password" }), _jsx("input", { className: a.input, type: "password", value: password, onChange: (e) => setPassword(e.target.value), required: true, minLength: 8 })] }), _jsx("button", { type: "submit", className: `${a.btn} ${a.btnPrimary} ${a.btnBlock}`, disabled: loading, children: loading ? "Signing in…" : "Sign in →" }), error && _jsx("div", { className: a.alert, children: error }), _jsxs("p", { style: { fontFamily: "var(--font-mono)", fontSize: 11, letterSpacing: "0.12em", color: "var(--ink-2)", textTransform: "uppercase", textAlign: "center", marginTop: 12 }, children: ["no account? ", _jsx(Link, { to: "/register", className: "bare", style: { color: "var(--burgundy)", borderBottom: "1px solid currentColor" }, children: "request one" })] })] })] }));
}
