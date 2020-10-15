package dbs

import (
	"strconv"

	"github.com/elastic/beats/v7/libbeat/common/cfgwarn"
	"github.com/elastic/beats/v7/metricbeat/mb"
	"github.com/elastic/beats/v7/metricbeat/module/redis"
	"github.com/pkg/errors"
)

// init registers the MetricSet with the central registry as soon as the program
// starts. The New function will be called later to instantiate an instance of
// the MetricSet for each host defined in the module's configuration. After the
// MetricSet has been created then Fetch will begin to be called periodically.
func init() {
	mb.Registry.MustAddMetricSet("redis", "dbs", New)
}

// MetricSet holds any configuration or state information. It must implement
// the mb.MetricSet interface. And this is best achieved by embedding
// mb.BaseMetricSet because it implements all of the required mb.MetricSet
// interface methods except for Fetch.
type MetricSet struct {
	*redis.MetricSet
}

// New creates a new instance of the MetricSet. New is responsible for unpacking
// any MetricSet specific configuration options if there are any.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	cfgwarn.Beta("The redis dbs metricset is beta.")

	config := struct{}{}
	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	ms, err := redis.NewMetricSet(base)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create 'key' metricset")
	}

	return &MetricSet{
		MetricSet: ms,
	}, nil
}

// Fetch methods implements the data gathering and data conversion to the right
// format. It publishes the event which is then forwarded to the output. In case
// of an error set the Error field of mb.Event or simply call report.Error().
func (m *MetricSet) Fetch(report mb.ReporterV2) error {
	conn := m.Connection()
	defer func() {
		if err := conn.Close(); err != nil {
			m.Logger().Debug(errors.Wrapf(err, "failed to release connection"))
		}
	}()

	// Fetch default INFO.
	size, err := redis.FetchDbsize(conn)
	if err != nil {
		return errors.Wrap(err, "failed to fetch dbsize info")
	}

	info := make(map[string]interface{})
	info["dbsize"] = strconv.FormatInt(size, 10)

	m.Logger().Debugf("Redis DBSIZE from %s: %+v", m.Host(), info)

	mb.Event{
		MetricSetFields: info,
	}

	return nil
}
