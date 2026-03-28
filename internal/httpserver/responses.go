package httpserver

import (
	"time"

	"github.com/htahta103/taskmanagerv2/internal/store"
)

func formatUser(u store.User) map[string]any {
	return map[string]any{
		"id":         u.ID.String(),
		"email":      u.Email,
		"name":       u.Name,
		"created_at": u.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at": u.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func formatTag(t store.Tag) map[string]any {
	return map[string]any{
		"id":         t.ID.String(),
		"name":       t.Name,
		"created_at": t.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func formatTags(tags []store.Tag) []any {
	out := make([]any, 0, len(tags))
	for _, t := range tags {
		out = append(out, formatTag(t))
	}
	return out
}

func formatProject(p store.Project) map[string]any {
	m := map[string]any{
		"id":         p.ID.String(),
		"name":       p.Name,
		"archived":   p.Archived,
		"created_at": p.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at": p.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if p.ArchivedAt != nil {
		m["archived_at"] = p.ArchivedAt.UTC().Format(time.RFC3339Nano)
	} else {
		m["archived_at"] = nil
	}
	return m
}

func formatTask(t store.Task, tags []store.Tag) map[string]any {
	m := map[string]any{
		"id":           t.ID.String(),
		"title":        t.Title,
		"status":       t.Status,
		"focus_bucket": t.FocusBucket,
		"created_at":   t.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":   t.UpdatedAt.UTC().Format(time.RFC3339Nano),
		"tags":         formatTags(tags),
	}
	if t.Description != nil {
		m["description"] = *t.Description
	} else {
		m["description"] = nil
	}
	if t.Priority != nil {
		m["priority"] = *t.Priority
	} else {
		m["priority"] = nil
	}
	if t.DueDate != nil {
		m["due_date"] = t.DueDate.UTC().Format("2006-01-02")
	} else {
		m["due_date"] = nil
	}
	if t.ProjectID != nil {
		m["project_id"] = t.ProjectID.String()
	} else {
		m["project_id"] = nil
	}
	if t.AssigneeID != nil {
		m["assignee_id"] = t.AssigneeID.String()
	} else {
		m["assignee_id"] = nil
	}
	return m
}
