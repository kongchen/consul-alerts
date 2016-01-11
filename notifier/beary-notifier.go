package notifier

import (
	"bytes"
	"fmt"
	"strings"

	"io/ioutil"

	"encoding/json"
	"net/http"

	log "github.com/AcalephStorage/consul-alerts/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

type BearyNotifier struct {
	ClusterName string            `json:"-"`
	Url         string            `json:"-"`
	IconUrl     string            `json:"icon_url"`
	IconEmoji   string            `json:"icon_emoji"`
	Text        string            `json:"text,omitempty"`
	Attachments []bearyattachment `json:"attachments,omitempty"`
	Detailed    bool              `json:"-"`
}

type bearyattachment struct {
	Color    string   `json:"color"`
	Title    string   `json:"title"`
	Pretext  string   `json:"pretext"`
	Text     string   `json:"text"`
	MrkdwnIn []string `json:"mrkdwn_in"`
}

func (beary *BearyNotifier) Notify(messages Messages) bool {

	if beary.Detailed {
		return beary.notifyDetailed(messages)
	} else {
		return beary.notifySimple(messages)
	}

}

func (beary *BearyNotifier) notifySimple(messages Messages) bool {

	overallStatus, pass, warn, fail := messages.Summary()

	text := fmt.Sprintf(header, beary.ClusterName, overallStatus, fail, warn, pass)

	for _, message := range messages {
		text += fmt.Sprintf("\n%s:%s:%s is %s.", message.Node, message.Service, message.Check, message.Status)
		text += fmt.Sprintf("\n%s", message.Output)
	}

	beary.Text = text

	return beary.postToBeary()

}

func (beary *BearyNotifier) notifyDetailed(messages Messages) bool {

	overallStatus, pass, warn, fail := messages.Summary()

	var emoji, color string
	switch overallStatus {
	case SYSTEM_HEALTHY:
		emoji = ":white_check_mark:"
		color = "good"
	case SYSTEM_UNSTABLE:
		emoji = ":question:"
		color = "warning"
	case SYSTEM_CRITICAL:
		emoji = ":x:"
		color = "danger"
	default:
		emoji = ":question:"
	}
	title := "Consul monitoring report"
	pretext := fmt.Sprintf("%s %s is *%s*", emoji, beary.ClusterName, overallStatus)

	detailedBody := ""
	detailedBody += fmt.Sprintf("*Changes:* Fail = %d, Warn = %d, Pass = %d", fail, warn, pass)
	detailedBody += fmt.Sprintf("\n")

	for _, message := range messages {
		detailedBody += fmt.Sprintf("\n*[%s:%s]* %s is *%s.*", message.Node, message.Service, message.Check, message.Status)
		detailedBody += fmt.Sprintf("\n`%s`", strings.TrimSpace(message.Output))
	}

	a := bearyattachment{
		Color:    color,
		Title:    title,
		Pretext:  pretext,
		Text:     detailedBody,
		MrkdwnIn: []string{"text", "pretext"},
	}
	beary.Attachments = []bearyattachment{a}

	return beary.postToBeary()

}

func (beary *BearyNotifier) postToBeary() bool {

	data, err := json.Marshal(beary)
	if err != nil {
		log.Println("Unable to marshal beary payload:", err)
		return false
	}
	log.Debugf("struct = %+v, json = %s", beary, string(data))

	b := bytes.NewBuffer(data)
	if res, err := http.Post(beary.Url, "application/json", b); err != nil {
		log.Println("Unable to send data to beary:", err)
		return false
	} else {
		defer res.Body.Close()
		statusCode := res.StatusCode
		if statusCode != 200 {
			body, _ := ioutil.ReadAll(res.Body)
			log.Println("Unable to notify beary:", string(body))
			return false
		} else {
			log.Println("Beary notification sent.")
			return true
		}
	}

}
