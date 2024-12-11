package grpcerrors

import (
	"strings"

	"github.com/bennyscetbun/xxxyourappyyy/backend/generated/rpc/apiproto"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/logger"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ztrue/tracerr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func GormToGRPCError(err error, uniqueConstraints map[string]string) error {
	pgerr, ok := err.(*pgconn.PgError)
	if ok {
		// https: //www.postgresql.org/docs/current/errcodes-appendix.html
		if pgerr.Code == "23505" {
			for k, v := range uniqueConstraints {
				if strings.Contains(pgerr.ConstraintName, k) {
					return ErrorFieldViolationAlreadyTaken(v)
				}
			}
			return ErrorFieldViolationAlreadyTaken(pgerr.ConstraintName)
		}
	}
	switch err {
	case gorm.ErrRecordNotFound:
		return ErrorNotFound()
	default:
		tracerr.Print(err)
		return ErrorInternal(true)
	}
}

func buildError(code codes.Code, msg string, details *apiproto.ErrorInfo) error {
	st := status.New(code, msg)
	if details == nil {
		return st.Err()
	}
	ret, err := st.WithDetails(details)
	if err != nil {
		logger.Error(err)
		return st.Err()
	}
	return ret.Err()
}

func ErrorFieldViolation(field string, typ apiproto.ErrorFieldViolationType) error {
	br := &apiproto.ErrorInfo{
		Type:           apiproto.ErrorType_ERROR_FIELD_VIOLATION,
		ViolationType:  typ,
		ViolationField: field,
	}
	return buildError(codes.InvalidArgument, typ.Enum().String()+":"+field, br)
}

func ErrorFieldViolationBadFormat(field string) error {
	return ErrorFieldViolation(field, apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_BAD_FORMAT)
}

func ErrorFieldViolationEmpty(field string) error {
	return ErrorFieldViolation(field, apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_CANT_BE_EMPTY)
}

func ErrorFieldViolationAlreadyTaken(field string) error {
	return ErrorFieldViolation(field, apiproto.ErrorFieldViolationType_ERROR_FIELD_VIOLATION_TYPE_ALREADY_TAKEN)
}

func ErrorInternal(retryAble bool) error {
	return buildError(codes.Internal, "Internal Error", &apiproto.ErrorInfo{
		Type:      apiproto.ErrorType_ERROR_INTERNAL,
		Retryable: retryAble,
	})
}

func ErrorNotFound() error {
	return buildError(codes.NotFound, "not found", &apiproto.ErrorInfo{
		Type: apiproto.ErrorType_ERROR_NOT_FOUND,
	})
}

func ErrorInvalidToken() error {
	return buildError(codes.PermissionDenied, "invalid token", &apiproto.ErrorInfo{
		Type: apiproto.ErrorType_ERROR_INVALID_TOKEN,
	})
}

func ErrorPermissionDenied() error {
	return buildError(codes.PermissionDenied, "permission denied", &apiproto.ErrorInfo{
		Type: apiproto.ErrorType_ERROR_PERMISSION_DENIED,
	})
}

func ErrorUnauthenticated() error {
	return buildError(codes.Unauthenticated, "unauthenticated", &apiproto.ErrorInfo{
		Type: apiproto.ErrorType_ERROR_UNAUTHENTICATED,
	})
}
