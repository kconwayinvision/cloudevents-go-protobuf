package pbcloudevents

import (
	"net/url"
	"sync"
	"testing"
	stdtime "time"

	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/event/datacodec"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/stretchr/testify/require"

	"github.com/kconwayinvision/cloudevents-go-protobuf/internal/pb"
)

var registerForTests = &sync.Once{} //nolint:gochecknoglobals

func register() {
	format.Add(FormatProtobuf{})
	format.Add(FormatProtobufJSON{})
	datacodec.AddDecoder(ContentTypeProtobuf, Decode)
	datacodec.AddEncoder(ContentTypeProtobuf, Encode)
}

func TestRoundTripProtobuf(t *testing.T) {
	registerForTests.Do(register)
	const test = "test"
	e := event.New()
	e.SetID(test)
	e.SetDataContentType(ContentTypeProtobuf)
	e.SetTime(stdtime.Date(2020, 1, 1, 1, 1, 1, 1, stdtime.UTC))
	e.SetExtension(test, test)
	e.SetExtension("int", 1)
	e.SetExtension("bool", true)
	e.SetExtension("URI", &url.URL{
		Host: "test-uri",
	})
	e.SetExtension("URIRef", types.URIRef{URL: url.URL{
		Host: "test-uriref",
	}})
	e.SetExtension("bytes", []byte(test))
	e.SetExtension("timestamp", stdtime.Date(2020, 2, 1, 1, 1, 1, 1, stdtime.UTC))
	e.SetSubject(test)
	e.SetSource(test)
	e.SetType(test)
	dataObj := &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeBoolean{
			CeBoolean: true,
		},
	}
	e.SetDataSchema("type.googleapis.com/" + string(dataObj.ProtoReflect().Descriptor().FullName())) // simulate the anypb behavior
	require.NoError(t, e.SetData(ContentTypeProtobuf, dataObj))
	pbe, err := sdkToProto(&e)
	require.NoError(t, err)
	e2, err := protoToSDK(pbe)
	require.NoError(t, err)

	eb, err := e.MarshalJSON()
	require.NoError(t, err)
	eb2, err := e2.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, string(eb), string(eb2))
	dataObj2 := &pb.CloudEventAttributeValue{}
	require.NoError(t, e2.DataAs(dataObj2))
	require.IsType(t, &pb.CloudEventAttributeValue_CeBoolean{}, dataObj2.Attr)
}

func TestRoundTripTextData(t *testing.T) {
	registerForTests.Do(register)
	test := "test"
	e := event.New()
	e.SetID(test)
	e.SetDataContentType(ContentTypeProtobuf)
	e.SetSource(test)
	e.SetType(test)
	value := `hello test!`
	require.NoError(t, e.SetData("text/plain", value))
	pbe, err := sdkToProto(&e)
	require.NoError(t, err)
	e2, err := protoToSDK(pbe)
	require.NoError(t, err)

	var value2 string
	require.NoError(t, e2.DataAs(&value2))
	require.Equal(t, value, value2)
}

type testObj struct {
	Value string `json:"value"`
}

func TestRoundTripBinaryData(t *testing.T) {
	registerForTests.Do(register)
	test := "test"
	e := event.New()
	e.SetID(test)
	e.SetDataContentType(ContentTypeProtobuf)
	e.SetSource(test)
	e.SetType(test)
	value := testObj{Value: test}
	require.NoError(t, e.SetData("application/json", value))
	pbe, err := sdkToProto(&e)
	require.NoError(t, err)
	e2, err := protoToSDK(pbe)
	require.NoError(t, err)

	var value2 testObj
	require.NoError(t, e2.DataAs(&value2))
	require.Equal(t, value, value2)
}
