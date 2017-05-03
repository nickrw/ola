package ola

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Universe represents a DMX universe within the OLA daemon.
type Universe struct {
	Number int
	OLAD   string
	ports  [512]uint8
}

// Render will write the current Universe channel states to the OLA daemon
// via the OLA HTTP interface (set_dmx). The entire universe is written at once.
func (universe *Universe) Render() {
	csvPorts := strings.Join(sliceItoa(universe.ports[:]), ",")
	postDataValues := url.Values{}
	postDataValues.Add("u", strconv.Itoa(universe.Number))
	postDataValues.Add("d", csvPorts)
	_, err := http.PostForm(universe.OLAD+"/set_dmx", postDataValues)
	if err != nil {
		log.Printf("Request failed: %s", err.Error())
	} else {
		log.Printf("Sent set_dmx request to OLA daemon: %s", csvPorts)
	}
}

// SetChannel updates the absolute DMX value of the given channel number
// Note that channel numbers begin at one, not zero.
// This function does not render the change to the DMX universe.
func (universe *Universe) SetChannel(ch int, val uint8) error {
	err := validateChannelNumber(ch)
	if err != nil {
		return err
	}
	universe.ports[ch-1] = val
	return nil
}

// SetChannelPercent updates the DMX value of the given channel number by percentage
// Percentages will be converted to an absolute DMX value (0 to 255)
// This function does not render the change to the DMX universe
func (universe *Universe) SetChannelPercent(ch int, pct float64) error {
	if pct < 0.0 || pct > 1.0 {
		return errors.New("Percentage value must be between 0.0 and 1.0")
	}
	scaledBrightness := uint8(pct * 255.0)
	return universe.SetChannel(ch, scaledBrightness)
}

// GetChannel returns the absolute DMX value of the given channel number
func (universe *Universe) GetChannel(ch int) (uint8, error) {
	err := validateChannelNumber(ch)
	if err != nil {
		return 0, err
	}
	return universe.ports[ch-1], err
}

// GetChannelPercent returns a float percentage representation of the given channel number
func (universe *Universe) GetChannelPercent(ch int) (float64, error) {
	val, err := universe.GetChannel(ch)
	if err != nil {
		return 0.0, err
	}
	return float64(val) / 255.0, nil
}

func validateChannelNumber(ch int) error {
	if ch < 1 {
		return fmt.Errorf("Channel number %d is too low (1-512)", ch)
	} else if ch > 512 {
		return fmt.Errorf("Channel number %d is too high (1-512)", ch)
	}
	return nil
}

// aliceItoa returns a slice of the Universe's DMX Ports with the
// absolute DMX values converted to strings
func sliceItoa(channels []uint8) []string {
	strPorts := make([]string, len(channels))
	for i, v := range channels {
		strPorts[i] = strconv.Itoa(int(v))
	}
	return strPorts[:]
}
