import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
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
import s from "./styles/app.module.css";
function Header() {
    const { user, logout } = useAuth();
    const [time, setTime] = useState(() => new Date());
    useEffect(() => {
        const id = setInterval(() => setTime(new Date()), 1000);
        return () => clearInterval(id);
    }, []);
    return (_jsxs("header", { className: s.header, children: [_jsx(DepartureTicker, {}), _jsxs("div", { className: s.headerInner, children: [_jsxs(Link, { to: "/", className: `${s.brand} bare`, children: [_jsxs("span", { className: s.name, children: ["Compagnie ", _jsx("em", { children: "Internationale" }), " des Trains"] }), _jsx("span", { className: s.tag, children: "Service Continental \u00B7 est. mcmxxiv" })] }), _jsxs("nav", { className: s.nav, children: [_jsx(NavLink, { end: true, to: "/", className: ({ isActive }) => `${s.navLink} bare ${isActive ? s.navLinkActive : ""}`, children: "Search" }), _jsx(NavLink, { to: "/bookings", className: ({ isActive }) => `${s.navLink} bare ${isActive ? s.navLinkActive : ""}`, children: "Bookings" }), _jsx(NavLink, { to: "/notifications", className: ({ isActive }) => `${s.navLink} bare ${isActive ? s.navLinkActive : ""}`, children: "Bulletin" })] }), _jsxs("div", { className: s.headerRight, children: [_jsxs("span", { className: s.clock, children: [time.toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit", hour12: false }), " UTC"] }), user ? (_jsxs("div", { style: { display: "flex", gap: 10, alignItems: "center" }, children: [_jsx("span", { className: s.userPill, children: user.email }), _jsx("button", { onClick: logout, className: `${s.btn} ${s.btnGhost}`, style: { padding: "6px 12px", fontSize: 10 }, children: "Log out" })] })) : (_jsx(Link, { to: "/login", className: `${s.btn}`, style: { padding: "6px 12px", fontSize: 10 }, children: "Sign in" }))] })] })] }));
}
function DepartureTicker() {
    const items = [
        "Astana → Almaty · Talgo · 08·15 · platform 3",
        "Atyrau → Aktobe · Tulpar · 09·40 · platform 1",
        "Shymkent → Kyzylorda · Express · 11·00 · platform 5",
        "Almaty → Bishkek · Sapsan · 13·22 · platform 2",
        "Karaganda → Astana · Spanish · 17·15 · platform 4",
    ];
    return (_jsx("div", { className: s.tickerBar, children: _jsx("div", { className: s.ticker, children: [...items, ...items].map((t, i) => (_jsx("span", { className: s.tickerItem, children: t }, i))) }) }));
}
function Footer() {
    const loc = useLocation();
    return (_jsx("footer", { className: s.footer, children: _jsxs("div", { className: s.footerInner, children: [_jsx("span", { children: "Compagnie Internationale des Trains \u00B7 Capstone" }), _jsx("span", { children: loc.pathname.replace("/", "") || "search" }), _jsx("span", { children: "Built with gRPC \u00B7 NATS \u00B7 Postgres \u00B7 Redis" })] }) }));
}
function Protected({ children }) {
    const { user } = useAuth();
    const loc = useLocation();
    if (!user)
        return _jsx(Navigate, { to: "/login", state: { from: loc.pathname }, replace: true });
    return children;
}
export function App() {
    return (_jsxs("div", { className: s.shell, children: [_jsx(Header, {}), _jsx("main", { className: s.main, children: _jsxs(Routes, { children: [_jsx(Route, { path: "/", element: _jsx(Search, {}) }), _jsx(Route, { path: "/login", element: _jsx(Login, {}) }), _jsx(Route, { path: "/register", element: _jsx(Register, {}) }), _jsx(Route, { path: "/trains/:id", element: _jsx(TrainDetail, {}) }), _jsx(Route, { path: "/bookings", element: _jsx(Protected, { children: _jsx(Bookings, {}) }) }), _jsx(Route, { path: "/bookings/:id", element: _jsx(Protected, { children: _jsx(BookingDetail, {}) }) }), _jsx(Route, { path: "/notifications", element: _jsx(Protected, { children: _jsx(Notifications, {}) }) })] }) }), _jsx(Footer, {})] }));
}
