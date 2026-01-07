package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

// CompanyInfo 简化的公司信息模型
type CompanyInfo struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
	LogoURL   string `json:"logo_url"`
	Industry  string `json:"industry"`
	Location  string `json:"location"`
}

// Job 数据模型映射 zervigo_jobs 表
type Job struct {
	ID                 uint           `json:"id" gorm:"primaryKey"`
	Title              string         `json:"title" gorm:"size:200;not null"`
	Description        string         `json:"description" gorm:"type:text;not null"`
	Requirements       string         `json:"requirements" gorm:"type:text"`
	Responsibilities   string         `json:"responsibilities" gorm:"type:text"`
	CompanyID          int64          `json:"companyId" gorm:"column:company_id;not null"`
	CompanyName        string         `json:"companyName" gorm:"column:company_name"`
	CompanyLogo        *string        `json:"companyLogo" gorm:"column:company_logo"`
	JobCategory        string         `json:"jobCategory" gorm:"column:job_category"`
	JobSubcategory     string         `json:"jobSubcategory" gorm:"column:job_subcategory"`
	JobLevel           string         `json:"jobLevel" gorm:"column:job_level"`
	WorkType           string         `json:"workType" gorm:"column:work_type"`
	WorkLocation       string         `json:"workLocation" gorm:"column:work_location"`
	WorkAddress        string         `json:"workAddress" gorm:"column:work_address"`
	RemoteAllowed      bool           `json:"remoteAllowed" gorm:"column:remote_allowed"`
	SalaryMin          *int           `json:"salaryMin" gorm:"column:salary_min"`
	SalaryMax          *int           `json:"salaryMax" gorm:"column:salary_max"`
	SalaryCurrency     string         `json:"salaryCurrency" gorm:"column:salary_currency"`
	SalaryPeriod       string         `json:"salaryPeriod" gorm:"column:salary_period"`
	SalaryNegotiable   bool           `json:"salaryNegotiable" gorm:"column:salary_negotiable"`
	ExperienceRequired string         `json:"experienceRequired" gorm:"column:experience_required"`
	EducationRequired  string         `json:"educationRequired" gorm:"column:education_required"`
	SkillsRequired     pq.StringArray `json:"skillsRequired" gorm:"type:text[];column:skills_required"`
	LanguagesRequired  pq.StringArray `json:"languagesRequired" gorm:"type:text[];column:languages_required"`
	Benefits           pq.StringArray `json:"benefits" gorm:"type:text[];column:benefits"`
	Perks              pq.StringArray `json:"perks" gorm:"type:text[];column:perks"`
	Status             string         `json:"status" gorm:"column:status"`
	IsFeatured         bool           `json:"isFeatured" gorm:"column:is_featured"`
	IsUrgent           bool           `json:"isUrgent" gorm:"column:is_urgent"`
	ViewCount          int64          `json:"viewCount" gorm:"column:view_count"`
	ApplyCount         int64          `json:"applyCount" gorm:"column:apply_count"`
	FavoriteCount      int64          `json:"favoriteCount" gorm:"column:favorite_count"`
	PublishAt          *time.Time     `json:"publishAt" gorm:"column:publish_at"`
	ExpireAt           *time.Time     `json:"expireAt" gorm:"column:expire_at"`
	CreatedAt          time.Time      `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt          time.Time      `json:"updatedAt" gorm:"column:updated_at"`
	CreatedBy          int64          `json:"createdBy" gorm:"column:created_by"`
}

func (Job) TableName() string {
	return "zervigo_jobs"
}

// JobApplication 映射职位申请表
type JobApplication struct {
	ID                uint       `json:"id" gorm:"primaryKey"`
	JobID             uint       `json:"jobId" gorm:"column:job_id;not null"`
	UserID            uint       `json:"userId" gorm:"column:user_id;not null"`
	ResumeID          *uint      `json:"resumeId" gorm:"column:resume_id"`
	CoverLetter       string     `json:"coverLetter" gorm:"column:cover_letter"`
	ApplicationSource string     `json:"applicationSource" gorm:"column:application_source"`
	Status            string     `json:"status" gorm:"column:status"`
	ApplicationStage  string     `json:"applicationStage" gorm:"column:application_stage"`
	ReviewedBy        *uint      `json:"reviewedBy" gorm:"column:reviewed_by"`
	ReviewedAt        *time.Time `json:"reviewedAt" gorm:"column:reviewed_at"`
	ReviewNotes       string     `json:"reviewNotes" gorm:"column:review_notes"`
	AppliedAt         time.Time  `json:"appliedAt" gorm:"column:applied_at"`
	UpdatedAt         time.Time  `json:"updatedAt" gorm:"column:updated_at"`
}

func (JobApplication) TableName() string {
	return "zervigo_job_applications"
}

// JobFavorite 映射职位收藏表
type JobFavorite struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId" gorm:"column:user_id;not null"`
	JobID     uint      `json:"jobId" gorm:"column:job_id;not null"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
}

func (JobFavorite) TableName() string {
	return "zervigo_job_favorites"
}

// JobFilter 查询过滤条件
type JobFilter struct {
	Page       int
	PageSize   int
	Keyword    string
	Categories []string
	WorkType   string
	Experience string
	Status     string
}

// JobSummary 列表展示数据
type JobSummary struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Company     string   `json:"company"`
	CompanyLogo string   `json:"companyLogo"`
	Location    string   `json:"location"`
	Description string   `json:"description"`
	Salary      string   `json:"salary"`
	Experience  string   `json:"experience"`
	Education   string   `json:"education"`
	Tags        []string `json:"tags"`
	PostTime    string   `json:"postTime"`
	IsActive    bool     `json:"isActive"`
}

// JobDetail 明细数据
type JobDetail struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Requirements     string     `json:"requirements"`
	Responsibilities string     `json:"responsibilities"`
	CompanyID        int64      `json:"companyId"`
	CompanyName      string     `json:"companyName"`
	CompanyLogo      string     `json:"companyLogo"`
	Location         string     `json:"location"`
	WorkType         string     `json:"workType"`
	SalaryMin        *int       `json:"salaryMin"`
	SalaryMax        *int       `json:"salaryMax"`
	SalaryCurrency   string     `json:"salaryCurrency"`
	SalaryPeriod     string     `json:"salaryPeriod"`
	Experience       string     `json:"experience"`
	Education        string     `json:"education"`
	Skills           []string   `json:"skills"`
	Benefits         []string   `json:"benefits"`
	Perks            []string   `json:"perks"`
	Status           string     `json:"status"`
	PublishAt        *time.Time `json:"publishAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// JobListResult 列表结果
type JobListResult struct {
	Items      []JobSummary `json:"items"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"pageSize"`
	TotalPages int          `json:"totalPages"`
}

// CreateJobRequest 创建职位请求
type CreateJobRequest struct {
	Title            string   `json:"title" binding:"required"`
	Description      string   `json:"description" binding:"required"`
	Requirements     string   `json:"requirements"`
	Responsibilities string   `json:"responsibilities"`
	CompanyID        int64    `json:"companyId" binding:"required"`
	CompanyName      string   `json:"companyName" binding:"required"`
	CompanyLogo      string   `json:"companyLogo"`
	Location         string   `json:"location" binding:"required"`
	WorkType         string   `json:"workType"`
	SalaryMin        *int     `json:"salaryMin"`
	SalaryMax        *int     `json:"salaryMax"`
	SalaryCurrency   string   `json:"salaryCurrency"`
	SalaryPeriod     string   `json:"salaryPeriod"`
	Experience       string   `json:"experience"`
	Education        string   `json:"education"`
	Tags             []string `json:"tags"`
	Skills           []string `json:"skills"`
	Benefits         []string `json:"benefits"`
	Perks            []string `json:"perks"`
	Status           string   `json:"status"`
	CreatedBy        int64    `json:"createdBy"`
}

// UpdateJobRequest 更新职位请求
type UpdateJobRequest struct {
	Title            *string  `json:"title"`
	Description      *string  `json:"description"`
	Requirements     *string  `json:"requirements"`
	Responsibilities *string  `json:"responsibilities"`
	CompanyName      *string  `json:"companyName"`
	CompanyLogo      *string  `json:"companyLogo"`
	Location         *string  `json:"location"`
	WorkType         *string  `json:"workType"`
	SalaryMin        *int     `json:"salaryMin"`
	SalaryMax        *int     `json:"salaryMax"`
	SalaryCurrency   *string  `json:"salaryCurrency"`
	SalaryPeriod     *string  `json:"salaryPeriod"`
	Experience       *string  `json:"experience"`
	Education        *string  `json:"education"`
	Tags             []string `json:"tags"`
	Skills           []string `json:"skills"`
	Benefits         []string `json:"benefits"`
	Perks            []string `json:"perks"`
	Status           *string  `json:"status"`
}

// FavoriteRequest 收藏请求
type FavoriteRequest struct {
	JobID  uint `json:"jobId" binding:"required"`
	UserID uint `json:"userId"`
}

// JobStats 用户侧职位统计
type JobStats struct {
	Applied   int `json:"applied"`
	Interview int `json:"interview"`
	Offers    int `json:"offers"`
	Favorites int `json:"favorites"`
}

// 辅助方法
func (j Job) toSummary() JobSummary {
	salaryDisplay := "面议"
	if j.SalaryMin != nil || j.SalaryMax != nil {
		min := 0
		max := 0
		if j.SalaryMin != nil {
			min = *j.SalaryMin
		}
		if j.SalaryMax != nil {
			max = *j.SalaryMax
		}
		currency := j.SalaryCurrency
		if currency == "" {
			currency = "CNY"
		}
		salaryDisplay = formatSalary(min, max, currency, j.SalaryPeriod)
	}

	postTime := j.CreatedAt.Format("2006-01-02")
	if j.PublishAt != nil {
		postTime = j.PublishAt.Format("2006-01-02")
	}

	tags := copyStringSlice(j.SkillsRequired)
	if len(tags) == 0 {
		if j.JobCategory != "" {
			tags = append(tags, j.JobCategory)
		}
		if j.JobSubcategory != "" {
			tags = append(tags, j.JobSubcategory)
		}
		if j.JobLevel != "" {
			tags = append(tags, j.JobLevel)
		}
	}

	companyLogo := ""
	if j.CompanyLogo != nil {
		companyLogo = *j.CompanyLogo
	}

	isActive := strings.EqualFold(j.Status, JobStatusPublished) || strings.EqualFold(j.Status, JobStatusOpen)

	return JobSummary{
		ID:          formatUintID(j.ID),
		Title:       j.Title,
		Company:     j.CompanyName,
		CompanyLogo: companyLogo,
		Location:    j.WorkLocation,
		Description: j.Description,
		Salary:      salaryDisplay,
		Experience:  j.ExperienceRequired,
		Education:   j.EducationRequired,
		Tags:        tags,
		PostTime:    postTime,
		IsActive:    isActive,
	}
}

func (j Job) toDetail() JobDetail {
	skills := copyStringSlice(j.SkillsRequired)
	if len(skills) == 0 && j.JobCategory != "" {
		skills = append(skills, j.JobCategory)
	}

	companyLogo := ""
	if j.CompanyLogo != nil {
		companyLogo = *j.CompanyLogo
	}

	return JobDetail{
		ID:               formatUintID(j.ID),
		Title:            j.Title,
		Description:      j.Description,
		Requirements:     j.Requirements,
		Responsibilities: j.Responsibilities,
		CompanyID:        j.CompanyID,
		CompanyName:      j.CompanyName,
		CompanyLogo:      companyLogo,
		Location:         j.WorkLocation,
		WorkType:         j.WorkType,
		SalaryMin:        j.SalaryMin,
		SalaryMax:        j.SalaryMax,
		SalaryCurrency:   j.SalaryCurrency,
		SalaryPeriod:     j.SalaryPeriod,
		Experience:       j.ExperienceRequired,
		Education:        j.EducationRequired,
		Skills:           skills,
		Benefits:         copyStringSlice(j.Benefits),
		Perks:            copyStringSlice(j.Perks),
		Status:           j.Status,
		PublishAt:        j.PublishAt,
		UpdatedAt:        j.UpdatedAt,
	}
}

func formatUintID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

func copyStringSlice(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

func formatSalary(min, max int, currency, period string) string {
	if min == 0 && max == 0 {
		return "面议"
	}
	if max == 0 {
		return fmt.Sprintf("%s %d/%s", currency, min, periodValue(period))
	}
	if min == 0 {
		return fmt.Sprintf("%s %d/%s", currency, max, periodValue(period))
	}
	return fmt.Sprintf("%s %d-%d/%s", currency, min, max, periodValue(period))
}

func periodValue(period string) string {
	switch strings.ToLower(period) {
	case "yearly":
		return "年"
	case "hourly":
		return "小时"
	default:
		return "月"
	}
}

const (
	JobStatusDraft     = "draft"
	JobStatusPublished = "published"
	JobStatusPaused    = "paused"
	JobStatusClosed    = "closed"
	JobStatusOpen      = "open"
)
