#!/bin/bash

echo "🚀 Zervigo MVP Go-Zero代码生成工具"
echo "📋 使用goctl工具生成标准Go-Zero微服务代码"

# 检查goctl是否安装
if ! command -v goctl &> /dev/null; then
    echo "❌ goctl未安装，正在安装..."
    go install github.com/zeromicro/go-zero/tools/goctl@latest
    if [ $? -ne 0 ]; then
        echo "❌ goctl安装失败，请手动安装"
        exit 1
    fi
    echo "✅ goctl安装成功"
fi

# 进入项目根目录
cd "$(dirname "$0")/.."

echo "📁 当前目录: $(pwd)"

# 创建必要的目录
mkdir -p service/{auth,user,job,resume,company,ai,blockchain}
mkdir -p rpc/{auth,user,job,resume,company,ai,blockchain}
mkdir -p model/{authmodel,usermodel,jobmodel,resumemodel,companymodel,blockchainmodel}

echo "🔨 开始生成代码..."

# 生成API服务
echo "1. 生成认证服务API..."
goctl api go -api api/auth.api -dir service/auth --style=goZero

echo "2. 生成用户服务API..."
goctl api go -api api/user.api -dir service/user --style=goZero

echo "3. 生成职位服务API..."
goctl api go -api api/job.api -dir service/job --style=goZero

echo "4. 生成简历服务API..."
goctl api go -api api/resume.api -dir service/resume --style=goZero

# 生成RPC服务
echo "5. 生成认证RPC服务..."
goctl rpc protoc rpc/auth/auth.proto --go_out=./rpc/auth --go-grpc_out=./rpc/auth --zrpc_out=./rpc/auth

# 生成数据模型
echo "6. 生成数据模型..."
goctl model mysql datasource -url="root:dev_password@tcp(localhost:3306)/zervigo_mvp" -table="user" -dir="./model/usermodel" --style=goZero
goctl model mysql datasource -url="root:dev_password@tcp(localhost:3306)/zervigo_mvp" -table="job" -dir="./model/jobmodel" --style=goZero
goctl model mysql datasource -url="root:dev_password@tcp(localhost:3306)/zervigo_mvp" -table="resume" -dir="./model/resumemodel" --style=goZero

# 生成Dockerfile
echo "7. 生成Dockerfile..."
goctl docker -go service/auth/main.go
goctl docker -go service/user/main.go
goctl docker -go service/job/main.go
goctl docker -go service/resume/main.go

# 生成K8s配置
echo "8. 生成Kubernetes配置..."
goctl kube deploy -name auth-api -namespace zervigo-mvp -image auth-api:latest -o deploy/k8s/auth-api.yaml
goctl kube deploy -name user-api -namespace zervigo-mvp -image user-api:latest -o deploy/k8s/user-api.yaml
goctl kube deploy -name job-api -namespace zervigo-mvp -image job-api:latest -o deploy/k8s/job-api.yaml
goctl kube deploy -name resume-api -namespace zervigo-mvp -image resume-api:latest -o deploy/k8s/resume-api.yaml

echo "✅ 代码生成完成！"
echo ""
echo "📊 生成的文件结构："
echo "   service/          - HTTP API服务"
echo "   rpc/             - RPC服务"
echo "   model/           - 数据模型"
echo "   deploy/k8s/      - Kubernetes配置"
echo ""
echo "🚀 下一步："
echo "   1. 检查生成的代码"
echo "   2. 配置数据库连接"
echo "   3. 实现业务逻辑"
echo "   4. 启动服务测试"
echo ""
echo "💡 使用说明："
echo "   # 启动认证服务"
echo "   cd service/auth && go run main.go"
echo ""
echo "   # 启动用户服务"
echo "   cd service/user && go run main.go"
echo ""
echo "   # 启动职位服务"
echo "   cd service/job && go run main.go"
echo ""
echo "   # 启动简历服务"
echo "   cd service/resume && go run main.go"
