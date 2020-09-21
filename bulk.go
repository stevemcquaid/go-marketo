package marketo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
)

type Error struct {
	Message    string
	StatusCode int
	Body       string
}

func (e Error) Error() string {
	return e.Message
}

// BatchStatus describes the possible states for an import batch
type BatchStatus string

const (
	BatchComplete  = "Complete"
	BatchQueued    = "Queued"
	BatchImporting = "Importing"
	BatchFailed    = "Failed"
)

// BatchResult contains the details of a batch, returned by the Create
// & Get functions
type BatchResult struct {
	BatchID        int    `json:"batchId"`
	ImportID       string `json:"importId"`
	Status         string `json:"status"`
	LeadsProcessed int    `json:"numOfLeadsProcessed"`
	Failures       int    `json:"numOfRowsFailed"`
	Warnings       int    `json:"numOfRowsWithWarning"`
	Message        string `json:"message"`
}

// LeadImportResponse is returned from bulk lead import operations
type LeadImportResponse struct {
	RequestID string        `json:"requestId"`
	Success   bool          `json:"success"`
	Result    []BatchResult `json:"result"`
}

// ImportAPI provides access to the HubSpot import API
type ImportAPI struct {
	*Client
}

// NewImportAPI returns a new instance of the import API, configured
// using the provided options
func NewImportAPI(c *Client) *ImportAPI {
	return &ImportAPI{c}
}

// Create uploads a new file for importing, returning the new
// asynchronous import
func (i *ImportAPI) Create(ctx context.Context, name string, file io.Reader) (*LeadImportResponse, error) {
	buffer := &strings.Builder{}
	mpWriter := multipart.NewWriter(buffer)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="file"; filename="%s"`, "lead.csv"))

	fileWriter, err := mpWriter.CreatePart(h)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return nil, err
	}

	mpWriter.Close()
	request, err := http.NewRequest(http.MethodPost, i.url("bulk", "v1", "leads.json?format=csv"), bytes.NewBufferString(buffer.String()))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", mpWriter.FormDataContentType())

	resp, err := i.Client.doRequest(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, Error{
			Message:    "error creating import",
			Body:       string(body),
			StatusCode: resp.StatusCode,
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	results := &LeadImportResponse{}
	err = json.Unmarshal(body, results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// Get retrieves an existing import by its batch ID
func (i *ImportAPI) Get(ctx context.Context, id int) (*LeadImportResponse, error) {
	request, err := http.NewRequest(
		http.MethodGet, i.url("bulk", "v1", "leads", "batch", fmt.Sprintf("%d.json", id)), nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := i.Client.doRequest(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, Error{
			Message:    "error creating import",
			Body:       string(body),
			StatusCode: resp.StatusCode,
		}
	}

	result := &LeadImportResponse{}
	reader := json.NewDecoder(resp.Body)
	err = reader.Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
