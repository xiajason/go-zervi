package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/szjason72/zervigo/shared/core/context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// JobService 封装职位相关数据库操作
type JobService struct {
	db         *gorm.DB
	dialect    string
	isPostgres bool
}

func NewJobService(db *gorm.DB) *JobService {
	dialect := strings.ToLower(db.Dialector.Name())
	return &JobService{
		db:         db,
		dialect:    dialect,
		isPostgres: strings.Contains(dialect, "postgres"),
	}
}

// ListJobs 返回分页后的职位列表（自动过滤租户）
func (s *JobService) ListJobs(ctx context.Context, filter JobFilter) (JobListResult, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	// 从context获取租户ID
	tenantID, err := context.GetTenantID(ctx)
	if err != nil {
		// 如果没有租户ID，使用默认租户
		tenantID = 1
	}

	query := s.db.WithContext(ctx).Model(&Job{}).Where("tenant_id = ?", tenantID)

	if filter.Keyword != "" {
		keyword := "%" + strings.ToLower(filter.Keyword) + "%"
		if s.isPostgres {
			query = query.Where("LOWER(title) LIKE ? OR LOWER(company_name) LIKE ?", keyword, keyword)
		} else {
			query = query.Where("title LIKE ? OR company_name LIKE ?", keyword, keyword)
		}
	}

	if filter.WorkType != "" {
		query = query.Where("work_type = ?", filter.WorkType)
	}
	if filter.Experience != "" {
		query = query.Where("experience_required = ?", filter.Experience)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	} else {
		query = query.Where("status IN ?", []string{JobStatusDraft, JobStatusPublished, JobStatusOpen, JobStatusPaused})
	}

	if len(filter.Categories) > 0 {
		normalized := normalizeTags(filter.Categories)
		if s.isPostgres {
			query = query.Where("skills_required @> ?", pq.StringArray(normalized))
		} else {
			for _, tag := range normalized {
				like := "%" + tag + "%"
				query = query.Where("skills_required LIKE ? OR job_category = ? OR job_subcategory = ?", like, tag, tag)
			}
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return JobListResult{}, err
	}

	if total == 0 {
		return JobListResult{Items: []JobSummary{}, Total: 0, Page: filter.Page, PageSize: filter.PageSize, TotalPages: 0}, nil
	}

	totalPages := int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize))
	if totalPages == 0 {
		totalPages = 1
	}
	if filter.Page > totalPages {
		filter.Page = totalPages
	}

	orderExpr := "created_at DESC"
	if s.isPostgres {
		orderExpr = "COALESCE(publish_at, created_at) DESC"
	}

	var jobs []Job
	if err := query.Order(orderExpr).Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize).Find(&jobs).Error; err != nil {
		return JobListResult{}, err
	}

	summaries := make([]JobSummary, len(jobs))
	for i, job := range jobs {
		summaries[i] = job.toSummary()
	}

	return JobListResult{
		Items:      summaries,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetJob 获取职位详情（自动过滤租户）
func (s *JobService) GetJob(ctx context.Context, id uint) (JobDetail, error) {
	// 从context获取租户ID
	tenantID, err := context.GetTenantID(ctx)
	if err != nil {
		tenantID = 1 // 默认租户
	}

	var job Job
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&job).Error; err != nil {
		return JobDetail{}, err
	}
	return job.toDetail(), nil
}

// CreateJob 创建新职位（自动设置租户ID）
func (s *JobService) CreateJob(ctx context.Context, req CreateJobRequest) (JobDetail, error) {
	// 从context获取租户ID
	tenantID, err := context.GetTenantID(ctx)
	if err != nil {
		tenantID = 1 // 默认租户
	}

	job := Job{
		TenantID:           tenantID, // 自动设置租户ID
		Title:              req.Title,
		Description:        req.Description,
		Requirements:       req.Requirements,
		Responsibilities:   req.Responsibilities,
		CompanyID:          req.CompanyID,
		CompanyName:        req.CompanyName,
		CompanyLogo:        stringPtrOrNil(req.CompanyLogo),
		WorkType:           req.WorkType,
		WorkLocation:       req.Location,
		SalaryMin:          req.SalaryMin,
		SalaryMax:          req.SalaryMax,
		SalaryCurrency:     req.SalaryCurrency,
		SalaryPeriod:       req.SalaryPeriod,
		ExperienceRequired: req.Experience,
		EducationRequired:  req.Education,
		Status:             defaultStatus(req.Status),
		CreatedBy:          req.CreatedBy,
	}

	if len(req.Tags) > 0 {
		job.SkillsRequired = pq.StringArray(normalizeTags(req.Tags))
	}
	if len(req.Skills) > 0 {
		job.SkillsRequired = pq.StringArray(normalizeTags(req.Skills))
	}
	if len(req.Benefits) > 0 {
		job.Benefits = pq.StringArray(normalizeTags(req.Benefits))
	}
	if len(req.Perks) > 0 {
		job.Perks = pq.StringArray(normalizeTags(req.Perks))
	}

	if job.Status == JobStatusPublished || job.Status == JobStatusOpen {
		now := time.Now()
		job.PublishAt = &now
	}

	if err := s.db.WithContext(ctx).Create(&job).Error; err != nil {
		return JobDetail{}, err
	}

	return job.toDetail(), nil
}

