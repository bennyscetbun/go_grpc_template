package api

import (
	"context"
	"testing"

	"github.com/bennyscetbun/xxxyourappyyy/backend/generated/rpc/apiproto"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/random"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func signupAndGetCtxWithValidRequest(client apiproto.ApiClient, t *testing.T, testServersInfo *testServers, ctx context.Context, verify bool, purgemessage bool) context.Context {
	return signupAndGetCtx(client, t, testServersInfo, ctx, verify, purgemessage, validSignupRequest())
}

func signupAndGetCtx(client apiproto.ApiClient, t *testing.T, testServersInfo *testServers, ctx context.Context, verify bool, purgemessage bool, req *apiproto.SignupRequest) context.Context {
	resp, err := client.Signup(ctx, req)
	if !assert.NoError(t, err) {
		t.FailNow()
		return nil
	}
	ctx = testhelpers.AddTokenToContext(ctx, resp.Token)
	if verify {
		verifyLastEmailSignup(t, testServersInfo, ctx, client)
	} else if purgemessage {
		purgeOneMessage(t, testServersInfo)
	}
	return ctx
}

func TestChangeEmail(t *testing.T) {
	ctx := context.Background()

	client, testServersInfo, closer, err := server(t)
	if !assert.NoError(t, err) {
		return
	}
	defer closer()

	{
		_, err := client.ChangeEmail(ctx, &apiproto.ChangeEmailRequest{})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:      apiproto.ErrorType_ERROR_UNAUTHENTICATED,
			Retryable: false,
		}, err) {
			return
		}
	}
	ctx = signupAndGetCtxWithValidRequest(client, t, testServersInfo, ctx, false, false)
	{
		_, err := client.ChangeEmail(ctx, &apiproto.ChangeEmailRequest{})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
			ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_CANT_BE_EMPTY,
			ViolationField: "new_email",
			Retryable:      false,
		}, err) {
			return
		}
	}
	{
		_, err := client.ChangeEmail(ctx, &apiproto.ChangeEmailRequest{
			NewEmail: "prout",
		})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
			ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_BAD_FORMAT,
			ViolationField: "new_email",
			Retryable:      false,
		}, err) {
			return
		}
	}
	verifiedEmail := verifyLastEmailSignup(t, testServersInfo, ctx, client)
	newEmail := "A" + validSignupRequest().Email
	{

		{
			resp, err := client.ChangeEmail(ctx, &apiproto.ChangeEmailRequest{NewEmail: newEmail})
			if !assert.NoError(t, err) {
				return
			}
			if !assert.Equal(t, *resp.UserInfo.NewEmail, newEmail) {
				return
			}
			if !assert.Equal(t, *resp.UserInfo.VerifiedEmail, verifiedEmail) {
				return
			}
			purgeOneMessage(t, testServersInfo)
		}
		{
			ctx2 := signupAndGetCtxWithValidRequest(client, t, testServersInfo, context.Background(), true, false)
			_, err := client.ChangeEmail(ctx2, &apiproto.ChangeEmailRequest{NewEmail: newEmail})
			if !AssertErrorInfo(t, &apiproto.ErrorInfo{
				Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
				ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_ALREADY_TAKEN,
				ViolationField: "new_email",
				Retryable:      false,
			}, err) {
				return
			}
		}
		{
			resp, err := client.ChangeEmail(ctx, &apiproto.ChangeEmailRequest{NewEmail: newEmail})
			if !assert.NoError(t, err) {
				return
			}
			if !assert.Equal(t, *resp.UserInfo.NewEmail, newEmail) {
				return
			}
			if !assert.Equal(t, *resp.UserInfo.VerifiedEmail, verifiedEmail) {
				return
			}
		}
		newVerifiedEmail := verifyLastEmailSignup(t, testServersInfo, ctx, client)
		if !assert.Equal(t, newVerifiedEmail, newEmail) {
			return
		}
		{
			_, err := client.ChangeEmail(ctx, &apiproto.ChangeEmailRequest{NewEmail: newEmail})
			if !AssertErrorInfo(t, &apiproto.ErrorInfo{
				Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
				ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_ALREADY_TAKEN,
				ViolationField: "new_email",
				Retryable:      false,
			}, err) {
				return
			}
		}
	}

	{
		newNewEmail := "B" + validSignupRequest().Email
		resp, err := client.ChangeEmail(ctx, &apiproto.ChangeEmailRequest{NewEmail: newNewEmail})
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, *resp.UserInfo.NewEmail, newNewEmail) {
			return
		}
		purgeOneMessage(t, testServersInfo)
	}
	{
		ctx2 := signupAndGetCtxWithValidRequest(client, t, testServersInfo, context.Background(), true, false)
		_, err := client.ChangeEmail(ctx2, &apiproto.ChangeEmailRequest{NewEmail: newEmail})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
			ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_ALREADY_TAKEN,
			ViolationField: "new_email",
			Retryable:      false,
		}, err) {
			return
		}
	}
}

