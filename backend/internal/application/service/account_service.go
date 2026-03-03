package service

import (
	"context"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/google/uuid"
)

type AccountService struct {
	accountRepo  port.AccountRepository
	movementRepo port.MovementRepository
}

func NewAccountService(accountRepo port.AccountRepository, movementRepo port.MovementRepository) *AccountService {
	return &AccountService{
		accountRepo:  accountRepo,
		movementRepo: movementRepo,
	}
}

func (s *AccountService) GetByClientID(ctx context.Context, clientID uuid.UUID) (*model.CurrentAccount, error) {
	return s.accountRepo.FindByClientID(ctx, clientID)
}

func (s *AccountService) GetMovements(ctx context.Context, accountID uuid.UUID, offset, limit int) ([]model.Movement, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.movementRepo.FindByAccountID(ctx, accountID, offset, limit)
}
