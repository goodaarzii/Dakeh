package postgres

import (
	"context"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
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
		Street:        u.Street,
		PostalCode:    u.PostalCode,
		Lat:           u.Lat,
		Lng:           u.Lng,
		IsDefault:     u.IsDefault,
	})
	if err != nil {
		return domain.Address{}, err
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
		return domain.Address{}, err
	}
	return toDomainAddress(addressByID), nil

}

func (r *AddressRepo) GetListOfAddresses(ctx context.Context, userID uuid.UUID) ([]domain.Address, error) {
	addresses, err := r.q.ListAddressesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return toDomainAddresses(addresses), err

}

func (r *AddressRepo) Delete(ctx context.Context, id uuid.UUID) (int64, error) {
	address, err := r.q.SoftDeleteAddress(ctx, sqlc.SoftDeleteAddressParams{
		ID:     id,
		UserID: id,
	})
	if err != nil {
		return -1, err
	}

	return address, err

}

func (r *AddressRepo) Update(ctx context.Context, u domain.Address) (domain.Address, error) {

	row, err := r.q.UpdateAddress(ctx, sqlc.UpdateAddressParams{
		Label:         u.Label,
		RecipientName: u.RecipientName,
		Phone:         u.Phone,
		Country:       u.Country,
		Province:      "",
		City:          u.City,
		Street:        u.Street,
		PostalCode:    u.PostalCode,
		Lat:           u.Lat,
		Lng:           u.Lng,
		ID:            u.ID,
		UserID:        u.UserID,
	})
	if err != nil {
		return domain.Address{}, err
	}

	return toDomainAddress(row), nil
}

func toDomainAddress(u sqlc.Address) domain.Address {
	return domain.Address{
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
	}
}

func toDomainAddresses(rows []sqlc.Address) []domain.Address {
	out := make([]domain.Address, len(rows))
	for i, row := range rows {
		out[i] = toDomainAddress(row)
	}
	return out
}
