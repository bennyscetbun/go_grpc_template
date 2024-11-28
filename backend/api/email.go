package api

import (
	"bytes"
	"context"
	"time"

	"github.com/bennyscetbun/xxx_your_app_xxx/backend/generated/database/dbmodels"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/generated/database/dbqueries"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/emails"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/environment"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/random"
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
