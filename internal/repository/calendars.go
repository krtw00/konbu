package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CalendarRow struct {
	ID         uuid.UUID
	OwnerID    uuid.UUID
	Name       string
	IsDefault  bool
	IsExternal bool
	Color      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

const calendarCols = `id, owner_id, name, is_default, is_external, color, created_at, updated_at`

func scanCalendar(row interface{ Scan(...any) error }) (CalendarRow, error) {
	var c CalendarRow
	err := row.Scan(&c.ID, &c.OwnerID, &c.Name, &c.IsDefault, &c.IsExternal, &c.Color, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (q *Queries) ListCalendarsByUser(ctx context.Context, userID uuid.UUID) ([]CalendarRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT `+calendarCols+`
		 FROM calendars
		 WHERE owner_id = $1 AND deleted_at IS NULL
		 ORDER BY is_default DESC, name ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calendars []CalendarRow
	for rows.Next() {
		c, err := scanCalendar(rows)
		if err != nil {
			return nil, err
		}
		calendars = append(calendars, c)
	}
	return calendars, rows.Err()
}

func (q *Queries) GetCalendarByID(ctx context.Context, calendarID uuid.UUID) (CalendarRow, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT `+calendarCols+`
		 FROM calendars WHERE id = $1 AND deleted_at IS NULL`, calendarID)
	return scanCalendar(row)
}

func (q *Queries) CreateCalendar(ctx context.Context, ownerID uuid.UUID, name, color string, isDefault bool) (CalendarRow, error) {
	return q.createCalendar(ctx, ownerID, name, color, isDefault, false)
}

// CreateExternalCalendar creates a read-only calendar backed by an external
// iCal subscription (is_external = true). is_default is always false.
func (q *Queries) CreateExternalCalendar(ctx context.Context, ownerID uuid.UUID, name, color string) (CalendarRow, error) {
	return q.createCalendar(ctx, ownerID, name, color, false, true)
}

func (q *Queries) createCalendar(ctx context.Context, ownerID uuid.UUID, name, color string, isDefault, isExternal bool) (CalendarRow, error) {
	row := q.db.QueryRowContext(ctx,
		`INSERT INTO calendars (owner_id, name, color, is_default, is_external)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+calendarCols,
		ownerID, name, color, isDefault, isExternal)
	return scanCalendar(row)
}

func (q *Queries) UpdateCalendar(ctx context.Context, calendarID uuid.UUID, name, color string) (CalendarRow, error) {
	row := q.db.QueryRowContext(ctx,
		`UPDATE calendars SET name = $2, color = $3, updated_at = now()
		 WHERE id = $1 AND deleted_at IS NULL
		 RETURNING `+calendarCols,
		calendarID, name, color)
	return scanCalendar(row)
}

func (q *Queries) SoftDeleteCalendar(ctx context.Context, calendarID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE calendars SET deleted_at = now() WHERE id = $1`, calendarID)
	return err
}

func (q *Queries) GetDefaultCalendar(ctx context.Context, userID uuid.UUID) (CalendarRow, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT `+calendarCols+`
		 FROM calendars
		 WHERE owner_id = $1 AND is_default = TRUE AND deleted_at IS NULL`, userID)
	return scanCalendar(row)
}
