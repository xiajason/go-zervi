# GoZervi 快速推送到 GitHub (xiajason 账号)

## 🚀 快速步骤

### 1. 获取 xiajason 的 Personal Access Token

1. 登录 xiajason 账号：https://github.com/login
2. 访问：https://github.com/settings/tokens
3. 点击 "Generate new token" → "Generate new token (classic)"
4. 勾选 `repo` 权限
5. 复制生成的 Token

### 2. 在 GitHub 上创建仓库

1. 使用 xiajason 账号访问：https://github.com/new
2. 仓库名称：`GoZervi`（或你喜欢的名字）
3. 描述：`GoZervi 智能化 SaaS 服务系统`
4. 选择 Public 或 Private
5. **不要**勾选 "Initialize this repository with a README"
6. 点击 "Create repository"

### 3. 推送代码

#### 方式 A: 使用脚本（推荐）

```bash
cd /Users/szjason72/gozervi/zervigo.demo
./scripts/push-to-github.sh https://github.com/xiajason/GoZervi.git main
```

推送时输入：
- Username: `xiajason`
- Password: 粘贴你的 Token

#### 方式 B: 直接推送（如果已有 Token）

```bash
cd /Users/szjason72/gozervi/zervigo.demo

# 替换 YOUR_TOKEN 为你的实际 Token
git remote add origin https://xiajason:YOUR_TOKEN@github.com/xiajason/GoZervi.git

# 切换到 main 分支（如果需要）
git checkout main

# 推送
git push -u origin main
```

## ✅ 完成

推送成功后，访问：https://github.com/xiajason/GoZervi

## 📝 后续推送

设置完成后，后续推送只需：

```bash
git add .
git commit -m "你的提交信息"
git push
```

## 🔐 安全提示

推送成功后，建议移除 URL 中的 Token：

```bash
git remote set-url origin https://github.com/xiajason/GoZervi.git
```

后续推送时会使用 credential helper 存储的凭据。

