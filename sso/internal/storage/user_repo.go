package storage

import (
	"context"
	"errors"

	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	tableUsers    = "users"
	tableUserInfo = "user_info"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepo struct {
	log *zap.Logger
	db  *gorm.DB
}

func NewUserRepo(db *gorm.DB, log *zap.Logger) *UserRepo {

	return &UserRepo{
		log: log,
		db:  db,
	}
}

func (repo *UserRepo) Save(ctx context.Context, loginData *UserLoginModel, infoData *UserInfoModel) error {
	const op = "UserRepo.Save"
	tracer := otel.Tracer("database/postgres")
	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "INSERT"),
			attribute.String("user.email", loginData.Email),
			attribute.String("user.login", loginData.Login),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log).With(zap.String("op", op))
	if repo.db == nil {
		span.SetStatus(codes.Error, "nil db")
		log.Error("nil db")
		return gorm.ErrInvalidDB
	}

	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(tableUsers).
			Select("email", "user_login", "pass_hash", "role").
			Create(loginData).Error; err != nil {
			return err
		}

		infoData.UserID = loginData.UserID
		if err := tx.Table(tableUserInfo).
			Select("first_name", "last_name", "id").
			Create(infoData).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "transaction failed")
		log.Error("save transaction failed",
			zap.Int64("user_id", loginData.UserID),
			zap.String("email", loginData.Email),
			zap.String("login", loginData.Login),
			zap.Error(err),
		)
		return err
	}

	span.SetAttributes(attribute.Int64("user.id", loginData.UserID))
	return nil
}

func (repo *UserRepo) SaveUserGoogle(ctx context.Context, loginData *UserLoginModel, infoData *UserInfoModel) (int64, error) {
	const op = "UserRepo.SaveUserGoogle"
	tracer := otel.Tracer("database/postgres")
	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "INSERT_GOOGLE"),
			attribute.String("user.email", loginData.Email),
			attribute.String("user.login", loginData.Login),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log).With(zap.String("op", op))
	if repo.db == nil {
		span.SetStatus(codes.Error, "nil db")
		log.Error("nil db")
		return 0, gorm.ErrInvalidDB
	}

	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(tableUsers).
			Select("email", "google_id", "role", "user_login").
			Create(loginData).Error; err != nil {
			return err
		}

		infoData.UserID = loginData.UserID

		if err := tx.Table(tableUserInfo).
			Select("first_name", "last_name", "id").
			Create(infoData).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "transaction failed")
		log.Error("save google user transaction failed",
			zap.String("email", loginData.Email),
			zap.String("login", loginData.Login),
			zap.Error(err),
		)
		return 0, err
	}

	span.SetAttributes(attribute.Int64("user.id", loginData.UserID))
	return loginData.UserID, nil
}

func (repo *UserRepo) CreateUserByPhone(ctx context.Context, in CreateUserByPhone) (int64, error) {
	const op = "UserRepo.CreateUserByPhone"
	tracer := otel.Tracer("database/postgres")

	phone := in.Phone
	role := in.Role

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "INSERT_PHONE"),
			attribute.String("user.phone", phone),
			attribute.String("user.role", role),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log).With(zap.String("op", op))

	if repo.db == nil {
		span.SetStatus(codes.Error, "nil db")
		log.Error("nil db")
		return 0, gorm.ErrInvalidDB
	}
	if phone == "" {
		span.SetStatus(codes.Error, "empty phone")
		log.Error("empty phone")
		return 0, errors.New("phone is required")
	}
	if role == "" {
		span.SetStatus(codes.Error, "empty role")
		log.Error("empty role")
		return 0, errors.New("role is required")
	}

	newUser := UserLoginModel{
		Phone: phone,
		Role:  role,
	}

	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(tableUsers).
			Select("phone", "role").
			Create(&newUser).Error; err != nil {
			return err
		}

		userInfo := UserInfoModel{UserID: newUser.UserID}
		if err := tx.Table(tableUserInfo).
			Select("id").
			Create(&userInfo).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "transaction failed")
		log.Error("create user by phone failed",
			zap.String("phone", phone),
			zap.String("role", role),
			zap.Error(err),
		)
		return 0, err
	}

	span.SetAttributes(attribute.Int64("user.id", newUser.UserID))
	return newUser.UserID, nil
}

func (repo *UserRepo) GetByEmail(ctx context.Context, email string) (*UserLoginModel, error) {
	const op = "UserRepo.GetByEmail"
	tracer := otel.Tracer("database/postgres")
	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(attribute.String("user.email", email)),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log).With(zap.String("op", op))
	if repo.db == nil {
		span.SetStatus(codes.Error, "nil db")
		log.Error("nil db")
		return nil, gorm.ErrInvalidDB
	}

	var model UserLoginModel
	err := repo.db.WithContext(ctx).
		Table(tableUsers).
		Where("email = ?", email).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			span.SetStatus(codes.Error, "not found")
			return nil, ErrUserNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "db error")
		log.Error("get by email failed", zap.Error(err))
		return nil, err
	}

	return &model, nil
}

