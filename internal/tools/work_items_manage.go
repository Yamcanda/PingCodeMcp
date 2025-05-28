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
	workItemsAPIURL = baseUrl + "/v1/project/work_items"
	workItemTimeout = 15 * time.Second
)

func init() {
	toolsFns = append(toolsFns, func() *[]MCPTool {
		return &[]MCPTool{
			NewCreateWorkItemTool(),
		}
	})
}

// CreateWorkItemRequest 创建工作项的请求结构
type CreateWorkItemRequest struct {
	// 工作项负责人的id
	AssigneeID *string `json:"assignee_id,omitempty"`
	// 看板的id。该字段只有项目类型为kanban时有效
	BoardID *string `json:"board_id,omitempty"`
	// 工作项的描述
	Description *string `json:"description,omitempty"`
	// 工作项的截止时间
	EndAt *float64 `json:"end_at,omitempty"`
	// 看板栏的id。该字段只有项目类型为kanban时有效
	EntryID *string `json:"entry_id,omitempty"`
	// 工作项的预估工时
	EstimatedWorkload *float64 `json:"estimated_workload,omitempty"`
	// 工作项的父工作项的id
	ParentID *string `json:"parent_id,omitempty"`
	// 工作项关注人的id列表
	ParticipantIDS []string `json:"participant_ids,omitempty"`
	// 工作项优先级的id
	PriorityID *string `json:"priority_id,omitempty"`
	// 项目的id
	ProjectID string `json:"project_id"`
	// 工作项属性的键值对集合
	Properties *WorkItemProperties `json:"properties,omitempty"`
	// 工作项的剩余工时
	RemainingWorkload *float64 `json:"remaining_workload,omitempty"`
	// 所属迭代id。该字段只有项目类型为scrum时有效
	SprintID *string `json:"sprint_id,omitempty"`
	// 工作项的开始时间
	StartAt *float64 `json:"start_at,omitempty"`
	// 工作项状态的id
	StateID *string `json:"state_id,omitempty"`
	// 工作项的故事点
	StoryPoints *float64 `json:"story_points,omitempty"`
	// 泳道的id。该字段只有项目类型为kanban时有效
	SwimlaneID *string `json:"swimlane_id,omitempty"`
	// 工作项的标题
	Title string `json:"title"`
	// 工作项类型的id
	TypeID string `json:"type_id"`
	// 所属发布的id
	VersionID *string `json:"version_id,omitempty"`
}

// WorkItemProperties 工作项属性的键值对集合
type WorkItemProperties struct {
	// 工作项属性prop_a
	PropA *string `json:"prop_a,omitempty"`
	// 工作项属性prop_b
	PropB *string `json:"prop_b,omitempty"`
}

