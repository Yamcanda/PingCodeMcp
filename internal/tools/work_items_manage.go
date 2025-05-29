package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"PingCodeMcp/internal/auth"
	"PingCodeMcp/internal/models"
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
			NewUpdateWorkItemTool(),
			NewListWorkItemsTool(),
			NewDeleteWorkItemTool(),
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

	var errorRes models.ErrorResponse
	if err := json.Unmarshal([]byte(respStr), &errorRes); err == nil && errorRes.Code != "" {
		requestLogger.With("success", false, "error", errorRes.Message).Error("创建工作项API调用失败")
		return mcp.NewToolResultError(errorRes.Message), nil
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

	timeoutCtx, cancel := context.WithTimeout(ctx, workItemTimeout)
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

// UpdateWorkItemRequest 更新工作项的请求结构
type UpdateWorkItemRequest struct {
	// 项目的id
	ProjectID *string `json:"project_id,omitempty"`
	// 工作项类型的id
	TypeID *string `json:"type_id,omitempty"`
	// 工作项的标题
	Title *string `json:"title,omitempty"`
	// 工作项的描述
	Description *string `json:"description,omitempty"`
	// 工作项的开始时间
	StartAt *float64 `json:"start_at,omitempty"`
	// 工作项的截止时间
	EndAt *float64 `json:"end_at,omitempty"`
	// 工作项状态的id
	StateID *string `json:"state_id,omitempty"`
	// 工作项的父工作项的id
	ParentID *string `json:"parent_id,omitempty"`
	// 所属迭代id
	SprintID *string `json:"sprint_id,omitempty"`
	// 所属发布的id
	VersionID *string `json:"version_id,omitempty"`
	// 看板的id
	BoardID *string `json:"board_id,omitempty"`
	// 看板栏的id
	EntryID *string `json:"entry_id,omitempty"`
	// 泳道的id
	SwimlaneID *string `json:"swimlane_id,omitempty"`
	// 工作项优先级的id
	PriorityID *string `json:"priority_id,omitempty"`
	// 工作项负责人的id
	AssigneeID *string `json:"assignee_id,omitempty"`
	// 工作项关注人的id列表
	ParticipantIDS []string `json:"participant_ids,omitempty"`
	// 工作项的故事点
	StoryPoints *float64 `json:"story_points,omitempty"`
	// 工作项的预估工时
	EstimatedWorkload *float64 `json:"estimated_workload,omitempty"`
	// 工作项的剩余工时
	RemainingWorkload *float64 `json:"remaining_workload,omitempty"`
	// 工作项属性的键值对集合
	Properties *WorkItemProperties `json:"properties,omitempty"`
}

// UpdateWorkItemResponse 更新工作项的响应结构
type UpdateWorkItemResponse struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
	Type       string `json:"type"`
	StartAt    int64  `json:"start_at"`
	EndAt      int64  `json:"end_at"`
	ParentID   string `json:"parent_id"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
	IsArchived int    `json:"is_archived"`
	IsDeleted  int    `json:"is_deleted"`

	Project struct {
		ID         string `json:"id"`
		URL        string `json:"url"`
		Identifier string `json:"identifier"`
		Name       string `json:"name"`
		Type       string `json:"type"`
	} `json:"project"`

	Parent *struct {
		ID         string `json:"id"`
		URL        string `json:"url"`
		Identifier string `json:"identifier"`
		Title      string `json:"title"`
		Type       string `json:"type"`
		StartAt    int64  `json:"start_at"`
		EndAt      int64  `json:"end_at"`
		ParentID   string `json:"parent_id"`
	} `json:"parent"`

	Assignee *struct {
		ID          string `json:"id"`
		URL         string `json:"url"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Avatar      string `json:"avatar"`
	} `json:"assignee"`

	Version *struct {
		ID      string `json:"id"`
		URL     string `json:"url"`
		Name    string `json:"name"`
		StartAt int64  `json:"start_at"`
		EndAt   int64  `json:"end_at"`
		Stage   struct {
			ID    string `json:"id"`
			URL   string `json:"url"`
			Name  string `json:"name"`
			Type  string `json:"type"`
			Color string `json:"color"`
		} `json:"stage"`
	} `json:"version"`

	Sprint *struct {
		ID      string `json:"id"`
		URL     string `json:"url"`
		Name    string `json:"name"`
		StartAt int64  `json:"start_at"`
		EndAt   int64  `json:"end_at"`
		Status  string `json:"status"`
	} `json:"sprint"`

	State struct {
		ID    string `json:"id"`
		URL   string `json:"url"`
		Name  string `json:"name"`
		Type  string `json:"type"`
		Color string `json:"color"`
	} `json:"state"`

	Priority *struct {
		ID   string `json:"id"`
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"priority"`

	StoryPoints  int64       `json:"story_points"`
	Description  string      `json:"description"`
	CompletedAt  int64       `json:"completed_at"`
	Properties   interface{} `json:"properties"`
	Tags         []string    `json:"tags"`
	Participants []struct {
		ID   string `json:"id"`
		URL  string `json:"url"`
		Type string `json:"type"`
		User struct {
			ID          string `json:"id"`
			URL         string `json:"url"`
			Name        string `json:"name"`
			DisplayName string `json:"display_name"`
			Avatar      string `json:"avatar"`
		} `json:"user"`
	} `json:"participants"`

	CreatedBy struct {
		ID          string `json:"id"`
		URL         string `json:"url"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Avatar      string `json:"avatar"`
	} `json:"created_by"`

	UpdatedBy struct {
		ID          string `json:"id"`
		URL         string `json:"url"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Avatar      string `json:"avatar"`
	} `json:"updated_by"`
}

// UpdateWorkItemTool 更新工作项工具
type UpdateWorkItemTool struct {
	name        string
	description string
}

// NewUpdateWorkItemTool 创建新的工作项更新工具实例
func NewUpdateWorkItemTool() MCPTool {
	return &UpdateWorkItemTool{
		name:        "update_work_item",
		description: "Update a work item",
	}
}

func (t *UpdateWorkItemTool) GetName() string {
	return t.name
}

func (t *UpdateWorkItemTool) GetDescription() string {
	return t.description
}

func (t *UpdateWorkItemTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription(t.description),
		mcp.WithString("work_item_id", mcp.Description("工作项的id"), mcp.Required()),
		mcp.WithString("project_id", mcp.Description("项目的id")),
		mcp.WithString("title", mcp.Description("工作项的标题")),
		mcp.WithString("type_id", mcp.Description("工作项类型的id")),
		mcp.WithString("description", mcp.Description("工作项的描述")),
		mcp.WithNumber("start_at", mcp.Description("工作项的开始时间（时间戳）")),
		mcp.WithNumber("end_at", mcp.Description("工作项的截止时间（时间戳）")),
		mcp.WithString("state_id", mcp.Description("工作项状态的id")),
		mcp.WithString("parent_id", mcp.Description("工作项的父工作项的id")),
		mcp.WithString("sprint_id", mcp.Description("所属迭代id")),
		mcp.WithString("version_id", mcp.Description("所属发布的id")),
		mcp.WithString("board_id", mcp.Description("看板的id")),
		mcp.WithString("entry_id", mcp.Description("看板栏的id")),
		mcp.WithString("swimlane_id", mcp.Description("泳道的id")),
		mcp.WithString("priority_id", mcp.Description("工作项优先级的id")),
		mcp.WithString("assignee_id", mcp.Description("工作项负责人的id")),
		mcp.WithArray("participant_ids", mcp.Description("工作项关注人的id列表")),
		mcp.WithNumber("story_points", mcp.Description("工作项的故事点")),
		mcp.WithNumber("estimated_workload", mcp.Description("工作项的预估工时")),
		mcp.WithNumber("remaining_workload", mcp.Description("工作项的剩余工时")),
		mcp.WithString("prop_a", mcp.Description("工作项属性prop_a")),
		mcp.WithString("prop_b", mcp.Description("工作项属性prop_b")),
	)
}

func (t *UpdateWorkItemTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.New()
	defer log.Sync()

	requestLogger := log.With("tool", "update_work_item")
	requestLogger.Info("开始更新工作项")

	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	workItemID, updateRequest, err := t.parseUpdateWorkItemArguments(request.GetArguments())
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	respStr, err := t.doUpdateWorkItemRequest(ctx, token, workItemID, updateRequest, requestLogger)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var errorRes models.ErrorResponse
	if err := json.Unmarshal([]byte(respStr), &errorRes); err == nil && errorRes.Code != "" {
		requestLogger.With("success", false, "error", errorRes.Message).Error("更新工作项API调用失败")
		return mcp.NewToolResultError(errorRes.Message), nil
	}

	var resp UpdateWorkItemResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("解析响应失败: %v", err)), nil
	}

	result := t.formatUpdateWorkItemResponse(resp)
	return mcp.NewToolResultText(result), nil
}

func (t *UpdateWorkItemTool) parseUpdateWorkItemArguments(args interface{}) (string, *UpdateWorkItemRequest, error) {
	if args == nil {
		return "", nil, fmt.Errorf("缺少必需的参数")
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return "", nil, fmt.Errorf("参数格式错误")
	}

	// 检查必需参数
	workItemID, ok := argsMap["work_item_id"].(string)
	if !ok || workItemID == "" {
		return "", nil, fmt.Errorf("work_item_id 是必需的参数")
	}

	req := &UpdateWorkItemRequest{}

	// 解析可选参数
	if projectID, exists := argsMap["project_id"]; exists {
		if str, ok := projectID.(string); ok && str != "" {
			req.ProjectID = &str
		}
	}

	if title, exists := argsMap["title"]; exists {
		if str, ok := title.(string); ok && str != "" {
			req.Title = &str
		}
	}

	if typeID, exists := argsMap["type_id"]; exists {
		if str, ok := typeID.(string); ok && str != "" {
			req.TypeID = &str
		}
	}

	if description, exists := argsMap["description"]; exists {
		if str, ok := description.(string); ok && str != "" {
			req.Description = &str
		}
	}

	if startAt, exists := argsMap["start_at"]; exists {
		if num, ok := startAt.(float64); ok {
			req.StartAt = &num
		}
	}

	if endAt, exists := argsMap["end_at"]; exists {
		if num, ok := endAt.(float64); ok {
			req.EndAt = &num
		}
	}

	if stateID, exists := argsMap["state_id"]; exists {
		if str, ok := stateID.(string); ok && str != "" {
			req.StateID = &str
		}
	}

	if parentID, exists := argsMap["parent_id"]; exists {
		if str, ok := parentID.(string); ok && str != "" {
			req.ParentID = &str
		}
	}

	if sprintID, exists := argsMap["sprint_id"]; exists {
		if str, ok := sprintID.(string); ok && str != "" {
			req.SprintID = &str
		}
	}

	if versionID, exists := argsMap["version_id"]; exists {
		if str, ok := versionID.(string); ok && str != "" {
			req.VersionID = &str
		}
	}

	if boardID, exists := argsMap["board_id"]; exists {
		if str, ok := boardID.(string); ok && str != "" {
			req.BoardID = &str
		}
	}

	if entryID, exists := argsMap["entry_id"]; exists {
		if str, ok := entryID.(string); ok && str != "" {
			req.EntryID = &str
		}
	}

	if swimlaneID, exists := argsMap["swimlane_id"]; exists {
		if str, ok := swimlaneID.(string); ok && str != "" {
			req.SwimlaneID = &str
		}
	}

	if priorityID, exists := argsMap["priority_id"]; exists {
		if str, ok := priorityID.(string); ok && str != "" {
			req.PriorityID = &str
		}
	}

	if assigneeID, exists := argsMap["assignee_id"]; exists {
		if str, ok := assigneeID.(string); ok && str != "" {
			req.AssigneeID = &str
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

	if storyPoints, exists := argsMap["story_points"]; exists {
		if num, ok := storyPoints.(float64); ok {
			req.StoryPoints = &num
		}
	}

	if estimatedWorkload, exists := argsMap["estimated_workload"]; exists {
		if num, ok := estimatedWorkload.(float64); ok {
			req.EstimatedWorkload = &num
		}
	}

	if remainingWorkload, exists := argsMap["remaining_workload"]; exists {
		if num, ok := remainingWorkload.(float64); ok {
			req.RemainingWorkload = &num
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

	return workItemID, req, nil
}

func (t *UpdateWorkItemTool) doUpdateWorkItemRequest(ctx context.Context, token string, workItemID string, updateRequest *UpdateWorkItemRequest, log *logger.Logger) (string, error) {
	headers := map[string]string{
		"Authorization": token,
		"Content-Type":  "application/json",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, workItemTimeout)
	defer cancel()

	// 构建API URL
	updateURL := fmt.Sprintf("%s/%s", workItemsAPIURL, workItemID)

	// 序列化请求体
	requestBody, err := json.Marshal(updateRequest)
	if err != nil {
		log.With("error", err.Error()).Error("序列化请求体失败")
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	body, _, err := utils.DoPatch(timeoutCtx, updateURL, headers, requestBody)
	if err != nil {
		return "", fmt.Errorf("更新工作项请求失败: %v", err)
	}

	log.With("url", updateURL, "response_size_bytes", len(body), "success", true).Debug("更新工作项API请求成功")
	return string(body), nil
}

func (t *UpdateWorkItemTool) formatUpdateWorkItemResponse(resp UpdateWorkItemResponse) string {
	var result strings.Builder
	result.WriteString("✅ 工作项更新成功！\n\n")
	result.WriteString(fmt.Sprintf("📋 标题: %s\n", resp.Title))
	result.WriteString(fmt.Sprintf("🆔 工作项ID: %s\n", resp.ID))
	result.WriteString(fmt.Sprintf("🏷️ 标识符: %s\n", resp.Identifier))
	result.WriteString(fmt.Sprintf("📂 项目: %s (%s)\n", resp.Project.Name, resp.Project.Identifier))
	result.WriteString(fmt.Sprintf("🔗 项目ID: %s\n", resp.Project.ID))
	result.WriteString(fmt.Sprintf("📊 类型: %s\n", resp.Type))

	if resp.Description != "" {
		result.WriteString(fmt.Sprintf("📝 描述: %s\n", resp.Description))
	}

	if resp.Assignee != nil {
		result.WriteString(fmt.Sprintf("👤 负责人: %s (%s)\n", resp.Assignee.DisplayName, resp.Assignee.Name))
		result.WriteString(fmt.Sprintf("🔗 负责人ID: %s\n", resp.Assignee.ID))
	}

	result.WriteString(fmt.Sprintf("📈 状态: %s (%s)\n", resp.State.Name, resp.State.Type))
	result.WriteString(fmt.Sprintf("🎨 状态颜色: %s\n", resp.State.Color))

	if resp.Priority != nil {
		result.WriteString(fmt.Sprintf("⚡ 优先级: %s\n", resp.Priority.Name))
	}

	if resp.Sprint != nil {
		result.WriteString(fmt.Sprintf("🏃 迭代: %s (%s)\n", resp.Sprint.Name, resp.Sprint.Status))
	}

	if resp.Version != nil {
		result.WriteString(fmt.Sprintf("📦 版本: %s (%s)\n", resp.Version.Name, resp.Version.Stage.Name))
	}

	if resp.Parent != nil {
		result.WriteString(fmt.Sprintf("👨‍👩‍👧‍👦 父工作项: %s (%s)\n", resp.Parent.Title, resp.Parent.Identifier))
	}

	if resp.StoryPoints > 0 {
		result.WriteString(fmt.Sprintf("📈 故事点: %d\n", resp.StoryPoints))
	}

	if resp.StartAt > 0 {
		startTime := time.Unix(resp.StartAt, 0)
		result.WriteString(fmt.Sprintf("🚀 开始时间: %s\n", startTime.Format("2006-01-02 15:04:05")))
	}

	if resp.EndAt > 0 {
		endTime := time.Unix(resp.EndAt, 0)
		result.WriteString(fmt.Sprintf("🏁 截止时间: %s\n", endTime.Format("2006-01-02 15:04:05")))
	}

	if resp.CompletedAt > 0 {
		completedTime := time.Unix(resp.CompletedAt, 0)
		result.WriteString(fmt.Sprintf("✅ 完成时间: %s\n", completedTime.Format("2006-01-02 15:04:05")))
	}

	if len(resp.Participants) > 0 {
		result.WriteString(fmt.Sprintf("👥 关注人数: %d\n", len(resp.Participants)))
		for i, participant := range resp.Participants {
			if i < 3 { // 只显示前3个关注人
				result.WriteString(fmt.Sprintf("  • %s (%s)\n", participant.User.DisplayName, participant.User.Name))
			}
		}
		if len(resp.Participants) > 3 {
			result.WriteString(fmt.Sprintf("  • ... 还有 %d 人\n", len(resp.Participants)-3))
		}
	}

	if len(resp.Tags) > 0 {
		result.WriteString(fmt.Sprintf("🏷️ 标签: %s\n", strings.Join(resp.Tags, ", ")))
	}

	result.WriteString(fmt.Sprintf("👤 创建人: %s (%s)\n", resp.CreatedBy.DisplayName, resp.CreatedBy.Name))
	result.WriteString(fmt.Sprintf("👤 更新人: %s (%s)\n", resp.UpdatedBy.DisplayName, resp.UpdatedBy.Name))

	createdTime := time.Unix(resp.CreatedAt, 0)
	result.WriteString(fmt.Sprintf("🕐 创建时间: %s\n", createdTime.Format("2006-01-02 15:04:05")))

	updatedTime := time.Unix(resp.UpdatedAt, 0)
	result.WriteString(fmt.Sprintf("🕐 更新时间: %s\n", updatedTime.Format("2006-01-02 15:04:05")))

	result.WriteString(fmt.Sprintf("🔗 工作项链接: %s\n", resp.URL))

	return result.String()
}

// ListWorkItemsResponse 获取工作项列表的响应结构
type ListWorkItemsResponse struct {
	PageSize  int                `json:"page_size"`
	PageIndex int                `json:"page_index"`
	Total     int                `json:"total"`
	Values    []WorkItemListItem `json:"values"`
}

// WorkItemListItem 工作项列表项
type WorkItemListItem struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Identifier  string `json:"identifier"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	StartAt     int64  `json:"start_at"`
	EndAt       int64  `json:"end_at"`
	ParentID    string `json:"parent_id"`
	Description string `json:"description"`
	CompletedAt int64  `json:"completed_at"`
	StoryPoints int64  `json:"story_points"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	IsArchived  int    `json:"is_archived"`
	IsDeleted   int    `json:"is_deleted"`

	Project struct {
		ID         string `json:"id"`
		URL        string `json:"url"`
		Identifier string `json:"identifier"`
		Name       string `json:"name"`
		Type       string `json:"type"`
	} `json:"project"`

	Assignee *struct {
		ID          string `json:"id"`
		URL         string `json:"url"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Avatar      string `json:"avatar"`
	} `json:"assignee"`

	State struct {
		ID    string `json:"id"`
		URL   string `json:"url"`
		Name  string `json:"name"`
		Type  string `json:"type"`
		Color string `json:"color"`
	} `json:"state"`

	Priority *struct {
		ID   string `json:"id"`
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"priority"`

	Parent *struct {
		ID         string `json:"id"`
		URL        string `json:"url"`
		Identifier string `json:"identifier"`
		Title      string `json:"title"`
		Type       string `json:"type"`
		StartAt    int64  `json:"start_at"`
		EndAt      int64  `json:"end_at"`
		ParentID   string `json:"parent_id"`
	} `json:"parent"`

	Version *struct {
		ID      string `json:"id"`
		URL     string `json:"url"`
		Name    string `json:"name"`
		StartAt int64  `json:"start_at"`
		EndAt   int64  `json:"end_at"`
		Stage   struct {
			ID    string `json:"id"`
			URL   string `json:"url"`
			Name  string `json:"name"`
			Type  string `json:"type"`
			Color string `json:"color"`
		} `json:"stage"`
	} `json:"version"`

	Sprint *struct {
		ID      string `json:"id"`
		URL     string `json:"url"`
		Name    string `json:"name"`
		StartAt int64  `json:"start_at"`
		EndAt   int64  `json:"end_at"`
		Status  string `json:"status"`
	} `json:"sprint"`

	Phase *struct {
		ID         string `json:"id"`
		URL        string `json:"url"`
		Title      string `json:"title"`
		Identifier string `json:"identifier"`
	} `json:"phase"`

	Properties interface{} `json:"properties"`

	Tags []struct {
		ID   string `json:"id"`
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"tags"`

	Participants []struct {
		ID   string `json:"id"`
		URL  string `json:"url"`
		Type string `json:"type"`
		User struct {
			ID          string `json:"id"`
			URL         string `json:"url"`
			Name        string `json:"name"`
			DisplayName string `json:"display_name"`
			Avatar      string `json:"avatar"`
		} `json:"user"`
	} `json:"participants"`

	CreatedBy struct {
		ID          string `json:"id"`
		URL         string `json:"url"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Avatar      string `json:"avatar"`
	} `json:"created_by"`

	UpdatedBy struct {
		ID          string `json:"id"`
		URL         string `json:"url"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Avatar      string `json:"avatar"`
	} `json:"updated_by"`
}

