# feat/paste-to-pane 開發規格

> **前置文件：** [clipnote 專案解說與 Go 語言入門指南](project-guide.md)

<a id="目錄"></a>

## 目錄

- [功能概述](#功能概述)
- [異動檔案總覽](#異動檔案總覽)
- [各檔案異動細節](#各檔案異動細節)
  - [keys.go — 按鍵定義](#keysgo--按鍵定義)
  - [clipboard.go — 核心邏輯](#clipboardgo--核心邏輯)
  - [update.go — 按鍵處理](#updatego--按鍵處理)
  - [view.go — UI 更新](#viewgo--ui-更新)
  - [main.go — CLI 說明](#maingo--cli-說明)
- [架構設計理由](#架構設計理由)
  - [Bubbletea (Elm Architecture) 拆分原則](#bubbletea-elm-architecture-拆分原則)
  - [五步驟對應表](#五步驟對應表)
  - [clipboard.go 的職責歸屬](#clipboardgo-的職責歸屬)

---

## 功能概述

在 clipnote TUI 的 Browse 模式下，按 `P` 鍵將已標記（marked）的內容以 Markdown 格式貼入左側 tmux pane。

[回到目錄](#目錄)

---

## 異動檔案總覽

| 檔案 | 異動類型 | 說明 |
|------|----------|------|
| `keys.go` | 新增欄位 | 定義 `P` 鍵綁定 |
| `clipboard.go` | 新增函式 | 實作 paste-to-pane 核心邏輯 |
| `update.go` | 新增 case | 接線按鍵與函式 |
| `view.go` | 修改 | 狀態列與說明頁加入 `P` |
| `main.go` | 修改 | `--help` 加入 `P` 說明 |

[回到目錄](#目錄)

---

## 各檔案異動細節

### keys.go -- 按鍵定義

新增 `PasteToPane` 欄位，綁定 `P` 鍵。

### clipboard.go -- 核心邏輯

新增函式：

```go
func (m *Model) PasteMarksToPane() string
```

執行流程：

1. 呼叫 `ExportMarks()` 取得 Markdown 文字
2. `tmux set-buffer <text>` 寫入 tmux buffer
3. `tmux paste-buffer -t <left-pane>` 貼到左 pane

### update.go -- 按鍵處理

在 `handleBrowseMode` 的 switch 中新增：

```go
case key.Matches(msg, keys.PasteToPane):
```

呼叫 `PasteMarksToPane()` 執行貼上動作。

### view.go -- UI 更新

- 狀態列新增 `[P]paste`
- 說明頁新增 `P -- Paste marks to left pane`

### main.go -- CLI 說明

`--help` 輸出中加入 `P` 鍵的功能說明。

[回到目錄](#目錄)

---

## 架構設計理由

### Bubbletea (Elm Architecture) 拆分原則

本功能的檔案拆分遵循 Bubbletea 框架的 Elm Architecture 模式：

```
Model（資料） -> Update（邏輯） -> View（顯示）
```

新增一個按鍵功能，就是沿著這個架構走一遍。

### 五步驟對應表

| 步驟 | 檔案 | 為什麼要改 |
|------|------|------------|
| 1 | keys.go | 定義按鍵綁定（什麼鍵觸發什麼） |
| 2 | clipboard.go | 實作核心邏輯（實際做的事） |
| 3 | update.go | 把按鍵和邏輯接起來（按 P -> 呼叫函式） |
| 4 | view.go | 讓使用者看到這個功能存在（狀態列、說明頁） |
| 5 | main.go | `--help` 也要同步更新 |

步驟 1-3 是功能本身必要的，少一個就不能動。步驟 4-5 是 UI 告知，少了功能還是能用但使用者不知道有這個鍵。

### clipboard.go 的職責歸屬

核心邏輯放在 `clipboard.go` 而不是 `update.go`，是因為專案已經把 `CopyMarksToClipboard()` 放在那裡了。`PasteMarksToPane()` 性質相同（都是把 marks 送到外部），放一起保持一致。

[回到目錄](#目錄)
