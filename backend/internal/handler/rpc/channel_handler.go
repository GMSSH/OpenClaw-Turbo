package rpc

import (

	"guanxi/eazy-claw/internal/service"
	"guanxi/eazy-claw/pkg/rpcutil"

	"github.com/DemonZack/simplejrpc-go/net/gsock"
)

// GetChannels 获取所有通道
func (s *Server) GetChannels(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	result, err := service.NewChannelService().GetChannels()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SaveChannel 保存通道配置
func (s *Server) SaveChannel(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewChannelService().SaveChannel(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteChannel 删除通道
func (s *Server) DeleteChannel(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewChannelService().DeleteChannel(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ToggleChannel 启用/禁用通道
func (s *Server) ToggleChannel(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewChannelService().ToggleChannel(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}
