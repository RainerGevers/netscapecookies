package netscapecookies

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

func ApplyCookiesToClient(client *http.Client, reader io.Reader) error {
	jar := client.Jar
	if jar == nil {
		jar, _ = cookiejar.New(nil)
		client.Jar = jar
	}

	cookies, err := readNetscapeCookies(reader)
	if err != nil {
		return err
	}

	for _, cookie := range cookies {
		u, err := url.Parse(fmt.Sprintf("https://%s", cookie.Domain))
		if err != nil {
			return fmt.Errorf("failed to parse URL for domain %s: %v", cookie.Domain, err)
		}

		jar.SetCookies(u, []*http.Cookie{cookie})
	}

	return nil
}

func readNetscapeCookies(reader io.Reader) ([]*http.Cookie, error) {
	var cookies []*http.Cookie
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}

		expires, _ := time.Parse("2006-01-02T15:04:05Z", fields[4])
		cookie := &http.Cookie{
			Domain:   fields[0],
			Path:     fields[2],
			Secure:   fields[3] == "TRUE",
			Expires:  expires,
			Name:     fields[5],
			Value:    fields[6],
			HttpOnly: true,
		}
		cookies = append(cookies, cookie)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return cookies, nil
}
