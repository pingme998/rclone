package jwtutil

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/pingme998/rclone/fs"
	"github.com/pingme998/rclone/fs/config/configmap"
	"github.com/pingme998/rclone/lib/oauthutil"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"
)

// RandomHex creates a random string of the given length
func RandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Config configures rclone using JWT
func Config(id, name string, claims *jws.ClaimSet, header *jws.Header, queryParams map[string]string, privateKey *rsa.PrivateKey, m configmap.Mapper, client *http.Client) (err error) {
	payload, err := jws.Encode(header, claims, privateKey)
	if err != nil {
		return errors.Wrap(err, "jwtutil: failed to encode payload")
	}
	req, err := http.NewRequest("POST", claims.Aud, nil)
	if err != nil {
		return errors.Wrap(err, "jwtutil: failed to create new request")
	}
	q := req.URL.Query()
	q.Add("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	q.Add("assertion", payload)
	for key, value := range queryParams {
		q.Add(key, value)
	}
	queryString := q.Encode()

	req, err = http.NewRequest("POST", claims.Aud, bytes.NewBuffer([]byte(queryString)))
	if err != nil {
		return errors.Wrap(err, "jwtutil: failed to create new request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "jwtutil: failed making auth request")
	}

	s, err := bodyToString(resp.Body)
	if err != nil {
		fs.Debugf(nil, "jwtutil: failed to get response body")
	}
	if resp.StatusCode != 200 {
		err = errors.New(resp.Status)
		return errors.Wrap(err, "jwtutil: failed making auth request")
	}
	defer func() {
		deferedErr := resp.Body.Close()
		if deferedErr != nil {
			err = errors.Wrap(err, "jwtutil: failed to close resp.Body")
		}
	}()

	result := &response{}
	err = json.NewDecoder(strings.NewReader(s)).Decode(result)
	if result.AccessToken == "" && err == nil {
		err = errors.New("No AccessToken in Response")
	}
	if err != nil {
		return errors.Wrap(err, "jwtutil: failed to get token")
	}
	token := &oauth2.Token{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
	}
	e := result.ExpiresIn
	if e != 0 {
		token.Expiry = time.Now().Add(time.Duration(e) * time.Second)
	}
	return oauthutil.PutToken(name, m, token, true)
}

func bodyToString(responseBody io.Reader) (bodyString string, err error) {
	bodyBytes, err := ioutil.ReadAll(responseBody)
	if err != nil {
		return "", err
	}
	bodyString = string(bodyBytes)
	fs.Debugf(nil, "jwtutil: Response Body: "+bodyString)
	return bodyString, nil
}

type response struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
