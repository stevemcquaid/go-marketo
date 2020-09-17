# go-marketo

[![GoDoc](https://godoc.org/github.com/polytomic/go-marketo?status.svg)](https://godoc.org/github.com/polytomic/go-marketo)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/polytomic/go-marketo/master/LICENSE)

Inspired by [FrenchBen/goketo](https://github.com/FrenchBen/goketo)
and derived from
[SpeakData/minimarketo](https://github.com/SpeakData/goketo),
**go-marketo** provides a thin wrapper around Marketo's REST APIs,
along with some utility structs for handling responses.


## Installation

```bash
go get github.com/polytomic/go-marketo
```

## Usage

Basic operations are:
1. Create a client
2. Make a http call (Marketo API only supports GET, POST, DELETE) with url string and data in []byte if needed
3. Check "success" and parse "result" with your struct

First, create a client.
In this example, I'm passing configuration through environment variables.
```go
config := marketo.ClientConfig{
    ID:       os.Getenv("MARKETO_ID"),
    Secret:   os.Getenv("MARKETO_SECRET"),
    Endpoint: os.Getenv("MARKETO_URL"), // https://XXX-XXX-XXX.mktorest.com
    Debug:    true,
}
client, err := marketo.NewClient(config)
if err != nil {
    log.Fatal(err)
}
```

Then, call Marketo supported http calls: GET, POST, or DELETE.

Find a lead
```go
path := "/rest/v1/leads.json?"
v := url.Values{
    "filterType":   {"email"},
    "filterValues": {"tester@example.com"},
    "fields":       {"email"},
}
response, err := client.Get(path + v.Encode())
if err != nil {
    log.Fatal(err)
}
if !response.Success {
    log.Fatal(response.Errors)
}
var leads []marketo.LeadResult
if err = json.Unmarshal(response.Result, &leads); err != nil {
    log.Fatal(err)
}
for _, lead := range leads {
    fmt.Printf("%+v", lead)
}
```

Create a lead
```go
path := "/rest/v1/leads.json"
type Input struct {
    Email     string `json:"email"`
    FirstName string `json:"firstName"`
    LastName  string `json:"lastName"`
}
type CreateData struct {
    Action      string  `json:"action"`
    LookupField string  `json:"lookupField"`
    Input       []Input `json:"input"`
}
data := CreateData{
    "createOnly",
    "email",
    []Input{
        Input{"tester@example.com", "John", "Doe"},
    },
}

dataInBytes, err := json.Marshal(data)
response, err := client.Post(path, dataInBytes)
if err != nil {
    log.Fatal(err)
}
if !response.Success {
    log.Fatal(response.Errors)
}
var createResults []marketo.RecordResult
if err = json.Unmarshal(response.Result, &createResults); err != nil {
    log.Fatal(err)
}
for _, result := range createResults {
    fmt.Printf("%+v", result)
}
```

## JSON Response

go-marketo defines the common Marketo response format.
This covers most of the API responses.

```go
type Response struct {
	RequestID     string `json:"requestId"`
	Success       bool   `json:"success"`
	NextPageToken string `json:"nextPageToken,omitempty"`
	MoreResult    bool   `json:"moreResult,omitempty"`
	Errors        []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors,omitempty"`
	Result   json.RawMessage `json:"result,omitempty"`
	Warnings []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"warning,omitempty"`
}
```

Your job is to parse "Result".

As for convenience, go-marketo defines two commonly used "result" format.

```go
// Find lead returns "result" in this format
type LeadResult struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Created   string `json:"createdAt"`
	Updated   string `json:"updatedAt"`
}

// Create/update lead uses this format
type RecordResult struct {
	ID      int    `json:"id"`
	Status  string `json:"status"`
	Reasons []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"reasons,omitempty"`
}
```

## License

MIT

