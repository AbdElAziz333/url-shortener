import { useState, type FormEvent } from "react"
import { toISO } from "../utils/date"

interface UpdateExpiryProps {
    // Resolves to true when the update succeeded.
    onSubmit: (expires_at: string) => Promise<boolean>
}

export default function UpdateExpiry({ onSubmit }: UpdateExpiryProps) {
    const [open, setOpen] = useState(false)
    const [expiry, setExpiry] = useState("")
    const [submitting, setSubmitting] = useState(false)

    function close() {
        setOpen(false)
        setExpiry("")
    }

    async function handleSubmit(e: FormEvent) {
        e.preventDefault()
        const expires_at = toISO(expiry)
        if (!expires_at) return

        setSubmitting(true)
        const ok = await onSubmit(expires_at)
        setSubmitting(false)
        if (ok) close()
    }

    if (!open) {
        return <button onClick={() => setOpen(true)}>update expiry</button>
    }

    return (
        <form onSubmit={handleSubmit}>
            <input
                type="date"
                value={expiry}
                onChange={(e) => setExpiry(e.target.value)}
                required
            />
            <button type="submit" disabled={submitting}>save</button>
            <button type="button" onClick={close} disabled={submitting}>cancel</button>
        </form>
    )
}