package domain

import "context"

type RoomStorage interface {
	Create(ctx context.Context, room Room) (string, error)
	GetRoom(ctx context.Context, id string) (*Room, error)
	UpdateRoom(ctx context.Context, id string, room Room) error
}
