# cloudevents-go-protobuf


**NOTE: This code is functional but meant for demonstration until it's decided
whether or not this will be contributed to the Go SDK directly or maintained as
a 3rd party module. If this ends up forever being a 3rd party module then I
will delete this repo and create another under my company's Github org.**

## Protobuf bindings for CloudEvents

This package implements the interfaces required to add protobuf support to the
v2 CloudEvents SDK.

## Usage

### Registering The Extensions

CloudEvents uses a global registry to map content types to extensions. The
easiest way to do this is to import the init modules from this package:

```golang
import _ "github.com/kconwayinvision/cloudevents-go-protobuf/init"
```

Alternatively, if you want finer grained control over the timing then you can
put the equivalent code in your project somewhere:

```golang
package main

import (
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/event/datacodec"
	pbcloudevents "github.com/kconwayinvision/cloudevents-go-protobuf"
)

func main() {
    // ...
	format.Add(pbcloudevents.FormatProtobuf{})
	format.Add(pbcloudevents.FormatProtobufJSON{})
	datacodec.AddDecoder(pbcloudevents.ContentTypeProtobuf, pbcloudevents.Decode)
    datacodec.AddEncoder(pbcloudevents.ContentTypeProtobuf, pbcloudevents.Encode)
    // ...
}
```

### Using With HTTP Transport

#### HTTP Client

```golang
package main

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
    cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
    _ "github.com/kconwayinvision/cloudevents-go-protobuf/init"
    pbcloudevents "github.com/kconwayinvision/cloudevents-go-protobuf"
    mypb "github.com/someone/my-pb-repo"
)

func main() {
    target := "http://localhost"
    p, err := cehttp.New(
        cehttp.WithTarget(target),
        cehttp.WithHeader("Content-Type", pbcloudevents.MediaTypeProtobuf),
        // Optionally swap for the JSON protobuf encoding for human readability.
        // cehttp.WithHeader("Content-Type", pbcloudevents.MediaTypeProtobufJSON),
    )
	if err != nil {
		log.Fatalf("Failed to create protocol, %v", err)
    }
    c, err := cloudevents.NewClient(
        p,
		cloudevents.WithTimeNow(),
	)
	if err != nil {
		log.Fatalf("Failed to create client, %v", err)
    }
    payload := &mypb.Greeting{
        Greeting: "Hello you!"
    }
    event := cloudevents.NewEvent()
    event.SetID("xyz")
    event.SetType("greeting.v1")
    event.SetSource("main")
    if err = event.SetData(pbcloudevents.ContentTypeProtobuf, payload); err != nil {
        log.Fatalf("failed to set protobuf data, %v", err)
    }
    result := c.Send(context.Background(), event)
    if cloudevents.IsUndelivered(result) {
        log.Printf("Failed to deliver request: %v", result)
    } else {
        // Event was delivered, but possibly not accepted and without a response.
        log.Printf("Event delivered at %s, Acknowledged==%t ", time.Now(), cloudevents.IsACK(result))
        var httpResult *cehttp.Result
        if cloudevents.ResultAs(result, &httpResult) {
            log.Printf("Response status code %d", httpResult.StatusCode)
        }
    }
}
```

### HTTP Server

```golang
package main

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
    cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
    _ "github.com/kconwayinvision/cloudevents-go-protobuf/init"
    pbcloudevents "github.com/kconwayinvision/cloudevents-go-protobuf"
    mypb "github.com/someone/my-pb-repo"
)

func main() {
    target := "http://localhost"
    p, err := cehttp.New(
        cehttp.WithPath("/events"),
        cehttp.WithPort(8080),
        cehttp.WithShutdownTimeout(30*time.Second),
        // The server automatically decodes using the incoming Content-Type
        // headers. No special handling is needed to enable protobuf support
        // other than registering the extensions.
    )
	if err != nil {
		log.Fatalf("Failed to create protocol, %v", err)
    }
    c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatalf("Failed to create client, %v", err)
    }
    log.Printf("will listen on :8080\n")
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(context.Background(), receive))
}

func receive(ctx context.Context, event cloudevents.Event) error {
    expected := &mypb.Greeting{}
    if err := event.DataAs(expected); err != nil {
        return fmt.Errorf("did not get expected greeting event. got %s", event.DataSchema())
    }
    return nil
}
```
