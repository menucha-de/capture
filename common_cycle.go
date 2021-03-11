package capture

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

var ch chan State

type CommonCycle struct {
	mu               sync.Mutex
	state            State
	report           CycleSpec
	initiation       Initiation
	termination      Termination
	terminator       string
	initiator        string
	hasStartTriggers bool
	stopchan         chan struct{}
	stoppedchan      chan struct{}
	hasStopTriggers  bool
}

func newCommonCycle(report CycleSpec) (*CommonCycle, error) {
	p := &CommonCycle{hasStartTriggers: false, hasStopTriggers: false}
	p.report = report
	p.state = UNREQUESTED
	err := p.validateConfiguration()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (c *CommonCycle) validateConfiguration() error {
	if c.report.Duration <= 0 && c.report.Interval <= 0 && !c.hasStopTriggers && !c.report.WhenDataAvailable {
		return errors.New("No stop condition specifed")
	}

	return nil
}
func (c *CommonCycle) stopSpec() {

	close(c.stopchan) // tell it to stop
	<-c.stoppedchan   // wait for it to have stopped
	fmt.Println("Stopped.")
	//only after to avoid race condition
	if ACTIVE == c.state {
		c.termination = UNREQUESTEDTERMINATION
	}
	c.state = UNREQUESTED
	lg.Trace("Cycle "+c.report.Name+" state changed to ", c.state)

}

func (c *CommonCycle) update(report CycleSpec) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.report = report
	err := c.validateConfiguration()
	return err
}

func (c *CommonCycle) evaluateCycleState() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if UNDEFINED != c.state {
		if c.hasEnabledSubscribers() && UNREQUESTED == c.state {
			c.state = STATEREQUESTED
			lg.Trace("Cycle "+c.report.Name+" state changed to ", c.state)

			if !c.hasStartTriggers {
				c.initiation = REQUESTED
				if c.state == STATEREQUESTED {
					c.state = ACTIVE
					var d time.Duration
					d = time.Duration(c.report.RepeatPeriod) * time.Millisecond
					if c.report.RepeatPeriod < c.report.Duration {
						d = time.Duration(c.report.Duration) * time.Millisecond
					}
					c.stopchan = make(chan struct{})
					c.stoppedchan = make(chan struct{})
					report := c.report
					c.start(d, report)

				}
			}

		}
		//i Don't think we need this anymore
		if !c.hasEnabledSubscribers() && UNREQUESTED != c.state {

			c.stopSpec()

		}

	}
}

func (c *CommonCycle) hasEnabledSubscribers() bool {
	subscriptorsIds, err := GetSubscriptorsIds(c.report.ID)
	if err != nil {
		return false

	}
	for _, v := range subscriptorsIds {
		sub, err := GetSubscriptor(c.report.ID, v)
		if err == nil {
			if sub.Enable && c.report.Enabled {
				return true
			}
		}

	}
	return false
}

func (c *CommonCycle) start(t time.Duration, report CycleSpec) {

	go func() { // work in background
		// close the stoppedchan when this func
		// exits

		defer close(c.stoppedchan)
		// TODO: do setup work

		var ticker *time.Ticker
		if t > 0 && !report.WhenDataAvailable {
			ticker = time.NewTicker(t)
		}
		startTime := time.Now().UnixNano() / 1000000
		defer func() {
			// TODO: do teardown work
			if ticker != nil {
				ticker.Stop()
			}

		}()
		channels := make(map[string]chan CaptureData)
		for _, v := range report.FieldSubscriptions {
			for _, val := range v {

				ch := Pub.subscribe(val)
				channels[val] = ch
			}

		}
		defer func() {
			for i, v := range channels {
				Pub.unSubscribe(i, v)
			}
		}()
		c.createReport(channels, startTime, report, t)
		if ticker == nil {
			return
		}
		for {
			select {
			default:
			case <-ticker.C:
				startTime := time.Now().UnixNano() / 1000000
				c.initiation = REPEAT_PERIOD
				c.termination = DURATION
				c.createReport(channels, startTime, report, t)

			case <-c.stopchan:
				return
			}
		}
	}()
}
func (c *CommonCycle) createReport(channels map[string]chan CaptureData, startTime int64, report CycleSpec, repeat time.Duration) {
	d := time.Now().Add(time.Duration(report.Duration) * time.Millisecond)
	var ctx context.Context
	var cancel context.CancelFunc
	if !report.WhenDataAvailable {
		ctx, cancel = context.WithDeadline(context.Background(), d)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	host, _ := os.Hostname()
	areport := AdapterReport{
		Termination:   c.termination,
		Terminator:    c.terminator,
		Initiation:    c.initiation,
		Initiator:     c.initiator,
		ReportName:    report.Name,
		Date:          time.Now().UTC(),
		ApplicationID: host,
	}
	areport.Devices = make([]*DeviceReport, 0)
	dev, _ := config.GetDevice("gpio")
	ll := dev.Name
	if dev.Label != "" {
		ll = dev.Label
	}

	devrep := DeviceReport{Name: ll, Fields: make([]*FieldReport, 0)}
	areport.Devices = append(areport.Devices, &devrep)

	defer cancel()
	defer func() {
		creationTime := time.Now().UnixNano() / 1000000
		totalMilliseconds := creationTime - startTime
		areport.TotalMilliseconds = totalMilliseconds

		if len(devrep.Fields) != 0 || report.ReportIfEmpty {
			c.sendReport(areport, report.ID)
		}

	}()
	timeout := time.After(repeat)
	for {
		select {

		case <-ctx.Done():
			return
		case <-c.stopchan:
			fmt.Println("stoppppppp")
			return
		default:
			for _, ch := range channels {
				select {
				case msg := <-ch:
					lg.Debug("Received", msg)
					ff := config.getDeviceFieldLabel("gpio", msg.Field)
					f := FieldReport{Date: time.Now().UTC(), Name: ff, Value: msg.Value}
					devrep.Fields = append(devrep.Fields, &f)
					if report.WhenDataAvailable {
						areport.Termination = DATAAVAILABLE
						creationTime := time.Now().UnixNano() / 1000000
						areport.TotalMilliseconds = creationTime - startTime
						c.sendReport(areport, report.ID)
						devrep.Fields = make([]*FieldReport, 0)
						startTime = time.Now().UnixNano() / 1000000
						areport.Date = time.Now().UTC()
						//return
					}
				case <-timeout:
					if report.WhenDataAvailable {
						timeout = time.After(repeat)
						areport.Termination = DURATION
						creationTime := time.Now().UnixNano() / 1000000
						areport.TotalMilliseconds = creationTime - startTime
						c.sendReport(areport, report.ID)
						devrep.Fields = make([]*FieldReport, 0)
						startTime = time.Now().UnixNano() / 1000000
						areport.Date = time.Now().UTC()
						//return
					}
				case <-ctx.Done():
					return
				case <-c.stopchan:
					fmt.Println("stoppppppp")
					return
				default:
					time.Sleep(50 * time.Millisecond)
				}
			}
		}
	}

}

func (c *CommonCycle) sendReport(report AdapterReport, id string) {
	out, err := json.Marshal(report)
	if err != nil {
		lg.Error(err)
		return
	}
	subscriptors, err := GetSubscriptorsIds(id)
	if err != nil {
		return
	}
	for _, v := range subscriptors {
		sub, err := GetSubscriptor(id, v)

		if err == nil && sub.Enable {

			go func() {
				sub.SendReport(string(out))
			}()

		}
	}

}
