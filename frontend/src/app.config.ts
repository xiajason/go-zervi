export default defineAppConfig({
  pages: [
    'pages/index/index',
    'pages/login/index',
    'pages/register/index',
    'pages/profile/index',
    'pages/resume/index',
    'pages/job/index',
    'pages/company/index',
    'pages/chat/index',
    'pages/search/index'
  ],
  window: {
    backgroundTextStyle: 'light',
    navigationBarBackgroundColor: '#fff',
    navigationBarTitleText: 'Zervigo MVP',
    navigationBarTextStyle: 'black'
  },
  tabBar: {
    color: '#666',
    selectedColor: '#1296db',
    backgroundColor: '#fafafa',
    borderStyle: 'black',
    list: [
      {
        pagePath: 'pages/index/index',
        iconPath: 'assets/icons/home.png',
        selectedIconPath: 'assets/icons/home-active.png',
        text: '首页'
      },
      {
        pagePath: 'pages/job/index',
        iconPath: 'assets/icons/job.png',
        selectedIconPath: 'assets/icons/job-active.png',
        text: '职位'
      },
      {
        pagePath: 'pages/resume/index',
        iconPath: 'assets/icons/resume.png',
        selectedIconPath: 'assets/icons/resume-active.png',
        text: '简历'
      },
      {
        pagePath: 'pages/chat/index',
        iconPath: 'assets/icons/chat.png',
        selectedIconPath: 'assets/icons/chat-active.png',
        text: 'AI助手'
      },
      {
        pagePath: 'pages/profile/index',
        iconPath: 'assets/icons/profile.png',
        selectedIconPath: 'assets/icons/profile-active.png',
        text: '我的'
      }
    ]
  },
  networkTimeout: {
    request: 10000,
    downloadFile: 10000
  },
  debug: true,
  sitemapLocation: 'sitemap.json'
})
