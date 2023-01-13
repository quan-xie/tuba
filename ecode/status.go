package ecode

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/status"

	"github.com/quan-xie/tuba/ecode/types"
	"google.golang.org/protobuf/proto"
)

// Error new status with code and message
func Error(code Code, message string) *Status {
	return &Status{s: &types.Status{Code: int32(code.Code()), Message: message}}
}

// Errorf new status with code and message
func Errorf(code Code, format string, args ...interface{}) *Status {
	return Error(code, fmt.Sprintf(format, args...))
}

var _ Codes = &Status{}

// Status statusError is an alias of a status proto
// implement ecode.Codes
type Status struct {
	s *types.Status
}

// Error implement error
func (s *Status) Error() string {
	return s.Message()
}

// Code return error code
func (s *Status) Code() int32 {
	return s.s.Code
}

// Message return error message for developer
func (s *Status) Message() string {
	if s.s.Message == "" {
		return strconv.Itoa(int(s.s.Code))
	}
	return s.s.Message
}

func (s *Status) Details() []interface{} {
	if s == nil || s.s == nil {
		return nil
	}
	details := make([]interface{}, 0, len(s.s.Details))
	for _, any := range s.s.Details {
		detail, err := any.UnmarshalNew()
		if err != nil {
			details = append(details, err)
			continue
		}
		details = append(details, detail)
	}
	return details
}

// Proto return origin protobuf message
func (s *Status) Proto() *types.Status {
	return s.s
}

// FromCode create status from ecode
func FromCode(code Code) *Status {
	return &Status{s: &types.Status{Code: int32(code), Message: code.Message()}}
}

// FromProto new status from grpc detail
func FromProto(pbMsg proto.Message) Codes {
	if msg, ok := pbMsg.(*types.Status); ok {
		if msg.Message == "" || msg.Message == strconv.FormatInt(int64(msg.Code), 10) {
			// NOTE: if message is empty convert to pure Code, will get message from config center.
			return Code(msg.Code)
		}
		return &Status{s: msg}
	}
	return Errorf(ServerErr, "invalid proto message get %v", pbMsg)
}

func GRPCStatus(err error, code Code) error {
	if err == nil {
		return nil
	}
	s := &spb.Status{
		Code:    code.Code(),
		Message: code.Message(),
	}

	ec, ok := errors.Cause(err).(Codes)
	if ok {
		s.Code = ec.Code()
		s.Message = ec.Message()
	}

	return status.FromProto(s).Err()
}

// FromStatus .
func FromStatus(err error) error {
	if err == nil {
		return nil
	}

	statusCode, ok := status.FromError(err)

	if !ok {
		return err
	}

	return Code(statusCode.Proto().Code)
}
