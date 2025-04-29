package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/entity"
)

type registrationRepository struct {
	db *sqlx.DB
}

func NewRegistrationRepository(db *sqlx.DB) contract.IRegistrationRepository {
	return &registrationRepository{
		db: db,
	}
}

func (r *registrationRepository) createRegistration(ctx context.Context, tx sqlx.ExtContext,
	registration *entity.Registration) error {

	_, err := sqlx.NamedExecContext(
		ctx,
		tx,
		`INSERT INTO registrations (
			conference_id, user_id
		) VALUES (
			:conference_id, :user_id
		)`,
		registration,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *registrationRepository) CreateRegistration(ctx context.Context, registration *entity.Registration) error {
	return r.createRegistration(ctx, r.db, registration)
}

func (r *registrationRepository) GetRegisteredUsersByConference(ctx context.Context,
	conferenceID uuid.UUID, lazy dto.LazyLoadQuery) ([]entity.User, dto.LazyLoadResponse, error) {

	var users []entity.User
	var args []interface{}
	args = append(args, conferenceID)
	argCount := 1

	query := `SELECT id, name FROM users
        WHERE id IN (
            SELECT user_id FROM registrations
            WHERE conference_id = $1
        )`

	// Add pagination filters
	if lazy.AfterID != uuid.Nil {
		query += fmt.Sprintf(" AND id > $%d", argCount+1)
		args = append(args, lazy.AfterID)
		argCount++
	}
	if lazy.BeforeID != uuid.Nil {
		query += fmt.Sprintf(" AND id < $%d", argCount+1)
		args = append(args, lazy.BeforeID)
		argCount++
	}

	// Add ordering and limit
	if lazy.BeforeID != uuid.Nil {
		query += " ORDER BY id DESC"
	} else {
		query += " ORDER BY id ASC"
	}
	query += fmt.Sprintf(" LIMIT $%d", argCount+1)
	args = append(args, lazy.Limit+1) // Request one extra record to determine if there are more results

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, dto.LazyLoadResponse{}, fmt.Errorf("failed to query registered users: %w", err)
	}
	defer rows.Close()

	// Scan results
	for rows.Next() {
		var user entity.User
		if err2 := rows.Scan(&user.ID, &user.Name); err2 != nil {
			return nil, dto.LazyLoadResponse{}, fmt.Errorf("failed to scan user: %w", err2)
		}
		users = append(users, user)
	}

	if err2 := rows.Err(); err2 != nil {
		return nil, dto.LazyLoadResponse{}, fmt.Errorf("error iterating users: %w", err2)
	}

	// Prepare response
	lazyResp := dto.LazyLoadResponse{
		HasMore: false,
		FirstID: nil,
		LastID:  nil,
	}

	if len(users) > 0 {
		// Check if we got an extra record
		if len(users) > lazy.Limit {
			lazyResp.HasMore = true
			if lazy.BeforeID != uuid.Nil {
				users = users[1:] // Remove first record when paginating backwards
			} else {
				users = users[:lazy.Limit] // Remove last record when paginating forwards
			}
		}

		// For BeforeID, reverse the final result set to maintain ascending order
		if lazy.BeforeID != uuid.Nil {
			for i := 0; i < len(users)/2; i++ {
				j := len(users) - 1 - i
				users[i], users[j] = users[j], users[i]
			}
		}

		lazyResp.FirstID = users[0].ID
		lazyResp.LastID = users[len(users)-1].ID
	}

	return users, lazyResp, nil
}

