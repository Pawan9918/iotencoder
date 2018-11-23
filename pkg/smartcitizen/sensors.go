package smartcitizen

import (
	"encoding/json"

	"github.com/pkg/errors"
	null "gopkg.in/guregu/null.v3"
)

// Sensor is a type we use to parse the sensor type json published by SmartCitizen.
type Sensor struct {
	ID          int         `json:"id"`
	UUID        string      `json:"uuid"`
	ParentID    null.Int    `json:"parent_id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Unit        null.String `json:"unit"`
}

// Sensors is a function that returns a list of Sensor instances read from the
// static copy of SmartCitizen's sensor list we maintain locally. This map can
// then be used by the pipeline in order to create richer data.
func Sensors() (map[int]*Sensor, error) {
	sensorBytes, err := Asset("sensors.json")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read sensors.json")
	}

	var sensorList []Sensor

	err = json.Unmarshal(sensorBytes, &sensorList)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal data")
	}

	sensors := map[int]*Sensor{}

	for _, sensor := range sensorList {
		sensors[sensor.ID] = &sensor
	}

	return sensors, nil
}
