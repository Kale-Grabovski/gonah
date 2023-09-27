package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Kale-Grabovski/gonah/src/domain"
)

func TestUser(t *testing.T) {
	client := httpClient{}

	// CREATE
	record := domain.User{
		Login: "Alice",
	}
	httpBody, err := json.Marshal(record)
	require.NoError(t, err)
	resp, respBody, err := client.sendJsonReq(http.MethodPost, "http://localhost:8877/api/v1/users", httpBody)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	respUser := domain.User{}
	err = json.Unmarshal(respBody, &respUser)
	require.NoError(t, err)
	require.NotEqual(t, 0, respUser.Id)

	// LIST
	resp, respBody, err = client.sendJsonReq(http.MethodGet, "http://localhost:8877/api/v1/users", nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var users []domain.User
	err = json.Unmarshal(respBody, &users)
	require.Equal(t, 1, len(users))

	// READ
	resp, respBody, err = client.sendJsonReq(http.MethodGet, fmt.Sprintf("http://localhost:8877/api/v1/users/%d", respUser.Id), []byte{})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.Unmarshal(respBody, &record)
	require.NoError(t, err)
	require.Equal(t, respUser.Id, record.Id)
	require.Equal(t, "Alice", record.Login)
}
