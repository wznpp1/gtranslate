package github.com/wznpp1/gtranslate

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/text/language"

	gout2 "github.com/guonaihong/gout"
	"github.com/robertkrimen/otto"
)

var ttk otto.Value

func init() {
	ttk, _ = otto.ToValue("0")
}

const (
	defaultNumberOfRetries = 2
)

func translate(text, from, to string, withVerification bool, tries int, delay time.Duration) (string, error) {
	if tries == 0 {
		tries = defaultNumberOfRetries
	}

	if withVerification {
		if _, err := language.Parse(from); err != nil && from != "auto" {
			log.Println("[WARNING], '" + from + "' is a invalid language, switching to 'auto'")
			from = "auto"
		}
		if _, err := language.Parse(to); err != nil {
			log.Println("[WARNING], '" + to + "' is a invalid language, switching to 'en'")
			to = "en"
		}
	}

	t, _ := otto.ToValue(text)

	urll := fmt.Sprintf("https://translate.%s/translate_a/single", GoogleHost)

	token := get(t, ttk)

	data := map[string]string{
		"client": "gtx",
		"sl":     from,
		"tl":     to,
		"hl":     to,
		// "dt":     []string{"at", "bd", "ex", "ld", "md", "qca", "rw", "rm", "ss", "t"},
		"ie":   "UTF-8",
		"oe":   "UTF-8",
		"otf":  "1",
		"ssel": "0",
		"tsel": "0",
		"kc":   "7",
		"q":    text,
	}

	u, err := url.Parse(urll)
	if err != nil {
		return "", nil
	}

	parameters := url.Values{}

	for k, v := range data {
		parameters.Add(k, v)
	}
	for _, v := range []string{"at", "bd", "ex", "ld", "md", "qca", "rw", "rm", "ss", "t"} {
		parameters.Add("dt", v)
	}

	parameters.Add("tk", token)
	u.RawQuery = parameters.Encode()

	var resp []interface{}

	for tries > 0 {
		StatusCode := 0
		err := gout2.
			New().
			GET(u.String()).
			SetSOCKS5("127.0.0.1:1100").
			Code(&StatusCode).
			BindJSON(&resp).
			Do()

		if err != nil {
			return "", err
		}

		if StatusCode == http.StatusOK {
			break
		}

		if StatusCode == http.StatusForbidden {
			tries--
			time.Sleep(delay)
		}
	}

	Data, err := json.Marshal(resp)
	if err == nil {
		File2, err := os.OpenFile(fmt.Sprintf("%d_translate.json", int(time.Now().UnixNano())), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

		if err != nil {
			fmt.Println("Open File2 error")
		}
		File2.Write(Data)
		File2.Close()
	}

	responseText := ""
	for _, obj := range resp[0].([]interface{}) {
		if len(obj.([]interface{})) == 0 {
			break
		}

		t, ok := obj.([]interface{})[0].(string)
		if ok {
			responseText += t
		}
	}

	return responseText, nil
}
