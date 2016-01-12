package notifier

import (
	log "github.com/AcalephStorage/consul-alerts/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/influxdb/influxdb/client/v2"
	"time"
)

type InfluxdbNotifier struct {
	Host       string
	Username   string
	Password   string
	Database   string
	SeriesName string
}

func (influxdb *InfluxdbNotifier) Notify(messages Messages) bool {

	// Make client
	influxdbClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     influxdb.Host,
		Username: influxdb.Username,
		Password: influxdb.Password,
	})

	if err != nil {
		log.Println("unable to access influxdb. can't send notification. ", err)
		return false
	}

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  influxdb.Database,
		Precision: "ms",
	})

	influxdb.toBatchPoints(messages, bp)
	err = influxdbClient.Write(bp)

	if err != nil {
		log.Println("unable to send notifications: ", err)
		return false
	}

	log.Println("influxdb notification sent.")
	return true
}

func (influxdb *InfluxdbNotifier) toBatchPoints(messages Messages, bp client.BatchPoints) {

	seriesName := influxdb.SeriesName

	for index, message := range messages {
		tags := map[string]string{
			"node":    message.Node,
			"service": message.Service,
			"status":  message.Status,
			"serviceId": message.ServiceId,
		}
		fields := map[string]interface{}{
			"checks": message.Check,
			"notes":  message.Notes,
			"output": message.Output,
		}

		p, err := client.NewPoint(seriesName, tags, fields, message.Timestamp)
		if err != nil {
			log.Println("Error: ", err.Error())
		}
		log.Debugf("%s", index)
		bp.AddPoint(p)
	}
}
