package pbcloudevents

import (
	"fmt"
	"net/url"
	stdtime "time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kconwayinvision/cloudevents-go-protobuf/internal/pb"
)

const (
	datacontenttype = "datacontenttype"
	dataschema      = "dataschema"
	subject         = "subject"
	time            = "time"
)

var (
	zeroTime = stdtime.Time{} //nolint:gochecknoglobals
)

// convert an SDK event to a protobuf variant of the event that can be marshaled.
func sdkToProto(e *event.Event) (*pb.CloudEvent, error) {
	container := &pb.CloudEvent{
		Id:          e.ID(),
		Source:      e.Source(),
		SpecVersion: e.SpecVersion(),
		Type:        e.Type(),
		Attributes:  make(map[string]*pb.CloudEventAttributeValue),
	}
	if e.DataContentType() != "" {
		container.Attributes[datacontenttype], _ = attributeFor(e.DataContentType())
	}
	if e.DataSchema() != "" {
		container.Attributes[dataschema], _ = attributeFor(e.DataSchema())
	}
	if e.Subject() != "" {
		container.Attributes[subject], _ = attributeFor(e.Subject())
	}
	if e.Time() != zeroTime {
		container.Attributes[time], _ = attributeFor(e.Time())
	}
	for name, value := range e.Extensions() {
		attr, err := attributeFor(value)
		if err != nil {
			return nil, fmt.Errorf("failed to encode attribute %s: %s", name, err)
		}
		container.Attributes[name] = attr
	}
	container.Data = &pb.CloudEvent_BinaryData{
		BinaryData: e.Data(),
	}
	if e.DataContentType() == ContentTypeProtobuf {
		anymsg := &anypb.Any{}
		if err := proto.Unmarshal(e.Data(), anymsg); err != nil {
			if e.DataSchema() == "" {
				return nil, fmt.Errorf("cannot encode direct protobuf message without dataschema. set dataschema to the appropriate protobuf type like type.googleapis.com/packge.v1.Type or make sure you are using the appropriate data content type %s", ContentTypeProtobuf)
			}
			anymsg.TypeUrl = e.DataSchema()
			anymsg.Value = e.Data()
		}
		container.Data = &pb.CloudEvent_ProtoData{
			ProtoData: anymsg,
		}
	}
	return container, nil
}

func attributeFor(v interface{}) (*pb.CloudEventAttributeValue, error) {
	vv, err := types.Validate(v)
	if err != nil {
		return nil, err
	}
	attr := &pb.CloudEventAttributeValue{}
	switch vt := vv.(type) {
	case bool:
		attr.Attr = &pb.CloudEventAttributeValue_CeBoolean{
			CeBoolean: vt,
		}
	case int32:
		attr.Attr = &pb.CloudEventAttributeValue_CeInteger{
			CeInteger: vt,
		}
	case string:
		attr.Attr = &pb.CloudEventAttributeValue_CeString{
			CeString: vt,
		}
	case []byte:
		attr.Attr = &pb.CloudEventAttributeValue_CeBytes{
			CeBytes: vt,
		}
	case types.URI:
		attr.Attr = &pb.CloudEventAttributeValue_CeUri{
			CeUri: vt.String(),
		}
	case types.URIRef:
		attr.Attr = &pb.CloudEventAttributeValue_CeUriRef{
			CeUriRef: vt.String(),
		}
	case types.Timestamp:
		attr.Attr = &pb.CloudEventAttributeValue_CeTimestamp{
			CeTimestamp: timestamppb.New(vt.Time),
		}
	default:
		return nil, fmt.Errorf("unsupported attribute type: %T", v)
	}
	return attr, nil
}

func valueFrom(attr *pb.CloudEventAttributeValue) (interface{}, error) {
	var v interface{}
	switch vt := attr.Attr.(type) {
	case *pb.CloudEventAttributeValue_CeBoolean:
		v = vt.CeBoolean
	case *pb.CloudEventAttributeValue_CeInteger:
		v = vt.CeInteger
	case *pb.CloudEventAttributeValue_CeString:
		v = vt.CeString
	case *pb.CloudEventAttributeValue_CeBytes:
		v = vt.CeBytes
	case *pb.CloudEventAttributeValue_CeUri:
		uri, err := url.Parse(vt.CeUri)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URI value %s: %s", vt.CeUri, err.Error())
		}
		v = uri
	case *pb.CloudEventAttributeValue_CeUriRef:
		uri, err := url.Parse(vt.CeUriRef)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URIRef value %s: %s", vt.CeUriRef, err.Error())
		}
		v = types.URIRef{URL: *uri}
	case *pb.CloudEventAttributeValue_CeTimestamp:
		v = vt.CeTimestamp.AsTime()
	default:
		return nil, fmt.Errorf("unsupported attribute type: %T", vt)
	}
	return types.Validate(v)
}

// Convert from a protobuf variant into the generic, SDK event.
func protoToSDK(container *pb.CloudEvent) (*event.Event, error) {
	e := event.New()
	e.SetID(container.Id)
	e.SetSource(container.Source)
	e.SetSpecVersion(container.SpecVersion)
	e.SetType(container.Type)
	// NOTE: There are some issues around missing content type values that are
	// still unresolved. It's an optional field and if unset then it is implied
	// that the encoding used for the envelope was used for the data. However,
	// there is no mapping that exists between content types and the envelope
	// media type. It's also clear what should happen if the content type is
	// unset but it's known that the content type is not the same as the
	// envelope. For example, it is acceptable to write a JSON value as the
	// binary value but encode the envelope with protobuf. In the case of
	// this protobuf encoding implementation, protobuf values are _always_
	// stored as protobuf using the ProtoData. Any use of binary or text data
	// means the value was not protobuf and if content type is not set then have
	// no way of knowing what it actually is. To handle this we're going to set
	// the content type to whatever we find in the attributes with a default to
	// empty string.
	contentType := ""
	if container.Attributes != nil {
		attr := container.Attributes[datacontenttype]
		if attr != nil {
			if stattr, ok := attr.Attr.(*pb.CloudEventAttributeValue_CeString); ok {
				contentType = stattr.CeString
			}
		}
	}
	switch dt := container.Data.(type) {
	case *pb.CloudEvent_BinaryData:
		if err := e.SetData(contentType, dt.BinaryData); err != nil {
			return nil, fmt.Errorf("failed to convert binary type (%s) data: %s", contentType, err)
		}
	case *pb.CloudEvent_TextData:
		if err := e.SetData(contentType, dt.TextData); err != nil {
			return nil, fmt.Errorf("failed to convert text type (%s) data: %s", contentType, err)
		}
	case *pb.CloudEvent_ProtoData:
		if err := e.SetData(ContentTypeProtobuf, dt.ProtoData); err != nil {
			return nil, fmt.Errorf("failed to convert protobuf type (%s) data: %s", contentType, err)
		}
	}
	for name, value := range container.Attributes {
		v, err := valueFrom(value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert attribute %s: %s", name, err)
		}
		switch name {
		case datacontenttype:
			vs, _ := v.(string)
			e.SetDataContentType(vs)
		case dataschema:
			vs, _ := v.(string)
			e.SetDataSchema(vs)
		case subject:
			vs, _ := v.(string)
			e.SetSubject(vs)
		case time:
			vs, _ := v.(types.Timestamp)
			e.SetTime(vs.Time)
		default:
			e.SetExtension(name, v)
		}
	}
	return &e, nil
}