// UpdateJob 更新职位（自动校验租户ID）
func (s *JobService) UpdateJob(ctx context.Context, id uint, req UpdateJobRequest) (JobDetail, error) {
	// 从context获取租户ID
	tenantID, err := context.GetTenantID(ctx)
	if err != nil {
		tenantID = 1 // 默认租户
	}

	var job Job
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&job).Error; err != nil {
		return JobDetail{}, err
	}

	applyString := func(value **string, src *string) {
		if src != nil {
			if strings.TrimSpace(*src) == "" {
				*value = nil
			} else {
				copy := strings.TrimSpace(*src)
				*value = &copy
			}
		}
	}

	if req.Title != nil {
		job.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		job.Description = strings.TrimSpace(*req.Description)
	}
	if req.Requirements != nil {
		job.Requirements = strings.TrimSpace(*req.Requirements)
	}
	if req.Responsibilities != nil {
		job.Responsibilities = strings.TrimSpace(*req.Responsibilities)
	}
	if req.CompanyName != nil {
		job.CompanyName = strings.TrimSpace(*req.CompanyName)
	}
	applyString(&job.CompanyLogo, req.CompanyLogo)
	if req.Location != nil {
		job.WorkLocation = strings.TrimSpace(*req.Location)
	}
	if req.WorkType != nil {
		job.WorkType = strings.TrimSpace(*req.WorkType)
	}
	if req.SalaryMin != nil {
		job.SalaryMin = req.SalaryMin
	}
	if req.SalaryMax != nil {
		job.SalaryMax = req.SalaryMax
	}
	if req.SalaryCurrency != nil {
		job.SalaryCurrency = strings.TrimSpace(*req.SalaryCurrency)
	}
	if req.SalaryPeriod != nil {
		job.SalaryPeriod = strings.TrimSpace(*req.SalaryPeriod)
	}
	if req.Experience != nil {
		job.ExperienceRequired = strings.TrimSpace(*req.Experience)
	}
	if req.Education != nil {
		job.EducationRequired = strings.TrimSpace(*req.Education)
	}
	if req.Tags != nil && len(req.Tags) > 0 {
		job.SkillsRequired = pq.StringArray(normalizeTags(req.Tags))
	}
	if req.Skills != nil && len(req.Skills) > 0 {
		job.SkillsRequired = pq.StringArray(normalizeTags(req.Skills))
	}
	if req.Benefits != nil {
		job.Benefits = pq.StringArray(normalizeTags(req.Benefits))
	}
	if req.Perks != nil {
		job.Perks = pq.StringArray(normalizeTags(req.Perks))
	}
	if req.Status != nil {
		job.Status = strings.TrimSpace(*req.Status)
	}

	if err := s.db.WithContext(ctx).Save(&job).Error; err != nil {
		return JobDetail{}, err
	}

	return job.toDetail(), nil
}

// ToggleFavorite 切换收藏状态
func (s *JobService) ToggleFavorite(ctx context.Context, userID, jobID uint, favorite bool) error {
	if userID == 0 {
		return errors.New("user ID required for favorite operations")
	}

	if favorite {
		favoriteRecord := JobFavorite{
			UserID:    userID,
			JobID:     jobID,
			CreatedAt: time.Now(),
		}
		return s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&favoriteRecord).Error
	}

	return s.db.WithContext(ctx).Where("user_id = ? AND job_id = ?", userID, jobID).Delete(&JobFavorite{}).Error
}

