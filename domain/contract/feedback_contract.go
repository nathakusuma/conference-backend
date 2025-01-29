package contract

import (
	"context"
	"github.com/google/uuid"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/entity"
)

type IFeedbackRepository interface {
	CreateFeedback(ctx context.Context, feedback *entity.Feedback) error
	GetFeedbacksByConferenceID(ctx context.Context, conferenceID uuid.UUID,
		lazyReq dto.LazyLoadQuery) ([]entity.Feedback, dto.LazyLoadResponse, error)
	DeleteFeedback(ctx context.Context, id uuid.UUID) error
}

type IFeedbackService interface {
	CreateFeedback(ctx context.Context, userID, conferenceID uuid.UUID, comment string) (uuid.UUID, error)
	GetFeedbacksByConferenceID(ctx context.Context, conferenceID uuid.UUID,
		lazyReq dto.LazyLoadQuery) ([]dto.FeedbackResponse, dto.LazyLoadResponse, error)
	DeleteFeedback(ctx context.Context, id uuid.UUID) error
}
