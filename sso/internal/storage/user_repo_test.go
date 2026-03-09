package storage

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func strPtr(s string) *string {
	return &s
}

func (UserLoginModel) TableName() string {
	return "users"
}

func (UserInfoModel) TableName() string {
	return "user_info"
}

func setupInMemoryDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:memdb_%d?mode=memory&cache=shared", time.Now().UnixNano())

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: nil,
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&UserLoginModel{}, &UserInfoModel{})
	require.NoError(t, err)
	return db
}

func TestUserRepo_Integration_Transactions(t *testing.T) {
	db := setupInMemoryDB(t)
	repo := NewUserRepo(db, zaptest.NewLogger(t))
	ctx := context.Background()

	t.Run("Save Success", func(t *testing.T) {
		login := &UserLoginModel{
			Email:    "test@mail.com",
			Login:    "testuser",
			PassHash: []byte("secret"),
			Role:     "user",
		}
		info := &UserInfoModel{
			Name:     "Ivan",
			LastName: "Testov",
		}

		err := repo.Save(ctx, login, info)
		assert.NoError(t, err)
		assert.NotZero(t, login.UserID)

		var savedUser UserLoginModel
		err = db.Table("users").First(&savedUser, "email = ?", login.Email).Error
		assert.NoError(t, err)
		assert.Equal(t, "testuser", savedUser.Login)

		var savedInfo UserInfoModel
		err = db.Table("user_info").First(&savedInfo, "id = ?", login.UserID).Error
		assert.NoError(t, err)
		assert.Equal(t, "Ivan", savedInfo.Name)
	})

	t.Run("SaveUserGoogle Success", func(t *testing.T) {
		login := &UserLoginModel{
			Email:    "google@mail.com",
			Login:    "g_user",
			GoogleID: strPtr("gid-123"),
			Role:     "user",
		}
		info := &UserInfoModel{Name: "Goo", LastName: "Gle"}

		id, err := repo.SaveUserGoogle(ctx, login, info)
		assert.NoError(t, err)
		assert.NotZero(t, id)

		var check UserLoginModel
		db.Table("users").First(&check, id)
		require.NotNil(t, check.GoogleID)
		assert.Equal(t, "gid-123", *check.GoogleID)
	})

	t.Run("CreateUserByPhone Success", func(t *testing.T) {
		in := CreateUserByPhone{
			Phone: "+380000000000",
			Role:  "user",
		}

		id, err := repo.CreateUserByPhone(ctx, in)
		assert.NoError(t, err)

		var check UserLoginModel
		err = db.Table("users").First(&check, id).Error
		assert.NoError(t, err)
		assert.Equal(t, in.Phone, check.Phone)
		assert.Equal(t, in.Role, check.Role)
	})

	t.Run("Transaction Rollback Logic", func(t *testing.T) {

		_ = db.Migrator().DropTable(&UserInfoModel{})
		login := &UserLoginModel{Email: "fail@mail.com", Login: "fail"}
		info := &UserInfoModel{Name: "Fail"}

		err := repo.Save(ctx, login, info)
		assert.Error(t, err)

		var count int64
		db.Table("users").Where("email = ?", "fail@mail.com").Count(&count)
		assert.Equal(t, int64(0), count)

		_ = db.AutoMigrate(&UserInfoModel{})
	})
}

func TestUserRepo_Integration_ConcurrentWrites(t *testing.T) {
	db := setupInMemoryDB(t)
	repo := NewUserRepo(db, zaptest.NewLogger(t))
	ctx := context.Background()

	workers := 10
	var wg sync.WaitGroup
	wg.Add(workers)
	errChan := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func(idx int) {
			defer wg.Done()
			login := &UserLoginModel{
				Email:    "conc" + strconv.Itoa(idx) + "@test.com",
				Login:    "user" + strconv.Itoa(idx),
				PassHash: []byte("pass"),
			}
			info := &UserInfoModel{Name: "U", LastName: "S"}
			if err := repo.Save(ctx, login, info); err != nil {
				errChan <- err
			}
		}(i)
	}
	wg.Wait()
	close(errChan)

	for err := range errChan {
		assert.NoError(t, err)
	}

	var count int64
	db.Table("users").Count(&count)

	assert.Equal(t, int64(workers), count)
}

func TestUserRepo_Integration_Updates(t *testing.T) {
	db := setupInMemoryDB(t)
	repo := NewUserRepo(db, zaptest.NewLogger(t))
	ctx := context.Background()

	user := &UserLoginModel{Login: "updater", Role: "user"}
	db.Table("users").Create(user)

	t.Run("SetGoogleID", func(t *testing.T) {
		newGID := "new-gid-999"
		err := repo.SetGoogleID(ctx, user.UserID, newGID)
		assert.NoError(t, err)

		var check UserLoginModel
		db.Table("users").First(&check, user.UserID)
		require.NotNil(t, check.GoogleID)
		assert.Equal(t, newGID, *check.GoogleID)
	})

	t.Run("UpdatePassHashByID", func(t *testing.T) {
		newHash := []byte("new-hash-secure")
		err := repo.UpdatePassHash(ctx, user.UserID, newHash)
		assert.NoError(t, err)

		var check UserLoginModel
		db.Table("users").First(&check, user.UserID)
		assert.Equal(t, newHash, check.PassHash)
	})

	t.Run("Update NonExistent User", func(t *testing.T) {
		err := repo.SetGoogleID(ctx, 999999, "ghost")
		assert.NoError(t, err)
	})
}

