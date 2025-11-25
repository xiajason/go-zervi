/**
 * 菜单服务类 - 本地演示版本
 * 不依赖后端API，使用本地模拟数据
 * 保留所有智能功能和优雅降级策略
 */

import { ref, type Ref } from 'vue'
import router from '@/router'
import { ElMessage } from 'element-plus'

export interface MenuItem {
  id: number
  pid: number
  model_id?: number
  title: string
  path: string
  icon?: string
  component_path?: string  // 组件路径
  sort_num?: number
  status?: number
  children?: MenuItem[]
}

export interface MenuTreeNode extends MenuItem {
  children?: MenuTreeNode[]
}

/**
 * 菜单服务类
 */
export class MenuService {
  // 菜单数据
  menuList = ref<MenuItem[]>([])
  menuTree = ref<MenuTreeNode[]>([])
  
  // 当前选中状态
  activeMenuPath = ref<string>('')
  
  // 面包屑
  breadcrumbs = ref<string[]>([])
  
  // 是否正在加载
  loading = ref<boolean>(false)
  
  // 【P1集成】服务组合信息
  serviceCombination = ref<string>('minimal')
  availableServices = ref<string[]>([])
  
  // 【P1集成】是否使用Router Service（智能模式）
  useRouterService = ref<boolean>(true)

  constructor() {
    // 初始化时可以预加载菜单
  }

  /**
   * 加载菜单数据 - 本地演示版本
   */
  async loadMenu(): Promise<void> {
    try {
      this.loading.value = true
      
      console.log('🚀 本地演示模式：加载模拟菜单数据')
      
      // 使用本地模拟数据
      const mockMenuData = this.getMockMenuData()
      
      this.menuList.value = mockMenuData
      
      // 构建树形结构
      this.menuTree.value = this.buildMenuTree(this.menuList.value)
      
      // 注册动态路由
      this.registerRoutes(this.menuTree.value)
      
      console.log('✅ 本地模拟菜单加载成功', this.menuTree.value)
      
      // 设置服务组合信息
      this.serviceCombination.value = 'all_services'
      this.availableServices.value = ['system', 'jobs', 'resume', 'company']
      
      console.log('🎯 当前服务组合: 完整模式 (all_services)')
      
    } catch (error: any) {
      console.error('❌ 菜单加载失败:', error)
      
      // 使用默认菜单（优雅降级）
      this.useDefaultMenu()
      
      if (import.meta.env.DEV) {
        ElMessage.error('菜单加载失败：' + error.message)
      }
    } finally {
      this.loading.value = false
    }
  }
  
  /**
   * 获取本地模拟菜单数据
   */
  private getMockMenuData(): MenuItem[] {
    return [
      // 系统管理
      {
        id: 1,
        pid: 0,
        title: '系统管理',
        path: '/system',
        icon: 'Setting',
        status: 10
      },
      {
        id: 2,
        pid: 1,
        title: '用户管理',
        path: '/system/users',
        icon: 'User',
        component_path: 'system/Users',
        status: 10
      },
      {
        id: 3,
        pid: 1,
        title: '角色管理',
        path: '/system/roles',
        icon: 'UserFilled',
        component_path: 'system/Roles',
        status: 10
      },
      {
        id: 4,
        pid: 1,
        title: '权限管理',
        path: '/system/permissions',
        icon: 'Lock',
        component_path: 'system/Permissions',
        status: 10
      },
      
      // 职位管理
      {
        id: 5,
        pid: 0,
        title: '职位管理',
        path: '/jobs',
        icon: 'BriefcaseFilled',
        status: 10
      },
      {
        id: 6,
        pid: 5,
        title: '职位列表',
        path: '/jobs/list',
        icon: 'List',
        component_path: 'jobs/JobList',
        status: 10
      },
      {
        id: 7,
        pid: 5,
        title: '职位分类',
        path: '/jobs/categories',
        icon: 'Operation',
        component_path: 'jobs/Categories',
        status: 10
      },
      
      // 简历管理
      {
        id: 8,
        pid: 0,
        title: '简历管理',
        path: '/resume',
        icon: 'DocumentFilled',
        status: 10
      },
      {
        id: 9,
        pid: 8,
        title: '简历库',
        path: '/resume/library',
        icon: 'Document',
        component_path: 'resume/Library',
        status: 10
      },
      {
        id: 10,
        pid: 8,
        title: '简历解析',
        path: '/resume/parse',
        icon: 'Operation',
        component_path: 'resume/Parse',
        status: 10
      },
      
      // 企业管理
      {
        id: 11,
        pid: 0,
        title: '企业管理',
        path: '/company',
        icon: 'OfficeBuilding',
        status: 10
      },
      {
        id: 12,
        pid: 11,
        title: '企业列表',
        path: '/company/list',
        icon: 'List',
        component_path: 'company/CompanyList',
        status: 10
      },
      {
        id: 13,
        pid: 11,
        title: '企业认证',
        path: '/company/auth',
        icon: 'Lock',
        component_path: 'company/Auth',
        status: 10
      },
      
      // 数据统计
      {
        id: 14,
        pid: 0,
        title: '数据统计',
        path: '/statistics',
        icon: 'DataAnalysis',
        status: 10
      },
      {
        id: 15,
        pid: 14,
        title: '用户统计',
        path: '/statistics/user',
        icon: 'User',
        component_path: 'statistics/UserStats',
        status: 10
      },
      {
        id: 16,
        pid: 14,
        title: '业务统计',
        path: '/statistics/business',
        icon: 'Operation',
        component_path: 'statistics/BusinessStats',
        status: 10
      }
    ]
  }
  
