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

// ToggleBuiltinSkill 通过 WS 网关启用/禁用内置技能
func (s *Server) ToggleBuiltinSkill(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewSkillService().ToggleBuiltinSkill(args)
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

// IsClawHubInstalled 检测 clawhub 是否已全局安装
func (s *Server) IsClawHubInstalled(req *gsock.Request) (any, error) {
	return service.NewSkillService().IsClawHubInstalled(), nil
}

// InstallClawHub 全局安装 clawhub
func (s *Server) InstallClawHub(req *gsock.Request) (any, error) {
	return service.NewSkillService().InstallClawHub()
}

// GetActiveSkillCount 获取启用的能力数总计
func (s *Server) GetActiveSkillCount(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)
	return service.NewSkillService().GetActiveSkillCount()
}

// ListEnvVars 读取 .env 环境变量
func (s *Server) ListEnvVars(req *gsock.Request) (any, error) {
	return service.NewEnvVarService().ListEnvVars()
}

// SaveEnvVars 保存 .env 环境变量
func (s *Server) SaveEnvVars(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewEnvVarService().SaveEnvVars(args)
}

// GetModelsConfig 获取模型配置
func (s *Server) GetModelsConfig(req *gsock.Request) (any, error) {
	return service.NewModelService().GetModelsConfig()
}

// SaveModelsConfig 保存模型配置
func (s *Server) SaveModelsConfig(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewModelService().SaveModelsConfig(args)
}

// GetSkillsDir 获取技能目录路径
func (s *Server) GetSkillsDir(req *gsock.Request) (any, error) {
	return service.NewSkillService().GetSkillsDir(), nil
}

// ListRecommendedSkills 获取推荐技能列表（来自 skill.json + scripts/）
func (s *Server) ListRecommendedSkills(req *gsock.Request) (any, error) {
	return service.NewSkillService().ListRecommendedSkills()
}

// InstallRecommendedSkill 安装推荐技能（解压 zip 到 skills 目录）
func (s *Server) InstallRecommendedSkill(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewSkillService().InstallRecommendedSkill(args)
}

// UninstallRecommendedSkill 卸载推荐技能（删除 skills 目录中对应文件夹）
func (s *Server) UninstallRecommendedSkill(req *gsock.Request) (any, error) {
	var args map[string]any
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	return service.NewSkillService().UninstallRecommendedSkill(args)
}
