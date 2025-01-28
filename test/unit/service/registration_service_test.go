package service

import (
	"context"
	"github.com/nathakusuma/astungkara/domain/enum"
	"github.com/nathakusuma/astungkara/internal/app/registration/service"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/entity"
	"github.com/nathakusuma/astungkara/domain/errorpkg"
	appmocks "github.com/nathakusuma/astungkara/test/unit/mocks/app"
	_ "github.com/nathakusuma/astungkara/test/unit/setup"
	"github.com/stretchr/testify/assert"
)

type registrationServiceMocks struct {
	registrationRepo *appmocks.MockIRegistrationRepository
	conferenceSvc    *appmocks.MockIConferenceService
}

func setupRegistrationServiceTest(t *testing.T) (contract.IRegistrationService, *registrationServiceMocks) {
	mocks := &registrationServiceMocks{
		registrationRepo: appmocks.NewMockIRegistrationRepository(t),
		conferenceSvc:    appmocks.NewMockIConferenceService(t),
	}

	svc := service.NewRegistrationService(mocks.registrationRepo, mocks.conferenceSvc)

	return svc, mocks
}

func Test_RegistrationService_Register(t *testing.T) {
	conferenceID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		conference := &dto.ConferenceResponse{
			ID: conferenceID,
			Host: &dto.UserResponse{
				ID: hostID,
			},
			Seats:    100,
			StartsAt: &now,
			EndsAt:   &futureTime,
		}

		// Mock getting conference
		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		// Mock checking if user is registered
		mocks.registrationRepo.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(false, nil)

		// Mock counting registrations
		mocks.registrationRepo.EXPECT().
			CountRegistrationsByConference(ctx, conferenceID).
			Return(50, nil)

		// Mock checking for conflicting registrations
		mocks.registrationRepo.EXPECT().
			GetConflictingRegistrations(ctx, userID, *conference.StartsAt, *conference.EndsAt).
			Return([]entity.Conference{}, nil)

		// Mock creating registration
		mocks.registrationRepo.EXPECT().
			CreateRegistration(ctx, &entity.Registration{
				ConferenceID: conferenceID,
				UserID:       userID,
			}).
			Return(nil)

		err := svc.Register(ctx, conferenceID, userID)
		assert.NoError(t, err)
	})

	t.Run("error - conference not found", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(nil, errorpkg.ErrNotFound)

		err := svc.Register(ctx, conferenceID, userID)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
	})

	t.Run("error - host attempting to register", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		conference := &dto.ConferenceResponse{
			ID: conferenceID,
			Host: &dto.UserResponse{
				ID: userID, // Host is the same as registering user
			},
			StartsAt: &now,
			EndsAt:   &futureTime,
		}

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		err := svc.Register(ctx, conferenceID, userID)
		assert.ErrorIs(t, err, errorpkg.ErrHostCannotRegister)
	})

	t.Run("error - conference ended", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		pastTime := now.Add(-48 * time.Hour)
		conference := &dto.ConferenceResponse{
			ID: conferenceID,
			Host: &dto.UserResponse{
				ID: hostID,
			},
			StartsAt: &pastTime,
			EndsAt:   &now,
		}

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		err := svc.Register(ctx, conferenceID, userID)
		assert.ErrorIs(t, err, errorpkg.ErrConferenceEnded)
	})

	t.Run("error - already registered", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		conference := &dto.ConferenceResponse{
			ID: conferenceID,
			Host: &dto.UserResponse{
				ID: hostID,
			},
			StartsAt: &now,
			EndsAt:   &futureTime,
		}

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		mocks.registrationRepo.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(true, nil)

		err := svc.Register(ctx, conferenceID, userID)
		assert.ErrorIs(t, err, errorpkg.ErrUserAlreadyRegisteredToConference)
	})

	t.Run("error - conference full", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		conference := &dto.ConferenceResponse{
			ID: conferenceID,
			Host: &dto.UserResponse{
				ID: hostID,
			},
			Seats:    100,
			StartsAt: &now,
			EndsAt:   &futureTime,
		}

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		mocks.registrationRepo.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(false, nil)

		mocks.registrationRepo.EXPECT().
			CountRegistrationsByConference(ctx, conferenceID).
			Return(100, nil) // Full capacity

		err := svc.Register(ctx, conferenceID, userID)
		assert.ErrorIs(t, err, errorpkg.ErrConferenceFull)
	})

	t.Run("error - conflicting registrations", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		conference := &dto.ConferenceResponse{
			ID: conferenceID,
			Host: &dto.UserResponse{
				ID: hostID,
			},
			Seats:    100,
			StartsAt: &now,
			EndsAt:   &futureTime,
		}

		conflictingConference := entity.Conference{
			ID:    uuid.New(),
			Title: "Conflicting Conference",
		}

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		mocks.registrationRepo.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(false, nil)

		mocks.registrationRepo.EXPECT().
			CountRegistrationsByConference(ctx, conferenceID).
			Return(50, nil)

		mocks.registrationRepo.EXPECT().
			GetConflictingRegistrations(ctx, userID, *conference.StartsAt, *conference.EndsAt).
			Return([]entity.Conference{conflictingConference}, nil)

		err := svc.Register(ctx, conferenceID, userID)
		assert.ErrorIs(t, err, errorpkg.ErrConflictingRegistrations)
	})

	t.Run("error - internal server error on registration check", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		conference := &dto.ConferenceResponse{
			ID: conferenceID,
			Host: &dto.UserResponse{
				ID: hostID,
			},
			StartsAt: &now,
			EndsAt:   &futureTime,
		}

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		mocks.registrationRepo.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(false, errorpkg.ErrInternalServer)

		err := svc.Register(ctx, conferenceID, userID)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - internal server error on count registrations", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		conference := &dto.ConferenceResponse{
			ID: conferenceID,
			Host: &dto.UserResponse{
				ID: hostID,
			},
			StartsAt: &now,
			EndsAt:   &futureTime,
		}

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		mocks.registrationRepo.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(false, nil)

		mocks.registrationRepo.EXPECT().
			CountRegistrationsByConference(ctx, conferenceID).
			Return(0, errorpkg.ErrInternalServer)

		err := svc.Register(ctx, conferenceID, userID)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - internal server error on get conflicting registrations", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		conference := &dto.ConferenceResponse{
			ID: conferenceID,
			Host: &dto.UserResponse{
				ID: hostID,
			},
			Seats:    100,
			StartsAt: &now,
			EndsAt:   &futureTime,
		}

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		mocks.registrationRepo.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(false, nil)

		mocks.registrationRepo.EXPECT().
			CountRegistrationsByConference(ctx, conferenceID).
			Return(50, nil)

		mocks.registrationRepo.EXPECT().
			GetConflictingRegistrations(ctx, userID, *conference.StartsAt, *conference.EndsAt).
			Return(nil, errorpkg.ErrInternalServer)

		err := svc.Register(ctx, conferenceID, userID)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - internal server error on create registration", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		conference := &dto.ConferenceResponse{
			ID: conferenceID,
			Host: &dto.UserResponse{
				ID: hostID,
			},
			Seats:    100,
			StartsAt: &now,
			EndsAt:   &futureTime,
		}

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		mocks.registrationRepo.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(false, nil)

		mocks.registrationRepo.EXPECT().
			CountRegistrationsByConference(ctx, conferenceID).
			Return(50, nil)

		mocks.registrationRepo.EXPECT().
			GetConflictingRegistrations(ctx, userID, *conference.StartsAt, *conference.EndsAt).
			Return([]entity.Conference{}, nil)

		mocks.registrationRepo.EXPECT().
			CreateRegistration(ctx, &entity.Registration{
				ConferenceID: conferenceID,
				UserID:       userID,
			}).
			Return(errorpkg.ErrInternalServer)

		err := svc.Register(ctx, conferenceID, userID)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

