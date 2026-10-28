import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { getCurrentUser, hasStoredToken, login as loginService, logout as logoutService } from "../services/authService";
import type { User, LoginRequest } from "../types/authTypes";

interface AuthContextType {
    user: User | null
    loading: boolean
    isAuthenticated: boolean
    login: (credentials: LoginRequest) => Promise<void>
    logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null)
    const [loading, setLoading] = useState(true)

    // Restore session on initial mount, only if there's a token to restore from
    useEffect(() => {
        if (!hasStoredToken()) {
            setLoading(false)
            return
        }

        getCurrentUser()
            .then(setUser)
            .catch(() => setUser(null))
            .finally(() => setLoading(false))
    }, [])

    const login = useCallback(async (credentials: LoginRequest) => {
        await loginService(credentials)
        setUser(await getCurrentUser())
    }, [])

    const logout = useCallback(async () => {
        try {
            await logoutService()
        } finally {
            setUser(null)
        }
    }, [])

    const value = useMemo(
        () => ({ user, loading, isAuthenticated: !!user, login, logout }),
        [user, loading, login, logout],
    )

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error("useAuth must be used within an AuthProvider");
    }
    return context;
}