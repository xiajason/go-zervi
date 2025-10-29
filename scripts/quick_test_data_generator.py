#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
快速测试数据生成器
用于Job Matching功能验证
"""

import asyncio
import json
import sys
import os
from datetime import datetime

# 添加项目路径
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

async def create_test_data():
    """创建快速验证测试数据"""
    
    # ==================== 测试简历数据 ====================
    test_resumes = [
        {
            "id": 1,
            "user_id": 1,
            "title": "硬件工程师简历 - LCD TV",
            "file_name": "H003--LCD TV 硬件工程师.DOC",
            "parsed_data": {
                "personal_info": {
                    "name": "张三",
                    "gender": "男",
                    "age": 28,
                    "location": "深圳",
                    "email": "zhangsan@example.com",
                    "phone": "13800138000"
                },
                "education": [
                    {
                        "school": "某理工大学",
                        "degree": "本科",
                        "major": "电子工程",
                        "graduation_year": "2018"
                    }
                ],
                "work_experience": [
                    {
                        "company": "某电子公司",
                        "position": "硬件工程师",
                        "duration": "3年",
                        "responsibilities": "负责LCD TV硬件设计、电路设计、PCB布局、产品调试"
                    }
                ],
                "skills": [
                    "硬件设计",
                    "LCD TV",
                    "电路设计",
                    "PCB Layout",
                    "Allegro",
                    "OrCAD",
                    "信号完整性分析",
                    "EMC设计"
                ],
                "keywords": ["硬件工程师", "LCD TV", "电路设计", "PCB", "消费电子"]
            },
            "raw_content": """
姓名：张三
性别：男
年龄：28岁
地点：深圳
联系方式：13800138000

教育背景：
某理工大学 电子工程 本科 2018年毕业

工作经历：
某电子公司 硬件工程师 2018-2021（3年）
- 负责LCD TV硬件设计和开发
- 完成多款产品的电路设计和PCB布局
- 参与产品测试和调试工作
- 解决硬件相关技术问题

专业技能：
- 熟练掌握硬件设计流程
- 精通LCD TV相关技术
- 熟练使用Allegro、OrCAD等EDA工具
- 了解信号完整性分析和EMC设计
            """
        },
        {
            "id": 2,
            "user_id": 2,
            "title": "销售经理简历",
            "file_name": "简历1106--销售经理.xls",
            "parsed_data": {
                "personal_info": {
                    "name": "李四",
                    "gender": "女",
                    "age": 32,
                    "location": "北京",
                    "email": "lisi@example.com",
                    "phone": "13900139000"
                },
                "education": [
                    {
                        "school": "某财经大学",
                        "degree": "本科",
                        "major": "市场营销",
                        "graduation_year": "2014"
                    }
                ],
                "work_experience": [
                    {
                        "company": "某科技公司",
                        "position": "销售经理",
                        "duration": "5年",
                        "responsibilities": "负责华北区销售管理、客户开发、团队管理、销售目标达成"
                    }
                ],
                "skills": [
                    "销售管理",
                    "客户开发",
                    "团队管理",
                    "市场分析",
                    "商务谈判",
                    "大客户维护"
                ],
                "keywords": ["销售经理", "销售管理", "客户开发", "团队管理"]
            },
            "raw_content": """
姓名：李四
性别：女
年龄：32岁
地点：北京
联系方式：13900139000

教育背景：
某财经大学 市场营销 本科 2014年毕业

工作经历：
某科技公司 销售经理 2016-2021（5年）
- 负责华北区销售业务管理
- 完成年度销售目标150%
- 管理销售团队15人
- 开发大客户20+家

专业技能：
- 5年以上销售管理经验
- 熟悉B2B销售流程
- 优秀的客户开发和维护能力
- 良好的团队管理能力
            """
        },
        {
            "id": 3,
            "user_id": 3,
            "title": "软件工程师简历",
            "file_name": "316317189.pdf",
            "parsed_data": {
                "personal_info": {
                    "name": "王五",
                    "gender": "男",
                    "age": 26,
                    "location": "上海",
                    "email": "wangwu@example.com",
                    "phone": "13700137000"
                },
                "education": [
                    {
                        "school": "某交通大学",
                        "degree": "硕士",
                        "major": "计算机科学",
                        "graduation_year": "2020"
                    }
                ],
                "work_experience": [
                    {
                        "company": "某互联网公司",
                        "position": "后端工程师",
                        "duration": "3年",
                        "responsibilities": "负责后端服务开发、数据库设计、API开发、系统优化"
                    }
                ],
                "skills": [
                    "Python",
                    "Java",
                    "Go",
                    "MySQL",
                    "Redis",
                    "微服务架构",
                    "分布式系统",
                    "高并发"
                ],
                "keywords": ["软件工程师", "后端开发", "Python", "微服务", "分布式"]
            },
            "raw_content": """
