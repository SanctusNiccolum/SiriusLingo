package db

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/elgris/stom"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const UsersTable = "users"

const (
	UsersID                 = "users_id_pk"
	UsersUsername           = "users_username"
	UsersPasswordHash       = "users_password_hash"
	UsersEmail              = "users_email"
	UsersRoleID             = "users_role_id_fk"
	UsersAccessTokenSecret  = "users_access_token_secret"
	UsersRefreshTokenSecret = "users_refresh_token_secret"
	UsersAuthTime           = "users_auth_time"
	UsersCreatedAt          = "users_created_at"
	UsersUpdatedAt          = "users_updated_at"
)

type User struct {
	ID                 int64      `db:"users_id_pk"`
	Username           string     `db:"users_username" insert:"users_username" update:"users_username"`
	Password           string     `db:"users_password_hash" insert:"users_password_hash" update:"users_password_hash"`
	Email              string     `db:"users_email" insert:"users_email"`
	RoleID             int64      `db:"users_roles_id_fk" insert:"users_roles_id_fk"`
	AccessTokenSecret  string     `db:"users_access_token_secret" insert:"users_access_token_secret"`
	RefreshTokenSecret string     `db:"users_refresh_token_secret" insert:"users_refresh_token_secret"`
	AccessTokenJTI     *string    `db:"users_access_token_jti" updateAuth:"users_access_token_jti"`
	RefreshTokenJTI    *string    `db:"users_refresh_token_jti" updateAuth:"users_refresh_token_jti"`
	AuthTime           *time.Time `db:"users_auth_time" insert:"users_auth_time"`
	CreatedAt          *time.Time `db:"users_created_at"`
	UpdatedAt          *time.Time `db:"users_updated_at" update:"users_updated_at" updateAuth:"users_updated_at"`
}

var (
	stomUserSelect     = stom.MustNewStom(User{}).SetTag(selectTag)
	stomUserInsert     = stom.MustNewStom(User{}).SetTag(insertTag)
	stomUserUpdate     = stom.MustNewStom(User{}).SetTag(updateTag)
	stomUserAuthUpdate = stom.MustNewStom(User{}).SetTag(updateAuthTag)
)

func (u *User) columns(pref string) []string {
	return colNamesWithPref(stomUserSelect.TagValues(), pref)
}

type UserQuery interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	ExistsByUsernameOrEmail(ctx context.Context, username string, email string) (bool, error)
	Insert(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, user *User, id int64) (*User, error)
	UpdateAuthTime(ctx context.Context, id int64) (*User, error)
	UpdateLoginOrLogout(ctx context.Context, user *User, id int64) (*User, error)
	Delete(ctx context.Context, id int64) error
}

