# 修復 status bar 溢出右側 pane 邊界

> **前置文件：** [feat/paste-to-pane 開發規格](feat-paste-to-pane.md)

<a id="目錄"></a>

## 📑 目錄

- [問題描述](#問題描述)
- [修改檔案](#修改檔案)
- [修改方案](#修改方案)
- [驗證步驟](#驗證步驟)

---

## 問題描述

TUI 底部的 status bar 快捷鍵說明文字超出右側 pane 的可見範圍，最後的 `[q]quit` 被截斷。

**根本原因：** `renderStatusBar()`（`view.go:158`）中，當 `left + right` 總寬度超過 `m.width` 時，gap 被設為 0 但文字未截斷，導致內容溢出 pane 邊界。

**討論過的替代方案：**

- 換行顯示 — 可行但會佔用額外一行高度
- 新增「看更多」彈出視窗 — 過度設計，因為已有 `?` help overlay 提供完整按鍵說明

**選定方案：** 截斷 + 依賴現有 `?` help overlay。寬度不足時用 `…` 截斷左側 help 文字，右側行號/marks 資訊優先保留。使用者按 `?` 即可查看完整按鍵說明（`renderHelp()` 已實作於 `view.go:170`）。

[⬆ 回到目錄](#目錄)

---

## 修改檔案

- `view.go` — `renderStatusBar()` 函式（第 158-168 行）

[⬆ 回到目錄](#目錄)

---

## 修改方案

1. 先計算 `right`（行號 + marks 資訊）的寬度，此部分優先顯示
2. 計算左側可用寬度 = `m.width - rightWidth`
3. 若左側 help 文字超過可用寬度，使用 `truncateLine` 截斷（已有此工具函式在 `view.go:197`）
4. gap 計算邏輯保持不變

```go
func (m Model) renderStatusBar() string {
	leftText := "  [r]capture [m]mark [c]note [S]export [P]paste [?]help [q]quit"
	right := statusStyle.Render(fmt.Sprintf("L%d/%d  Marks: %d  ", m.cursorLine+1, len(m.lines), len(m.marks)))

	rightW := lipgloss.Width(right)
	maxLeft := m.width - rightW
	if maxLeft < 0 {
		maxLeft = 0
	}
	left := statusStyle.Render(truncateLine(leftText, maxLeft))

	gap := m.width - lipgloss.Width(left) - rightW
	if gap < 0 {
		gap = 0
	}
	return left + strings.Repeat(" ", gap) + right
}
```

[⬆ 回到目錄](#目錄)

---

## 驗證步驟

1. `cd ~/Desktop/ai/clipnote && go build -o clipnote .` 確認編譯通過
2. `./clipnote` 啟動後，縮小終端機寬度，確認 status bar 不再溢出且 `…` 截斷正確顯示
3. 按 `?` 確認 help overlay 仍正常顯示完整按鍵說明

[⬆ 回到目錄](#目錄)