姓名：王五
性别：男
年龄：26岁
地点：上海
联系方式：13700137000

教育背景：
某交通大学 计算机科学 硕士 2020年毕业

工作经历：
某互联网公司 后端工程师 2020-2023（3年）
- 负责核心业务系统的后端开发
- 完成多个高并发系统的架构设计
- 优化数据库性能，提升系统响应速度50%
- 参与微服务架构改造

专业技能：
- 精通Python、Java、Go等编程语言
- 熟悉MySQL、Redis、MongoDB等数据库
- 了解微服务架构和分布式系统
- 熟悉高并发系统设计
            """
        }
    ]
    
    # ==================== 测试职位数据 ====================
    test_jobs = [
        {
            "id": 1,
            "company_id": 1,
            "company_name": "某知名电子公司",
            "title": "高级硬件工程师（LCD TV方向）",
            "location": "深圳",
            "salary_range": "15K-25K",
            "experience_requirement": "3-5年",
            "education_requirement": "本科及以上",
            "job_description": """
岗位职责：
1. 负责LCD TV产品的硬件设计和开发
2. 完成电路设计、PCB布局和产品调试
3. 解决产品开发过程中的技术问题
4. 参与新产品的方案评审和技术支持

任职要求：
1. 本科及以上学历，电子、通信等相关专业
2. 3年以上LCD TV或消费电子硬件设计经验
3. 熟练使用Allegro、OrCAD等EDA工具
4. 了解信号完整性和EMC设计
5. 具备良好的团队协作能力
            """,
            "requirements": [
                "本科及以上学历",
                "3年以上LCD TV硬件设计经验",
                "熟练使用Allegro、OrCAD",
                "了解信号完整性和EMC设计",
                "良好的团队协作能力"
            ],
            "keywords": ["硬件工程师", "LCD TV", "电路设计", "PCB", "消费电子", "Allegro"]
        },
        {
            "id": 2,
            "company_id": 2,
            "company_name": "建设银行",
            "title": "销售经理",
            "location": "北京",
            "salary_range": "15K-30K",
            "experience_requirement": "5年以上",
            "education_requirement": "本科及以上",
            "job_description": """
岗位职责：
1. 负责区域内的销售业务管理和客户开发
2. 制定并执行销售计划，完成销售目标
3. 管理销售团队，提升团队业绩
4. 维护重要客户关系，开发新客户

任职要求：
1. 本科及以上学历，金融、市场营销等相关专业优先
2. 5年以上销售管理经验，有金融行业经验优先
3. 具备优秀的客户开发和维护能力
4. 良好的团队管理和领导能力
5. 具备良好的商务谈判和沟通能力
            """,
            "requirements": [
                "本科及以上学历",
                "5年以上销售管理经验",
                "金融行业经验优先",
                "客户开发和维护能力",
                "团队管理能力",
                "商务谈判能力"
            ],
            "keywords": ["销售经理", "销售管理", "客户开发", "团队管理", "金融"]
        },
        {
            "id": 3,
            "company_id": 3,
            "company_name": "某互联网公司",
            "title": "高级后端工程师（Python）",
            "location": "上海",
            "salary_range": "25K-40K",
            "experience_requirement": "3-5年",
            "education_requirement": "本科及以上",
            "job_description": """
岗位职责：
1. 负责核心业务系统的后端开发和架构设计
2. 参与微服务架构的设计和实施
3. 优化系统性能，提升用户体验
4. 解决高并发场景下的技术难题

