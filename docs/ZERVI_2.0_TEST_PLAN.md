# Zervi 2.0 智能服务编排测试计划

**日期**: 2025-10-30  
**目标**: 测试7种服务组合

---

## 📋 测试组合列表

### 1. 单个服务组合 (3种)

#### 1.1 job_only
- **命令**: `./scripts/start-services.sh job_only`
- **应该启动**: auth + user + job
- **不应该启动**: resume, company

#### 1.2 resume_only
- **命令**: `./scripts/start-services.sh resume_only`
- **应该启动**: auth + user + resume
- **不应该启动**: job, company

#### 1.3 company_only
- **命令**: `./scripts/start-services.sh company_only`
- **应该启动**: auth + user + company
- **不应该启动**: job, resume

---

### 2. 两个服务组合 (3种)

#### 2.1 job_resume
- **命令**: `./scripts/start-services.sh job_resume`
- **应该启动**: auth + user + job + resume
- **不应该启动**: company

#### 2.2 job_company
- **命令**: `./scripts/start-services.sh job_company`
- **应该启动**: auth + user + job + company
- **不应该启动**: resume

#### 2.3 resume_company
- **命令**: `./scripts/start-services.sh resume_company`
- **应该启动**: auth + user + resume + company
- **不应该启动**: job

---

### 3. 所有服务组合 (1种)

#### 3.1 all_services
- **命令**: `./scripts/start-services.sh all_services`
- **应该启动**: auth + user + job + resume + company

---

## ✅ 测试步骤

### 每个测试的步骤

1. **停止所有服务**
   ```bash
   ./scripts/stop-local-services.sh
   ```

2. **启动指定组合**
   ```bash
   ./scripts/start-services.sh <composition>
   ```

3. **验证启动的服务**
   ```bash
   ps aux | grep "go run main.go" | grep -v grep
   ```

4. **健康检查**
   ```bash
   curl http://localhost:8207/health
   curl http://localhost:8082/health
   curl http://localhost:8084/health
   curl http://localhost:8085/health
   curl http://localhost:8083/health
   ```

5. **停止服务**
   ```bash
   ./scripts/stop-local-services.sh
   ```

---

## 📊 测试记录

### Test 1: job_only

**时间**: 2025-10-30  
**状态**: ⏳ 待测试

**预期结果**:
- ✅ auth-service 启动
- ✅ user-service 启动
- ✅ job-service 启动
- ❌ resume-service 不启动
- ❌ company-service 不启动

---

### Test 2: resume_only

**时间**: 2025-10-30  
**状态**: ⏳ 待测试

**预期结果**:
- ✅ auth-service 启动
- ✅ user-service 启动
- ✅ resume-service 启动
- ❌ job-service 不启动
- ❌ company-service 不启动

---

### Test 3: company_only

**时间**: 2025-10-30  
**状态**: ⏳ 待测试

**预期结果**:
- ✅ auth-service 启动
- ✅ user-service 启动
- ✅ company-service 启动
- ❌ job-service 不启动
- ❌ resume-service 不启动

---

### Test 4: job_resume

**时间**: 2025-10-30  
**状态**: ⏳ 待测试

**预期结果**:
- ✅ auth-service 启动
- ✅ user-service 启动
- ✅ job-service 启动
- ✅ resume-service 启动
- ❌ company-service 不启动

---

### Test 5: job_company

**时间**: 2025-10-30  
**状态**: ⏳ 待测试

**预期结果**:
- ✅ auth-service 启动
- ✅ user-service 启动
- ✅ job-service 启动
- ✅ company-service 启动
- ❌ resume-service 不启动

---

### Test 6: resume_company

**时间**: 2025-10-30  
**状态**: ⏳ 待测试

**预期结果**:
- ✅ auth-service 启动
- ✅ user-service 启动
- ✅ resume-service 启动
- ✅ company-service 启动
- ❌ job-service 不启动

---

### Test 7: all_services

**时间**: 2025-10-30  
**状态**: ⏳ 待测试

**预期结果**:
- ✅ auth-service 启动
- ✅ user-service 启动
- ✅ job-service 启动
- ✅ resume-service 启动
- ✅ company-service 启动

---

## 🎯 成功标准

### 每个测试必须通过

1. ✅ 服务按正确顺序启动
2. ✅ 只有需要的服务被启动
3. ✅ 所有启动的服务健康检查通过
4. ✅ 脚本输出清晰易懂
5. ✅ 无错误信息

---

## 📝 测试报告

**测试完成后更新此处**

