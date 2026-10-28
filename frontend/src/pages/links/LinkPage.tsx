import { useEffect, useState } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { deleteLink, getLink, updateExpiry } from "../../services/linkService"
import { formatDate } from "../../utils/date"
import UpdateExpiry from "../../components/UpdateExpiry"
import type { LinkDto } from "../../types/linkTypes"

const SHORT_URL_BASE = "http://localhost:8082"

export default function LinkPage() {
    const { code } = useParams<{ code: string }>()
    const navigate = useNavigate()

    const [link, setLink] = useState<LinkDto | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState("")

    useEffect(() => {
        if (!code) return
        let cancelled = false

        async function load(code: string) {
            setLoading(true)
            setError("")

            try {
                const data = await getLink(code)
                if (!cancelled) setLink(data)
            } catch (err) {
                console.error("error fetching link:", err)
                if (!cancelled) setError("Failed to load link")
            } finally {
                if (!cancelled) setLoading(false)
            }
        }

        load(code)
        return () => {
            cancelled = true
        }
    }, [code])

    async function handleUpdateExpiry(expires_at: string): Promise<boolean> {
        if (!code) return false

        setError("")
        try {
            setLink(await updateExpiry(code, { expires_at }))
            return true
        } catch (err) {
            console.error("error updating expiry:", err)
            setError("Failed to update expiry")
            return false
        }
    }

    async function handleDelete() {
        if (!code) return

        setError("")
        try {
            await deleteLink(code)
            navigate("/links")
        } catch (err) {
            console.error("error deleting link:", err)
            setError("Failed to delete link")
        }
    }

    if (loading) return <p>loading...</p>

    const shortUrl = link ? `${SHORT_URL_BASE}/${link.code}` : ""

    return (
        <div>
            <h1>Link</h1>

            {error && <p role="alert">{error}</p>}

            {link && (
                <>
                    <p>
                        shortened url:{" "}
                        <a href={shortUrl} target="_blank" rel="noopener noreferrer">{shortUrl}</a>
                    </p>
                    <p>
                        original url: {" "}
                        <a href={link.original_url} target="_blank" rel="noopener noreferrer">{link.original_url}</a>
                    </p>
                    <p>active: {String(link.is_active)}</p>
                    <p>created at: {formatDate(link.created_at)}</p>
                    <p>expires at: {link.expires_at ? formatDate(link.expires_at) : "never"}</p>

                    <UpdateExpiry onSubmit={handleUpdateExpiry} />

                    <button onClick={handleDelete}>delete</button>
                </>
            )}

            <br />
            <br />
            <Link to="/links">My Links</Link>
            <br />
            <Link to="/home">Home Page</Link>
        </div>
    )
}