// ListWorkItemsTool 获取工作项列表工具
type ListWorkItemsTool struct {
	name        string
	description string
}

// NewListWorkItemsTool 创建新的工作项列表获取工具实例
func NewListWorkItemsTool() MCPTool {
	return &ListWorkItemsTool{
		name:        "list_work_items",
		description: "Get work items list",
	}
}

func (t *ListWorkItemsTool) GetName() string {
	return t.name
}

func (t *ListWorkItemsTool) GetDescription() string {
	return t.description
}

func (t *ListWorkItemsTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription(t.description),
		mcp.WithString("identifier", mcp.Description("工作项编号")),
		mcp.WithString("project_ids", mcp.Description("项目的id，使用','分割，最多只能20个")),
		mcp.WithString("type_ids", mcp.Description("工作项类型的id，使用','分割，最多只能20个")),
		mcp.WithString("parent_ids", mcp.Description("父工作项的id，使用','分割，最多只能20个")),
		mcp.WithString("assignee_ids", mcp.Description("工作项负责人的id，使用','分割，最多只能20个")),
		mcp.WithString("state_ids", mcp.Description("工作项状态的id，使用','分割，最多只能20个")),
		mcp.WithString("start_between", mcp.Description("开始时间介于的时间范围，通过','分割起始时间")),
		mcp.WithString("end_between", mcp.Description("结束时间介于的时间范围，通过','分割起始时间")),
		mcp.WithString("priority_ids", mcp.Description("工作项优先级的id，使用','分割，最多只能20个")),
		mcp.WithString("bug_type_ids", mcp.Description("缺陷类别的id，使用','分割，最多只能20个")),
		mcp.WithString("sprint_ids", mcp.Description("迭代的id，使用','分割，最多只能20个")),
		mcp.WithString("board_ids", mcp.Description("看板的id，使用','分割，最多只能20个")),
		mcp.WithString("entry_ids", mcp.Description("看板栏的id，使用','分割，最多只能20个")),
		mcp.WithString("tag_ids", mcp.Description("工作项标签的id，使用','分割，最多只能20个")),
		mcp.WithString("swimlane_ids", mcp.Description("泳道的id，使用','分割，最多只能20个")),
		mcp.WithString("phase_ids", mcp.Description("所属计划的id，使用','分割，最多只能20个")),
		mcp.WithString("version_ids", mcp.Description("发布的id，使用','分割，最多只能20个")),
		mcp.WithString("created_by_ids", mcp.Description("创建人的id，使用','分割，最多只能20个")),
		mcp.WithString("created_between", mcp.Description("创建时间介于的时间范围，通过','分割起始时间")),
		mcp.WithString("updated_between", mcp.Description("更新时间介于的时间范围，通过','分割起始时间")),
		mcp.WithString("participant_id", mcp.Description("工作项关注人的id")),
		mcp.WithString("keywords", mcp.Description("关键字。支持工作项编号和工作项标题")),
		mcp.WithString("include_deleted", mcp.Description("是否查询已删除的工作项。该值默认为false")),
		mcp.WithString("include_archived", mcp.Description("是否查询已归档的工作项。该值默认为false")),
		mcp.WithNumber("page_size", mcp.Description("每页数量，默认30")),
		mcp.WithNumber("page_index", mcp.Description("页码，从0开始，默认0")),
	)
}

