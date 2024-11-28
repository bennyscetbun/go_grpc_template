package api

import (
	"context"
	"time"

	"github.com/bennyscetbun/xxx_your_app_xxx/backend/generated/database/dbmodels"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/generated/database/dbqueries"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/generated/rpc/apiproto"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/apihelpers"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/grpcerrors"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/logger"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/passwd"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/random"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	authorizationHeader = "authorization"
	tokenExpireAfter    = time.Minute * 10
)

func (g *GRPCServer) refreshToken(ctx context.Context, userId string, db *dbqueries.Query) (string, error) {
	token := random.RandString(20)
	ut := db.UserToken
	if err := ut.WithContext(ctx).Create(&dbmodels.UserToken{ID: token, UserID: userId, ExpiredAt: time.Now().Add(tokenExpireAfter)}); err != nil {
		return "", err
	}
	return token, nil
}

// Login implements protobuf.ApiServer.
func (g *GRPCServer) Login(ctx context.Context, req *apiproto.LoginRequest) (*apiproto.LoginReply, error) {
	if req.Identifier == "" {
		return nil, grpcerrors.ErrorFieldViolationEmpty("identifier")
	}
	if req.Password == "" {
		return nil, grpcerrors.ErrorFieldViolationEmpty("password")
	}
	u := g.DBQueries.User
	user, err := u.WithContext(ctx).Where(u.VerifiedEmail.Eq(req.Identifier)).Or(u.Username.Eq(req.Identifier)).Or(u.NewEmail.Eq(req.Identifier)).First()
	if err != nil {
		logger.Errorln(err)
		return nil, grpcerrors.GormToGRPCError(err, nil)
	}
	if err := passwd.CheckPasswd(req.Password, user.Pswhash); err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, grpcerrors.ErrorPermissionDenied()
	} else if err != nil {
		return nil, grpcerrors.ErrorInternal(true)
	}

	token, err := g.refreshToken(ctx, user.ID, g.DBQueries)
	if err != nil {
		logger.Errorln(err)
		return nil, grpcerrors.ErrorInternal(true)
	}
	return &apiproto.LoginReply{
		UserInfo: apihelpers.UserDbModelToProto(user),
		Token:    token,
	}, nil
}

