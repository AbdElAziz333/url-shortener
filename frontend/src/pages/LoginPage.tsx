
import { Link } from "react-router-dom"
import { useForm, type SubmitHandler } from "react-hook-form"
import type { LoginRequest } from "../types/authTypes"
import { useAuth } from "../context/AuthContext";
import { EMAIL_PATTERN } from "../utils/validation";

export default function LoginPage() {
    const { login } = useAuth();
    const {register, handleSubmit, formState: {errors, isSubmitting}} = useForm<LoginRequest>()

    const onSubmit: SubmitHandler<LoginRequest> = async (data) => {
        try {
            await login(data)
        } catch (err) {
            console.error("Login failed:", err)
        }
    }

    return (
        <div>
            <h1>Login Page</h1>
            <form onSubmit={handleSubmit(onSubmit)}>
                <input type="email" placeholder="Enter email" {...register("email", {
                    required: "Email is required",
                    pattern: {value: EMAIL_PATTERN, message: "Invalid email format"},
                })} />
                {errors.email && <div>{errors.email.message}</div>}

                <input type="password" placeholder="Enter password" {...register("password", {
                    required: "Password is required",
                    minLength: {value: 8, message: "Password must be at least 8 characters"}
                })} />
                {errors.password && <div>{errors.password.message}</div>}

                <button type="submit" disabled={isSubmitting}>Login</button>
            </form>
            <br />
            <Link to={"/signup"}>don't have an account yet? create one!</Link>
            <br />
            <Link to={"/home"}>Home page</Link>
        </div>
    )
}