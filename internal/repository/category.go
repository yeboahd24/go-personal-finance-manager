package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/yeboahd24/personal-finance-manager/internal/errors"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *model.Category) error
	GetCategoryByID(ctx context.Context, id string) (*model.Category, error)
	GetCategories(ctx context.Context, userID string) ([]*model.Category, error)
	UpdateCategory(ctx context.Context, category *model.Category) error
	DeleteCategory(ctx context.Context, id string) error
	InitializeDefaultCategories(ctx context.Context, userID string) error
}

type CategorySQL struct {
	db *sql.DB
	tx *sql.Tx
}

func (r *CategorySQL) query() QueryExecutor {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *CategorySQL) CreateCategory(ctx context.Context, category *model.Category) error {
	query := `
		INSERT INTO categories (id, name, type, icon, color, parent_id, user_id, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at`

	log.Printf("Creating category: %+v", category)
	err := r.query().QueryRowContext(
		ctx,
		query,
		category.Name,
		category.Type,
		category.Icon,
		category.Color,
		category.ParentID,
		category.UserID,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		log.Printf("Error creating category: %v", err)
		return errors.Wrap(err, "Failed to create category", 500)
	}
	log.Printf("Created category with ID: %s", category.ID)
	return nil
}

func (r *CategorySQL) GetCategoryByID(ctx context.Context, id string) (*model.Category, error) {
	category := &model.Category{}
	query := `
		SELECT id, name, type, icon, color, parent_id, user_id, created_at, updated_at
		FROM categories
		WHERE id = $1`

	err := r.query().QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Type,
		&category.Icon,
		&category.Color,
		&category.ParentID,
		&category.UserID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get category", 500)
	}
	return category, nil
}

func (r *CategorySQL) GetCategories(ctx context.Context, userID string) ([]*model.Category, error) {
	query := `
		SELECT id, name, type, icon, color, parent_id, user_id, created_at, updated_at
		FROM categories
		WHERE user_id = $1
		ORDER BY COALESCE(parent_id, id), name`

	rows, err := r.query().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get categories", 500)
	}
	defer rows.Close()

	var categories []*model.Category
	for rows.Next() {
		category := &model.Category{}
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Type,
			&category.Icon,
			&category.Color,
			&category.ParentID,
			&category.UserID,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to scan category", 500)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *CategorySQL) UpdateCategory(ctx context.Context, category *model.Category) error {
	query := `
		UPDATE categories
		SET name = $2, icon = $3, color = $4
		WHERE id = $1
		RETURNING updated_at`

	err := r.query().QueryRowContext(
		ctx,
		query,
		category.ID,
		category.Name,
		category.Icon,
		category.Color,
	).Scan(&category.UpdatedAt)

	if err == sql.ErrNoRows {
		return errors.ErrNotFound
	}
	if err != nil {
		return errors.Wrap(err, "Failed to update category", 500)
	}
	return nil
}