  /**
   * 【演示版本】检测当前的服务组合
   */
  private async detectServiceCombination(): Promise<void> {
    // 本地演示版本，直接返回完整模式
    this.serviceCombination.value = 'all_services'
    this.availableServices.value = ['system', 'jobs', 'resume', 'company']
    
    console.log('🎯 本地演示模式：完整服务组合 (all_services)')
    console.log('📋 可用服务:', this.availableServices.value)
    
    // 显示提示
    this.showCombinationTip()
  }
  
  /**
   * 【P1集成】显示服务组合提示
   */
  private showCombinationTip(): void {
    const tips: Record<string, string> = {
      'minimal': '基础模式：仅系统管理功能可用',
      'job_only': '职位模式：职位管理功能已启用',
      'resume_only': '简历模式：简历管理功能已启用',
      'company_only': '企业模式：企业管理功能已启用',
      'job_resume': '职位+简历模式',
      'job_company': '职位+企业模式',
      'resume_company': '简历+企业模式',
      'all_services': '完整模式：所有业务功能已启用'
    }
    
    const tip = tips[this.serviceCombination.value]
    if (tip && import.meta.env.DEV) {
      console.log(`💡 ${tip}`)
    }
  }
  
  /**
   * 使用默认菜单（降级方案）
   */
  private useDefaultMenu(): void {
    console.log('🔄 使用默认菜单')
    
    // 设置默认菜单数据
    const defaultMenus: MenuItem[] = [
      {
        id: 1,
        pid: 0,
        title: '系统管理',
        path: '/system',
        icon: 'Setting',
        status: 10
      },
      {
        id: 2,
        pid: 1,
        title: '用户管理',
        path: '/system/users',
        icon: 'User',
        component_path: 'system/Users',
        status: 10
      },
      {
        id: 3,
        pid: 1,
        title: '角色管理',
        path: '/system/roles',
        icon: 'UserFilled',
        component_path: 'system/Roles',
        status: 10
      },
      {
        id: 4,
        pid: 1,
        title: '权限管理',
        path: '/system/permissions',
        icon: 'Lock',
        component_path: 'system/Permissions',
        status: 10
      }
    ]
    
    this.menuList.value = defaultMenus
    this.menuTree.value = this.buildMenuTree(defaultMenus)
    this.registerRoutes(this.menuTree.value)
    
    console.log('✅ 默认菜单已加载')
  }

  /**
   * 构建菜单树
   */
  private buildMenuTree(menuList: MenuItem[]): MenuTreeNode[] {
    const menuMap = new Map<number, MenuTreeNode>()
    const rootMenus: MenuTreeNode[] = []

    // 第一遍：创建所有节点
    menuList.forEach(menu => {
      menuMap.set(menu.id, { ...menu, children: [] })
    })

    // 第二遍：建立父子关系
    menuList.forEach(menu => {
      const node = menuMap.get(menu.id)!
      if (menu.pid === 0) {
        // 顶级菜单
        rootMenus.push(node)
      } else {
        // 子菜单
        const parent = menuMap.get(menu.pid)
        if (parent) {
          if (!parent.children) {
            parent.children = []
          }
          parent.children.push(node)
        }
      }
    })

    return rootMenus
  }

