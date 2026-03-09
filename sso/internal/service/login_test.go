package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	eventv1 "github.com/turtlepavlo/proto-contract/gen/go/analytics/event/v1"
	"github.com/turtlepavlo/sso/internal/domain"
	"github.com/turtlepavlo/sso/internal/storage"
	"github.com/turtlepavlo/sso/pkg/provider/google"
)

const loginTestTimeout = 1200 * time.Millisecond

var (
	errDBDown     = errors.New("db down")
	errRedisDown  = errors.New("redis down")
	errInsertFail = errors.New("insert failed")
)

type loginUserRepoMock struct {
	mu sync.Mutex

	getByEmailFn func(ctx context.Context, email string) (*storage.UserLoginModel, error)
	getByLoginFn func(ctx context.Context, login string) (*storage.UserLoginModel, error)
	getByPhoneFn func(ctx context.Context, phone string) (*storage.UserLoginModel, error)

	createUserByPhoneFn func(ctx context.Context, in storage.CreateUserByPhone) (int64, error)
	setGoogleIDFn       func(ctx context.Context, userID int64, googleID string) error
	saveUserGoogleFn    func(ctx context.Context, login *storage.UserLoginModel, info *storage.UserInfoModel) (int64, error)

	lastEmail string
	lastLogin string
	lastPhone string

	setGoogleCalled bool
}

func (m *loginUserRepoMock) Save(ctx context.Context, login *storage.UserLoginModel, info *storage.UserInfoModel) error {
	return errors.New("not used in login tests")
}
func (m *loginUserRepoMock) GetByEmail(ctx context.Context, email string) (*storage.UserLoginModel, error) {
	m.mu.Lock()
	m.lastEmail = email
	fn := m.getByEmailFn
	m.mu.Unlock()
	if fn == nil {
		return nil, errors.New("getByEmailFn not set")
	}
	return fn(ctx, email)
}
func (m *loginUserRepoMock) GetByID(ctx context.Context, id int64) (*storage.UserLoginModel, error) {
	return nil, errors.New("not used in login tests")
}
func (m *loginUserRepoMock) GetByLogin(ctx context.Context, login string) (*storage.UserLoginModel, error) {
	m.mu.Lock()
	m.lastLogin = login
	fn := m.getByLoginFn
	m.mu.Unlock()
	if fn == nil {
		return nil, errors.New("getByLoginFn not set")
	}
	return fn(ctx, login)
}
func (m *loginUserRepoMock) SaveUserGoogle(ctx context.Context, login *storage.UserLoginModel, info *storage.UserInfoModel) (int64, error) {
	fn := m.saveUserGoogleFn
	if fn == nil {
		return 0, errors.New("saveUserGoogleFn not set")
	}
	return fn(ctx, login, info)
}
func (m *loginUserRepoMock) SetGoogleID(ctx context.Context, userID int64, googleID string) error {
	m.mu.Lock()
	m.setGoogleCalled = true
	fn := m.setGoogleIDFn
	m.mu.Unlock()
	if fn == nil {
		return nil
	}
	return fn(ctx, userID, googleID)
}
func (m *loginUserRepoMock) GetByPhone(ctx context.Context, phone string) (*storage.UserLoginModel, error) {
	m.mu.Lock()
	m.lastPhone = phone
	fn := m.getByPhoneFn
	m.mu.Unlock()
	if fn == nil {
		return nil, errors.New("getByPhoneFn not set")
	}
	return fn(ctx, phone)
}
func (m *loginUserRepoMock) CreateUserByPhone(ctx context.Context, in storage.CreateUserByPhone) (int64, error) {
	fn := m.createUserByPhoneFn
	if fn == nil {
		return 0, errors.New("createUserByPhoneFn not set")
	}
	return fn(ctx, in)
}
func (m *loginUserRepoMock) GetByGoogleID(ctx context.Context, googleID string) (*storage.UserLoginModel, error) {
	return nil, errors.New("not used")
}
func (m *loginUserRepoMock) UpdatePassHash(ctx context.Context, userID int64, passHash []byte) error {
	return errors.New("not used in login tests")
}

type loginCacheRepoMock struct {
	mu sync.Mutex

	data      map[string]int64
	getErr    error
	setErr    error
	deleteErr error

	lastSetKey   string
	lastSetValue int64

	deleteCalled bool
}

func newLoginCacheRepoMock() *loginCacheRepoMock {
	return &loginCacheRepoMock{data: make(map[string]int64)}
}

