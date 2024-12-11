package api

import (
	"context"
	"time"

	"github.com/bennyscetbun/xxxyourappyyy/backend/generated/database/dbmodels"
	"github.com/bennyscetbun/xxxyourappyyy/backend/generated/rpc/apiproto"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/apihelpers"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/grpcerrors"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/logger"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/passwd"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ztrue/tracerr"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	authorizationHeader = "authorization"
	tokenExpireAfter    = time.Minute * 10
	tokenUserIDKey      = "user_id"
)

func (g *GRPCServer) refreshToken(_ context.Context, userID string) (string, error) {
	// build JWT with necessary claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		tokenUserIDKey: userID,
		"iat":          time.Now().Unix(),
		"exp":          time.Now().Add(tokenExpireAfter).Unix(), // expire after 15 minutes.
		"other":        time.Now().UnixMilli(),
	}, nil)

	// sign token using the server's secret key.
	signed, err := token.SignedString(g.TokenSecret)
	if err != nil {
		return "", tracerr.Errorf("failed to sign JWT: %w", err)
	}
	return signed, nil
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

	token, err := g.refreshToken(ctx, user.ID)
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
	token, err := g.refreshToken(ctx, userId)
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

func (g *GRPCServer) RefreshToken(ctx context.Context, req *apiproto.RefreshTokenRequest) (*apiproto.RefreshTokenReply, error) {
	userId, ok := ctx.Value(apihelpers.UserIdContextKey).(string)
	if !ok {
		logger.Errorln("Cant get the userid from the context")
		return nil, grpcerrors.ErrorInternal(true)
	}
	token, err := g.refreshToken(ctx, userId)
	if err != nil {
		logger.Errorln(err)
		return nil, grpcerrors.ErrorInternal(true)
	}
	return &apiproto.RefreshTokenReply{
		Token: token,
	}, nil
}
