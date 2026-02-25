# 🚀 OpenClaw-Turbo

**🦞 你的私人 AI 助手 (Personal AI Assistant) | 建立在 [GMSSH](https://web.gmssh.com) 生态基础上的全能管理应用**

---

**更安全 · 更懂中文 · 适配国内 IM 的基础设施**

**OpenClaw-Turbo** 极大地简化了 AI 智能体的部署与管理流程。无论你是追求极致性能的源码编译控，还是偏好容器化的 DevOps 工程师，OpenClaw-Turbo 都能为你提供直观的图形化界面，让你彻底告别繁琐的终端配置。

[🌐 官网](https://www.gmssh.com) • [English README](https://www.google.com/search?q=%23)

---

### 📥 下载与使用

您可以按照以下步骤在 Linux 服务器或 Linux 工作站上快速部署：

1. **访问平台**：登录 **GMSSH 桌面**（在线版地址：[https://web.gmssh.com](https://web.gmssh.com)）。
2. **获取客户端**：下载并安装 GMSSH 客户端 [https://www.gm.cn/client-download](https://www.gm.cn/client-download)。
3. **一键安装**：进入 **应用中心**，搜索 **“GMClaw”**，即可实现开箱即用。
<img width="1457" height="910" alt="image" src="https://github.com/user-attachments/assets/60ad00e5-4a4a-488f-bc2a-f3b3598d7515" />

---

### ⚡️ 源码下载与开发

如需进行二次开发或源码调试，请参考以下流程：

#### 1. 克隆仓库

```bash
git clone https://github.com/GMSSH/OpenClaw-Turbo
cd OpenClaw-Turbo

```

#### 2. 后端开发 (Backend)

```bash
cd backend
go mod tidy
go run main.go

```

#### 3. 前端开发 (Frontend)

```bash
cd ../frontend  # 假设前端目录名为 frontend 或 pnpm 所在目录
pnpm install
pnpm run dev

```
参考[GMSSH开发者文档](https://doc-dev.gmssh.com/)在GM环境中启动调试
---

### ✨ 核心特性

#### 📦 极简部署 · 环境自愈

* **双模安装**：完美支持 **Docker 容器化部署** 与 **原生 Shell 脚本编译部署**。
* **自动环境检测**：内置环境预检逻辑，自动识别并补全服务器缺失的依赖环境，确保安装流程零报错。

#### 🛠️ 可视化配置中心

* **实时控制台**：一键监控网关状态（Token 认证、LAN 绑定模式）、运行时间及 Web/通讯端口信息。
* **AI 模型管理**：图形化配置 **DeepSeek-Chat** 等模型，直观修改 API Key、128K 上下文窗口及输出限制。
* **Agent 人格管理**：内置编辑器，可视化编辑 **Identity (我是谁)**、**User (你是谁)** 及 **Soul (怎么聊)**，深度定制 AI 的身份、记忆与灵魂。

#### 🔗 多渠道社交接入（可视化）

* 无需修改配置文件，通过图形化界面即可快速接入主流社交平台：
* **企业微信 / QQ 机器人**
* **钉钉 / 飞书**
* *即将支持：WhatsApp、Telegram、iMessage*

---

#### 🧠 能力与技能中心 (Skills)

* **内置技能管理**：直观开启或安装 `1password`、`apple-notes`、`clawhub` 等 51 项内置技能。
* **技能市场**：支持从社区市场搜索（如 `github`, `video`, `notion`）并一键安装扩展能力。

#### 📅 定时任务 (Cron)

* **可视化调度**：支持通过 **Cron 表达式**、固定间隔或一次性触发创建自动化任务。
* **多样化请求**：可设置定时发送文本指令（如早间摘要、定时清理）或触发特定 Agent 行为。

---

### 🖥️ 界面预览

 **控制台**  运行状态、模型参数与网关网络一目了然 
<img width="1458" height="924" alt="image" src="https://github.com/user-attachments/assets/e762e4fa-0dd2-4a29-9de2-629542cd231a" />
 **三方平台接入**  一键接入QQ、企业微信、钉钉、飞书
<img width="1459" height="882" alt="image" src="https://github.com/user-attachments/assets/bcbdecb2-f62c-45e0-a28e-4a72b9f52a67" />
 **人格管理**  像写文档一样打磨 AI 的身份标签 
<img width="1464" height="915" alt="image" src="https://github.com/user-attachments/assets/b16ad2fd-dd95-4a83-a462-900998038411" />
 **能力中心**  内置 51+ 种技能点击即用，支持市场扩展 
<img width="1461" height="915" alt="image" src="https://github.com/user-attachments/assets/6f1e44bc-3b0c-4cf5-8e38-40d383654f5c" />
 **定时任务**  强大的自动化任务编辑弹窗，支持周期性指令触发 
<img width="1459" height="920" alt="image" src="https://github.com/user-attachments/assets/86a76088-ce22-4537-96f2-4d12ef117d64" />

---

### 🌟 即将到来：赛博员工 (Cyber Employees)

我们正在基于 **GMSSH-Openclaw 生态** 开发**“赛博员工”**模块——这是一种全新的智能体形态，旨在为企业流程部署专门化的虚拟数字员工。敬请期待！

### 🙏 致谢 (Credits)

本项目的顺利开发离不开开源社区的贡献，特别感谢以下优秀项目：

* **[openclaw-china](https://github.com/BytePioneer-AI/openclaw-china)**：提供了优秀的 IM 插件支持，极大丰富了国内社交平台的接入能力。
* **[openclaw-zh](https://github.com/dongshuyan/openclaw-zh)**：提供了高质量的编译源代码及汉化支持，为本项目的基础设施搭建提供了重要参考。

---
## Star History


[![Star History Chart](https://api.star-history.com/svg?repos=GMSSH/OpenClaw-Turbo&type=date&legend=top-left)](https://www.star-history.com/#GMSSH/OpenClaw-Turbo&type=date&legend=top-left)
---
*Happy Coding!*

*The GMSSH Team*