func (t *ListWorkItemsTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.New()
	defer log.Sync()

	requestLogger := log.With("tool", "list_work_items")
	requestLogger.Info("开始获取工作项列表")

	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	queryParams := t.parseListWorkItemsArguments(request.GetArguments())

	respStr, err := t.doListWorkItemsRequest(ctx, token, queryParams, requestLogger)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var resp ListWorkItemsResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("解析响应失败: %v", err)), nil
	}

	result := t.formatListWorkItemsResponse(resp)
	return mcp.NewToolResultText(result), nil
}

func (t *ListWorkItemsTool) parseListWorkItemsArguments(args interface{}) map[string]string {
	queryParams := make(map[string]string)

	if args == nil {
		return queryParams
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return queryParams
	}

	// 解析字符串参数
	stringParams := []string{
		"identifier", "project_ids", "type_ids", "parent_ids", "assignee_ids",
		"state_ids", "start_between", "end_between", "priority_ids", "bug_type_ids",
		"sprint_ids", "board_ids", "entry_ids", "tag_ids", "swimlane_ids",
		"phase_ids", "version_ids", "created_by_ids", "created_between",
		"updated_between", "participant_id", "keywords", "include_deleted", "include_archived",
	}

	for _, param := range stringParams {
		if value, exists := argsMap[param]; exists {
			if str, ok := value.(string); ok && str != "" {
				queryParams[param] = str
			}
		}
	}

	// 解析数字参数
	if pageSize, exists := argsMap["page_size"]; exists {
		if num, ok := pageSize.(float64); ok {
			queryParams["page_size"] = fmt.Sprintf("%.0f", num)
		}
	}

	if pageIndex, exists := argsMap["page_index"]; exists {
		if num, ok := pageIndex.(float64); ok {
			queryParams["page_index"] = fmt.Sprintf("%.0f", num)
		}
	}

	return queryParams
}

