---
name: remnawave-visual-style-guide
description: 'Design and implement UI components, pages, widgets, and layouts strictly adhering to the Remnawave Cyber Dark Glassmorphism design system. Use when styling or reviewing web applications, widgets, subscription pages, modals, cards, buttons, or dashboards requiring deep dark backgrounds (#161B23), 60x60px grid patterns, ambient glow radial gradients, backdrop-filter glassmorphism, Unbounded and Montserrat typography, neon Cyan (#22D3EE) accents, and metric info-blocks.'
argument-hint: 'UI component, page, or widget description'
user-invocable: true
---

# Remnawave Subscription Page — Visual Style Guide & Design System

Полная спецификация дизайн-системы и визуального стиля проекта [remnawave/subscription-page](https://github.com/remnawave/subscription-page).  
Этот документ предназначен для разработчиков и языковых моделей (LLM / AI-ассистентов) для точного воспроизведения визуального стиля в новых интерфейсах, виджетах и веб-приложениях.

---

## 1. Концепция и эстетика

- **Название стиля**: _Cyber Dark Glassmorphism / Modern Tech Minimalist_.
- **Основные принципы**:
  1. **Глубокий темный фон**: база `#161B23` (GitHub Dark / Slate) с фиксированной сеткой `60x60px`.
  2. **Атмосферное фоновое свечение (Ambient Glow)**: два мягких радиальных градиента (Violet и Cyan) с низкой прозрачностью, создающие эффект глубины.
  3. **Ультра-легкий глассморфизм (Glassmorphism)**: полупрозрачные фоны карточек `rgba(255, 255, 255, 0.02)`, тонкие границы `rgba(255, 255, 255, 0.08)`, размытие `backdrop-filter: blur(10px)`.
  4. **Неоновый акцентный фокус**: фирменный оттенок **Cyan (`#22D3EE`)** в качестве основного акцента для активных состояний, фокусов и индикаторов.
  5. **Контрастная современная типографика**: футуристичный дисплейный шрифт `Unbounded` для заголовков и геометричный чистый `Montserrat` для основного текста.
  6. **Центрированный поток (Single-Column Stream)**: максимальная ширина контентной зоны `800px`, мобильная адаптивность "из коробки".

---

## 2. Цветовая палитра и токены

### 2.1. Базовые фоновые цвета

| Токен                   | HEX / RGBA                  | Назначение                                                   |
| :---------------------- | :-------------------------- | :----------------------------------------------------------- |
| **`bg-canvas`**         | `#161B23`                   | Главный цвет фона страницы (Body Canvas)                     |
| **`bg-header`**         | `rgba(22, 27, 35, 0.80)`    | Липкая шапка (Sticky Header) с `backdrop-filter: blur(12px)` |
| **`bg-modal`**          | `rgba(22, 27, 35, 0.95)`    | Модальные окна с `backdrop-filter: blur(20px)`               |
| **`card-bg`**           | `rgba(255, 255, 255, 0.02)` | Базовый фон карточек и интерактивных плашек                  |
| **`card-border`**       | `rgba(255, 255, 255, 0.08)` | Стандартная тонкая граница                                   |
| **`card-border-hover`** | `rgba(34, 211, 238, 0.30)`  | Граница карточки при наведении (Cyan Hover)                  |

### 2.2. Нейтральная темная шкала (GitHub Dark / Primer)

| Ступень      | HEX       | Применение                                        |
| :----------- | :-------- | :------------------------------------------------ |
| **`dark.0`** | `#c9d1d9` | Высококонтрастный текст                           |
| **`dark.1`** | `#b1bac4` | Основной вторичный текст                          |
| **`dark.2`** | `#8b949e` | Приглушенный текст (`text-dimmed` / `text-muted`) |
| **`dark.3`** | `#6e7681` | Неактивные иконки и плейсхолдеры                  |
| **`dark.4`** | `#484f58` | Разделители и второстепенные бордеры              |
| **`dark.5`** | `#30363d` | Границы контролов и инпутов                       |
| **`dark.6`** | `#21262d` | Промежуточный темный слой элементов               |
| **`dark.7`** | `#161b22` | Подложка для выпадающих списков и модалок         |
| **`dark.8`** | `#0d1117` | Глубокий темный фон                               |
| **`dark.9`** | `#010409` | Абсолютно темный                                  |

### 2.3. Акцентные цвета и функциональные градиенты

| Цвет               | RGB             | HEX       | Soft Gradient (Cards / InfoBlock)                                                     | Solid Gradient (ThemeIcon / Status)                 |
| :----------------- | :-------------- | :-------- | :------------------------------------------------------------------------------------ | :-------------------------------------------------- |
| **Cyan (Primary)** | `34, 211, 238`  | `#22D3EE` | `linear-gradient(135deg, rgba(34, 211, 238, 0.1) 0%, rgba(6, 182, 212, 0.05) 100%)`   | `linear-gradient(135deg, #182B36 0%, #152229 100%)` |
| **Blue**           | `34, 139, 230`  | `#228BE6` | `linear-gradient(135deg, rgba(34, 139, 230, 0.1) 0%, rgba(28, 126, 214, 0.05) 100%)`  | `linear-gradient(135deg, #182638 0%, #151F2C 100%)` |
| **Teal**           | `32, 201, 151`  | `#20C997` | `linear-gradient(135deg, rgba(32, 201, 151, 0.1) 0%, rgba(18, 184, 134, 0.05) 100%)`  | `linear-gradient(135deg, #182B2A 0%, #152220 100%)` |
| **Green**          | `64, 192, 87`   | `#40C057` | `linear-gradient(135deg, rgba(64, 192, 87, 0.1) 0%, rgba(55, 178, 77, 0.05) 100%)`    | `linear-gradient(135deg, #1A2B20 0%, #17211B 100%)` |
| **Yellow**         | `250, 176, 5`   | `#FAB005` | `linear-gradient(135deg, rgba(250, 176, 5, 0.1) 0%, rgba(245, 159, 0, 0.05) 100%)`    | `linear-gradient(135deg, #2A2817 0%, #211F15 100%)` |
| **Orange**         | `253, 126, 20`  | `#FD7E14` | `linear-gradient(135deg, rgba(253, 126, 20, 0.1) 0%, rgba(247, 103, 7, 0.05) 100%)`   | `linear-gradient(135deg, #2A2118 0%, #211B15 100%)` |
| **Red**            | `250, 82, 82`   | `#FA5252` | `linear-gradient(135deg, rgba(250, 82, 82, 0.1) 0%, rgba(224, 49, 49, 0.05) 100%)`    | `linear-gradient(135deg, #2A1D1D 0%, #211818 100%)` |
| **Violet**         | `151, 117, 250` | `#9775FA` | `linear-gradient(135deg, rgba(151, 117, 250, 0.1) 0%, rgba(132, 94, 247, 0.05) 100%)` | `linear-gradient(135deg, #211E30 0%, #1C1A25 100%)` |

---

## 3. Типографика

- **Заголовки (Headings - `h1`–`h6`, `Title`)**:  
  `font-family: 'Unbounded', Vazirmatn, sans-serif; font-weight: 600..700;`  
  _Характер: футуристичный, широкий, технологичный гео-гротеск._
- **Основной текст (Body Text)**:  
  `font-family: 'Montserrat', Vazirmatn, sans-serif; font-weight: 400..600;`  
  _Характер: нейтральный, открытый, легко читаемый._
- **Моноширинный текст (Code / Keys / Links)**:  
  `font-family: 'Fira Mono', monospace; font-weight: 400..500;`

```html
<!-- Подключение шрифтов -->
<link rel="preconnect" href="https://fonts.googleapis.com" />
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
<link
	href="https://fonts.googleapis.com/css2?family=Fira+Mono:wght@400;500;700&family=Montserrat:wght@400;500;600;700&family=Unbounded:wght@500;600;700&family=Vazirmatn:wght@400;600;700&display=swap"
	rel="stylesheet"
/>
```

---

## 4. Скругления и геометрия (Border Radius & Sizing)

- **`radius-sm` (4px)**: Внутренние бейджи иконки, микро-тэги.
- **`radius-md` (8px)**: Инпуты, `InfoBlock`, селекторы приложений, модальные окна, списки ключей.
- **`radius-lg` (16px)**: Главные внешние карточки (`Card`), кнопки (`Button`), аккордеоны.
- **`radius-xl / full` (9999px)**: Круглые статус-иконки (`ThemeIcon`), круглые индикаторы (`Badge`).

---

## 5. Полный CSS Core (Готовые стили для внедрения)

```css
/* ============================================================
   1. ГЛОБАЛЬНЫЙ ФОН И СЕТКА
   ============================================================ */
html,
body {
	min-height: 100vh;
	margin: 0;
	padding: 0;
	background-color: #161b23;
	background-image:
		linear-gradient(to right, rgba(128, 128, 128, 0.05) 1px, transparent 1px),
		linear-gradient(to bottom, rgba(128, 128, 128, 0.05) 1px, transparent 1px);
	background-size: 60px 60px;
	background-attachment: fixed;
	color: #ffffff;
	font-family: 'Montserrat', sans-serif;
	-webkit-font-smoothing: antialiased;
}

/* Фоновые светящиеся сферы */
.animated-background {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	pointer-events: none;
	z-index: 0;
	background:
		radial-gradient(
			ellipse 60% 40% at 25% 40%,
			rgba(151, 117, 250, 0.08),
			transparent 50%
		),
		radial-gradient(
			ellipse 50% 50% at 85% 85%,
			rgba(34, 211, 238, 0.06),
			transparent 50%
		);
}

/* ============================================================
   2. МАКЕТ И СТРУКТУРА
   ============================================================ */
.app-layout {
	display: flex;
	flex-direction: column;
	min-height: 100vh;
	position: relative;
	z-index: 1;
}

.main-container {
	max-width: 800px;
	width: 100%;
	margin: 0 auto;
	padding: 24px 16px;
	box-sizing: border-box;
	animation: fadeIn 0.5s ease-in-out;
}

@keyframes fadeIn {
	from {
		opacity: 0;
		transform: translateY(6px);
	}
	to {
		opacity: 1;
		transform: translateY(0);
	}
}

/* ============================================================
   3. ЛИПКИЙ HEADER
   ============================================================ */
.header-wrapper {
	position: sticky;
	top: 0;
	z-index: 100;
	backdrop-filter: blur(12px);
	-webkit-backdrop-filter: blur(12px);
	background: rgba(22, 27, 35, 0.8);
	border-bottom: 1px solid rgba(255, 255, 255, 0.05);
	padding: 12px 16px;
}

@media (min-width: 768px) {
	.header-wrapper {
		border-radius: 0 0 16px 16px;
	}
}

/* ============================================================
   4. БАЗОВАЯ КАРТОЧКА (GLASS CARD)
   ============================================================ */
.glass-card {
	position: relative;
	background: rgba(255, 255, 255, 0.02);
	border: 1px solid rgba(255, 255, 255, 0.08);
	border-radius: 16px;
	overflow: hidden;
	backdrop-filter: blur(10px);
	-webkit-backdrop-filter: blur(10px);
	padding: 20px;
	transition:
		background 0.3s ease,
		border-color 0.3s ease,
		box-shadow 0.3s ease;
}

.glass-card:hover {
	border-color: rgba(34, 211, 238, 0.3);
	box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
}

/* ============================================================
   5. ИНТЕРАКТИВНЫЕ КНОПКИ ВЫБОРА (APP SELECTOR)
   ============================================================ */
.apps-grid {
	display: flex;
	flex-wrap: wrap;
	gap: 8px;
	margin-bottom: 16px;
}

.app-button {
	position: relative;
	display: flex;
	align-items: center;
	flex: 1 1 auto;
	padding: 0 48px 0 16px;
	height: 48px;
	min-width: 0;
	overflow: hidden;
	background: rgba(255, 255, 255, 0.02);
	border: 1px solid rgba(255, 255, 255, 0.08);
	border-left: 3px solid rgba(255, 255, 255, 0.08);
	border-radius: 8px;
	cursor: pointer;
	transition: all 0.2s ease;
	user-select: none;
}

.app-button:hover {
	background: rgba(255, 255, 255, 0.05);
	border-left-color: rgba(255, 255, 255, 0.25);
}

.app-button-active,
.app-button-active:hover {
	background: linear-gradient(
		90deg,
		rgba(34, 211, 238, 0.12) 0%,
		rgba(34, 211, 238, 0.04) 100%
	);
	border-color: rgba(34, 211, 238, 0.2);
	border-left-color: #22d3ee;
	box-shadow:
		inset 4px 0 12px -4px rgba(34, 211, 238, 0.3),
		0 0 12px rgba(34, 211, 238, 0.1);
}

.app-button .app-name {
	color: rgba(255, 255, 255, 0.85);
	font-size: 15px;
	font-weight: 600;
	white-space: nowrap;
	z-index: 1;
}

.app-button-active .app-name {
	color: #67e8f9;
}

/* Водяной знак SVG в кнопке */
.app-button .bg-icon {
	position: absolute;
	right: -12px;
	top: 50%;
	transform: translateY(-50%);
	width: 56px;
	height: 56px;
	opacity: 0.25;
	pointer-events: none;
	mask-image: linear-gradient(to right, transparent 0%, black 40%, black 100%);
	-webkit-mask-image: linear-gradient(
		to right,
		transparent 0%,
		black 40%,
		black 100%
	);
}

/* Золотой маркер Featured */
.featured-badge {
	position: absolute;
	top: 4px;
	left: 4px;
	width: 8px;
	height: 8px;
	border-radius: 50%;
	background: #efbf04;
	box-shadow: 0 0 6px rgba(239, 191, 4, 0.6);
}

/* ============================================================
   6. МИНИ-БЛОКИ МЕТРИК (INFO BLOCKS)
   ============================================================ */
.info-block {
	padding: 10px 12px;
	border-radius: 8px;
	display: flex;
	flex-direction: column;
	gap: 4px;
}

.info-block-cyan {
	background: linear-gradient(
		135deg,
		rgba(34, 211, 238, 0.1) 0%,
		rgba(6, 182, 212, 0.05) 100%
	);
	border: 1px solid rgba(34, 211, 238, 0.2);
}

.info-block-green {
	background: linear-gradient(
		135deg,
		rgba(64, 192, 87, 0.1) 0%,
		rgba(55, 178, 77, 0.05) 100%
	);
	border: 1px solid rgba(64, 192, 87, 0.2);
}

.info-block-orange {
	background: linear-gradient(
		135deg,
		rgba(253, 126, 20, 0.1) 0%,
		rgba(247, 103, 7, 0.05) 100%
	);
	border: 1px solid rgba(253, 126, 20, 0.2);
}

.info-block-red {
	background: linear-gradient(
		135deg,
		rgba(250, 82, 82, 0.1) 0%,
		rgba(224, 49, 49, 0.05) 100%
	);
	border: 1px solid rgba(250, 82, 82, 0.2);
}

/* ============================================================
   7. КНОПКИ ДЕЙСТВИЯ (BUTTONS)
   ============================================================ */
.btn-primary {
	background: rgba(34, 211, 238, 0.15);
	color: #22d3ee;
	border: 1px solid rgba(34, 211, 238, 0.3);
	border-radius: 8px;
	padding: 8px 16px;
	font-weight: 600;
	cursor: pointer;
	display: inline-flex;
	align-items: center;
	gap: 8px;
	transition: all 0.2s ease;
}

.btn-primary:hover {
	background: rgba(34, 211, 238, 0.25);
	border-color: rgba(34, 211, 238, 0.5);
	box-shadow: 0 0 12px rgba(34, 211, 238, 0.2);
}

.btn-icon {
	background: rgba(255, 255, 255, 0.02);
	border: 1px solid rgba(255, 255, 255, 0.1);
	border-radius: 8px;
	width: 40px;
	height: 40px;
	display: inline-flex;
	align-items: center;
	justify-content: center;
	color: #ffffff;
	cursor: pointer;
	transition: all 0.2s ease;
}

.btn-icon:hover {
	background: rgba(255, 255, 255, 0.06);
	border-color: rgba(34, 211, 238, 0.4);
}
```

---

## 6. System Prompt для ИИ (AI System Instruction)

Скопируйте данный блок в инструкцию любой языковой модели (Claude, GPT, Gemini, DeepSeek), чтобы она создавала страницы в точности по стилю Remnawave:

```text
You are an expert Frontend Engineer and UI/UX Designer.
You must design all components, pages, and widgets strictly adhering to the "Remnawave Cyber Glassmorphism" Design System:

1. AESTHETICS & THEME:
- Dark theme only. Canvas background is `#161B23`.
- Fixed background grid: 60px x 60px lines using `linear-gradient(to right/bottom, rgba(128,128,128, 0.05) 1px, transparent 1px)`.
- Ambient glow background orbs: Top-left Violet (`rgba(151,117,250, 0.08)`), bottom-right Cyan (`rgba(34,211,238, 0.06)`).
- Cards & Containers: Glassmorphism with `background: rgba(255, 255, 255, 0.02)`, `border: 1px solid rgba(255, 255, 255, 0.08)`, `border-radius: 16px`, `backdrop-filter: blur(10px)`. On hover: `border-color: rgba(34, 211, 238, 0.30)`, `box-shadow: 0 20px 40px rgba(0, 0, 0, 0.30)`.

2. TYPOGRAPHY:
- Headings (H1-H6, Titles): 'Unbounded', sans-serif; font-weight: 600-700.
- Body Text: 'Montserrat', sans-serif; font-weight: 400-600.
- Code / Monospace: 'Fira Mono', monospace; font-weight: 400-500.

3. ACCENTS & PALETTE:
- Primary Accent: Neon Cyan `#22D3EE` (rgb: 34, 211, 238).
- Semantic Accents: Green (`#40C057`), Orange (`#FD7E14`), Red (`#FA5252`), Yellow (`#FAB005`), Blue (`#228BE6`), Violet (`#9775FA`).
- Metric badges / subcards use soft 135deg gradients (`rgba(color, 0.1)` to `rgba(color, 0.05)`) with matching `1px solid rgba(color, 0.2)` borders.

4. INTERACTIVE PATTERNS:
- App switcher tabs: 48px height, 3px left border. Inactive: `border-left: 3px solid rgba(255,255,255,0.08)`. Active: `background: linear-gradient(90deg, rgba(34,211,238,0.12) 0%, rgba(34,211,238,0.04) 100%)`, `border-left-color: #22d3ee`, `box-shadow: inset 4px 0 12px -4px rgba(34,211,238,0.3)`. Right side has a cropped watermark SVG with linear-gradient alpha mask.
- Buttons & ActionIcons: Outline or Light variant, border-radius 8px or 16px, subtle glow on hover.

5. LAYOUT:
- Centered stream layout, max-width `800px`, sticky frosted glass header (`backdrop-filter: blur(12px)`), responsive gaps (16px / 24px).
```