func Test_RegistrationService_GetRegisteredUsersByConference(t *testing.T) {
	conferenceID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()

	t.Run("success - admin role", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		ctx := context.WithValue(context.Background(), "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleAdmin)

		lazyReq := dto.LazyLoadQuery{
			Limit: 10,
		}

		conference := &dto.ConferenceResponse{
			ID: conferenceID,
			Host: &dto.UserResponse{
				ID: hostID,
			},
		}

		users := []entity.User{
			{
				ID:    uuid.New(),
				Name:  "User 1",
				Email: "user1@example.com",
			},
			{
				ID:    uuid.New(),
				Name:  "User 2",
				Email: "user2@example.com",
			},
		}

		lazyResp := dto.LazyLoadResponse{
			HasMore: false,
			LastID:  users[1].ID,
		}

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		mocks.registrationRepo.EXPECT().
			GetRegisteredUsersByConference(ctx, conferenceID, lazyReq).
			Return(users, lazyResp, nil)

		result, resultLazy, err := svc.GetRegisteredUsersByConference(ctx, conferenceID, lazyReq)
		assert.NoError(t, err)
		assert.Equal(t, len(users), len(result))
		assert.Equal(t, lazyResp, resultLazy)
		assert.Equal(t, users[0].ID, result[0].ID)
		assert.Equal(t, users[0].Name, result[0].Name)
		assert.Equal(t, users[0].Email, result[0].Email)
	})

	t.Run("success - host accessing own conference", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		ctx := context.WithValue(context.Background(), "user.id", hostID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleUser)

		lazyReq := dto.LazyLoadQuery{
			Limit: 10,
		}

		conference := &dto.ConferenceResponse{
			ID: conferenceID,
			Host: &dto.UserResponse{
				ID: hostID,
			},
		}

		users := []entity.User{
			{
				ID:    uuid.New(),
				Name:  "User 1",
				Email: "user1@example.com",
			},
		}

		lazyResp := dto.LazyLoadResponse{
			HasMore: false,
			LastID:  users[0].ID,
		}

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		mocks.registrationRepo.EXPECT().
			GetRegisteredUsersByConference(ctx, conferenceID, lazyReq).
			Return(users, lazyResp, nil)

		result, resultLazy, err := svc.GetRegisteredUsersByConference(ctx, conferenceID, lazyReq)
		assert.NoError(t, err)
		assert.Equal(t, len(users), len(result))
		assert.Equal(t, lazyResp, resultLazy)
	})

	t.Run("error - invalid pagination", func(t *testing.T) {
		svc, _ := setupRegistrationServiceTest(t)

		ctx := context.WithValue(context.Background(), "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleAdmin)

		lazyReq := dto.LazyLoadQuery{
			Limit:    10,
			AfterID:  uuid.New(),
			BeforeID: uuid.New(),
		}

		result, resultLazy, err := svc.GetRegisteredUsersByConference(ctx, conferenceID, lazyReq)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidPagination)
		assert.Empty(t, result)
		assert.Empty(t, resultLazy)
	})

	t.Run("error - conference not found", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		ctx := context.WithValue(context.Background(), "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleAdmin)

		lazyReq := dto.LazyLoadQuery{
			Limit: 10,
		}

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(nil, errorpkg.ErrNotFound)

		result, resultLazy, err := svc.GetRegisteredUsersByConference(ctx, conferenceID, lazyReq)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
		assert.Empty(t, result)
		assert.Empty(t, resultLazy)
	})

	t.Run("error - forbidden user", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		ctx := context.WithValue(context.Background(), "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleUser)

		lazyReq := dto.LazyLoadQuery{
			Limit: 10,
		}

		conference := &dto.ConferenceResponse{
			ID: conferenceID,
			Host: &dto.UserResponse{
				ID: hostID, // Different from requester
			},
		}

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		result, resultLazy, err := svc.GetRegisteredUsersByConference(ctx, conferenceID, lazyReq)
		assert.ErrorIs(t, err, errorpkg.ErrForbiddenUser)
		assert.Empty(t, result)
		assert.Empty(t, resultLazy)
	})

	t.Run("error - internal server error", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		ctx := context.WithValue(context.Background(), "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleAdmin)

		lazyReq := dto.LazyLoadQuery{
			Limit: 10,
		}

		conference := &dto.ConferenceResponse{
			ID: conferenceID,
			Host: &dto.UserResponse{
				ID: hostID,
			},
		}

		mocks.conferenceSvc.EXPECT().
			GetConferenceByID(ctx, conferenceID).
			Return(conference, nil)

		mocks.registrationRepo.EXPECT().
			GetRegisteredUsersByConference(ctx, conferenceID, lazyReq).
			Return(nil, dto.LazyLoadResponse{}, errorpkg.ErrInternalServer)

		result, resultLazy, err := svc.GetRegisteredUsersByConference(ctx, conferenceID, lazyReq)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
		assert.Empty(t, result)
		assert.Empty(t, resultLazy)
	})
}

