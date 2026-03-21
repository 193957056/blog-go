package models

import (
	"time"

	"lumina-blog/config"
	"gorm.io/gorm"
)

type Post struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Slug        string         `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	Content     string         `gorm:"type:text" json:"content"`
	Excerpt     string         `gorm:"size:500" json:"excerpt"`
	Cover       string         `gorm:"size:500" json:"cover"`
	Status      string         `gorm:"size:20;default:'draft'" json:"status"` // draft, published
	ViewCount   int            `gorm:"default:0" json:"view_count"`
	ReadTime    int            `json:"read_time"` // in minutes
	Lang        string         `gorm:"size:10;default:'zh-CN'" json:"lang"` // Language: zh-CN, en, ja, etc.
	CategoryID  uint           `json:"category_id"`
	Category    Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tags        []Tag          `gorm:"many2many:post_tags;" json:"tags,omitempty"`
	
	// SEO fields (computed, not stored in DB)
	CanonicalURL   string `json:"canonical_url,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
}

type Category struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Slug      string         `gorm:"size:50;uniqueIndex;not null" json:"slug"`
	Color     string         `gorm:"size:20" json:"color"`
	Posts     []Post         `json:"posts,omitempty"`
}

type Tag struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Slug      string         `gorm:"size:50;uniqueIndex;not null" json:"slug"`
	Color     string         `gorm:"size:20" json:"color"`
	Posts     []Post         `gorm:"many2many:post_tags;" json:"posts,omitempty"`
}

// SeedData creates default categories and tags
func SeedData() {
	// Create categories first
	categories := []Category{
		{Name: "技术", Slug: "tech", Color: "#7C3AED"},
		{Name: "设计", Slug: "design", Color: "#06B6D4"},
		{Name: "生活", Slug: "life", Color: "#10B981"},
		{Name: "随笔", Slug: "essay", Color: "#F59E0B"},
	}

	for _, cat := range categories {
		var exists Category
		if err := config.DB.Where("slug = ?", cat.Slug).First(&exists).Error; err != nil {
			config.DB.Create(&cat)
		}
	}

	// Create tags
	tags := []Tag{
		{Name: "Vue", Slug: "vue", Color: "#42B883"},
		{Name: "Go", Slug: "go", Color: "#00ADD8"},
		{Name: "前端", Slug: "frontend", Color: "#E34F26"},
		{Name: "后端", Slug: "backend", Color: "#1572B6"},
		{Name: "UI/UX", Slug: "ui-ux", Color: "#FF61F6"},
	}

	for _, tag := range tags {
		var exists Tag
		if err := config.DB.Where("slug = ?", tag.Slug).First(&exists).Error; err != nil {
			config.DB.Create(&tag)
		}
	}

	// Create sample posts
	var catTech, catDesign Category
	config.DB.Where("slug = ?", "tech").First(&catTech)
	config.DB.Where("slug = ?", "design").First(&catDesign)

	var tagVue, tagGo, tagFrontend Tag
	config.DB.Where("slug = ?", "vue").First(&tagVue)
	config.DB.Where("slug = ?", "go").First(&tagGo)
	config.DB.Where("slug = ?", "frontend").First(&tagFrontend)

	samplePosts := []Post{
		{
			Title:       "Vue 3 组合式 API 完全指南",
			Slug:        "vue3-composition-api-guide",
			Content:     "# Vue 3 组合式 API 完全指南\n\nVue 3 引入的组合式 API 是近年来最大的架构变革。本文将深入探讨其设计理念与实践。\n\n## 什么是组合式 API？\n\n组合式 API 是一组附加的 API，允许我们使用函数而不是声明选项的方式来组织组件逻辑。\n\n```javascript\nimport { ref, computed, onMounted } from 'vue'\n\nexport default {\n  setup() {\n    const count = ref(0)\n    const doubled = computed(() => count.value * 2)\n    \n    onMounted(() => {\n      console.log('Component mounted!')\n    })\n    \n    return { count, doubled }\n  }\n}\n```\n\n## 为什么选择组合式 API？\n\n1. **更好的逻辑复用** - 组合函数可以轻松共享状态逻辑\n2. **更灵活的组织** - 按逻辑而非选项组织代码\n3. **更好的 TypeScript 支持** - 类型推断更加自然",
			Excerpt:     "Vue 3 引入的组合式 API 是近年来最大的架构变革。本文将深入探讨其设计理念与实践。",
			Cover:       "https://images.unsplash.com/photo-1555066931-4365d14bab8c?w=800",
			Status:      "published",
			ViewCount:   128,
			Lang:        "zh-CN",
			CategoryID:  catTech.ID,
			Tags:        []Tag{tagVue, tagFrontend},
		},
		{
			Title:       "Go 语言并发编程实战",
			Slug:        "go-concurrent-programming",
			Content:     "# Go 语言并发编程实战\n\nGo 语言的 goroutine 和 channel 使得并发编程变得优雅而简单。\n\n## Goroutine 基础\n\ngoroutine 是由 Go 运行时管理的轻量级线程。\n\n```go\nfunc main() {\n    go sayHello()\n    time.Sleep(1 * time.Second)\n}\n\nfunc sayHello() {\n    fmt.Println(\"Hello, World!\")\n}\n```\n\n## Channel 用法\n\nchannel 是 goroutine 之间的通信机制。\n\n```go\nch := make(chan int)\ngo func() {\n    ch <- 42\n}()\nvalue := <-ch\n```",
			Excerpt:     "Go 语言的 goroutine 和 channel 使得并发编程变得优雅而简单。",
			Cover:       "https://images.unsplash.com/photo-1555066931-4365d14bab8c?w=800",
			Status:      "published",
			ViewCount:   256,
			Lang:        "zh-CN",
			CategoryID:  catTech.ID,
			Tags:        []Tag{tagGo},
		},
		{
			Title:       "现代 UI 设计趋势 2024",
			Slug:        "ui-design-trends-2024",
			Content:     "# 现代 UI 设计趋势 2024\n\n设计界正在经历一场静默的革命。让我们来看看今年的主要趋势。\n\n## 1. 玻璃拟态 (Glassmorphism)\n\n模糊背景效果继续流行，带来层次感和深度。\n\n## 2. 超级渐变\n\n比以往更大胆的多色彩渐变成为主角。\n\n## 3. 微动画\n\n细节动画提升用户体验和交互感。",
			Excerpt:     "设计界正在经历一场静默的革命。让我们来看看今年的主要趋势。",
			Cover:       "https://images.unsplash.com/photo-1561070791-2526d30994b5?w=800",
			Status:      "published",
			ViewCount:   89,
			Lang:        "zh-CN",
			CategoryID:  catDesign.ID,
			Tags:        []Tag{tagFrontend},
		},
	}

	for _, post := range samplePosts {
		var exists Post
		if err := config.DB.Where("slug = ?", post.Slug).First(&exists).Error; err != nil {
			post.ReadTime = len(post.Content) / 200
			if post.ReadTime < 1 {
				post.ReadTime = 1
			}
			config.DB.Create(&post)
		}
	}
}
