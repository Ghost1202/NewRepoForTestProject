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

const resetTestTimeout = 1200 * time.Millisecond

var (
	errRedisDown3  = errors.New("redis down")
	errDBDown3     = errors.New("db down")
	errUpdateFail3 = errors.New("update failed")
)

type resetUserRepoMock struct {
	getByEmailFn     func(ctx context.Context, email string) (*storage.UserLoginModel, error)
	getByPhoneFn     func(ctx context.Context, phone string) (*storage.UserLoginModel, error)
	updatePassHashFn func(ctx context.Context, userID int64, passHash []byte) error
}

func (m *resetUserRepoMock) Save(ctx context.Context, login *storage.UserLoginModel, info *storage.UserInfoModel) error {
	return errors.New("not used")
}
func (m *resetUserRepoMock) GetByEmail(ctx context.Context, email string) (*storage.UserLoginModel, error) {
	if m.getByEmailFn == nil {
		return nil, errors.New("getByEmailFn not set")
	}
	return m.getByEmailFn(ctx, email)
}
func (m *resetUserRepoMock) GetByID(ctx context.Context, id int64) (*storage.UserLoginModel, error) {
	return nil, errors.New("not used")
}
func (m *resetUserRepoMock) GetByLogin(ctx context.Context, login string) (*storage.UserLoginModel, error) {
	return nil, errors.New("not used")
}
func (m *resetUserRepoMock) SaveUserGoogle(ctx context.Context, login *storage.UserLoginModel, info *storage.UserInfoModel) (int64, error) {
	return 0, errors.New("not used")
}
func (m *resetUserRepoMock) SetGoogleID(ctx context.Context, userID int64, googleID string) error {
	return errors.New("not used")
}
func (m *resetUserRepoMock) GetByPhone(ctx context.Context, phone string) (*storage.UserLoginModel, error) {
	if m.getByPhoneFn == nil {
		return nil, errors.New("getByPhoneFn not set")
	}
	return m.getByPhoneFn(ctx, phone)
}
func (m *resetUserRepoMock) CreateUserByPhone(ctx context.Context, in storage.CreateUserByPhone) (int64, error) {
	return 0, errors.New("not used")
}
func (m *resetUserRepoMock) GetByGoogleID(ctx context.Context, googleID string) (*storage.UserLoginModel, error) {
	return nil, errors.New("not used")
}
func (m *resetUserRepoMock) UpdatePassHash(ctx context.Context, userID int64, passHash []byte) error {
	if m.updatePassHashFn == nil {
		return errors.New("updatePassHashFn not set")
	}
	return m.updatePassHashFn(ctx, userID, passHash)
}

type resetCacheRepoMock struct {
	mu sync.Mutex

	data      map[string]int64
	getErr    error
	setErr    error
	deleteErr error

	deleteCalled bool
}

func newResetCacheRepoMock() *resetCacheRepoMock {
	return &resetCacheRepoMock{data: make(map[string]int64)}
}

func (m *resetCacheRepoMock) SetCode(ctx context.Context, key string, value int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.setErr != nil {
		return m.setErr
	}
	m.data[key] = value
	return nil
}
func (m *resetCacheRepoMock) GetCode(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getErr != nil {
		return 0, m.getErr
	}
	v, ok := m.data[key]
	if !ok {
		return 0, storage.ErrKeyNotFound
	}
	return v, nil
}
func (m *resetCacheRepoMock) DeleteCode(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteCalled = true
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.data, key)
	return nil
}

type resetProducerMock struct {
	loginCh chan *eventv1.LoginEvent
}

func newResetProducerMock() *resetProducerMock {
	return &resetProducerMock{loginCh: make(chan *eventv1.LoginEvent, 16)}
}
func (p *resetProducerMock) PublishLoginEvent(ctx context.Context, event *eventv1.LoginEvent) error {
	p.loginCh <- event
	return nil
}
func (p *resetProducerMock) PublishRegisterEvent(ctx context.Context, event *eventv1.RegisterEvent) error {
	return nil
}

type resetOAuthProviderMock struct{}

func (o *resetOAuthProviderMock) AuthCodeURL(ctx context.Context, state string) string { return "" }
func (o *resetOAuthProviderMock) FetchUser(ctx context.Context, code string) (*google.UserInfo, error) {
	return nil, errors.New("not used")
}

type resetSMSProviderMock struct {
	err error
}

func (s *resetSMSProviderMock) SendSMS(ctx context.Context, msg domain.SMSMessage) error {
	return s.err
}

type resetEmailProviderMock struct {
	err error
}

