import { jsx as _jsx } from "react/jsx-runtime";
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { api, setAuthToken } from "./api";
const Ctx = createContext(null);
const TOKEN_KEY = "ap2.token";
const USER_KEY = "ap2.user";
export function AuthProvider({ children }) {
    const [user, setUser] = useState(() => {
        const raw = window.localStorage.getItem(USER_KEY);
        return raw ? JSON.parse(raw) : null;
    });
    const [token, setToken] = useState(() => window.localStorage.getItem(TOKEN_KEY));
    const [loading, setLoading] = useState(false);
    useEffect(() => {
        setAuthToken(token);
    }, [token]);
    const persist = useCallback((res) => {
        if (res) {
            window.localStorage.setItem(TOKEN_KEY, res.access_token);
            window.localStorage.setItem(USER_KEY, JSON.stringify(res.user));
            setToken(res.access_token);
            setUser(res.user);
        }
        else {
            window.localStorage.removeItem(TOKEN_KEY);
            window.localStorage.removeItem(USER_KEY);
            setToken(null);
            setUser(null);
        }
    }, []);
    const login = useCallback(async (email, password) => {
        setLoading(true);
        try {
            const res = await api("/v1/users/login", {
                method: "POST",
                body: { email, password },
                auth: false,
            });
            persist(res);
        }
        finally {
            setLoading(false);
        }
    }, [persist]);
    const register = useCallback(async (email, password, fullName) => {
        setLoading(true);
        try {
            await api("/v1/users/register", {
                method: "POST",
                body: { email, password, full_name: fullName },
                auth: false,
            });
            
            const res = await api("/v1/users/login", {
                method: "POST",
                body: { email, password },
                auth: false,
            });
            persist(res);
        }
        finally {
            setLoading(false);
        }
    }, [persist]);
    const logout = useCallback(() => persist(null), [persist]);
    const value = useMemo(() => ({ user, token, loading, login, register, logout }), [user, token, loading, login, register, logout]);
    return _jsx(Ctx.Provider, { value: value, children: children });
}
export function useAuth() {
    const v = useContext(Ctx);
    if (!v)
        throw new Error("useAuth outside AuthProvider");
    return v;
}
