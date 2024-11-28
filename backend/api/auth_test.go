package api

import (
	"context"
	"testing"

	"github.com/bennyscetbun/xxx_your_app_xxx/backend/generated/rpc/apiproto"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/random"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/testhelpers"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func validSignupRequest() *apiproto.SignupRequest {
	name := "A" + random.RandString(10)
	email := name + "@xxx_your_app_xxx.com"
	passwd := random.RandString(12) + "@A1a"
	return &apiproto.SignupRequest{
		Username: name,
		Email:    email,
		Password: passwd,
	}
}

func TestSignup(t *testing.T) {
	ctx := context.Background()

	client, _, closer, err := server(t)
	if !assert.NoError(t, err) {
		return
	}
	defer closer()

	name := "A" + random.RandString(10)
	email := name + "@xxx_your_app_xxx.com"
	passwd := random.RandString(12) + "@A1a"
	{
		ret, err := client.Signup(ctx, &apiproto.SignupRequest{
			Username: name,
			Email:    email,
			Password: passwd,
		})
		if !assert.Nil(t, err) {
			return
		}
		if !(assert.Equal(t, name, ret.UserInfo.Username) &&
			assert.Equal(t, email, *ret.UserInfo.NewEmail) &&
			assert.Empty(t, ret.UserInfo.VerifiedEmail) &&
			assert.False(t, ret.UserInfo.IsVerified) &&
			assert.NotEmpty(t, ret.UserInfo.UserId) &&
			assert.NotEmpty(t, ret.Token)) {
			return
		}

	}

	if _, err := client.Signup(ctx, &apiproto.SignupRequest{
		Username: name,
		Password: passwd,
	}); !AssertErrorInfo(t, &apiproto.ErrorInfo{
		Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
		ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_CANT_BE_EMPTY,
		ViolationField: "email",
		Retryable:      false,
	}, err) {
		return
	}

	if _, err := client.Signup(ctx, &apiproto.SignupRequest{
		Email:    email,
		Password: passwd,
	}); !AssertErrorInfo(t, &apiproto.ErrorInfo{
		Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
		ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_CANT_BE_EMPTY,
		ViolationField: "username",
		Retryable:      false,
	}, err) {
		return
	}

	if _, err := client.Signup(ctx, &apiproto.SignupRequest{
		Email:    email,
		Username: name,
	}); !AssertErrorInfo(t, &apiproto.ErrorInfo{
		Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
		ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_CANT_BE_EMPTY,
		ViolationField: "password",
		Retryable:      false,
	}, err) {
		return
	}

	if _, err := client.Signup(ctx, &apiproto.SignupRequest{
		Username: name + "poupou",
		Email:    email,
		Password: passwd,
	}); !AssertErrorInfo(t, &apiproto.ErrorInfo{
		Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
		ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_ALREADY_TAKEN,
		ViolationField: "email",
		Retryable:      false,
	}, err) {
		return
	}

	if _, err := client.Signup(ctx, &apiproto.SignupRequest{
		Username: name,
		Email:    "poupou" + email,
		Password: passwd,
	}); !AssertErrorInfo(t, &apiproto.ErrorInfo{
		Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
		ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_ALREADY_TAKEN,
		ViolationField: "username",
		Retryable:      false,
	}, err) {
		return
	}

	{
		ret, err := client.Signup(ctx, &apiproto.SignupRequest{
			Username: "poupou" + name,
			Email:    "poupou" + email,
			Password: passwd,
		})
		if !(assert.Nil(t, err) &&
			assert.Equal(t, "poupou"+name, ret.UserInfo.Username) &&
			assert.Equal(t, "poupou"+email, *ret.UserInfo.NewEmail) &&
			assert.Empty(t, ret.UserInfo.VerifiedEmail) &&
			assert.False(t, ret.UserInfo.IsVerified) &&
			assert.NotEmpty(t, ret.UserInfo.UserId) &&
			assert.NotEmpty(t, ret.Token)) {
			return
		}
	}

	if _, err := client.Signup(ctx, &apiproto.SignupRequest{
		Username: "a" + random.RandString(13),
		Email:    "thisisnotavalidemail",
		Password: passwd,
	}); !AssertErrorInfo(t, &apiproto.ErrorInfo{
		Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
		ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_BAD_FORMAT,
		ViolationField: "email",
		Retryable:      false,
	}, err) {
		return
	}

	if _, err := client.Signup(ctx, &apiproto.SignupRequest{
		Username: "a" + random.RandString(60),
		Email:    "a" + email,
		Password: passwd,
	}); !AssertErrorInfo(t, &apiproto.ErrorInfo{
		Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
		ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_BAD_FORMAT,
		ViolationField: "username",
		Retryable:      false,
	}, err) {
		return
	}

	if _, err := client.Signup(ctx, &apiproto.SignupRequest{
		Username: "a" + random.RandString(1),
		Email:    "a" + email,
		Password: passwd,
	}); !AssertErrorInfo(t, &apiproto.ErrorInfo{
		Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
		ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_BAD_FORMAT,
		ViolationField: "username",
		Retryable:      false,
	}, err) {
		return
	}

	if _, err := client.Signup(ctx, &apiproto.SignupRequest{
		Username: "a" + random.RandString(10),
		Email:    "a" + email,
		Password: "A2d@bdwsiubefdwiubewiuewfbuiefwbiufewbiuewfbiuwefbiuwefbfuiewbwefuiwfeubiwfebiufweiubwfedsfddfsfdssdffsdsdfsfdiubibwuefibwfebkwfjebwefkjwefbkjwefbwefjkwfebjkfwebkjwefbjkfwebjk",
	}); !AssertErrorInfo(t, &apiproto.ErrorInfo{
		Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
		ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_BAD_FORMAT,
		ViolationField: "password",
		Retryable:      false,
	}, err) {
		return
	}
	if _, err := client.Signup(ctx, &apiproto.SignupRequest{
		Username: "a" + random.RandString(10),
		Email:    "a" + email,
		Password: "1223467dbywiwe3983289",
	}); !AssertErrorInfo(t, &apiproto.ErrorInfo{
		Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
		ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_BAD_FORMAT,
		ViolationField: "password",
		Retryable:      false,
	}, err) {
		return
	}
}

func TestLogin(t *testing.T) {
	ctx := context.Background()

	client, _, closer, err := server(t)
	if !assert.NoError(t, err) {
		return
	}
	defer closer()

	name := "A" + random.RandString(10)
	email := name + "@xxx_your_app_xxx.com"
	passwd := random.RandString(12) + "@A1a"

	{
		_, err := client.Login(ctx, &apiproto.LoginRequest{
			Identifier: name,
			Password:   passwd,
		})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:      apiproto.ErrorType_ERROR_NOT_FOUND,
			Retryable: false,
		}, err) {
			return
		}
	}
	{
		_, err := client.Login(ctx, &apiproto.LoginRequest{
			Identifier: email,
			Password:   passwd,
		})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:      apiproto.ErrorType_ERROR_NOT_FOUND,
			Retryable: false,
		}, err) {
			return
		}
	}
	var userId string
	{
		resp, err := client.Signup(ctx, &apiproto.SignupRequest{
			Username: name,
			Email:    email,
			Password: passwd,
		})
		if !assert.NoError(t, err) {
			return
		}
		userId = resp.UserInfo.UserId
	}
	{
		resp, err := client.Login(ctx, &apiproto.LoginRequest{
			Identifier: name,
			Password:   passwd,
		})
		if !(assert.NoError(t, err) && assert.Equal(t, userId, resp.UserInfo.UserId)) {
			return
		}
	}
	{
		resp, err := client.Login(ctx, &apiproto.LoginRequest{
			Identifier: email,
			Password:   passwd,
		})
		if !(assert.NoError(t, err) && assert.Equal(t, userId, resp.UserInfo.UserId)) {
			return
		}
	}
	{
		_, err := client.Login(ctx, &apiproto.LoginRequest{
			Identifier: name,
			Password:   passwd + "a",
		})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:      apiproto.ErrorType_ERROR_PERMISSION_DENIED,
			Retryable: false,
		}, err) {
			return
		}
	}
}

