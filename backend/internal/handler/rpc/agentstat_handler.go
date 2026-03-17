package rpc

import (
	"guanxi/eazy-claw/internal/dto"
	"guanxi/eazy-claw/internal/service"
	"guanxi/eazy-claw/pkg/rpcutil"

	"github.com/DemonZack/simplejrpc-go/net/gsock"
)

// GetAgentStatuses 获取所有 Agent 实时状态（working/online/idle/offline）
func (s *Server) GetAgentStatuses(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	result, err := service.NewAgentStatService().GetAgentStatuses()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAgentTokenStats 获取 Agent Token 消耗统计
func (s *Server) GetAgentTokenStats(req *gsock.Request) (any, error) {
	var args dto.GetAgentTokenStatsReq
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewAgentStatService().GetAgentTokenStats(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetGatewayHealth 探测 Gateway 健康状态
func (s *Server) GetGatewayHealth(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	result, err := service.NewAgentStatService().GetGatewayHealth()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// TestPlatformConn 测试平台连通性
func (s *Server) TestPlatformConn(req *gsock.Request) (any, error) {
	var args dto.TestPlatformConnReq
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewAgentStatService().TestPlatformConn(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAgentSessions 获取指定 Agent 的会话列表
func (s *Server) GetAgentSessions(req *gsock.Request) (any, error) {
	var args dto.GetAgentSessionsReq
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewAgentStatService().GetAgentSessions(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}
