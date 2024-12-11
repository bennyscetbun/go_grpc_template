package api

import (
	"bytes"
	"context"
	"time"

	"github.com/bennyscetbun/xxxyourappyyy/backend/generated/database/dbmodels"
	"github.com/bennyscetbun/xxxyourappyyy/backend/generated/database/dbqueries"
	"github.com/bennyscetbun/xxxyourappyyy/backend/generated/rpc/apiproto"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/apihelpers"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/emails"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/environment"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/grpcerrors"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/random"
	"github.com/ztrue/tracerr"
)

const (
	verificationExpireAfter = time.Minute * 60
)

var (
	verificationFromEmail = environment.MustGetenvString("VERIFICATION_EMAIL", "example@example.com")
)

func (g *GRPCServer) insertVerifyEmail(ctx context.Context, userID, email string, db *dbqueries.Query) (func() error, error) {
	token := random.RandString(20)
	ev := db.EmailVerification
	if err := ev.WithContext(ctx).Create(&dbmodels.EmailVerification{ID: token, UserID: userID, Email: email, ExpiredAt: time.Now().Add(verificationExpireAfter)}); err != nil {
		return nil, tracerr.Wrap(err)
	}

	return func() error {
		verificationTmplFileName, err := environment.GetenvString("VERIFICATION_TEMPLATE_FILE", "verification_email.tmpl.html")
		if err != nil {
			return tracerr.Wrap(err)
		}
		var tpl bytes.Buffer
		if err := g.Templates.ExecuteTemplate(&tpl, verificationTmplFileName, map[string]string{
			"Host":  environment.MustGetenvString("EMAIL_VERIFICATION_HOST", "http://localhost:8080"),
			"Token": token,
			"Email": email,
		}); err != nil {
			return tracerr.Wrap(err)
		}
		if err := emails.SendEmail(ctx, verificationFromEmail, email, tpl.String()); err != nil {
			return tracerr.Wrap(err)
		}
		return nil
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