func (r *registrationRepository) GetRegisteredConferencesByUser(ctx context.Context, userID uuid.UUID,
	includePast bool, lazy dto.LazyLoadQuery) ([]entity.Conference, dto.LazyLoadResponse, error) {

	var conferences []entity.Conference
	var args []interface{}
	args = append(args, userID)
	argCount := 1

	query := `SELECT
        c.id, c.title, c.description, c.speaker_name, c.speaker_title,
        c.target_audience, c.prerequisites, c.seats, c.starts_at, c.ends_at,
        c.host_id, c.status, c.created_at, c.updated_at, u.name AS host_name
    FROM conferences c
    JOIN users u ON c.host_id = u.id
    JOIN registrations r ON c.id = r.conference_id
    WHERE r.user_id = $1`

	// Add filter for past conferences
	if !includePast {
		query += fmt.Sprintf(" AND c.ends_at > NOW()")
	}

	// Add pagination filters
	if lazy.AfterID != uuid.Nil {
		query += fmt.Sprintf(" AND c.starts_at > (SELECT starts_at FROM conferences WHERE id = $%d)", argCount+1)
		args = append(args, lazy.AfterID)
		argCount++
	}
	if lazy.BeforeID != uuid.Nil {
		query += fmt.Sprintf(" AND c.starts_at < (SELECT starts_at FROM conferences WHERE id = $%d)", argCount+1)
		args = append(args, lazy.BeforeID)
		argCount++
	}

	// Add ordering and limit
	if lazy.BeforeID != uuid.Nil {
		query += " ORDER BY c.starts_at DESC"
	} else {
		query += " ORDER BY c.starts_at ASC"
	}
	query += fmt.Sprintf(" LIMIT $%d", argCount+1)
	args = append(args, lazy.Limit+1) // Request one extra record to determine if there are more results

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, dto.LazyLoadResponse{}, fmt.Errorf("failed to query registered conferences: %w", err)
	}
	defer rows.Close()

	// Scan results
	for rows.Next() {
		var conf entity.Conference
		var hostName string
		if err := rows.Scan(
			&conf.ID, &conf.Title, &conf.Description, &conf.SpeakerName, &conf.SpeakerTitle,
			&conf.TargetAudience, &conf.Prerequisites, &conf.Seats, &conf.StartsAt, &conf.EndsAt,
			&conf.HostID, &conf.Status, &conf.CreatedAt, &conf.UpdatedAt, &hostName,
		); err != nil {
			return nil, dto.LazyLoadResponse{}, fmt.Errorf("failed to scan conference: %w", err)
		}
		conf.Host.ID = conf.HostID
		conf.Host.Name = hostName
		conferences = append(conferences, conf)
	}

	if err := rows.Err(); err != nil {
		return nil, dto.LazyLoadResponse{}, fmt.Errorf("error iterating conferences: %w", err)
	}

	// Prepare response
	lazyResp := dto.LazyLoadResponse{
		HasMore: false,
		FirstID: nil,
		LastID:  nil,
	}

	if len(conferences) > 0 {
		// Check if we got an extra record
		if len(conferences) > lazy.Limit {
			lazyResp.HasMore = true
			if lazy.BeforeID != uuid.Nil {
				conferences = conferences[1:] // Remove first record when paginating backwards
			} else {
				conferences = conferences[:lazy.Limit] // Remove last record when paginating forwards
			}
		}

		// For BeforeID, reverse the final result set to maintain ascending order
		if lazy.BeforeID != uuid.Nil {
			for i := 0; i < len(conferences)/2; i++ {
				j := len(conferences) - 1 - i
				conferences[i], conferences[j] = conferences[j], conferences[i]
			}
		}

		lazyResp.FirstID = conferences[0].ID
		lazyResp.LastID = conferences[len(conferences)-1].ID
	}

	return conferences, lazyResp, nil
}

func (r *registrationRepository) IsUserRegisteredToConference(ctx context.Context, conferenceID,
	userID uuid.UUID) (bool, error) {

	var exists bool
	if err := r.db.GetContext(
		ctx,
		&exists,
		`SELECT EXISTS (
    			SELECT 1 FROM registrations
    			WHERE conference_id = $1
    			AND user_id = $2
    		)`,
		conferenceID, userID,
	); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *registrationRepository) GetConflictingRegistrations(ctx context.Context, userID uuid.UUID, startsAt,
	endsAt time.Time) ([]entity.Conference, error) {

	var conferences []entity.Conference

	query := `
        SELECT
            c.id,
            c.title,
            c.starts_at,
            c.ends_at
        FROM registrations r
        JOIN conferences c ON r.conference_id = c.id
        WHERE r.user_id = $1
            AND c.deleted_at IS NULL
            AND (
                ($2 BETWEEN c.starts_at AND c.ends_at)
                OR
                ($3 BETWEEN c.starts_at AND c.ends_at)
                OR
                (c.starts_at BETWEEN $2 AND $3)
            )`

	if err := r.db.SelectContext(ctx, &conferences, query, userID, startsAt, endsAt); err != nil {
		return nil, err
	}

	return conferences, nil
}

func (r *registrationRepository) CountRegistrationsByConference(ctx context.Context,
	conferenceID uuid.UUID) (int, error) {
	var count int
	if err := r.db.GetContext(
		ctx,
		&count,
		`SELECT COUNT(*) FROM registrations WHERE conference_id = $1`,
		conferenceID,
	); err != nil {
		return 0, err
	}

	return count, nil
}
