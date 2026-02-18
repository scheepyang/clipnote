# ai-review 浮動視窗功能規劃

<a id="目錄"></a>

## 目錄

- [背景與目標](#背景與目標)
- [前置變更參考](#前置變更參考)
- [修改範圍](#修改範圍)
- [實作方案](#實作方案)
  - [浮窗渲染函式](#浮窗渲染函式)
  - [Model 狀態擴充](#model-狀態擴充)
  - [快捷鍵浮窗取代全屏 help](#快捷鍵浮窗取代全屏-help)
  - [註記查看浮窗](#註記查看浮窗)
  - [精簡狀態列](#精簡狀態列)
  - [渲染流程](#渲染流程)
- [新增快捷鍵](#新增快捷鍵)
- [驗證步驟](#驗證步驟)

---

## 背景與目標

ai-review 的底部狀態列塞了 7 個快捷鍵提示（共 67 字元），在窄視窗下被 `truncateLine()` 截斷，導致快捷鍵文字和右側行號資訊黏在一起，難以閱讀。

此外，右側面板的註記內容受限於面板寬度而被截斷，無法查看完整文字。

本次修改目標：

1. 快捷鍵提示改為浮窗顯示（按 `?` 觸發），取代目前的全屏 help
2. 新增註記內容浮窗，可查看完整文字
3. 精簡狀態列，只保留最常用的操作提示

[回到目錄](#目錄)

---

## 前置變更參考

本功能建立在以下已完成的變更之上：

- **註記匯出前綴標記**（`marks.go`）：在 `ExportMarks()` 中為使用者註記添加 `[Q]` 前綴，透過 `noteExportPrefix` 常數控制。匯出格式為 `> [Q] 使用者的追問`。此變更僅影響匯出格式，不影響本次浮窗功能的 UI 顯示。
- **標記符號顏色區分**（[feat-mark-color-distinction.md](feat-mark-color-distinction.md)）：純標記顯示粉紅色 `●`，有筆記的標記顯示黃色 `●`，方便一眼辨識。

[回到目錄](#目錄)

---

## 修改範圍

| 檔案 | 修改內容 |
|------|----------|
| `view.go` | 新增 `renderOverlay()` 浮窗渲染函式，修改 `View()`、`renderStatusBar()`，移除 `renderHelp()` 全屏邏輯 |
| `model.go` | 新增 `overlayKind` 型別與 `overlayType` 狀態欄位 |
| `update.go` | 修改 `?` 鍵行為，新增 `v` 鍵觸發註記查看 |
| `keys.go` | 新增 `v` 鍵綁定 |

[回到目錄](#目錄)

---

## 實作方案

### 浮窗渲染函式

新增 `renderOverlay(base, content string, width, height int) string`：

- 將 `content` 渲染成帶邊框的 lipgloss 面板
- 置中在 `base` 上方，逐行替換中間區域的字元
- 邊框使用既有的 `borderStyle`，背景色 `"235"`（深灰）提高辨識度

視覺效果：

```
┌──────────────────────┐
│  ai-review shortcuts │
│                      │
│  r  capture          │
│  m  mark             │
│  c  note             │
│  ...                 │
│                      │
│  press any key...    │
└──────────────────────┘
```

[回到目錄](#目錄)

### Model 狀態擴充

在 `model.go` 新增浮窗狀態管理：

```go
type overlayKind int

const (
    overlayNone overlayKind = iota
    overlayHelp
    overlayNote
)
```

`Model` struct 新增欄位：

```go
overlayType overlayKind
```

[回到目錄](#目錄)

### 快捷鍵浮窗取代全屏 help

目前按 `?` 會呼叫 `renderHelp()` 接管整個畫面。改為：

- `update.go`：按 `?` 時切換 `overlayType` 為 `overlayHelp`（取代 `showHelp`）
- `view.go`：不再用 `showHelp` 判斷全屏，改用 `renderOverlay()` 疊加浮窗

快捷鍵內容沿用現有 `renderHelp()` 的文字，只是改變呈現方式。

[回到目錄](#目錄)

### 註記查看浮窗

游標在有註記的標記行上時，按 `v` 開啟浮窗顯示完整註記：

- `update.go`：按 `v` 時檢查當前行是否有 note，有則設 `overlayType` 為 `overlayNote`
- `view.go`：`overlayNote` 時渲染該 note 的完整文字於浮窗中

[回到目錄](#目錄)

### 精簡狀態列

```go
// 修改前（7 項，67 字元）
leftText := "  [?]help [q]quit [r]capture [m]mark [c]note [S]export [P]paste"

// 修改後（4 項，42 字元）
leftText := "  ? help | q quit | m mark | S export"
```

其餘快捷鍵（`r` capture、`c` note、`P` paste）可透過 `?` 浮窗查看。

[回到目錄](#目錄)

### 渲染流程

`View()` 修改後的邏輯：

```go
func (m Model) View() string {
    // 1. 正常渲染 panels + statusBar
    result := lipgloss.JoinVertical(...)

    // 2. 如果有浮窗，覆蓋上去
    if m.overlayType != overlayNone {
        content := m.overlayContent()
        result = renderOverlay(result, content, m.width, m.height)
    }
    return result
}
```

`inputMode` 和 `captureInput` 的狀態列行為維持不變。

[回到目錄](#目錄)

---

## 新增快捷鍵

| 鍵 | 動作 |
|----|------|
| `v` | 查看當前行的完整註記（浮窗） |

`?` 鍵功能不變（顯示快捷鍵），只是改為浮窗呈現。

[回到目錄](#目錄)

---

## 驗證步驟

1. `cd ~/Desktop/ai/ai-review && go build -o ai-review .`
2. 按 `?` 確認浮窗顯示快捷鍵（非全屏），按任意鍵關閉
3. 標記一行並加上長註記，游標移到該行按 `v` 確認浮窗顯示完整註記
4. 確認狀態列精簡後在窄視窗下不截斷
5. 確認浮窗關閉後畫面完全恢復正常

[回到目錄](#目錄)
