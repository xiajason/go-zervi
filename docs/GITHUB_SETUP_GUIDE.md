# 📦 GitHub 仓库上传指南

## 快速开始

### 方式一：使用自动化脚本（推荐）

```bash
# 在项目根目录运行
./scripts/setup-github-repo.sh
```

脚本会自动引导您完成：
1. 检查 Git 状态
2. 配置远程仓库
3. 提交未跟踪的文件
4. 推送到 GitHub

---

### 方式二：手动操作

#### 步骤 1: 在 GitHub 上创建仓库

1. 访问 https://github.com/new
2. 填写仓库信息：
   - **Repository name**: `go-zervi-framework` 或您喜欢的名称
   - **Description**: `Go-Zervi Framework - 一个创新的Go微服务框架`
   - **Visibility**: 选择 Public 或 Private
   - **⚠️ 不要**勾选 "Initialize this repository with a README"
3. 点击 "Create repository"

#### 步骤 2: 配置远程仓库

```bash
# 添加远程仓库（替换为您的仓库 URL）
git remote add origin https://github.com/您的用户名/仓库名.git

# 或使用 SSH（如果您配置了 SSH 密钥）
git remote add origin git@github.com:您的用户名/仓库名.git

# 验证远程仓库
git remote -v
```

#### 步骤 3: 提交并推送代码

```bash
# 查看当前状态
git status

# 添加所有文件
git add .

# 提交更改
git commit -m "chore: initial commit - Go-Zervi Framework"

# 推送到 GitHub（首次推送需要设置 upstream）
git push -u origin main
# 如果您的默认分支是 master，使用：
# git push -u origin master
```

---

## 注意事项

### 1. 大文件处理

项目可能包含一些较大的文件（如 `.gocache`、编译产物等）。`.gitignore` 已经配置了忽略规则，但如果您之前已经提交了大文件，需要清理：

```bash
# 查看大文件
git ls-files | xargs du -h | sort -rh | head -20

# 如果发现需要忽略的大文件已经在 Git 历史中
git rm --cached 文件路径
git commit -m "chore: remove large files"
```

### 2. 敏感信息检查

推送前请确保没有提交敏感信息：

```bash
# 检查 .env 文件
grep -r "password\|secret\|key" --include="*.env" --include="*.yaml" --include="*.yml" .

# 检查配置文件中的敏感数据
git diff HEAD
```

### 3. 分支管理

如果您的项目使用不同的分支名称：

```bash
# 查看当前分支
git branch

# 推送指定分支
git push -u origin <分支名>
```

---

## 常见问题

### Q1: 推送时提示认证失败

**解决方法**：

1. **使用 HTTPS**：需要配置 Personal Access Token
   ```bash
   # 生成 Token: https://github.com/settings/tokens
   # 权限至少需要: repo
   git remote set-url origin https://您的用户名:TOKEN@github.com/用户名/仓库名.git
   ```

2. **使用 SSH**：需要配置 SSH 密钥
   ```bash
   # 生成 SSH 密钥
   ssh-keygen -t ed25519 -C "your_email@example.com"
   
   # 添加 SSH 密钥到 ssh-agent
   eval "$(ssh-agent -s)"
   ssh-add ~/.ssh/id_ed25519
   
   # 将公钥添加到 GitHub: https://github.com/settings/keys
   cat ~/.ssh/id_ed25519.pub
   ```

### Q2: 推送时提示仓库不存在

**解决方法**：
- 确保已在 GitHub 上创建仓库
- 检查仓库名称和 URL 是否正确
- 确保您有仓库的写入权限

### Q3: 推送速度很慢

**解决方法**：
- 检查 `.gitignore` 是否正确配置，避免推送大文件
- 考虑使用 SSH 而不是 HTTPS
- 如果项目很大，考虑使用 Git LFS 管理大文件

---

## 上传后的下一步

1. **设置仓库描述和主题**
   - 在 GitHub 仓库页面点击 ⚙️ Settings
   - 填写详细描述
   - 添加相关主题标签

2. **添加 README 徽章**
   - 在 README.md 中添加状态徽章
   - 展示项目版本、构建状态等

3. **配置 GitHub Actions**
   - 创建 `.github/workflows/ci.yml`
   - 设置自动化测试和构建

4. **创建 Releases**
   - 为重要版本创建 Release
   - 添加 changelog 和发布说明

---

## 需要帮助？

如果遇到问题，可以：
1. 查看 Git 日志：`git log --oneline`
2. 检查远程配置：`git remote -v`
3. 查看帮助：`git help push`

或参考官方文档：
- [GitHub 文档](https://docs.github.com/)
- [Git 官方文档](https://git-scm.com/doc)
