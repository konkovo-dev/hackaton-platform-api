package mappers

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TimeToProto(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
