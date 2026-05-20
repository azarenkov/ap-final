import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../lib/auth";
import a from "../styles/app.module.css";
export function Register() {
    const { register, loading } = useAuth();
    const nav = useNavigate();
    const [fullName, setFullName] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState(null);
    async function onSubmit(e) {
        e.preventDefault();
        setError(null);
        try {
            await register(email, password, fullName);
            nav("/", { replace: true });
        }
        catch (err) {
            setError(err instanceof Error ? err.message : "Registration failed");
        }
    }
    return (_jsxs("div", { style: { maxWidth: 480, margin: "48px auto" }, children: [_jsxs("header", { className: a.pageHeader, style: { marginBottom: 40 }, children: [_jsx("span", { className: "eyebrow", children: "New traveller \u00B7 subscription required" }), _jsxs("h1", { className: a.pageTitle, style: { fontSize: "clamp(48px,6vw,80px)" }, children: ["Request a ", _jsx("em", { children: "passport" }), "."] }), _jsx("p", { className: a.pageLede, children: "Eight letters minimum. Your reservations follow your account." })] }), _jsxs("form", { onSubmit: onSubmit, className: a.fieldset, children: [_jsxs("label", { className: a.field, children: [_jsx("span", { className: a.label, children: "full name" }), _jsx("input", { className: a.input, value: fullName, onChange: (e) => setFullName(e.target.value), required: true, autoFocus: true })] }), _jsxs("label", { className: a.field, children: [_jsx("span", { className: a.label, children: "email" }), _jsx("input", { className: a.input, type: "email", value: email, onChange: (e) => setEmail(e.target.value), required: true })] }), _jsxs("label", { className: a.field, children: [_jsx("span", { className: a.label, children: "password" }), _jsx("input", { className: a.input, type: "password", value: password, onChange: (e) => setPassword(e.target.value), required: true, minLength: 8 })] }), _jsx("button", { type: "submit", className: `${a.btn} ${a.btnPrimary} ${a.btnBlock}`, disabled: loading, children: loading ? "Issuing pass…" : "Open account →" }), error && _jsx("div", { className: a.alert, children: error }), _jsxs("p", { style: { fontFamily: "var(--font-mono)", fontSize: 11, letterSpacing: "0.12em", color: "var(--ink-2)", textTransform: "uppercase", textAlign: "center", marginTop: 12 }, children: ["already have one? ", _jsx(Link, { to: "/login", className: "bare", style: { color: "var(--burgundy)", borderBottom: "1px solid currentColor" }, children: "sign in" })] })] })] }));
}
