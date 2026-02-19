# 區分 mark-only 與 mark-with-note 的標記符號顏色

<a id="目錄"></a>

## 目錄

- [背景與目標](#背景與目標)
- [修改範圍](#修改範圍)
- [實作細節](#實作細節)
- [相關文件](#相關文件)
- [驗證步驟](#驗證步驟)

---

## 背景與目標

`m`（純標記）和 `c`（標記+筆記）在左側面板都顯示相同的粉紅色 `●`，無法一眼辨識哪些行有附加筆記。

本次修改目標：用不同顏色區分兩種標記：

- **粉紅色 `●`**（色碼 212）：純標記（`m`）
- **黃色 `●`**（色碼 220）：有筆記的標記（`c`）

[⬆ 回到目錄](#目錄)

---

## 修改範圍

| 檔案 | 修改內容 |
|------|----------|
| `view.go` | 新增 `noteMarkSymbol` 變數，修改 `renderContent()` 根據是否有 note 選擇符號 |
| `marks.go` | 新增 `GetMark()` 方法，回傳指定行的 `*Mark`（無則回傳 nil） |

[⬆ 回到目錄](#目錄)

---

## 實作細節

### `view.go` — 新增 noteMarkSymbol

```go
noteMarkSymbol = lipgloss.NewStyle().
    Foreground(lipgloss.Color("220")).  // 黃色（有筆記）
    Render("●")
```

### `view.go` — renderContent() 符號選擇邏輯

```go
if mk := m.GetMark(i); mk != nil {
    if mk.Note != "" {
        mark = noteMarkSymbol + " "
    } else {
        mark = markSymbol + " "
    }
}
```

### `marks.go` — GetMark() 方法

```go
func (m Model) GetMark(line int) *Mark {
    for i := range m.marks {
        if m.marks[i].Line == line {
            return &m.marks[i]
        }
    }
    return nil
}
```

[⬆ 回到目錄](#目錄)

---

## 相關文件

- [feat-overlay-window.md](feat-overlay-window.md) — 浮動視窗功能（同期 TUI 改進）

[⬆ 回到目錄](#目錄)

---

## 驗證步驟

1. `cd ~/Desktop/ai/clipnote && go build -o clipnote . && ./clipnote`
2. 按 `m` 標記一行 → 粉紅色 ●
3. 按 `c` 標記另一行並輸入筆記 → 黃色 ●
4. 對已有純標記的行按 `c` 加筆記 → 符號變為黃色

[⬆ 回到目錄](#目錄)
