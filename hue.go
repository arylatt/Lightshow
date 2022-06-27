package lightshow

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
)

func GetApplicationID(host, user string) (id string, err error) {
	reqURL, err := url.Parse(fmt.Sprintf("https://%s/auth/v1", host))
	if err != nil {
		return
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return
	}

	req.Header.Add("hue-application-key", user)

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	return resp.Header.Get("hue-application-id"), nil
}
