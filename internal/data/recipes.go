package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/svoglimacci/forks/internal/validator"
)

type RecipeModel struct {
	DB *sql.DB
}

func (r RecipeModel) GetAll(title string, categories []string, filters Filters) ([]*Recipe, Metadata, error) {
	query := fmt.Sprintf(`
	SELECT count(*) OVER(), id, created_at, title, description, categories, prep_time, cooking_time, servings, version
	FROM recipes
        WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '') 
        AND (categories @> $2 OR $2 = '{}')    
	ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{title, categories, filters.limit(), filters.offset()}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	m := pgtype.NewMap()
	recipes := []*Recipe{}
	totalRecords := 0
	for rows.Next() {
		var recipe Recipe

		err := rows.Scan(
			&totalRecords,
			&recipe.ID,
			&recipe.CreatedAt,
			&recipe.Title,
			&recipe.Description,
			m.SQLScanner(&recipe.Categories),
			&recipe.PrepTime,
			&recipe.CookingTime,
			&recipe.Servings,
			&recipe.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		recipes = append(recipes, &recipe)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return recipes, metadata, nil
}

func (r RecipeModel) Insert(recipe *Recipe) error {
	query := `
		INSERT INTO recipes (title, description, categories, prep_time, cooking_time, servings)
		VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, created_at, version
	`

	args := []any{recipe.Title, recipe.Description, recipe.Categories, recipe.CookingTime, recipe.PrepTime, recipe.Servings}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return r.DB.QueryRowContext(ctx, query, args...).Scan(&recipe.ID, &recipe.CreatedAt, &recipe.Version)
}

func (r RecipeModel) Get(id int64) (*Recipe, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
	SELECT id, created_at, title, description, categories , prep_time, cooking_time, servings, version
	FROM recipes
	WHERE id = $1`

	var recipe Recipe
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()
	m := pgtype.NewMap()

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&recipe.ID,
		&recipe.CreatedAt,
		&recipe.Title,
		&recipe.Description,
		m.SQLScanner(&recipe.Categories),
		&recipe.PrepTime,
		&recipe.CookingTime,
		&recipe.Servings,
		&recipe.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &recipe, nil
}

func (r RecipeModel) Update(recipe *Recipe) error {
	query := `
	UPDATE recipes
	SET title = $1, description = $2, categories = $3, prep_time = $4, cooking_time = $5, servings = $6, version = version + 1
	WHERE id = $7 AND version = $8
	RETURNING version`

	args := []any{
		recipe.Title,
		recipe.Description,
		recipe.Categories,
		recipe.PrepTime,
		recipe.CookingTime,
		recipe.Servings,
		recipe.ID,
		recipe.Version,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, args...).Scan(&recipe.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (r RecipeModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
	DELETE FROM recipes
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

type Recipe struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"-"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Categories  []string  `json:"categories,omitempty"`
	PrepTime    int32     `json:"prep_time,omitzero"`
	CookingTime int32     `json:"cooking_time,omitzero"`
	Servings    int32     `json:"servings,omitzero"`
	Version     int32     `json:"version"`
	UpdatedAt   int32     `json:"-"`
}

func ValidateRecipe(v *validator.Validator, recipe *Recipe) {
	v.Check(recipe.Title != "", "title", "must be provided")
	v.Check(len(recipe.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(recipe.PrepTime != 0, "prep time", "must be provided")
	v.Check(recipe.PrepTime > 0, "prep time", "must be a positive integer")

	v.Check(recipe.CookingTime != 0, "cooking time", "must be provided")
	v.Check(recipe.CookingTime > 0, "cooking time", "must be a positive integer")

	v.Check(recipe.Categories != nil, "categories", "must be provided")
	v.Check(len(recipe.Categories) >= 1, "categories", "must contain at least 1 category")
	v.Check(len(recipe.Categories) <= 5, "categories", "must not contain more than 5 categories")
	v.Check(validator.Unique(recipe.Categories), "categories", "must not contain duplicate values")

	v.Check(recipe.Servings != 0, "servings", "must be provided")
	v.Check(recipe.Servings > 0, "servings", "must be a positive integer")
}
