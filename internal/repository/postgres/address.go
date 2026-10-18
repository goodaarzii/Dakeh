package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AddressRepo struct {
	q sqlc.Querier
}

func NewAddressRepo(q sqlc.Querier) *AddressRepo {
	return &AddressRepo{q: q}
}

func (r *AddressRepo) Create(ctx context.Context, u domain.Address) (domain.Address, error) {

	row, err := r.q.CreateAddress(ctx, sqlc.CreateAddressParams{
		UserID:        u.UserID,
		Label:         u.Label,
		RecipientName: u.RecipientName,
		Phone:         u.Phone,
		Country:       u.Country,
		City:          u.City,
		Province:      u.Province,
		Street:        u.Street,
		PostalCode:    u.PostalCode,
		Lat:           u.Lat,
		Lng:           u.Lng,
		IsDefault:     u.IsDefault,
	})
	if err != nil {
		return domain.Address{}, mapAddressErr(err)
	}
	return toDomainAddress(row), nil
}

func (r *AddressRepo) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (domain.Address, error) {

	addressByID, err := r.q.GetAddressByID(ctx,
		sqlc.GetAddressByIDParams{
			ID:     id,
			UserID: userID,
		})
	if err != nil {
		return domain.Address{}, mapAddressErr(err)
	}
	return toDomainAddress(addressByID), nil

}

func (r *AddressRepo) GetUserAddresses(ctx context.Context, userID uuid.UUID) ([]domain.Address, error) {
	addresses, err := r.q.ListAddressesByUser(ctx, userID)
	if err != nil {
		return nil, mapAddressErr(err)
	}

	return toDomainAddresses(addresses), err

}

func (r *AddressRepo) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (int64, error) {
	address, err := r.q.SoftDeleteAddress(ctx, sqlc.SoftDeleteAddressParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return -1, mapAddressErr(err)
	}

	return address, err

}

func (r *AddressRepo) Update(ctx context.Context, u domain.Address) (domain.Address, error) {

	row, err := r.q.UpdateAddress(ctx, sqlc.UpdateAddressParams{
		Label:         u.Label,
		RecipientName: u.RecipientName,
		Phone:         u.Phone,
		Country:       u.Country,
		Province:      u.Province,
		City:          u.City,
		Street:        u.Street,
		PostalCode:    u.PostalCode,
		Lat:           u.Lat,
		Lng:           u.Lng,
		ID:            u.ID,
		UserID:        u.UserID,
	})
	if err != nil {
		return domain.Address{}, mapAddressErr(err)
	}

	return toDomainAddress(row), nil
}

func toDomainAddress(u sqlc.Address) domain.Address {
	return domain.Address{
		ID:            u.ID,
		Province:      u.Province,
		UserID:        u.UserID,
		Label:         u.Label,
		RecipientName: u.RecipientName,
		Phone:         u.Phone,
		Country:       u.Country,
		City:          u.City,
		Street:        u.Street,
		PostalCode:    u.PostalCode,
		Lat:           u.Lat,
		Lng:           u.Lng,
		IsDefault:     u.IsDefault,
		UpdatedAt:     u.UpdatedAt,
		CreatedAt:     u.CreatedAt,
	}
}

func toDomainAddresses(rows []sqlc.Address) []domain.Address {
	out := make([]domain.Address, len(rows))
	for i, row := range rows {
		out[i] = toDomainAddress(row)
	}
	return out
}

func mapAddressErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound // also: someone else's address, because every query includes user_id
	}
	switch code, constraint := pgConstraint(err); code {
	case pgUniqueViolation:
		if constraint == "uq_addresses_one_default_per_user" {
			return domain.ErrConflict // this user already has a default address
		}
	case pgForeignKeyViolation:
		if constraint == "addresses_user_id_fkey" {
			return domain.ErrUnauthorized // the token is valid but the user no longer exists
		}
	case pgCheckViolation:
		return domain.ErrInvalidInput // lat or lng out of range
	}
	return fmt.Errorf("address repo: %w", err)
}