// Signup implements protobuf.ApiServer.
func (g *GRPCServer) Signup(ctx context.Context, req *apiproto.SignupRequest) (ret *apiproto.SignupReply, err error) {
	if req.Email == "" {
		return nil, grpcerrors.ErrorFieldViolationEmpty("email")
	}
	if !apihelpers.IsValidEmail(req.Email) {
		return nil, grpcerrors.ErrorFieldViolationBadFormat("email")
	}

	if req.Username == "" {
		return nil, grpcerrors.ErrorFieldViolationEmpty("username")
	}
	if !apihelpers.IsValidUsername(req.Username) {
		return nil, grpcerrors.ErrorFieldViolationBadFormat("username")
	}

	if req.Password == "" {
		return nil, grpcerrors.ErrorFieldViolationEmpty("password")
	}
	if !apihelpers.IsValidPassword(req.Password) {
		return nil, grpcerrors.ErrorFieldViolationBadFormat("password")
	}

	olduser, err := g.DBQueries.User.WithContext(ctx).Where(g.DBQueries.User.NewEmail.Eq(req.Email)).Or(g.DBQueries.User.VerifiedEmail.Eq(req.Email)).Or(g.DBQueries.User.Username.Eq(req.Username)).First()
	switch err {
	case gorm.ErrRecordNotFound:
		olduser = nil
	case nil:
		if (olduser.NewEmail != nil && *olduser.NewEmail == req.Email) ||
			(olduser.VerifiedEmail != nil && *olduser.VerifiedEmail == req.Email) {
			return nil, grpcerrors.ErrorFieldViolationAlreadyTaken("email")
		}
		if olduser.Username == req.Username {
			return nil, grpcerrors.ErrorFieldViolationAlreadyTaken("username")
		}
		logger.ShouldNeverHappen()
		return nil, grpcerrors.ErrorFieldViolationAlreadyTaken("")
	default:
		return nil, grpcerrors.GormToGRPCError(err, nil)
	}

	tx := g.DBQueries.Begin()
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	u := tx.User
	userId := uuid.New().String()
	pswHash, err := passwd.HashPasswd(req.Password)
	if err != nil {
		logger.Errorln(err)
		return nil, grpcerrors.ErrorInternal(true)
	}

	user := &dbmodels.User{
		ID:            userId,
		Username:      req.Username,
		VerifiedEmail: nil,
		NewEmail:      &req.Email,
		IsVerified:    false,
		Pswhash:       pswHash,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := u.WithContext(ctx).Create(user); err != nil {
		return nil, grpcerrors.GormToGRPCError(err, map[string]string{
			"email_key":    "email",
			"username_key": "username",
		})
	}

	//make the token first so if the user is not created it s fine
	token, err := g.refreshToken(ctx, userId, tx.Query)
	if err != nil {
		logger.Errorln(err)
		return nil, grpcerrors.ErrorInternal(true)
	}
	sendEmailFunc, err := g.insertVerifyEmail(ctx, userId, req.Email, tx.Query)
	if err != nil {
		logger.Errorln(err)
		return nil, grpcerrors.ErrorInternal(true)
	}
	if err := tx.Commit(); err != nil {
		return nil, grpcerrors.GormToGRPCError(err, map[string]string{
			"email_key":    "email",
			"username_key": "username",
		})
	}
	if err := sendEmailFunc(); err != nil {
		logger.Errorln(err)
		return nil, grpcerrors.ErrorInternal(true)
	}
	return &apiproto.SignupReply{
		UserInfo: apihelpers.UserDbModelToProto(user),
		Token:    token,
	}, nil
}

// RefreshToken implements protobuf.ApiServer.
func (g *GRPCServer) RefreshToken(ctx context.Context, req *apiproto.RefreshTokenRequest) (*apiproto.RefreshTokenReply, error) {
	userId, ok := ctx.Value(apihelpers.UserIdContextKey).(string)
	if !ok {
		logger.Errorln("Cant get the userid from the context")
		return nil, grpcerrors.ErrorInternal(true)
	}
	token, err := g.refreshToken(ctx, userId, g.DBQueries)
	if err != nil {
		logger.Errorln(err)
		return nil, grpcerrors.ErrorInternal(true)
	}
	return &apiproto.RefreshTokenReply{
		Token: token,
	}, nil
}

func (g *GRPCServer) VerifyEmail(ctx context.Context, req *apiproto.VerifyEmailRequest) (ret *apiproto.VerifyEmailReply, err error) {
	if req.Email == "" {
		return nil, grpcerrors.ErrorFieldViolationEmpty("email")
	}
	if req.VerifyId == "" {
		return nil, grpcerrors.ErrorFieldViolationEmpty("verify_id")
	}

	tx := g.DBQueries.Begin()
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	ev := tx.EmailVerification
	verif, err := ev.WithContext(ctx).Where(ev.Email.Eq(req.Email), ev.ID.Eq(req.VerifyId)).First()
	if err != nil {
		return nil, grpcerrors.GormToGRPCError(err, nil)
	}
	if verif.UsedAt != nil {
		return nil, grpcerrors.ErrorFieldViolationAlreadyTaken("verify_id")
	}
	curTime := time.Now()
	verif.UsedAt = &curTime
	if err := ev.WithContext(ctx).Save(verif); err != nil {
		return nil, grpcerrors.GormToGRPCError(err, nil)
	}
	u := tx.User
	user, err := u.WithContext(ctx).Where(u.ID.Eq(verif.UserID)).First()
	if err != nil {
		return nil, grpcerrors.GormToGRPCError(err, nil)
	}
	user.IsVerified = true
	user.VerifiedEmail = &req.Email
	if err := u.WithContext(ctx).Save(user); err != nil {
		return nil, grpcerrors.ErrorInternal(true)
	}
	if err := tx.Commit(); err != nil {
		return nil, grpcerrors.GormToGRPCError(err, nil)
	}
	return &apiproto.VerifyEmailReply{UserInfo: apihelpers.UserDbModelToProto(user)}, nil
}
