package postgres

import (
	"context"
	"database/sql"
	"errors"
	"microservice-roomservice/internal/domain/entities"

	_ "github.com/lib/pq"
)

// OrderRepositoryPostgres implements the OrderRepository interface
type OrderRepositoryPostgres struct {
	db *sql.DB
}

// NewOrderRepository creates a new PostgreSQL order repository
func NewOrderRepository(db *sql.DB) *OrderRepositoryPostgres {
	return &OrderRepositoryPostgres{
		db: db,
	}
}

// Save saves an order and its items in a transaction
func (r *OrderRepositoryPostgres) Save(ctx context.Context, order *entities.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert order
	orderQuery := `
		INSERT INTO orders (id, room_id, reservation_id, billing_folder_id, billing_ticket_id, total_price, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = tx.ExecContext(ctx, orderQuery,
		order.ID,
		order.RoomID,
		order.ReservationID,
		order.BillingFolderID,
		order.BillingTicketID,
		order.TotalPrice,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Insert order items
	itemQuery := `
		INSERT INTO order_items (id, order_id, menu_item_id, quantity, price)
		VALUES ($1, $2, $3, $4, $5)
	`

	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, itemQuery,
			item.ID,
			order.ID,
			item.MenuItemID,
			item.Quantity,
			item.Price,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// FindByID finds an order by ID
func (r *OrderRepositoryPostgres) FindByID(ctx context.Context, id string) (*entities.Order, error) {
	// Query order
	orderQuery := `
		SELECT id, room_id, reservation_id, billing_folder_id, billing_ticket_id, total_price, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var order entities.Order
	err := r.db.QueryRowContext(ctx, orderQuery, id).Scan(
		&order.ID,
		&order.RoomID,
		&order.ReservationID,
		&order.BillingFolderID,
		&order.BillingTicketID,
		&order.TotalPrice,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("order not found")
	}
	if err != nil {
		return nil, err
	}

	// Query order items
	itemsQuery := `
		SELECT id, menu_item_id, quantity, price
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entities.OrderItem
	for rows.Next() {
		var item entities.OrderItem
		if err := rows.Scan(
			&item.ID,
			&item.MenuItemID,
			&item.Quantity,
			&item.Price,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	order.Items = items
	return &order, nil
}

// FindByRoomID finds all orders for a specific room
func (r *OrderRepositoryPostgres) FindByRoomID(ctx context.Context, roomID string) ([]entities.Order, error) {
	query := `
		SELECT id, room_id, reservation_id, billing_folder_id, billing_ticket_id, total_price, status, created_at, updated_at
		FROM orders
		WHERE room_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []entities.Order
	for rows.Next() {
		var order entities.Order
		if err := rows.Scan(
			&order.ID,
			&order.RoomID,
			&order.ReservationID,
			&order.BillingFolderID,
			&order.BillingTicketID,
			&order.TotalPrice,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Load order items
		items, err := r.loadOrderItems(ctx, order.ID)
		if err != nil {
			return nil, err
		}
		order.Items = items

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// FindByReservationID finds all orders for a specific reservation
func (r *OrderRepositoryPostgres) FindByReservationID(ctx context.Context, reservationID string) ([]entities.Order, error) {
	query := `
		SELECT id, room_id, reservation_id, billing_folder_id, billing_ticket_id, total_price, status, created_at, updated_at
		FROM orders
		WHERE reservation_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, reservationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []entities.Order
	for rows.Next() {
		var order entities.Order
		if err := rows.Scan(
			&order.ID,
			&order.RoomID,
			&order.ReservationID,
			&order.BillingFolderID,
			&order.BillingTicketID,
			&order.TotalPrice,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Load order items
		items, err := r.loadOrderItems(ctx, order.ID)
		if err != nil {
			return nil, err
		}
		order.Items = items

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// Update updates an order
func (r *OrderRepositoryPostgres) Update(ctx context.Context, order *entities.Order) error {
	query := `
		UPDATE orders
		SET room_id = $2, reservation_id = $3, billing_folder_id = $4, billing_ticket_id = $5, total_price = $6, status = $7, updated_at = $8
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		order.ID,
		order.RoomID,
		order.ReservationID,
		order.BillingFolderID,
		order.BillingTicketID,
		order.TotalPrice,
		order.Status,
		order.UpdatedAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("order not found")
	}

	return nil
}

// Delete deletes an order by ID
func (r *OrderRepositoryPostgres) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM orders WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("order not found")
	}

	return nil
}

// FindByStatus finds all orders with a specific status
func (r *OrderRepositoryPostgres) FindByStatus(ctx context.Context, status entities.OrderStatus) ([]entities.Order, error) {
	query := `
		SELECT id, room_id, reservation_id, billing_folder_id, billing_ticket_id, total_price, status, created_at, updated_at
		FROM orders
		WHERE status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []entities.Order
	for rows.Next() {
		var order entities.Order
		if err := rows.Scan(
			&order.ID,
			&order.RoomID,
			&order.ReservationID,
			&order.BillingFolderID,
			&order.BillingTicketID,
			&order.TotalPrice,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Load order items
		items, err := r.loadOrderItems(ctx, order.ID)
		if err != nil {
			return nil, err
		}
		order.Items = items

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// loadOrderItems is a helper function to load order items
func (r *OrderRepositoryPostgres) loadOrderItems(ctx context.Context, orderID string) ([]entities.OrderItem, error) {
	query := `
		SELECT id, menu_item_id, quantity, price
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entities.OrderItem
	for rows.Next() {
		var item entities.OrderItem
		if err := rows.Scan(
			&item.ID,
			&item.MenuItemID,
			&item.Quantity,
			&item.Price,
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
