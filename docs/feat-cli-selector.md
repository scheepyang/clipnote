# 實作多 CLI 選擇選單

<a id="目錄"></a>

## 目錄

- [背景與目標](#背景與目標)
- [前置變更參考](#前置變更參考)
- [實作方案](#實作方案)
  - [修改檔案](#修改檔案)
  - [步驟 1：新增 selector.go](#步驟-1新增-selectorgo)
  - [步驟 2：修改 main.go 的 runLauncher](#步驟-2修改-maingo-的-runlauncher)
  - [步驟 3：runSelector 函式](#步驟-3runselector-函式)
- [驗證](#驗證)

---

## 背景與目標

clipnote 啟動時會透過 `detectCLIs()` 掃描系統上已安裝的 AI CLI（claude, gemini, codex, aider），但偵測到多個 CLI 時直接取第一個，沒有讓使用者選擇。`main.go:47` 有 TODO 標記此功能未完成。

**目標：** 偵測到多個 CLI 時，顯示互動式選擇選單讓使用者挑選要啟動的 CLI。

[回到目錄](#目錄)

---

## 前置變更參考

- [feat-reopen-pane.md](feat-reopen-pane.md) — 重新開啟 pane 功能規劃（上一份計劃）

[回到目錄](#目錄)

---

## 實作方案

使用 `bubbletea` 建立簡易選擇 TUI（專案已引用 bubbletea + bubbles，不需新增相依）。

### 修改檔案

| 檔案 | 動作 | 說明 |
|------|------|------|
| `main.go` | 修改 | `runLauncher()` 多 CLI 時啟動選擇 TUI |
| `selector.go` | 新增 | 選擇選單的 bubbletea Model |

### 步驟 1：新增 selector.go

建立 `selectorModel` 結構，包含：

- `choices []string` — 偵測到的 CLI 清單
- `cursor int` — 目前游標位置
- `selected string` — 使用者選定的 CLI

按鍵綁定：

- `j/k` 或方向鍵上下移動
- `Enter` 確認選擇
- `q/Esc` 退出

View 呈現方式：

```
Select AI CLI to launch:

  > claude
    gemini
    codex
    aider

[j/k] move  [enter] select  [q] quit
```

### 步驟 2：修改 main.go 的 runLauncher

```go
func runLauncher() {
    clis := detectCLIs()
    if len(clis) == 0 {
        fmt.Fprintln(os.Stderr, "No installed AI CLI found (claude, gemini, codex, aider)")
        os.Exit(1)
    }

    var cli string
    if len(clis) == 1 {
        cli = clis[0]
    } else {
        cli = runSelector(clis)  // 啟動選擇 TUI
        if cli == "" {
            return  // 使用者按 q 取消
        }
    }

    if err := launchSession(cli); err != nil {
        fmt.Fprintf(os.Stderr, "Launch failed: %v\n", err)
        os.Exit(1)
    }
}
```

### 步驟 3：runSelector 函式

放在 `selector.go` 中，啟動 bubbletea program 並回傳使用者選擇的 CLI 名稱。不使用 AltScreen（選單很小，不需要全螢幕）。

[回到目錄](#目錄)

---

## 驗證

```bash
cd ~/Desktop/ai/clipnote && go build -o bin/clipnote .
# 啟動後應顯示選擇選單（前提：系統有多個 CLI）
./bin/clipnote
```

[回到目錄](#目錄)
