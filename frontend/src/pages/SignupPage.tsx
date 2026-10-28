
import { Link } from "react-router-dom"
import { useForm, type SubmitHandler } from "react-hook-form"
import { signup } from "../services/authService"
import type { SignupRequest } from "../types/authTypes"
import { useAuth } from "../context/AuthContext"
import { EMAIL_PATTERN } from "../utils/validation"

export default function SignupPage() {
    const { login } = useAuth()
    const {register, handleSubmit, formState: { errors, isSubmitting }} = useForm<SignupRequest>()

    const onSubmit: SubmitHandler<SignupRequest> = async (data) => {
        try {
            await signup(data)
            await login(data)
        } catch (err) {
            console.error(`Signup Failed: ${err}`)
        }
    }

    return (
        <div>
            <h1>Signup Page</h1>
            <form onSubmit={handleSubmit(onSubmit)}>
                <input type="email" placeholder="Email" {...register("email", {
                    required: "Email is required",
                    maxLength: {value: 255, message: "Email must be at most 255 characters"},
                    pattern: {value: EMAIL_PATTERN, message: "Email should be valid"}
                })} />
                {errors.email && <div>{errors.email.message}</div>}

                <input type="password" placeholder="Password" {...register("password", {
                    required: "Password is required",
                    minLength: {value: 8, message: "Password must be at least 8 characters"},
                    maxLength: {value: 72, message: "Password must be at most 72 characters"}
                })} />
                {errors.password && <div>{errors.password.message}</div>}

                <button type="submit" disabled={isSubmitting}>signup</button>
            </form>
            <br />
            <Link to="/login">have an account? login!</Link>
            <br />
            <Link to="/home">Home page</Link>
        </div>
    )
}