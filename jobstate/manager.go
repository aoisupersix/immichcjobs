package jobstate

import (
	"encoding/json"
	"os"
	"time"
)

type JobStates map[string]*time.Time

// ReadLastCreated reads the last created timestamp from last_created.json file.
func ReadLastCreated(name string) (*time.Time, error) {
	var states JobStates
	data, err := os.ReadFile("last_created.json")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &states); err != nil {
		return nil, err
	}
	return states[name], nil
}

// WriteLastCreated writes the last created timestamp to last_created.json file.
func WriteLastCreated(name string, t *time.Time) error {
	var states JobStates

	data, err := os.ReadFile("last_created.json")
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &states); err != nil {
			return err
		}
	} else {
		states = make(JobStates)
	}

	states[name] = t

	newData, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("last_created.json", newData, 0644)
}
