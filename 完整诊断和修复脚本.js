// ============================================
// VueCMF 完整诊断和修复脚本
// ============================================
// 在浏览器控制台中执行此脚本

console.clear();
console.log('=== VueCMF 完整诊断和修复 ===\n');
console.log('开始时间:', new Date().toLocaleTimeString());
console.log('━'.repeat(60));

// ============================================
// 第1步：检查localStorage
// ============================================
console.log('\n📦 第1步：检查 localStorage');
console.log('─'.repeat(60));

const token = localStorage.getItem('vuecmf_token');
const menuData = localStorage.getItem('vuecmf_menu');
const apiMapsStr = localStorage.getItem('vuecmf_api_maps');

console.log('Token:', token ? '✅ 存在 (' + token.substring(0, 20) + '...)' : '❌ 不存在');
console.log('Menu:', menuData ? '✅ 存在' : '❌ 不存在');
console.log('API Maps:', apiMapsStr ? '✅ 存在' : '❌ 不存在');

let apiMaps = {};
if (apiMapsStr) {
    try {
        apiMaps = JSON.parse(apiMapsStr);
        console.log('\nAPI映射详情:');
        console.log('  admin.index:', apiMaps.admin?.index || '❌ 未配置');
        console.log('  roles.index:', apiMaps.roles?.index || '❌ 未配置');
        console.log('  permissions.index:', apiMaps.permissions?.index || '❌ 未配置');
    } catch (e) {
        console.error('❌ API映射解析失败:', e);
    }
}

// ============================================
// 第2步：重新获取菜单和API映射
// ============================================
console.log('\n🔄 第2步：重新获取菜单和API映射');
console.log('─'.repeat(60));

async function refreshMenuAndApiMaps() {
    try {
        const response = await fetch('http://localhost:9000/api/v1/menu/nav', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ data: { username: 'admin' } })
        });
        
        const data = await response.json();
        
        if (data.code === 0) {
            console.log('✅ 菜单API响应成功');
            console.log('   菜单数量:', Object.keys(data.data.nav_menu || {}).length);
            
            // 保存到localStorage
            localStorage.setItem('vuecmf_menu', JSON.stringify(data.data.nav_menu));
            localStorage.setItem('vuecmf_api_maps', JSON.stringify(data.data.api_maps));
            
            console.log('\n新的API映射:');
            console.log('  admin.index:', data.data.api_maps.admin?.index || '❌ 未配置');
            console.log('  roles.index:', data.data.api_maps.roles?.index || '❌ 未配置');
            console.log('  permissions.index:', data.data.api_maps.permissions?.index || '❌ 未配置');
            
            return data.data.api_maps;
        } else {
            console.error('❌ 菜单API失败:', data.message || data.msg);
            return null;
        }
    } catch (error) {
        console.error('❌ 菜单API请求失败:', error);
        return null;
    }
}

