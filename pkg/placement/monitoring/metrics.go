/*
Copyright 2021 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package monitoring

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var (
	meter = otel.Meter("dapr/placement")

	runtimesTotal       metric.Int64Gauge
	actorRuntimesTotal  metric.Int64Gauge
	actorHeartbeatTotal metric.Int64Gauge

	// Metrics attributes
	appIDKey     = attribute.Key("app_id")
	actorTypeKey = attribute.Key("actor_type")
	hostNameKey  = attribute.Key("host_name")
	namespaceKey = attribute.Key("host_namespace")
	podNameKey   = attribute.Key("pod_name")
)

// InitMetrics initializes the placement service metrics.
func InitMetrics() error {
	// Create a new Prometheus exporter
	promExporter, err := prometheus.New(prometheus.WithNamespace("dapr"))
	if err != nil {
		return fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	// Create a new MeterProvider
	res, err := resource.New(context.Background(),
		resource.WithAttributes(semconv.ServiceNameKey.String("dapr-placement")),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(promExporter),
	)

	// Set the global MeterProvider
	otel.SetMeterProvider(mp)

	// Create metrics
	runtimesTotal, err = meter.Int64Gauge("placement.runtimes_total",
		metric.WithDescription("The total number of runtimes reported to placement service."))
	if err != nil {
		return fmt.Errorf("failed to create runtimes_total metric: %w", err)
	}

	actorRuntimesTotal, err = meter.Int64Gauge("placement.actor_runtimes_total",
		metric.WithDescription("The total number of actor runtimes reported to placement service."))
	if err != nil {
		return fmt.Errorf("failed to create actor_runtimes_total metric: %w", err)
	}

	actorHeartbeatTotal, err = meter.Int64Gauge("placement.actor_heartbeat_timestamp",
		metric.WithDescription("The actor's heartbeat timestamp (in seconds) was last reported to the placement service."))
	if err != nil {
		return fmt.Errorf("failed to create actor_heartbeat_timestamp metric: %w", err)
	}

	return nil
}

// RecordRuntimesCount records the number of connected runtimes.
func RecordRuntimesCount(count int, ns string) {
	//runtimesTotal.Record(context.Background(), int64(count), metric.WithAttributes(namespaceKey.String(ns)))
	runtimesTotal.Record(context.Background(), int64(count), metric.WithAttributes(namespaceKey.String(ns)))
}

// RecordActorRuntimesCount records the number of actor-hosting runtimes.
func RecordActorRuntimesCount(count int, ns string) {
	actorRuntimesTotal.Record(context.Background(), int64(count), metric.WithAttributes(namespaceKey.String(ns)))
}

// RecordActorHeartbeat records the actor heartbeat, in seconds since epoch, with actor type, host and pod name.
func RecordActorHeartbeat(appID, actorType, host, ns, pod string, heartbeatTime time.Time) {
	actorHeartbeatTotal.Record(context.Background(), heartbeatTime.Unix(), metric.WithAttributes(namespaceKey.String(ns), appIDKey.String(appID), actorTypeKey.String(actorType), hostNameKey.String(host), podNameKey.String(pod)))
}
