package init

import (
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/event/datacodec"

	pbcloudevents "github.com/kconwayinvision/cloudevents-go-protobuf"
)

func init() {
	format.Add(pbcloudevents.FormatProtobuf{})
	format.Add(pbcloudevents.FormatProtobufJSON{})
	datacodec.AddDecoder(pbcloudevents.ContentTypeProtobuf, pbcloudevents.Decode)
	datacodec.AddEncoder(pbcloudevents.ContentTypeProtobuf, pbcloudevents.Encode)
}
