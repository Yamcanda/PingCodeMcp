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
	// API请求超时时间
	apiTimeout = 15 * time.Second
)

// 注册工具
func init() {
	toolsFns = append(toolsFns, func() *[]MCPTool {
		return &[]MCPTool{
			NewProjectSprintWorkitemsTool(),
		}
	})
}

// 定义了认证接口响应的结构
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type ProjectListValues struct {
	Values []Project `json:"values"`
}

// 迭代
type Sprints struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Project Project `json:"project"`
	Status  string  `json:"status"`
}

type SprintsListValues struct {
	Values []Sprints `json:"values"`
}

// 工作项
type WorkItems struct {
	ID          string  `json:"id"`
	Identifier  string  `json:"identifier"`
	Title       string  `json:"title"`
	Type        string  `json:"type"`
	StartAt     *int64  `json:"start_at"`
	EndAt       *int64  `json:"end_at"`
	Project     Project `json:"project"`
	Description string  `json:"description"`
	State       State   `json:"state"`
	Assignee    UserBy  `json:"assignee"`
	CreatedAt   int64   `json:"created_at"`
	CreatedBy   UserBy  `json:"created_by"`
	UpdatedAt   int64   `json:"updated_at"`
	UpdatedBy   UserBy  `json:"updated_by"`
}

type WorkItemsValues struct {
	Values []WorkItems `json:"values"`
}