func TestUserRepo_Mocks_GetMethods(t *testing.T) {
	setupMock := func(t *testing.T) (*UserRepo, sqlmock.Sqlmock) {
		sqlDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{Logger: nil})
		require.NoError(t, err)
		return NewUserRepo(gormDB, zaptest.NewLogger(t)), mock
	}

	t.Run("GetByEmail Success", func(t *testing.T) {
		repo, mock := setupMock(t)
		email := "find@me.com"
		rows := sqlmock.NewRows([]string{"id", "email", "user_login"}).
			AddRow(10, email, "found_user")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1`)).
			WithArgs(email, 1).
			WillReturnRows(rows)

		res, err := repo.GetByEmail(context.Background(), email)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, "found_user", res.Login)
	})

	t.Run("GetByEmail Not Found", func(t *testing.T) {
		repo, mock := setupMock(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1`)).
			WithArgs("missing@me.com", 1).
			WillReturnRows(sqlmock.NewRows(nil))

		res, err := repo.GetByEmail(context.Background(), "missing@me.com")
		require.ErrorIs(t, err, ErrUserNotFound)
		assert.Nil(t, res)
	})

	t.Run("GetByGoogleID Success", func(t *testing.T) {
		repo, mock := setupMock(t)
		gid := "gid-555"
		rows := sqlmock.NewRows([]string{"id", "google_id", "user_login"}).
			AddRow(55, gid, "google_user")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE google_id = $1`)).
			WithArgs(gid, 1).
			WillReturnRows(rows)

		res, err := repo.GetByGoogleID(context.Background(), gid)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, gid, *res.GoogleID)
	})

	t.Run("GetByPhone Success", func(t *testing.T) {
		repo, mock := setupMock(t)
		phone := "+123456789"
		rows := sqlmock.NewRows([]string{"id", "phone", "user_login"}).
			AddRow(77, phone, "phone_user")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE phone = $1`)).
			WithArgs(phone, 1).
			WillReturnRows(rows)

		res, err := repo.GetByPhone(context.Background(), phone)
		require.NoError(t, err)
		assert.Equal(t, phone, res.Phone)
	})

	t.Run("GetByLogin Success", func(t *testing.T) {
		repo, mock := setupMock(t)
		login := "superadmin"
		rows := sqlmock.NewRows([]string{"id", "user_login"}).
			AddRow(88, login)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE user_login = $1`)).
			WithArgs(login, 1).
			WillReturnRows(rows)

		res, err := repo.GetByLogin(context.Background(), login)
		require.NoError(t, err)
		assert.Equal(t, login, res.Login)
	})

	t.Run("GetByID DB Error", func(t *testing.T) {
		repo, mock := setupMock(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1`)).
			WithArgs(999, 1).
			WillReturnError(sql.ErrConnDone)

		res, err := repo.GetByID(context.Background(), 999)
		require.ErrorIs(t, err, sql.ErrConnDone)
		assert.Nil(t, res)
	})

	t.Run("Duplicate Constraint Error", func(t *testing.T) {
		repo, mock := setupMock(t)

		mock.ExpectBegin()
		mock.ExpectQuery(".*").WillReturnError(fmt.Errorf("duplicate key value violates unique constraint"))
		mock.ExpectRollback()

		err := repo.Save(context.Background(), &UserLoginModel{Email: "dup@e.com"}, &UserInfoModel{})
		assert.Error(t, err)
	})
}

func TestUserRepo_ContextCancellation(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{Logger: nil})
	repo := NewUserRepo(gormDB, zaptest.NewLogger(t))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mock.ExpectQuery(".*").WillReturnError(context.Canceled)

	_, err := repo.GetByEmail(ctx, "test@mail.com")
	assert.Error(t, err)
}

func TestUserRepo_NilChecks(t *testing.T) {
	repo := NewUserRepo(nil, zap.NewNop())
	ctx := context.Background()

	assert.ErrorIs(t, repo.Save(ctx, &UserLoginModel{}, &UserInfoModel{}), gorm.ErrInvalidDB)
	_, err := repo.GetByID(ctx, 1)
	assert.ErrorIs(t, err, gorm.ErrInvalidDB)
	err = repo.SetGoogleID(ctx, 1, "asd")
	assert.ErrorIs(t, err, gorm.ErrInvalidDB)
}

func BenchmarkUserRepo_Save_Mock(b *testing.B) {
	sqlDB, mock, _ := sqlmock.New()
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{Logger: nil})
	repo := NewUserRepo(gormDB, zaptest.NewLogger(nil))
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i))
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i))
		mock.ExpectCommit()

		_ = repo.Save(ctx, &UserLoginModel{Login: "bench"}, &UserInfoModel{})
	}
}
