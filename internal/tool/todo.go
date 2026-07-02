package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/opentmd/opentmd/internal/llm"
)

type todoItem struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

type todoList struct {
	mu     sync.Mutex
	items  []todoItem
	nextID int
}

func newTodoList() *todoList {
	return &todoList{nextID: 1}
}

type todoArgs struct {
	Action  string `json:"action"`
	Content string `json:"content"`
	ID      int    `json:"id"`
	Status  string `json:"status"`
}

func (r *Runtime) registerTodo() {
	r.Register(todoToolDef(), r.todoTool)
}

func todoToolDef() llm.ToolDefinition {
	return llm.ToolDefinition{
		Name: "todo",
		Description: "Manage a task list to track progress on multi-step work. " +
			"Use 'add' to create tasks, 'update' to change status, and 'list' to show all tasks.",
		Parameters: objParam(map[string]any{
			"action":  strParam("Action: add, update, or list"),
			"content": strParam("Task description (required for add)"),
			"id":      intParam("Task ID (required for update)"),
			"status":  strParam("New status: pending, in_progress, or completed (required for update)"),
		}, "action"),
	}
}

func (r *Runtime) todoTool(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[todoArgs](args, "todo")
	if err != nil {
		return "", err
	}
	switch a.Action {
	case "add":
		content := strings.TrimSpace(a.Content)
		if content == "" {
			content = "Untitled task"
		}
		return r.todos.add(content), nil
	case "update":
		if a.ID <= 0 {
			return "", fmt.Errorf("todo: id required for update")
		}
		status := strings.TrimSpace(a.Status)
		if status == "" {
			return "", fmt.Errorf("todo: status required for update")
		}
		return r.todos.update(a.ID, status)
	case "list":
		return r.todos.format(), nil
	default:
		return "", fmt.Errorf("todo: unknown action %q (use add, update, or list)", a.Action)
	}
}

func (tl *todoList) add(content string) string {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	id := tl.nextID
	tl.nextID++
	tl.items = append(tl.items, todoItem{ID: id, Content: content, Status: "pending"})
	return fmt.Sprintf("Added task #%d: %s", id, content)
}

func (tl *todoList) update(id int, status string) (string, error) {
	switch status {
	case "pending", "in_progress", "completed":
	default:
		return "", fmt.Errorf("todo: invalid status %q", status)
	}
	tl.mu.Lock()
	defer tl.mu.Unlock()
	for i := range tl.items {
		if tl.items[i].ID == id {
			tl.items[i].Status = status
			return fmt.Sprintf("Updated task #%d to %s", id, status), nil
		}
	}
	return "", fmt.Errorf("todo: task #%d not found", id)
}

func (tl *todoList) format() string {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	if len(tl.items) == 0 {
		return "No tasks."
	}
	var sb strings.Builder
	for _, item := range tl.items {
		icon := "[ ]"
		switch item.Status {
		case "completed":
			icon = "[x]"
		case "in_progress":
			icon = "[>]"
		}
		fmt.Fprintf(&sb, "%s %d. %s\n", icon, item.ID, item.Content)
	}
	return strings.TrimSuffix(sb.String(), "\n")
}
