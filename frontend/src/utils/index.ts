import Taro from '@tarojs/taro'

// 工具函数集合
export class Utils {
  // 格式化时间
  static formatTime(timestamp: number): string {
    const date = new Date(timestamp)
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    
    return `${year}-${month}-${day} ${hours}:${minutes}`
  }

  // 格式化日期
  static formatDate(timestamp: number): string {
    const date = new Date(timestamp)
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    
    return `${year}-${month}-${day}`
  }

  // 相对时间
  static getRelativeTime(timestamp: number): string {
    const now = Date.now()
    const diff = now - timestamp
    
    const minute = 60 * 1000
    const hour = 60 * minute
    const day = 24 * hour
    const week = 7 * day
    const month = 30 * day
    const year = 365 * day

    if (diff < minute) {
      return '刚刚'
    } else if (diff < hour) {
      return `${Math.floor(diff / minute)}分钟前`
    } else if (diff < day) {
      return `${Math.floor(diff / hour)}小时前`
    } else if (diff < week) {
      return `${Math.floor(diff / day)}天前`
    } else if (diff < month) {
      return `${Math.floor(diff / week)}周前`
    } else if (diff < year) {
      return `${Math.floor(diff / month)}个月前`
    } else {
      return `${Math.floor(diff / year)}年前`
    }
  }

  // 防抖函数
  static debounce(func: Function, wait: number) {
    let timeout: NodeJS.Timeout
    return function executedFunction(...args: any[]) {
      const later = () => {
        clearTimeout(timeout)
        func(...args)
      }
      clearTimeout(timeout)
      timeout = setTimeout(later, wait)
    }
  }

  // 节流函数
  static throttle(func: Function, limit: number) {
    let inThrottle: boolean
    return function executedFunction(...args: any[]) {
      if (!inThrottle) {
        func.apply(this, args)
        inThrottle = true
        setTimeout(() => inThrottle = false, limit)
      }
    }
  }

  // 深拷贝
  static deepClone(obj: any): any {
    if (obj === null || typeof obj !== 'object') return obj
    if (obj instanceof Date) return new Date(obj.getTime())
    if (obj instanceof Array) return obj.map(item => Utils.deepClone(item))
    if (typeof obj === 'object') {
      const clonedObj: any = {}
      for (const key in obj) {
        if (obj.hasOwnProperty(key)) {
          clonedObj[key] = Utils.deepClone(obj[key])
        }
      }
      return clonedObj
    }
  }

  // 生成UUID
  static generateUUID(): string {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
      const r = Math.random() * 16 | 0
      const v = c === 'x' ? r : (r & 0x3 | 0x8)
      return v.toString(16)
    })
  }

  // 验证邮箱
  static validateEmail(email: string): boolean {
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    return re.test(email)
  }

  // 验证手机号
  static validatePhone(phone: string): boolean {
    const re = /^1[3-9]\d{9}$/
    return re.test(phone)
  }

  // 验证密码强度
  static validatePassword(password: string): { valid: boolean; message: string } {
    if (password.length < 6) {
      return { valid: false, message: '密码长度不能少于6位' }
    }
    if (password.length > 20) {
      return { valid: false, message: '密码长度不能超过20位' }
    }
    if (!/(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/.test(password)) {
      return { valid: false, message: '密码必须包含大小写字母和数字' }
    }
    return { valid: true, message: '密码强度符合要求' }
  }

  // 显示Toast
  static showToast(title: string, icon: 'success' | 'error' | 'loading' | 'none' = 'none') {
    Taro.showToast({
      title,
      icon,
      duration: 2000
    })
  }

  // 显示Loading
  static showLoading(title: string = '加载中...') {
    Taro.showLoading({ title })
  }

  // 隐藏Loading
  static hideLoading() {
    Taro.hideLoading()
  }

  // 显示确认对话框
  static showModal(title: string, content: string): Promise<boolean> {
    return new Promise((resolve) => {
      Taro.showModal({
        title,
        content,
        success: (res) => {
          resolve(res.confirm)
        }
      })
    })
  }

  // 获取系统信息
  static getSystemInfo() {
    return Taro.getSystemInfoSync()
  }

  // 检查网络状态
  static checkNetworkStatus(): Promise<boolean> {
    return new Promise((resolve) => {
      Taro.getNetworkType({
        success: (res) => {
          resolve(res.networkType !== 'none')
        },
        fail: () => {
          resolve(false)
        }
      })
    })
  }

  // 图片压缩
  static compressImage(imagePath: string, quality: number = 0.8): Promise<string> {
    return new Promise((resolve, reject) => {
      Taro.compressImage({
        src: imagePath,
        quality,
        success: (res) => {
          resolve(res.tempFilePath)
        },
        fail: reject
      })
    })
  }

  // 选择图片
  static chooseImage(count: number = 1): Promise<string[]> {
    return new Promise((resolve, reject) => {
      Taro.chooseImage({
        count,
        sizeType: ['compressed'],
        sourceType: ['album', 'camera'],
        success: (res) => {
          resolve(res.tempFilePaths)
        },
        fail: reject
      })
    })
  }

  // 预览图片
  static previewImage(urls: string[], current?: string) {
    Taro.previewImage({
      urls,
      current: current || urls[0]
    })
  }

  // 复制到剪贴板
  static copyToClipboard(data: string): Promise<boolean> {
    return new Promise((resolve) => {
      Taro.setClipboardData({
        data,
        success: () => resolve(true),
        fail: () => resolve(false)
      })
    })
  }

  // 获取位置信息
  static getLocation(): Promise<{ latitude: number; longitude: number }> {
    return new Promise((resolve, reject) => {
      Taro.getLocation({
        type: 'wgs84',
        success: (res) => {
          resolve({
            latitude: res.latitude,
            longitude: res.longitude
          })
        },
        fail: reject
      })
    })
  }
}

export default Utils
