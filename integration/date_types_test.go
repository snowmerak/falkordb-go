package integration_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	createGraph()

	queries := []struct {
		query string
		check func(interface{}) error
	}{
		{
			query: "RETURN date('2023-01-01')",
			check: func(v interface{}) error {
				ts, ok := v.(time.Time)
				if !ok {
					return fmt.Errorf("expected time.Time, got %T", v)
				}
				// 2023-01-01 UTC = 1672531200
				if ts.Unix() != 1672531200 {
					return fmt.Errorf("expected unix 1672531200, got %d", ts.Unix())
				}
				return nil
			},
		},
		{
			query: "RETURN localdatetime('2023-01-01T12:00:00')",
			check: func(v interface{}) error {
				ts, ok := v.(time.Time)
				if !ok {
					return fmt.Errorf("expected time.Time, got %T", v)
				}
				// 2023-01-01 12:00:00 UTC = 1672574400
				if ts.Unix() != 1672574400 {
					return fmt.Errorf("expected unix 1672574400, got %d", ts.Unix())
				}
				return nil
			},
		},
		{
			query: "RETURN localtime('12:00:00')",
			check: func(v interface{}) error {
				ts, ok := v.(time.Time)
				if !ok {
					return fmt.Errorf("expected time.Time, got %T", v)
				}
				// 1900-01-01 12:00:00 UTC = -2208945600
				if ts.Unix() != -2208945600 {
					return fmt.Errorf("expected unix -2208945600, got %d", ts.Unix())
				}
				return nil
			},
		},
		{
			query: "RETURN duration({hours: 1})",
			check: func(v interface{}) error {
				d, ok := v.(time.Duration)
				if !ok {
					return fmt.Errorf("expected time.Duration, got %T", v)
				}
				if d != time.Hour {
					return fmt.Errorf("expected 1h, got %v", d)
				}
				return nil
			},
		},
	}

	for _, q := range queries {
		if q.query == "" {
			continue
		}
		t.Run(q.query, func(t *testing.T) {
			res, err := graphInstance.Query(q.query, nil, nil)
			assert.Nil(t, err)
			assert.False(t, res.Empty(), "Expecting resultset to have records")

			val := res.Results()[0].Values()[0]

			if q.check != nil {
				err := q.check(val)
				assert.Nil(t, err)
			}
		})
	}
}
