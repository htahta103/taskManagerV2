import { FormEvent, useState } from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";
import { ApiError } from "../auth/api";
import { useAuth } from "../auth/AuthContext";

export function SignupPage() {
  const { signup, user, loading } = useAuth();
  const navigate = useNavigate();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  if (loading) {
    return (
      <div className="page page--center">
        <p className="muted">Loading…</p>
      </div>
    );
  }

  if (user) {
    return <Navigate to="/" replace />;
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setFormError(null);
    setSubmitting(true);
    try {
      await signup({ name, email, password });
      navigate("/", { replace: true });
    } catch (err) {
      if (err instanceof ApiError) setFormError(err.message);
      else if (err instanceof Error) setFormError(err.message);
      else setFormError("Could not create account");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="page page--narrow page--center">
      <div className="card">
        <h1 className="card__title">Create account</h1>
        <p className="muted card__subtitle">Start with name, email, and a password.</p>
        <form className="stack" onSubmit={onSubmit}>
          <label className="field">
            <span className="field__label">Name</span>
            <input
              className="input"
              type="text"
              name="name"
              autoComplete="name"
              required
              maxLength={120}
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </label>
          <label className="field">
            <span className="field__label">Email</span>
            <input
              className="input"
              type="email"
              name="email"
              autoComplete="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </label>
          <label className="field">
            <span className="field__label">Password</span>
            <input
              className="input"
              type="password"
              name="password"
              autoComplete="new-password"
              required
              minLength={8}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </label>
          {formError ? <p className="error">{formError}</p> : null}
          <button className="btn btn--primary" type="submit" disabled={submitting}>
            {submitting ? "Creating…" : "Create account"}
          </button>
        </form>
        <p className="muted footnote">
          Already have an account? <Link to="/login">Sign in</Link>
        </p>
      </div>
    </div>
  );
}
