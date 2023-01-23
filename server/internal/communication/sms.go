package communication

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

)

type fast2smsRes struct {
	Return     bool
	Request_id string
	Message    []string
}

func SendSMS(text string, number string) error {
	endpoint := "https://www.fast2sms.com/dev/bulkV2?authorization=9xcUMv1KXIzuobRnVHLB8keJEmSGNPp53Asgqhitf6FwlT4QdWOzCsD1ah2vPji36AnepxBfqblITLkt&sender_id=CANDLY&message=" + text + "&route=v3&numbers=" + number
	res, err := http.Get(endpoint)
	if err != nil {
		return err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return readErr
	}

	bodyStruct := fast2smsRes{}
	jsonErr := json.Unmarshal(body, &bodyStruct)
	if jsonErr != nil {
		return fmt.Errorf("%w; %s", err, string(body))
	}

	if bodyStruct.Return {
		return nil
	}

	return errors.New("Unable to send sms;" + string(body))

}
