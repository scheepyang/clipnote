# clipnote 包裝為 Claude Code Plugin

<a id="目錄"></a>

## 目錄

- [背景說明](#背景說明)
- [執行順序](#執行順序)
- [目錄結構變更](#目錄結構變更)
- [步驟 1：推到 GitHub private repo](#步驟-1推到-github-private-repo)
- [步驟 2：建立 GitHub Actions workflow](#步驟-2建立-github-actions-workflow)
- [步驟 3：建立 Plugin manifest](#步驟-3建立-plugin-manifest)
- [步驟 4：建立 Marketplace 配置](#步驟-4建立-marketplace-配置)
- [步驟 5：建立 conductor.json](#步驟-5建立-conductorjson)
- [步驟 6：建立 Skill](#步驟-6建立-skill)
- [步驟 7：建立 hooks.json](#步驟-7建立-hooksjson)
- [步驟 8：建立 setup.sh](#步驟-8建立-setupsh)
- [步驟 9：更新 .gitignore](#步驟-9更新-gitignore)
- [修改檔案清單](#修改檔案清單)
- [驗證](#驗證)

---

## 背景說明

將 clipnote（Go TUI 標註工具）包裝成 Claude Code Plugin，目標是分發給其他使用者。參照 claude-mem plugin 結構，採用兩層目錄架構（頂層 marketplace + `plugin/` 子目錄）。

使用 GitHub Actions 預編譯 macOS/Linux binary，使用者不需要 Go 環境。先推到 GitHub private repo 建好 CI，測試通過後再切 public。

[回到目錄](#目錄)

---

## 執行順序

1. 推到 GitHub private repo
2. 建立 GitHub Actions CI，push tag 觸發跨平台編譯
3. 建立 plugin 目錄結構與所有檔案
4. 本地測試 `claude --plugin-dir`
5. 測試通過後切 public

[回到目錄](#目錄)

---

## 目錄結構變更

在現有 clipnote repo 中新增以下檔案：

```
clipnote/
├── .claude-plugin/
│   ├── plugin.json              <-- 新增：Plugin manifest
│   └── marketplace.json         <-- 新增：Marketplace 配置
├── conductor.json               <-- 新增：setup script 定義
├── plugin/
│   ├── .claude-plugin/
│   │   └── plugin.json          <-- 新增：Plugin 內的 manifest
│   ├── skills/
│   │   └── launch/
│   │       └── SKILL.md         <-- 新增：/clipnote:launch 斜線命令
│   ├── hooks/
│   │   └── hooks.json           <-- 新增：Setup + SessionStart hook
│   ├── scripts/
│   │   └── setup.sh             <-- 新增：偵測平台、下載 binary
│   └── bin/                     <-- gitignore，setup.sh 下載 binary 放這裡
├── .github/
│   └── workflows/
│       └── release.yml          <-- 新增：GitHub Actions 跨平台編譯
├── main.go, session.go ...      <-- 不動
└── bin/                         <-- 本地開發用（已在 .gitignore）
```

[回到目錄](#目錄)

---

## 步驟 1：推到 GitHub private repo

- 建立 GitHub private repo `clipnote`
- 推送現有程式碼

[回到目錄](#目錄)

---

## 步驟 2：建立 GitHub Actions workflow

新增 `.github/workflows/release.yml`：

- 觸發條件：push tag `v*`
- 矩陣編譯：`GOOS=darwin,linux` x `GOARCH=amd64,arm64`
- 上傳 4 個 binary 到 GitHub Release
- 推送後 push tag `v0.1.0` 測試 CI

[回到目錄](#目錄)

---

## 步驟 3：建立 Plugin manifest

新增 `.claude-plugin/plugin.json`（頂層）和 `plugin/.claude-plugin/plugin.json`（plugin 內），內容相同：

```json
{
  "name": "clipnote",
  "version": "0.1.0",
  "description": "AI CLI output annotation tool - mark, annotate, and export AI responses in a tmux split pane",
  "author": { "name": "nevertomica" },
  "keywords": ["review", "annotation", "tmux", "tui"]
}
```

[回到目錄](#目錄)

---

## 步驟 4：建立 Marketplace 配置

新增 `.claude-plugin/marketplace.json`：

```json
{
  "name": "nevertomica",
  "owner": { "name": "nevertomica" },
  "plugins": [
    {
      "name": "clipnote",
      "version": "0.1.0",
      "source": "./plugin",
      "description": "AI CLI output annotation tool - mark, annotate, and export AI responses in a tmux split pane"
    }
  ]
}
```

[回到目錄](#目錄)

---

## 步驟 5：建立 conductor.json

新增 `conductor.json`：

```json
{
  "scripts": {
    "setup": "bash plugin/scripts/setup.sh"
  }
}
```

[回到目錄](#目錄)

---

## 步驟 6：建立 Skill

新增 `plugin/skills/launch/SKILL.md`：

- 自動觸發（不設 `disable-model-invocation`）
- description 限定在使用者明確要求啟動時觸發
- 內容指示 Claude 用 Bash 執行 `${CLAUDE_PLUGIN_ROOT}/bin/clipnote` 啟動 tmux session
- 附上快捷鍵說明供 Claude 告知使用者

```yaml
---
name: launch
description: Launch clipnote tmux annotation session. Use when
  the user explicitly asks to start, open, or launch clipnote.
---
```

[回到目錄](#目錄)

---

## 步驟 7：建立 hooks.json

新增 `plugin/hooks/hooks.json`，Setup + SessionStart 兩個 hook 都觸發 setup.sh：

- **Setup**：plugin 安裝時執行一次，首次下載 binary
- **SessionStart**：每次啟動 Claude Code 時檢查 binary 是否存在

```json
{
  "hooks": {
    "Setup": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "bash ${CLAUDE_PLUGIN_ROOT}/scripts/setup.sh",
            "timeout": 120
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "matcher": "startup|clear|compact",
        "hooks": [
          {
            "type": "command",
            "command": "bash ${CLAUDE_PLUGIN_ROOT}/scripts/setup.sh",
            "timeout": 120
          }
        ]
      }
    ]
  }
}
```

[回到目錄](#目錄)

---

## 步驟 8：建立 setup.sh

新增 `plugin/scripts/setup.sh`：

- 檢查 tmux 是否安裝（唯一的環境依賴）
- 檢查 binary 是否已存在，存在就跳過
- 偵測平台（`uname -s` + `uname -m`）
- 從 GitHub Release 下載對應的預編譯 binary 到 `${CLAUDE_PLUGIN_ROOT}/bin/`
- `chmod +x`
- 不依賴 Go 環境；想自行編譯的使用者參照 README

[回到目錄](#目錄)

---

## 步驟 9：更新 .gitignore

加入 `plugin/bin/`，避免下載的 binary 進入版本控制。

[回到目錄](#目錄)

---

## 修改檔案清單

| 檔案 | 動作 | 說明 |
|------|------|------|
| `.github/workflows/release.yml` | 新增 | CI 跨平台編譯 |
| `.claude-plugin/plugin.json` | 新增 | Plugin manifest（頂層） |
| `.claude-plugin/marketplace.json` | 新增 | Marketplace 配置 |
| `conductor.json` | 新增 | setup script 定義 |
| `plugin/.claude-plugin/plugin.json` | 新增 | Plugin manifest（plugin 內） |
| `plugin/skills/launch/SKILL.md` | 新增 | 斜線命令定義（自動觸發） |
| `plugin/hooks/hooks.json` | 新增 | Setup + SessionStart hook |
| `plugin/scripts/setup.sh` | 新增 | 偵測平台、下載 binary |
| `.gitignore` | 修改 | 加入 `plugin/bin/` |

[回到目錄](#目錄)

---

## 驗證

```bash
# 1. 推送 tag 觸發 GitHub Actions，確認 Release 產出 4 個 binary
git tag v0.1.0 && git push origin v0.1.0

# 2. 測試 setup.sh 能正確偵測平台並下載 binary
bash plugin/scripts/setup.sh
ls -la plugin/bin/clipnote

# 3. 本地測試 plugin 載入
claude --plugin-dir ~/Desktop/ai/clipnote

# 4. 在 Claude Code 中測試 /clipnote:launch
```

[回到目錄](#目錄)
