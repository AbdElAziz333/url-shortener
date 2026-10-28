import { useEffect, useState } from "react"
import { Link } from "react-router-dom"
import { deleteLink, getLinks } from "../../services/linkService"
import type { LinkDto } from "../../types/linkTypes"
import { formatDate } from "../../utils/date"
import { useAuth } from "../../context/AuthContext"

const PAGE_SIZE = 10

export default function LinksPage() {
    const {logout} = useAuth()
    const [links, setLinks] = useState<LinkDto[]>([])
    const [page, setPage] = useState(1)
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        let cancelled = false // ignore stale responses when page changes / unmounts

        async function load() {
            setLoading(true)

            try {
                // const data = await getLinks(page, PAGE_SIZE)
                const data = await getLinks(PAGE_SIZE, (page - 1) * PAGE_SIZE)
                if (!cancelled) {
                    setLinks(data)
                }
            } catch (err) {
                if (!cancelled) {
                    console.error("error fetching links:", err)
                }
            } finally {
                if (!cancelled) {
                    setLoading(false)
                }
            }
        }

        load()
        return () => {
            cancelled = true
        }
    }, [page])

    async function handleDelete(code: string) {
        try {
            await deleteLink(code)
            setLinks((prev) => prev.filter((l) => l.code !== code))
        } catch (err) {
            console.error("error deleting link:", err)
        }
    }

    async function handleLogout() {
        try {
            await logout()
        } catch (err) {
            console.error("error logging out:", err)
        }
    }

    return (
        <div>
            <h1>My Links</h1>

            {loading && <p>loading...</p>}
            {!loading && links.length === 0 && <p>no links</p>}

            <ul>
                {links.map((link) => (
                    <li key={link.code}>
                        <Link to={`/link/${link.code}`}>{link.code}</Link>
                        <div>
                            original url: {" "}
                            <a href={link.original_url} target="_blank" rel="noopener noreferrer">{link.original_url}</a>
                            </div>
                        <div>expires at: {link.expires_at ? formatDate(link.expires_at) : "never"}</div>
                        <div>created at: {formatDate(link.created_at)}</div>
                        <button onClick={() => handleDelete(link.code)}>delete</button>
                    </li>
                ))}
            </ul>

            <button onClick={() => setPage((p) => p - 1)} disabled={page === 1 || loading}>
                previous
            </button>
            <span> page {page} </span>
            <button onClick={() => setPage((p) => p + 1)} disabled={links.length < PAGE_SIZE || loading}>
                next
            </button>
            <br />
            <br />
            <Link to="/link/create">Create link</Link>
            <br />
            <br />
            <button onClick={handleLogout}>Logout</button>
        </div>
    )
}