func Test_RegistrationService_GetRegisteredConferencesByUser(t *testing.T) {
	userID := uuid.New()
	requesterID := uuid.New()

	t.Run("success - admin role", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		ctx := context.WithValue(context.Background(), "user.id", requesterID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleAdmin)

		lazyReq := dto.LazyLoadQuery{
			Limit: 10,
		}

		conferences := []entity.Conference{
			{
				ID:          uuid.New(),
				Title:       "Conference 1",
				Description: "Description 1",
			},
			{
				ID:          uuid.New(),
				Title:       "Conference 2",
				Description: "Description 2",
			},
		}

		lazyResp := dto.LazyLoadResponse{
			HasMore: false,
			LastID:  conferences[1].ID,
		}

		mocks.registrationRepo.EXPECT().
			GetRegisteredConferencesByUser(ctx, userID, true, lazyReq).
			Return(conferences, lazyResp, nil)

		result, resultLazy, err := svc.GetRegisteredConferencesByUser(ctx, userID, true, lazyReq)
		assert.NoError(t, err)
		assert.Equal(t, len(conferences), len(result))
		assert.Equal(t, lazyResp, resultLazy)
		assert.Equal(t, conferences[0].ID, result[0].ID)
		assert.Equal(t, conferences[0].Title, result[0].Title)
		assert.Equal(t, conferences[0].Description, result[0].Description)
	})

	t.Run("success - user accessing own conferences", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		ctx := context.WithValue(context.Background(), "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleUser)

		lazyReq := dto.LazyLoadQuery{
			Limit: 10,
		}

		conferences := []entity.Conference{
			{
				ID:          uuid.New(),
				Title:       "Conference 1",
				Description: "Description 1",
			},
		}

		lazyResp := dto.LazyLoadResponse{
			HasMore: false,
			LastID:  conferences[0].ID,
		}

		mocks.registrationRepo.EXPECT().
			GetRegisteredConferencesByUser(ctx, userID, false, lazyReq).
			Return(conferences, lazyResp, nil)

		result, resultLazy, err := svc.GetRegisteredConferencesByUser(ctx, userID, false, lazyReq)
		assert.NoError(t, err)
		assert.Equal(t, len(conferences), len(result))
		assert.Equal(t, lazyResp, resultLazy)
	})

	t.Run("error - invalid pagination", func(t *testing.T) {
		svc, _ := setupRegistrationServiceTest(t)

		ctx := context.WithValue(context.Background(), "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleUser)

		lazyReq := dto.LazyLoadQuery{
			Limit:    10,
			AfterID:  uuid.New(),
			BeforeID: uuid.New(),
		}

		result, resultLazy, err := svc.GetRegisteredConferencesByUser(ctx, userID, false, lazyReq)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidPagination)
		assert.Empty(t, result)
		assert.Empty(t, resultLazy)
	})

	t.Run("error - forbidden user", func(t *testing.T) {
		svc, _ := setupRegistrationServiceTest(t)

		ctx := context.WithValue(context.Background(), "user.id", requesterID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleUser)

		lazyReq := dto.LazyLoadQuery{
			Limit: 10,
		}

		result, resultLazy, err := svc.GetRegisteredConferencesByUser(ctx, userID, false, lazyReq)
		assert.ErrorIs(t, err, errorpkg.ErrForbiddenUser)
		assert.Empty(t, result)
		assert.Empty(t, resultLazy)
	})

	t.Run("error - internal server error", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		ctx := context.WithValue(context.Background(), "user.id", userID)
		ctx = context.WithValue(ctx, "user.role", enum.RoleUser)

		lazyReq := dto.LazyLoadQuery{
			Limit: 10,
		}

		mocks.registrationRepo.EXPECT().
			GetRegisteredConferencesByUser(ctx, userID, false, lazyReq).
			Return(nil, dto.LazyLoadResponse{}, errorpkg.ErrInternalServer)

		result, resultLazy, err := svc.GetRegisteredConferencesByUser(ctx, userID, false, lazyReq)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
		assert.Empty(t, result)
		assert.Empty(t, resultLazy)
	})
}

func Test_RegistrationService_IsUserRegisteredToConference(t *testing.T) {
	conferenceID := uuid.New()
	userID := uuid.New()

	t.Run("success - registered", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		ctx := context.Background()

		mocks.registrationRepo.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(true, nil)

		ok, err := svc.IsUserRegisteredToConference(ctx, conferenceID, userID)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("success - not registered", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		ctx := context.Background()

		mocks.registrationRepo.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(false, nil)

		ok, err := svc.IsUserRegisteredToConference(ctx, conferenceID, userID)
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("error - internal server error", func(t *testing.T) {
		svc, mocks := setupRegistrationServiceTest(t)

		ctx := context.Background()

		mocks.registrationRepo.EXPECT().
			IsUserRegisteredToConference(ctx, conferenceID, userID).
			Return(false, errorpkg.ErrInternalServer)

		ok, err := svc.IsUserRegisteredToConference(ctx, conferenceID, userID)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
		assert.False(t, ok)
	})
}
