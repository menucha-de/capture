package capture

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	transport "github.com/peramic/transport"
)

type CycleConfiguration struct {
	Devices      map[string]Device                           `json:"devices"`
	Cycles       map[string]CycleSpec                        `json:"cycles"`
	Subscriptors map[string]map[string]transport.Subscriptor `json:"subscriptors"`
	mu           sync.RWMutex
}

func (c *CycleConfiguration) serialize() {
	f, err := os.Create(filename) ///, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		lg.WithError(err).Error("Failed to create or open configuration file")
	} else {
		enc := json.NewEncoder(f)
		enc.SetIndent("", " ")
		enc.Encode(c)
	}
	defer f.Close()
}
func (c *CycleConfiguration) AddDevice(dev Device) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if dev.ID != "" {
		return "", errors.New("Device ID must not be set ")
	}
	if dev.Name == "" {
		return "", errors.New("Device Name must  be set ")
	}
	for _, v := range c.Devices {
		if v.Name == dev.Name {
			return "", errors.New("Device with name " + dev.Name + " allready exists!")
		}
	}
	/*var id guuid.UUID
	for {
		id = guuid.New()
		_, ok := c.Devices[id.String()]
		if !ok {
			dev.ID = id.String()
			break
		}
	}*/
	id := "gpio"
	dev.ID = id
	c.Devices[id] = dev
	c.serialize()
	return id, nil
}
func (c *CycleConfiguration) GetDevice(id string) (Device, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	dev, ok := c.Devices[id]
	if !ok {
		return Device{}, errors.New("Device with ID " + id + " does not exist")
	}
	return dev, nil
}
func (c *CycleConfiguration) getDeviceFieldLabel(id string, idd string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	dev, ok := c.Devices[id]
	if !ok {
		return ""
	}
	f, ok := dev.Fields[idd]
	if !ok {
		return ""
	}
	if f.Label != "" {
		return f.Label
	}
	return f.Name

}
func (c *CycleConfiguration) SetDevice(dev Device) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	old, ok := c.Devices[dev.ID]

	if !ok {
		return errors.New("Device with ID " + dev.ID + " does not exist")
	}

	if old.Enabled && dev.Enabled {
		return errors.New("Device with ID " + dev.ID + " is in use")
	}
	if dev.Name == "" {
		return errors.New("Device Name must  be set ")
	}
	for _, v := range c.Devices {
		if v.Name == dev.Name && v.ID != dev.ID {
			return errors.New("Device with name " + dev.Name + " allready exists!")
		}
	}

	c.Devices[dev.ID] = dev
	c.serialize()
	return nil
}
func (c *CycleConfiguration) SetDeviceField(devId string, field Field) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	dev, ok := c.Devices[devId]

	if !ok {
		return errors.New("Device with ID " + dev.ID + " does not exist")
	}

	_, fok := dev.Fields[field.ID]
	if !fok {
		return errors.New("Field " + field.Name + " does not exist")
	}

	c.Devices[devId].Fields[field.ID] = field
	c.serialize()
	return nil
}
func (c *CycleConfiguration) deleteDevice(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.Devices[id]

	if !ok {
		return errors.New("Device with ID " + id + " does not exist")

	}
	for _, v := range c.Cycles {
		for d := range v.FieldSubscriptions {
			if d == id {
				return errors.New("device is in use by " + v.Name)
			}
		}
	}
	delete(c.Devices, id)
	c.serialize()
	return nil
}
func (c *CycleConfiguration) addCycle(cycle CycleSpec) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cycle.Name == "" {
		return errors.New("Cycle Name must  be set ")
	}
	for _, v := range c.Cycles {
		if v.Name == cycle.Name {
			return errors.New("Cycle with name " + cycle.Name + " allready exists!")
		}
	}

	if cycle.Enabled && len(cycle.FieldSubscriptions) == 0 {
		return errors.New("Cycle must have at least one field")
	}
	c.Cycles[cycle.ID] = cycle
	c.serialize()
	return nil
}

func (c *CycleConfiguration) setCycle(cycle CycleSpec) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Cycles[cycle.ID] = cycle
	c.serialize()

}
func (c *CycleConfiguration) deleteCycle(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.Cycles, id)
	delete(c.Subscriptors, id)
	c.serialize()
	return
}

func (c *CycleConfiguration) addSubscriptor(id string, sub transport.Subscriptor) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.Cycles[id]

	if !ok {
		return errors.New("Report with ID " + id + " does not exist")

	}

	if c.Subscriptors[id] == nil {
		c.Subscriptors[id] = make(map[string]transport.Subscriptor)
	}
	c.Subscriptors[id][sub.ID] = sub
	c.serialize()
	return nil
}
func (c *CycleConfiguration) removeSubscriptor(id string, subID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.Cycles[id]

	if !ok {
		return errors.New("Report with ID " + id + " does not exist")

	}

	delete(c.Subscriptors[id], subID)
	c.serialize()
	return nil
}
func (c *CycleConfiguration) setSubscriptor(id string, sub transport.Subscriptor) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.Cycles[id]
	if !ok {
		return errors.New("Report with ID " + id + " does not exist")

	}
	if c.Subscriptors[id] == nil {
		c.Subscriptors[id] = make(map[string]transport.Subscriptor)
	}
	c.Subscriptors[id][sub.ID] = sub
	c.serialize()
	return nil
}
func GetConfig() *CycleConfiguration {
	return config
}
