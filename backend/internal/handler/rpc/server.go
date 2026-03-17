package rpc

import (
	"errors"
	"os"

	"github.com/DemonZack/simplejrpc-go"
	"github.com/DemonZack/simplejrpc-go/core"
	"github.com/DemonZack/simplejrpc-go/core/config"
	"github.com/DemonZack/simplejrpc-go/core/gi18n"
	"github.com/DemonZack/simplejrpc-go/net/gsock"
	"github.com/DemonZack/simplejrpc-go/os/gpath"
)

// Server RPC服务器
type Server struct {
}

// NewServer 创建RPC服务器实例
func NewServer() *Server {

	return &Server{}
}

// RegisterHandles 注册所有RPC处理函数
func (s *Server) RegisterHandles(ds interface {
	RegisterHandle(name string, handler func(*gsock.Request) (any, error), middlewares ...gsock.RPCMiddleware)
},
) {
	// 注册基础接口
	ds.RegisterHandle("ping", s.Ping)

	// 部署管理接口
	ds.RegisterHandle("checkEnvironment", s.CheckEnvironment)
	ds.RegisterHandle("generateToken", s.GenerateToken)
	ds.RegisterHandle("deploy", s.Deploy)
	ds.RegisterHandle("getDeployLogs", s.GetDeployLogs)
	ds.RegisterHandle("getClawStatus", s.GetClawStatus)
	ds.RegisterHandle("isClawInstalled", s.IsClawInstalled)
	ds.RegisterHandle("checkPorts", s.CheckPorts)
	ds.RegisterHandle("getClawConfig", s.GetClawConfig)
	ds.RegisterHandle("stopClaw", s.StopClaw)
	ds.RegisterHandle("restartClaw", s.RestartClaw)
	ds.RegisterHandle("uninstallClaw", s.UninstallClaw)
	ds.RegisterHandle("testApiConnection", s.TestApiConnection)
	ds.RegisterHandle("updateModelConfig", s.UpdateModelConfig)
	ds.RegisterHandle("updateMemoryConfig", s.UpdateMemoryConfig)
	ds.RegisterHandle("installNodeEnv", s.InstallNodeEnv)
	ds.RegisterHandle("getClawWsInfo", s.GetClawWsInfo)
	ds.RegisterHandle("getWsProxyStatus", s.GetWsProxyStatus)
	ds.RegisterHandle("toggleWsProxy", s.ToggleWsProxy)
	ds.RegisterHandle("getRecentLogs", s.GetRecentLogs)
	ds.RegisterHandle("getOpenClawConfigDirPath", s.GetOpenClawConfigDirPath)


	// Agent管理接口
	ds.RegisterHandle("getAgentFiles", s.GetAgentFiles)
	ds.RegisterHandle("saveAgentFile", s.SaveAgentFile)
	ds.RegisterHandle("resetAgentFile", s.ResetAgentFile)
	ds.RegisterHandle("getAgentTemplates", s.GetAgentTemplates)
	ds.RegisterHandle("applyAgentTemplate", s.ApplyAgentTemplate)

	// 多Agent管理接口
	ds.RegisterHandle("listAgents", s.ListAgents)
	ds.RegisterHandle("createAgent", s.CreateAgent)
	ds.RegisterHandle("updateAgent", s.UpdateAgent)
	ds.RegisterHandle("deleteAgent", s.DeleteAgent)
	ds.RegisterHandle("getAgentDetail", s.GetAgentDetail)
	ds.RegisterHandle("getConfiguredModels", s.GetConfiguredModels)

	// 通道管理接口
	ds.RegisterHandle("getChannels", s.GetChannels)
	ds.RegisterHandle("saveChannel", s.SaveChannel)
	ds.RegisterHandle("deleteChannel", s.DeleteChannel)
	ds.RegisterHandle("toggleChannel", s.ToggleChannel)
	ds.RegisterHandle("approvePairing", s.ApprovePairing)

	// 技能管理接口
	ds.RegisterHandle("searchSkills", s.SearchSkills)
	ds.RegisterHandle("inspectSkill", s.InspectSkill)
	ds.RegisterHandle("installSkill", s.InstallSkill)
	ds.RegisterHandle("uninstallSkill", s.UninstallSkill)
	ds.RegisterHandle("listInstalledSkills", s.ListInstalledSkills)
	ds.RegisterHandle("exploreSkills", s.ExploreSkills)
	ds.RegisterHandle("listBuiltinSkills", s.ListBuiltinSkills)
	ds.RegisterHandle("installBuiltinSkill", s.InstallBuiltinSkill)
	ds.RegisterHandle("uninstallBuiltinSkill", s.UninstallBuiltinSkill)
	ds.RegisterHandle("toggleBuiltinSkill", s.ToggleBuiltinSkill)
	ds.RegisterHandle("getActiveSkillCount", s.GetActiveSkillCount)
	ds.RegisterHandle("isClawHubInstalled", s.IsClawHubInstalled)
	ds.RegisterHandle("installClawHub", s.InstallClawHub)
	ds.RegisterHandle("listEnvVars", s.ListEnvVars)
	ds.RegisterHandle("saveEnvVars", s.SaveEnvVars)
	ds.RegisterHandle("getModelsConfig", s.GetModelsConfig)
	ds.RegisterHandle("saveModelsConfig", s.SaveModelsConfig)
	ds.RegisterHandle("getSkillsDir", s.GetSkillsDir)
	ds.RegisterHandle("listRecommendedSkills", s.ListRecommendedSkills)
	ds.RegisterHandle("installRecommendedSkill", s.InstallRecommendedSkill)
	ds.RegisterHandle("uninstallRecommendedSkill", s.UninstallRecommendedSkill)

	// 定时任务接口
	ds.RegisterHandle("cronStatus", s.CronStatus)
	ds.RegisterHandle("listCronJobs", s.ListCronJobs)
	ds.RegisterHandle("addCronJob", s.AddCronJob)
	ds.RegisterHandle("editCronJob", s.EditCronJob)
	ds.RegisterHandle("removeCronJob", s.RemoveCronJob)
	ds.RegisterHandle("enableCronJob", s.EnableCronJob)
	ds.RegisterHandle("disableCronJob", s.DisableCronJob)
	ds.RegisterHandle("runCronJob", s.RunCronJob)
	ds.RegisterHandle("getCronRuns", s.GetCronRuns)

	// Agent 实时状态接口
	ds.RegisterHandle("getAgentStatuses", s.GetAgentStatuses)
	ds.RegisterHandle("getAgentTokenStats", s.GetAgentTokenStats)
	ds.RegisterHandle("getGatewayHealth", s.GetGatewayHealth)
	ds.RegisterHandle("testPlatformConn", s.TestPlatformConn)
	ds.RegisterHandle("getAgentSessions", s.GetAgentSessions)
	ds.RegisterHandle("getAgentActivity", s.GetAgentActivity)
}

// Start 启动RPC服务器
func (s *Server) Start(sockPath string) error {
	_ = os.MkdirAll(sockPath, 0o755)
	ds := simplejrpc.NewDefaultServer(
		gsock.WithJsonRpcSimpleServiceHandler(gsock.NewJsonRpcSimpleServiceHandler()),
	)
	gpath.GmCfgPath = "./"
	gi18n.Instance().SetPath("./i18n")
	core.InitContainer(config.WithConfigEnvFormatterOptionFunc("test"))

	s.RegisterHandles(ds)

	return ds.StartServer(sockPath)
}

// Ping 心跳检测
func (s *Server) Ping(req *gsock.Request) (any, error) {
	// core.Container.Log().Info("收到Ping请求")
	return "pong", nil
}

func (s *Server) Hello(req *gsock.Request) (any, error) {
	// core.Container.Log().Info("收到Ping请求")
	return "Hello", nil
}

func (s *Server) Hello_ERROR(req *gsock.Request) (any, error) {
	// core.Container.Log().Info("收到Ping请求")
	return nil, errors.New("error msg")
}
