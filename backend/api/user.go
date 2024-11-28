package api

import (
	"context"

	"github.com/bennyscetbun/xxx_your_app_xxx/backend/generated/rpc/apiproto"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/apihelpers"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/grpcerrors"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/logger"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/passwd"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (g *GRPCServer) ChangeEmail(ctx context.Context, req *apiproto.ChangeEmailRequest) (ret *apiproto.ChangeEmailReply, err error) {
	constraintCheck := map[string]string{
		"email_key": "new_email",
	}

	if req.NewEmail == "" {
		return nil, grpcerrors.ErrorFieldViolationEmpty("new_email")
	}
	if !apihelpers.IsValidEmail(req.NewEmail) {
		return nil, grpcerrors.ErrorFieldViolationBadFormat("new_email")
	}
	userID, ok := ctx.Value(apihelpers.UserIdContextKey).(string)
	if !ok {
		return nil, grpcerrors.ErrorInternal(true)
	}
	olduser, err := g.DBQueries.User.WithContext(ctx).Where(g.DBQueries.User.NewEmail.Eq(req.NewEmail)).Or(g.DBQueries.User.VerifiedEmail.Eq(req.NewEmail)).First()
	switch err {
	case gorm.ErrRecordNotFound:
		olduser = nil
	case nil:
		if olduser.VerifiedEmail != nil && *olduser.VerifiedEmail == req.NewEmail {
			return nil, grpcerrors.ErrorFieldViolationAlreadyTaken("new_email")
		}
		if olduser.ID != userID && olduser.NewEmail != nil && *olduser.NewEmail == req.NewEmail {
			return nil, grpcerrors.ErrorFieldViolationAlreadyTaken("new_email")
		}
	default:
		return nil, grpcerrors.GormToGRPCError(err, constraintCheck)
	}

	u := g.DBQueries.User
	user, err := u.WithContext(ctx).Where(u.ID.Eq(userID)).First()
	if err != nil {
		return nil, grpcerrors.GormToGRPCError(err, constraintCheck)
	}
	tx := g.DBQueries.Begin()
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	user.NewEmail = &req.NewEmail
	if err := tx.User.WithContext(ctx).Save(user); err != nil {
		return nil, grpcerrors.GormToGRPCError(err, constraintCheck)
	}

	sendEmailFunc, err := g.insertVerifyEmail(ctx, userID, req.NewEmail, tx.Query)
	if err != nil {
		logger.Errorln(err)
		return nil, grpcerrors.ErrorInternal(true)
	}
	if err := tx.Commit(); err != nil {
		return nil, grpcerrors.GormToGRPCError(err, constraintCheck)
	}
	if err := sendEmailFunc(); err != nil {
		logger.Errorln(err)
		return nil, grpcerrors.ErrorInternal(true)
	}

	return &apiproto.ChangeEmailReply{
		UserInfo: apihelpers.UserDbModelToProto(user),
	}, nil
}

func (g *GRPCServer) ChangePassword(ctx context.Context, req *apiproto.ChangePasswordRequest) (*apiproto.ChangePasswordReply, error) {
	if req.OldPassword == "" {
		return nil, grpcerrors.ErrorFieldViolationEmpty("old_password")
	}
	if req.NewPassword == "" {
		return nil, grpcerrors.ErrorFieldViolationEmpty("new_password")
	}
	if !apihelpers.IsValidPassword(req.NewPassword) {
		return nil, grpcerrors.ErrorFieldViolationBadFormat("new_password")
	}
	userID, ok := ctx.Value(apihelpers.UserIdContextKey).(string)
	if !ok {
		return nil, grpcerrors.ErrorInternal(true)
	}
	u := g.DBQueries.User
	user, err := u.WithContext(ctx).Where(u.ID.Eq(userID)).First()
	if err != nil {
		logger.Errorln(err)
		return nil, grpcerrors.GormToGRPCError(err, nil)
	}
	if err := passwd.CheckPasswd(req.OldPassword, user.Pswhash); err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, grpcerrors.ErrorPermissionDenied()
	} else if err != nil {
		return nil, grpcerrors.ErrorInternal(true)
	}
	pswHash, err := passwd.HashPasswd(req.NewPassword)
	if err != nil {
		return nil, err
	}
	user.Pswhash = pswHash
	if err := u.WithContext(ctx).Save(user); err != nil {
		return nil, grpcerrors.GormToGRPCError(err, nil)
	}

	return &apiproto.ChangePasswordReply{}, nil
}

func (g *GRPCServer) ChangeUsername(ctx context.Context, req *apiproto.ChangeUsernameRequest) (*apiproto.ChangeUsernameReply, error) {
	if req.NewUsername == "" {
		return nil, grpcerrors.ErrorFieldViolationEmpty("new_username")
	}
	if !apihelpers.IsValidUsername(req.NewUsername) {
		return nil, grpcerrors.ErrorFieldViolationBadFormat("new_username")
	}
	userID, ok := ctx.Value(apihelpers.UserIdContextKey).(string)
	if !ok {
		return nil, grpcerrors.ErrorInternal(true)
	}
	u := g.DBQueries.User
	user, err := u.WithContext(ctx).Where(u.ID.Eq(userID)).First()
	if err != nil {
		logger.Errorln(err)
		return nil, grpcerrors.GormToGRPCError(err, nil)
	}
	user.Username = req.NewUsername
	if err := u.WithContext(ctx).Save(user); err != nil {
		logger.Errorln(err)
		return nil, grpcerrors.GormToGRPCError(err, map[string]string{
			"username_key": "new_username",
		})
	}

	return &apiproto.ChangeUsernameReply{
		UserInfo: apihelpers.UserDbModelToProto(user),
	}, nil
}
