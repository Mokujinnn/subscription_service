package service

import (
	"errors"
	"time"

	"subscription-service/internal/model"
	"subscription-service/internal/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type SubscriptionService struct {
	repo repository.Repository
	log  *logrus.Logger
}

func NewSubscriptionService(repo repository.Repository, log *logrus.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
		log:  log,
	}
}

func (s *SubscriptionService) Create(req model.CreateSubscriptionRequest) (*model.Subscription, error) {
	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		s.log.WithError(err).Error("Invalid start date format")
		return nil, errors.New("invalid start date format, use MM-YYYY")
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsed, err := time.Parse("01-2006", *req.EndDate)
		if err != nil {
			s.log.WithError(err).Error("Invalid end date format")
			return nil, errors.New("invalid end date format, use MM-YYYY")
		}

		if parsed.Before(startDate) {
			return nil, errors.New("end date cannot be before start date")
		}
		endDate = &parsed
	}

	now := time.Now()
	subscription := &model.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(subscription); err != nil {
		return nil, err
	}

	return subscription, nil
}

func (s *SubscriptionService) GetByID(id uuid.UUID) (*model.Subscription, error) {
	return s.repo.GetByID(id)
}

func (s *SubscriptionService) GetAll(filter model.SubscriptionFilter) ([]model.Subscription, error) {
	return s.repo.GetAll(filter)
}

func (s *SubscriptionService) Update(id uuid.UUID, req model.UpdateSubscriptionRequest) error {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("subscription not found")
	}

	if req.EndDate != nil {
		parsed, err := time.Parse("01-2006", *req.EndDate)
		if err != nil {
			s.log.WithError(err).Error("Invalid end date format")
			return errors.New("invalid end date format, use MM-YYYY")
		}

		if parsed.Before(existing.StartDate) {
			return errors.New("end date cannot be before start date")
		}
	}

	return s.repo.Update(id, req)
}

func (s *SubscriptionService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *SubscriptionService) GetTotalCost(filter model.SubscriptionFilter) (int, error) {
	return s.repo.GetTotalCost(filter)
}
