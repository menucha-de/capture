package capture

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/menucha-de/transport"
	"github.com/menucha-de/logging"
)

const filename = "./conf/capture/capture.json"
const dirname = "./conf/capture/"

var lg *logging.Logger = logging.GetLogger("capture")
var config *CycleConfiguration

type fieldSubscriptions struct {
	deviceSubscription map[string]map[string]int
	mu                 sync.Mutex
}

var devSubscriptions fieldSubscriptions

type mapCycles struct {
	cycles map[string]*CommonCycle
	mu     sync.Mutex
}

var cycles mapCycles

func init() {
	cycles = mapCycles{cycles: make(map[string]*CommonCycle)}
	ss := make(map[string]map[string]int)
	devSubscriptions = fieldSubscriptions{deviceSubscription: ss}
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		os.MkdirAll(dirname, 0700)
	}
	f, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err == nil {
		dec := json.NewDecoder(f)
		err = dec.Decode(&config)

		if err != nil {
			lg.WithError(err).Warning("Failed to parse config")
			config = initConfiguration()
		}
		if len(config.Cycles) > 0 {
			for _, v := range config.Cycles {

				if v.Enabled {
					enableFields(v)
				}
				Add(v)
				if config.Subscriptors != nil {
					s, ok := config.Subscriptors[v.ID]
					if ok {
						for _, sub := range s {
							AddSubscriptor(v.ID, sub)
						}
					}
				}
			}
		}

	} else {
		lg.WithError(err).Debug("Failed to read config")
		config = initConfiguration()
	}
	defer f.Close()

}

func initConfiguration() *CycleConfiguration {
	cycles := make(map[string]CycleSpec)
	devices := make(map[string]Device)
	subscriptors := make(map[string]map[string]transport.Subscriptor)
	config := CycleConfiguration{Cycles: cycles, Devices: devices, Subscriptors: subscriptors}

	return &config
}
func enableFields(report CycleSpec) {
	for id, fields := range report.FieldSubscriptions {
		dev, ok := config.Devices[id]
		if ok {

			for _, p := range fields {
				f, fok := dev.Fields[p]
				if fok {
					if f.Properties["direction"] == "INPUT" {
						f.Properties["enabled"] = "true"
						config.SetDeviceField(id, f)

						devSubscriptions.addField(id, f.ID)
					}
				}
			}
		}
	}
}
func disableFields(report CycleSpec) {
	for id, fields := range report.FieldSubscriptions {
		dev, ok := config.Devices[id]
		if ok {
			for _, p := range fields {
				f, fok := dev.Fields[p]
				if fok {
					devSubscriptions.deleteField(id, f.ID)
					//config.IoConfigurations[id-1].Enable = true

				}
				_, ok = devSubscriptions.deviceSubscription[id][f.ID]
				if !ok || devSubscriptions.deviceSubscription[id][f.ID] == 0 {
					f.Properties["enabled"] = "false"
					config.SetDeviceField(id, f)
				}
			}

		}

	}
}
func (s *fieldSubscriptions) addField(device string, field string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.deviceSubscription[device]
	if ok {
		_, okk := m[field]
		if okk {
			m[field]++
			s.deviceSubscription[device] = m
		} else {

			m[field] = 1
			s.deviceSubscription[device] = m
		}
	} else {
		fields := make(map[string]int)
		fields[field] = 1
		s.deviceSubscription[device] = fields
	}

}
func (s *fieldSubscriptions) deleteField(device string, field string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.deviceSubscription[device]
	if ok {
		_, okk := m[field]
		if okk {
			m[field]--
			if m[field] == 0 {
				delete(s.deviceSubscription[device], field)
			} else {
				s.deviceSubscription[device] = m
			}
		}
	}

}
func getCommonCycle(id string) (*CommonCycle, error) {
	cycles.mu.Lock()
	defer cycles.mu.Unlock()
	cycle, ok := cycles.cycles[id]
	if !ok {
		return nil, errors.New("Cycle spec with ID '" + id + "' does not exist")
	}
	return cycle, nil
}
func setCommonCycle(id string, commonCycle *CommonCycle) {
	cycles.mu.Lock()
	defer cycles.mu.Unlock()
	cycles.cycles[id] = commonCycle

}
func deleteCommonCycle(id string) {
	cycles.mu.Lock()
	defer cycles.mu.Unlock()
	delete(cycles.cycles, id)

}
