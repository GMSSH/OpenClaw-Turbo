package rpc

import (

	"guanxi/eazy-claw/internal/service"
	"guanxi/eazy-claw/pkg/rpcutil"

	"github.com/DemonZack/simplejrpc-go/net/gsock"
)

// SearchSkills 搜索技能
func (s *Server) SearchSkills(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewSkillService().SearchSkills(args)
}

// InspectSkill 查看技能详情
func (s *Server) InspectSkill(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewSkillService().InspectSkill(args)
}

// InstallSkill 安装技能
func (s *Server) InstallSkill(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewSkillService().InstallSkill(args)
}

// UninstallSkill 卸载技能
func (s *Server) UninstallSkill(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewSkillService().UninstallSkill(args)
}

// ListInstalledSkills 列出已安装技能
func (s *Server) ListInstalledSkills(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	return service.NewSkillService().ListInstalledSkills()
}

// ExploreSkills 浏览最新技能
func (s *Server) ExploreSkills(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	return service.NewSkillService().ExploreSkills()
}

// ListBuiltinSkills 列出内置技能
func (s *Server) ListBuiltinSkills(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	return service.NewSkillService().ListBuiltinSkills()
}

// InstallBuiltinSkill 安装内置技能
func (s *Server) InstallBuiltinSkill(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewSkillService().InstallBuiltinSkill(args)
}

// UninstallBuiltinSkill 卸载内置技能
func (s *Server) UninstallBuiltinSkill(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewSkillService().UninstallBuiltinSkill(args)
}

// GetActiveSkillCount 获取启用的能力数总计
func (s *Server) GetActiveSkillCount(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	return service.NewSkillService().GetActiveSkillCount()
}