func (m *loginCacheRepoMock) SetCode(ctx context.Context, key string, value int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastSetKey = key
	m.lastSetValue = value
	if m.setErr != nil {
		return m.setErr
	}
	m.data[key] = value
	return nil
}
func (m *loginCacheRepoMock) GetCode(ctx context.Context, key string) (int64, error) {
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
func (m *loginCacheRepoMock) DeleteCode(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteCalled = true
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.data, key)
	return nil
}

type loginProducerMock struct {
	loginCh    chan *eventv1.LoginEvent
	registerCh chan *eventv1.RegisterEvent

	loginErr    error
	registerErr error
}

func newLoginProducerMock() *loginProducerMock {
	return &loginProducerMock{
		loginCh:    make(chan *eventv1.LoginEvent, 16),
		registerCh: make(chan *eventv1.RegisterEvent, 16),
	}
}

func (p *loginProducerMock) PublishLoginEvent(ctx context.Context, event *eventv1.LoginEvent) error {
	if p.loginErr != nil {
		return p.loginErr
	}
	p.loginCh <- event
	return nil
}
func (p *loginProducerMock) PublishRegisterEvent(ctx context.Context, event *eventv1.RegisterEvent) error {
	if p.registerErr != nil {
		return p.registerErr
	}
	p.registerCh <- event
	return nil
}

type loginOAuthProviderMock struct {
	authURLFn   func(ctx context.Context, state string) string
	fetchUserFn func(ctx context.Context, code string) (*google.UserInfo, error)
}

func (m *loginOAuthProviderMock) AuthCodeURL(ctx context.Context, state string) string {
	if m.authURLFn == nil {
		return ""
	}
	return m.authURLFn(ctx, state)
}
func (m *loginOAuthProviderMock) FetchUser(ctx context.Context, code string) (*google.UserInfo, error) {
	if m.fetchUserFn == nil {
		return nil, errors.New("fetchUserFn not set")
	}
	return m.fetchUserFn(ctx, code)
}

type loginSMSProviderMock struct {
	last domain.SMSMessage
	err  error
}

func (s *loginSMSProviderMock) SendSMS(ctx context.Context, msg domain.SMSMessage) error {
	s.last = msg
	return s.err
}

type loginEmailProviderMock struct {
	last domain.EmailMessage
	err  error
}

func (e *loginEmailProviderMock) SendEmail(ctx context.Context, msg domain.EmailMessage) error {
	e.last = msg
	return e.err
}

func newAuthServiceForLoginTests(
	userRepo UserRepo,
	cacheRepo CacheRepo,
	producer Producer,
	oauth OAuthProvider,
	sms SMSProvider,
	email EmailProvider,
	otpMin, otpMax int64,
) *AuthService {
	log := zap.NewNop()
	return New(
		userRepo,
		cacheRepo,
		log,
		[]byte("secret"),
		15*time.Minute,
		producer,
		oauth,
		sms,
		email,
		otpMin,
		otpMax,
	)
}

