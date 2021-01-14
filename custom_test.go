package marketo

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

const (
	testHost = "https://marketo.testing"
)

func TestListCustomObjects(t *testing.T) {
	defer gock.Off()

	gock.New(testHost).
		Get("/identity/oauth/token").
		Reply(http.StatusOK).
		JSON(authResponseSuccess)
	gock.New(testHost).
		Get("/rest/v1/customobjects.json").
		Reply(http.StatusOK).
		File("test-fixtures/customobjects.json")

	client, err := NewClient(ClientConfig{
		ID:       clientID,
		Secret:   clientSecret,
		Endpoint: "https://marketo.testing",
		Debug:    true,
	})
	require.NoError(t, err)

	api := NewCustomObjectsAPI(client)
	objects, err := api.List(context.Background())

	assert.NoError(t, err)
	require.Len(t, objects, 1)
	assert.Equal(t, "Test Object", objects[0].DisplayName)

	assert.True(t, gock.IsDone())
}

func TestCustomObjectDescribe(t *testing.T) {
	t.Run("custom object", func(t *testing.T) {
		defer gock.Off()

		gock.New(testHost).
			Get("/identity/oauth/token").
			Reply(http.StatusOK).
			JSON(authResponseSuccess)
		gock.New(testHost).
			Get("/rest/v1/customobjects/testObject_c/describe.json").
			Reply(http.StatusOK).
			File("test-fixtures/testObject_c-describe.json")

		client, err := NewClient(ClientConfig{
			ID:       clientID,
			Secret:   clientSecret,
			Endpoint: "https://marketo.testing",
			Debug:    true,
		})
		require.NoError(t, err)

		api := NewCustomObjectsAPI(client)
		obj, err := api.Describe(context.Background(), "testObject_c")
		assert.NoError(t, err)

		assert.Len(t, obj.Fields, 6)

		t.Run("adds searchable tag to fields", func(t *testing.T) {
			var passed bool
			for _, f := range obj.Fields {
				if f.Name == "email" {
					assert.True(t, f.Searchable, "expected email to be searchable")
					passed = true
				}
			}
			assert.True(t, passed, "could not find email field")
		})

		assert.True(t, gock.IsDone())
	})

	t.Run("unknown object", func(t *testing.T) {
		defer gock.Off()

		gock.New(testHost).
			Get("/identity/oauth/token").
			Reply(http.StatusOK).
			JSON(authResponseSuccess)
		gock.New(testHost).
			Get("/rest/v1/customobjects/unknown/describe.json").
			Reply(http.StatusNotFound)

		client, err := NewClient(ClientConfig{
			ID:       clientID,
			Secret:   clientSecret,
			Endpoint: "https://marketo.testing",
			Debug:    true,
		})
		require.NoError(t, err)

		api := NewCustomObjectsAPI(client)
		objects, err := api.Describe(context.Background(), "unknown")
		assert.Error(t, err)
		assert.Nil(t, objects)

		assert.True(t, gock.IsDone())
	})
}

func TestFitlerCustomObjects(t *testing.T) {
	defer gock.Off()

	gock.New(testHost).
		Get("/identity/oauth/token").
		Reply(http.StatusOK).
		JSON(authResponseSuccess)
	gock.New(testHost).
		Post("/rest/v1/customobjects/testObject_c.json").
		AddMatcher(func(r *http.Request, tr *gock.Request) (bool, error) {
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "email", r.PostForm.Get("filterType"))
			return true, nil
		}).
		Reply(http.StatusOK).
		File("test-fixtures/filterCustomObject.json")

	client, err := NewClient(ClientConfig{
		ID:       clientID,
		Secret:   clientSecret,
		Endpoint: "https://marketo.testing",
		Debug:    true,
	})
	require.NoError(t, err)

	api := NewCustomObjectsAPI(client)
	leads, _, err := api.Filter(
		context.Background(),
		"testObject_c",
		FilterField("email"),
		FilterValues([]string{"nathan@polytomic.com", "ghalib@polytomic.com"}),
	)
	require.NoError(t, err)

	require.Len(t, leads, 1)
	assert.Equal(t, "nathan@polytomic.com", leads[0].Fields["email"])
	assert.True(t, gock.IsDone())
}
