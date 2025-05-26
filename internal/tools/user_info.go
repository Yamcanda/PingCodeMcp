package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"PingCodeMcp/internal/auth"
	"PingCodeMcp/internal/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	userInfoURL = "https://open.pingcode.com/v1/myself"
)

type userInfoResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
	Phone      string `json:"phone"`
	Status     string `json:"status"`
	Role       string `json:"role"`
	Department string `json:"department"`
}

// HandleUserInfoTool MCP工具：获取个人基本信息
func HandleUserInfoTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 从context中获取authkey
	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("missing or empty authorization token: %v", err)
	}

	// 确保token格式正确
	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	respStr, err := doUserInfoRequest(ctx, token)
	if err != nil {

		return nil, fmt.Errorf("failed to get user info: %v", err)
	}

	var resp userInfoResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// 格式化返回结果
	result := fmt.Sprintf(`用户基本信息:
		ID: %s
		姓名: %s
		邮箱: %s
		电话: %s
		状态: %s
		角色: %s
		部门: %s
		头像: %s`,
		resp.ID,
		resp.Name,
		resp.Email,
		resp.Phone,
		resp.Status,
		resp.Role,
		resp.Department,
		resp.Avatar,
	)

	return mcp.NewToolResultText(result), nil
}

func doUserInfoRequest(ctx context.Context, token string) (string, error) {
	headers := map[string]string{
		"Authorization": token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	body, _, err := utils.DoGet(ctx, userInfoURL, headers, nil)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