type WorkItemsWarp struct {
	ID           string `json:"id"`
	Identifier   string `json:"identifier"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	TypeName     string `json:"type_name"`
	StartAt      *int64 `json:"start_at"`
	EndAt        *int64 `json:"end_at"`
	ProjectName  string `json:"project"`
	Description  string `json:"description"`
	StateName    string `json:"state"`
	AssigneeName string `json:"assignee"`
	CreatedAt    int64  `json:"created_at"`
	CreatedBy    string `json:"created_by"`
	UpdatedAt    int64  `json:"updated_at"`
	UpdatedBy    string `json:"updated_by"`
}

// other
type State struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type UserBy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type ProjectSprintWorkitemListTool struct {
	name        string
	description string
}

type StoryTaskBugRequest struct {
	Identifier string  `json:"identifier"`
	SprintName string  `json:"sprint_name"`
	TypeIds    *string `json:"type_ids"`
}

// 创建信息工具实例
func NewProjectSprintWorkitemsTool() MCPTool {
	return &ProjectSprintWorkitemListTool{
		name:        "get_project_sprint_workitems",
		description: "获取 Pingcode 指定项目和迭代的所有工作项，可指定工作项类型，工作项类型包括用户故事、任务和缺陷",
	}
}

// 返回工具名称
func (t *ProjectSprintWorkitemListTool) GetName() string {
	return t.name
}

// 返回工具描述
func (t *ProjectSprintWorkitemListTool) GetDescription() string {
	return t.description
}

// 返回工具定义
func (t *ProjectSprintWorkitemListTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name, mcp.WithDescription(t.description),
		mcp.WithString("identifier", mcp.Description("项目标识 - 必填，要查询迭代的项目唯一标识，可通过 get_project_list 工具获取"), mcp.Required()),
		mcp.WithString("sprint_name", mcp.Description("迭代名称 - 必填，要查询迭代的项目唯一标识，可通过 get_project_list 工具获取"), mcp.Required()),
		mcp.WithString("type_ids", mcp.Description("工作项类型ID - 可选，更改工作项类型，如：epic(史诗)、feature(特性)、story(用户故事)、task(任务)、bug(缺陷)，以逗号分隔，为空获取所有类型工作项")),
	)
}

func (t *ProjectSprintWorkitemListTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.New()
	defer log.Sync()

	requestLogger := log.With("tool", "get_project_sprint_workitems", "request_id", fmt.Sprintf("user_%d", time.Now().UnixNano()))
	requestLogger.Info("开始获取信息")

	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		requestLogger.With("success", false, "error", "missing_token").Error("获取信息失败：缺少认证令牌")
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	storyTaskBugRequest, err := t.parseArguments(request.GetArguments())
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	workItemsWarps, err := t.doRequest(ctx, token, *storyTaskBugRequest, requestLogger)
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("信息API调用失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := t.formatResponse(workItemsWarps)

	if result != "" {
		requestLogger.Debug(result)
	} else {
		requestLogger.Warn("转换失败，空")
	}

	return mcp.NewToolResultText(result), nil
}

func (t *ProjectSprintWorkitemListTool) parseArguments(args interface{}) (*StoryTaskBugRequest, error) {
	if args == nil {
		return nil, fmt.Errorf("缺少必需的参数")
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("参数格式错误")
	}

	identifier, ok := argsMap["identifier"].(string)
	if !ok || identifier == "" {
		return nil, fmt.Errorf("identifier 是必需的参数")
	}

	sprintName, ok := argsMap["sprint_name"].(string)
	if !ok || sprintName == "" {
		return nil, fmt.Errorf("sprint_name 是必需的参数")
	}

	req := &StoryTaskBugRequest{
		Identifier: identifier,
		SprintName: sprintName,
	}

	// 解析可选参数
	if typeIds, exists := argsMap["type_ids"]; exists {
		if str, ok := typeIds.(string); ok && str != "" {
			req.TypeIds = &str
		}
	}

	return req, nil
}

func (t *ProjectSprintWorkitemListTool) doRequest(ctx context.Context, token string, storyTaskBugRequest StoryTaskBugRequest, log *logger.Logger) ([]WorkItemsWarp, error) {
	// 项目列表
	projects, err := getProjects(ctx, token, storyTaskBugRequest.Identifier, log)
	if err != nil {
		log.Errorf("获取项目列表失败: %v", err)
	}
	log.Infof("获取项目列表成功: %+v", projects.Values[0].ID)

	// 迭代列表
	sprints, err := getSprints(ctx, token, projects.Values[0].ID, storyTaskBugRequest.SprintName, log)
	if err != nil {
		log.Fatalf("获取迭代列表失败: %v", err)
	}

	if len(sprints.Values) > 0 {
		log.Infof("获取迭代列表成功: %+v", sprints.Values[0].ID)

		warpedWorkItems := []WorkItemsWarp{}

		if storyTaskBugRequest.TypeIds != nil && *storyTaskBugRequest.TypeIds != "" {
			typeIds := strings.Split(*storyTaskBugRequest.TypeIds, ",")

			for _, typeId := range typeIds {
				switch typeId {
				case "story", "用户故事":
					storyWorkItems, _ := getWorkItemsStory(ctx, token, projects.Values[0].ID, sprints.Values[0].ID, log)
					warpedWorkItems = append(warpedWorkItems, storyWorkItems...)
				case "task", "任务":
					taskWorkItems, _ := getWorkItemsTask(ctx, token, projects.Values[0].ID, sprints.Values[0].ID, log)
					warpedWorkItems = append(warpedWorkItems, taskWorkItems...)
				case "bug", "缺陷":
					bugWorkItems, _ := getWorkItemsBug(ctx, token, projects.Values[0].ID, sprints.Values[0].ID, log)
					warpedWorkItems = append(warpedWorkItems, bugWorkItems...)
				}
			}
		} else {
			storyWorkItems, _ := getWorkItemsStory(ctx, token, projects.Values[0].ID, sprints.Values[0].ID, log)
			taskWorkItems, _ := getWorkItemsTask(ctx, token, projects.Values[0].ID, sprints.Values[0].ID, log)
			bugWorkItems, _ := getWorkItemsBug(ctx, token, projects.Values[0].ID, sprints.Values[0].ID, log)

			warpedWorkItems = append(warpedWorkItems, storyWorkItems...)
			warpedWorkItems = append(warpedWorkItems, taskWorkItems...)
			warpedWorkItems = append(warpedWorkItems, bugWorkItems...)
		}

		return warpedWorkItems, nil
	} else {
		log.Warn("未获取到迭代列表")
	}

	return nil, nil
}

func (t *ProjectSprintWorkitemListTool) formatResponse(workItemsWarps []WorkItemsWarp) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("📋 工作项列表 (总计 %d 个工作项):\n\n", len(workItemsWarps)))

	if len(workItemsWarps) == 0 {
		result.WriteString("暂无工作项数据")
		return result.String()
	}

	for i, workItemsWarp := range workItemsWarps {
		result.WriteString(fmt.Sprintf("📂 %d. 工作项标识 %s\n", i+1, workItemsWarp.Identifier))
		result.WriteString(fmt.Sprintf("类型: %s (%s)\n", workItemsWarp.Type, workItemsWarp.TypeName))
		result.WriteString(fmt.Sprintf("状态: %s\n", workItemsWarp.StateName))
		result.WriteString(fmt.Sprintf("归属项目: %s\n", workItemsWarp.ProjectName))

		if workItemsWarp.Title != "" {
			result.WriteString(fmt.Sprintf("标题: %s\n", workItemsWarp.Title))
		}

		// if workItemsWarp.Description != "" {
		// 	result.WriteString(fmt.Sprintf("📝 描述: %s\n", workItemsWarp.Description))
		// }

		// 显示负责人
		if workItemsWarp.AssigneeName != "" {
			result.WriteString(fmt.Sprintf("   负责人: %s\n", workItemsWarp.AssigneeName))
		}

		// 显示时间信息
		if workItemsWarp.StartAt != nil {
			startTime := time.Unix(*workItemsWarp.StartAt, 0).Format("2006-01-02")
			result.WriteString(fmt.Sprintf("   开始时间: %s\n", startTime))
		}

		if workItemsWarp.EndAt != nil {
			endTime := time.Unix(*workItemsWarp.EndAt, 0).Format("2006-01-02")
			result.WriteString(fmt.Sprintf("   结束时间: %s\n", endTime))
		}

		// 显示创建信息
		createdTime := time.Unix(workItemsWarp.CreatedAt, 0).Format("2006-01-02 15:04:05")
		result.WriteString(fmt.Sprintf("   创建时间: %s\n", createdTime))
		result.WriteString(fmt.Sprintf("   创建者: %s\n", workItemsWarp.CreatedBy))

		// 显示更新信息
		updateTime := time.Unix(workItemsWarp.UpdatedAt, 0).Format("2006-01-02 15:04:05")
		result.WriteString(fmt.Sprintf("   更新时间: %s\n", updateTime))
		result.WriteString(fmt.Sprintf("   更新者: %s\n", workItemsWarp.UpdatedBy))
	}

	return result.String()
}

// 获取项目列表
func getProjects(ctx context.Context, token string, identifier string, log *logger.Logger) (*ProjectListValues, error) {
	apiURL := fmt.Sprintf("%s/v1/project/projects?identifier=%s", GetBaseUrl(), identifier)

	headers := map[string]string{
		"Authorization": token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	body, _, err := utils.DoGet(timeoutCtx, apiURL, headers, nil)
	if err != nil {
		log.With("url", apiURL, "success", false, "error", err.Error()).Error("HTTP请求失败")
		return nil, err
	}

	var projectListValues ProjectListValues
	if err := json.Unmarshal(body, &projectListValues); err != nil {
		return nil, fmt.Errorf("JSON解码失败: %w, 响应: %s", err, string(body))
	}

	return &projectListValues, nil
}

// 获取迭代列表
func getSprints(ctx context.Context, token string, projectId string, sprintsName string, log *logger.Logger) (*SprintsListValues, error) {
	apiURL := fmt.Sprintf("%s/v1/project/projects/%s/sprints?status=in_progress&name=%s", GetBaseUrl(), projectId, sprintsName)

	headers := map[string]string{
		"Authorization": token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	body, _, err := utils.DoGet(timeoutCtx, apiURL, headers, nil)

	if err != nil {
		log.With("url", apiURL, "success", false, "error", err.Error()).Error("HTTP请求失败")
		return nil, err
	}

	var sprintsListValues SprintsListValues
	if err := json.Unmarshal(body, &sprintsListValues); err != nil {
		return nil, fmt.Errorf("JSON解码失败: %w, 响应: %s", err, string(body))
	}

	return &sprintsListValues, nil
}

func getWorkItemsStory(ctx context.Context, token, projectIds string, sprintIds string, log *logger.Logger) ([]WorkItemsWarp, error) {
	page := 0

	warpedWorkItems := []WorkItemsWarp{}
	for {
		workItemsWarp, err := getWorkItemsTypeIds(ctx, token, page, projectIds, sprintIds, "story", log)
		if err != nil {
			log.Errorf("获取 story 失败: %v", err)
			break
		}

		warpedWorkItems = append(warpedWorkItems, workItemsWarp...)
		if len(workItemsWarp) < 100 {
			break
		}
		page++
	}

	return warpedWorkItems, nil
}

func getWorkItemsTask(ctx context.Context, token, projectIds string, sprintIds string, log *logger.Logger) ([]WorkItemsWarp, error) {
	page := 0

	warpedWorkItems := []WorkItemsWarp{}
	for {
		workItemsWarp, err := getWorkItemsTypeIds(ctx, token, page, projectIds, sprintIds, "task", log)
		if err != nil {
			log.Errorf("获取 task 失败: %v", err)
			break
		}

		warpedWorkItems = append(warpedWorkItems, workItemsWarp...)
		if len(workItemsWarp) < 100 {
			break
		}
		page++
	}

	return warpedWorkItems, nil
}

func getWorkItemsBug(ctx context.Context, token, projectIds string, sprintIds string, log *logger.Logger) ([]WorkItemsWarp, error) {
	page := 0

	warpedWorkItems := []WorkItemsWarp{}
	for {
		workItemsWarp, err := getWorkItemsTypeIds(ctx, token, page, projectIds, sprintIds, "bug", log)
		if err != nil {
			log.Errorf("获取 bug 失败: %v", err)
			break
		}

		warpedWorkItems = append(warpedWorkItems, workItemsWarp...)
		if len(workItemsWarp) < 100 {
			break
		}
		page++
	}

	return warpedWorkItems, nil
}

func getWorkItemsTypeIds(ctx context.Context, token string, pageIndex int, projectIds string, sprintIds string, typeId string, log *logger.Logger) ([]WorkItemsWarp, error) {
	workItems, err := getWorkItems(ctx, token, pageIndex, projectIds, sprintIds, typeId, log)
	if err != nil {
		log.Errorf("获取工作项列表失败: %v", err)
		return nil, err
	}

	typeName := "用户故事"
	switch typeId {
	case "task":
		typeName = "任务"
	case "bug":
		typeName = "缺陷"
	}

	if len(workItems.Values) > 0 {
		log.Infof("获取到 %d 个工作项", len(workItems.Values))
		warpedWorkItems := []WorkItemsWarp{}
		for _, item := range workItems.Values {
			warpedWorkItem := WorkItemsWarp{
				// ID:         item.ID,
				Identifier:  item.Identifier,
				Title:       item.Title,
				Type:        item.Type,
				TypeName:    typeName,
				StartAt:     item.StartAt,
				EndAt:       item.EndAt,
				ProjectName: item.Project.Name,
				// Description:  item.Description,
				StateName:    item.State.Name,
				AssigneeName: item.Assignee.DisplayName,
				CreatedAt:    item.CreatedAt,
				CreatedBy:    item.CreatedBy.DisplayName,
				UpdatedAt:    item.UpdatedAt,
				UpdatedBy:    item.UpdatedBy.DisplayName,
			}

			warpedWorkItems = append(warpedWorkItems, warpedWorkItem)
		}

		return warpedWorkItems, nil
	} else {
		log.Infof("获取工作项列表成功，但列表为空。工作项类型ID: %s", typeId)
	}
	return nil, nil
}

// 获取工作项列表
func getWorkItems(ctx context.Context, token string, pageIndex int, projectIds string, sprintIds string, typeIds string, log *logger.Logger) (*WorkItemsValues, error) {
	apiURL := fmt.Sprintf("%s/v1/project/work_items?page_index=%d&page_size=100&project_ids=%s&sprint_ids=%s&type_ids=%s", GetBaseUrl(), pageIndex, projectIds, sprintIds, typeIds)

	headers := map[string]string{
		"Authorization": token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	body, _, err := utils.DoGet(timeoutCtx, apiURL, headers, nil)

	if err != nil {
		log.With("url", apiURL, "success", false, "error", err.Error()).Error("HTTP请求失败")
		return nil, err
	}

	var workItemsValues WorkItemsValues
	if err := json.Unmarshal(body, &workItemsValues); err != nil {
		return nil, fmt.Errorf("JSON解码失败: %w, 响应: %s", err, string(body))
	}

	return &workItemsValues, nil
}
