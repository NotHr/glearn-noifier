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

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config holds all configuration values
type Config struct {
	Credentials struct {
		Username string `koanf:"username"`
		Password string `koanf:"password"`
	} `koanf:"credentials"`
	URLs struct {
		Base   string `koanf:"base"`
		GLearn string `koanf:"glearn"`
	} `koanf:"urls"`
	Notification struct {
		NtfyURL string        `koanf:"ntfy_url"`
		Delay   time.Duration `koanf:"check_delay"`
	} `koanf:"notification"`
}

// LoadConfig loads the configuration from config.toml
func LoadConfig() (*Config, error) {
	k := koanf.New(".")

	// Load TOML file
	if err := k.Load(file.Provider("config.toml"), toml.Parser()); err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	var config Config
	if err := k.Unmarshal("", &config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate required fields
	if config.Credentials.Username == "" || config.Credentials.Password == "" {
		return nil, fmt.Errorf("username and password must be set in config.toml")
	}

	// Set defaults if not specified
	if config.URLs.Base == "" {
		config.URLs.Base = "https://login.gitam.edu"
	}
	if config.URLs.GLearn == "" {
		config.URLs.GLearn = "https://glearn.gitam.edu"
	}
	if config.Notification.NtfyURL == "" {
		config.Notification.NtfyURL = "https://ntfy.sh/nothrglearn"
	}
	if config.Notification.Delay == 0 {
		config.Notification.Delay = 5 * time.Minute
	}

	return &config, nil
}

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
	config    *Config
}

const (
	loginPath = "/Login.aspx"
	homePath  = "/Student/std_course_details"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
)

// NewClient creates a new authenticated client
func NewClient(config *Config) (*Client, error) {
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
		baseURL:   config.URLs.Base,
		glearnURL: config.URLs.GLearn,
		config:    config,
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
	resp, err := c.http.Get(c.baseURL + loginPath)
	if err != nil {
		return fmt.Errorf("fetching login page: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading login page: %w", err)
	}

	form, err := extractFormValues(string(bodyBytes))
	if err != nil {
		return fmt.Errorf("extracting form values: %w", err)
	}

	form.Username = creds.Username
	form.Password = creds.Password

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

// ParseAssignments scrapes and identifies assignment sections
func ParseAssignments(html string) []string {
	re := regexp.MustCompile(`<h5 class="cardTitle">.*?Scheduled assignments.*?</h5>.*?<div>(.*?)</div>`)
	matches := re.FindAllStringSubmatch(html, -1)

	assignments := []string{}
	for _, match := range matches {
		if len(match) > 1 {
			assignments = append(assignments, strings.TrimSpace(match[1]))
		}
	}

	return assignments
}

// SendNotification sends a notification via ntfy
func (c *Client) SendNotification(message string) error {
	resp, err := http.Post(c.config.Notification.NtfyURL, "text/plain", strings.NewReader(message))
	if err != nil {
		return fmt.Errorf("sending notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification failed with status: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Creating client: %v", err)
	}

	creds := Credentials{
		Username: config.Credentials.Username,
		Password: config.Credentials.Password,
	}

	if err := client.Login(creds); err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	fmt.Println("Logged in successfully. Starting periodic checks...")

	var lastAssignments []string
	for {
		homeContent, err := client.FetchHomePage()
		if err != nil {
			log.Printf("Error fetching home page: %v", err)
		} else {
			assignments := ParseAssignments(homeContent)
			if !isSameAssignments(lastAssignments, assignments) {
				fmt.Println("New or updated assignments found!")
				fmt.Println(assignments)

				message := fmt.Sprintf("New or updated assignments detected: %v", assignments)
				if err := client.SendNotification(message); err != nil {
					log.Printf("Error sending notification: %v", err)
				}

				lastAssignments = assignments
			} else {
				fmt.Println("No updates in assignments.")
			}
		}
		time.Sleep(config.Notification.Delay)
	}
}

func isSameAssignments(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
