package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/entity"
	"github.com/nathakusuma/astungkara/domain/enum"
	"github.com/nathakusuma/astungkara/domain/errorpkg"
	"github.com/nathakusuma/astungkara/internal/app/conference/service"
	appmocks "github.com/nathakusuma/astungkara/test/unit/mocks/app"
	pkgmocks "github.com/nathakusuma/astungkara/test/unit/mocks/pkg"
	_ "github.com/nathakusuma/astungkara/test/unit/setup" // Initialize test environment
	"github.com/stretchr/testify/assert"
)

type conferenceServiceMocks struct {
	conferenceRepo *appmocks.MockIConferenceRepository
	uuid           *pkgmocks.MockIUUID
}

func setupConferenceServiceTest(t *testing.T) (contract.IConferenceService, *conferenceServiceMocks) {
	mocks := &conferenceServiceMocks{
		conferenceRepo: appmocks.NewMockIConferenceRepository(t),
		uuid:           pkgmocks.NewMockIUUID(t),
	}

	svc := service.NewConferenceService(mocks.conferenceRepo, mocks.uuid)

	return svc, mocks
}

func Test_ConferenceService_CreateConferenceProposal(t *testing.T) {
	conferenceID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)
	laterTime := now.Add(48 * time.Hour)

	ctx := context.WithValue(context.Background(), "user.id", userID)

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		req := &dto.CreateConferenceProposalRequest{
			Title:          "Test Conference",
			Description:    "Test Description",
			SpeakerName:    "Test Speaker",
			SpeakerTitle:   "Test Title",
			TargetAudience: "Test Audience",
			Prerequisites:  nil,
			Seats:          100,
			StartsAt:       futureTime,
			EndsAt:         laterTime,
		}

		// Mock checking for active proposals
		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, &dto.GetConferenceQuery{
				Limit:       1,
				HostID:      &userID,
				Status:      enum.ConferencePending,
				IncludePast: false,
				OrderBy:     "created_at",
				Order:       "desc",
			}).
			Return([]entity.Conference{}, dto.LazyLoadResponse{}, nil)

		// Mock checking for time conflicts
		mocks.conferenceRepo.EXPECT().
			GetConferencesConflictingWithTime(ctx, req.StartsAt, req.EndsAt, uuid.Nil).
			Return([]entity.Conference{}, nil)

		// Mock UUID generation
		mocks.uuid.EXPECT().
			NewV7().
			Return(conferenceID, nil)

		// Mock conference creation
		mocks.conferenceRepo.EXPECT().
			CreateConference(ctx, &entity.Conference{
				ID:             conferenceID,
				Title:          req.Title,
				Description:    req.Description,
				SpeakerName:    req.SpeakerName,
				SpeakerTitle:   req.SpeakerTitle,
				TargetAudience: req.TargetAudience,
				Prerequisites:  req.Prerequisites,
				Seats:          req.Seats,
				StartsAt:       req.StartsAt,
				EndsAt:         req.EndsAt,
				HostID:         userID,
				Status:         enum.ConferencePending,
			}).
			Return(nil)

		resultID, err := svc.CreateConferenceProposal(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, conferenceID, resultID)
	})

	t.Run("error - missing user ID in context", func(t *testing.T) {
		svc, _ := setupConferenceServiceTest(t)

		req := &dto.CreateConferenceProposalRequest{}
		resultID, err := svc.CreateConferenceProposal(context.Background(), req)
		assert.Equal(t, uuid.Nil, resultID)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - user has active proposal", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		req := &dto.CreateConferenceProposalRequest{
			StartsAt: futureTime,
			EndsAt:   laterTime,
		}

		existingConference := entity.Conference{
			ID:        uuid.New(),
			Title:     "Existing Conference",
			Status:    enum.ConferencePending,
			CreatedAt: time.Now(),
		}

		// Mock checking for active proposals
		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, &dto.GetConferenceQuery{
				Limit:       1,
				HostID:      &userID,
				Status:      enum.ConferencePending,
				IncludePast: false,
				OrderBy:     "created_at",
				Order:       "desc",
			}).
			Return([]entity.Conference{existingConference}, dto.LazyLoadResponse{}, nil)

		resultID, err := svc.CreateConferenceProposal(ctx, req)
		assert.Equal(t, uuid.Nil, resultID)
		assert.ErrorIs(t, err, errorpkg.ErrUserHasActiveProposal)
	})

	t.Run("error - time window conflict", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		req := &dto.CreateConferenceProposalRequest{
			StartsAt: futureTime,
			EndsAt:   laterTime,
		}

		// Mock checking for active proposals
		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, &dto.GetConferenceQuery{
				Limit:       1,
				HostID:      &userID,
				Status:      enum.ConferencePending,
				IncludePast: false,
				OrderBy:     "created_at",
				Order:       "desc",
			}).
			Return([]entity.Conference{}, dto.LazyLoadResponse{}, nil)

		conflictingConference := entity.Conference{
			ID:          uuid.New(),
			Description: "Conflicting Conference",
			StartsAt:    futureTime,
			EndsAt:      laterTime,
		}

		// Mock checking for time conflicts
		mocks.conferenceRepo.EXPECT().
			GetConferencesConflictingWithTime(ctx, req.StartsAt, req.EndsAt, uuid.Nil).
			Return([]entity.Conference{conflictingConference}, nil)

		resultID, err := svc.CreateConferenceProposal(ctx, req)
		assert.Equal(t, uuid.Nil, resultID)
		assert.ErrorIs(t, err, errorpkg.ErrTimeWindowConflict)
	})

	t.Run("error - UUID generation fails", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		req := &dto.CreateConferenceProposalRequest{
			StartsAt: futureTime,
			EndsAt:   laterTime,
		}

		// Mock checking for active proposals
		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, &dto.GetConferenceQuery{
				Limit:       1,
				HostID:      &userID,
				Status:      enum.ConferencePending,
				IncludePast: false,
				OrderBy:     "created_at",
				Order:       "desc",
			}).
			Return([]entity.Conference{}, dto.LazyLoadResponse{}, nil)

		// Mock checking for time conflicts
		mocks.conferenceRepo.EXPECT().
			GetConferencesConflictingWithTime(ctx, req.StartsAt, req.EndsAt, uuid.Nil).
			Return([]entity.Conference{}, nil)

		// Mock UUID generation failure
		mocks.uuid.EXPECT().
			NewV7().
			Return(uuid.Nil, errors.New("uuid generation failed"))

		resultID, err := svc.CreateConferenceProposal(ctx, req)
		assert.Equal(t, uuid.Nil, resultID)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - start time in past", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		req := &dto.CreateConferenceProposalRequest{
			StartsAt: now.Add(-24 * time.Hour), // Past time
			EndsAt:   laterTime,
		}

		// Mock checking for active proposals
		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, &dto.GetConferenceQuery{
				Limit:       1,
				HostID:      &userID,
				Status:      enum.ConferencePending,
				IncludePast: false,
				OrderBy:     "created_at",
				Order:       "desc",
			}).
			Return([]entity.Conference{}, dto.LazyLoadResponse{}, nil)

		// Mock checking for time conflicts
		mocks.conferenceRepo.EXPECT().
			GetConferencesConflictingWithTime(ctx, req.StartsAt, req.EndsAt, uuid.Nil).
			Return([]entity.Conference{}, nil)

		// Mock UUID generation
		mocks.uuid.EXPECT().
			NewV7().
			Return(conferenceID, nil)

		resultID, err := svc.CreateConferenceProposal(ctx, req)
		assert.Equal(t, uuid.Nil, resultID)
		assert.ErrorIs(t, err, errorpkg.ErrTimeAlreadyPassed)
	})

	t.Run("error - end time before start time", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		req := &dto.CreateConferenceProposalRequest{
			StartsAt: laterTime,
			EndsAt:   futureTime, // Before start time
		}

		// Mock checking for active proposals
		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, &dto.GetConferenceQuery{
				Limit:       1,
				HostID:      &userID,
				Status:      enum.ConferencePending,
				IncludePast: false,
				OrderBy:     "created_at",
				Order:       "desc",
			}).
			Return([]entity.Conference{}, dto.LazyLoadResponse{}, nil)

		// Mock checking for time conflicts
		mocks.conferenceRepo.EXPECT().
			GetConferencesConflictingWithTime(ctx, req.StartsAt, req.EndsAt, uuid.Nil).
			Return([]entity.Conference{}, nil)

		// Mock UUID generation
		mocks.uuid.EXPECT().
			NewV7().
			Return(conferenceID, nil)

		resultID, err := svc.CreateConferenceProposal(ctx, req)
		assert.Equal(t, uuid.Nil, resultID)
		assert.ErrorIs(t, err, errorpkg.ErrEndTimeBeforeStart)
	})

	t.Run("error - conference creation fails", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		req := &dto.CreateConferenceProposalRequest{
			Title:          "Test Conference",
			Description:    "Test Description",
			SpeakerName:    "Test Speaker",
			SpeakerTitle:   "Test Title",
			TargetAudience: "Test Audience",
			Prerequisites:  nil,
			Seats:          100,
			StartsAt:       futureTime,
			EndsAt:         laterTime,
		}

		// Mock checking for active proposals
		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, &dto.GetConferenceQuery{
				Limit:       1,
				HostID:      &userID,
				Status:      enum.ConferencePending,
				IncludePast: false,
				OrderBy:     "created_at",
				Order:       "desc",
			}).
			Return([]entity.Conference{}, dto.LazyLoadResponse{}, nil)

		// Mock checking for time conflicts
		mocks.conferenceRepo.EXPECT().
			GetConferencesConflictingWithTime(ctx, req.StartsAt, req.EndsAt, uuid.Nil).
			Return([]entity.Conference{}, nil)

		// Mock UUID generation
		mocks.uuid.EXPECT().
			NewV7().
			Return(conferenceID, nil)

		// Mock conference creation failure
		mocks.conferenceRepo.EXPECT().
			CreateConference(ctx, &entity.Conference{
				ID:             conferenceID,
				Title:          req.Title,
				Description:    req.Description,
				SpeakerName:    req.SpeakerName,
				SpeakerTitle:   req.SpeakerTitle,
				TargetAudience: req.TargetAudience,
				Prerequisites:  req.Prerequisites,
				Seats:          req.Seats,
				StartsAt:       req.StartsAt,
				EndsAt:         req.EndsAt,
				HostID:         userID,
				Status:         enum.ConferencePending,
			}).
			Return(errors.New("database error"))

		resultID, err := svc.CreateConferenceProposal(ctx, req)
		assert.Equal(t, uuid.Nil, resultID)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - get conferences fails", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		req := &dto.CreateConferenceProposalRequest{
			StartsAt: futureTime,
			EndsAt:   laterTime,
		}

		// Mock checking for active proposals fails
		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, &dto.GetConferenceQuery{
				Limit:       1,
				HostID:      &userID,
				Status:      enum.ConferencePending,
				IncludePast: false,
				OrderBy:     "created_at",
				Order:       "desc",
			}).
			Return(nil, dto.LazyLoadResponse{}, errors.New("database error"))

		resultID, err := svc.CreateConferenceProposal(ctx, req)
		assert.Equal(t, uuid.Nil, resultID)
		assert.Error(t, err)
	})

	t.Run("error - get conflicting conferences fails", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		req := &dto.CreateConferenceProposalRequest{
			StartsAt: futureTime,
			EndsAt:   laterTime,
		}

		// Mock checking for active proposals
		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, &dto.GetConferenceQuery{
				Limit:       1,
				HostID:      &userID,
				Status:      enum.ConferencePending,
				IncludePast: false,
				OrderBy:     "created_at",
				Order:       "desc",
			}).
			Return([]entity.Conference{}, dto.LazyLoadResponse{}, nil)

		// Mock checking for time conflicts fails
		mocks.conferenceRepo.EXPECT().
			GetConferencesConflictingWithTime(ctx, req.StartsAt, req.EndsAt, uuid.Nil).
			Return(nil, errors.New("database error"))

		resultID, err := svc.CreateConferenceProposal(ctx, req)
		assert.Equal(t, uuid.Nil, resultID)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

