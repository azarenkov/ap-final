import { Link, NavLink, Navigate, Route, Routes, useLocation } from "react-router-dom";
import { useEffect, useState } from "react";

import { useAuth } from "./lib/auth";
import { Search } from "./pages/Search";
import { TrainDetail } from "./pages/TrainDetail";
import { Login } from "./pages/Login";
import { Register } from "./pages/Register";
import { Bookings } from "./pages/Bookings";
import { BookingDetail } from "./pages/BookingDetail";
import { Notifications } from "./pages/Notifications";
import { Admin } from "./pages/Admin";
import s from "./styles/app.module.css";

function Header() {
    const { user, logout } = useAuth();
    const [time, setTime] = useState(() => new Date());
    useEffect(() => {
        const id = setInterval(() => setTime(new Date()), 1000);
        return () => clearInterval(id);
    }, []);

    return (
        <header className={s.header}>
            <DepartureTicker />
            <div className={s.headerInner}>
                <Link to="/" className={`${s.brand} bare`}>
                    <span className={s.name}>
                        Compagnie <em>Internationale</em> des Trains
                    </span>
                    <span className={s.tag}>Service Continental · est. mcmxxiv</span>
                </Link>
                <nav className={s.nav}>
                    <NavLink end to="/" className={({ isActive }) => `${s.navLink} bare ${isActive ? s.navLinkActive : ""}`}>
                        Search
                    </NavLink>
                    <NavLink to="/bookings" className={({ isActive }) => `${s.navLink} bare ${isActive ? s.navLinkActive : ""}`}>
                        Bookings
                    </NavLink>
                    <NavLink to="/notifications" className={({ isActive }) => `${s.navLink} bare ${isActive ? s.navLinkActive : ""}`}>
                        Bulletin
                    </NavLink>
                    {user && (
                        <NavLink to="/admin" className={({ isActive }) => `${s.navLink} bare ${isActive ? s.navLinkActive : ""}`}>
                            Admin
                        </NavLink>
                    )}
                </nav>
                <div className={s.headerRight}>
                    <span className={s.clock}>
                        {time.toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit", hour12: false })} UTC
                    </span>
                    {user ? (
                        <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
                            <span className={s.userPill}>{user.email}</span>
                            <button onClick={logout} className={`${s.btn} ${s.btnGhost}`} style={{ padding: "6px 12px", fontSize: 10 }}>
                                Log out
                            </button>
                        </div>
                    ) : (
                        <Link to="/login" className={`${s.btn}`} style={{ padding: "6px 12px", fontSize: 10 }}>
                            Sign in
                        </Link>
                    )}
                </div>
            </div>
        </header>
    );
}

function DepartureTicker() {
    const items = [
        "Astana → Almaty · Talgo · 08·15 · platform 3",
        "Atyrau → Aktobe · Tulpar · 09·40 · platform 1",
        "Shymkent → Kyzylorda · Express · 11·00 · platform 5",
        "Almaty → Bishkek · Sapsan · 13·22 · platform 2",
        "Karaganda → Astana · Spanish · 17·15 · platform 4",
    ];
    return (
        <div className={s.tickerBar}>
            <div className={s.ticker}>
                {[...items, ...items].map((t, i) => (
                    <span key={i} className={s.tickerItem}>
                        {t}
                    </span>
                ))}
            </div>
        </div>
    );
}

function Footer() {
    const loc = useLocation();
    return (
        <footer className={s.footer}>
            <div className={s.footerInner}>
                <span>Compagnie Internationale des Trains · Capstone</span>
                <span>{loc.pathname.replace("/", "") || "search"}</span>
                <span>Built with gRPC · NATS · Postgres · Redis</span>
            </div>
        </footer>
    );
}

function Protected({ children }: { children: JSX.Element }) {
    const { user } = useAuth();
    const loc = useLocation();
    if (!user) return <Navigate to="/login" state={{ from: loc.pathname }} replace />;
    return children;
}

export function App() {
    return (
        <div className={s.shell}>
            <Header />
            <main className={s.main}>
                <Routes>
                    <Route path="/" element={<Search />} />
                    <Route path="/login" element={<Login />} />
                    <Route path="/register" element={<Register />} />
                    <Route path="/trains/:id" element={<TrainDetail />} />
                    <Route path="/bookings" element={<Protected><Bookings /></Protected>} />
                    <Route path="/bookings/:id" element={<Protected><BookingDetail /></Protected>} />
                    <Route path="/notifications" element={<Protected><Notifications /></Protected>} />
                    <Route path="/admin" element={<Protected><Admin /></Protected>} />
                </Routes>
            </main>
            <Footer />
        </div>
    );
}