// ============================================
// 第3步：测试所有CRUD API
// ============================================
async function testAllApis(apiMaps) {
    console.log('\n🧪 第3步：测试所有CRUD API');
    console.log('─'.repeat(60));
    
    const tests = [
        { name: '用户管理', table: 'admin', apiPath: apiMaps.admin?.index },
        { name: '角色管理', table: 'roles', apiPath: apiMaps.roles?.index },
        { name: '权限管理', table: 'permissions', apiPath: apiMaps.permissions?.index }
    ];
    
    for (const test of tests) {
        console.log(`\n测试 ${test.name} (${test.table}):`);
        
        if (!test.apiPath) {
            console.error('  ❌ API路径未配置');
            continue;
        }
        
        console.log('  API路径:', test.apiPath);
        
        try {
            // 方法1：使用API映射路径
            const url1 = 'http://localhost:9000' + test.apiPath;
            console.log('  尝试方法1:', url1);
            
            const response1 = await fetch(url1, {
                method: 'POST',
                headers: { 
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer ' + (token || '')
                },
                body: JSON.stringify({
                    data: {
                        table_name: test.table,
                        page: 1,
                        page_size: 20
                    }
                })
            });
            
            const data1 = await response1.json();
            
            if (data1.code === 0) {
                console.log('  ✅ 成功! 数据量:', data1.data?.total || 0);
                console.log('  数据预览:', data1.data?.list?.slice(0, 2) || []);
            } else {
                console.error('  ❌ 失败 (code=' + data1.code + '):', data1.message || data1.msg);
                
                // 方法2：尝试直接路径
                const url2 = `http://localhost:9000/api/v1/${test.table}/index`;
                console.log('  尝试方法2:', url2);
                
                const response2 = await fetch(url2, {
                    method: 'POST',
                    headers: { 
                        'Content-Type': 'application/json',
                        'Authorization': 'Bearer ' + (token || '')
                    },
                    body: JSON.stringify({
                        data: {
                            table_name: test.table,
                            page: 1,
                            page_size: 20
                        }
                    })
                });
                
                const data2 = await response2.json();
                
                if (data2.code === 0) {
                    console.log('  ✅ 方法2成功! 数据量:', data2.data?.total || 0);
                } else {
                    console.error('  ❌ 方法2也失败:', data2.message || data2.msg);
                }
            }
        } catch (error) {
            console.error('  ❌ 请求异常:', error);
        }
    }
}

// ============================================
// 第4步：检查VueCMF组件状态
// ============================================
console.log('\n🔍 第4步：检查VueCMF组件状态');
console.log('─'.repeat(60));

function checkVueCMFState() {
    // 检查Vue实例
    if (typeof window.__VUE_DEVTOOLS_GLOBAL_HOOK__ !== 'undefined') {
        console.log('✅ Vue DevTools已连接');
    } else {
        console.log('⚠️  Vue DevTools未连接');
    }
    
    // 检查Pinia store
    if (typeof window.__PINIA__ !== 'undefined') {
        console.log('✅ Pinia store存在');
        // 尝试获取store中的api_maps
        try {
            const stores = window.__PINIA__.state.value;
            console.log('Store keys:', Object.keys(stores));
        } catch (e) {
            console.log('无法读取store详情');
        }
    } else {
        console.log('⚠️  Pinia store不存在');
    }
}

checkVueCMFState();

// ============================================
// 执行诊断
// ============================================
console.log('\n' + '━'.repeat(60));
console.log('开始执行诊断...\n');

(async () => {
    // 获取新的API映射
    const newApiMaps = await refreshMenuAndApiMaps();
    
    if (newApiMaps) {
        // 测试所有API
        await testAllApis(newApiMaps);
        
        // ============================================
        // 第5步：给出修复建议
        // ============================================
        console.log('\n💡 第5步：修复建议');
        console.log('─'.repeat(60));
        
        const hasCorrectMappings = 
            newApiMaps.admin?.index === '/api/v1/admin/index' &&
            newApiMaps.roles?.index === '/api/v1/roles/index' &&
            newApiMaps.permissions?.index === '/api/v1/permissions/index';
        
        if (hasCorrectMappings) {
            console.log('✅ API映射配置正确');
            console.log('\n建议操作:');
            console.log('1. 刷新页面: location.reload()');
            console.log('2. 或重新点击左侧菜单项');
            console.log('3. 如果还不行，检查网络标签中的实际请求');
        } else {
            console.log('❌ API映射配置不正确');
            console.log('\n需要修复数据库中的API映射表');
            console.log('请联系管理员运行修复SQL脚本');
        }
        
        console.log('\n' + '━'.repeat(60));
        console.log('诊断完成！');
        console.log('如果数据仍然不显示，请:');
        console.log('1. 查看"网络"标签中的实际API请求');
        console.log('2. 检查响应内容');
        console.log('3. 查看是否有JavaScript错误');
    } else {
        console.error('\n❌ 无法获取API映射，请检查后端服务是否正常运行');
    }
})();





