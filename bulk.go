package marketo

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
)

// BatchStatus describes the possible states for an import batch
type BatchStatus string

const (
	BatchComplete  = "Complete"
	BatchQueued    = "Queued"
	BatchImporting = "Importing"
	BatchFailed    = "Failed"
)

const (
	createLeadImport      = "create lead import"
	getLeadImport         = "get lead import"
	getLeadImportFailures = "get lead import failures"
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

// ImportAPI provides access to the Marketo import API
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
func (i *ImportAPI) Create(ctx context.Context, file io.Reader) (*LeadImportResponse, error) {
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
		return nil, handleError(createLeadImport, resp)
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
		return nil, handleError(getLeadImport, resp)
	}

	result := &LeadImportResponse{}
	reader := json.NewDecoder(resp.Body)
	err = reader.Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// LeadImportFailure contains a single lead record failure, along with
// the reason for failure.
type LeadImportFailure struct {
	Reason string
	Fields map[string]interface{}
}

// Failures returns the list of failed recrods for an import
func (i *ImportAPI) Failures(ctx context.Context, id int) ([]LeadImportFailure, error) {
	request, err := http.NewRequest(
		http.MethodGet, i.url("bulk", "v1", "leads", "batch", fmt.Sprintf("%d", id), "failures.json"), nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := i.Client.doRequest(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		// no errors
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, handleError(getLeadImportFailures, resp)
	}

	reader := csv.NewReader(resp.Body)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	failures := []LeadImportFailure{}
	record, err := reader.Read()
	for err == nil {
		failure := LeadImportFailure{
			Reason: record[len(header)-1],
			Fields: map[string]interface{}{},
		}
		for i := 0; i < len(header)-1; i++ {
			failure.Fields[header[i]] = record[i]
		}
		failures = append(failures, failure)
		record, err = reader.Read()
	}
	return failures, nil
}
