import type { LoginRequest, SignupRequest, User } from "../types/authTypes"
import { ACCESS_TOKEN_KEY, authServiceApi } from "./api"

export async function signup(request: SignupRequest): Promise<void> {
    await authServiceApi.post("/signup", request)
}

export async function login(request: LoginRequest): Promise<void> {
    const response = await authServiceApi.post("/login", request)
    const token = response.headers["access_token"]
    if (token) {
        localStorage.setItem(ACCESS_TOKEN_KEY, token)
    }
}

export async function logout(): Promise<void> {
    try {
        await authServiceApi.post("/logout")
    } finally {
        localStorage.removeItem(ACCESS_TOKEN_KEY)
    }
}

export async function getCurrentUser(): Promise<User> {
    try {
        return (await authServiceApi.get("/me")).data.data
    } catch (err) {
        // The stored token is expired or invalid, so drop it.
        localStorage.removeItem(ACCESS_TOKEN_KEY)
        throw err
    }
}

export async function refresh(): Promise<void> {
    await authServiceApi.post("/refresh")
}

export function hasStoredToken(): boolean {
    return !!localStorage.getItem(ACCESS_TOKEN_KEY)
}