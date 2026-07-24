package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend/config"
	"backend/models"
	"backend/utils"
)

type publicMember struct {
	Username         string  `json:"username"`
	AvatarURL        *string `json:"avatar_url"`
	CurrentPosition  string  `json:"current_position"`
	CurrentWorkplace string  `json:"current_workplace"`
	PortfolioCount   int64   `json:"portfolio_count"`
	PostCount        int64   `json:"post_count"`
}

type publicUser struct {
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

type publicPostSummary struct {
	Id          uint       `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Excerpt     string     `json:"excerpt"`
	CoverURL    *string    `json:"cover_url"`
	PublishedAt *time.Time `json:"published_at"`
}

type publicContact struct {
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Github   string `json:"github"`
	Linkedin string `json:"linkedin"`
	Website  string `json:"website"`
}

type publicProfileResponse struct {
	User         publicUser                  `json:"user"`
	Professional professionalProfileResponse `json:"professional"`
	Contact      publicContact               `json:"contact"`
	Portfolios   []models.Portfolio          `json:"portfolios"`
	Posts        []publicPostSummary         `json:"posts"`
	TotalPosts   int64                       `json:"total_posts"`
	TotalPages   int                         `json:"total_pages"`
	CurrentPage  int                         `json:"current_page"`
}

func usernameOf(u models.User) string {
	if u.Username != nil {
		return *u.Username
	}
	return ""
}

func ListPublicMembers(c *gin.Context) {
	var users []models.User
	config.DB.
		Where("is_active = ? AND username IS NOT NULL", true).
		Where("id IN (SELECT DISTINCT user_id FROM portfolios)").
		Where("id IN (SELECT DISTINCT user_id FROM blog_posts WHERE status = ?)", "published").
		Order("id asc").
		Find(&users)

	members := make([]publicMember, 0, len(users))
	for _, u := range users {
		var profile models.ProfessionalProfile
		config.DB.Where("user_id = ?", u.Id).First(&profile)

		var pfCount, postCount int64
		config.DB.Model(&models.Portfolio{}).Where("user_id = ?", u.Id).Count(&pfCount)
		config.DB.Model(&models.BlogPost{}).Where("user_id = ? AND status = ?", u.Id, "published").Count(&postCount)

		members = append(members, publicMember{
			Username:         usernameOf(u),
			AvatarURL:        u.AvatarURL,
			CurrentPosition:  profile.CurrentPosition,
			CurrentWorkplace: profile.CurrentWorkplace,
			PortfolioCount:   pfCount,
			PostCount:        postCount,
		})
	}

	utils.SuccessResponse(c, 200, gin.H{"members": members})
}

func GetPublicProfile(c *gin.Context) {
	username := strings.ToLower(strings.TrimSpace(c.Param("username")))

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	const postLimit = 4

	var user models.User
	if err := config.DB.Where("username = ? AND is_active = ?", username, true).First(&user).Error; err != nil {
		utils.ErrorResponse(c, 404, "Member tidak ditemukan")
		return
	}

	var profile models.ProfessionalProfile
	config.DB.Where("user_id = ?", user.Id).First(&profile)

	var skills []models.Skill
	config.DB.Where("user_id = ?", user.Id).Order("id asc").Find(&skills)
	var experiences []models.WorkExperience
	config.DB.Where("user_id = ?", user.Id).Order("id desc").Find(&experiences)
	var subjects []models.Subject
	config.DB.Where("user_id = ?", user.Id).Order("id asc").Find(&subjects)
	if skills == nil {
		skills = []models.Skill{}
	}
	if experiences == nil {
		experiences = []models.WorkExperience{}
	}
	if subjects == nil {
		subjects = []models.Subject{}
	}

	var contact models.Contact
	config.DB.Where("user_id = ?", user.Id).First(&contact)

	var portfolios []models.Portfolio
	config.DB.Where("user_id = ?", user.Id).Order("id desc").Find(&portfolios)
	if portfolios == nil {
		portfolios = []models.Portfolio{}
	}

	var totalPosts int64
	config.DB.Model(&models.BlogPost{}).Where("user_id = ? AND status = ?", user.Id, "published").Count(&totalPosts)

	var posts []models.BlogPost
	config.DB.Where("user_id = ? AND status = ?", user.Id, "published").
		Order("published_at desc").
		Offset((page - 1) * postLimit).Limit(postLimit).Find(&posts)

	totalPages := int((totalPosts + postLimit - 1) / postLimit)
	if totalPages < 1 {
		totalPages = 1
	}

	summaries := make([]publicPostSummary, 0, len(posts))
	for _, p := range posts {
		summaries = append(summaries, publicPostSummary{
			Id: p.Id, Slug: p.Slug, Title: p.Title, Excerpt: p.Excerpt,
			CoverURL: p.CoverURL, PublishedAt: p.PublishedAt,
		})
	}

	utils.SuccessResponse(c, 200, publicProfileResponse{
		User: publicUser{Username: usernameOf(user), AvatarURL: user.AvatarURL, CreatedAt: user.CreatedAt},
		Professional: professionalProfileResponse{
			CurrentWorkplace: profile.CurrentWorkplace,
			CurrentPosition:  profile.CurrentPosition,
			Skills:           skills,
			Experiences:      experiences,
			Subjects:         subjects,
		},
		Contact: publicContact{
			Phone:    contact.Phone,
			Email:    contact.Email,
			Address:  contact.Address,
			Github:   contact.Github,
			Linkedin: contact.Linkedin,
			Website:  contact.Website,
		},
		Portfolios:  portfolios,
		Posts:       summaries,
		TotalPosts:  totalPosts,
		TotalPages:  totalPages,
		CurrentPage: page,
	})
}

func GetPublicPost(c *gin.Context) {
	username := strings.ToLower(strings.TrimSpace(c.Param("username")))
	slug := c.Param("slug")

	var user models.User
	if err := config.DB.Where("username = ? AND is_active = ?", username, true).First(&user).Error; err != nil {
		utils.ErrorResponse(c, 404, "Member tidak ditemukan")
		return
	}

	var post models.BlogPost
	if err := config.DB.
		Where("user_id = ? AND slug = ? AND status = ?", user.Id, slug, "published").
		Preload("Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position asc") }).
		First(&post).Error; err != nil {
		utils.ErrorResponse(c, 404, "Blog tidak ditemukan")
		return
	}

	utils.SuccessResponse(c, 200, gin.H{
		"author": publicUser{Username: usernameOf(user), AvatarURL: user.AvatarURL, CreatedAt: user.CreatedAt},
		"post":   post,
	})
}

type publicArticleAuthor struct {
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
	Position  string  `json:"position"`
}

type publicArticle struct {
	Id          uint                `json:"id"`
	Slug        string              `json:"slug"`
	Title       string              `json:"title"`
	Excerpt     string              `json:"excerpt"`
	CoverURL    *string             `json:"cover_url"`
	PublishedAt *time.Time          `json:"published_at"`
	ReadMinutes int                 `json:"read_minutes"`
	Author      publicArticleAuthor `json:"author"`
}

func readMinutes(p models.BlogPost) int {
	words := 0
	for _, b := range p.Blocks {
		if b.Type == "text" {
			words += len(strings.Fields(b.Text))
		}
	}
	m := (words + 199) / 200
	if m < 1 {
		m = 1
	}
	return m
}

func mapArticles(posts []models.BlogPost) []publicArticle {
	out := make([]publicArticle, 0, len(posts))
	userCache := map[uint]models.User{}
	posCache := map[uint]string{}
	for _, p := range posts {
		u, ok := userCache[p.UserId]
		if !ok {
			config.DB.First(&u, p.UserId)
			userCache[p.UserId] = u
		}
		pos, ok := posCache[p.UserId]
		if !ok {
			var prof models.ProfessionalProfile
			config.DB.Where("user_id = ?", p.UserId).First(&prof)
			pos = prof.CurrentPosition
			posCache[p.UserId] = pos
		}
		out = append(out, publicArticle{
			Id:          p.Id,
			Slug:        p.Slug,
			Title:       p.Title,
			Excerpt:     p.Excerpt,
			CoverURL:    p.CoverURL,
			PublishedAt: p.PublishedAt,
			ReadMinutes: readMinutes(p),
			Author: publicArticleAuthor{
				Username:  usernameOf(u),
				AvatarURL: u.AvatarURL,
				Position:  pos,
			},
		})
	}
	return out
}

func ListPublicPosts(c *gin.Context) {
	const limit = 9

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}

	featured := []publicArticle{}
	var featuredIDs []uint
	if page == 1 {
		weekAgo := time.Now().AddDate(0, 0, -7)
		var feat []models.BlogPost
		config.DB.Where("status = ? AND published_at >= ?", "published", weekAgo).
			Preload("Blocks").Order("published_at desc").Limit(3).Find(&feat)
		if len(feat) == 0 {
			config.DB.Where("status = ?", "published").
				Preload("Blocks").Order("published_at desc").Limit(3).Find(&feat)
		}
		featured = mapArticles(feat)
		for _, f := range feat {
			featuredIDs = append(featuredIDs, f.Id)
		}
	}

	baseQuery := config.DB.Where("status = ?", "published")
	if len(featuredIDs) > 0 {
		baseQuery = baseQuery.Where("id NOT IN ?", featuredIDs)
	}

	var total int64
	baseQuery.Model(&models.BlogPost{}).Count(&total)

	var posts []models.BlogPost
	baseQuery.
		Preload("Blocks").
		Order("published_at desc").
		Offset((page - 1) * limit).Limit(limit).Find(&posts)

	totalPages := int((total + limit - 1) / limit)
	if totalPages < 1 {
		totalPages = 1
	}

	utils.SuccessResponse(c, 200, gin.H{
		"articles":    mapArticles(posts),
		"featured":    featured,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
	})
}
