package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	eventv1 "github.com/turtlepavlo/proto-contract/gen/go/analytics/event/v1"
	"github.com/turtlepavlo/sso/internal/domain"
	"github.com/turtlepavlo/sso/internal/storage"
	"github.com/turtlepavlo/sso/pkg/provider/google"
)

const createTestTimeout = 1200 * time.Millisecond

var (
	errDBDown2     = errors.New("db down")
	errInsertFail2 = errors.New("insert failed")
)

type createUserRepoMock struct {
	mu sync.Mutex

	getByEmailFn func(ctx context.Context, email string) (*storage.UserLoginModel, error)
	getByLoginFn func(ctx context.Context, login string) (*storage.UserLoginModel, error)
	saveFn       func(ctx context.Context, login *storage.UserLoginModel, info *storage.UserInfoModel) error

	lastEmail string
	lastLogin string
}

func (m *createUserRepoMock) Save(ctx context.Context, login *storage.UserLoginModel, info *storage.UserInfoModel) error {
	fn := m.saveFn
	if fn == nil {
		return errors.New("saveFn not set")
	}
	return fn(ctx, login, info)
}
func (m *createUserRepoMock) GetByEmail(ctx context.Context, email string) (*storage.UserLoginModel, error) {
	m.mu.Lock()
	m.lastEmail = email
	fn := m.getByEmailFn
	m.mu.Unlock()
	if fn == nil {
		return nil, errors.New("getByEmailFn not set")
	}
	return fn(ctx, email)
}
func (m *createUserRepoMock) GetByID(ctx context.Context, id int64) (*storage.UserLoginModel, error) {
	return nil, errors.New("not used")
}
func (m *createUserRepoMock) GetByLogin(ctx context.Context, login string) (*storage.UserLoginModel, error) {
	m.mu.Lock()
	m.lastLogin = login
	fn := m.getByLoginFn
	m.mu.Unlock()
	if fn == nil {
		return nil, errors.New("getByLoginFn not set")
	}
	return fn(ctx, login)
}
func (m *createUserRepoMock) SaveUserGoogle(ctx context.Context, login *storage.UserLoginModel, info *storage.UserInfoModel) (int64, error) {
	return 0, errors.New("not used")
}
func (m *createUserRepoMock) SetGoogleID(ctx context.Context, userID int64, googleID string) error {
	return errors.New("not used")
}
func (m *createUserRepoMock) GetByPhone(ctx context.Context, phone string) (*storage.UserLoginModel, error) {
	return nil, errors.New("not used")
}
func (m *createUserRepoMock) CreateUserByPhone(ctx context.Context, in storage.CreateUserByPhone) (int64, error) {
	return 0, errors.New("not used")
}
func (m *createUserRepoMock) GetByGoogleID(ctx context.Context, googleID string) (*storage.UserLoginModel, error) {
	return nil, errors.New("not used")
}
func (m *createUserRepoMock) UpdatePassHash(ctx context.Context, userID int64, passHash []byte) error {
	return errors.New("not used")
}

type createCacheRepoMock struct{}

func (c *createCacheRepoMock) SetCode(ctx context.Context, key string, value int64) error {
	return errors.New("not used")
}
func (c *createCacheRepoMock) GetCode(ctx context.Context, key string) (int64, error) {
	return 0, errors.New("not used")
}
func (c *createCacheRepoMock) DeleteCode(ctx context.Context, key string) error {
	return errors.New("not used")
}

type createProducerMock struct {
	registerCh chan *eventv1.RegisterEvent
}

func newCreateProducerMock() *createProducerMock {
	return &createProducerMock{registerCh: make(chan *eventv1.RegisterEvent, 16)}
}
func (p *createProducerMock) PublishLoginEvent(ctx context.Context, event *eventv1.LoginEvent) error {
	return nil
}
func (p *createProducerMock) PublishRegisterEvent(ctx context.Context, event *eventv1.RegisterEvent) error {
	p.registerCh <- event
	return nil
}

type createOAuthProviderMock struct{}

func (o *createOAuthProviderMock) AuthCodeURL(ctx context.Context, state string) string {
	return ""
}
func (o *createOAuthProviderMock) FetchUser(ctx context.Context, code string) (*google.UserInfo, error) {
	return nil, errors.New("not used")
}

type createSMSProviderMock struct{}

func (s *createSMSProviderMock) SendSMS(ctx context.Context, msg domain.SMSMessage) error { return nil }

type createEmailProviderMock struct{}

func (e *createEmailProviderMock) SendEmail(ctx context.Context, msg domain.EmailMessage) error {
	return nil
}

func newAuthServiceForCreateTests(userRepo UserRepo, producer Producer) *AuthService {
	log := zap.NewNop()
	return New(
		userRepo,
		&createCacheRepoMock{},
		log,
		[]byte("secret"),
		10*time.Minute,
		producer,
		&createOAuthProviderMock{},
		&createSMSProviderMock{},
		&createEmailProviderMock{},
		1000,
		9999,
	)
}

