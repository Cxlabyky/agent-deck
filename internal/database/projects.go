package database

import (
	"database/sql"
	"fmt"
	"time"
)

// CreateProject creates a new project.
func (db *DB) CreateProject(p *Project) error {
	if p.ID == "" {
		p.ID = generateID()
	}
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err := db.conn.Exec(`
		INSERT INTO projects (id, name, path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, p.ID, p.Name, p.Path, p.CreatedAt, p.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

// GetProject retrieves a project by ID.
func (db *DB) GetProject(id string) (*Project, error) {
	p := &Project{}
	err := db.conn.QueryRow(`
		SELECT id, name, path, created_at, updated_at
		FROM projects WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &p.Path, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return p, nil
}

// GetProjectByName retrieves a project by name.
func (db *DB) GetProjectByName(name string) (*Project, error) {
	p := &Project{}
	err := db.conn.QueryRow(`
		SELECT id, name, path, created_at, updated_at
		FROM projects WHERE name = ?
	`, name).Scan(&p.ID, &p.Name, &p.Path, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return p, nil
}

// UpdateProject updates a project.
func (db *DB) UpdateProject(p *Project) error {
	p.UpdatedAt = time.Now()
	result, err := db.conn.Exec(`
		UPDATE projects SET name = ?, path = ?, updated_at = ?
		WHERE id = ?
	`, p.Name, p.Path, p.UpdatedAt, p.ID)

	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("project not found: %s", p.ID)
	}
	return nil
}

// ListProjects returns all projects.
func (db *DB) ListProjects() ([]*Project, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, path, created_at, updated_at
		FROM projects ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		p := &Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Path, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, nil
}

// DeleteProject deletes a project and all associated data (cascades).
func (db *DB) DeleteProject(id string) error {
	result, err := db.conn.Exec("DELETE FROM projects WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("project not found: %s", id)
	}
	return nil
}
