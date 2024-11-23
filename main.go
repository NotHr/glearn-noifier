package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Credentials stores login information
type Credentials struct {
	Username string
	Password string
}

// LoginForm represents the form data needed for login
type LoginForm struct {
	ViewState          string
	EventValidation    string
	ViewStateGenerator string
	Username           string
	Password           string
}

// Client wraps http.Client with additional functionality
type Client struct {
	http      *http.Client
	baseURL   string
	glearnURL string
}

const (
	loginPath = "/Login.aspx"
	homePath  = "/Student/std_course_details"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
)

// NewClient creates a new authenticated client
func NewClient(baseURL string, glearnURL string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Client{
		http:      client,
		baseURL:   baseURL,
		glearnURL: glearnURL,
	}, nil
}

// extractFormValues extracts hidden form values from HTML
func extractFormValues(body string) (LoginForm, error) {
	form := LoginForm{
		ViewStateGenerator: "C2EE9ABB", // Default value
	}

	patterns := map[string]*string{
		`id="__VIEWSTATE" value="(.*?)"`:       &form.ViewState,
		`id="__EVENTVALIDATION" value="(.*?)"`: &form.EventValidation,
	}

	for pattern, target := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(body)
		if len(matches) < 2 {
			return form, fmt.Errorf("could not find pattern: %s", pattern)
		}
		*target = matches[1]
	}

	return form, nil
}

// Login performs the login process
func (c *Client) Login(creds Credentials) error {
	// Get login page
	resp, err := c.http.Get(c.baseURL + loginPath)
	if err != nil {
		return fmt.Errorf("fetching login page: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading login page: %w", err)
	}

	// Extract form values
	form, err := extractFormValues(string(bodyBytes))
	if err != nil {
		return fmt.Errorf("extracting form values: %w", err)
	}

	form.Username = creds.Username
	form.Password = creds.Password

	// Prepare login data
	formData := url.Values{
		"__EVENTTARGET":        {""},
		"__EVENTARGUMENT":      {""},
		"__VIEWSTATE":          {form.ViewState},
		"__VIEWSTATEGENERATOR": {form.ViewStateGenerator},
		"__EVENTVALIDATION":    {form.EventValidation},
		"txtusername":          {form.Username},
		"password":             {form.Password},
		"Submit":               {"Login"},
	}

	// Create login request
	req, err := http.NewRequest(
		"POST",
		c.baseURL+loginPath,
		bytes.NewBufferString(formData.Encode()),
	)
	if err != nil {
		return fmt.Errorf("creating login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.baseURL+loginPath)
	req.Header.Set("User-Agent", userAgent)

	// Perform login
	resp, err = c.http.Do(req)
	if err != nil {
		return fmt.Errorf("performing login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	return nil
}

// FetchHomePage retrieves the home page content
func (c *Client) FetchHomePage() (string, error) {
	fmt.Println(c.glearnURL + homePath)
	resp, err := c.http.Get(c.glearnURL + homePath)
	if err != nil {
		return "", fmt.Errorf("fetching home page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading home page: %w", err)
	}

	return string(body), nil
}

func main() {
	creds := Credentials{
		Username: "username",
		Password: "password",
	}

	client, err := NewClient("https://login.gitam.edu", "https://glearn.gitam.edu")
	if err != nil {
		log.Fatalf("Creating client: %v", err)
	}

	if err := client.Login(creds); err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	homeContent, err := client.FetchHomePage()
	if err != nil {
		log.Fatalf("Fetching home page: %v", err)
	}

	fmt.Println("Successfully logged in and fetched home page:")
	fmt.Println(strings.TrimSpace(homeContent))
}

