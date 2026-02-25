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