func TestAuthService_Login_Table(t *testing.T) {
	pass := "p@ss"
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.MinCost)
	require.NoError(t, err)

	tests := []struct {
		name       string
		identifier string
		password   string
		isEmail    bool

		repoSetup func(m *loginUserRepoMock)
		wantErr   error
		check     func(t *testing.T, repo *loginUserRepoMock, sess domain.AuthSession)
	}{
		{
			name:       "invalid_empty_identifier_email",
			identifier: "   ",
			password:   "x",
			isEmail:    true,
			repoSetup: func(m *loginUserRepoMock) {
				m.getByEmailFn = func(ctx context.Context, email string) (*storage.UserLoginModel, error) {
					return nil, storage.ErrUserNotFound
				}
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:       "invalid_empty_password",
			identifier: "user@mail.com",
			password:   "   ",
			isEmail:    true,
			repoSetup: func(m *loginUserRepoMock) {
				m.getByEmailFn = func(ctx context.Context, email string) (*storage.UserLoginModel, error) {
					return &storage.UserLoginModel{
						UserID:   10,
						Email:    email,
						Login:    "l",
						PassHash: hash,
						Role:     defaultUserRole,
					}, nil
				}
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:       "email_not_found_becomes_invalid_credentials",
			identifier: "USER@MAIL.COM",
			password:   pass,
			isEmail:    true,
			repoSetup: func(m *loginUserRepoMock) {
				m.getByEmailFn = func(ctx context.Context, email string) (*storage.UserLoginModel, error) {
					return nil, storage.ErrUserNotFound
				}
			},
			wantErr: ErrInvalidCredentials,
			check: func(t *testing.T, repo *loginUserRepoMock, _ domain.AuthSession) {
				require.Equal(t, "user@mail.com", repo.lastEmail)
			},
		},
		{
			name:       "login_not_found_becomes_invalid_credentials",
			identifier: "  myLogin ",
			password:   pass,
			isEmail:    false,
			repoSetup: func(m *loginUserRepoMock) {
				m.getByLoginFn = func(ctx context.Context, login string) (*storage.UserLoginModel, error) {
					return nil, storage.ErrUserNotFound
				}
			},
			wantErr: ErrInvalidCredentials,
			check: func(t *testing.T, repo *loginUserRepoMock, _ domain.AuthSession) {
				require.Equal(t, "myLogin", repo.lastLogin)
			},
		},
		{
			name:       "repo_error_is_returned",
			identifier: "user@mail.com",
			password:   pass,
			isEmail:    true,
			repoSetup: func(m *loginUserRepoMock) {
				m.getByEmailFn = func(ctx context.Context, email string) (*storage.UserLoginModel, error) {
					return nil, errDBDown
				}
			},
			wantErr: errDBDown,
		},
		{
			name:       "wrong_password",
			identifier: "user@mail.com",
			password:   "wrong",
			isEmail:    true,
			repoSetup: func(m *loginUserRepoMock) {
				m.getByEmailFn = func(ctx context.Context, email string) (*storage.UserLoginModel, error) {
					return &storage.UserLoginModel{
						UserID:   10,
						Email:    email,
						Login:    "l",
						PassHash: hash,
						Role:     defaultUserRole,
					}, nil
				}
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:       "success_by_email",
			identifier: " USER@MAIL.COM ",
			password:   pass,
			isEmail:    true,
			repoSetup: func(m *loginUserRepoMock) {
				m.getByEmailFn = func(ctx context.Context, email string) (*storage.UserLoginModel, error) {
					return &storage.UserLoginModel{
						UserID:   11,
						Email:    email,
						Login:    "john",
						PassHash: hash,
						Role:     defaultUserRole,
					}, nil
				}
			},
			wantErr: nil,
			check: func(t *testing.T, repo *loginUserRepoMock, sess domain.AuthSession) {
				require.Equal(t, int64(11), sess.UserID)
				require.Equal(t, defaultUserRole, sess.Role)
				require.NotEmpty(t, sess.AccessToken)
				require.Equal(t, "john", sess.Login)
				require.Equal(t, "user@mail.com", repo.lastEmail)
			},
		},
		{
			name:       "success_by_login",
			identifier: "  john  ",
			password:   pass,
			isEmail:    false,
			repoSetup: func(m *loginUserRepoMock) {
				m.getByLoginFn = func(ctx context.Context, login string) (*storage.UserLoginModel, error) {
					return &storage.UserLoginModel{
						UserID:   12,
						Email:    "x@x.com",
						Login:    login,
						PassHash: hash,
						Role:     defaultUserRole,
					}, nil
				}
			},
			wantErr: nil,
			check: func(t *testing.T, repo *loginUserRepoMock, sess domain.AuthSession) {
				require.Equal(t, int64(12), sess.UserID)
				require.Equal(t, "john", repo.lastLogin)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &loginUserRepoMock{}
			cache := newLoginCacheRepoMock()
			prod := newLoginProducerMock()
			oauth := &loginOAuthProviderMock{}
			sms := &loginSMSProviderMock{}
			email := &loginEmailProviderMock{}

			if tt.repoSetup != nil {
				tt.repoSetup(userRepo)
			}
			svc := newAuthServiceForLoginTests(userRepo, cache, prod, oauth, sms, email, 1000, 9999)

			sess, err := svc.Login(context.Background(), tt.identifier, tt.password, tt.isEmail)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, userRepo, sess)
			}

			select {
			case ev := <-prod.loginCh:
				require.Equal(t, sess.UserID, ev.UserId)
			case <-time.After(loginTestTimeout):
				t.Fatalf("expected login event publish")
			}
		})
	}
}