func (repo *UserRepo) GetByPhone(ctx context.Context, phone string) (*UserLoginModel, error) {
	const op = "UserRepo.GetByPhone"
	tracer := otel.Tracer("database/postgres")
	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(attribute.String("user.phone", phone)),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log).With(zap.String("op", op))
	if repo.db == nil {
		span.SetStatus(codes.Error, "nil db")
		log.Error("nil db")
		return nil, gorm.ErrInvalidDB
	}

	var model UserLoginModel
	err := repo.db.WithContext(ctx).
		Table(tableUsers).
		Where("phone = ?", phone).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			span.SetStatus(codes.Error, "not found")
			return nil, ErrUserNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "db error")
		log.Error("get by phone failed", zap.Error(err))
		return nil, err
	}

	return &model, nil
}

func (repo *UserRepo) GetByID(ctx context.Context, id int64) (*UserLoginModel, error) {
	const op = "UserRepo.GetByID"
	tracer := otel.Tracer("database/postgres")
	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(attribute.Int64("user.id", id)),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log).With(zap.String("op", op))
	if repo.db == nil {
		span.SetStatus(codes.Error, "nil db")
		log.Error("nil db")
		return nil, gorm.ErrInvalidDB
	}

	var model UserLoginModel
	err := repo.db.WithContext(ctx).
		Table(tableUsers).
		Where("id = ?", id).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			span.SetStatus(codes.Error, "not found")
			return nil, ErrUserNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "db error")
		log.Error("get by id failed", zap.Error(err))
		return nil, err
	}

	return &model, nil
}

func (repo *UserRepo) GetByLogin(ctx context.Context, loginStr string) (*UserLoginModel, error) {
	const op = "UserRepo.GetByLogin"
	tracer := otel.Tracer("database/postgres")
	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(attribute.String("user.login", loginStr)),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log).With(zap.String("op", op))
	if repo.db == nil {
		span.SetStatus(codes.Error, "nil db")
		log.Error("nil db")
		return nil, gorm.ErrInvalidDB
	}

	var model UserLoginModel
	err := repo.db.WithContext(ctx).
		Table(tableUsers).
		Where("user_login = ?", loginStr).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			span.SetStatus(codes.Error, "not found")
			return nil, ErrUserNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "db error")
		log.Error("get by login failed", zap.Error(err))
		return nil, err
	}

	return &model, nil
}

func (repo *UserRepo) GetByGoogleID(ctx context.Context, googleID string) (*UserLoginModel, error) {
	const op = "UserRepo.GetByGoogleID"
	tracer := otel.Tracer("database/postgres")
	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(attribute.String("user.google_id", googleID)),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log).With(zap.String("op", op))
	if repo.db == nil {
		span.SetStatus(codes.Error, "nil db")
		log.Error("nil db")
		return nil, gorm.ErrInvalidDB
	}

	var model UserLoginModel
	err := repo.db.WithContext(ctx).
		Table(tableUsers).
		Where("google_id = ?", googleID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			span.SetStatus(codes.Error, "not found")
			return nil, ErrUserNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "db error")
		log.Error("get by google id failed", zap.Error(err))
		return nil, err
	}

	return &model, nil
}

func (repo *UserRepo) SetGoogleID(ctx context.Context, userID int64, googleID string) error {
	const op = "UserRepo.SetGoogleID"
	tracer := otel.Tracer("database/postgres")
	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.Int64("user.id", userID),
			attribute.String("user.google_id", googleID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log).With(zap.String("op", op))
	if repo.db == nil {
		span.SetStatus(codes.Error, "nil db")
		log.Error("nil db")
		return gorm.ErrInvalidDB
	}

	err := repo.db.WithContext(ctx).
		Table(tableUsers).
		Where("id = ?", userID).
		Update("google_id", googleID).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db error")
		log.Error("set google id failed", zap.Error(err), zap.Int64("user_id", userID))
		return err
	}

	return nil
}

func (repo *UserRepo) UpdatePassHash(ctx context.Context, userID int64, passHash []byte) error {
	const op = "UserRepo.UpdatePassHash"
	tracer := otel.Tracer("database/postgres")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "UPDATE"),
			attribute.Int64("user.id", userID),
		),
	)
	defer span.End()
	log := telemetry.WithTrace(ctx, repo.log).With(zap.String("op", op))

	if repo.db == nil {
		span.SetStatus(codes.Error, "nil db")
		log.Error("nil db")
		return gorm.ErrInvalidDB
	}

	err := repo.db.WithContext(ctx).
		Table(tableUsers).
		Where("id = ?", userID).
		Update("pass_hash", passHash).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db error")
		log.Error("update pass_hash failed", zap.Error(err), zap.Int64("user_id", userID))
		return err
	}

	return nil
}
