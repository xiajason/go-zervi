// ============================================
// VueCMF 一键完整修复脚本
// ============================================
// 在VueCMF前端页面 (http://localhost:8081) 的控制台中执行

console.clear();
console.log('╔════════════════════════════════════════╗');
console.log('║   🔧 VueCMF 一键完整修复脚本          ║');
console.log('╚════════════════════════════════════════╝');
console.log('');

async function fullRepair() {
  try {
    // ========================================
    // 步骤1: 清除所有缓存
    // ========================================
    console.log('📦 步骤1: 清除所有缓存');
    console.log('─'.repeat(50));
    
    // 清除localStorage
    const keysToRemove = [];
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key && key.startsWith('vuecmf_')) {
        keysToRemove.push(key);
      }
    }
    keysToRemove.forEach(key => {
      localStorage.removeItem(key);
      console.log('  ✅ 已删除:', key);
    });
    
    // 清除sessionStorage
    sessionStorage.clear();
    console.log('  ✅ sessionStorage已清除');
    
    // 清除Cookie
    document.cookie.split(";").forEach(c => {
      document.cookie = c.replace(/^ +/, "").replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/");
    });
    console.log('  ✅ Cookie已清除\n');
    
    // ========================================
    // 步骤2: 检查登录状态
    // ========================================
    console.log('🔐 步骤2: 检查登录状态');
    console.log('─'.repeat(50));
    
    const token = localStorage.getItem('vuecmf_token');
    console.log('  Token:', token ? '✅ 存在' : '❌ 不存在');
    
    if (!token) {
      console.log('\n  ⚠️  需要重新登录！');
      console.log('  3秒后跳转到登录页...\n');
      setTimeout(() => {
        location.href = '/';
      }, 3000);
      return;
    }
    
    console.log('  ✅ 已登录\n');
    
    // ========================================
    // 步骤3: 获取菜单和API映射
    // ========================================
    console.log('📋 步骤3: 获取最新的菜单和API映射');
    console.log('─'.repeat(50));
    
    const menuRes = await fetch('http://localhost:9000/api/v1/menu/nav', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({data: {username: 'admin'}})
    });
    
    const menuData = await menuRes.json();
    
    if (menuData.code !== 0) {
      console.error('  ❌ 获取菜单失败:', menuData.message || menuData.msg);
      return;
    }
    
    console.log('  ✅ 菜单获取成功');
    console.log('  菜单数量:', Object.keys(menuData.data.nav_menu).length);
    
    // 保存到localStorage
    localStorage.setItem('vuecmf_menu', JSON.stringify(menuData.data.nav_menu));
    localStorage.setItem('vuecmf_api_maps', JSON.stringify(menuData.data.api_maps));
    console.log('  ✅ 已保存到localStorage\n');
    
    console.log('  API映射:');
    console.log('    • admin.list:', menuData.data.api_maps.admin?.list || '❌');
    console.log('    • roles.list:', menuData.data.api_maps.roles?.list || '❌');
    console.log('    • permissions.list:', menuData.data.api_maps.permissions?.list || '❌');
    console.log('');
    
    // ========================================
    // 步骤4: 测试新的数据格式
    // ========================================
    console.log('🧪 步骤4: 测试新的响应格式（三层嵌套）');
    console.log('─'.repeat(50));
    
    const rolesApiUrl = menuData.data.api_maps.roles?.list;
    if (!rolesApiUrl) {
      console.error('  ❌ 未找到roles.list映射');
      return;
    }
    
    console.log('  测试URL:', 'http://localhost:9000' + rolesApiUrl);
    
    const rolesRes = await fetch('http://localhost:9000' + rolesApiUrl, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({data: {table_name: 'roles', page: 1, page_size: 20}})
    });
    
    const rolesData = await rolesRes.json();
    
    console.log('\n  响应结构检查:');
    console.log('    • code:', rolesData.code === 0 ? '✅ 0' : '❌ ' + rolesData.code);
    console.log('    • data:', rolesData.data ? '✅ 存在' : '❌ 不存在');
    console.log('    • data.data:', rolesData.data?.data ? '✅ 存在' : '❌ 不存在');
    console.log('    • field_info:', rolesData.data?.data?.field_info ? '✅ 存在' : '❌ 不存在');
    console.log('    • list:', rolesData.data?.data?.list ? '✅ 存在' : '❌ 不存在');
    console.log('');
    console.log('  数据详情:');
    console.log('    • 字段数量:', rolesData.data?.data?.field_info?.length || 0);
    console.log('    • 数据总数:', rolesData.data?.data?.total || 0);
    console.log('    • 列表数量:', rolesData.data?.data?.list?.length || 0);
    
    if (rolesData.data?.data?.field_info && rolesData.data?.data?.list) {
      console.log('\n  ✅ 格式完全正确！');
      console.log('\n  角色数据预览:');
      rolesData.data.data.list.forEach((role, idx) => {
        console.log(`    ${idx+1}. ${role.role_name} - ${role.description}`);
      });
      
      console.log('\n  字段配置预览:');
      rolesData.data.data.field_info.forEach((field, idx) => {
        console.log(`    ${idx+1}. ${field.label} (${field.prop})`);
      });
      
      // ========================================
      // 步骤5: 测试所有表
      // ========================================
      console.log('\n🔍 步骤5: 测试所有管理页面');
      console.log('─'.repeat(50));
      
      const tables = [
        {name: '用户管理', table: 'admin'},
        {name: '角色管理', table: 'roles'},
        {name: '权限管理', table: 'permissions'}
      ];
      
      for (const t of tables) {
        const apiUrl = menuData.data.api_maps[t.table]?.list;
        if (apiUrl) {
          const res = await fetch('http://localhost:9000' + apiUrl, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({data: {table_name: t.table, page: 1, page_size: 20}})
          });
          const data = await res.json();
          const hasCorrectFormat = data.data?.data?.field_info && data.data?.data?.list;
          console.log(`  ${hasCorrectFormat ? '✅' : '❌'} ${t.name}: 字段=${data.data?.data?.field_info?.length || 0}, 数据=${data.data?.data?.total || 0}`);
        } else {
          console.error(`  ❌ ${t.name}: 未找到API映射`);
        }
      }
      
      // ========================================
      // 完成
      // ========================================
      console.log('\n' + '═'.repeat(50));
      console.log('🎉🎉🎉 修复完成！所有格式正确！');
      console.log('═'.repeat(50));
      console.log('\n正在刷新页面...');
      console.log('刷新后请点击"系统管理" → "角色管理"查看数据\n');
      
      setTimeout(() => {
        location.reload();
      }, 3000);
      
    } else {
      console.error('\n❌ 响应格式仍然不对！');
      console.log('\n完整响应:');
      console.log(JSON.stringify(rolesData, null, 2));
    }
    
  } catch (error) {
    console.error('❌ 修复过程出错:', error);
    console.log('\n建议:');
    console.log('1. 检查后端服务是否运行');
    console.log('2. 手动刷新页面并重新登录');
  }
}

// 执行修复
fullRepair();





