package traces

import (
	"github.com/quan-xie/tuba/log"

	"contrib.go.opencensus.io/exporter/zipkin"
	openzipkin "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/trace"
)

type Config struct {
	ServiceName string
	HostPort    string
	ReporterURL string
	Sampler     trace.Sampler
}

func NewZipkinReporter(c *Config) reporter.Reporter {
	reporter := zipkinHTTP.NewReporter(c.ReporterURL)
	endpoint, err := openzipkin.NewEndpoint(c.ServiceName, c.HostPort)
	if err != nil {
		log.Fatalf("Failed to create the local zipkinEndpoint: %v", err)
	}
	ze := zipkin.NewExporter(reporter, endpoint)
	trace.RegisterExporter(ze)

	//  Configure 100% sample rate, otherwise, few traces will be sampled.
	trace.ApplyConfig(trace.Config{DefaultSampler: c.Sampler})
	return reporter
}