func Test_ConferenceService_GetConferenceByID(t *testing.T) {
	ctx := context.Background()
	conferenceID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	conference := &entity.Conference{
		ID:             conferenceID,
		Title:          "Test Conference",
		Description:    "Test Description",
		SpeakerName:    "Test Speaker",
		SpeakerTitle:   "Test Title",
		TargetAudience: "Test Audience",
		Prerequisites:  nil,
		Seats:          100,
		StartsAt:       now.Add(24 * time.Hour),
		EndsAt:         now.Add(48 * time.Hour),
		HostID:         userID,
		Status:         enum.ConferenceApproved,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	t.Run("success - admin role", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)
		ctx := context.WithValue(context.WithValue(ctx, "user.id", userID), "user.role", enum.RoleAdmin)

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		result, err := svc.GetConferenceByID(ctx, conferenceID)
		assert.NoError(t, err)
		assert.Equal(t, conference.ID, result.ID)
		assert.Equal(t, conference.Title, result.Title)
		assert.Equal(t, conference.Status, result.Status)
	})

	t.Run("success - host user", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)
		ctx := context.WithValue(context.WithValue(ctx, "user.id", userID), "user.role", enum.RoleUser)

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		result, err := svc.GetConferenceByID(ctx, conferenceID)
		assert.NoError(t, err)
		assert.Equal(t, conference.ID, result.ID)
	})

	t.Run("success - non-host user accessing approved conference", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)
		otherUserID := uuid.New()
		ctx := context.WithValue(context.WithValue(ctx, "user.id", otherUserID), "user.role", enum.RoleUser)

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		result, err := svc.GetConferenceByID(ctx, conferenceID)
		assert.NoError(t, err)
		assert.Equal(t, conference.ID, result.ID)
	})

	t.Run("error - non-host user accessing pending conference", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)
		otherUserID := uuid.New()
		ctx := context.WithValue(context.WithValue(ctx, "user.id", otherUserID), "user.role", enum.RoleUser)

		pendingConference := *conference
		pendingConference.Status = enum.ConferencePending

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(&pendingConference, nil)

		result, err := svc.GetConferenceByID(ctx, conferenceID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, errorpkg.ErrForbiddenUser)
	})

	t.Run("error - conference not found", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)
		ctx := context.WithValue(context.WithValue(ctx, "user.id", userID), "user.role", enum.RoleUser)

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(nil, sql.ErrNoRows)

		result, err := svc.GetConferenceByID(ctx, conferenceID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
	})

	t.Run("error - repository error", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)
		ctx := context.WithValue(context.WithValue(ctx, "user.id", userID), "user.role", enum.RoleUser)

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(nil, errors.New("repository error"))

		result, err := svc.GetConferenceByID(ctx, conferenceID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

func Test_ConferenceService_GetConferences(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()

	t.Run("success - user role viewing approved conferences", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		ctx = context.WithValue(ctx, "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleUser)

		query := &dto.GetConferenceQuery{
			Limit:  10,
			Status: enum.ConferenceApproved,
		}

		now := time.Now()
		conferences := []entity.Conference{
			{
				ID:        uuid.New(),
				Title:     "Test Conference 1",
				Status:    enum.ConferenceApproved,
				StartsAt:  now,
				EndsAt:    now.Add(time.Hour),
				CreatedAt: now,
				UpdatedAt: now,
				Host: entity.User{
					ID: userID,
				},
			},
		}

		expectedResponse := []dto.ConferenceResponse{
			{
				ID:         conferences[0].ID,
				Title:      conferences[0].Title,
				Status:     conferences[0].Status,
				StartsAt:   &conferences[0].StartsAt,
				EndsAt:     &conferences[0].EndsAt,
				CreatedAt:  &conferences[0].CreatedAt,
				UpdatedAt:  &conferences[0].UpdatedAt,
				SeatsTaken: new(int),
				Host: &dto.UserResponse{
					ID: userID,
				},
			},
		}

		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, query).
			Return(conferences, dto.LazyLoadResponse{HasMore: false}, nil)

		result, lazy, err := svc.GetConferences(ctx, query)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
		assert.Equal(t, dto.LazyLoadResponse{HasMore: false}, lazy)
	})

	t.Run("success - admin viewing all conferences", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		ctx = context.WithValue(ctx, "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleAdmin)

		query := &dto.GetConferenceQuery{
			Limit:  10,
			Status: enum.ConferencePending,
		}

		conferences := []entity.Conference{
			{
				ID:     uuid.New(),
				Title:  "Test Conference 1",
				Status: enum.ConferencePending,
				Host: entity.User{
					ID: otherUserID,
				},
			},
		}

		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, query).
			Return(conferences, dto.LazyLoadResponse{HasMore: false}, nil)

		result, _, err := svc.GetConferences(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("error - pagination includes beforeID and afterID", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		ctx = context.WithValue(ctx, "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleUser)

		query := &dto.GetConferenceQuery{
			Limit:    10,
			Status:   enum.ConferenceApproved,
			BeforeID: &uuid.Nil,
			AfterID:  &uuid.Nil,
		}

		result, lazy, err := svc.GetConferences(ctx, query)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidPagination)
		assert.Empty(t, result)
		assert.Equal(t, dto.LazyLoadResponse{}, lazy)

		// Verify that the repository was not called
		mocks.conferenceRepo.AssertNotCalled(t, "GetConferences")
	})

	t.Run("error - user trying to view other's pending conferences", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		ctx = context.WithValue(ctx, "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleUser)

		query := &dto.GetConferenceQuery{
			Limit:  10,
			Status: enum.ConferencePending,
			HostID: &otherUserID,
		}

		result, lazy, err := svc.GetConferences(ctx, query)
		assert.ErrorIs(t, err, errorpkg.ErrForbiddenUser)
		assert.Empty(t, result)
		assert.Equal(t, dto.LazyLoadResponse{}, lazy)

		// Verify that the repository was not called
		mocks.conferenceRepo.AssertNotCalled(t, "GetConferences")
	})

	t.Run("error - repository error", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		ctx = context.WithValue(ctx, "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleUser)

		query := &dto.GetConferenceQuery{
			Limit:  10,
			Status: enum.ConferenceApproved,
		}

		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, query).
			Return(nil, dto.LazyLoadResponse{}, errors.New("repository error"))

		result, lazy, err := svc.GetConferences(ctx, query)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
		assert.Empty(t, result)
		assert.Equal(t, dto.LazyLoadResponse{}, lazy)
	})

	t.Run("success - system role (no role in context)", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		// Don't set role in context
		ctx = context.WithValue(ctx, "user.id", userID)

		query := &dto.GetConferenceQuery{
			Limit:  10,
			Status: enum.ConferencePending,
		}

		conferences := []entity.Conference{
			{
				ID:     uuid.New(),
				Title:  "Test Conference 1",
				Status: enum.ConferencePending,
			},
		}

		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, query).
			Return(conferences, dto.LazyLoadResponse{HasMore: false}, nil)

		result, _, err := svc.GetConferences(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("success - user viewing own pending conferences", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		ctx = context.WithValue(ctx, "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleUser)

		query := &dto.GetConferenceQuery{
			Limit:  10,
			Status: enum.ConferencePending,
			HostID: &userID,
		}

		conferences := []entity.Conference{
			{
				ID:     uuid.New(),
				Title:  "Test Conference 1",
				Status: enum.ConferencePending,
				Host: entity.User{
					ID: userID,
				},
			},
		}

		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, query).
			Return(conferences, dto.LazyLoadResponse{HasMore: false}, nil)

		result, _, err := svc.GetConferences(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("success - user viewing pending conferences without hostID", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		ctx = context.WithValue(ctx, "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleUser)

		query := &dto.GetConferenceQuery{
			Limit:  10,
			Status: enum.ConferencePending,
		}

		// The service should automatically set HostID to the requester's ID
		expectedQuery := &dto.GetConferenceQuery{
			Limit:  10,
			Status: enum.ConferencePending,
			HostID: &userID,
		}

		conferences := []entity.Conference{
			{
				ID:     uuid.New(),
				Title:  "Test Conference 1",
				Status: enum.ConferencePending,
				Host: entity.User{
					ID: userID,
				},
			},
		}

		mocks.conferenceRepo.EXPECT().
			GetConferences(ctx, expectedQuery).
			Return(conferences, dto.LazyLoadResponse{HasMore: false}, nil)

		result, _, err := svc.GetConferences(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

func Test_ConferenceService_UpdateConference(t *testing.T) {
	conferenceID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)
	ctx := context.WithValue(context.Background(), "user.id", userID)
	ctx = context.WithValue(ctx, "user.role", enum.RoleUser)

	t.Run("success - update basic info", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		originalConference := &entity.Conference{
			ID:       conferenceID,
			HostID:   userID,
			StartsAt: futureTime,
			EndsAt:   futureTime.Add(time.Hour),
			Status:   enum.ConferencePending,
		}

		newTitle := "Updated Title"
		req := dto.UpdateConferenceRequest{
			Title: &newTitle,
		}

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(originalConference, nil)

		updatedConference := *originalConference
		updatedConference.Title = newTitle

		mocks.conferenceRepo.EXPECT().
			UpdateConference(ctx, &updatedConference).
			Return(nil)

		err := svc.UpdateConference(ctx, conferenceID, req)
		assert.NoError(t, err)
	})

	t.Run("error - conference not found", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(nil, sql.ErrNoRows)

		err := svc.UpdateConference(ctx, conferenceID, dto.UpdateConferenceRequest{})
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
	})

	t.Run("error - forbidden user", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		differentUserID := uuid.New()
		conference := &entity.Conference{
			ID:     conferenceID,
			HostID: differentUserID,
		}

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		err := svc.UpdateConference(ctx, conferenceID, dto.UpdateConferenceRequest{})
		assert.ErrorIs(t, err, errorpkg.ErrForbiddenUser)
	})

	t.Run("error - update past conference", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		pastTime := now.Add(-24 * time.Hour)
		conference := &entity.Conference{
			ID:       conferenceID,
			HostID:   userID,
			StartsAt: pastTime,
			EndsAt:   pastTime.Add(time.Hour),
		}

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		err := svc.UpdateConference(ctx, conferenceID, dto.UpdateConferenceRequest{})
		assert.ErrorIs(t, err, errorpkg.ErrUpdatePastConference)
	})

	t.Run("error - update rejected conference", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		conference := &entity.Conference{
			ID:       conferenceID,
			HostID:   userID,
			StartsAt: futureTime,
			EndsAt:   futureTime.Add(time.Hour),
			Status:   enum.ConferenceRejected,
		}

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		err := svc.UpdateConference(ctx, conferenceID, dto.UpdateConferenceRequest{})
		assert.ErrorIs(t, err, errorpkg.ErrUpdateNotPendingConference)
	})

	t.Run("error - time already passed", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		conference := &entity.Conference{
			ID:       conferenceID,
			HostID:   userID,
			StartsAt: futureTime,
			EndsAt:   futureTime.Add(time.Hour),
			Status:   enum.ConferencePending,
		}

		pastTime := now.Add(-1 * time.Hour)
		req := dto.UpdateConferenceRequest{
			StartsAt: &pastTime,
			EndsAt:   &futureTime,
		}

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		err := svc.UpdateConference(ctx, conferenceID, req)
		assert.ErrorIs(t, err, errorpkg.ErrTimeAlreadyPassed)
	})

	t.Run("error - end time before start time", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		conference := &entity.Conference{
			ID:       conferenceID,
			HostID:   userID,
			StartsAt: futureTime,
			EndsAt:   futureTime.Add(time.Hour),
			Status:   enum.ConferencePending,
		}

		endTime := futureTime
		startTime := endTime.Add(time.Hour)
		req := dto.UpdateConferenceRequest{
			StartsAt: &startTime,
			EndsAt:   &endTime,
		}

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		err := svc.UpdateConference(ctx, conferenceID, req)
		assert.ErrorIs(t, err, errorpkg.ErrEndTimeBeforeStart)
	})

	t.Run("error - time window conflict", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		conference := &entity.Conference{
			ID:       conferenceID,
			HostID:   userID,
			StartsAt: futureTime,
			EndsAt:   futureTime.Add(time.Hour),
			Status:   enum.ConferencePending,
		}

		newStart := futureTime.Add(2 * time.Hour)
		newEnd := newStart.Add(time.Hour)
		req := dto.UpdateConferenceRequest{
			StartsAt: &newStart,
			EndsAt:   &newEnd,
		}

		conflictingConference := entity.Conference{
			ID:          uuid.New(),
			Title:       "Conflicting Conference",
			Description: "Conflict",
			StartsAt:    newStart,
			EndsAt:      newEnd,
		}

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		mocks.conferenceRepo.EXPECT().
			GetConferencesConflictingWithTime(ctx, newStart, newEnd, conferenceID).
			Return([]entity.Conference{conflictingConference}, nil)

		err := svc.UpdateConference(ctx, conferenceID, req)
		assert.ErrorIs(t, err, errorpkg.ErrTimeWindowConflict)
	})

	t.Run("error - internal server error on get conference", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(nil, errors.New("database error"))

		err := svc.UpdateConference(ctx, conferenceID, dto.UpdateConferenceRequest{})
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - internal server error on update", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		conference := &entity.Conference{
			ID:       conferenceID,
			HostID:   userID,
			StartsAt: futureTime,
			EndsAt:   futureTime.Add(time.Hour),
			Status:   enum.ConferencePending,
		}

		newTitle := "Updated Title"
		req := dto.UpdateConferenceRequest{
			Title: &newTitle,
		}

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		updatedConference := *conference
		updatedConference.Title = newTitle

		mocks.conferenceRepo.EXPECT().
			UpdateConference(ctx, &updatedConference).
			Return(errors.New("database error"))

		err := svc.UpdateConference(ctx, conferenceID, req)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - internal server error on get conflicting conferences", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		conference := &entity.Conference{
			ID:       conferenceID,
			HostID:   userID,
			StartsAt: futureTime,
			EndsAt:   futureTime.Add(time.Hour),
			Status:   enum.ConferencePending,
		}

		newStart := futureTime.Add(2 * time.Hour)
		newEnd := newStart.Add(time.Hour)
		req := dto.UpdateConferenceRequest{
			StartsAt: &newStart,
			EndsAt:   &newEnd,
		}

		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		mocks.conferenceRepo.EXPECT().
			GetConferencesConflictingWithTime(ctx, newStart, newEnd, conferenceID).
			Return(nil, errors.New("database error"))

		err := svc.UpdateConference(ctx, conferenceID, req)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

