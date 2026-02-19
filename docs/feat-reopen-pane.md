# 為 clipnote tmux session 綁定 prefix+a 重新開啟 annotation pane

<a id="目錄"></a>

## 目錄

- [背景與目標](#背景與目標)
- [相關功能文件](#相關功能文件)
- [修改範圍](#修改範圍)
- [實作方案](#實作方案)
- [驗證步驟](#驗證步驟)

---

## 背景與目標

右側 annotation pane 按 `q` 退出後，pane 關閉，使用者只剩左側 CLI。目前沒有方法在不重啟整個 session 的情況下重新開啟右側 pane。

透過在 `session.go` 的 `launchSession()` 中加入 `tmux bind-key`，讓使用者可以按 `Ctrl+b a` 重新開啟 annotation pane。

[⬆ 回到目錄](#目錄)

---

## 相關功能文件

- [feat-mark-color-distinction.md](feat-mark-color-distinction.md)
- [feat-overlay-window.md](feat-overlay-window.md)

[⬆ 回到目錄](#目錄)

---

## 修改範圍

| 檔案 | 修改內容 |
|------|----------|
| `session.go` | 在 pane border 設定後新增 `tmux bind-key` 呼叫 |

[⬆ 回到目錄](#目錄)

---

## 實作方案

在 `launchSession()` 中，於 mouse/border 設定之後、attach 之前加入：

```go
// bind prefix+a to reopen annotation pane if closed
// 用 sh -c 包裝，避免 tmux 重新解析時誤讀 --internal-watch 的 --
reopenCmd := fmt.Sprintf("%s --internal-watch %s:0.0", self, sessionName)
exec.Command("tmux", "bind-key", "a",
    "split-window", "-h", "-t", sessionName+":0.0", "-l", "45%",
    "sh", "-c", reopenCmd).Run()
```

**注意事項：**

- `tmux bind-key` 是全域的（無法限定 session），但因為每次啟動都會 `kill-session` 重建，不會有衝突
- 綁定使用 `self`（當前執行檔路徑）確保重新開啟的 pane 使用同一個 binary
- `--internal-watch` 搭配 `sessionName+":0.0"` 讓新 pane 監聯左側第一個 pane
- **使用 `sh -c` 包裝**：tmux `bind-key` 在按鍵觸發時會重新解析命令字串，`--internal-watch` 的 `--` 前綴可能被 tmux 當作選項結束標記而靜默失敗。透過 `sh -c` 讓 shell 負責解析參數，繞過此問題

[⬆ 回到目錄](#目錄)

---

## 驗證步驟

```bash
cd ~/Desktop/ai/clipnote && go build -o clipnote . && ./clipnote
```

1. 右側 pane 按 `q` 退出 → pane 關閉
2. 在左側按 `Ctrl+b a` → 右側 annotation pane 重新開啟
3. 重新開啟的 pane 功能正常（可 capture、mark 等）

[⬆ 回到目錄](#目錄)
