package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CalendarRow struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Name      string
	IsDefault bool
	Color     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (q *Queries) ListCalendarsByUser(ctx context.Context, userID uuid.UUID) ([]CalendarRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, owner_id, name, is_default, color, created_at, updated_at
		 FROM calendars
		 WHERE owner_id = $1 AND deleted_at IS NULL
		 ORDER BY is_default DESC, name ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calendars []CalendarRow
	for rows.Next() {
		var c CalendarRow
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.Name, &c.IsDefault, &c.Color, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		calendars = append(calendars, c)
	}
	return calendars, rows.Err()
}

func (q *Queries) GetCalendarByID(ctx context.Context, calendarID uuid.UUID) (CalendarRow, error) {
	var c CalendarRow
	err := q.db.QueryRowContext(ctx,
		`SELECT id, owner_id, name, is_default, color, created_at, updated_at
		 FROM calendars WHERE id = $1 AND deleted_at IS NULL`, calendarID).
		Scan(&c.ID, &c.OwnerID, &c.Name, &c.IsDefault, &c.Color, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (q *Queries) CreateCalendar(ctx context.Context, ownerID uuid.UUID, name, color string, isDefault bool) (CalendarRow, error) {
	var c CalendarRow
	err := q.db.QueryRowContext(ctx,
		`INSERT INTO calendars (owner_id, name, color, is_default)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, owner_id, name, is_default, color, created_at, updated_at`,
		ownerID, name, color, isDefault).
		Scan(&c.ID, &c.OwnerID, &c.Name, &c.IsDefault, &c.Color, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (q *Queries) UpdateCalendar(ctx context.Context, calendarID uuid.UUID, name, color string) (CalendarRow, error) {
	var c CalendarRow
	err := q.db.QueryRowContext(ctx,
		`UPDATE calendars SET name = $2, color = $3, updated_at = now()
		 WHERE id = $1 AND deleted_at IS NULL
		 RETURNING id, owner_id, name, is_default, color, created_at, updated_at`,
		calendarID, name, color).
		Scan(&c.ID, &c.OwnerID, &c.Name, &c.IsDefault, &c.Color, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (q *Queries) SoftDeleteCalendar(ctx context.Context, calendarID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE calendars SET deleted_at = now() WHERE id = $1`, calendarID)
	return err
}

func (q *Queries) GetDefaultCalendar(ctx context.Context, userID uuid.UUID) (CalendarRow, error) {
	var c CalendarRow
	err := q.db.QueryRowContext(ctx,
		`SELECT id, owner_id, name, is_default, color, created_at, updated_at
		 FROM calendars
		 WHERE owner_id = $1 AND is_default = TRUE AND deleted_at IS NULL`, userID).
		Scan(&c.ID, &c.OwnerID, &c.Name, &c.IsDefault, &c.Color, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}
