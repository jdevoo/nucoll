package twitter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jdevoo/nucoll/util"
)

// NucollTransport holds the config and the structure to deal with throttling
type NucollTransport struct {
	Config    *util.NucollConfig
	Transport http.RoundTripper
}

// APIError to hold message and code
type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// APIErrors holds a collection of errors
type APIErrors struct {
	Errors []APIError `json:"errors"`
}

// RoundTrip intercepts API responses and checks if a throttling pause is required
func (t *NucollTransport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	if t.Config.AccessToken != "" {
		req.Header.Add("Authorization", "Bearer "+t.Config.AccessToken)
	}
RT:
	for res, err = t.Transport.RoundTrip(req); err == nil; {
		switch res.StatusCode {
		case http.StatusUnauthorized:
			log.Println(res.Status)
			break RT
		case http.StatusServiceUnavailable:
			fallthrough
		case http.StatusTooManyRequests:
			x, err := strconv.ParseInt(res.Header.Get("x-rate-limit-reset"), 10, 64)
			if err != nil {
				return nil, err
			}
			win := time.Unix(x+5, 0)
			log.Printf("response code %d received; waiting until %s to resume", res.StatusCode, win.Format("15:04:05"))
			time.Sleep(time.Until(win))
			res, err = t.Transport.RoundTrip(req)
		case http.StatusOK:
			break RT
		default:
			err = errors.New(res.Status)
			break RT
		}
	}
	return res, err
}

func (t *NucollTransport) getToken(consumerKey string, consumerSecret string) error {
	var endpoint = "https://api.twitter.com/oauth2/token"
	req, _ := http.NewRequest("POST", endpoint, strings.NewReader("grant_type=client_credentials"))
	req.SetBasicAuth(consumerKey, consumerSecret)
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	res, err := t.RoundTrip(req)
	defer res.Body.Close()
	if err != nil {
		return err
	}
	body, _ := ioutil.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		var te APIErrors
		if err := json.Unmarshal(body, &te); err != nil {
			return errors.New(res.Status)
		}
		buf := bytes.NewBufferString("")
		for i := range te.Errors {
			fmt.Fprintf(buf, "%s (%d)\n", te.Errors[i].Message, te.Errors[i].Code)
		}
		return errors.New(buf.String())
	}
	var conf util.TwitterConfig
	if err := json.Unmarshal(body, &conf); err != nil {
		return err
	}
	if conf.TokenType != "bearer" {
		return errors.New("invalid token type")
	}
	t.Config.TokenType = conf.TokenType
	t.Config.AccessToken = conf.AccessToken
	return nil
}

// NewClient implements application-only authentication for CLI usage
// https://developer.twitter.com/en/docs/basics/authentication/overview/application-only
func NewClient() (*http.Client, error) {
	t := &NucollTransport{}
	t.Transport = http.DefaultTransport

	var err error
	if t.Config, err = util.ReadConfig(); err != nil {
		return nil, err
	}
	if t.Config.TokenType == "" || t.Config.AccessToken == "" {
		fmt.Println(`
===TWITTER API AUTHENTICATION SETUP==============================
Open the following link and register this application...
>>> https://apps.twitter.com/`)
		fmt.Print("What is the consumer key? ")
		reader := bufio.NewReader(os.Stdin)
		consumerKey, _ := reader.ReadString('\n')
		consumerKey = strings.TrimSpace(consumerKey)
		fmt.Print("What is the consumer secret? ")
		consumerSecret, _ := reader.ReadString('\n')
		consumerSecret = strings.TrimSpace(consumerSecret)
		if err = t.getToken(consumerKey, consumerSecret); err != nil {
			return nil, err
		}
		if err = util.WriteConfig(t.Config); err != nil {
			return nil, err
		}
	}

	return &http.Client{Transport: t}, nil
}
