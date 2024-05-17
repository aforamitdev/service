package user

import (
	"context"
	"database/sql"
	"log"
	"service2/business/auth"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrNotFound is used when a specific User is requested but does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrAuthenticationFailure occurs when a user attempts to authenticate but
	// anything goes wrong.
	ErrAuthenticationFailure = errors.New("authentication failed")

	// ErrForbidden occurs when a user tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("attempted action is not allowed")
)

type User struct {
	log *log.Logger
	db  *sqlx.DB
}

func New(log *log.Logger, db *sqlx.DB) User {
	return User{log: log, db: db}
}

func (u User) Create(ctx context.Context, traceID string, nu NewUser, now time.Time) (Info, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)

	if err != nil {
		return Info{}, errors.Wrap(err, "generating password ")
	}
	usr := Info{
		ID:           uuid.New().String(),
		Name:         nu.Name,
		Email:        nu.Email,
		PasswordHash: hash,
		Roles:        nu.Roles,
		DateCreated:  now.UTC(),
		DateUpdated:  now.UTC(),
	}

	const q = `INSERT INTO users (user_id,name,email,password_hash,roles,date_created,date_updated) VALUES ($1,$2,$3,$4,$5,$6,$7)`

	u.log.Printf("%s : %s :query : %s", traceID, "user.create", "kl")

	if _, err := u.db.ExecContext(ctx, q, usr.ID, usr.Email, usr.PasswordHash, usr.Roles, usr.DateCreated, usr.DateUpdated); err != nil {
		return Info{}, errors.Wrap(err, "inserting user")
	}
	return usr, nil

}

func (u User) Update(ctx context.Context, claims auth.Claims, userID string, uu UpdateUser, now time.Time) error {

	usr, err := u.One(ctx, claims, userID)

	if err != nil {
		return err
	}

	if uu.Name != nil {
		usr.Name = *uu.Name
	}

	if uu.Email != nil {
		usr.Email = *uu.Email
	}
	if uu.Roles != nil {
		usr.Roles = uu.Roles
	}

	if uu.Password != nil {

		pw, err := bcrypt.GenerateFromPassword([]byte(*uu.Password), bcrypt.DefaultCost)

		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}

		usr.PasswordHash = pw

	}

	usr.DateUpdated = now

	const q = `UPDATE users SET "name"=$2,"email"=$3,"roles"=$4,"password_hash"=$4,"date_updated"=$6 WHERE user_id=$1`

	if _, err := u.db.ExecContext(ctx, q, userID, uu.Name, uu.Email, uu.Roles, usr.PasswordHash, usr.DateUpdated); err != nil {
		return errors.Wrap(err, "update user=")
	}

	return nil

}

func (u User) One(ctx context.Context, claims auth.Claims, userId string) (Info, error) {

	if _, err := uuid.Parse(userId); err != nil {
		return Info{}, ErrInvalidID
	}

	if !claims.HasRoles(auth.RoleAdmin) && claims.Subject != userId {
		return Info{}, ErrForbidden
	}

	const q = `SELECT * FROM users WHERE user_id=$1`

	var i Info
	if err := u.db.GetContext(ctx, &i, userId); err != nil {
		if err == sql.ErrNoRows {

			return Info{}, ErrNotFound
		}
		return Info{}, errors.Wrapf(err, "selected user %q", userId)
	}
	return i, nil

}

func (u User) Delete(ctx context.Context, userID string) error {

	if _, err := uuid.Parse(userID); err != nil {
		return ErrInvalidID
	}

	const q = `DELETE FROM users WHERE user_id=$1`

	if _, err := u.db.ExecContext(ctx, q, userID); err != nil {
		return errors.Wrapf(err, "delete user %d", userID)
	}
	return nil
}

func (u User) Query(ctx context.Context, tractId string) ([]Info, error) {

	const q = `SELECT * FROM users`

	users := []Info{}

	if err := u.db.SelectContext(ctx, &users, q); err != nil {
		return nil, errors.Wrap(err, "selecting users")

	}
	return users, nil

}
