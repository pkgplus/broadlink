package broadlink

import (
	"testing"
	"time"
)

func TestDiscover(t *testing.T) {
	Discover(5 * time.Second)
}
