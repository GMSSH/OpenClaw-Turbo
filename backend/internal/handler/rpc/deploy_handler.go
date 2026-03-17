package rpc

import (

	"guanxi/eazy-claw/internal/dto"
	"guanxi/eazy-claw/internal/service"
	"guanxi/eazy-claw/pkg/rpcutil"

	"github.com/DemonZack/simplejrpc-go/net/gsock"
)

// CheckEnvironment 检测Docker环境
func (s *Server) CheckEnvironment(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	result, err := service.NewDeployService().CheckEnvironment()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GenerateToken 生成随机访问Token
func (s *Server) GenerateToken(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	result, err := service.NewDeployService().GenerateToken()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Deploy 执行部署
func (s *Server) Deploy(req *gsock.Request) (any, error) {
	var args dto.DeployReq
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	// 根据部署模式分流
	if args.DeployMode == "local" {
		result, err := service.NewDeployService().DeployLocal(args)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	result, err := service.NewDeployService().Deploy(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// InstallNodeEnv 安装 Node.js 环境
func (s *Server) InstallNodeEnv(req *gsock.Request) (any, error) {
	result, err := service.NewDeployService().InstallNodeEnv()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetDeployLogs 获取部署日志
func (s *Server) GetDeployLogs(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	result, err := service.NewDeployService().GetDeployLogs()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetClawStatus 获取OpenClaw运行状态
func (s *Server) GetClawStatus(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	result, err := service.NewDeployService().GetClawStatus()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// IsClawInstalled 检查 OpenClaw 是否已安装
func (s *Server) IsClawInstalled(req *gsock.Request) (any, error) {
	return service.NewDeployService().IsClawInstalled(), nil
}

// CheckPorts 检测端口是否被占用
func (s *Server) CheckPorts(req *gsock.Request) (any, error) {
	var args dto.CheckPortsReq
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewDeployService().CheckPorts(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetClawConfig 获取OpenClaw配置信息
func (s *Server) GetClawConfig(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	result, err := service.NewDeployService().GetClawConfig()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// StopClaw 停止容器
func (s *Server) StopClaw(req *gsock.Request) (any, error) {
	result, err := service.NewDeployService().StopClaw()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// RestartClaw 重启容器
func (s *Server) RestartClaw(req *gsock.Request) (any, error) {
	result, err := service.NewDeployService().RestartClaw()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UninstallClaw 卸载容器、镜像和数据
func (s *Server) UninstallClaw(req *gsock.Request) (any, error) {
	result, err := service.NewDeployService().UninstallClaw()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// TestApiConnection 测试AI API连通性
func (s *Server) TestApiConnection(req *gsock.Request) (any, error) {
	var args dto.TestApiReq
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewDeployService().TestApiConnection(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateModelConfig 切换AI模型配置
func (s *Server) UpdateModelConfig(req *gsock.Request) (any, error) {
	var args dto.UpdateModelReq
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewDeployService().UpdateModelConfig(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateMemoryConfig 更新记忆配置
func (s *Server) UpdateMemoryConfig(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	result, err := service.NewDeployService().UpdateMemoryConfig(args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetClawWsInfo 获取 OpenClaw WebSocket 连接信息
func (s *Server) GetClawWsInfo(req *gsock.Request) (any, error) {
	result, err := service.NewDeployService().GetClawWsInfo()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetWsProxyStatus 获取 WS 代理状态
func (s *Server) GetWsProxyStatus(req *gsock.Request) (any, error) {
	return service.NewDeployService().GetWsProxyStatus()
}

// ToggleWsProxy 开启/关闭 WS 代理
func (s *Server) ToggleWsProxy(req *gsock.Request) (any, error) {
	var params map[string]any
	if err := rpcutil.ParseParams(req, &params); err != nil {
		return nil, err
	}
	return service.NewDeployService().ToggleWsProxy(params)
}

// GetRecentLogs 获取 OpenClaw 近期日志
func (s *Server) GetRecentLogs(req *gsock.Request) (any, error) {
	var params map[string]any
	if err := rpcutil.ParseParams(req, &params); err != nil {
		params = map[string]any{}
	}
	return service.NewDeployService().GetRecentLogs(params)
}

// GetOpenClawConfigDirPath 获取 OpenClaw 配置目录路径
func (s *Server) GetOpenClawConfigDirPath(req *gsock.Request) (any, error) {
	return service.NewDeployService().GetOpenClawConfigDirPath(), nil
}