func (e *resetEmailProviderMock) SendEmail(ctx context.Context, msg domain.EmailMessage) error {
	return e.err
}

func newAuthServiceForResetTests(
	userRepo UserRepo,
	cacheRepo CacheRepo,
	prod Producer,
	smsErr error,
	emailErr error,
	otpMin, otpMax int64,
) *AuthService {
	log := zap.NewNop()
	return New(
		userRepo,
		cacheRepo,
		log,
		[]byte("secret"),
		10*time.Minute,
		prod,
		&resetOAuthProviderMock{},
		&resetSMSProviderMock{err: smsErr},
		&resetEmailProviderMock{err: emailErr},
		otpMin,
		otpMax,
	)
}

func TestAuthService_ResetEmailConfirm_Table(t *testing.T) {
	tests := []struct {
		name string

		email       string
		code        int64
		newPassword string

		cacheSeed func(cache *resetCacheRepoMock, svc *AuthService, email string, code int64)
		repoSetup func(repo *resetUserRepoMock, email string)

		wantErr error
		wantEv  bool
	}{
		{
			name:        "invalid_input",
			email:       "   ",
			newPassword: "   ",
			wantErr:     ErrInvalidCredentials,
		},
		{
			name:        "cache_get_error",
			email:       "a@b.com",
			code:        1111,
			newPassword: "newpass",
			cacheSeed: func(cache *resetCacheRepoMock, svc *AuthService, email string, code int64) {
				cache.getErr = errRedisDown3
			},
			repoSetup: func(repo *resetUserRepoMock, email string) {},
			wantErr:   errRedisDown3,
		},
		{
			name:        "repo_error",
			email:       "a@b.com",
			code:        1111,
			newPassword: "newpass",
			cacheSeed: func(cache *resetCacheRepoMock, svc *AuthService, email string, code int64) {
				key := svc.convToOTP.KeyResetEmail(email)
				cache.data[key] = code
			},
			repoSetup: func(repo *resetUserRepoMock, email string) {
				repo.getByEmailFn = func(ctx context.Context, e string) (*storage.UserLoginModel, error) {
					return nil, errDBDown3
				}
			},
			wantErr: errDBDown3,
		},
		{
			name:        "update_passhash_error",
			email:       "a@b.com",
			code:        1111,
			newPassword: "newpass",
			cacheSeed: func(cache *resetCacheRepoMock, svc *AuthService, email string, code int64) {
				key := svc.convToOTP.KeyResetEmail(email)
				cache.data[key] = code
			},
			repoSetup: func(repo *resetUserRepoMock, email string) {
				repo.getByEmailFn = func(ctx context.Context, e string) (*storage.UserLoginModel, error) {
					return &storage.UserLoginModel{UserID: 10, Email: e, Role: defaultUserRole, Login: "l"}, nil
				}
				repo.updatePassHashFn = func(ctx context.Context, userID int64, passHash []byte) error {
					return errUpdateFail3
				}
			},
			wantErr: errUpdateFail3,
		},
		{
			name:        "success_publishes_login_event",
			email:       "A@B.COM",
			code:        1111,
			newPassword: "newpass",
			cacheSeed: func(cache *resetCacheRepoMock, svc *AuthService, email string, code int64) {
				key := svc.convToOTP.KeyResetEmail(email)
				cache.data[key] = code
			},
			repoSetup: func(repo *resetUserRepoMock, email string) {
				repo.getByEmailFn = func(ctx context.Context, e string) (*storage.UserLoginModel, error) {
					return &storage.UserLoginModel{UserID: 11, Email: e, Role: defaultUserRole, Login: "john"}, nil
				}
				repo.updatePassHashFn = func(ctx context.Context, userID int64, passHash []byte) error {
					return nil
				}
			},
			wantEv: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := &resetUserRepoMock{}
			cache := newResetCacheRepoMock()
			prod := newResetProducerMock()

			svc := newAuthServiceForResetTests(repo, cache, prod, nil, nil, 1000, 9999)

			emailNorm := svc.convToNorm.Email(tt.email)

			if tt.cacheSeed != nil {
				tt.cacheSeed(cache, svc, emailNorm, tt.code)
			}
			if tt.repoSetup != nil {
				tt.repoSetup(repo, emailNorm)
			} else {
				repo.updatePassHashFn = func(ctx context.Context, userID int64, passHash []byte) error { return nil }
				repo.getByEmailFn = func(ctx context.Context, e string) (*storage.UserLoginModel, error) {
					return &storage.UserLoginModel{UserID: 1, Email: e, Role: defaultUserRole, Login: "l"}, nil
				}
			}

			_, err := svc.ResetEmailConfirm(context.Background(), tt.email, tt.code, tt.newPassword)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			if tt.wantEv {
				select {
				case <-prod.loginCh:
				case <-time.After(resetTestTimeout):
					t.Fatalf("expected login event publish")
				}
			}
		})
	}
}

