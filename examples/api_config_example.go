package main

import (
	"fmt"
	"log"

	"PingCodeMcp/internal/config"
)

func main() {
	fmt.Println("=== API 配置文件使用示例 ===")

	// 1. 加载配置（自动处理 API 配置文件）
	cfg := config.Load()

	fmt.Printf("配置版本: %s\n", cfg.Version)
	fmt.Printf("服务器名称: %s\n", cfg.Server.Name)
	fmt.Printf("API配置文件: %s\n", cfg.APIConfigFile)

	// 2. 显示 API 配置信息
	fmt.Println("\n=== API 配置信息 ===")
	fmt.Printf("IP查询服务URL: %s\n", cfg.API.IPSearch.URL)
	fmt.Printf("IP查询超时: %s\n", cfg.API.IPSearch.Timeout)
	fmt.Printf("用户信息服务URL: %s\n", cfg.API.UserInfo.URL)
	fmt.Printf("用户信息超时: %s\n", cfg.API.UserInfo.Timeout)

	// 3. 显示认证配置信息
	fmt.Println("\n=== 认证配置信息 ===")
	fmt.Printf("Token头名称: %s\n", cfg.Auth.Token.HeaderName)
	fmt.Printf("Token前缀: %s\n", cfg.Auth.Token.Prefix)
	fmt.Printf("会话超时: %s\n", cfg.Auth.Session.Timeout)

	// 4. 直接加载 API 配置文件示例
	fmt.Println("\n=== 直接加载 API 配置文件 ===")
	if cfg.APIConfigFile != "" {
		apiConfig, err := config.LoadAPIConfig(cfg.APIConfigFile)
		if err != nil {
			log.Printf("加载API配置文件失败: %v", err)
		} else {
			fmt.Printf("API配置版本: %s\n", apiConfig.Version)
			fmt.Printf("企业用户服务URL: %s\n", apiConfig.API.EnterpriseUsers.URL)
		}
	}

	// 5. 演示环境变量覆盖
	fmt.Println("\n=== 环境变量覆盖示例 ===")
	fmt.Println("可以通过以下环境变量覆盖配置:")
	fmt.Println("export PINGCODE_MCP_API_CONFIG_FILE=\"custom_api.yaml\"")
	fmt.Println("export PINGCODE_MCP_API_IP_SEARCH_URL=\"https://custom-api.com/ip\"")
	fmt.Println("export PINGCODE_MCP_API_USER_INFO_URL=\"https://custom-api.com/user\"")
}