func (t *ListWorkItemsTool) doListWorkItemsRequest(ctx context.Context, token string, queryParams map[string]string, log *logger.Logger) (string, error) {
	headers := map[string]string{
		"Authorization": token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, workItemTimeout)
	defer cancel()

	body, _, err := utils.DoGet(timeoutCtx, workItemsAPIURL, headers, queryParams)
	if err != nil {
		return "", fmt.Errorf("获取工作项列表请求失败: %v", err)
	}

	log.With("url", workItemsAPIURL, "response_size_bytes", len(body), "success", true).Debug("获取工作项列表API请求成功")
	return string(body), nil
}

func (t *ListWorkItemsTool) formatListWorkItemsResponse(resp ListWorkItemsResponse) string {
	var result strings.Builder
	result.WriteString("📋 工作项列表\n\n")
	result.WriteString(fmt.Sprintf("📊 总计: %d 个工作项\n", resp.Total))
	result.WriteString(fmt.Sprintf("📄 当前页: %d，每页: %d 个\n\n", resp.PageIndex+1, resp.PageSize))

	if len(resp.Values) == 0 {
		result.WriteString("暂无工作项数据\n")
		return result.String()
	}

	for i, item := range resp.Values {
		result.WriteString(fmt.Sprintf("--- 工作项 %d ---\n", i+1))
		result.WriteString(fmt.Sprintf("🆔 ID: %s\n", item.ID))
		result.WriteString(fmt.Sprintf("🏷️ 标识符: %s\n", item.Identifier))
		result.WriteString(fmt.Sprintf("📋 标题: %s\n", item.Title))
		result.WriteString(fmt.Sprintf("📂 项目: %s (%s)\n", item.Project.Name, item.Project.Identifier))
		result.WriteString(fmt.Sprintf("📊 类型: %s\n", item.Type))

		if item.Assignee != nil {
			result.WriteString(fmt.Sprintf("👤 负责人: %s (%s)\n", item.Assignee.DisplayName, item.Assignee.Name))
		} else {
			result.WriteString("👤 负责人: 未分配\n")
		}

		result.WriteString(fmt.Sprintf("📈 状态: %s (%s)\n", item.State.Name, item.State.Type))

		if item.Priority != nil {
			result.WriteString(fmt.Sprintf("⚡ 优先级: %s\n", item.Priority.Name))
		}

		if item.Parent != nil {
			result.WriteString(fmt.Sprintf("👨‍👩‍👧‍👦 父工作项: %s (%s)\n", item.Parent.Title, item.Parent.Identifier))
		}

		if item.Sprint != nil {
			result.WriteString(fmt.Sprintf("🏃 迭代: %s (%s)\n", item.Sprint.Name, item.Sprint.Status))
		}

		if item.Version != nil {
			result.WriteString(fmt.Sprintf("📦 版本: %s (%s)\n", item.Version.Name, item.Version.Stage.Name))
		}

		if item.Phase != nil {
			result.WriteString(fmt.Sprintf("📅 阶段: %s (%s)\n", item.Phase.Title, item.Phase.Identifier))
		}

		if item.StoryPoints > 0 {
			result.WriteString(fmt.Sprintf("📈 故事点: %d\n", item.StoryPoints))
		}

		if item.StartAt > 0 {
			startTime := time.Unix(item.StartAt, 0)
			result.WriteString(fmt.Sprintf("🚀 开始时间: %s\n", startTime.Format("2006-01-02 15:04:05")))
		}

		if item.EndAt > 0 {
			endTime := time.Unix(item.EndAt, 0)
			result.WriteString(fmt.Sprintf("🏁 截止时间: %s\n", endTime.Format("2006-01-02 15:04:05")))
		}

		if item.CompletedAt > 0 {
			completedTime := time.Unix(item.CompletedAt, 0)
			result.WriteString(fmt.Sprintf("✅ 完成时间: %s\n", completedTime.Format("2006-01-02 15:04:05")))
		}

		if len(item.Tags) > 0 {
			var tagNames []string
			for _, tag := range item.Tags {
				tagNames = append(tagNames, tag.Name)
			}
			result.WriteString(fmt.Sprintf("🏷️ 标签: %s\n", strings.Join(tagNames, ", ")))
		}

		if len(item.Participants) > 0 {
			result.WriteString(fmt.Sprintf("👥 关注人数: %d\n", len(item.Participants)))
		}

		if item.Description != "" {
			// 限制描述长度显示
			desc := item.Description
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			result.WriteString(fmt.Sprintf("📝 描述: %s\n", desc))
		}

		createdTime := time.Unix(item.CreatedAt, 0)
		result.WriteString(fmt.Sprintf("🕐 创建时间: %s\n", createdTime.Format("2006-01-02 15:04:05")))

		updatedTime := time.Unix(item.UpdatedAt, 0)
		result.WriteString(fmt.Sprintf("🕐 更新时间: %s\n", updatedTime.Format("2006-01-02 15:04:05")))

		result.WriteString(fmt.Sprintf("🔗 链接: %s\n", item.URL))

		if item.IsArchived == 1 {
			result.WriteString("📦 状态: 已归档\n")
		}
		if item.IsDeleted == 1 {
			result.WriteString("🗑️ 状态: 已删除\n")
		}

		result.WriteString("\n")
	}

	return result.String()
}

