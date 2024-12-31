package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/yeboahd24/personal-finance-manager/internal/model"
)

// GoalRepository defines all goal-related database operations
type GoalRepository interface {
	CreateGoal(ctx context.Context, goal *model.Goal) error
	GetGoalByID(ctx context.Context, id string) (*model.Goal, error)
	GetUserGoals(ctx context.Context, userID string) ([]*model.Goal, error)
	UpdateGoal(ctx context.Context, goal *model.Goal) error
	DeleteGoal(ctx context.Context, id string, userID string) error
}

// GoalSQL handles goal-related database operations
type GoalSQL struct {
	db *sql.DB
	tx *sql.Tx
}

// SQLExecutor interface for database operations
type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// CreateGoal creates a new financial goal
func (r *GoalSQL) CreateGoal(ctx context.Context, goal *model.Goal) error {
	if goal.ID == "" {
		goal.ID = uuid.New().String()
	}
	now := time.Now()
	goal.CreatedAt = now
	goal.UpdatedAt = now

	query := `
		INSERT INTO goals (id, user_id, name, target_amount, current_amount, deadline, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	var executor SQLExecutor
	if r.tx != nil {
		executor = r.tx
	} else {
		executor = r.db
	}

	_, err := executor.ExecContext(ctx, query,
		goal.ID,
		goal.UserID,
		goal.Name,
		goal.TargetAmount,
		goal.CurrentAmount,
		goal.Deadline,
		goal.CreatedAt,
		goal.UpdatedAt,
	)
	return err
}

// GetGoalByID retrieves a goal by its ID
func (r *GoalSQL) GetGoalByID(ctx context.Context, id string) (*model.Goal, error) {
	query := `
		SELECT id, user_id, name, target_amount, current_amount, deadline, created_at, updated_at
		FROM goals
		WHERE id = $1
	`

	var executor SQLExecutor
	if r.tx != nil {
		executor = r.tx
	} else {
		executor = r.db
	}

	goal := &model.Goal{}
	err := executor.QueryRowContext(ctx, query, id).Scan(
		&goal.ID,
		&goal.UserID,
		&goal.Name,
		&goal.TargetAmount,
		&goal.CurrentAmount,
		&goal.Deadline,
		&goal.CreatedAt,
		&goal.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return goal, err
}

// GetUserGoals retrieves all goals for a user
func (r *GoalSQL) GetUserGoals(ctx context.Context, userID string) ([]*model.Goal, error) {
	query := `
		SELECT id, user_id, name, target_amount, current_amount, deadline, created_at, updated_at
		FROM goals
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	var executor SQLExecutor
	if r.tx != nil {
		executor = r.tx
	} else {
		executor = r.db
	}

	rows, err := executor.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []*model.Goal
	for rows.Next() {
		goal := &model.Goal{}
		err := rows.Scan(
			&goal.ID,
			&goal.UserID,
			&goal.Name,
			&goal.TargetAmount,
			&goal.CurrentAmount,
			&goal.Deadline,
			&goal.CreatedAt,
			&goal.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		goals = append(goals, goal)
	}
	return goals, rows.Err()
}

// UpdateGoal updates an existing goal
func (r *GoalSQL) UpdateGoal(ctx context.Context, goal *model.Goal) error {
	goal.UpdatedAt = time.Now()

	query := `
		UPDATE goals
		SET name = $1,
			target_amount = $2,
			current_amount = $3,
			deadline = $4,
			updated_at = $5
		WHERE id = $6 AND user_id = $7
	`

	var executor SQLExecutor
	if r.tx != nil {
		executor = r.tx
	} else {
		executor = r.db
	}

	_, err := executor.ExecContext(ctx, query,
		goal.Name,
		goal.TargetAmount,
		goal.CurrentAmount,
		goal.Deadline,
		goal.UpdatedAt,
		goal.ID,
		goal.UserID,
	)
	return err
}

// DeleteGoal deletes a goal
func (r *GoalSQL) DeleteGoal(ctx context.Context, id string, userID string) error {
	query := `DELETE FROM goals WHERE id = $1 AND user_id = $2`

	var executor SQLExecutor
	if r.tx != nil {
		executor = r.tx
	} else {
		executor = r.db
	}

	_, err := executor.ExecContext(ctx, query, id, userID)
	return err
}