func TestAuthService_LoginWithPhone_Table(t *testing.T) {
	tests := []struct {
		name      string
		phone     string
		code      int64
		cacheSeed func(cache *loginCacheRepoMock, svc *AuthService, phone string, code int64)

		userRepoSetup func(repo *loginUserRepoMock, phone string)
		wantErr       error
		wantNewUser   bool
	}{
		{
			name:    "invalid_phone",
			phone:   "   ",
			code:    1111,
			wantErr: ErrInvalidCredentials,
		},
		{
			name:  "otp_not_found",
			phone: "+380501112233",
			code:  1111,
			cacheSeed: func(cache *loginCacheRepoMock, svc *AuthService, phone string, code int64) {
			},
			userRepoSetup: func(repo *loginUserRepoMock, phone string) {},
			wantErr:       ErrCodeNotFound,
		},
		{
			name:  "cache_get_error",
			phone: "+380501112233",
			code:  1111,
			cacheSeed: func(cache *loginCacheRepoMock, svc *AuthService, phone string, code int64) {
				cache.getErr = errRedisDown
			},
			userRepoSetup: func(repo *loginUserRepoMock, phone string) {},
			wantErr:       errRedisDown,
		},
		{
			name:  "otp_mismatch",
			phone: "+380501112233",
			code:  2222,
			cacheSeed: func(cache *loginCacheRepoMock, svc *AuthService, phone string, code int64) {
				key := svc.convToOTP.KeyLoginPhone(phone)
				cache.data[key] = 1111
			},
			userRepoSetup: func(repo *loginUserRepoMock, phone string) {},
			wantErr:       ErrInvalidCredentials,
		},
		{
			name:  "user_not_found_create_new_user_register_event",
			phone: "+380501112233",
			code:  1111,
			cacheSeed: func(cache *loginCacheRepoMock, svc *AuthService, phone string, code int64) {
				key := svc.convToOTP.KeyLoginPhone(phone)
				cache.data[key] = code
			},
			userRepoSetup: func(repo *loginUserRepoMock, phone string) {
				repo.getByPhoneFn = func(ctx context.Context, ph string) (*storage.UserLoginModel, error) {
					return nil, storage.ErrUserNotFound
				}
				repo.createUserByPhoneFn = func(ctx context.Context, _ storage.CreateUserByPhone) (int64, error) {
					return 77, nil
				}
			},
			wantErr:     nil,
			wantNewUser: true,
		},
		{
			name:  "create_user_error",
			phone: "+380501112233",
			code:  1111,
			cacheSeed: func(cache *loginCacheRepoMock, svc *AuthService, phone string, code int64) {
				key := svc.convToOTP.KeyLoginPhone(phone)
				cache.data[key] = code
			},
			userRepoSetup: func(repo *loginUserRepoMock, phone string) {
				repo.getByPhoneFn = func(ctx context.Context, ph string) (*storage.UserLoginModel, error) {
					return nil, storage.ErrUserNotFound
				}
				repo.createUserByPhoneFn = func(ctx context.Context, _ storage.CreateUserByPhone) (int64, error) {
					return 0, errInsertFail
				}
			},
			wantErr: errInsertFail,
		},
		{
			name:  "existing_user_login_event",
			phone: "+380501112233",
			code:  1111,
			cacheSeed: func(cache *loginCacheRepoMock, svc *AuthService, phone string, code int64) {
				key := svc.convToOTP.KeyLoginPhone(phone)
				cache.data[key] = code
			},
			userRepoSetup: func(repo *loginUserRepoMock, phone string) {
				repo.getByPhoneFn = func(ctx context.Context, ph string) (*storage.UserLoginModel, error) {
					return &storage.UserLoginModel{
						UserID: 88,
						Phone:  ph,
						Role:   defaultUserRole,
					}, nil
				}
			},
			wantErr:     nil,
			wantNewUser: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &loginUserRepoMock{}
			cache := newLoginCacheRepoMock()
			prod := newLoginProducerMock()
			oauth := &loginOAuthProviderMock{}
			sms := &loginSMSProviderMock{}
			email := &loginEmailProviderMock{}

			svc := newAuthServiceForLoginTests(userRepo, cache, prod, oauth, sms, email, 1000, 9999)

			if tt.cacheSeed != nil {
				tt.cacheSeed(cache, svc, svc.convToNorm.Phone(tt.phone), tt.code)
			}
			if tt.userRepoSetup != nil {
				tt.userRepoSetup(userRepo, svc.convToNorm.Phone(tt.phone))
			}

			_, err := svc.LoginWithPhone(context.Background(), tt.phone, tt.code)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			if tt.wantNewUser {
				select {
				case <-prod.registerCh:
				case <-time.After(loginTestTimeout):
					t.Fatalf("expected register event publish")
				}
			} else {
				select {
				case <-prod.loginCh:
				case <-time.After(loginTestTimeout):
					t.Fatalf("expected login event publish")
				}
			}
		})
	}
}

