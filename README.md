# go-webflow

A Webflow API client enabling Go programs to interact with Webflow in a simple and uniform way.

Inspired by [go-github](https://github.com/google/go-github).


## Supported APIs

- [x] [Meta](https://developers.webflow.com/#meta)
- [x] [Sites](https://developers.webflow.com/#sites)
- [x] [Domains](https://developers.webflow.com/#domains)
- [x] [Collections](https://developers.webflow.com/#collections)
- [x] [Items](https://developers.webflow.com/#items)
- [ ] [Ecommerce](https://developers.webflow.com/#ecommerce)
- [ ] [Webhooks](https://developers.webflow.com/#webhooks)

## Usage

```go
import "github.com/aesadde/go-webflow"
```

Then construct a new client providing your [Webflow API Key](https://developers.webflow.com/#authentication) and use
the client to interact with the different parts of the [API](#supported-apis).

For example you can use the sites and domains APIs to publish a site
```go
    client := webflow.NewClient("YOUR_API_KEY")

    siteId := "YOUR_SITE_ID"

    site, err := client.Sites.GetSite(context.Background(), siteId)
    if err != nil {
        return
    }

    domains, err := client.Domains.ListDomains(context.Background(), site.Id, nil)
    if err != nil {
        return
    }

    toPublish := make([]string, len(domains))
    for i, d := range domains {
        toPublish[i] = d.Name
    }

    published, err := client.Sites.PublishSite(context.Background(), siteId, toPublish)
    if err != nil {
        return
    }
    fmt.Println(published)
```

## Examples

A full list of examples can be found in the [examples](./examples) folder.


## Issues
If you have an issue: report it on the [issue tracker](https://github.com/aesadde/go-webflow/issues).

## Author
Alberto Sadde ([@aesadde](https://github.com/aesadde)).