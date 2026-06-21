# KubeOps 设计系统

## 设计方向
渐变现代风 - 深色主题，渐变卡片，毛玻璃效果，现代感强

## 设计参数
- DESIGN_VARIANCE: 7
- MOTION_INTENSITY: 5
- VISUAL_DENSITY: 6

## 颜色系统

### 主色调
- 主色：#6366f1 (靛蓝)
- 主色渐变：linear-gradient(135deg, #6366f1, #8b5cf6)
- 成功：#22c55e
- 警告：#f59e0b
- 危险：#ef4444
- 信息：#3b82f6

### 背景色
- 主背景：#0f172a (深蓝黑)
- 卡片背景：rgba(30, 41, 59, 0.8)
- 毛玻璃：backdrop-filter: blur(12px)
- 悬浮：rgba(51, 65, 85, 0.5)

### 文本色
- 主文本：#f1f5f9
- 次要文本：#94a3b8
- 禁用文本：#475569

### 边框
- 默认：1px solid rgba(148, 163, 184, 0.1)
- 悬浮：1px solid rgba(99, 102, 241, 0.5)
- 聚焦：1px solid #6366f1

## 间距系统
- xs: 4px
- sm: 8px
- md: 16px
- lg: 24px
- xl: 32px
- 2xl: 48px

## 圆角
- sm: 6px
- md: 8px
- lg: 12px
- xl: 16px
- full: 9999px

## 阴影
- sm: 0 1px 2px rgba(0, 0, 0, 0.3)
- md: 0 4px 6px rgba(0, 0, 0, 0.4)
- lg: 0 10px 15px rgba(0, 0, 0, 0.5)
- glow: 0 0 20px rgba(99, 102, 241, 0.3)

## 字体
- 标题：Inter, system-ui, sans-serif (600-700)
- 正文：Inter, system-ui, sans-serif (400-500)
- 等宽：JetBrains Mono, monospace

## 动效
- 过渡：transition: all 0.2s ease
- 悬浮：transform: translateY(-2px)
- 聚焦：box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.3)