func (r *CategorySQL) DeleteCategory(ctx context.Context, id string) error {
	// First check if category has any transactions
	var count int
	err := r.query().QueryRowContext(ctx, "SELECT COUNT(*) FROM transactions WHERE category_id = $1", id).Scan(&count)
	if err != nil {
		return errors.Wrap(err, "Failed to check category usage", 500)
	}
	if count > 0 {
		return errors.New("Cannot delete category with associated transactions", 400)
	}

	// Then check if category has any child categories
	err = r.query().QueryRowContext(ctx, "SELECT COUNT(*) FROM categories WHERE parent_id = $1", id).Scan(&count)
	if err != nil {
		return errors.Wrap(err, "Failed to check child categories", 500)
	}
	if count > 0 {
		return errors.New("Cannot delete category with child categories", 400)
	}

	query := "DELETE FROM categories WHERE id = $1"
	result, err := r.query().ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "Failed to delete category", 500)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "Failed to get rows affected", 500)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *CategorySQL) InitializeDefaultCategories(ctx context.Context, userID string) error {
	// Create parent categories first
	parentCategories := []model.Category{
		{Name: "Income", Type: "income", UserID: userID},
		{Name: "Housing", Type: "expense", UserID: userID},
		{Name: "Transportation", Type: "expense", UserID: userID},
		{Name: "Food & Dining", Type: "expense", UserID: userID},
		{Name: "Shopping", Type: "expense", UserID: userID},
		{Name: "Bills & Utilities", Type: "expense", UserID: userID},
		{Name: "Healthcare", Type: "expense", UserID: userID},
		{Name: "Entertainment", Type: "expense", UserID: userID},
	}

	// Map to store parent category IDs
	parentIDs := make(map[string]string)

	// Create parent categories
	for _, category := range parentCategories {
		if err := r.CreateCategory(ctx, &category); err != nil {
			return fmt.Errorf("failed to create parent category: %w", err)
		}
		parentIDs[category.Name] = category.ID
	}

	// Create subcategories
	subcategories := []model.Category{
		// Income subcategories
		{Name: "Salary", Type: "income", UserID: userID, ParentID: func() *string { id := parentIDs["Income"]; return &id }()},
		{Name: "Investments", Type: "income", UserID: userID, ParentID: func() *string { id := parentIDs["Income"]; return &id }()},
		{Name: "Freelance", Type: "income", UserID: userID, ParentID: func() *string { id := parentIDs["Income"]; return &id }()},
		{Name: "Other Income", Type: "income", UserID: userID, ParentID: func() *string { id := parentIDs["Income"]; return &id }()},

		// Housing subcategories
		{Name: "Rent/Mortgage", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Housing"]; return &id }()},
		{Name: "Property Tax", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Housing"]; return &id }()},
		{Name: "Home Insurance", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Housing"]; return &id }()},
		{Name: "Home Maintenance", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Housing"]; return &id }()},

		// Transportation subcategories
		{Name: "Public Transit", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Transportation"]; return &id }()},
		{Name: "Fuel", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Transportation"]; return &id }()},
		{Name: "Car Insurance", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Transportation"]; return &id }()},
		{Name: "Car Maintenance", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Transportation"]; return &id }()},
		{Name: "Parking", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Transportation"]; return &id }()},

		// Food & Dining subcategories
		{Name: "Groceries", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Food & Dining"]; return &id }()},
		{Name: "Restaurants", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Food & Dining"]; return &id }()},
		{Name: "Coffee Shops", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Food & Dining"]; return &id }()},
		{Name: "Food Delivery", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Food & Dining"]; return &id }()},

		// Shopping subcategories
		{Name: "Clothing", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Shopping"]; return &id }()},
		{Name: "Electronics", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Shopping"]; return &id }()},
		{Name: "Home Goods", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Shopping"]; return &id }()},
		{Name: "Personal Care", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Shopping"]; return &id }()},

		// Bills & Utilities subcategories
		{Name: "Electricity", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Bills & Utilities"]; return &id }()},
		{Name: "Water", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Bills & Utilities"]; return &id }()},
		{Name: "Internet", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Bills & Utilities"]; return &id }()},
		{Name: "Phone", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Bills & Utilities"]; return &id }()},
		{Name: "Streaming Services", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Bills & Utilities"]; return &id }()},

		// Healthcare subcategories
		{Name: "Health Insurance", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Healthcare"]; return &id }()},
		{Name: "Doctor", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Healthcare"]; return &id }()},
		{Name: "Pharmacy", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Healthcare"]; return &id }()},
		{Name: "Dental", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Healthcare"]; return &id }()},

		// Entertainment subcategories
		{Name: "Movies", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Entertainment"]; return &id }()},
		{Name: "Games", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Entertainment"]; return &id }()},
		{Name: "Hobbies", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Entertainment"]; return &id }()},
		{Name: "Sports", Type: "expense", UserID: userID, ParentID: func() *string { id := parentIDs["Entertainment"]; return &id }()},
	}

	// Create subcategories
	for _, category := range subcategories {
		if err := r.CreateCategory(ctx, &category); err != nil {
			return fmt.Errorf("failed to create subcategory: %w", err)
		}
	}

	return nil
}
