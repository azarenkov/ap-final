import { createContext, ReactNode, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { api, setAuthToken } from "./api";

export interface User {
    id: string;
    email: string;
    full_name: string;
    verified: boolean;
}

interface LoginResponse {
    access_token: string;
    expires_at: string;
    user: User;
}

interface AuthCtx {
    user: User | null;
    token: string | null;
    loading: boolean;
    login: (email: string, password: string) => Promise<void>;
    register: (email: string, password: string, fullName: string) => Promise<void>;
    logout: () => void;
}

const Ctx = createContext<AuthCtx | null>(null);

const TOKEN_KEY = "ap2.token";
const USER_KEY = "ap2.user";

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(() => {
        const raw = window.localStorage.getItem(USER_KEY);
        return raw ? JSON.parse(raw) : null;
    });
    const [token, setToken] = useState<string | null>(() => window.localStorage.getItem(TOKEN_KEY));
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        setAuthToken(token);
    }, [token]);

    const persist = useCallback((res: LoginResponse | null) => {
        if (res) {
            window.localStorage.setItem(TOKEN_KEY, res.access_token);
            window.localStorage.setItem(USER_KEY, JSON.stringify(res.user));
            setToken(res.access_token);
            setUser(res.user);
        } else {
            window.localStorage.removeItem(TOKEN_KEY);
            window.localStorage.removeItem(USER_KEY);
            setToken(null);
            setUser(null);
        }
    }, []);

    const login = useCallback(async (email: string, password: string) => {
        setLoading(true);
        try {
            const res = await api<LoginResponse>("/v1/users/login", {
                method: "POST",
                body: { email, password },
                auth: false,
            });
            persist(res);
        } finally {
            setLoading(false);
        }
    }, [persist]);

    const register = useCallback(async (email: string, password: string, fullName: string) => {
        setLoading(true);
        try {
            await api<User>("/v1/users/register", {
                method: "POST",
                body: { email, password, full_name: fullName },
                auth: false,
            });
            
            const res = await api<LoginResponse>("/v1/users/login", {
                method: "POST",
                body: { email, password },
                auth: false,
            });
            persist(res);
        } finally {
            setLoading(false);
        }
    }, [persist]);

    const logout = useCallback(() => persist(null), [persist]);

    const value = useMemo<AuthCtx>(
        () => ({ user, token, loading, login, register, logout }),
        [user, token, loading, login, register, logout],
    );
    return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useAuth() {
    const v = useContext(Ctx);
    if (!v) throw new Error("useAuth outside AuthProvider");
    return v;
}
