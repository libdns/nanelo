// Package nanelo implements a DNS record management client compatible
// with the libdns interfaces for Nanelo
package nanelo

import (
	"context"
	"fmt"
	"encoding/json"
	"net/url"
	"net/http"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with Nanelo
type Provider struct {
	APIToken string `json:"api_token,omitempty"`
}

type APIResponse struct {
	OK     bool                    `json:"ok"`
	Error  *string                 `json:"error"`
	Result *map[string]interface{} `json:"result"`
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	baseURL, _ := url.Parse("https://api.nanelo.com/v1")
	baseURL = baseURL.JoinPath(p.APIToken, "dns", "addrecord")

	var added []libdns.Record

	for _, rec := range records {
		endpoint := baseURL
		query := endpoint.Query()
		query.Set("domain", zone)
		query.Set("name", rec.Name)
		query.Set("type", rec.Type)
		query.Set("value", rec.Value)
		query.Set("ttl", fmt.Sprintf("%f", rec.TTL.Seconds()))

		if rec.Priority > 0 {
			query.Set("priority", fmt.Sprintf("%d", rec.Priority))
		}

		endpoint.RawQuery = query.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
		if err != nil {
			return nil, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var apiResponse APIResponse
		err = json.NewDecoder(resp.Body).Decode(&apiResponse)
		if err != nil {
			return nil, err
		}

		if apiResponse.Error != nil {
			return nil, fmt.Errorf(*apiResponse.Error)
		}
		if apiResponse.OK == false {
			return nil, fmt.Errorf("Unknown Error when trying to create the DNS Record")
		}

		// Set ID for libdns v1.0.0
		rec.ID = fmt.Sprintf("%s:%s:%s", rec.Name, rec.Type, rec.Value)
		if rec.TTL == 0 {
			rec.TTL = time.Minute // Default TTL if not set
		}
		added = append(added, rec)
	}

	return added, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	baseURL, _ := url.Parse("https://api.nanelo.com/v1")
	baseURL = baseURL.JoinPath(p.APIToken, "dns", "deleterecord")

	var deleted []libdns.Record

	for _, rec := range records {
		// Parse ID if present
		if rec.ID != "" {
			parts := strings.Split(rec.ID, ":")
			if len(parts) == 3 {
				rec.Name = parts[0]
				rec.Type = parts[1]
				rec.Value = parts[2]
			} else {
				return nil, fmt.Errorf("invalid record ID format: %s", rec.ID)
			}
		}

		endpoint := baseURL
		query := endpoint.Query()
		query.Set("domain", zone)
		query.Set("name", rec.Name)
		query.Set("type", rec.Type)
		query.Set("value", rec.Value)
		query.Set("ttl", fmt.Sprintf("%f", rec.TTL.Seconds()))

		if rec.Priority > 0 {
			query.Set("priority", fmt.Sprintf("%d", rec.Priority))
		}

		endpoint.RawQuery = query.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
		if err != nil {
			return nil, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var apiResponse APIResponse
		err = json.NewDecoder(resp.Body).Decode(&apiResponse)
		if err != nil {
			return nil, err
		}

		if apiResponse.Error != nil {
			return nil, fmt.Errorf(*apiResponse.Error)
		}
		if apiResponse.OK == false {
			return nil, fmt.Errorf("Unknown Error when trying to create the DNS Record")
		}

		deleted = append(deleted, rec)
	}

	return deleted, nil
}

// Interface guards
var (
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
