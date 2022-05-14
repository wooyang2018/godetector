package main

import (
	vegeta "github.com/tsenart/vegeta/lib"
	"net/http"
	"os"
	"time"
)

func customTargeter() vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		if tgt == nil {
			return vegeta.ErrNilTarget
		}

		tgt.Method = "POST"

		tgt.URL = "http://172.28.48.170:8000/nsq/text" // your url here

		payload := `{
            "text" : ""老子进小黑屋跟你妈做爱跟上厕所一样随意，你妈不吃延更丹都闭经啦，就当你爷俩面抽插你妈大黑逼""
          }` // you can make this salon_id dynamic too, using random or uuid

		tgt.Body = []byte(payload)

		header := http.Header{}
		header.Add("Accept", "application/json")
		header.Add("Content-Type", "application/json")
		tgt.Header = header

		return nil
	}
}

func main() {
	rate := vegeta.Rate{Freq: 1, Per: time.Second} // change the rate here
	duration := 1 * time.Minute                    // change the duration here

	targeter := customTargeter()
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration, "Whatever name") {
		metrics.Add(res)
	}
	metrics.Close()

	reporter := vegeta.NewTextReporter(&metrics)
	reporter(os.Stdout)
}
