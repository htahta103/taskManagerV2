export function SearchPage() {
  return (
    <div className="page">
      <header className="page__header">
        <h1>Search</h1>
        <p className="muted">Find tasks by title or description.</p>
      </header>
      <section className="empty">
        <p>Search will query the API when the task endpoints are wired up.</p>
      </section>
    </div>
  );
}
