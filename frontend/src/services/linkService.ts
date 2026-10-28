import type { LinkDto, CreateLinkRequest, UpdateExpiryRequest } from "../types/linkTypes";
import { linkServiceApi } from "./api";

export async function getLinks(limit: number, offset: number): Promise<LinkDto[]> {
    return (await linkServiceApi.get("", { params: { limit, offset } })).data.data
}

export async function getLink(code: string): Promise<LinkDto> {
    return (await linkServiceApi.get(code)).data.data
}

export async function createLink(request: CreateLinkRequest): Promise<LinkDto> {
    return (await linkServiceApi.post("", request)).data.data
}

export async function updateExpiry(code: string, request: UpdateExpiryRequest): Promise<LinkDto> {
    return (await linkServiceApi.patch(`${code}/expiry`, request)).data.data
}

export async function deleteLink(code: string): Promise<void> {
    await linkServiceApi.delete(code)
}