任职要求：
1. 本科及以上学历，计算机相关专业
2. 3年以上Python后端开发经验
3. 熟悉MySQL、Redis、MongoDB等数据库
4. 了解微服务架构和分布式系统
5. 有高并发系统设计和优化经验
6. 良好的代码规范和文档习惯
            """,
            "requirements": [
                "本科及以上学历",
                "3年以上Python后端开发经验",
                "熟悉MySQL、Redis、MongoDB",
                "微服务架构和分布式系统",
                "高并发系统经验",
                "良好的代码规范"
            ],
            "keywords": ["后端工程师", "Python", "微服务", "分布式系统", "高并发", "MySQL", "Redis"]
        }
    ]
    
    return test_resumes, test_jobs

async def insert_test_data_to_db():
    """将测试数据插入到数据库"""
    import asyncpg
    import aiomysql
    
    # 配置
    mysql_config = {
        'host': 'localhost',
        'port': 3306,
        'user': 'root',
        'password': 'test_mysql_password',
        'database': 'jobfirst'
    }
    
    postgres_config = {
        'host': 'localhost',
        'port': 5432,
        'user': 'future_user',
        'password': 'f_postgres_password_2025',
        'database': 'jobfirst_vector'
    }
    
    test_resumes, test_jobs = await create_test_data()
    
    print("=" * 60)
    print("开始插入测试数据")
    print("=" * 60)
    
    # 生成简单的向量数据（模拟）
    def generate_mock_vector(text, dim=1536):
        """生成模拟向量数据"""
        import hashlib
        import numpy as np
        
        # 使用文本哈希生成确定性的向量
        hash_obj = hashlib.md5(text.encode())
        seed = int(hash_obj.hexdigest()[:8], 16)
        np.random.seed(seed)
        
        # 生成归一化向量
        vector = np.random.randn(dim)
        vector = vector / np.linalg.norm(vector)
        return vector.tolist()
    
    try:
        # 连接PostgreSQL
        postgres_conn = await asyncpg.connect(
            host=postgres_config['host'],
            port=postgres_config['port'],
            user=postgres_config['user'],
            password=postgres_config['password'],
            database=postgres_config['database']
        )
        
        print("\n✅ PostgreSQL连接成功")
        
        # 清空现有测试数据
        await postgres_conn.execute("DELETE FROM resume_vectors WHERE resume_id <= 3")
        await postgres_conn.execute("DELETE FROM job_vectors WHERE job_id <= 3")
        print("✅ 清空现有测试数据")
        
        # 插入简历向量
        print("\n插入简历向量数据...")
        for resume in test_resumes:
            content_vector = generate_mock_vector(resume['raw_content'])
            skills_vector = generate_mock_vector(' '.join(resume['parsed_data']['skills']))
            experience_vector = generate_mock_vector(
                ' '.join([exp['responsibilities'] for exp in resume['parsed_data']['work_experience']])
            )
            
            await postgres_conn.execute("""
                INSERT INTO resume_vectors 
                (resume_id, user_id, content_vector, skills_vector, experience_vector)
                VALUES ($1, $2, $3, $4, $5)
            """, resume['id'], resume['user_id'], content_vector, skills_vector, experience_vector)
            
            print(f"  ✅ 简历 #{resume['id']}: {resume['title']}")
        
        # 插入职位向量
        print("\n插入职位向量数据...")
        for job in test_jobs:
            title_vector = generate_mock_vector(job['title'])
            description_vector = generate_mock_vector(job['job_description'])
            requirements_vector = generate_mock_vector(' '.join(job['requirements']))
            
            await postgres_conn.execute("""
                INSERT INTO job_vectors 
                (job_id, company_id, title_vector, description_vector, requirements_vector)
                VALUES ($1, $2, $3, $4, $5)
            """, job['id'], job['company_id'], title_vector, description_vector, requirements_vector)
            
            print(f"  ✅ 职位 #{job['id']}: {job['title']}")
        
        # 验证数据
        resume_count = await postgres_conn.fetchval("SELECT COUNT(*) FROM resume_vectors WHERE resume_id <= 3")
        job_count = await postgres_conn.fetchval("SELECT COUNT(*) FROM job_vectors WHERE job_id <= 3")
        
        print("\n" + "=" * 60)
        print("测试数据插入完成！")
        print("=" * 60)
        print(f"✅ 简历向量数据: {resume_count} 条")
        print(f"✅ 职位向量数据: {job_count} 条")
        
        # 关闭连接
        await postgres_conn.close()
        
        return True
        
    except Exception as e:
        print(f"\n❌ 插入数据失败: {e}")
        import traceback
        traceback.print_exc()
        return False

async def generate_test_report():
    """生成测试数据报告"""
    test_resumes, test_jobs = await create_test_data()
    
    report = f"""
# Job Matching快速测试数据报告

**生成时间**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}  
**测试类型**: 快速验证测试

---

## 📊 测试数据概览

### 简历数据 (3份)

| ID | 标题 | 来源文件 | 关键技能 |
|----|------|---------|---------|
"""
    
    for resume in test_resumes:
        skills = ', '.join(resume['parsed_data']['skills'][:5])
        report += f"| {resume['id']} | {resume['title']} | {resume['file_name']} | {skills}... |\n"
    
    report += """
### 职位数据 (3个)

