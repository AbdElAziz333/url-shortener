
import { useNavigate } from "react-router-dom"
import { useForm, type SubmitHandler } from "react-hook-form"
import { createLink } from "../../services/linkService"
import { toISO } from "../../utils/date"
import type { CreateLinkRequest } from "../../types/linkTypes"

export default function CreateLinkPage() {
    const navigate = useNavigate()
    const {
        register,
        handleSubmit,
        setError,
        formState: { errors, isSubmitting },
    } = useForm<CreateLinkRequest>()

    const onSubmit: SubmitHandler<CreateLinkRequest> = async (data) => {
        try {
            const link = await createLink({
                original_url: data.original_url,
                expires_at: toISO(data.expires_at),
            })
            navigate(`/link/${link.code}`)
        } catch (err) {
            console.error("error creating link:", err)
            setError("root", { message: "Failed to create link" })
        }
    }   

    return (
        <div>
            <h1>Create Link</h1>
            <form onSubmit={handleSubmit(onSubmit)}>
                <input
                    type="url"
                    placeholder="https://example.com"
                    {...register("original_url", {
                        required: "url is required",
                        pattern: { value: /^https?:\/\/.+/, message: "must start with http:// or https://" },
                        maxLength: { value: 2048, message: "url is too long" },
                    })}
                />
                {errors.original_url && <p role="alert">{errors.original_url.message}</p>}

                <input
                    type="date"
                    {...register("expires_at", {
                        validate: (v) => !v || new Date(v) > new Date() || "must be in the future",
                    })}
                />
                {errors.expires_at && <p role="alert">{errors.expires_at.message}</p>}

                {errors.root && <p role="alert">{errors.root.message}</p>}

                <button type="submit" disabled={isSubmitting}>create link</button>
            </form>
        </div>
    )
}