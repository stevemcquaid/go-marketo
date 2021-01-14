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

func TestFilterLeads(t *testing.T) {
	defer gock.Off()

	gock.New(testHost).
		Get("/identity/oauth/token").
		Reply(http.StatusOK).
		JSON(authResponseSuccess)
	gock.New(testHost).
		Post("/rest/v1/leads.json").
		AddMatcher(func(r *http.Request, tr *gock.Request) (bool, error) {
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "email", r.PostForm.Get("filterType"))
			return true, nil
		}).
		Reply(http.StatusOK).
		File("test-fixtures/filterLeads.json")

	client, err := NewClient(ClientConfig{
		ID:       clientID,
		Secret:   clientSecret,
		Endpoint: "https://marketo.testing",
		Debug:    true,
	})
	require.NoError(t, err)

	api := NewLeadAPI(client)
	leads, _, err := api.Filter(
		context.Background(),
		FilterField("email"),
		FilterValues([]string{"nathan@polytomic.com", "ghalib@polytomic.com"}),
	)
	require.NoError(t, err)

	assert.Len(t, leads, 2)
	assert.True(t, gock.IsDone())
}

func TestFilterLeads_withFields(t *testing.T) {
	defer gock.Off()

	gock.New(testHost).
		Get("/identity/oauth/token").
		Reply(http.StatusOK).
		JSON(authResponseSuccess)
	gock.New(testHost).
		Post("/rest/v1/leads.json").
		AddMatcher(func(r *http.Request, tr *gock.Request) (bool, error) {
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "email", r.PostForm.Get("filterType"))
			assert.Equal(t, "department,company,firstName", r.PostForm.Get("fields"))
			return true, nil
		}).
		Reply(http.StatusOK).
		File("test-fixtures/filterLeads-fields.json")

	client, err := NewClient(ClientConfig{
		ID:       clientID,
		Secret:   clientSecret,
		Endpoint: "https://marketo.testing",
		Debug:    true,
	})
	require.NoError(t, err)

	api := NewLeadAPI(client)
	leads, _, err := api.Filter(
		context.Background(),
		FilterField("email"),
		FilterValues([]string{"nathan@polytomic.com", "ghalib@polytomic.com"}),
		GetFields("department", "company", "firstName"),
	)
	require.NoError(t, err)

	require.Len(t, leads, 2)
	assert.Equal(t, "Polytomic", leads[0].Fields["company"])

	assert.True(t, gock.IsDone())
}
