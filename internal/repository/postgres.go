package repository

import (
	"database/sql"
	"fmt"
	"time"

	"subscription-service/internal/model"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Repository interface {
	Create(sub *model.Subscription) error
	GetByID(id uuid.UUID) (*model.Subscription, error)
	GetAll(filter model.SubscriptionFilter) ([]model.Subscription, error)
	Update(id uuid.UUID, req model.UpdateSubscriptionRequest) error
	Delete(id uuid.UUID) error
	GetTotalCost(filter model.SubscriptionFilter) (int, error)
}

type PostgresRepository struct {
	db  *sql.DB
	log *logrus.Logger
}

func NewPostgresRepository(connStr string, log *logrus.Logger) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresRepository{db: db, log: log}, nil
}

func (r *PostgresRepository) Create(sub *model.Subscription) error {
	query := `
        INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	_, err := r.db.Exec(query,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
		sub.CreatedAt,
		sub.UpdatedAt,
	)

	if err != nil {
		r.log.WithError(err).Error("Failed to create subscription")
		return err
	}

	r.log.WithFields(logrus.Fields{
		"id":      sub.ID,
		"user_id": sub.UserID,
		"service": sub.ServiceName,
	}).Info("Subscription created successfully")

	return nil
}

func (r *PostgresRepository) GetByID(id uuid.UUID) (*model.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at 
              FROM subscriptions WHERE id = $1`

	var sub model.Subscription
	err := r.db.QueryRow(query, id).Scan(
		&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
		&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.WithError(err).WithField("id", id).Error("Failed to get subscription by ID")
		return nil, err
	}

	return &sub, nil
}

func (r *PostgresRepository) GetAll(filter model.SubscriptionFilter) ([]model.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at 
              FROM subscriptions WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *filter.UserID)
		argCount++
	}

	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", argCount)
		args = append(args, *filter.ServiceName)
		argCount++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND start_date >= $%d", argCount)
		args = append(args, parseMonthYear(*filter.StartDate))
		argCount++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND (end_date <= $%d OR end_date IS NULL)", argCount)
		args = append(args, parseMonthYear(*filter.EndDate))
		argCount++
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		r.log.WithError(err).Error("Failed to get all subscriptions")
		return nil, err
	}
	defer rows.Close()

	var subscriptions []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		err := rows.Scan(
			&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
			&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			r.log.WithError(err).Error("Failed to scan subscription row")
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

func (r *PostgresRepository) Update(id uuid.UUID, req model.UpdateSubscriptionRequest) error {
	query := "UPDATE subscriptions SET updated_at = $2"
	args := []interface{}{id, time.Now()}
	argCount := 3

	if req.ServiceName != nil {
		query += fmt.Sprintf(", service_name = $%d", argCount)
		args = append(args, *req.ServiceName)
		argCount++
	}

	if req.Price != nil {
		query += fmt.Sprintf(", price = $%d", argCount)
		args = append(args, *req.Price)
		argCount++
	}

	if req.EndDate != nil {
		endDate := parseMonthYear(*req.EndDate)
		query += fmt.Sprintf(", end_date = $%d", argCount)
		args = append(args, endDate)
		argCount++
	}

	query += " WHERE id = $1"

	result, err := r.db.Exec(query, args...)
	if err != nil {
		r.log.WithError(err).WithField("id", id).Error("Failed to update subscription")
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	r.log.WithFields(logrus.Fields{
		"id":            id,
		"rows_affected": rowsAffected,
	}).Info("Subscription updated successfully")

	return nil
}

func (r *PostgresRepository) Delete(id uuid.UUID) error {
	query := "DELETE FROM subscriptions WHERE id = $1"

	result, err := r.db.Exec(query, id)
	if err != nil {
		r.log.WithError(err).WithField("id", id).Error("Failed to delete subscription")
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	r.log.WithFields(logrus.Fields{
		"id":            id,
		"rows_affected": rowsAffected,
	}).Info("Subscription deleted successfully")

	return nil
}

func (r *PostgresRepository) GetTotalCost(filter model.SubscriptionFilter) (int, error) {
	query := `SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *filter.UserID)
		argCount++
	}

	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", argCount)
		args = append(args, *filter.ServiceName)
		argCount++
	}

	if filter.StartDate != nil && filter.EndDate != nil {
		startDate := parseMonthYear(*filter.StartDate)
		endDate := parseMonthYear(*filter.EndDate)

		query += fmt.Sprintf(" AND start_date <= $%d AND (end_date >= $%d OR end_date IS NULL)", argCount, argCount+1)
		args = append(args, endDate, startDate)
	}

	var total int
	err := r.db.QueryRow(query, args...).Scan(&total)
	if err != nil {
		r.log.WithError(err).Error("Failed to calculate total cost")
		return 0, err
	}

	return total, nil
}

func parseMonthYear(monthYear string) time.Time {
	t, err := time.Parse("01-2006", monthYear)
	if err != nil {
		return time.Time{}
	}
	return t
}
