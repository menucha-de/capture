package capture

import (
	"errors"

	guuid "github.com/google/uuid"
	transport "github.com/peramic/transport"
)

func GetCycles() map[string]CycleSpec {
	return config.Cycles
}
func HasEnabledReports() bool {
	config.mu.Lock()
	defer config.mu.Unlock()
	for _, report := range config.Cycles {
		if report.Enabled {
			return true
		}
	}
	return false
}
func GetCycle(id string) (CycleSpec, error) {
	cycle, ok := config.Cycles[id]
	if !ok {
		return CycleSpec{}, errors.New("Cycle with ID " + id + " does not exist")
	}
	return cycle, nil
}
func GetDevices() map[string]Device {
	config.mu.RLock()
	defer config.mu.RUnlock()
	return config.Devices
}
func GetDevice(id string) (Device, error) {

	return config.GetDevice(id)
}
func GetSubscriptors(id string) (map[string]transport.Subscriptor, error) {
	config.mu.RLock()
	defer config.mu.RUnlock()
	_, ok := config.Cycles[id]
	if !ok {
		return nil, errors.New("Report with ID " + id + " does not exist")

	}
	s, _ := config.Subscriptors[id]
	return s, nil

}
func GetSubscriptorsIds(id string) ([]string, error) {
	config.mu.RLock()
	defer config.mu.RUnlock()
	_, ok := config.Cycles[id]
	if !ok {
		return nil, errors.New("Report with ID " + id + " does not exist")

	}
	s, _ := config.Subscriptors[id]
	keys := make([]string, len(s))

	i := 0
	for k := range s {
		keys[i] = k
		i++
	}
	return keys, nil
}
func GetSubscriptor(id string, subID string) (transport.Subscriptor, error) {
	config.mu.RLock()
	defer config.mu.RUnlock()
	_, ok := config.Cycles[id]

	if !ok {
		return transport.Subscriptor{}, errors.New("Report with ID " + id + " does not exist")

	}
	s, ok := config.Subscriptors[id]
	if !ok {
		return transport.Subscriptor{}, errors.New("Report with ID " + id + " doesn't have subscriptors")

	}
	sub, ok := s[subID]
	if !ok {
		return transport.Subscriptor{}, errors.New("Subscriptor with ID " + subID + " doesn't exist")

	}
	return sub, nil
}
func Add(spec CycleSpec) error { //only at start

	c, err := newCommonCycle(spec)
	// start cycle if cycle is enabled
	if err != nil {
		lg.Error(err)
		return err
	}
	setCommonCycle(spec.ID, c)

	if spec.Enabled {
		c.evaluateCycleState()
	}
	return nil
}
func AddSubscriptor(id string, subscriptor transport.Subscriptor) error {
	cycle, err := getCommonCycle(id)
	if err != nil {
		return nil
	}

	err = transport.AddSubscriptor(subscriptor)
	if err != nil {
		lg.Error(err)
		return err
	}

	cycle.evaluateCycleState()
	return nil
}
func Define(spec CycleSpec) (string, error) {

	// create id
	if spec.ID != "" {
		return "", errors.New("Report ID must noy be set")
	}

	var id guuid.UUID
	for {
		id = guuid.New()
		_, ok := config.Cycles[id.String()]
		if !ok {
			spec.ID = id.String()
			break
		}

	}
	c, err := newCommonCycle(spec)
	// start cycle if cycle is enabled
	if err != nil {
		return "", err
	}
	err = config.addCycle(spec)
	if err != nil {
		return "", err
	}
	setCommonCycle(id.String(), c)

	// return generated id
	if spec.Enabled {
		enableFields(spec)
		c.evaluateCycleState()
	}
	return spec.ID, nil
}
func Update(spec CycleSpec) error {

	// update cycle
	cycle, err := getCommonCycle(spec.ID)
	if err != nil {
		return err
	}
	current := config.Cycles[spec.ID]
	if spec.Name == "" {
		return errors.New("Cycle Name must  be set ")
	}
	for _, v := range config.Cycles {
		if v.Name == spec.Name && v.ID != spec.ID {
			return errors.New("Cycle with name " + spec.Name + " allready exists!")
		}
	}

	if spec.Enabled && len(spec.FieldSubscriptions) == 0 {
		return errors.New("Cycle must have at least one field")
	}

	if current.Enabled {
		if spec.Enabled {
			return errors.New("Cannot update active cycle")
		}
		//disable cycle
		disableFields(current)
		cycle.update(spec)
		cycle.evaluateCycleState()

	} else {
		if spec.Enabled {
			//enable cycle
			err := cycle.update(spec)
			if err != nil {
				return err
			}
			enableFields(spec)
			cycle.evaluateCycleState()
		} else {
			err := cycle.update(spec)
			if err != nil {
				return err
			}
		}
	}
	config.setCycle(spec)
	return nil
}
func Undefine(id string) error {

	// update cycle
	cycle, err := getCommonCycle(id)
	if err != nil {
		return err
	}
	current := config.Cycles[id]
	if current.Enabled {
		disableFields(current)
		cycle.evaluateCycleState()
	}
	for _, v := range config.Subscriptors[id] {
		UndefineSubscriptor(id, v.ID)
	}
	config.deleteCycle(id)

	deleteCommonCycle(id)
	return nil
}
func DefineSubscriptor(id string, subscriptor transport.Subscriptor) (string, error) {
	cycle, err := getCommonCycle(id)
	if err != nil {
		return "", err
	}

	subID, err := transport.DefineSubscriptor(subscriptor)
	if err != nil {
		return "", err
	}
	subscriptor.ID = subID
	config.addSubscriptor(id, subscriptor)
	//could be error?
	cycle.evaluateCycleState()
	return subID, nil
}
func UndefineSubscriptor(id string, subID string) error {
	cycle, err := getCommonCycle(id)
	if err != nil {
		return err
	}
	err = transport.DeleteSubscriptor(subID)
	if err != nil {
		return err
	}
	err = config.removeSubscriptor(id, subID)
	if err != nil {
		return err
	}
	cycle.evaluateCycleState()
	return nil
}
func UpdateSubscriptor(id string, subID string, spec transport.Subscriptor) error {

	// update subscriptor
	old, err := GetSubscriptor(id, subID)

	if err != nil {
		return err
	}
	if old.ID != subID {
		return errors.New("Subscriptor ID does not match")
	}
	cycle, err := getCommonCycle(id)
	if err != nil {
		return err
	}
	err = transport.UpdateSubscriptor(spec)
	if err != nil {
		return err
	}
	config.setSubscriptor(id, spec)
	if err != nil {
		return err
	}
	cycle.evaluateCycleState()
	return nil
}