func TestChangePassword(t *testing.T) {
	ctx := context.Background()

	client, testServersInfo, closer, err := server(t)
	if !assert.NoError(t, err) {
		return
	}
	defer closer()
	{
		_, err := client.ChangePassword(ctx, &apiproto.ChangePasswordRequest{})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:      apiproto.ErrorType_ERROR_UNAUTHENTICATED,
			Retryable: false,
		}, err) {
			return
		}
	}

	signupRequest := validSignupRequest()
	ctx = signupAndGetCtx(client, t, testServersInfo, ctx, false, false, signupRequest)
	{
		_, err := client.ChangePassword(ctx, &apiproto.ChangePasswordRequest{})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
			ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_CANT_BE_EMPTY,
			ViolationField: "old_password",
			Retryable:      false,
		}, err) {
			return
		}
	}
	{
		_, err := client.ChangePassword(ctx, &apiproto.ChangePasswordRequest{
			OldPassword: random.RandString(10),
		})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
			ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_CANT_BE_EMPTY,
			ViolationField: "new_password",
			Retryable:      false,
		}, err) {
			return
		}
	}
	{
		_, err := client.ChangePassword(ctx, &apiproto.ChangePasswordRequest{
			NewPassword: random.RandString(10),
		})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
			ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_CANT_BE_EMPTY,
			ViolationField: "old_password",
			Retryable:      false,
		}, err) {
			return
		}
	}
	{
		_, err := client.ChangePassword(ctx, &apiproto.ChangePasswordRequest{
			OldPassword: random.RandString(13) + "@sH1",
			NewPassword: random.RandString(10),
		})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
			ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_BAD_FORMAT,
			ViolationField: "new_password",
			Retryable:      false,
		}, err) {
			return
		}
	}
	{
		_, err := client.ChangePassword(ctx, &apiproto.ChangePasswordRequest{
			OldPassword: random.RandString(13) + "@sH1",
			NewPassword: random.RandString(13) + "@sH1",
		})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:      apiproto.ErrorType_ERROR_PERMISSION_DENIED,
			Retryable: false,
		}, err) {
			return
		}
	}
	{
		newPasswd := random.RandString(13) + "@sH1"
		_, err := client.ChangePassword(ctx, &apiproto.ChangePasswordRequest{
			OldPassword: signupRequest.Password,
			NewPassword: newPasswd,
		})
		if !assert.NoError(t, err) {
			return
		}
		{
			_, err := client.Login(context.Background(), &apiproto.LoginRequest{
				Identifier: signupRequest.Email,
				Password:   signupRequest.Password,
			})
			if !AssertErrorInfo(t, &apiproto.ErrorInfo{
				Type:      apiproto.ErrorType_ERROR_PERMISSION_DENIED,
				Retryable: false,
			}, err) {
				return
			}
		}
		{
			loginResp, err := client.Login(context.Background(), &apiproto.LoginRequest{
				Identifier: signupRequest.Email,
				Password:   newPasswd,
			})
			if !assert.NoError(t, err) {
				return
			}
			if !assert.Equal(t, signupRequest.Username, loginResp.UserInfo.Username) {
				return
			}
			if !assert.NotNil(t, loginResp.UserInfo.NewEmail) {
				return
			}
			if !assert.Equal(t, signupRequest.Email, *loginResp.UserInfo.NewEmail) {
				return
			}
		}
	}
}

func TestChangeUsername(t *testing.T) {
	ctx := context.Background()

	client, testServersInfo, closer, err := server(t)
	if !assert.NoError(t, err) {
		return
	}
	defer closer()
	{
		_, err := client.ChangeUsername(ctx, &apiproto.ChangeUsernameRequest{})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:      apiproto.ErrorType_ERROR_UNAUTHENTICATED,
			Retryable: false,
		}, err) {
			return
		}
	}
	signupRequest := validSignupRequest()
	ctx = signupAndGetCtx(client, t, testServersInfo, ctx, false, false, signupRequest)
	{
		_, err := client.ChangeUsername(ctx, &apiproto.ChangeUsernameRequest{})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
			ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_CANT_BE_EMPTY,
			ViolationField: "new_username",
			Retryable:      false,
		}, err) {
			return
		}
	}
	{
		_, err := client.ChangeUsername(ctx, &apiproto.ChangeUsernameRequest{
			NewUsername: signupRequest.Username,
		})
		if !assert.NoError(t, err) {
			return
		}
	}
	{
		_, err := client.ChangeUsername(ctx, &apiproto.ChangeUsernameRequest{
			NewUsername: "a",
		})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
			ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_BAD_FORMAT,
			ViolationField: "new_username",
			Retryable:      false,
		}, err) {
			return
		}
	}
	{
		_, err := client.ChangeUsername(ctx, &apiproto.ChangeUsernameRequest{
			NewUsername: "0" + random.RandString(10),
		})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
			ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_BAD_FORMAT,
			ViolationField: "new_username",
			Retryable:      false,
		}, err) {
			return
		}
	}
	{
		ctx2 := signupAndGetCtxWithValidRequest(client, t, testServersInfo, context.Background(), false, false)
		_, err := client.ChangeUsername(ctx2, &apiproto.ChangeUsernameRequest{
			NewUsername: signupRequest.Username,
		})
		if !AssertErrorInfo(t, &apiproto.ErrorInfo{
			Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
			ViolationType:  apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_ALREADY_TAKEN,
			ViolationField: "new_username",
			Retryable:      false,
		}, err) {
			return
		}
	}
	{
		newUsername := "a" + signupRequest.Username
		resp, err := client.ChangeUsername(ctx, &apiproto.ChangeUsernameRequest{
			NewUsername: newUsername,
		})
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, newUsername, resp.UserInfo.Username) {
			return
		}
		if !assert.NotNil(t, resp.UserInfo.NewEmail) {
			return
		}
		if !assert.Equal(t, signupRequest.Email, *resp.UserInfo.NewEmail) {
			return
		}
	}
}
