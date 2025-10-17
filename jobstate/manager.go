package jobstate

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

type JobStates map[string]*time.Time

func ReadLastCreated(name string, dir string) (*time.Time, error) {
	var states JobStates
	dir = strings.TrimSuffix(dir, "/")
	data, err := os.ReadFile(dir + "/last_created.json")
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
func WriteLastCreated(name string, dir string, t *time.Time) error {
	var states JobStates

	dir = strings.TrimSuffix(dir, "/")
	data, err := os.ReadFile(dir + "/last_created.json")
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
	return os.WriteFile(dir+"/last_created.json", newData, 0644)
}
