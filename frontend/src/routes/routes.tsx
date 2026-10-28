import { createBrowserRouter, Navigate } from "react-router-dom"
import HomePage from "../pages/HomePage"
import SignupPage from "../pages/SignupPage"
import LoginPage from "../pages/LoginPage"
import CreateLinkPage from "../pages/links/CreateLinkPage"
import LinkPage from "../pages/links/LinkPage"
import LinksPage from "../pages/links/LinksPage"
import NotFoundPage from "../pages/NotFoundPage"
import { PublicRoute } from "./PublicRoute"
import { ProtectedRoute } from "./ProtectedRoute"

export const router = createBrowserRouter([
    {
        errorElement: <NotFoundPage />,
        children: [
            { path: "/", element: <Navigate to="/home" replace /> },
            {
                element: <PublicRoute />,
                children: [
                    { path: "/home", element: <HomePage />},
                    { path: "/signup", element: <SignupPage /> },
                    { path: "/login", element: <LoginPage /> },
                ],
            },
            {
                element: <ProtectedRoute />,
                children: [
                    { path: "/links", element: <LinksPage /> },
                    { path: "/link/create", element: <CreateLinkPage /> },
                    { path: "/link/:code", element: <LinkPage /> },
                ],
            },
        ],
    },
])