// EnsureSeedData 注入默认职位数据
func (s *JobService) EnsureSeedData(ctx context.Context) error {
	// 从context获取租户ID
	tenantID, err := context.GetTenantID(ctx)
	if err != nil {
		tenantID = 1 // 默认租户
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&Job{}).Where("tenant_id = ?", tenantID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now()
	seedJobs := []Job{
		{
			TenantID:           tenantID, // 设置租户ID
			Title:              "资深后端工程师",
			Description:        "负责核心微服务的设计与实现，参与架构评审和性能优化。",
			Requirements:       "精通 Go 语言，熟悉微服务架构，具备数据库优化经验。",
			Responsibilities:   "负责后端服务开发、代码评审与上线保障。",
			CompanyID:          1001,
			CompanyName:        "Zervigo 科技",
			WorkType:           "full-time",
			WorkLocation:       "上海·浦东",
			SalaryMin:          intPtr(30000),
			SalaryMax:          intPtr(45000),
			SalaryCurrency:     "CNY",
			SalaryPeriod:       "monthly",
			ExperienceRequired: "5+年",
			EducationRequired:  "本科及以上",
			SkillsRequired:     pq.StringArray{"Go", "Microservices", "PostgreSQL"},
			Benefits:           pq.StringArray{"五险一金", "年度体检", "弹性工作"},
			Status:             JobStatusPublished,
			CreatedBy:          1,
			PublishAt:          &now,
		},
		{
			TenantID:           tenantID, // 设置租户ID
			Title:              "前端开发工程师",
			Description:        "负责跨端产品的前端开发，配合设计与后端完成需求交付。",
			Requirements:       "熟悉 React/Taro，掌握 TypeScript，具备组件化开发经验。",
			Responsibilities:   "实现高可用前端页面，优化性能与体验。",
			CompanyID:          1002,
			CompanyName:        "星火互动",
			WorkType:           "full-time",
			WorkLocation:       "北京·望京",
			SalaryMin:          intPtr(18000),
			SalaryMax:          intPtr(28000),
			SalaryCurrency:     "CNY",
			SalaryPeriod:       "monthly",
			ExperienceRequired: "3+年",
			EducationRequired:  "本科",
			SkillsRequired:     pq.StringArray{"React", "TypeScript", "Taro"},
			Benefits:           pq.StringArray{"餐补", "下午茶", "年度旅游"},
			Status:             JobStatusPublished,
			CreatedBy:          1,
			PublishAt:          &now,
		},
		{
			TenantID:           tenantID, // 设置租户ID
			Title:              "数据分析师",
			Description:        "负责招聘数据分析，搭建数据指标体系，为业务决策提供支持。",
			Requirements:       "熟练使用 SQL、Python，具备数据建模与可视化经验。",
			Responsibilities:   "产出分析报告，设计数据监控看板。",
			CompanyID:          1003,
			CompanyName:        "星途数据",
			WorkType:           "full-time",
			WorkLocation:       "深圳·南山",
			SalaryMin:          intPtr(20000),
			SalaryMax:          intPtr(32000),
			SalaryCurrency:     "CNY",
			SalaryPeriod:       "monthly",
			ExperienceRequired: "4+年",
			EducationRequired:  "硕士",
			SkillsRequired:     pq.StringArray{"SQL", "Python", "Tableau"},
			Benefits:           pq.StringArray{"十五薪", "补充医疗", "股票期权"},
			Status:             JobStatusPublished,
			CreatedBy:          1,
			PublishAt:          &now,
		},
	}

	for _, job := range seedJobs {
		jobCopy := job
		if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&jobCopy).Error; err != nil {
			return err
		}
	}
	return nil
}

func normalizeTags(tags []string) []string {
	unique := make(map[string]struct{})
	for _, tag := range tags {
		t := strings.TrimSpace(tag)
		if t == "" {
			continue
		}
		t = strings.Title(strings.ToLower(t))
		unique[t] = struct{}{}
	}
	result := make([]string, 0, len(unique))
	for tag := range unique {
		result = append(result, tag)
	}
	return result
}

func stringPtrOrNil(value string) *string {
	v := strings.TrimSpace(value)
	if v == "" {
		return nil
	}
	return &v
}

func intPtr(value int) *int {
	return &value
}

func defaultStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return JobStatusDraft
	}
	return strings.TrimSpace(status)
}
