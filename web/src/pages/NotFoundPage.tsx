import { Link } from "react-router-dom";

export function NotFoundPage() {
  return (
    <div className="page page--center">
      <h1>Page not found</h1>
      <p className="muted">
        <Link to="/">Back to Inbox</Link>
      </p>
    </div>
  );
}