func TestAuthService_CreateUser_Table(t *testing.T) {
	tests := []struct {
		name string
		in   domain.CreateUser

		repoSetup func(repo *createUserRepoMock)
		wantErr   error
		wantID    int64
		check     func(t *testing.T, repo *createUserRepoMock)
		wantEv    bool
	}{
		{
			name:    "invalid_email_required",
			in:      domain.CreateUser{Email: "   ", Login: "l", Password: "p"},
			wantErr: ErrInvalidCredentials,
			wantEv:  false,
		},
		{
			name:    "invalid_password_required",
			in:      domain.CreateUser{Email: "a@b.com", Login: "l", Password: "   "},
			wantErr: ErrInvalidCredentials,
			wantEv:  false,
		},
		{
			name:    "invalid_login_required",
			in:      domain.CreateUser{Email: "a@b.com", Login: "   ", Password: "p"},
			wantErr: ErrInvalidCredentials,
			wantEv:  false,
		},
		{
			name: "duplicate_email",
			in:   domain.CreateUser{Email: "A@B.COM", Login: "l", Password: "p"},
			repoSetup: func(repo *createUserRepoMock) {
				repo.getByEmailFn = func(ctx context.Context, email string) (*storage.UserLoginModel, error) {
					return &storage.UserLoginModel{UserID: 1}, nil
				}
			},
			wantErr: ErrUserAlreadyExists,
			wantEv:  false,
			check: func(t *testing.T, repo *createUserRepoMock) {
				require.Equal(t, "a@b.com", repo.lastEmail)
			},
		},
		{
			name: "get_by_email_repo_error",
			in:   domain.CreateUser{Email: "a@b.com", Login: "l", Password: "p"},
			repoSetup: func(repo *createUserRepoMock) {
				repo.getByEmailFn = func(ctx context.Context, email string) (*storage.UserLoginModel, error) {
					return nil, errDBDown2
				}
			},
			wantErr: errDBDown2,
			wantEv:  false,
		},
		{
			name: "duplicate_login",
			in:   domain.CreateUser{Email: "a@b.com", Login: "  LOGIN ", Password: "p"},
			repoSetup: func(repo *createUserRepoMock) {
				repo.getByEmailFn = func(ctx context.Context, email string) (*storage.UserLoginModel, error) {
					return nil, storage.ErrUserNotFound
				}
				repo.getByLoginFn = func(ctx context.Context, login string) (*storage.UserLoginModel, error) {
					return &storage.UserLoginModel{UserID: 2}, nil
				}
			},
			wantErr: ErrUserAlreadyExists,
			wantEv:  false,
			check: func(t *testing.T, repo *createUserRepoMock) {
				require.Equal(t, "LOGIN", repo.lastLogin)
			},
		},
		{
			name: "get_by_login_repo_error",
			in:   domain.CreateUser{Email: "a@b.com", Login: "login", Password: "p"},
			repoSetup: func(repo *createUserRepoMock) {
				repo.getByEmailFn = func(ctx context.Context, email string) (*storage.UserLoginModel, error) {
					return nil, storage.ErrUserNotFound
				}
				repo.getByLoginFn = func(ctx context.Context, login string) (*storage.UserLoginModel, error) {
					return nil, errDBDown2
				}
			},
			wantErr: errDBDown2,
			wantEv:  false,
		},
		{
			name: "save_error",
			in:   domain.CreateUser{Email: "a@b.com", Login: "login", Password: "p"},
			repoSetup: func(repo *createUserRepoMock) {
				repo.getByEmailFn = func(ctx context.Context, email string) (*storage.UserLoginModel, error) {
					return nil, storage.ErrUserNotFound
				}
				repo.getByLoginFn = func(ctx context.Context, login string) (*storage.UserLoginModel, error) {
					return nil, storage.ErrUserNotFound
				}
				repo.saveFn = func(ctx context.Context, login *storage.UserLoginModel, info *storage.UserInfoModel) error {
					return errInsertFail2
				}
			},
			wantErr: errInsertFail2,
			wantEv:  false,
		},
		{
			name: "success",
			in:   domain.CreateUser{Email: "A@B.COM", Login: "  login  ", Password: "p", Name: " John ", LastName: " Doe "},
			repoSetup: func(repo *createUserRepoMock) {
				repo.getByEmailFn = func(ctx context.Context, email string) (*storage.UserLoginModel, error) {
					return nil, storage.ErrUserNotFound
				}
				repo.getByLoginFn = func(ctx context.Context, login string) (*storage.UserLoginModel, error) {
					return nil, storage.ErrUserNotFound
				}
				repo.saveFn = func(ctx context.Context, login *storage.UserLoginModel, info *storage.UserInfoModel) error {
					login.UserID = 1234
					return nil
				}
			},
			wantErr: nil,
			wantID:  1234,
			wantEv:  true,
			check: func(t *testing.T, repo *createUserRepoMock) {
				require.Equal(t, "a@b.com", repo.lastEmail)
				require.Equal(t, "login", repo.lastLogin)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := &createUserRepoMock{}
			prod := newCreateProducerMock()

			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}

			svc := newAuthServiceForCreateTests(repo, prod)

			id, err := svc.CreateUser(context.Background(), tt.in)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)

				select {
				case <-prod.registerCh:
					t.Fatalf("unexpected register event publish")
				default:
				}

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantID, id)

			if tt.check != nil {
				tt.check(t, repo)
			}

			if tt.wantEv {
				select {
				case ev := <-prod.registerCh:
					require.Equal(t, id, ev.UserId)
				case <-time.After(createTestTimeout):
					t.Fatalf("expected register event publish")
				}
			}
		})
	}
}

func TestConvToStorage_FromCreateUser_Normalization(t *testing.T) {
	n := NewConvToNormalizer()
	c := NewConvToStorage(n)

	in := domain.CreateUser{
		Email:    " USER@MAIL.COM ",
		Login:    "  login  ",
		Name:     " John ",
		LastName: " Doe ",
	}
	loginModel, infoModel := c.FromCreateUser(in, []byte{1, 2, 3})

	require.Equal(t, "user@mail.com", loginModel.Email)
	require.Equal(t, "login", loginModel.Login)
	require.Equal(t, defaultUserRole, loginModel.Role)

	require.Equal(t, "John", infoModel.Name)
	require.Equal(t, "Doe", infoModel.LastName)
}
