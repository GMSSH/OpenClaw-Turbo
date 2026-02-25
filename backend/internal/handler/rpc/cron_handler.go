package rpc

import (
	"guanxi/eazy-claw/internal/service"
	"guanxi/eazy-claw/pkg/rpcutil"

	"github.com/DemonZack/simplejrpc-go/net/gsock"
)

// CronStatus 获取调度器状态
func (s *Server) CronStatus(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	return service.NewCronService().CronStatus()
}

// ListCronJobs 列出定时任务
func (s *Server) ListCronJobs(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	return service.NewCronService().ListCronJobs()
}

// AddCronJob 新增定时任务
func (s *Server) AddCronJob(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewCronService().AddCronJob(args)
}

// EditCronJob 编辑定时任务
func (s *Server) EditCronJob(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewCronService().EditCronJob(args)
}

// RemoveCronJob 删除定时任务
func (s *Server) RemoveCronJob(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewCronService().RemoveCronJob(args)
}

// EnableCronJob 启用定时任务
func (s *Server) EnableCronJob(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewCronService().EnableCronJob(args)
}

// DisableCronJob 禁用定时任务
func (s *Server) DisableCronJob(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewCronService().DisableCronJob(args)
}

// RunCronJob 手动执行定时任务
func (s *Server) RunCronJob(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewCronService().RunCronJob(args)
}

// GetCronRuns 获取运行历史
func (s *Server) GetCronRuns(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewCronService().GetCronRuns(args)
}