func TestAuthService_ConfirmOTP_Table(t *testing.T) {
	tests := []struct {
		name string
		in   domain.OTPConfirm

		cacheSeed func(cache *loginCacheRepoMock, svc *AuthService, destination string, code int64)

		userRepoSetup func(repo *loginUserRepoMock, destination string)
		wantErr       error
		wantNewUser   bool
	}{
		{
			name:    "destination_empty",
			in:      domain.OTPConfirm{Type: typePhone, Destination: "   ", Code: 1111},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:    "unsupported_type_email",
			in:      domain.OTPConfirm{Type: "email", Destination: "a@b.com", Code: 1111},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "otp_not_found",
			in:   domain.OTPConfirm{Type: typePhone, Destination: "+380501112233", Code: 1111},
			cacheSeed: func(cache *loginCacheRepoMock, svc *AuthService, destination string, code int64) {
			},
			userRepoSetup: func(repo *loginUserRepoMock, destination string) {},
			wantErr:       ErrCodeNotFound,
		},
		{
			name: "cache_get_error",
			in:   domain.OTPConfirm{Type: typePhone, Destination: "+380501112233", Code: 1111},
			cacheSeed: func(cache *loginCacheRepoMock, svc *AuthService, destination string, code int64) {
				cache.getErr = errRedisDown
			},
			userRepoSetup: func(repo *loginUserRepoMock, destination string) {},
			wantErr:       errRedisDown,
		},
		{
			name: "otp_invalid",
			in:   domain.OTPConfirm{Type: typePhone, Destination: "+380501112233", Code: 2222},
			cacheSeed: func(cache *loginCacheRepoMock, svc *AuthService, destination string, code int64) {
				key := svc.convToOTP.KeyLoginPhone(destination)
				cache.data[key] = 1111
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "new_user_register_event",
			in:   domain.OTPConfirm{Type: typePhone, Destination: "+380501112233", Code: 1111},
			cacheSeed: func(cache *loginCacheRepoMock, svc *AuthService, destination string, code int64) {
				key := svc.convToOTP.KeyLoginPhone(destination)
				cache.data[key] = code
			},
			userRepoSetup: func(repo *loginUserRepoMock, destination string) {
				repo.getByPhoneFn = func(ctx context.Context, ph string) (*storage.UserLoginModel, error) {
					return nil, storage.ErrUserNotFound
				}
				repo.createUserByPhoneFn = func(ctx context.Context, _ storage.CreateUserByPhone) (int64, error) {
					return 901, nil
				}
			},
			wantErr:     nil,
			wantNewUser: true,
		},
		{
			name: "existing_user_login_event",
			in:   domain.OTPConfirm{Type: typePhone, Destination: "+380501112233", Code: 1111},
			cacheSeed: func(cache *loginCacheRepoMock, svc *AuthService, destination string, code int64) {
				key := svc.convToOTP.KeyLoginPhone(destination)
				cache.data[key] = code
			},
			userRepoSetup: func(repo *loginUserRepoMock, destination string) {
				repo.getByPhoneFn = func(ctx context.Context, ph string) (*storage.UserLoginModel, error) {
					return &storage.UserLoginModel{UserID: 902, Phone: ph, Role: defaultUserRole}, nil
				}
			},
			wantErr:     nil,
			wantNewUser: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &loginUserRepoMock{}
			cache := newLoginCacheRepoMock()
			prod := newLoginProducerMock()
			oauth := &loginOAuthProviderMock{}
			sms := &loginSMSProviderMock{}
			email := &loginEmailProviderMock{}

			svc := newAuthServiceForLoginTests(userRepo, cache, prod, oauth, sms, email, 1000, 9999)

			destNorm := svc.convToNorm.DestinationByType(tt.in.Destination, tt.in.Type)

			if tt.cacheSeed != nil {
				tt.cacheSeed(cache, svc, destNorm, tt.in.Code)
			}
			if tt.userRepoSetup != nil {
				tt.userRepoSetup(userRepo, destNorm)
			}

			_, err := svc.ConfirmOTP(context.Background(), tt.in)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			if tt.wantNewUser {
				select {
				case <-prod.registerCh:
				case <-time.After(loginTestTimeout):
					t.Fatalf("expected register event publish")
				}
			} else {
				select {
				case <-prod.loginCh:
				case <-time.After(loginTestTimeout):
					t.Fatalf("expected login event publish")
				}
			}
		})
	}
}
