-- VueCMF 菜单表
CREATE TABLE IF NOT EXISTS vuecmf_menu (
    menu_id SERIAL PRIMARY KEY,
    pid INTEGER DEFAULT 0,
    model_id INTEGER,
    title VARCHAR(50) NOT NULL,
    path VARCHAR(100),
    icon VARCHAR(50),
    sort_num INTEGER DEFAULT 0,
    status INTEGER DEFAULT 10,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_vuecmf_menu_pid ON vuecmf_menu(pid);
CREATE INDEX IF NOT EXISTS idx_vuecmf_menu_status ON vuecmf_menu(status);

-- 插入基础菜单数据
INSERT INTO vuecmf_menu (pid, model_id, title, path, icon, sort_num, status) VALUES
(0, 1, '首页', '/dashboard', 'HomeFilled', 1, 10),
(0, 2, '系统管理', '/system', 'Setting', 2, 10),
(2, 3, '用户管理', '/system/users', 'User', 1, 10),
(2, 4, '角色管理', '/system/roles', 'UserFilled', 2, 10),
(2, 5, '权限管理', '/system/permissions', 'Lock', 3, 10),
(0, 6, '职位管理', '/jobs', 'BriefcaseFilled', 3, 10),
(0, 7, '简历管理', '/resumes', 'DocumentFilled', 4, 10),
(0, 8, '企业管理', '/companies', 'OfficeBuilding', 5, 10)
ON CONFLICT DO NOTHING;

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_vuecmf_menu_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_vuecmf_menu_updated_at ON vuecmf_menu;
CREATE TRIGGER trigger_update_vuecmf_menu_updated_at
    BEFORE UPDATE ON vuecmf_menu
    FOR EACH ROW
    EXECUTE FUNCTION update_vuecmf_menu_updated_at();

-- 输出初始化信息
DO $$
BEGIN
    RAISE NOTICE 'VueCMF 菜单表初始化完成！';
    RAISE NOTICE '已添加 % 条菜单', (SELECT COUNT(*) FROM vuecmf_menu);
END $$;

