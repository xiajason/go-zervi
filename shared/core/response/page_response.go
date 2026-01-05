package response

// PageResponse 分页响应格式
type PageResponse struct {
	List     interface{} `json:"list"`      // 数据列表（必须是list，不是records）
	Total    int64       `json:"total"`     // 总记录数
	PageNum  int         `json:"pageNum"`   // 当前页码
	PageSize int         `json:"pageSize"`  // 每页大小
	Pages    int         `json:"pages"`     // 总页数
}

// NewPageResponse 创建分页响应
func NewPageResponse(list interface{}, total int64, pageNum, pageSize int) *PageResponse {
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}
	
	return &PageResponse{
		List:     list,
		Total:    total,
		PageNum:  pageNum,
		PageSize: pageSize,
		Pages:    pages,
	}
}

// PageRequest 分页请求参数
type PageRequest struct {
	PageNum  int `json:"pageNum" form:"pageNum"`   // 当前页码，默认1
	PageSize int `json:"pageSize" form:"pageSize"` // 每页大小，默认10
}

// GetOffset 获取偏移量
func (p *PageRequest) GetOffset() int {
	if p.PageNum <= 0 {
		p.PageNum = 1
	}
	return (p.PageNum - 1) * p.GetPageSize()
}

// GetPageSize 获取每页大小
func (p *PageRequest) GetPageSize() int {
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100 // 限制最大每页100条
	}
	return p.PageSize
}
