import { Link } from "react-router-dom"

export default function HomePage() {
    return (
        <div>
            <h1>Home Page</h1>
            <nav>
                <ul>
                    <li><Link to="/signup">Signup</Link></li>
                    <li><Link to="/login">Login</Link></li>
                    <li><Link to="/links">My Links</Link></li>
                    <li><Link to="/link/create">Create New Link</Link></li>
                </ul>
            </nav>
        </div>
    )
}