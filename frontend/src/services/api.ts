import axios from "axios"

// why this is undefined?
const API_GATEWAY_URL = import.meta.env.VITE_API_GATEWAY_URL

export const ACCESS_TOKEN_KEY = "access_token"

function createApi(path: string) {
    const instance = axios.create({
        baseURL: `${API_GATEWAY_URL}${path}`,
        withCredentials: true,
    })

    instance.interceptors.request.use((config) => {
        const token = localStorage.getItem(ACCESS_TOKEN_KEY)
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    })

    return instance
}

export const authServiceApi = createApi("/auth")
export const linkServiceApi = createApi("/shortener/api/links")