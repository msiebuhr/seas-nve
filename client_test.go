package seasnve

import (
	"os"
	"testing"
	"time"
)

func TestFailedLogin(t *testing.T) {
	c := NewClient()

	err := c.Login("bad-username", "wrong-password")

	if err == nil {
		t.Fatalf("Expected error, got nothing")
	}

	if err.Error() != "Not authorized. Wrong username or password" {
		t.Fatalf("Expected error `Not authorized. Wrong username or password`, got `%s`", err.Error())
	}
}

func TestFull(t *testing.T) {
	email := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")

	c := NewClient()

	err := c.Login(email, password)
	if err != nil {
		t.Fatalf("Unexpected error logging in: %s", err)
	}

	// Sub-tests for various parts of the setup
	t.Run("Metering", func(t *testing.T) {
		m, err := c.Metering()
		if err != nil {
			t.Fatalf("Unexpected error getting basic metering: %s", err)
		}

		// Let's dump output if we're verbose. Nice for exploration...
		t.Logf("Output %+v\n", m)

		if len(m.MeteringPoints) == 0 {
			t.Fatalf("Not no metering points, expected at least one")
		}

		for _, mp := range m.MeteringPoints {
			if mp.ConsumptionYearToDate == 0 {
				t.Errorf("Unexpected low ConsumptionYearToDate: ", mp.ConsumptionYearToDate)
			}

			if mp.MeteringPoint == "" {
				t.Fatalf("Unexpected empty MeteringPoint")
			}

			// Now we can get detailed data from each meter...
			t.Run("MeteringPoints", func(t *testing.T) {
				p, err := c.MeteringPoints(
					mp.MeteringPoint,
					time.Now(),
					time.Now().Add(-4*24*time.Hour),
					AGGREGATION_DAY,
				)

				if err != nil {
					t.Fatalf("Unexpected error getting points: %s", err)
				}

				// Let's dump output if we're verbose. Nice for exploration...
				t.Logf("Output %+v\n", p)

				// It has a metering-point similar to the one we asked about
				if len(p.MeteringPoints) != 1 {
					t.Fatalf("Expected one MeteringPoint, got %d", len(p.MeteringPoints))
				}

				if p.MeteringPoints[0].MeteringPoint != mp.MeteringPoint {
					t.Errorf(
						"Expected MeteringPoint %s, got %s",
						mp.MeteringPoint,
						p.MeteringPoints[0].MeteringPoint,
					)
				}

				// Has values for the days we've asked for
				if len(p.MeteringPoints[0].Values) != 5 {
					t.Errorf("Expected five datapoints, got %d", len(p.MeteringPoints[0].Values))
				}
			})
		}
	})

	t.Run("Management", func(t *testing.T) {
		m, err := c.Management()
		if err != nil {
			t.Fatalf("Unexpected error getting management data: %s", err)
		}

		// Let's dump output if we're verbose. Nice for exploration...
		t.Logf("Output %+v\n", m)

		// TODO(msiebuhr): Wat to test?
	})
}