  /**
   * 注册动态路由
   * 借鉴 VueCMF 的智能组件查找机制
   */
  private registerRoutes(menuTree: MenuTreeNode[]): void {
    // 加载所有模板组件（使用 glob 导入）
    const modules = import.meta.glob('@/views/template/**/*.vue')
    
    if (import.meta.env.DEV) {
      console.log('[路由加载] 可用的模板文件:', Object.keys(modules))
    }

    const registerNode = (menu: MenuTreeNode) => {
      if (menu.children && menu.children.length > 0) {
        // 有子菜单，递归注册
        menu.children.forEach(child => registerNode(child))
      } else {
        // 叶子节点，注册路由
        this.registerSingleRoute(menu, modules)
      }
    }

    menuTree.forEach(menu => registerNode(menu))
    
    console.log('✅ 动态路由注册完成')
  }

  /**
   * 注册单个路由
   */
  private registerSingleRoute(
    menu: MenuTreeNode, 
    modules: Record<string, () => Promise<any>>
  ): void {
    // 确定组件路径
    const componentPath = menu.component_path || this.getDefaultComponentPath(menu.path)
    
    // 智能查找组件（借鉴 VueCMF 的方法）
    const componentKey = Object.keys(modules).find(key => 
      key.includes(componentPath) || 
      key.endsWith(`/${componentPath}.vue`)
    )

    // 开发环境警告
    if (!componentKey && import.meta.env.DEV) {
      console.warn(`[路由加载] 找不到模板文件: ${componentPath}，将使用通用模板`)
    }

    // 动态添加路由
    try {
      router.addRoute('Layout', {
        path: menu.path,
        name: this.getRouteName(menu.path),
        component: componentKey 
          ? modules[componentKey] 
          : () => import('@/views/template/Common.vue'),
        meta: {
          title: menu.title,
          icon: menu.icon,
          menuId: menu.id,
          requiresAuth: true
        }
      })

      if (import.meta.env.DEV) {
        console.log(`[路由注册] ${menu.title} -> ${menu.path}`, componentKey ? '✅' : '⚠️ 使用通用模板')
      }
    } catch (error) {
      console.error(`[路由注册失败] ${menu.title}:`, error)
    }
  }

  /**
   * 根据路径推断默认组件路径
   */
  private getDefaultComponentPath(path: string): string {
    // /system/users -> system/Users
    const parts = path.split('/').filter(p => p)
    if (parts.length === 0) return 'Common'
    
    // 将最后一部分首字母大写
    const lastPart = parts[parts.length - 1]
    if (!lastPart) return 'Common'
    
    const capitalized = lastPart.charAt(0).toUpperCase() + lastPart.slice(1)
    
    if (parts.length === 1) {
      return capitalized
    } else {
      return parts.slice(0, -1).join('/') + '/' + capitalized
    }
  }

  /**
   * 获取路由名称
   */
  private getRouteName(path: string): string {
    return path.replace(/\//g, '-').replace(/^-/, '')
  }

  /**
   * 设置当前激活菜单
   */
  setActiveMenu(path: string): void {
    this.activeMenuPath.value = path
  }

  /**
   * 更新面包屑
   */
  updateBreadcrumbs(path: string): void {
    const breadcrumbs: string[] = []
    
    const findPath = (menus: MenuTreeNode[], targetPath: string, parents: string[] = []): boolean => {
      for (const menu of menus) {
        const currentPath = [...parents, menu.title]
        
        if (menu.path === targetPath) {
          breadcrumbs.push(...currentPath)
          return true
        }
        
        if (menu.children && menu.children.length > 0) {
          if (findPath(menu.children, targetPath, currentPath)) {
            return true
          }
        }
      }
      return false
    }

    findPath(this.menuTree.value, path)
    this.breadcrumbs.value = breadcrumbs
  }

  /**
   * 刷新菜单
   */
  async refresh(): Promise<void> {
    await this.loadMenu()
  }
}

// 导出单例
export const menuService = new MenuService()


