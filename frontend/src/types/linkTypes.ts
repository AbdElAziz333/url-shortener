export interface LinkDto {
    code: string
    original_url: string
    expires_at?: string
    is_active: boolean
    created_at: string
}
 
export interface CreateLinkRequest {
    original_url: string
    expires_at?: string
}
 
export interface UpdateExpiryRequest {
    expires_at: string
}