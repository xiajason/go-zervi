<template>
  <DynamicTable
    table-name="roles"
    model-label="角色"
    api-path="/api/v1/roles/index"
    :can-create="true"
    :can-update="true"
    :can-delete="true"
  />
</template>

<script setup lang="ts">
import DynamicTable from '../../components/DynamicTable.vue'
</script>

<style scoped>
/* 
  这就是VueCMF的精华！
  一个组件 + 几行配置 = 完整的CRUD界面
  
  但我们做得更好：
  - 代码清晰易懂（不是黑盒）
  - 格式简洁明了（不是四层嵌套）
  - 完全可控可扩展（不受框架限制）
*/
</style>





