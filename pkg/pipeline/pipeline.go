package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	zenroom "github.com/DECODEproject/zenroom-go"
	kitlog "github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	datastore "github.com/thingful/twirp-datastore-go"

	"github.com/DECODEproject/iotencoder/pkg/lua"
	"github.com/DECODEproject/iotencoder/pkg/postgres"
	"github.com/DECODEproject/iotencoder/pkg/smartcitizen"
)

var (
	// datastoreErrorCounter is a prometheus counter recording a count of any
	// errors that occur when writing to the datastore
	datastoreErrorCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "decode",
			Subsystem: "encoder",
			Name:      "datastore_errors",
			Help:      "Count of errors writing to datastore",
		},
	)

	// zenroomErrorCounter is a prometheus counter recording a count of any errors
	// that occur when invoking zenroom.
	zenroomErrorCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "decode",
			Subsystem: "encoder",
			Name:      "zenroom_errors",
			Help:      "Count of errors invoking zenroom",
		},
	)

	// datastoreWriteHistogram is a prometheus histogram recording successful
	// writes to the datastore. We use the default bucket distributions.
	datastoreWriteHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "decode",
			Subsystem: "encoder",
			Name:      "datastore_writes",
			Help:      "Datastore writes duration distribution",
		},
	)

	// zenroomHistogram is a prometheus histogram recording execution times of
	// calls to zenroom to exec some script.
	zenroomHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "decode",
			Subsystem: "encoder",
			Name:      "zenroom_exec",
			Help:      "Execution time of zenroom scripts",
		},
	)
)

func init() {
	prometheus.MustRegister(datastoreErrorCounter)
	prometheus.MustRegister(datastoreWriteHistogram)
	prometheus.MustRegister(zenroomErrorCounter)
	prometheus.MustRegister(zenroomHistogram)
}

// Processor is a type that encapsulates processing incoming events received
// from smartcitizen, and is responsible for enriching the data, applying any
// transformations to the data and then encrypting it using zenroom before
// writing it to the datastore.
type Processor struct {
	datastore datastore.Datastore
	logger    kitlog.Logger
	verbose   bool
	sensors   *smartcitizen.Smartcitizen
}

// NewProcessor is a constructor function that takes as input an instantiated
// datastore client, and a logger. It returns the instantiated processor which
// is ready for use. Note we pass in the datastore instance so that we can
// supply a mock for testing.
func NewProcessor(ds datastore.Datastore, verbose bool, logger kitlog.Logger) *Processor {
	logger = kitlog.With(logger, "module", "pipeline")

	logger.Log("msg", "creating processor")

	return &Processor{
		datastore: ds,
		logger:    logger,
		verbose:   verbose,
		sensors:   &smartcitizen.Smartcitizen{},
	}
}

// Process is the function that actually does the work of dispatching the
// received data to all destination streams after applying whatever processing
// the stream specifies. Currently we do the simplest thing of just writing the
// data directly to the datastore.
func (p *Processor) Process(device *postgres.Device, payload []byte) error {
	// check payload
	if payload == nil {
		return errors.New("empty payload received")
	}

	parsedDevice, err := p.sensors.ParseData(device, payload)
	if err != nil {
		return errors.Wrap(err, "failed to parse SmartCitizen data")
	}

	// pull encryption script from go-bindata asset
	script, err := lua.Asset("encrypt.lua")
	if err != nil {
		return errors.Wrap(err, "failed to read zenroom script")
	}

	// iterate over the configured streams for the device
	for _, stream := range device.Streams {
		if p.verbose {
			p.logger.Log("public_key", stream.PublicKey, "device_token", device.DeviceToken, "msg", "writing data")
		}

		keyString := fmt.Sprintf(
			`{"device_token":"%s","community_id":"%s","community_pubkey":"%s"}`,
			device.DeviceToken,
			stream.PolicyID,
			stream.PublicKey,
		)

		payloadBytes, err := json.Marshal(parsedDevice)
		if err != nil {
			return errors.Wrap(err, "failed to marshal parsed device")
		}

		start := time.Now()

		encodedPayload, err := zenroom.Exec(
			script,
			zenroom.WithKeys([]byte(keyString)),
			zenroom.WithData(payloadBytes),
			zenroom.WithVerbosity(1),
		)

		duration := time.Since(start)

		if err != nil {
			zenroomErrorCounter.Inc()
			return err
		}

		zenroomHistogram.Observe(duration.Seconds())

		fmt.Println(string(encodedPayload))

		start = time.Now()

		_, err = p.datastore.WriteData(context.Background(), &datastore.WriteRequest{
			PolicyId:    stream.PolicyID,
			DeviceToken: device.DeviceToken,
			Data:        []byte(encodedPayload),
		})

		duration = time.Since(start)

		if err != nil {
			datastoreErrorCounter.Inc()
			return err
		}

		datastoreWriteHistogram.Observe(duration.Seconds())
	}

	return nil
}
