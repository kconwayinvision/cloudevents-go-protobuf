package pbcloudevents

import (
	"github.com/cloudevents/sdk-go/v2/event"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/kconwayinvision/cloudevents-go-protobuf/internal/pb"
)

const (
	// MediaTypeProtobuf is used to indicate that the event is in protobuf
	// encoding. Using a non-standard name to guard against any breaking changes
	// in the official protobuf schema before it becomes 1.0. The short SHA
	// identifies the commit where protobuf was added to the specification.
	MediaTypeProtobuf = "application/cloudevents+protobuf-ad5f142"
	// MediaTypeProtobufJSON is a non-standard encoding that uses the protobuf
	// JSON encoding rather than binary. This is useful for maintaining human
	// readability of messages.
	MediaTypeProtobufJSON = "application/cloudevents+protobuf+json-ad5f142"
)

type FormatProtobuf struct{}

func (FormatProtobuf) MediaType() string {
	return MediaTypeProtobuf
}

func (FormatProtobuf) Marshal(e *event.Event) ([]byte, error) {
	pbe, err := sdkToProto(e)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(pbe)
}
func (FormatProtobuf) Unmarshal(b []byte, e *event.Event) error {
	pbe := &pb.CloudEvent{}
	if err := proto.Unmarshal(b, pbe); err != nil {
		return err
	}
	e2, err := protoToSDK(pbe)
	if err != nil {
		return err
	}
	*e = *e2
	return nil
}

type FormatProtobufJSON struct{}

func (FormatProtobufJSON) MediaType() string {
	return MediaTypeProtobuf
}

func (FormatProtobufJSON) Marshal(e *event.Event) ([]byte, error) {
	pbe, err := sdkToProto(e)
	if err != nil {
		return nil, err
	}
	return protojson.Marshal(pbe)
}
func (FormatProtobufJSON) Unmarshal(b []byte, e *event.Event) error {
	pbe := &pb.CloudEvent{}
	if err := protojson.Unmarshal(b, pbe); err != nil {
		return err
	}
	e2, err := protoToSDK(pbe)
	if err != nil {
		return err
	}
	*e = *e2
	return nil
}
