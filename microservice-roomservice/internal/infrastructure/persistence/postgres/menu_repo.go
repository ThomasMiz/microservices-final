package postgres

import (
	"context"
	"database/sql"
	"errors"
	"microservice-roomservice/internal/domain/entities"

	_ "github.com/lib/pq"
)

// MenuRepositoryPostgres implements the MenuRepository interface
type MenuRepositoryPostgres struct {
	db *sql.DB
}

// NewMenuRepository creates a new PostgreSQL menu repository
func NewMenuRepository(db *sql.DB) *MenuRepositoryPostgres {
	return &MenuRepositoryPostgres{
		db: db,
	}
}

// Save saves a menu item
func (r *MenuRepositoryPostgres) Save(ctx context.Context, item *entities.MenuItem) error {
	query := `
		INSERT INTO menu_items (id, description, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		item.ID,
		item.Description,
		item.Price,
		item.CreatedAt,
		item.UpdatedAt,
	)

	return err
}

// FindByID finds a menu item by ID
func (r *MenuRepositoryPostgres) FindByID(ctx context.Context, id string) (*entities.MenuItem, error) {
	query := `
		SELECT id, description, price, created_at, updated_at
		FROM menu_items
		WHERE id = $1
	`

	var item entities.MenuItem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Description,
		&item.Price,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("menu item not found")
	}
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// FindAll retrieves all menu items
func (r *MenuRepositoryPostgres) FindAll(ctx context.Context) ([]entities.MenuItem, error) {
	query := `
		SELECT id, description, price, created_at, updated_at
		FROM menu_items
		ORDER BY description
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entities.MenuItem
	for rows.Next() {
		var item entities.MenuItem
		if err := rows.Scan(
			&item.ID,
			&item.Description,
			&item.Price,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// Update updates a menu item
func (r *MenuRepositoryPostgres) Update(ctx context.Context, item *entities.MenuItem) error {
	query := `
		UPDATE menu_items
		SET description = $2, price = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		item.ID,
		item.Description,
		item.Price,
		item.UpdatedAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("menu item not found")
	}

	return nil
}

// Delete deletes a menu item by ID
func (r *MenuRepositoryPostgres) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM menu_items WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("menu item not found")
	}

	return nil
}

// ExistsByID checks if a menu item exists by ID
func (r *MenuRepositoryPostgres) ExistsByID(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM menu_items WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