// DeleteWorkItemResponse 删除工作项的响应结构
type DeleteWorkItemResponse struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Identifier  string `json:"identifier"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	StartAt     int64  `json:"start_at"`
	EndAt       int64  `json:"end_at"`
	Description string `json:"description"`
	CompletedAt int64  `json:"completed_at"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	IsArchived  int    `json:"is_archived"`
	IsDeleted   int    `json:"is_deleted"`

	Project struct {
		ID         string `json:"id"`
		URL        string `json:"url"`
		Identifier string `json:"identifier"`
		Name       string `json:"name"`
		Type       string `json:"type"`
	} `json:"project"`

	Assignee struct {
		ID          string `json:"id"`
		URL         string `json:"url"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Avatar      string `json:"avatar"`
	} `json:"assignee"`

	State struct {
		ID    string `json:"id"`
		URL   string `json:"url"`
		Name  string `json:"name"`
		Type  string `json:"type"`
		Color string `json:"color"`
	} `json:"state"`

	Priority struct {
		ID   string `json:"id"`
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"priority"`

	Board *struct {
		ID            string   `json:"id"`
		URL           string   `json:"url"`
		Name          string   `json:"name"`
		WorkItemTypes []string `json:"work_item_types"`
	} `json:"board"`

	Entry *struct {
		ID   string `json:"id"`
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"entry"`

	Swimlane *struct {
		ID   string `json:"id"`
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"swimlane"`

	Properties interface{} `json:"properties"`

	Tags []struct {
		ID   string `json:"id"`
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"tags"`

	Participants []struct {
		ID   string `json:"id"`
		URL  string `json:"url"`
		Type string `json:"type"`
		User struct {
			ID          string `json:"id"`
			URL         string `json:"url"`
			Name        string `json:"name"`
			DisplayName string `json:"display_name"`
			Avatar      string `json:"avatar"`
		} `json:"user"`
	} `json:"participants"`

	CreatedBy struct {
		ID          string `json:"id"`
		URL         string `json:"url"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Avatar      string `json:"avatar"`
	} `json:"created_by"`

	UpdatedBy struct {
		ID          string `json:"id"`
		URL         string `json:"url"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Avatar      string `json:"avatar"`
	} `json:"updated_by"`
}

// DeleteWorkItemTool 删除工作项工具
type DeleteWorkItemTool struct {
	name        string
	description string
}

// NewDeleteWorkItemTool 创建新的工作项删除工具实例
func NewDeleteWorkItemTool() MCPTool {
	return &DeleteWorkItemTool{
		name:        "delete_work_item",
		description: "Delete a work item",
	}
}

func (t *DeleteWorkItemTool) GetName() string {
	return t.name
}

func (t *DeleteWorkItemTool) GetDescription() string {
	return t.description
}

func (t *DeleteWorkItemTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription(t.description),
		mcp.WithString("work_item_id", mcp.Description("工作项的id"), mcp.Required()),
	)
}

func (t *DeleteWorkItemTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.New()
	defer log.Sync()

	requestLogger := log.With("tool", "delete_work_item")
	requestLogger.Info("开始删除工作项")

	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	workItemID, err := t.parseDeleteWorkItemArguments(request.GetArguments())
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	respStr, err := t.doDeleteWorkItemRequest(ctx, token, workItemID, requestLogger)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var resp DeleteWorkItemResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("解析响应失败: %v", err)), nil
	}

	result := t.formatDeleteWorkItemResponse(resp)
	return mcp.NewToolResultText(result), nil
}

func (t *DeleteWorkItemTool) parseDeleteWorkItemArguments(args interface{}) (string, error) {
	if args == nil {
		return "", fmt.Errorf("缺少必需的参数")
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("参数格式错误")
	}

	// 检查必需参数
	workItemID, ok := argsMap["work_item_id"].(string)
	if !ok || workItemID == "" {
		return "", fmt.Errorf("work_item_id 是必需的参数")
	}

	return workItemID, nil
}

func (t *DeleteWorkItemTool) doDeleteWorkItemRequest(ctx context.Context, token string, workItemID string, log *logger.Logger) (string, error) {
	headers := map[string]string{
		"Authorization": token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, workItemTimeout)
	defer cancel()

	// 构建API URL
	deleteURL := fmt.Sprintf("%s/%s", workItemsAPIURL, workItemID)

	body, _, err := utils.DoDelete(timeoutCtx, deleteURL, headers, nil)
	if err != nil {
		return "", fmt.Errorf("删除工作项请求失败: %v", err)
	}

	log.With("url", deleteURL, "response_size_bytes", len(body), "success", true).Debug("删除工作项API请求成功")
	return string(body), nil
}

func (t *DeleteWorkItemTool) formatDeleteWorkItemResponse(resp DeleteWorkItemResponse) string {
	var result strings.Builder
	result.WriteString("🗑️ 工作项删除成功！\n\n")
	result.WriteString(fmt.Sprintf("📋 标题: %s\n", resp.Title))
	result.WriteString(fmt.Sprintf("🆔 工作项ID: %s\n", resp.ID))
	result.WriteString(fmt.Sprintf("🏷️ 标识符: %s\n", resp.Identifier))
	result.WriteString(fmt.Sprintf("📂 项目: %s (%s)\n", resp.Project.Name, resp.Project.Identifier))
	result.WriteString(fmt.Sprintf("🔗 项目ID: %s\n", resp.Project.ID))
	result.WriteString(fmt.Sprintf("📊 类型: %s\n", resp.Type))

	if resp.Description != "" {
		// 限制描述长度显示
		desc := resp.Description
		if len(desc) > 100 {
			desc = desc[:100] + "..."
		}
		result.WriteString(fmt.Sprintf("📝 描述: %s\n", desc))
	}

	result.WriteString(fmt.Sprintf("👤 负责人: %s (%s)\n", resp.Assignee.DisplayName, resp.Assignee.Name))
	result.WriteString(fmt.Sprintf("🔗 负责人ID: %s\n", resp.Assignee.ID))

	result.WriteString(fmt.Sprintf("📈 状态: %s (%s)\n", resp.State.Name, resp.State.Type))
	result.WriteString(fmt.Sprintf("🎨 状态颜色: %s\n", resp.State.Color))

	result.WriteString(fmt.Sprintf("⚡ 优先级: %s\n", resp.Priority.Name))

	if resp.Board != nil {
		result.WriteString(fmt.Sprintf("📋 看板: %s\n", resp.Board.Name))
		if len(resp.Board.WorkItemTypes) > 0 {
			result.WriteString(fmt.Sprintf("📊 支持类型: %s\n", strings.Join(resp.Board.WorkItemTypes, ", ")))
		}
	}

	if resp.Entry != nil {
		result.WriteString(fmt.Sprintf("📂 看板栏: %s\n", resp.Entry.Name))
	}

	if resp.Swimlane != nil {
		result.WriteString(fmt.Sprintf("🏊 泳道: %s\n", resp.Swimlane.Name))
	}

	if resp.StartAt > 0 {
		startTime := time.Unix(resp.StartAt, 0)
		result.WriteString(fmt.Sprintf("🚀 开始时间: %s\n", startTime.Format("2006-01-02 15:04:05")))
	}

	if resp.EndAt > 0 {
		endTime := time.Unix(resp.EndAt, 0)
		result.WriteString(fmt.Sprintf("🏁 截止时间: %s\n", endTime.Format("2006-01-02 15:04:05")))
	}

	if resp.CompletedAt > 0 {
		completedTime := time.Unix(resp.CompletedAt, 0)
		result.WriteString(fmt.Sprintf("✅ 完成时间: %s\n", completedTime.Format("2006-01-02 15:04:05")))
	}

	if len(resp.Tags) > 0 {
		var tagNames []string
		for _, tag := range resp.Tags {
			tagNames = append(tagNames, tag.Name)
		}
		result.WriteString(fmt.Sprintf("🏷️ 标签: %s\n", strings.Join(tagNames, ", ")))
	}

	if len(resp.Participants) > 0 {
		result.WriteString(fmt.Sprintf("👥 关注人数: %d\n", len(resp.Participants)))
		for i, participant := range resp.Participants {
			if i < 3 { // 只显示前3个关注人
				result.WriteString(fmt.Sprintf("  • %s (%s)\n", participant.User.DisplayName, participant.User.Name))
			}
		}
		if len(resp.Participants) > 3 {
			result.WriteString(fmt.Sprintf("  • ... 还有 %d 人\n", len(resp.Participants)-3))
		}
	}

	result.WriteString(fmt.Sprintf("👤 创建人: %s (%s)\n", resp.CreatedBy.DisplayName, resp.CreatedBy.Name))
	result.WriteString(fmt.Sprintf("👤 更新人: %s (%s)\n", resp.UpdatedBy.DisplayName, resp.UpdatedBy.Name))

	createdTime := time.Unix(resp.CreatedAt, 0)
	result.WriteString(fmt.Sprintf("🕐 创建时间: %s\n", createdTime.Format("2006-01-02 15:04:05")))

	updatedTime := time.Unix(resp.UpdatedAt, 0)
	result.WriteString(fmt.Sprintf("🕐 更新时间: %s\n", updatedTime.Format("2006-01-02 15:04:05")))

	// 删除状态标识
	if resp.IsDeleted == 1 {
		result.WriteString("🗑️ 删除状态: 已删除\n")
	}
	if resp.IsArchived == 1 {
		result.WriteString("📦 归档状态: 已归档\n")
	}

	result.WriteString(fmt.Sprintf("🔗 工作项链接: %s\n", resp.URL))

	result.WriteString("\n⚠️ 注意: 工作项已被标记为删除状态，但数据仍保留在系统中。\n")

	return result.String()
}
