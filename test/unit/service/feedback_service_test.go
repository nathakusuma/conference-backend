package service

import (
	"context"
	"database/sql"
	"github.com/nathakusuma/astungkara/internal/app/feedback/service"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/entity"
	"github.com/nathakusuma/astungkara/domain/errorpkg"
	appmocks "github.com/nathakusuma/astungkara/test/unit/mocks/app"
	pkgmocks "github.com/nathakusuma/astungkara/test/unit/mocks/pkg"
	_ "github.com/nathakusuma/astungkara/test/unit/setup"
	"github.com/stretchr/testify/assert"
)

type feedbackServiceMocks struct {
	feedbackRepo    *appmocks.MockIFeedbackRepository
	registrationSvc *appmocks.MockIRegistrationService
	conferenceSvc   *appmocks.MockIConferenceService
	uuidGen         *pkgmocks.MockIUUID
}

func setupFeedbackServiceTest(t *testing.T) (contract.IFeedbackService, *feedbackServiceMocks) {
	mocks := &feedbackServiceMocks{
		feedbackRepo:    appmocks.NewMockIFeedbackRepository(t),
		registrationSvc: appmocks.NewMockIRegistrationService(t),
		conferenceSvc:   appmocks.NewMockIConferenceService(t),
		uuidGen:         pkgmocks.NewMockIUUID(t),
	}

	svc := service.NewFeedbackService(
		mocks.feedbackRepo,
		mocks.registrationSvc,
		mocks.conferenceSvc,
		mocks.uuidGen,
	)

	return svc, mocks
}

func Test_FeedbackService_CreateFeedback(t *testing.T) {
	userID := uuid.New()
	conferenceID := uuid.New()
	hostID := uuid.New()
	feedbackID := uuid.New()
	comment := "Great conference!"
	pastTime := time.Now().Add(-24 * time.Hour)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)

		// Mock checking if user is registered
		mocks.registrationSvc.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(true, nil)

		// Mock IsFeedbackGiven
		mocks.feedbackRepo.EXPECT().
			IsFeedbackGiven(ctx, userID, conferenceID).
			Return(false, nil)

		// Mock getting conference
		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(&dto.ConferenceResponse{
				ID: conferenceID,
				Host: &dto.UserResponse{
					ID: hostID,
				},
				EndsAt: &pastTime,
			}, nil)

		// Mock UUID generation
		mocks.uuidGen.EXPECT().
			NewV7().
			Return(feedbackID, nil)

		// Mock creating feedback
		mocks.feedbackRepo.EXPECT().
			CreateFeedback(ctx, &entity.Feedback{
				ID:           feedbackID,
				UserID:       userID,
				ConferenceID: conferenceID,
				Comment:      comment,
			}).
			Return(nil)

		id, err := svc.CreateFeedback(ctx, userID, conferenceID, comment)
		assert.NoError(t, err)
		assert.Equal(t, feedbackID, id)
	})

	t.Run("error - user not registered", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)

		mocks.registrationSvc.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(false, nil)

		id, err := svc.CreateFeedback(ctx, userID, conferenceID, comment)
		assert.ErrorIs(t, err, errorpkg.ErrUserNotRegisteredToConference)
		assert.Equal(t, uuid.Nil, id)
	})

	t.Run("error - registration check failed", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)

		mocks.registrationSvc.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(false, errorpkg.ErrInternalServer)

		id, err := svc.CreateFeedback(ctx, userID, conferenceID, comment)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
		assert.Equal(t, uuid.Nil, id)
	})

	t.Run("error - feedback already given", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)

		mocks.registrationSvc.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(true, nil)

		// Mock IsFeedbackGiven
		mocks.feedbackRepo.EXPECT().
			IsFeedbackGiven(ctx, userID, conferenceID).
			Return(true, nil)

		id, err := svc.CreateFeedback(ctx, userID, conferenceID, comment)
		assert.ErrorIs(t, err, errorpkg.ErrFeedbackAlreadyGiven)
		assert.Equal(t, uuid.Nil, id)
	})

	t.Run("error - IsFeedbackGiven failed", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)

		mocks.registrationSvc.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(true, nil)

		// Mock IsFeedbackGiven
		mocks.feedbackRepo.EXPECT().
			IsFeedbackGiven(ctx, userID, conferenceID).
			Return(false, errorpkg.ErrInternalServer)

		id, err := svc.CreateFeedback(ctx, userID, conferenceID, comment)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
		assert.Equal(t, uuid.Nil, id)
	})

	t.Run("error - conference not found", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)

		mocks.registrationSvc.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(true, nil)

		// Mock IsFeedbackGiven
		mocks.feedbackRepo.EXPECT().
			IsFeedbackGiven(ctx, userID, conferenceID).
			Return(false, nil)

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(nil, errorpkg.ErrNotFound)

		id, err := svc.CreateFeedback(ctx, userID, conferenceID, comment)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
		assert.Equal(t, uuid.Nil, id)
	})

	t.Run("error - host cannot give feedback", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)

		mocks.registrationSvc.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(true, nil)

		// Mock IsFeedbackGiven
		mocks.feedbackRepo.EXPECT().
			IsFeedbackGiven(ctx, userID, conferenceID).
			Return(false, nil)

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(&dto.ConferenceResponse{
				ID: conferenceID,
				Host: &dto.UserResponse{
					ID: userID, // User is the host
				},
				EndsAt: &pastTime,
			}, nil)

		id, err := svc.CreateFeedback(ctx, userID, conferenceID, comment)
		assert.ErrorIs(t, err, errorpkg.ErrHostCannotGiveFeedback)
		assert.Equal(t, uuid.Nil, id)
	})

	t.Run("error - conference not ended", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)
		futureTime := time.Now().Add(24 * time.Hour)

		mocks.registrationSvc.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(true, nil)

		// Mock IsFeedbackGiven
		mocks.feedbackRepo.EXPECT().
			IsFeedbackGiven(ctx, userID, conferenceID).
			Return(false, nil)

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(&dto.ConferenceResponse{
				ID: conferenceID,
				Host: &dto.UserResponse{
					ID: hostID,
				},
				EndsAt: &futureTime,
			}, nil)

		id, err := svc.CreateFeedback(ctx, userID, conferenceID, comment)
		assert.ErrorIs(t, err, errorpkg.ErrConferenceNotEnded)
		assert.Equal(t, uuid.Nil, id)
	})

	t.Run("error - uuid generation failed", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)

		mocks.registrationSvc.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(true, nil)

		// Mock IsFeedbackGiven
		mocks.feedbackRepo.EXPECT().
			IsFeedbackGiven(ctx, userID, conferenceID).
			Return(false, nil)

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(&dto.ConferenceResponse{
				ID: conferenceID,
				Host: &dto.UserResponse{
					ID: hostID,
				},
				EndsAt: &pastTime,
			}, nil)

		mocks.uuidGen.EXPECT().
			NewV7().
			Return(uuid.Nil, errorpkg.ErrInternalServer)

		id, err := svc.CreateFeedback(ctx, userID, conferenceID, comment)
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, id)
	})

	t.Run("error - create feedback failed", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)

		mocks.registrationSvc.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(true, nil)

		// Mock IsFeedbackGiven
		mocks.feedbackRepo.EXPECT().
			IsFeedbackGiven(ctx, userID, conferenceID).
			Return(false, nil)

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(&dto.ConferenceResponse{
				ID: conferenceID,
				Host: &dto.UserResponse{
					ID: hostID,
				},
				EndsAt: &pastTime,
			}, nil)

		mocks.uuidGen.EXPECT().
			NewV7().
			Return(feedbackID, nil)

		mocks.feedbackRepo.EXPECT().
			CreateFeedback(ctx, &entity.Feedback{
				ID:           feedbackID,
				UserID:       userID,
				ConferenceID: conferenceID,
				Comment:      comment,
			}).
			Return(errorpkg.ErrInternalServer)

		id, err := svc.CreateFeedback(ctx, userID, conferenceID, comment)
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, id)
	})
}

