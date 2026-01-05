# GoZervi 推送到 GitHub 指南

## 📋 当前状态

- ✅ 项目已经是 Git 仓库
- ✅ 有 `.gitignore` 文件
- ✅ 当前分支: `feature/oauth2-provider`
- ⚠️ 有未提交的更改
- ❌ 未配置远程仓库
- ✅ **账号选择**: 使用 `xiajason` 账号（保持项目历史一致性）

## 🚀 推送步骤

### 步骤 1: 提交当前更改（可选但推荐）

如果有未提交的更改，建议先提交：

```bash
cd /Users/szjason72/gozervi/zervigo.demo

# 查看更改
git status

# 添加更改
git add .

# 提交更改
git commit -m "chore: prepare for GitHub push"
```

### 步骤 2: 切换到 main 分支（推荐）

推送到 GitHub 时，通常使用 `main` 分支作为主分支：

```bash
# 切换到 main 分支
git checkout main

# 如果需要合并 feature 分支的更改
git merge feature/oauth2-provider
```

或者直接推送当前分支：

```bash
# 继续使用当前分支
git checkout feature/oauth2-provider
```

### 步骤 3: 在 GitHub 上创建仓库

**重要**: 使用 **xiajason** 账号创建仓库

1. **登录 xiajason 账号**: https://github.com/login
2. 访问：https://github.com/new
3. 仓库名称：`GoZervi` 或 `zervigo-demo`（你喜欢的名字）
4. 描述：`GoZervi 智能化 SaaS 服务系统`
5. 选择：Public 或 Private
6. **不要**勾选 "Initialize this repository with a README"
7. 点击 "Create repository"

### 步骤 4: 获取 xiajason 的 Personal Access Token

1. **登录 xiajason 账号**: https://github.com/login
2. 访问：https://github.com/settings/tokens
3. 点击 **"Generate new token"** → **"Generate new token (classic)"**
4. 设置 Token 信息：
   - **Note**: `GoZervi Project` (描述性名称)
   - **Expiration**: 选择过期时间（建议 90 天或自定义）
   - **Scopes**: 勾选 `repo` (完整仓库访问权限)
5. 点击 **"Generate token"**
6. **重要**: 立即复制 Token（只显示一次！）

### 步骤 5: 配置远程仓库并推送

#### 方法 A: 使用推送脚本（推荐）

```bash
cd /Users/szjason72/gozervi/zervigo.demo
./scripts/push-to-github.sh https://github.com/xiajason/GoZervi.git main
```

推送时会提示输入凭据：
- **Username**: `xiajason`
- **Password**: 粘贴你的 Personal Access Token

#### 方法 B: 手动配置

```bash
cd /Users/szjason72/gozervi/zervigo.demo

# 添加远程仓库（使用 xiajason 账号）
git remote add origin https://github.com/xiajason/GoZervi.git

# 推送到 GitHub（会提示输入凭据）
git push -u origin main
# 或者推送当前分支
git push -u origin feature/oauth2-provider
```

### 步骤 6: 认证

推送时会提示输入凭据：
- **Username**: `xiajason`
- **Password**: 你的 Personal Access Token（不是 GitHub 密码）

## 🔐 使用 Personal Access Token（推荐方式）

如果遇到认证问题，可以在 URL 中包含 Token：

```bash
# 替换 YOUR_TOKEN 为 xiajason 的 Token
git remote set-url origin https://xiajason:YOUR_TOKEN@github.com/xiajason/GoZervi.git
git push -u origin main
```

**安全提示**: 推送成功后，建议从 URL 中移除 Token：

```bash
# 移除 URL 中的 Token，改用 credential helper
git remote set-url origin https://github.com/xiajason/GoZervi.git
```

## 📝 推送多个分支

如果需要推送所有分支：

```bash
# 推送所有分支
git push --all origin

# 推送所有标签
git push --tags origin
```

## ⚠️ 注意事项

1. **敏感信息**: 确保 `.gitignore` 已正确配置，不会推送敏感信息（如 `.env`、密钥等）
2. **大文件**: 如果项目包含大文件，考虑使用 Git LFS
3. **私有仓库**: 如果包含敏感代码，建议使用 Private 仓库
4. **分支策略**: 建议使用 `main` 作为主分支，`develop` 作为开发分支

## 🔄 后续推送

设置完成后，后续推送只需：

```bash
git add .
git commit -m "你的提交信息"
git push
```

## 📚 相关文档

- [GitHub 设置指南](../TimesSquare/GITHUB_SETUP.md)
- [TimesSquare 与 GoZervi 的关系](../TimesSquare/TIMESQUARE_AND_GOZERVI.md)