type userQuery struct {
	runner *pgxpool.Pool
	sq     squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewUserQuery(runner *pgxpool.Pool, sq squirrel.StatementBuilderType, logger *zap.Logger) UserQuery {
	return &userQuery{
		runner: runner,
		sq:     sq,
		logger: logger,
	}
}

func (u *userQuery) GetByID(ctx context.Context, id int64) (*User, error) {
	u.logger.Debug("Fetching user by ID", zap.Int64("user_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, u.logger, u.runner)
	if err != nil {
		u.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return nil, fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	user := &User{}
	qb, args, err := u.sq.Select(user.columns("")...).
		From(UsersTable).
		Where(squirrel.Eq{UsersID: id}).
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	err = pgxscan.Get(ctx, conn, user, qb, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Int64("user_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Warn("Failed to fetch user", zap.Int64("user_id", id), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	u.logger.Info("User fetched successfully", zap.Int64("user_id", id))
	return user, nil
}

func (u *userQuery) GetByUsername(ctx context.Context, username string) (*User, error) {
	u.logger.Debug("Fetching user by username", zap.String("username", username))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, u.logger, u.runner)
	if err != nil {
		u.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return nil, fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	user := &User{}
	qb, args, err := u.sq.Select(user.columns("")...).
		From(UsersTable).
		Where(squirrel.Eq{UsersUsername: username}).
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	err = pgxscan.Get(ctx, conn, user, qb, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.String("username", username),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Warn("Failed to fetch user", zap.String("username", username), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	u.logger.Info("User fetched successfully", zap.String("username", username))
	return user, nil
}

func (u *userQuery) GetByEmail(ctx context.Context, email string) (*User, error) {
	u.logger.Debug("Fetching user by email", zap.String("email", email))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, u.logger, u.runner)
	if err != nil {
		u.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return nil, fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	user := &User{}
	qb, args, err := u.sq.Select(user.columns("")...).
		From(UsersTable).
		Where(squirrel.Eq{UsersEmail: email}).
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	err = pgxscan.Get(ctx, conn, user, qb, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.String("email", email),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Warn("Failed to fetch user", zap.String("email", email), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	u.logger.Info("User fetched successfully", zap.String("email", email))
	return user, nil
}

func (u *userQuery) ExistsByUsernameOrEmail(ctx context.Context, username, email string) (bool, error) {
	u.logger.Debug("Checking if user exists by username or email",
		zap.String("username", username),
		zap.String("email", email))

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, u.logger, u.runner)
	if err != nil {
		u.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return false, fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	var count int
	query, args, err := u.sq.Select("COUNT(*)").
		From(UsersTable).
		Where(squirrel.Or{
			squirrel.Eq{UsersUsername: username},
			squirrel.Eq{UsersEmail: email},
		}).
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	err = conn.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.String("username", username),
				zap.String("email", email),
				zap.String("error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Error("Failed to check user existence",
				zap.String("username", username),
				zap.String("email", email),
				zap.Error(err),
			)
		}
		return false, fmt.Errorf("failed to execute query: %w", err)
	}

	exists := count > 0
	if exists {
		u.logger.Info("User already exists",
			zap.String("username", username),
			zap.String("email", email),
		)
	}
	return exists, nil
}

func (u *userQuery) Insert(ctx context.Context, user *User) (*User, error) {
	u.logger.Debug("Inserting user", zap.Any("user", user))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, u.logger, u.runner)
	if err != nil {
		u.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return nil, fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	insertMap, err := stomUserInsert.ToMap(user)
	if err != nil {
		u.logger.Error("Failed to map struct", zap.Error(err))
		return nil, fmt.Errorf("failed to map struct: %w", err)
	}
	qb, args, err := u.sq.Insert(UsersTable).
		SetMap(insertMap).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Get(ctx, conn, user, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Any("user", user),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Error("Failed to insert user", zap.Any("user", user), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	u.logger.Info("User inserted successfully", zap.Int64("user_id", user.ID))
	return user, nil
}

func (u *userQuery) Update(ctx context.Context, user *User, id int64) (*User, error) {
	u.logger.Debug("Updating user", zap.Int64("user_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, u.logger, u.runner)
	if err != nil {
		u.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return nil, fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	updateMap, err := stomUserUpdate.ToMap(user)
	if err != nil {
		u.logger.Error("Failed to map struct", zap.Error(err))
		return nil, fmt.Errorf("failed to map struct: %w", err)
	}
	qb, args, err := u.sq.Update(UsersTable).
		SetMap(updateMap).
		Where(squirrel.Eq{UsersID: id}).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Get(ctx, conn, user, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Int64("user_id", user.ID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Error("Failed to update user", zap.Int64("user_id", user.ID), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	u.logger.Info("User updated successfully", zap.Int64("user_id", user.ID))
	return user, nil
}

func (u *userQuery) Delete(ctx context.Context, id int64) error {
	u.logger.Debug("Deleting user", zap.Int64("user_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, u.logger, u.runner)
	if err != nil {
		u.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	qb, args, err := u.sq.Delete(UsersTable).
		Where(squirrel.Eq{UsersID: id}).
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := conn.Exec(ctx, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Int64("user_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Error("Failed to delete user", zap.Int64("user_id", id), zap.Error(err))
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		u.logger.Warn("No user found to delete", zap.Int64("user_id", id))
		return fmt.Errorf("no user found with id %d", id)
	}

	u.logger.Info("User deleted successfully", zap.Int64("user_id", id))
	return nil
}

func (u *userQuery) UpdateLoginOrLogout(ctx context.Context, user *User, id int64) (*User, error) {
	u.logger.Debug("Updating user for auth", zap.Int64("user_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, u.logger, u.runner)
	if err != nil {
		u.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return nil, fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	updateMap, err := stomUserAuthUpdate.ToMap(user)
	if err != nil {
		u.logger.Error("Failed to map struct", zap.Error(err))
		return nil, fmt.Errorf("failed to map struct: %w", err)
	}
	qb, args, err := u.sq.Update(UsersTable).
		SetMap(updateMap).
		Where(squirrel.Eq{UsersID: id}).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Get(ctx, conn, user, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Int64("user_id", user.ID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Error("Failed to update user", zap.Int64("user_id", user.ID), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	u.logger.Info("User updated successfully", zap.Int64("user_id", user.ID))
	return user, nil
}

func (u *userQuery) UpdateAuthTime(ctx context.Context, id int64) (*User, error) {
	u.logger.Debug("Updating user auth time", zap.Int64("user_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := acquireHealthyConn(ctx, u.logger, u.runner)
	if err != nil {
		u.logger.Error("Failed to acquire healthy connection", zap.Error(err))
		return nil, fmt.Errorf("failed to acquire healthy connection: %w", err)
	}
	defer conn.Release()

	var user User
	qb, args, err := u.sq.Update(UsersTable).
		Set(UsersAuthTime, time.Now()).
		Where(squirrel.Eq{UsersID: id}).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	err = pgxscan.Get(ctx, conn, &user, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Int64("user_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Error("Failed to update user auth time",
				zap.Int64("user_id", id),
				zap.Error(err),
			)
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	u.logger.Info("User auth time updated successfully",
		zap.Int64("user_id", user.ID),
		zap.Time("new_auth_time", *user.AuthTime),
	)
	return &user, nil
}

func GenerateSecretKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", key), nil
}
