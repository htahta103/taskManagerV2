export function InboxPage() {
  return (
    <div className="page">
      <header className="page__header">
        <h1>Inbox</h1>
        <p className="muted">All open tasks will land here.</p>
      </header>
      <section className="empty">
        <p>No tasks yet — task list UI ships in the next milestone.</p>
      </section>
    </div>
  );
}
