package rpc

import (

	"guanxi/eazy-claw/internal/dto"
	"guanxi/eazy-claw/internal/service"
	"guanxi/eazy-claw/pkg/rpcutil"

	"github.com/DemonZack/simplejrpc-go/net/gsock"
)

// GetAgentFiles 获取所有Agent人格文件
func (s *Server) GetAgentFiles(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	result, err := service.NewAgentService().GetAgentFiles()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SaveAgentFile 保存Agent人格文件
func (s *Server) SaveAgentFile(req *gsock.Request) (any, error) {
	var args dto.AgentSaveReq
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewAgentService().SaveAgentFile(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ResetAgentFile 重置Agent人格文件为默认内容
func (s *Server) ResetAgentFile(req *gsock.Request) (any, error) {
	var args dto.AgentFileReq
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewAgentService().ResetAgentFile(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAgentTemplates 获取预设模板列表
func (s *Server) GetAgentTemplates(req *gsock.Request) (any, error) {
	result, err := service.NewAgentService().GetAgentTemplates()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ApplyAgentTemplate 应用预设模板
func (s *Server) ApplyAgentTemplate(req *gsock.Request) (any, error) {
	var args dto.ApplyTemplateReq
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewAgentService().ApplyAgentTemplate(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ========== 多 Agent 管理 ==========

// ListAgents 获取Agent列表
func (s *Server) ListAgents(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	result, err := service.NewAgentService().ListAgents()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateAgent 创建Agent
func (s *Server) CreateAgent(req *gsock.Request) (any, error) {
	var args dto.CreateAgentReq
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewAgentService().CreateAgent(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateAgent 更新Agent
func (s *Server) UpdateAgent(req *gsock.Request) (any, error) {
	var args dto.UpdateAgentReq
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewAgentService().UpdateAgent(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteAgent 删除Agent
func (s *Server) DeleteAgent(req *gsock.Request) (any, error) {
	var args dto.DeleteAgentReq
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewAgentService().DeleteAgent(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAgentDetail 获取Agent详情
func (s *Server) GetAgentDetail(req *gsock.Request) (any, error) {
	var args dto.GetAgentDetailReq
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewAgentService().GetAgentDetail(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetConfiguredModels 获取已配置的模型列表
func (s *Server) GetConfiguredModels(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	result, err := service.NewAgentService().GetConfiguredModels()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAgentActivity 获取 Agent 实时活跃状态和子 Agent（临时工）
func (s *Server) GetAgentActivity(req *gsock.Request) (any, error) {
	result, err := service.NewAgentService().GetAgentActivity()
	if err != nil {
		return nil, err
	}
	return result, nil
}
