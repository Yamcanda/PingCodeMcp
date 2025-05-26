package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"

	"PingCodeMcp/internal/auth"
	"PingCodeMcp/internal/models"
)

// HandleMakeAuthenticatedRequestTool is a tool that makes an authenticated request
// using the token from the context.
func HandleMakeAuthenticatedRequestTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	message, ok := request.GetArguments()["message"].(string)
	if !ok {
		return nil, fmt.Errorf("missing message")
	}
	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("missing token: %v", err)
	}
	// Now our tool can make a request with the token, irrespective of where it came from.
	resp, err := makeRequest(ctx, message, token)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(fmt.Sprintf("%+v", resp)), nil
}

// makeRequest makes a request to httpbin.org including the auth token in the request
// headers and the message in the query string.
func makeRequest(ctx context.Context, message, token string) (*models.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/anything", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	query := req.URL.Query()
	query.Add("message", message)
	req.URL.RawQuery = query.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var r *models.Response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	return r, nil
}
