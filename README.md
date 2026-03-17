# 🚀 OpenClaw-Turbo

🦞 **Your Personal AI Assistant**  
An All-in-One Management Application Built on the [GMSSH](https://web.gmssh.com) Ecosystem 

[![Chinese](https://img.shields.io/badge/Lang-%E7%AE%80%E4%BD%93%E4%B8%AD%E6%96%87-brightgreen)](https://github.com/GMSSH/OpenClaw-Turbo/blob/main/README.cn.md)
[![English](http://img.shields.io/badge/Lang-English-blue)](https://github.com/GMSSH/OpenClaw-Turbo/blob/main/README.md)
---

**More Secure · Better Chinese Understanding · Optimized for Domestic IM Infrastructure**

OpenClaw-Turbo dramatically simplifies the deployment and management of AI agents.  
Whether you are a performance-focused source-build enthusiast or a container-first DevOps engineer, OpenClaw-Turbo provides an intuitive graphical interface that frees you from complex terminal configurations.

---

### 📥 Download & Usage

You can quickly deploy on a Linux server or Linux workstation by following these steps:

### 1️⃣ Access the Platform  
Log in to the GMSSH Web Desktop:  https://web.gmssh.com  

### 2️⃣ Get the Client  
Download and install the GMSSH client:  https://www.gm.cn/client-download  

### 3️⃣ One-Click Installation  
Open **App Center** → Search for **"GMClaw"** → Install and start instantly.

> <img width="1299" height="702" alt="截屏2026-03-03 18 05 18" src="https://github.com/user-attachments/assets/ac17ee24-df2f-42e6-a7d8-a1c6269514bb" />

---

# ⚡️ Source Code & Development

If you would like to perform secondary development or debug the source code, please follow the steps below:

#### 1. Clone the Repository

```bash
git clone https://github.com/GMSSH/OpenClaw-Turbo
cd OpenClaw-Turbo

```

#### 2. Backend Development

```bash
cd backend
go mod tidy
go run main.go

```

#### 3. Frontend Development

```bash
cd ../frontend  # Adjust if your frontend directory name is different
pnpm install
pnpm run dev

```
👉 Please refer to the [GMSSH Developer Documentation](https://doc-dev.gmssh.com/) to start debugging within the GM environment.
---

### ✨ Core Features

#### 📦 Minimal Deployment · Self-Healing Environment

Dual Installation Modes
- Fully supports： **Docker container deployment** and Supports **native Shell script build deployment**
### Automatic Environment Detection：
- Built-in pre-installation checks
- Automatically detects missing dependencies
- Auto-completes required runtime environments
- Ensures a zero-error installation experience
  
#### 🛠️ Visual Configuration Center

### Real-Time Console：
- One-click gateway status monitoring  
- Token authentication visibility  
- LAN binding mode display  
- Runtime duration tracking  
- Web & communication port information  
### AI Model Management：
- Visual configuration for models such as **DeepSeek-Chat**
- Easy API key modification
- 128K context window configuration
- Output limit control
### Agent Persona Management：
- Built-in visual editor
- Customize:
  - **Identity** (Who am I)
  - **User** (Who are you)
  - **Soul** (How I communicate)
  - Deeply personalize your AI’s identity, memory, and behavior.

#### 🔗 Multi-Channel Social Integration (Visual)

Connect mainstream messaging platforms without editing configuration files:

- WeCom / QQ Bots
- DingTalk / Feishu

**Coming Soon:**
- WhatsApp  
- Telegram  
- iMessage  

---

#### 🧠 Skills & Capability Center

* **Built-in Skill Management**：Intuitively enable or install 51+ built-in skills, including:
  - 1Password
  - Apple Notes
  - ClawHub
  - And many more
* **Skill Marketplace**：Search community marketplace extensions (e.g., GitHub, Video, Notion)
- Install additional capabilities with one click

#### 📅 Scheduled Tasks (Cron)

* **Visual Task Scheduling**：Create automation tasks using:
  - Cron expressions
  - Fixed time intervals
  - One-time triggers
* **Flexible Task Actions**：Schedule text commands (e.g., morning summaries, routine cleanup)
Trigger specific Agent behaviors automatically

---

### 🖥️ Interface Preview

 **Dashboard Overview**  Runtime status at a glance、Model parameters clearly displayed、Gateway network information fully visible
 
<img width="1106" height="907" alt="控制台英文" src="https://github.com/user-attachments/assets/49119cac-034c-4ba0-9390-5817223fb43d" />

 **hird-Party Platform Integration**  - One-click access to:QQ、WeCom (Enterprise WeChat)、DingTalk、Feishu
 
<img width="1109" height="914" alt="三方平台接入英文" src="https://github.com/user-attachments/assets/f22d526a-c224-4c09-b7ba-72f4a2d2bc35" />

 **Persona Management**  Refine and shape your AI’s identity tags just like editing a document.
 
<img width="1108" height="920" alt="人格管理英文" src="https://github.com/user-attachments/assets/35ea8634-7253-45fe-829c-122f198fa5d3" />

 **Capability Center**  51+ built-in skills ready to use with one click ,Expand functionality through the integrated marketplace  
 
<img width="1112" height="915" alt="能力中心英文" src="https://github.com/user-attachments/assets/70d75376-30fc-44e4-9731-1eeb652418b4" />

 **Scheduled Tasks**  Powerful automation task editor with support for recurring command triggers.
 
<img width="1103" height="910" alt="定时任务英文" src="https://github.com/user-attachments/assets/97191d8a-e542-4ae9-a06c-a57e80f3f595" />
---

### 🌟Coming Soon: Cyber Employees

We are developing a **Cyber Employees** module based on the GMSSH-OpenClaw ecosystem ——A next-generation intelligent agent framework designed to deploy specialized virtual digital employees for enterprise workflows.

Stay tuned 🚀
<img width="1098" height="912" alt="赛博员工英文" src="https://github.com/user-attachments/assets/a760d9e1-de03-4bc6-b126-91d9ab180321" />

### 📜 License

**This project is licensed under the [GNU General Public License v3.0 (GPL-3.0)](https://www.google.com/search?q=https://www.gnu.org/licenses/gpl-3.0.html) .**

Key Terms Summary：

* **Free Use**：You may use this project for commercial purposes or modify it privately.
* **Source Code Disclosure**：If you distribute or publish a modified version of this project, you must release the source code under the same GPL-3.0 license.
* **Attribution Requirement**： The original copyright notice and license copy must be preserved.

---

### 🙏  Credits

The successful development of this project would not have been possible without the contributions of the open-source community. Special thanks to the following outstanding projects:

* **[openclaw-china](https://github.com/BytePioneer-AI/openclaw-china)**：Provided excellent IM plugin support, significantly enhancing integration capabilities with domestic social platforms.
* **[openclaw-zh](https://github.com/dongshuyan/openclaw-zh)**：Offered high-quality build source code and localization support, serving as an important reference for the infrastructure of this project.
---
## Star History


[![Star History Chart](https://api.star-history.com/svg?repos=GMSSH/OpenClaw-Turbo&type=date&legend=top-left)](https://www.star-history.com/#GMSSH/OpenClaw-Turbo&type=date&legend=top-left)
---
*Happy Coding!*

*The GMSSH Team*
