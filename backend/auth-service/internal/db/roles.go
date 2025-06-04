package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/elgris/stom"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const RolesTable = "roles"

const (
	RolesID          = "roles_id_pk"
	RolesCode        = "roles_code"
	RolesName        = "roles_name"
	RolesDescription = "roles_descr"
)

type Role struct {
	ID          int64  `db:"roles_id_pk" insert:"roles_id_pk"`
	Code        string `db:"roles_code" insert:"roles_code"`
	Name        string `db:"roles_name" insert:"roles_name"`
	Description string `db:"roles_descr" insert:"roles_descr"`
}

var (
	stomRoleSelect = stom.MustNewStom(Role{}).SetTag(selectTag)
	stomRoleInsert = stom.MustNewStom(Role{}).SetTag(insertTag)
	stomRoleUpdate = stom.MustNewStom(Role{}).SetTag(updateTag)
)

func (r *Role) columns(pref string) []string {
	return colNamesWithPref(stomRoleSelect.TagValues(), pref)
}

type RoleQuery interface {
	GetByID(ctx context.Context, id int64) (*Role, error)
	GetIDByCode(ctx context.Context, code int) (int64, error)
	GetIDByName(ctx context.Context, name string) (int64, error)
	Insert(ctx context.Context, role *Role) (*Role, error)
	Update(ctx context.Context, role *Role, id int64) (*Role, error)
	Delete(ctx context.Context, id int64) error
}

type roleQuery struct {
	runner *pgxpool.Pool
	sq     squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewRoleQuery(runner *pgxpool.Pool, sq squirrel.StatementBuilderType, logger *zap.Logger) RoleQuery {
	return &roleQuery{
		runner: runner,
		sq:     sq,
		logger: logger,
	}
}

func (r *roleQuery) GetByID(ctx context.Context, id int64) (*Role, error) {
	r.logger.Debug("Fetching role by ID", zap.Int64("role_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, r.logger, r.runner)
	if err != nil {
		r.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return nil, fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	role := &Role{}
	qb, args, err := r.sq.Select(role.columns("")...).
		From(RolesTable).
		Where(squirrel.Eq{RolesID: id}).
		ToSql()
	if err != nil {
		r.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	err = pgxscan.Get(ctx, conn, role, qb, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.logger.Warn("Database error",
				zap.Int64("role_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			r.logger.Warn("Failed to fetch role", zap.Int64("role_id", id), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	r.logger.Info("Role fetched successfully", zap.Int64("role_id", id))
	return role, nil
}

func (r *roleQuery) GetIDByName(ctx context.Context, name string) (int64, error) {
	name = strings.ToLower(name)
	r.logger.Debug("Fetching role ID by name", zap.String("name", name))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, r.logger, r.runner)
	if err != nil {
		r.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return 0, fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	var roleID int64
	qb, args, err := r.sq.Select("id").
		From(RolesTable).
		Where(squirrel.Eq{RolesName: name}).
		ToSql()
	if err != nil {
		r.logger.Error("Failed to build query", zap.Error(err))
		return 0, fmt.Errorf("failed to build query: %w", err)
	}
	err = conn.QueryRow(ctx, qb, args...).Scan(&roleID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.logger.Warn("Database error",
				zap.String("name", name),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			r.logger.Warn("Failed to fetch role ID", zap.String("name", name), zap.Error(err))
		}
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}
	r.logger.Info("Role ID fetched successfully", zap.String("name", name), zap.Int64("role_id", roleID))
	return roleID, nil
}

func (r *roleQuery) GetIDByCode(ctx context.Context, code int) (int64, error) {
	r.logger.Debug("Fetching role ID by code", zap.Int("code", code))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, r.logger, r.runner)
	if err != nil {
		r.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return 0, fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	var roleID int64
	qb, args, err := r.sq.Select("id").
		From(RolesTable).
		Where(squirrel.Eq{RolesCode: code}).
		ToSql()
	if err != nil {
		r.logger.Error("Failed to build query", zap.Error(err))
		return 0, fmt.Errorf("failed to build query: %w", err)
	}
	err = conn.QueryRow(ctx, qb, args...).Scan(&roleID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.logger.Warn("Database error",
				zap.Int("code", code),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			r.logger.Warn("Failed to fetch role ID", zap.Int("code", code), zap.Error(err))
		}
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}
	r.logger.Info("Role ID fetched successfully", zap.Int("code", code), zap.Int64("role_id", roleID))
	return roleID, nil
}

func (r *roleQuery) Insert(ctx context.Context, role *Role) (*Role, error) {
	r.logger.Debug("Inserting role", zap.Any("role", role))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, r.logger, r.runner)
	if err != nil {
		r.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return nil, fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	insertMap, err := stomRoleInsert.ToMap(role)
	if err != nil {
		r.logger.Error("Failed to map struct", zap.Error(err))
		return nil, fmt.Errorf("failed to map struct: %w", err)
	}
	qb, args, err := r.sq.Insert(RolesTable).
		SetMap(insertMap).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		r.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Get(ctx, conn, role, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.logger.Warn("Database error",
				zap.Any("role", role),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			r.logger.Error("Failed to insert role", zap.Any("role", role), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	r.logger.Info("Role inserted successfully", zap.Int64("role_id", role.ID))
	return role, nil
}

func (r *roleQuery) Update(ctx context.Context, role *Role, id int64) (*Role, error) {
	r.logger.Debug("Updating role", zap.Int64("role_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, r.logger, r.runner)
	if err != nil {
		r.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return nil, fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	updateMap, err := stomRoleUpdate.ToMap(role)
	if err != nil {
		r.logger.Error("Failed to map struct", zap.Error(err))
		return nil, fmt.Errorf("failed to map struct: %w", err)
	}
	qb, args, err := r.sq.Update(RolesTable).
		SetMap(updateMap).
		Where(squirrel.Eq{RolesID: id}).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		r.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Get(ctx, conn, role, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.logger.Warn("Database error",
				zap.Int64("role_id", role.ID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			r.logger.Error("Failed to update role", zap.Int64("role_id", role.ID), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	r.logger.Info("Role updated successfully", zap.Int64("role_id", role.ID))
	return role, nil
}

func (r *roleQuery) Delete(ctx context.Context, id int64) error {
	r.logger.Debug("Deleting role", zap.Int64("role_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, r.logger, r.runner)
	if err != nil {
		r.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	qb, args, err := r.sq.Delete(RolesTable).
		Where(squirrel.Eq{RolesID: id}).
		ToSql()
	if err != nil {
		r.logger.Error("Failed to build query", zap.Error(err))
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := conn.Exec(ctx, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.logger.Warn("Database error",
				zap.Int64("role_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			r.logger.Error("Failed to delete role", zap.Int64("role_id", id), zap.Error(err))
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("No role found to delete", zap.Int64("role_id", id))
		return fmt.Errorf("no role found with id %d", id)
	}

	r.logger.Info("Role deleted successfully", zap.Int64("role_id", id))
	return nil
}
