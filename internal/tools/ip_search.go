package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"PingCodeMcp/internal/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ipSearchURL = "https://whois.suyun.store/query"
	unknownAddr = "XX XX"
)

type ipSearchResponse struct {
	Status string `json:"status"`
	Data   struct {
		Country string `json:"country"`
		City    string `json:"city"`
	} `json:"data"`
}

// HandleIpSearchTool MCP工具：根据IP查询地理位置
func HandleIpSearchTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ip, ok := request.GetArguments()["ip"].(string)
	if !ok || strings.TrimSpace(ip) == "" {
		return nil, fmt.Errorf("missing or empty ip")
	}
	address := unknownAddr
	respStr, err := doIpSearch(ctx, ip)
	if err != nil {
		return mcp.NewToolResultText(address), nil
	}
	var resp ipSearchResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		return mcp.NewToolResultText(address), nil
	}
	if resp.Status == "Success" {
		region := resp.Data.Country
		city := resp.Data.City
		return mcp.NewToolResultText(fmt.Sprintf("%s %s", region, city)), nil
	}
	return mcp.NewToolResultText(address), nil
}

func doIpSearch(ctx context.Context, ip string) (string, error) {
	query := map[string]string{"ip": ip}
	headers := map[string]string{"Accept": "application/json"}
	body, _, err := utils.DoGet(ctx, ipSearchURL, headers, query)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
