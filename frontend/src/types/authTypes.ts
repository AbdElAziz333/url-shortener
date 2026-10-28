export interface SignupRequest {
    email: string
    password: string
}

export interface LoginRequest {
    email: string
    password: string
}

export interface User {
    id: string
    email: string
    isActive: boolean
    createdAt: string
    updatedAt: string
}