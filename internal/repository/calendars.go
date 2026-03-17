package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CalendarRow struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Name        string
	IsDefault   bool
	ShareToken  *string
	Color       string
	MemberCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CalendarMemberRow struct {
	CalendarID uuid.UUID
	UserID     uuid.UUID
	UserName   string
	UserEmail  string
	Role       string
	Color      string
	JoinedAt   time.Time
}

func (q *Queries) ListCalendarsByUser(ctx context.Context, userID uuid.UUID) ([]CalendarRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT c.id, c.owner_id, c.name, c.is_default, c.share_token, c.color,
		        (SELECT count(*) FROM calendar_members WHERE calendar_id = c.id) AS member_count,
		        c.created_at, c.updated_at
		 FROM calendars c
		 JOIN calendar_members cm ON cm.calendar_id = c.id
		 WHERE cm.user_id = $1 AND c.deleted_at IS NULL
		 ORDER BY c.is_default DESC, c.name ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calendars []CalendarRow
	for rows.Next() {
		var c CalendarRow
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.Name, &c.IsDefault, &c.ShareToken, &c.Color, &c.MemberCount, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		calendars = append(calendars, c)
	}
	return calendars, rows.Err()
}

func (q *Queries) GetCalendarByID(ctx context.Context, calendarID uuid.UUID) (CalendarRow, error) {
	var c CalendarRow
	err := q.db.QueryRowContext(ctx,
		`SELECT id, owner_id, name, is_default, share_token, color,
		        (SELECT count(*) FROM calendar_members WHERE calendar_id = $1) AS member_count,
		        created_at, updated_at
		 FROM calendars WHERE id = $1 AND deleted_at IS NULL`, calendarID).
		Scan(&c.ID, &c.OwnerID, &c.Name, &c.IsDefault, &c.ShareToken, &c.Color, &c.MemberCount, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (q *Queries) CreateCalendar(ctx context.Context, ownerID uuid.UUID, name, color string, isDefault bool) (CalendarRow, error) {
	var c CalendarRow
	err := q.db.QueryRowContext(ctx,
		`INSERT INTO calendars (owner_id, name, color, is_default)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, owner_id, name, is_default, share_token, color, 0, created_at, updated_at`,
		ownerID, name, color, isDefault).
		Scan(&c.ID, &c.OwnerID, &c.Name, &c.IsDefault, &c.ShareToken, &c.Color, &c.MemberCount, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (q *Queries) UpdateCalendar(ctx context.Context, calendarID uuid.UUID, name, color string) (CalendarRow, error) {
	var c CalendarRow
	err := q.db.QueryRowContext(ctx,
		`UPDATE calendars SET name = $2, color = $3, updated_at = now()
		 WHERE id = $1 AND deleted_at IS NULL
		 RETURNING id, owner_id, name, is_default, share_token, color,
		           (SELECT count(*) FROM calendar_members WHERE calendar_id = $1),
		           created_at, updated_at`,
		calendarID, name, color).
		Scan(&c.ID, &c.OwnerID, &c.Name, &c.IsDefault, &c.ShareToken, &c.Color, &c.MemberCount, &c.CreatedAt, &c.UpdatedAt)
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
		`SELECT c.id, c.owner_id, c.name, c.is_default, c.share_token, c.color,
		        (SELECT count(*) FROM calendar_members WHERE calendar_id = c.id) AS member_count,
		        c.created_at, c.updated_at
		 FROM calendars c
		 WHERE c.owner_id = $1 AND c.is_default = TRUE AND c.deleted_at IS NULL`, userID).
		Scan(&c.ID, &c.OwnerID, &c.Name, &c.IsDefault, &c.ShareToken, &c.Color, &c.MemberCount, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (q *Queries) SetShareToken(ctx context.Context, calendarID uuid.UUID, token string) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE calendars SET share_token = $2, updated_at = now() WHERE id = $1 AND deleted_at IS NULL`,
		calendarID, token)
	return err
}

func (q *Queries) ClearShareToken(ctx context.Context, calendarID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE calendars SET share_token = NULL, updated_at = now() WHERE id = $1 AND deleted_at IS NULL`,
		calendarID)
	return err
}

func (q *Queries) GetCalendarByToken(ctx context.Context, token string) (CalendarRow, error) {
	var c CalendarRow
	err := q.db.QueryRowContext(ctx,
		`SELECT id, owner_id, name, is_default, share_token, color,
		        (SELECT count(*) FROM calendar_members WHERE calendar_id = calendars.id) AS member_count,
		        created_at, updated_at
		 FROM calendars WHERE share_token = $1 AND deleted_at IS NULL`, token).
		Scan(&c.ID, &c.OwnerID, &c.Name, &c.IsDefault, &c.ShareToken, &c.Color, &c.MemberCount, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (q *Queries) ListCalendarMembers(ctx context.Context, calendarID uuid.UUID) ([]CalendarMemberRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT cm.calendar_id, cm.user_id, u.name, u.email, cm.role, cm.color, cm.joined_at
		 FROM calendar_members cm
		 JOIN users u ON u.id = cm.user_id
		 WHERE cm.calendar_id = $1
		 ORDER BY cm.joined_at ASC`, calendarID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []CalendarMemberRow
	for rows.Next() {
		var m CalendarMemberRow
		if err := rows.Scan(&m.CalendarID, &m.UserID, &m.UserName, &m.UserEmail, &m.Role, &m.Color, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (q *Queries) AddCalendarMember(ctx context.Context, calendarID, userID uuid.UUID, role, color string) error {
	_, err := q.db.ExecContext(ctx,
		`INSERT INTO calendar_members (calendar_id, user_id, role, color)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (calendar_id, user_id) DO NOTHING`,
		calendarID, userID, role, color)
	return err
}

func (q *Queries) UpdateCalendarMember(ctx context.Context, calendarID, userID uuid.UUID, role, color string) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE calendar_members SET role = $3, color = $4
		 WHERE calendar_id = $1 AND user_id = $2`,
		calendarID, userID, role, color)
	return err
}

func (q *Queries) RemoveCalendarMember(ctx context.Context, calendarID, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`DELETE FROM calendar_members WHERE calendar_id = $1 AND user_id = $2`,
		calendarID, userID)
	return err
}

func (q *Queries) GetCalendarMember(ctx context.Context, calendarID, userID uuid.UUID) (CalendarMemberRow, error) {
	var m CalendarMemberRow
	err := q.db.QueryRowContext(ctx,
		`SELECT cm.calendar_id, cm.user_id, u.name, u.email, cm.role, cm.color, cm.joined_at
		 FROM calendar_members cm
		 JOIN users u ON u.id = cm.user_id
		 WHERE cm.calendar_id = $1 AND cm.user_id = $2`,
		calendarID, userID).
		Scan(&m.CalendarID, &m.UserID, &m.UserName, &m.UserEmail, &m.Role, &m.Color, &m.JoinedAt)
	return m, err
}
