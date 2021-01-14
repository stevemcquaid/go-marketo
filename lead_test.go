package marketo

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestLeadDescribe(t *testing.T) {
	defer gock.Off()

	gock.New(testHost).
		Get("/identity/oauth/token").
		Reply(http.StatusOK).
		JSON(authResponseSuccess)
	gock.New(testHost).
		Get("/rest/v1/leads/describe2.json").
		Reply(http.StatusOK).
		File("test-fixtures/leads-describe2.json")

	client, err := NewClient(ClientConfig{
		ID:       clientID,
		Secret:   clientSecret,
		Endpoint: "https://marketo.testing",
		Debug:    true,
	})
	require.NoError(t, err)

	api := NewLeadAPI(client)
	fields, err := api.DescribeFields(context.Background())
	assert.NoError(t, err)
	assert.Len(t, fields, 90)

	t.Run("adds searchable tag to fields", func(t *testing.T) {
		var passed bool
		for _, f := range fields {
			if f.Name == "email" {
				assert.True(t, f.Searchable, "expected email to be searchable")
				passed = true
			}
		}
		assert.True(t, passed, "could not find email field")
	})

	assert.True(t, gock.IsDone())
}
