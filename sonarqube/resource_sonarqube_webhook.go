package sonarqube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// GetWebhooks for unmarshalling response body from geting webhooks
type GetWebhooks struct {
	Webhooks []Webhook `json:"webhooks"`
}

// Webhook type
type Webhook struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Secret string `json:"secret,omitempty"`
}

// CreateWebhookResponse struct
type CreateWebhookResponse struct {
	Webhook Webhook `json:"webhook"`
}

// Returns the resource represented by this file.
func resourceSonarqubeWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceSonarqubeWebhookCreate,
		Read:   resourceSonarqubeWebhookRead,
		Update: resourceSonarqubeWebhookUpdate,
		Delete: resourceSonarqubeWebhookDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSonarqubeWebhookImport,
		},

		// Define the fields of this schema.
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"project": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"secret": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceSonarqubeWebhookCreate(d *schema.ResourceData, m interface{}) error {
	sonarQubeURL := m.(*ProviderConfiguration).sonarQubeURL
	sonarQubeURL.Path = "api/webhooks/create"
	rawQuery := url.Values{
		"name": []string{d.Get("name").(string)},
		"url":  []string{d.Get("url").(string)},
	}

	if project, ok := d.GetOk("project"); ok {
		rawQuery.Add("project", project.(string))
	}
	if secret, ok := d.GetOk("secret"); ok {
		rawQuery.Add("secret", secret.(string))
	}

	sonarQubeURL.RawQuery = rawQuery.Encode()

	resp, err := httpRequestHelper(
		m.(*ProviderConfiguration).httpClient,
		"POST",
		sonarQubeURL.String(),
		http.StatusOK,
		"resourceSonarqubeWebhookCreate",
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	webhookResponse := CreateWebhookResponse{}
	err = json.NewDecoder(resp.Body).Decode(&webhookResponse)
	if err != nil {
		return fmt.Errorf("resourceSonarqubeWebhookCreate; Failed to decode json into struct: %+v", err)
	}

	if webhookResponse.Webhook.Key == "" {
		return fmt.Errorf("resourceSonarqubeWebhookCreate: Create response did not contain the webhook's key")
	}
	d.SetId(webhookResponse.Webhook.Key)
	return resourceSonarqubeWebhookRead(d, m)
}

func resourceSonarqubeWebhookRead(d *schema.ResourceData, m interface{}) error {
	sonarQubeURL := m.(*ProviderConfiguration).sonarQubeURL
	sonarQubeURL.Path = "api/webhooks/list"

	if project, ok := d.GetOk("project"); ok {
		sonarQubeURL.RawQuery = url.Values{
			"project": []string{project.(string)},
		}.Encode()
	}

	resp, err := httpRequestHelper(
		m.(*ProviderConfiguration).httpClient,
		"GET",
		sonarQubeURL.String(),
		http.StatusOK,
		"resourceSonarqubeWebhookRead",
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Decode response into struct
	getWebhooks := GetWebhooks{}
	err = json.NewDecoder(resp.Body).Decode(&getWebhooks)
	if err != nil {
		return fmt.Errorf("resourceSonarqubeWebhookRead: Failed to decode json into struct: %+v", err)
	}

	// Loop over all webhooks to see if the webhook we need exists.
	webhookFound := false
	for _, value := range getWebhooks.Webhooks {
		if d.Id() == value.Key {
			// If it does, set the values of that webhook
			d.SetId(value.Key)
			d.Set("key", value.Key)
			d.Set("name", value.Name)
			d.Set("url", value.URL)
			d.Set("secret", value.Secret)
			webhookFound = true
			break
		}
	}

	if !webhookFound {
		d.SetId("")
	}

	return nil
}

func resourceSonarqubeWebhookUpdate(d *schema.ResourceData, m interface{}) error {
	sonarQubeURL := m.(*ProviderConfiguration).sonarQubeURL
	sonarQubeURL.Path = "api/webhooks/update"
	rawQuery := url.Values{
		"webhook": []string{d.Id()},
		"name":    []string{d.Get("name").(string)},
		"url":     []string{d.Get("url").(string)},
	}

	if project, ok := d.GetOk("project"); ok {
		rawQuery.Add("project", project.(string))
	}
	if secret, ok := d.GetOk("secret"); ok {
		rawQuery.Add("secret", secret.(string))
	}

	sonarQubeURL.RawQuery = rawQuery.Encode()

	resp, err := httpRequestHelper(
		m.(*ProviderConfiguration).httpClient,
		"POST",
		sonarQubeURL.String(),
		http.StatusNoContent,
		"resourceSonarqubeWebhookUpdate",
	)
	if err != nil {
		return fmt.Errorf("Error updating Sonarqube webhook: %+v", err)
	}
	defer resp.Body.Close()

	return resourceSonarqubeWebhookRead(d, m)
}

func resourceSonarqubeWebhookDelete(d *schema.ResourceData, m interface{}) error {
	sonarQubeURL := m.(*ProviderConfiguration).sonarQubeURL
	sonarQubeURL.Path = "api/webhooks/delete"
	sonarQubeURL.RawQuery = url.Values{
		"webhook": []string{d.Id()},
	}.Encode()

	resp, err := httpRequestHelper(
		m.(*ProviderConfiguration).httpClient,
		"POST",
		sonarQubeURL.String(),
		http.StatusNoContent,
		"resourceSonarqubeWebhookDelete",
	)
	if err != nil {
		return fmt.Errorf("resourceSonarqubeWebhookDelete: Failed to delete webhook: %+v", err)
	}
	defer resp.Body.Close()

	return nil
}

func resourceSonarqubeWebhookImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := resourceSonarqubeWebhookRead(d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