func TestRefreshToken(t *testing.T) {
	ctx := context.Background()

	client, _, closer, err := server(t)
	if !assert.NoError(t, err) {
		return
	}
	defer closer()

	name := "A" + random.RandString(10)
	email := name + "@xxx_your_app_xxx.com"
	passwd := random.RandString(12) + "@A1a"

	{
		_, err := client.RefreshToken(ctx, &apiproto.RefreshTokenRequest{})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:      apiproto.ErrorType_ERROR_UNAUTHENTICATED,
			Retryable: false,
		}, err) {
			return
		}
	}
	var lastToken string
	{
		resp, err := client.Signup(ctx, &apiproto.SignupRequest{
			Username: name,
			Email:    email,
			Password: passwd,
		})
		if !assert.NoError(t, err) {
			return
		}
		lastToken = resp.Token
	}
	ctxWithToken := testhelpers.AddTokenToContext(ctx, lastToken)
	{
		resp, err := client.RefreshToken(ctxWithToken, &apiproto.RefreshTokenRequest{})
		if !(assert.NoError(t, err) && assert.NotEqual(t, lastToken, resp.Token)) {
			return
		}
	}
}

func TestVerifyEmail(t *testing.T) {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.MD{
		"origin": []string{"http://localhost"},
	})

	client, testServersInfo, closer, err := server(t)
	if !assert.NoError(t, err) {
		return
	}
	defer closer()

	name := "A" + random.RandString(10)
	email := name + "@xxx_your_app_xxx.com"
	passwd := random.RandString(12) + "@A1a"

	{
		{
			resp, err := client.Signup(ctx, &apiproto.SignupRequest{
				Username: name,
				Email:    email,
				Password: passwd,
			})
			if !assert.NoError(t, err) {
				return
			}
			if !assert.False(t, resp.UserInfo.IsVerified) {
				return
			}
		}
		tokenFromMail, emailFromMail := extractVerifyEmailInfo(t, testServersInfo)
		{
			_, err := client.VerifyEmail(ctx, &apiproto.VerifyEmailRequest{
				VerifyId: "pouet",
				Email:    emailFromMail,
			})
			if !AssertErrorInfo(t, &apiproto.ErrorInfo{
				Type:      apiproto.ErrorType_ERROR_NOT_FOUND,
				Retryable: false,
			}, err) {
				return
			}
		}
		{
			_, err := client.VerifyEmail(ctx, &apiproto.VerifyEmailRequest{
				VerifyId: tokenFromMail,
				Email:    "a" + emailFromMail,
			})
			if !AssertErrorInfo(t, &apiproto.ErrorInfo{
				Type:      apiproto.ErrorType_ERROR_NOT_FOUND,
				Retryable: false,
			}, err) {
				return
			}
		}
		{
			resp, err := client.VerifyEmail(ctx, &apiproto.VerifyEmailRequest{
				VerifyId: tokenFromMail,
				Email:    emailFromMail,
			})
			if !assert.NoError(t, err) {
				return
			}
			if !(assert.True(t, resp.UserInfo.IsVerified) && assert.Equal(t, email, *resp.UserInfo.VerifiedEmail)) {
				return
			}
		}
		{
			_, err := client.VerifyEmail(ctx, &apiproto.VerifyEmailRequest{
				VerifyId: tokenFromMail,
				Email:    emailFromMail,
			})
			if !AssertErrorInfo(t, &apiproto.ErrorInfo{
				Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
				ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_ALREADY_TAKEN,
				ViolationField: "verify_id",
				Retryable:      false,
			}, err) {
				return
			}
		}
	}
	{
		_, err := client.VerifyEmail(ctx, &apiproto.VerifyEmailRequest{
			VerifyId: "",
			Email:    email,
		})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
			ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_CANT_BE_EMPTY,
			ViolationField: "verify_id",
			Retryable:      false,
		}, err) {
			return
		}
	}
	{
		_, err := client.VerifyEmail(ctx, &apiproto.VerifyEmailRequest{
			VerifyId: random.RandString(10),
			Email:    "",
		})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
			ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_CANT_BE_EMPTY,
			ViolationField: "email",
			Retryable:      false,
		}, err) {
			return
		}
	}
	{
		_, err := client.VerifyEmail(ctx, &apiproto.VerifyEmailRequest{
			VerifyId: random.RandString(10),
			Email:    email,
		})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:      apiproto.ErrorType_ERROR_NOT_FOUND,
			Retryable: false,
		}, err) {
			return
		}
	}
}