func Test_FeedbackService_GetFeedbacksByConferenceID(t *testing.T) {
	conferenceID := uuid.New()
	userID := uuid.New()
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)

		feedbacks := []entity.Feedback{
			{
				ID:           uuid.New(),
				UserID:       userID,
				ConferenceID: conferenceID,
				Comment:      "Great conference!",
				CreatedAt:    now,
				User: &entity.User{
					ID:   userID,
					Name: "John Doe",
				},
			},
		}

		feedbackID := uuid.New()
		lazyReq := dto.LazyLoadQuery{
			AfterID:  uuid.Nil,
			BeforeID: uuid.Nil,
			Limit:    10,
		}

		lazyResp := dto.LazyLoadResponse{
			HasMore: true,
			FirstID: feedbackID,
			LastID:  feedbackID,
		}

		mocks.feedbackRepo.EXPECT().
			GetFeedbacksByConferenceID(ctx, conferenceID, lazyReq).
			Return(feedbacks, lazyResp, nil)

		resp, lazy, err := svc.GetFeedbacksByConferenceID(ctx, conferenceID, lazyReq)
		assert.NoError(t, err)
		assert.Equal(t, len(feedbacks), len(resp))
		assert.Equal(t, lazyResp, lazy)
		assert.Equal(t, feedbacks[0].ID, resp[0].ID)
		assert.Equal(t, feedbacks[0].Comment, resp[0].Comment)
		assert.Equal(t, feedbacks[0].User.Name, resp[0].User.Name)
	})

	t.Run("error - repository error", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)

		lazyReq := dto.LazyLoadQuery{
			AfterID:  uuid.Nil,
			BeforeID: uuid.Nil,
			Limit:    10,
		}

		mocks.feedbackRepo.EXPECT().
			GetFeedbacksByConferenceID(ctx, conferenceID, lazyReq).
			Return(nil, dto.LazyLoadResponse{}, errorpkg.ErrInternalServer)

		resp, lazy, err := svc.GetFeedbacksByConferenceID(ctx, conferenceID, lazyReq)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Empty(t, lazy)
	})
}

func Test_FeedbackService_DeleteFeedback(t *testing.T) {
	feedbackID := uuid.New()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)

		mocks.feedbackRepo.EXPECT().
			DeleteFeedback(ctx, feedbackID).
			Return(nil)

		err := svc.DeleteFeedback(ctx, feedbackID)
		assert.NoError(t, err)
	})

	t.Run("error - feedback not found", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)

		mocks.feedbackRepo.EXPECT().
			DeleteFeedback(ctx, feedbackID).
			Return(sql.ErrNoRows)

		err := svc.DeleteFeedback(ctx, feedbackID)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
	})

	t.Run("error - repository error", func(t *testing.T) {
		svc, mocks := setupFeedbackServiceTest(t)

		mocks.feedbackRepo.EXPECT().
			DeleteFeedback(ctx, feedbackID).
			Return(errorpkg.ErrInternalServer)

		err := svc.DeleteFeedback(ctx, feedbackID)
		assert.Error(t, err)
	})
}
