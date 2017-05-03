package ola

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const zero uint8 = 0
const full uint8 = 255

func TestUniverseTestSuite(t *testing.T) {
	suite.Run(t, new(UniverseTestSuite))
}

type UniverseTestSuite struct {
	suite.Suite
	uv     Universe
	assert *assert.Assertions
}

func (s *UniverseTestSuite) SetupTest() {
	s.uv = Universe{Number: 0, OLAD: "http://127.0.0.1:9090"}
	s.assert = s.Assert()
}

func (s *UniverseTestSuite) TestSetChannelPercentBadValues() {
	s.assert.Equal(s.uv.ports[0], zero)

	// Any value below 0% should be rejected
	err1 := s.uv.SetChannelPercent(1, -0.1)
	s.assert.NotNil(err1, "negative value should return an error")

	// Any value above 100% (> 1.0) should be rejected
	err2 := s.uv.SetChannelPercent(1, 1.1)
	s.assert.NotNil(err2, "too-large value should return an error")
}

func (s *UniverseTestSuite) TestSetChannelPercent() {
	s.assert.Equal(s.uv.ports[0], zero)

	err := s.uv.SetChannelPercent(1, 1.0)
	s.assert.Nil(err, "value of 1.0 shouldn't raise an error")
	s.assert.Equal(full, s.uv.ports[0], "Channel value should be 255")
}

func (s *UniverseTestSuite) TestGetChannel() {
	val, err := s.uv.GetChannel(1)
	s.assert.Nil(err, "Channel 1 shouldn't raise an error")
	s.assert.Equal(val, zero)

	err = s.uv.SetChannel(1, full)
	s.assert.Nil(err, "SetChannel on 1 shouldn't raise an error")

	val, _ = s.uv.GetChannel(1)
	s.assert.Equal(val, full)
}

func (s *UniverseTestSuite) TestGetChannelPercent() {
	s.uv.SetChannel(1, 255)
	val, err := s.uv.GetChannelPercent(1)
	s.assert.NoError(err)
	s.assert.Equal(val, 1.0, "foo")
}

func (s *UniverseTestSuite) TestGetChannelInvalid() {
	_, err := s.uv.GetChannel(0)
	s.assert.Error(err)
}

func (s *UniverseTestSuite) TestGetChannelPercentInvalid() {
	_, err := s.uv.GetChannelPercent(0)
	s.assert.Error(err)
}

func (s *UniverseTestSuite) TestSetChannelInvalid() {
	err := s.uv.SetChannel(0, 255)
	s.assert.Error(err)
}

func (s *UniverseTestSuite) TestChannelValidity() {
	err := validateChannelNumber(0)
	s.assert.EqualError(err, "Channel number 0 is too low (1-512)")

	err = validateChannelNumber(513)
	s.assert.EqualError(err, "Channel number 513 is too high (1-512)")

	for i := range s.uv.ports {
		err = validateChannelNumber(i + 1)
		s.assert.Nil(err, "Channel %d should be null", i+1)
	}
}

func renderAssertions(s *UniverseTestSuite, uv int, dmx []uint8) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		s.assert.Equal(strconv.Itoa(uv), r.FormValue("u"))
		strChannels := strings.Join(sliceItoa(dmx), ",")
		s.assert.Equal(strChannels, r.FormValue("d"))
	})
}

func (s *UniverseTestSuite) TestRenderZeros() {
	ts := httptest.NewServer(renderAssertions(s, 1, make([]uint8, 512)))
	s.uv.OLAD = ts.URL
	s.uv.Number = 1
	s.uv.Render()
}

func (s *UniverseTestSuite) TestRenderFull() {
	expectValues := make([]uint8, 512)
	for i := range expectValues {
		expectValues[i] = uint8(i % 256)
		s.uv.SetChannel(i+1, uint8(i%256))
	}
	ts := httptest.NewServer(renderAssertions(s, 0, expectValues))
	s.uv.OLAD = ts.URL
	s.uv.Render()
}

func TestSliceItoa(t *testing.T) {
	channels := make([]uint8, 512)
	strChannels := make([]string, 512)
	for i := range channels {
		channels[i] = uint8(i)
		strChannels[i] = strconv.Itoa(int(uint8(i)))
	}
	assert.Equal(t, strChannels, sliceItoa(channels))
}