func Test_ConferenceService_DeleteConference(t *testing.T) {
	conferenceID := uuid.New()
	userID := uuid.New()

	t.Run("success - admin user", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		// Set admin context
		ctx := context.WithValue(context.WithValue(context.Background(),
			"user.id", userID),
			"user.role", enum.RoleAdmin)

		conference := &entity.Conference{
			ID:     conferenceID,
			HostID: uuid.New(), // Different from requester
			Status: enum.ConferenceApproved,
		}

		// Expect conference retrieval
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		// Expect conference deletion
		mocks.conferenceRepo.EXPECT().
			DeleteConference(ctx, conferenceID).
			Return(nil)

		err := svc.DeleteConference(ctx, conferenceID)
		assert.NoError(t, err)
	})

	t.Run("success - conference host", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		// Set user context
		ctx := context.WithValue(context.WithValue(context.Background(),
			"user.id", userID),
			"user.role", enum.RoleUser)

		conference := &entity.Conference{
			ID:     conferenceID,
			HostID: userID, // Same as requester
			Status: enum.ConferenceApproved,
		}

		// Expect conference retrieval
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		// Expect conference deletion
		mocks.conferenceRepo.EXPECT().
			DeleteConference(ctx, conferenceID).
			Return(nil)

		err := svc.DeleteConference(ctx, conferenceID)
		assert.NoError(t, err)
	})

	t.Run("error - conference not found", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		// Set user context
		ctx := context.WithValue(context.WithValue(context.Background(),
			"user.id", userID),
			"user.role", enum.RoleUser)

		// Expect conference retrieval to return not found
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(nil, sql.ErrNoRows)

		err := svc.DeleteConference(ctx, conferenceID)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
	})

	t.Run("error - forbidden user", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		// Set user context
		ctx := context.WithValue(context.WithValue(context.Background(),
			"user.id", userID),
			"user.role", enum.RoleUser)

		conference := &entity.Conference{
			ID:     conferenceID,
			HostID: uuid.New(), // Different from requester
			Status: enum.ConferenceApproved,
		}

		// Expect conference retrieval
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		err := svc.DeleteConference(ctx, conferenceID)
		assert.ErrorIs(t, err, errorpkg.ErrForbiddenUser)
	})

	t.Run("error - get conference db error", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		// Set user context
		ctx := context.WithValue(context.WithValue(context.Background(),
			"user.id", userID),
			"user.role", enum.RoleUser)

		// Expect conference retrieval to return error
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(nil, errors.New("db error"))

		err := svc.DeleteConference(ctx, conferenceID)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - delete conference db error", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		// Set user context
		ctx := context.WithValue(context.WithValue(context.Background(),
			"user.id", userID),
			"user.role", enum.RoleUser)

		conference := &entity.Conference{
			ID:     conferenceID,
			HostID: userID, // Same as requester
			Status: enum.ConferenceApproved,
		}

		// Expect conference retrieval
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		// Expect conference deletion to fail
		mocks.conferenceRepo.EXPECT().
			DeleteConference(ctx, conferenceID).
			Return(errors.New("db error"))

		err := svc.DeleteConference(ctx, conferenceID)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

func Test_ConferenceService_UpdateConferenceStatus(t *testing.T) {
	ctx := context.Background()
	conferenceID := uuid.New()
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)

	t.Run("success - approve conference", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		conference := &entity.Conference{
			ID:       conferenceID,
			StartsAt: futureTime,
			EndsAt:   futureTime.Add(time.Hour),
			Status:   enum.ConferencePending,
		}

		// Expect conference retrieval
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		// Expect conflict check
		mocks.conferenceRepo.EXPECT().
			GetConferencesConflictingWithTime(ctx, conference.StartsAt, conference.EndsAt, conferenceID).
			Return([]entity.Conference{}, nil)

		// Expect conference update
		mocks.conferenceRepo.EXPECT().
			UpdateConference(ctx, conference).
			Return(nil)

		err := svc.UpdateConferenceStatus(ctx, conferenceID, enum.ConferenceApproved)
		assert.NoError(t, err)
		assert.Equal(t, enum.ConferenceApproved, conference.Status)
	})

	t.Run("success - reject conference", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		conference := &entity.Conference{
			ID:       conferenceID,
			StartsAt: now.Add(-time.Hour), // Past conference can be rejected
			EndsAt:   now,
			Status:   enum.ConferencePending,
		}

		// Expect conference retrieval
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		// Expect conference update
		mocks.conferenceRepo.EXPECT().
			UpdateConference(ctx, conference).
			Return(nil)

		err := svc.UpdateConferenceStatus(ctx, conferenceID, enum.ConferenceRejected)
		assert.NoError(t, err)
		assert.Equal(t, enum.ConferenceRejected, conference.Status)
	})

	t.Run("error - conference not found", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		// Expect conference retrieval to fail
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(nil, sql.ErrNoRows)

		err := svc.UpdateConferenceStatus(ctx, conferenceID, enum.ConferenceApproved)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
	})

	t.Run("error - database error on get conference", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		// Expect conference retrieval to fail
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(nil, errors.New("database error"))

		err := svc.UpdateConferenceStatus(ctx, conferenceID, enum.ConferenceApproved)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - conference status not pending", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		conference := &entity.Conference{
			ID:       conferenceID,
			StartsAt: futureTime,
			EndsAt:   futureTime.Add(time.Hour),
			Status:   enum.ConferenceApproved,
		}

		// Expect conference retrieval
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		err := svc.UpdateConferenceStatus(ctx, conferenceID, enum.ConferenceRejected)
		assert.ErrorIs(t, err, errorpkg.ErrUpdateNotPendingConference)
	})

	t.Run("error - past conference approval", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		conference := &entity.Conference{
			ID:       conferenceID,
			StartsAt: now.Add(-time.Hour), // Past conference
			EndsAt:   now,
			Status:   enum.ConferencePending,
		}

		// Expect conference retrieval
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		err := svc.UpdateConferenceStatus(ctx, conferenceID, enum.ConferenceApproved)
		assert.ErrorIs(t, err, errorpkg.ErrUpdatePastConferenceStatus)
	})

	t.Run("error - time window conflict", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		conference := &entity.Conference{
			ID:       conferenceID,
			StartsAt: futureTime,
			EndsAt:   futureTime.Add(time.Hour),
			Status:   enum.ConferencePending,
		}

		conflictingConference := entity.Conference{
			ID:          uuid.New(),
			Description: "Conflicting Conference",
			StartsAt:    futureTime,
			EndsAt:      futureTime.Add(time.Hour),
		}

		// Expect conference retrieval
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		// Expect conflict check to find conflicts
		mocks.conferenceRepo.EXPECT().
			GetConferencesConflictingWithTime(ctx, conference.StartsAt, conference.EndsAt, conferenceID).
			Return([]entity.Conference{conflictingConference}, nil)

		err := svc.UpdateConferenceStatus(ctx, conferenceID, enum.ConferenceApproved)
		assert.ErrorIs(t, err, errorpkg.ErrTimeWindowConflict)
	})

	t.Run("error - database error on conflict check", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		conference := &entity.Conference{
			ID:       conferenceID,
			StartsAt: futureTime,
			EndsAt:   futureTime.Add(time.Hour),
			Status:   enum.ConferencePending,
		}

		// Expect conference retrieval
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		// Expect conflict check to fail
		mocks.conferenceRepo.EXPECT().
			GetConferencesConflictingWithTime(ctx, conference.StartsAt, conference.EndsAt, conferenceID).
			Return(nil, errors.New("database error"))

		err := svc.UpdateConferenceStatus(ctx, conferenceID, enum.ConferenceApproved)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - database error on update", func(t *testing.T) {
		svc, mocks := setupConferenceServiceTest(t)

		conference := &entity.Conference{
			ID:       conferenceID,
			StartsAt: futureTime,
			EndsAt:   futureTime.Add(time.Hour),
			Status:   enum.ConferencePending,
		}

		// Expect conference retrieval
		mocks.conferenceRepo.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		// Expect conflict check
		mocks.conferenceRepo.EXPECT().
			GetConferencesConflictingWithTime(ctx, conference.StartsAt, conference.EndsAt, conferenceID).
			Return([]entity.Conference{}, nil)

		// Expect update to fail
		mocks.conferenceRepo.EXPECT().
			UpdateConference(ctx, conference).
			Return(errors.New("database error"))

		err := svc.UpdateConferenceStatus(ctx, conferenceID, enum.ConferenceApproved)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}
