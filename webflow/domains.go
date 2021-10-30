package webflow

import (
	"context"
	"fmt"
)

// DomainsService handles communication with the meta related
// methods of the Webflow API.
//
// ref: https://developers.webflow.com/#domains
type DomainsService service

type Domain struct {
	Id   string `json:"_id"`
	Name string `json:"name"`
}

type ListDomainsOptions struct {
	ListOptions
}

// ListDomains lists of all domains for the provided site.
// ref: https://developers.webflow.com/#list-domains
func (s *DomainsService) ListDomains(ctx context.Context, site string, opts *ListDomainsOptions) ([]*Domain, error) {

	u := fmt.Sprintf("/sites/%s/domains", site)

	u, err := addOptions(u, opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var domains []*Domain
	_, err = s.client.Do(ctx, req, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}