// CreateWorkItemResponse 创建工作项的响应结构
type CreateWorkItemResponse struct {
	ID         string `json:"id"`
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
	Type       string `json:"type"`
	URL        string `json:"url"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
	IsArchived int    `json:"is_archived"`
	IsDeleted  int    `json:"is_deleted"`

	Assignee *struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		URL         string `json:"url"`
	} `json:"assignee"`

	CreatedBy struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		URL         string `json:"url"`
	} `json:"created_by"`

	UpdatedBy struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		URL         string `json:"url"`
	} `json:"updated_by"`

	Project struct {
		ID         string `json:"id"`
		Identifier string `json:"identifier"`
		Name       string `json:"name"`
		Type       string `json:"type"`
		URL        string `json:"url"`
		IsArchived int    `json:"is_archived"`
		IsDeleted  int    `json:"is_deleted"`
	} `json:"project"`

	State struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Type  string `json:"type"`
		Color string `json:"color"`
		URL   string `json:"url"`
	} `json:"state"`

	Participants []interface{} `json:"participants"`
	Properties   interface{}   `json:"properties"`
	Tags         []interface{} `json:"tags"`
}

// WorkItemPropertiesResponse 响应中的工作项属性
type WorkItemPropertiesResponse struct {
	PropA string `json:"prop_a"`
	PropB string `json:"prop_b"`
}

// CreateWorkItemTool 创建工作项工具
type CreateWorkItemTool struct {
	name        string
	description string
}

// NewCreateWorkItemTool 创建新的工作项创建工具实例
func NewCreateWorkItemTool() MCPTool {
	return &CreateWorkItemTool{
		name:        "create_work_item",
		description: "Create a work item",
	}
}

func (t *CreateWorkItemTool) GetName() string {
	return t.name
}

func (t *CreateWorkItemTool) GetDescription() string {
	return t.description
}

func (t *CreateWorkItemTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription(t.description),
		mcp.WithString("project_id", mcp.Description("项目的id"), mcp.Required()),
		mcp.WithString("title", mcp.Description("工作项的标题"), mcp.Required()),
		mcp.WithString("type_id", mcp.Description("工作项类型的id"), mcp.Required()),
		mcp.WithString("assignee_id", mcp.Description("工作项负责人的id")),
		mcp.WithString("board_id", mcp.Description("看板的id（项目类型为kanban时有效）")),
		mcp.WithString("description", mcp.Description("工作项的描述")),
		mcp.WithNumber("end_at", mcp.Description("工作项的截止时间（时间戳）")),
		mcp.WithString("entry_id", mcp.Description("看板栏的id（项目类型为kanban时有效）")),
		mcp.WithNumber("estimated_workload", mcp.Description("工作项的预估工时")),
		mcp.WithString("parent_id", mcp.Description("工作项的父工作项的id")),
		mcp.WithArray("participant_ids", mcp.Description("工作项关注人的id列表")),
		mcp.WithString("priority_id", mcp.Description("工作项优先级的id")),
		mcp.WithNumber("remaining_workload", mcp.Description("工作项的剩余工时")),
		mcp.WithString("sprint_id", mcp.Description("所属迭代id（项目类型为scrum时有效）")),
		mcp.WithNumber("start_at", mcp.Description("工作项的开始时间（时间戳）")),
		mcp.WithString("state_id", mcp.Description("工作项状态的id")),
		mcp.WithNumber("story_points", mcp.Description("工作项的故事点")),
		mcp.WithString("swimlane_id", mcp.Description("泳道的id（项目类型为kanban时有效）")),
		mcp.WithString("version_id", mcp.Description("所属发布的id")),
		mcp.WithString("prop_a", mcp.Description("工作项属性prop_a")),
		mcp.WithString("prop_b", mcp.Description("工作项属性prop_b")),
	)
}

func (t *CreateWorkItemTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.New()
	defer log.Sync()

	requestLogger := log.With("tool", "create_work_item")
	requestLogger.Info("开始创建工作项")

	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	workItemRequest, err := t.parseCreateWorkItemArguments(request.GetArguments())
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 如果没有指定负责人，默认设置为当前用户
	if workItemRequest.AssigneeID == nil {
		currentUserID, err := t.getCurrentUserID(ctx, token, requestLogger)
		if err != nil {
			requestLogger.With("error", err.Error()).Warn("获取当前用户ID失败，将不设置默认负责人")
		} else {
			workItemRequest.AssigneeID = &currentUserID
			requestLogger.With("assignee_id", currentUserID).Info("未指定负责人，已设置为当前用户")
		}
	}

	respStr, err := t.doCreateWorkItemRequest(ctx, token, workItemRequest, requestLogger)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var resp CreateWorkItemResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("解析响应失败: %v", err)), nil
	}

	result := t.formatCreateWorkItemResponse(resp)
	return mcp.NewToolResultText(result), nil
}

func (t *CreateWorkItemTool) parseCreateWorkItemArguments(args interface{}) (*CreateWorkItemRequest, error) {
	if args == nil {
		return nil, fmt.Errorf("缺少必需的参数")
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("参数格式错误")
	}

	// 检查必需参数
	projectID, ok := argsMap["project_id"].(string)
	if !ok || projectID == "" {
		return nil, fmt.Errorf("project_id 是必需的参数")
	}

	title, ok := argsMap["title"].(string)
	if !ok || title == "" {
		return nil, fmt.Errorf("title 是必需的参数")
	}

	typeID, ok := argsMap["type_id"].(string)
	if !ok || typeID == "" {
		return nil, fmt.Errorf("type_id 是必需的参数")
	}

	req := &CreateWorkItemRequest{
		ProjectID: projectID,
		Title:     title,
		TypeID:    typeID,
	}

	// 解析可选参数
	if assigneeID, exists := argsMap["assignee_id"]; exists {
		if str, ok := assigneeID.(string); ok && str != "" {
			req.AssigneeID = &str
		}
	}

	if boardID, exists := argsMap["board_id"]; exists {
		if str, ok := boardID.(string); ok && str != "" {
			req.BoardID = &str
		}
	}

	if description, exists := argsMap["description"]; exists {
		if str, ok := description.(string); ok && str != "" {
			req.Description = &str
		}
	}

	if endAt, exists := argsMap["end_at"]; exists {
		if num, ok := endAt.(float64); ok {
			req.EndAt = &num
		}
	}

	if entryID, exists := argsMap["entry_id"]; exists {
		if str, ok := entryID.(string); ok && str != "" {
			req.EntryID = &str
		}
	}

	if estimatedWorkload, exists := argsMap["estimated_workload"]; exists {
		if num, ok := estimatedWorkload.(float64); ok {
			req.EstimatedWorkload = &num
		}
	}

	if parentID, exists := argsMap["parent_id"]; exists {
		if str, ok := parentID.(string); ok && str != "" {
			req.ParentID = &str
		}
	}

	if participantIDs, exists := argsMap["participant_ids"]; exists {
		if arr, ok := participantIDs.([]interface{}); ok {
			var ids []string
			for _, id := range arr {
				if str, ok := id.(string); ok && str != "" {
					ids = append(ids, str)
				}
			}
			req.ParticipantIDS = ids
		}
	}

	if priorityID, exists := argsMap["priority_id"]; exists {
		if str, ok := priorityID.(string); ok && str != "" {
			req.PriorityID = &str
		}
	}

	if remainingWorkload, exists := argsMap["remaining_workload"]; exists {
		if num, ok := remainingWorkload.(float64); ok {
			req.RemainingWorkload = &num
		}
	}

	if sprintID, exists := argsMap["sprint_id"]; exists {
		if str, ok := sprintID.(string); ok && str != "" {
			req.SprintID = &str
		}
	}

	if startAt, exists := argsMap["start_at"]; exists {
		if num, ok := startAt.(float64); ok {
			req.StartAt = &num
		}
	}

	if stateID, exists := argsMap["state_id"]; exists {
		if str, ok := stateID.(string); ok && str != "" {
			req.StateID = &str
		}
	}

	if storyPoints, exists := argsMap["story_points"]; exists {
		if num, ok := storyPoints.(float64); ok {
			req.StoryPoints = &num
		}
	}

	if swimlaneID, exists := argsMap["swimlane_id"]; exists {
		if str, ok := swimlaneID.(string); ok && str != "" {
			req.SwimlaneID = &str
		}
	}

	if versionID, exists := argsMap["version_id"]; exists {
		if str, ok := versionID.(string); ok && str != "" {
			req.VersionID = &str
		}
	}

	// 解析工作项属性
	if propA, exists := argsMap["prop_a"]; exists {
		if propB, existsB := argsMap["prop_b"]; existsB {
			if strA, okA := propA.(string); okA || propA == nil {
				if strB, okB := propB.(string); okB || propB == nil {
					req.Properties = &WorkItemProperties{}
					if okA && strA != "" {
						req.Properties.PropA = &strA
					}
					if okB && strB != "" {
						req.Properties.PropB = &strB
					}
				}
			}
		} else if str, ok := propA.(string); ok && str != "" {
			req.Properties = &WorkItemProperties{PropA: &str}
		}
	} else if propB, exists := argsMap["prop_b"]; exists {
		if str, ok := propB.(string); ok && str != "" {
			req.Properties = &WorkItemProperties{PropB: &str}
		}
	}

	return req, nil
}

func (t *CreateWorkItemTool) doCreateWorkItemRequest(ctx context.Context, token string, workItemRequest *CreateWorkItemRequest, log *logger.Logger) (string, error) {
	headers := map[string]string{
		"Authorization": token,
		"Content-Type":  "application/json",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, workItemTimeout)
	defer cancel()

	// 序列化请求体
	requestBody, err := json.Marshal(workItemRequest)
	if err != nil {
		log.With("error", err.Error()).Error("序列化请求体失败")
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	body, _, err := utils.DoPostJSON(timeoutCtx, workItemsAPIURL, headers, requestBody)
	if err != nil {
		return "", fmt.Errorf("创建工作项请求失败: %v", err)
	}

	log.With("url", workItemsAPIURL, "response_size_bytes", len(body), "success", true).Debug("创建工作项API请求成功")
	return string(body), nil
}

func (t *CreateWorkItemTool) formatCreateWorkItemResponse(resp CreateWorkItemResponse) string {
	var result strings.Builder
	result.WriteString("✅ 工作项创建成功！\n\n")
	result.WriteString(fmt.Sprintf("📋 标题: %s\n", resp.Title))
	result.WriteString(fmt.Sprintf("🆔 工作项ID: %s\n", resp.ID))
	result.WriteString(fmt.Sprintf("🏷️ 标识符: %s\n", resp.Identifier))
	result.WriteString(fmt.Sprintf("📂 项目: %s (%s)\n", resp.Project.Name, resp.Project.Identifier))
	result.WriteString(fmt.Sprintf("🔗 项目ID: %s\n", resp.Project.ID))
	result.WriteString(fmt.Sprintf("📊 类型: %s\n", resp.Type))

	if resp.Assignee != nil {
		result.WriteString(fmt.Sprintf("👤 负责人: %s (%s)\n", resp.Assignee.DisplayName, resp.Assignee.Name))
		result.WriteString(fmt.Sprintf("🔗 负责人ID: %s\n", resp.Assignee.ID))
	}

	result.WriteString(fmt.Sprintf("📈 状态: %s (%s)\n", resp.State.Name, resp.State.Type))
	result.WriteString(fmt.Sprintf("🎨 状态颜色: %s\n", resp.State.Color))

	result.WriteString(fmt.Sprintf("👤 创建人: %s (%s)\n", resp.CreatedBy.DisplayName, resp.CreatedBy.Name))

	createdTime := time.Unix(resp.CreatedAt, 0)
	result.WriteString(fmt.Sprintf("🕐 创建时间: %s\n", createdTime.Format("2006-01-02 15:04:05")))

	updatedTime := time.Unix(resp.UpdatedAt, 0)
	result.WriteString(fmt.Sprintf("🕐 更新时间: %s\n", updatedTime.Format("2006-01-02 15:04:05")))

	if len(resp.Participants) > 0 {
		result.WriteString(fmt.Sprintf("👥 关注人数: %d\n", len(resp.Participants)))
	}

	result.WriteString(fmt.Sprintf("🔗 工作项链接: %s\n", resp.URL))

	return result.String()
}

// getCurrentUserID 获取当前用户ID
func (t *CreateWorkItemTool) getCurrentUserID(ctx context.Context, token string, log *logger.Logger) (string, error) {
	headers := map[string]string{
		"Authorization": token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	log.Debug("获取当前用户信息")

	timeoutCtx, cancel := context.WithTimeout(ctx, userInfoTimeout)
	defer cancel()

	// 使用正确的用户信息API地址
	body, _, err := utils.DoGet(timeoutCtx, userInfoAPIURL, headers, nil)
	if err != nil {
		log.With("error", err.Error()).Error("获取用户信息API请求失败")
		return "", fmt.Errorf("获取用户信息失败: %v", err)
	}

	var userInfo userInfoResponse
	if err := json.Unmarshal(body, &userInfo); err != nil {
		log.With("error", err.Error()).Error("解析用户信息响应失败")
		return "", fmt.Errorf("解析用户信息失败: %v", err)
	}

	log.With("user_id", userInfo.ID, "display_name", userInfo.DisplayName).Debug("成功获取当前用户信息")
	return userInfo.ID, nil
}
