// 用户相关类型
export interface User {
  userId: number
  username: string
  email: string
  phone: string
  realName: string
  avatar: string
  gender: number
  birthday: string
  location: string
  bio: string
  status: number
  createTime: number
  updateTime: number
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
  expiresIn: number
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
  phone: string
  realName: string
}

// 职位相关类型
export interface Job {
  jobId: number
  companyId: number
  companyName: string
  jobTitle: string
  jobDescription: string
  jobRequirements: string
  jobType: string
  workLocation: string
  salaryMin: number
  salaryMax: number
  experience: string
  education: string
  skills: string[]
  benefits: string[]
  status: number
  createTime: number
  updateTime: number
}

export interface JobSearchRequest {
  keyword?: string
  location?: string
  jobType?: string
  salaryMin?: number
  salaryMax?: number
  experience?: string
  education?: string
  skills?: string[]
  page: number
  pageSize: number
}

export interface JobSearchResponse {
  list: Job[]
  total: number
  page: number
  pageSize: number
}

// 简历相关类型
export interface Resume {
  resumeId: number
  userId: number
  resumeName: string
  personalInfo: PersonalInfo
  workExperience: WorkExperience[]
  education: Education[]
  skills: Skill[]
  projects: Project[]
  certificates: Certificate[]
  status: number
  createTime: number
  updateTime: number
}

export interface PersonalInfo {
  realName: string
  email: string
  phone: string
  gender: number
  birthday: string
  location: string
  bio: string
  avatar: string
}

export interface WorkExperience {
  experienceId: number
  companyName: string
  position: string
  startDate: string
  endDate: string
  description: string
  achievements: string[]
}

export interface Education {
  educationId: number
  schoolName: string
  major: string
  degree: string
  startDate: string
  endDate: string
  description: string
}

export interface Skill {
  skillId: number
  skillName: string
  skillLevel: string
  category: string
}

export interface Project {
  projectId: number
  projectName: string
  projectDescription: string
  startDate: string
  endDate: string
  technologies: string[]
  responsibilities: string[]
}

export interface Certificate {
  certificateId: number
  certificateName: string
  issuingOrganization: string
  issueDate: string
  expiryDate: string
  certificateUrl: string
}

// 企业相关类型
export interface Company {
  companyId: number
  companyName: string
  companyLogo: string
  companyDescription: string
  industry: string
  companySize: string
  website: string
  address: string
  city: string
  province: string
  country: string
  contactPerson: string
  contactPhone: string
  contactEmail: string
  status: number
  verificationStatus: number
  createTime: number
  updateTime: number
}

// AI相关类型
export interface AIMatchRequest {
  resumeId: number
  jobId?: number
  matchType: string
  parameters?: Record<string, any>
}

export interface AIMatchResponse {
  matchId: number
  matchScore: number
  matchDetails: MatchDetails
  recommendations: Recommendation[]
  analysis: AnalysisResult
}

export interface MatchDetails {
  skillsMatch: number
  experienceMatch: number
  educationMatch: number
  locationMatch: number
  salaryMatch: number
  cultureMatch: number
}

export interface Recommendation {
  type: string
  title: string
  description: string
  priority: number
  action: string
}

export interface AnalysisResult {
  strengths: string[]
  weaknesses: string[]
  suggestions: string[]
  riskFactors: string[]
}

export interface AIChatRequest {
  message: string
  context?: string
  userId: number
  chatType: string
  sessionId?: string
}

export interface AIChatResponse {
  responseId: number
  message: string
  responseType: string
  suggestions?: string[]
  actions?: ChatAction[]
  sessionId: string
  timestamp: number
}

export interface ChatAction {
  actionType: string
  actionData: Record<string, any>
  description: string
}

// 区块链相关类型
export interface BlockchainTransaction {
  transactionId: string
  transactionType: string
  entityId: string
  versionSource: string
  oldValue: string
  newValue: string
  changeReason: string
  operatorId: string
  transactionHash: string
  transactionData: string
  status: string
  blockHeight: number
  createTime: number
  confirmTime: number
}

export interface BlockchainStats {
  totalTransactions: number
  totalBlocks: number
  activeTransactions: number
  pendingTransactions: number
  confirmedTransactions: number
  failedTransactions: number
  averageBlockTime: number
  networkHashRate: number
}

// 通用响应类型
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
  timestamp?: number
}

export interface PaginationResponse<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

// 页面状态类型
export interface PageState {
  loading: boolean
  error: string | null
  data: any
}

// 表单状态类型
export interface FormState {
  values: Record<string, any>
  errors: Record<string, string>
  touched: Record<string, boolean>
  submitting: boolean
}

// 导航类型
export interface TabBarItem {
  pagePath: string
  iconPath: string
  selectedIconPath: string
  text: string
}

// 主题类型
export interface Theme {
  primaryColor: string
  secondaryColor: string
  backgroundColor: string
  textColor: string
  borderColor: string
}

// 环境配置类型
export interface EnvConfig {
  NODE_ENV: string
  API_BASE_URL: string
  WS_BASE_URL: string
}