| ID | 岗位名称 | 公司 | 地点 | 薪资 |
|----|---------|------|------|------|
"""
    
    for job in test_jobs:
        report += f"| {job['id']} | {job['title']} | {job['company_name']} | {job['location']} | {job['salary_range']} |\n"
    
    report += """
---

## 🎯 预期匹配结果

### 高匹配度组合 (80%+)

1. **简历#1 (硬件工程师) ↔ 职位#1 (高级硬件工程师-LCD TV)**
   - 技能匹配: 95% (LCD TV、硬件设计、PCB、Allegro完全匹配)
   - 经验匹配: 100% (3年经验符合要求)
   - 教育匹配: 100% (本科符合要求)
   - 预期总分: **90-95%** ✅

2. **简历#2 (销售经理) ↔ 职位#2 (销售经理-建设银行)**
   - 技能匹配: 85% (销售管理、客户开发、团队管理)
   - 经验匹配: 100% (5年经验符合要求)
   - 教育匹配: 100% (本科符合要求)
   - 预期总分: **85-90%** ✅

3. **简历#3 (软件工程师) ↔ 职位#3 (高级后端工程师-Python)**
   - 技能匹配: 90% (Python、MySQL、Redis、微服务)
   - 经验匹配: 100% (3年经验符合要求)
   - 教育匹配: 100% (硕士超过要求)
   - 预期总分: **90-95%** ✅

### 中等匹配度组合 (50-70%)

4. **简历#1 (硬件工程师) ↔ 职位#3 (后端工程师)**
   - 技能匹配: 20% (技术栈完全不同)
   - 经验匹配: 50% (都是技术岗位)
   - 预期总分: **40-50%** ⚠️

5. **简历#2 (销售经理) ↔ 职位#1 (硬件工程师)**
   - 技能匹配: 10% (完全不同领域)
   - 经验匹配: 30% (都有团队协作)
   - 预期总分: **25-35%** ⚠️

---

## 🧪 测试执行步骤

### 1. 数据准备
```bash
python scripts/quick_test_data_generator.py
```

### 2. 执行匹配测试
```bash
# 测试简历#1匹配职位
curl -X POST http://localhost:8100/api/v1/ai/job-matching \\
  -H "Content-Type: application/json" \\
  -d '{"resume_id": 1, "limit": 3}'

# 测试简历#2匹配职位
curl -X POST http://localhost:8100/api/v1/ai/job-matching \\
  -H "Content-Type: application/json" \\
  -d '{"resume_id": 2, "limit": 3}'

# 测试简历#3匹配职位
curl -X POST http://localhost:8100/api/v1/ai/job-matching \\
  -H "Content-Type: application/json" \\
  -d '{"resume_id": 3, "limit": 3}'
```

### 3. 验证结果
- 检查匹配分数是否符合预期
- 验证排序是否正确
- 检查推荐理由是否合理

---

## ✅ 成功标准

| 指标 | 目标 | 验证方法 |
|------|------|---------|
| 高匹配准确率 | > 85% | 简历#1 ↔ 职位#1 得分 > 0.85 |
| 排序正确性 | 100% | 高匹配职位排在前面 |
| 响应时间 | < 2秒 | API响应时间测试 |
| 推荐合理性 | 主观评估 | 人工检查推荐理由 |

---

**📝 备注**: 
- 测试数据使用模拟向量（基于文本哈希生成）
- 实际生产环境应使用sentence-transformers生成真实向量
- 本测试主要验证系统功能和业务流程完整性
"""
    
    return report

if __name__ == "__main__":
    print("=" * 60)
    print("Job Matching 快速测试数据生成器")
    print("=" * 60)
    
    # 生成测试报告
    print("\n生成测试数据报告...")
    report = asyncio.run(generate_test_report())
    
    report_path = "/Users/szjason72/szbolent/Zervigo/快速测试数据报告.md"
    with open(report_path, 'w', encoding='utf-8') as f:
        f.write(report)
    print(f"✅ 测试报告已生成: {report_path}")
    
    # 插入测试数据
    print("\n是否立即插入测试数据到数据库? (y/n): ", end='')
    choice = input().strip().lower()
    
    if choice == 'y':
        success = asyncio.run(insert_test_data_to_db())
        if success:
            print("\n" + "=" * 60)
            print("✅ 所有准备工作完成！现在可以执行API测试了")
            print("=" * 60)
            print("\n📝 下一步:")
            print("1. 启动AI Service (如果未启动)")
            print("2. 执行Job Matching API测试")
            print("3. 查看测试结果")
        else:
            print("\n❌ 数据插入失败，请检查数据库连接")
    else:
        print("\n跳过数据插入，请手动执行测试数据插入")

