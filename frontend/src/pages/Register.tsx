import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";

import { useAuth } from "../lib/auth";
import a from "../styles/app.module.css";

export function Register() {
    const { register, loading } = useAuth();
    const nav = useNavigate();
    const [fullName, setFullName] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState<string | null>(null);

    async function onSubmit(e: FormEvent) {
        e.preventDefault();
        setError(null);
        try {
            await register(email, password, fullName);
            nav("/", { replace: true });
        } catch (err) {
            setError(err instanceof Error ? err.message : "Registration failed");
        }
    }

    return (
        <div style={{ maxWidth: 480, margin: "48px auto" }}>
            <header className={a.pageHeader} style={{ marginBottom: 40 }}>
                <span className="eyebrow">New traveller · subscription required</span>
                <h1 className={a.pageTitle} style={{ fontSize: "clamp(48px,6vw,80px)" }}>
                    Request a <em>passport</em>.
                </h1>
                <p className={a.pageLede}>Eight letters minimum. Your reservations follow your account.</p>
            </header>
            <form onSubmit={onSubmit} className={a.fieldset}>
                <label className={a.field}>
                    <span className={a.label}>full name</span>
                    <input className={a.input} value={fullName} onChange={(e) => setFullName(e.target.value)} required autoFocus />
                </label>
                <label className={a.field}>
                    <span className={a.label}>email</span>
                    <input className={a.input} type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
                </label>
                <label className={a.field}>
                    <span className={a.label}>password</span>
                    <input className={a.input} type="password" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} />
                </label>
                <button type="submit" className={`${a.btn} ${a.btnPrimary} ${a.btnBlock}`} disabled={loading}>
                    {loading ? "Issuing pass…" : "Open account →"}
                </button>
                {error && <div className={a.alert}>{error}</div>}
                <p style={{ fontFamily: "var(--font-mono)", fontSize: 11, letterSpacing: "0.12em", color: "var(--ink-2)", textTransform: "uppercase", textAlign: "center", marginTop: 12 }}>
                    already have one? <Link to="/login" className="bare" style={{ color: "var(--burgundy)", borderBottom: "1px solid currentColor" }}>sign in</Link>
                </p>
            </form>
        </div>
    );
}
