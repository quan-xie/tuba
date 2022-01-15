package traces

import (
	"github.com/quan-xie/tuba/log"

	"contrib.go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
)

type Config struct {
	ServiceName          string
	AgentEndpointURI     string
	CollectorEndpointURI string
}

func NewJaegerExporter(c *Config) {
	je, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint:     c.AgentEndpointURI,
		CollectorEndpoint: c.CollectorEndpointURI,
		Process:           jaeger.Process{ServiceName: c.ServiceName},
	})
	if err != nil {
		log.Fatalf("Failed to create the Jaeger exporter: %v", err)
	}

	// And now finally register it as a Trace Exporter
	trace.RegisterExporter(je)
}
