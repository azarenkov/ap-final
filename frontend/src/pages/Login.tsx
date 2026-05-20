import { FormEvent, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";

import { useAuth } from "../lib/auth";
import a from "../styles/app.module.css";

export function Login() {
    const { login, loading } = useAuth();
    const nav = useNavigate();
    const loc = useLocation();
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState<string | null>(null);

    async function onSubmit(e: FormEvent) {
        e.preventDefault();
        setError(null);
        try {
            await login(email, password);
            const from = (loc.state as { from?: string } | null)?.from ?? "/";
            nav(from, { replace: true });
        } catch (err) {
            setError(err instanceof Error ? err.message : "Sign-in failed");
        }
    }

    return (
        <div style={{ maxWidth: 480, margin: "48px auto" }}>
            <header className={a.pageHeader} style={{ marginBottom: 40 }}>
                <span className="eyebrow">Wagon-Lits · Authorised travellers only</span>
                <h1 className={a.pageTitle} style={{ fontSize: "clamp(48px,6vw,80px)" }}>
                    Welcome <em>aboard</em>.
                </h1>
                <p className={a.pageLede}>Sign in to retrieve your reservations and bulletin.</p>
            </header>
            <form onSubmit={onSubmit} className={a.fieldset}>
                <label className={a.field}>
                    <span className={a.label}>email</span>
                    <input className={a.input} type="email" value={email} onChange={(e) => setEmail(e.target.value)} required autoFocus />
                </label>
                <label className={a.field}>
                    <span className={a.label}>password</span>
                    <input className={a.input} type="password" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} />
                </label>
                <button type="submit" className={`${a.btn} ${a.btnPrimary} ${a.btnBlock}`} disabled={loading}>
                    {loading ? "Signing in…" : "Sign in →"}
                </button>
                {error && <div className={a.alert}>{error}</div>}
                <p style={{ fontFamily: "var(--font-mono)", fontSize: 11, letterSpacing: "0.12em", color: "var(--ink-2)", textTransform: "uppercase", textAlign: "center", marginTop: 12 }}>
                    no account? <Link to="/register" className="bare" style={{ color: "var(--burgundy)", borderBottom: "1px solid currentColor" }}>request one</Link>
                </p>
            </form>
        </div>
    );
}
