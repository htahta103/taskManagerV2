const apiBase =
  import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, "") || "http://localhost:8080";

export default function App() {
  return (
    <main className="app">
      <h1>Task Manager V2</h1>
      <p className="muted">
        API base: <code>{apiBase}</code>
      </p>
      <p>Scaffold ready — wire routes and data next.</p>
    </main>
  );
}
