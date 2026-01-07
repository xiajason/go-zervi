// 在浏览器控制台执行此脚本来诊断问题
console.clear();
console.log('=== VueCMF 前端诊断 ===\n');

// 1. 检查菜单数据
console.log('1. 检查 localStorage 中的菜单数据:');
const menuData = localStorage.getItem('vuecmf_menu');
if (menuData) {
  console.log('✅ 菜单数据存在');
  try {
    const parsed = JSON.parse(menuData);
    console.log('菜单数量:', Object.keys(parsed).length);
    console.log('菜单数据:', parsed);
  } catch (e) {
    console.error('❌ 菜单数据解析失败:', e);
  }
} else {
  console.warn('⚠️ localStorage 中没有菜单数据');
}

// 2. 检查API映射数据
console.log('\n2. 检查 localStorage 中的 API 映射:');
const apiMaps = localStorage.getItem('vuecmf_api_maps');
if (apiMaps) {
  console.log('✅ API映射存在');
  try {
    const parsed = JSON.parse(apiMaps);
    console.log('API映射数量:', Object.keys(parsed).length);
  } catch (e) {
    console.error('❌ API映射解析失败:', e);
  }
} else {
  console.warn('⚠️ localStorage 中没有 API 映射');
}

// 3. 检查token
console.log('\n3. 检查认证token:');
const token = localStorage.getItem('vuecmf_token');
if (token) {
  console.log('✅ Token存在:', token.substring(0, 20) + '...');
} else {
  console.warn('⚠️ Token不存在');
}

// 4. 检查Vue实例和store
console.log('\n4. 检查前端状态:');
if (typeof window !== 'undefined') {
  console.log('Window对象:', window.location.href);
  
  // 尝试访问Vue store（如果使用Pinia）
  if (window.__PINIA__) {
    console.log('✅ Pinia store存在');
  }
}

// 5. 测试菜单API
console.log('\n5. 测试菜单API请求:');
fetch('http://localhost:9000/api/v1/menu/nav', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ data: { username: 'admin' } })
})
.then(res => res.json())
.then(data => {
  console.log('✅ 菜单API响应成功');
  console.log('响应代码:', data.code);
  console.log('菜单数量:', Object.keys(data.data.nav_menu || {}).length);
  console.log('第一个菜单:', data.data.menu_order ? data.data.menu_order[0] : '未知');
  
  if (data.data.nav_menu) {
    const firstKey = data.data.menu_order ? data.data.menu_order[0] : Object.keys(data.data.nav_menu)[0];
    const firstMenu = data.data.nav_menu[firstKey];
    console.log('第一个菜单详情:', firstMenu);
    if (firstMenu && firstMenu.children) {
      console.log('✅ 第一个菜单有children:', Object.keys(firstMenu.children).length, '个');
    } else {
      console.warn('⚠️ 第一个菜单没有children');
    }
  }
})
.catch(err => {
  console.error('❌ 菜单API请求失败:', err);
});

// 6. 建议
console.log('\n=== 诊断建议 ===');
console.log('如果看到错误，请：');
console.log('1. 截图控制台所有输出');
console.log('2. 点击左上角汉堡菜单（☰）展开侧边栏');
console.log('3. 切换到"网络"标签查看API请求状态');
console.log('4. 尝试刷新页面（Ctrl+Shift+R 或 Cmd+Shift+R）');