func TestAuthService_ResetPhoneConfirm_Table(t *testing.T) {
	tests := []struct {
		name string

		phone       string
		code        int64
		newPassword string

		cacheSeed func(cache *resetCacheRepoMock, svc *AuthService, phone string, code int64)
		repoSetup func(repo *resetUserRepoMock, phone string)

		wantErr error
		wantEv  bool
	}{
		{
			name:        "invalid_input",
			phone:       "   ",
			newPassword: "   ",
			wantErr:     ErrInvalidCredentials,
		},
		{
			name:        "cache_get_error",
			phone:       "+380501112233",
			code:        1111,
			newPassword: "newpass",
			cacheSeed: func(cache *resetCacheRepoMock, svc *AuthService, phone string, code int64) {
				cache.getErr = errRedisDown3
			},
			repoSetup: func(repo *resetUserRepoMock, phone string) {},
			wantErr:   errRedisDown3,
		},
		{
			name:        "repo_error",
			phone:       "+380501112233",
			code:        1111,
			newPassword: "newpass",
			cacheSeed: func(cache *resetCacheRepoMock, svc *AuthService, phone string, code int64) {
				key := svc.convToOTP.KeyResetPhone(phone)
				cache.data[key] = code
			},
			repoSetup: func(repo *resetUserRepoMock, phone string) {
				repo.getByPhoneFn = func(ctx context.Context, p string) (*storage.UserLoginModel, error) {
					return nil, errDBDown3
				}
			},
			wantErr: errDBDown3,
		},
		{
			name:        "update_passhash_error",
			phone:       "+380501112233",
			code:        1111,
			newPassword: "newpass",
			cacheSeed: func(cache *resetCacheRepoMock, svc *AuthService, phone string, code int64) {
				key := svc.convToOTP.KeyResetPhone(phone)
				cache.data[key] = code
			},
			repoSetup: func(repo *resetUserRepoMock, phone string) {
				repo.getByPhoneFn = func(ctx context.Context, p string) (*storage.UserLoginModel, error) {
					return &storage.UserLoginModel{UserID: 20, Phone: p, Role: defaultUserRole, Login: "l"}, nil
				}
				repo.updatePassHashFn = func(ctx context.Context, userID int64, passHash []byte) error {
					return errUpdateFail3
				}
			},
			wantErr: errUpdateFail3,
		},
		{
			name:        "success_publishes_login_event",
			phone:       "  +380501112233 ",
			code:        1111,
			newPassword: "newpass",
			cacheSeed: func(cache *resetCacheRepoMock, svc *AuthService, phone string, code int64) {
				key := svc.convToOTP.KeyResetPhone(phone)
				cache.data[key] = code
			},
			repoSetup: func(repo *resetUserRepoMock, phone string) {
				repo.getByPhoneFn = func(ctx context.Context, p string) (*storage.UserLoginModel, error) {
					return &storage.UserLoginModel{UserID: 21, Phone: p, Role: defaultUserRole, Login: "john"}, nil
				}
				repo.updatePassHashFn = func(ctx context.Context, userID int64, passHash []byte) error {
					return nil
				}
			},
			wantEv: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := &resetUserRepoMock{}
			cache := newResetCacheRepoMock()
			prod := newResetProducerMock()

			svc := newAuthServiceForResetTests(repo, cache, prod, nil, nil, 1000, 9999)

			phoneNorm := svc.convToNorm.Phone(tt.phone)

			if tt.cacheSeed != nil {
				tt.cacheSeed(cache, svc, phoneNorm, tt.code)
			}
			if tt.repoSetup != nil {
				tt.repoSetup(repo, phoneNorm)
			} else {
				repo.updatePassHashFn = func(ctx context.Context, userID int64, passHash []byte) error { return nil }
				repo.getByPhoneFn = func(ctx context.Context, p string) (*storage.UserLoginModel, error) {
					return &storage.UserLoginModel{UserID: 1, Phone: p, Role: defaultUserRole, Login: "l"}, nil
				}
			}

			_, err := svc.ResetPhoneConfirm(context.Background(), tt.phone, tt.code, tt.newPassword)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			if tt.wantEv {
				select {
				case <-prod.loginCh:
				case <-time.After(resetTestTimeout):
					t.Fatalf("expected login event publish")
				}
			}
		})
	}
}
