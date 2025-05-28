package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"PingCodeMcp/internal/auth"
	"PingCodeMcp/internal/utils"
	"PingCodeMcp/pkg/logger"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	projectListAPIURL  = baseUrl + "/v1/project/projects"
	projectListTimeout = 15 * time.Second
)

func init() {
	toolsFns = append(toolsFns, func() *[]MCPTool {
		return &[]MCPTool{
			NewProjectListTool(),
		}
	})
}

type Project struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Identifier  string `json:"identifier"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	State       string `json:"state"`
	Description string `json:"description"`
}

type ProjectListResponse struct {
	PageIndex int       `json:"page_index"`
	PageSize  int       `json:"page_size"`
	Total     int       `json:"total"`
	Values    []Project `json:"values"`
}

type ProjectListTool struct {
	name        string
	description string
}

func NewProjectListTool() MCPTool {
	return &ProjectListTool{
		name:        "get_project_list",
		description: "Get the list of PingCode projects",
	}
}

func (t *ProjectListTool) GetName() string {
	return t.name
}

func (t *ProjectListTool) GetDescription() string {
	return t.description
}

func (t *ProjectListTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription(t.description),
		mcp.WithNumber("page_index", mcp.Description("页码，从1开始")),
		mcp.WithNumber("page_size", mcp.Description("每页数量，最大100")),
	)
}

func (t *ProjectListTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.New()
	defer log.Sync()

	requestLogger := log.With("tool", "get_project_list")
	requestLogger.Info("开始获取项目列表")

	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	params := t.parseArguments(request.GetArguments())
	respStr, err := t.doRequest(ctx, token, params, requestLogger)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var resp ProjectListResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := t.formatResponse(resp)
	return mcp.NewToolResultText(result), nil
}

func (t *ProjectListTool) parseArguments(args interface{}) map[string]string {
	params := map[string]string{
		"page_index": "1",
		"page_size":  "20",
	}

	if args == nil {
		return params
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return params
	}

	if page, exists := argsMap["page_index"]; exists {
		if pageFloat, ok := page.(float64); ok {
			params["page_index"] = fmt.Sprintf("%.0f", pageFloat)
		}
	}

	if pageSize, exists := argsMap["page_size"]; exists {
		if pageSizeFloat, ok := pageSize.(float64); ok {
			params["page_size"] = fmt.Sprintf("%.0f", pageSizeFloat)
		}
	}

	return params
}

func (t *ProjectListTool) doRequest(ctx context.Context, token string, params map[string]string, log *logger.Logger) (string, error) {
	headers := map[string]string{
		"Authorization": token,
		"Content-Type":  "application/json",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, projectListTimeout)
	defer cancel()

	body, _, err := utils.DoGet(timeoutCtx, projectListAPIURL, headers, params)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (t *ProjectListTool) formatResponse(resp ProjectListResponse) string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("项目列表 (第%d页，共%d个项目):\n\n", resp.PageIndex, resp.Total))

	for i, project := range resp.Values {
		result.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, project.Name, project.Identifier))
		result.WriteString(fmt.Sprintf("   类型: %s\n", project.Type))
		result.WriteString(fmt.Sprintf("   状态: %s\n", project.State))
		if project.Description != "" {
			result.WriteString(fmt.Sprintf("   描述: %s\n", project.Description))
		}
		result.WriteString(fmt.Sprintf("   项目ID: %s\n\n", project.ID))
	}

	return result.String()
}
