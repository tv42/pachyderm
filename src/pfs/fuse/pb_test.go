package fuse_test

import (
	"io"
	"time"

	"github.com/pachyderm/pachyderm/src/pfs"
	"go.pedge.io/pb/go/google/protobuf"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// Helpers for mocking grpc APIs in tests

// PBTimestamp converts a time into protobuf format.
func PBTimestamp(t time.Time) *google_protobuf.Timestamp {
	return &google_protobuf.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}

// getFileClient is a mock streaming result from
// pfs.APIClient.GetFile.
type getFileClient struct {
	data []byte
}

var _ pfs.API_GetFileClient = (*getFileClient)(nil)

func (g *getFileClient) Recv() (*google_protobuf.BytesValue, error) {
	if g.data == nil {
		return nil, io.EOF
	}
	b := &google_protobuf.BytesValue{
		Value: g.data,
	}
	g.data = nil
	return b, nil
}

func (*getFileClient) Header() (metadata.MD, error) {
	panic("not implemented in mock")
}

func (*getFileClient) Trailer() metadata.MD {
	panic("not implemented in mock")
}

func (*getFileClient) CloseSend() error {
	panic("not implemented in mock")
}

func (*getFileClient) Context() context.Context {
	panic("not implemented in mock")
}

func (*getFileClient) SendMsg(m interface{}) error {
	panic("not implemented in mock")
}

func (*getFileClient) RecvMsg(m interface{}) error {
	panic("not implemented in mock")
}
