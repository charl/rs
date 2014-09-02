package rs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var (
	client = &http.Client{}
)

// A Rackspace account that is the minimum requirement to
// authenticate with the Rackspace service.
type Account struct {
	user   string
	apiKey string
}

// Creates a new account.
func NewAccount(user, apiKey string) *Account {
	return &Account{user: user, apiKey: apiKey}
}

// The service endpoints that make up the service catalog.
type Endpoint struct {
	InternalUrl string `json:"internalURL"`
	PublicUrl   string `json:"publicURL"`
	Region      string `json:"region"`
	TenantId    string `json:"tenantId"`
}

// The list od services our permissions allow us access to.
type ServiceCatalog struct {
	Endpoints []Endpoint `json:"endpoints"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
}

// The tennant data linked to an authenticated session token.
type Tenant struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// The authenticated session token data.
type Token struct {
	AuthenticatedBy []string `json:"RAX-AUTH:authenticatedBy"`
	Expires         string   `json:"expires"`
	Id              string   `json:"id"`
	Tenant          Tenant   `json:"tenant"`
}

// The roles an authenticated user has access to.
type Role struct {
	Description string `json:"description"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	TenantId    string `json:"tenanId"`
}

// The user profile data for this authenticated request.
type User struct {
	DefaultRegion string `json:"RAX-AUTH:defaultRegion"`
	Id            string `json:"id"`
	Name          string `json:"name"`
	Roles         []Role
}

// The permissions allowed for this authtnetication request.
type Access struct {
	ServiceCatalog []ServiceCatalog `json:"serviceCatalog"`
	Token          Token            `json:"token"`
	User           User             `json:"user"`
}

// The identity data returned from an authentication request.
type IdentityData struct {
	Access Access `json:"access"`
}

// An authentication identity used in conjunctioon with an
// account to create an authenticated session.
type Identity struct {
	url     string
	account Account
	Access  Access
}

// Creates a new identity.
func NewIdentity(url string, account Account) *Identity {
	return &Identity{url: url, account: account}
}

// Create an authenticated session.
func (i *Identity) Authenticate() error {
	// Fire off the API authentication request.
	data := fmt.Sprintf(`{"auth": {"RAX-KSKEY:apiKeyCredentials": {"username": "%s", "apiKey": "%s"}}}`, i.account.user, i.account.apiKey)
	resp, err := http.Post(i.url, "application/json", strings.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Decode the json response.
	var a IdentityData
	dec := json.NewDecoder(resp.Body)
	for {
		if err := dec.Decode(&a); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}
	i.Access = a.Access

	return nil
}

// A Rackspace container.
type Container struct {
	Name         string `json:"name"`
	Count        int64  `json:"count"`
	Bytes        int64  `json:"bytes"`
	Uri          string `json:"cdn_uri"`
	StreamingUri string `json:"cdn_streaming_uri"`
	IosUri       string `json:"cdn_ios_uri"`
	SslUri       string `json:"cdn_ssl_uri"`
	Enabled      bool   `json:"cdn_enabled"`
	Ttl          int64  `json:"ttl"`
	LogRetention bool   `json:"log_retention"`
}

// Get a list of all containers.
func AllContainers(endpoint, authToken string) (*[]Container, error) {
	var containers *[]Container

	req, err := http.NewRequest("GET", endpoint+"?format=json", nil)
	if err != nil {
		return containers, err
	}
	req.Header.Add("X-Auth-Token", authToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return containers, err
	}
	defer resp.Body.Close()

	// Decode the json response.
	dec := json.NewDecoder(resp.Body)
	for {
		if err := dec.Decode(&containers); err == io.EOF {
			break
		} else if err != nil {
			return containers, err
		}
	}

	return containers, nil
}

// Check for the existence of a container.
func ContainerExists(endpoint, authToken, name string) bool {
	containers, err := AllContainers(endpoint, authToken)
	if err != nil {
		log.Printf("Error: ContainerExists: %", err.Error())
		return false
	}

	for _, c := range *containers {
		if c.Name == name {
			return true
		}
	}

	return false
}

// Create a new container.
func CreateContainer(endpoint, authToken, name string) error {
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", endpoint, name), nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-Auth-Token", authToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		return fmt.Errorf("Error: cannot create container: %s: %d", name, resp.StatusCode)
	}

	return nil
}
