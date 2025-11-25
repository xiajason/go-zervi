<template>
  <div class="resumes-page">
    <DynamicTable
      table-name="resume"
      model-label="简历"
      api-path="/api/v1/resume/index"
      :can-create="true"
      :can-update="true"
      :can-delete="true"
    />
  </div>
</template>

<script setup lang="ts">
import DynamicTable from '@/components/DynamicTable.vue'
</script>

<style scoped>
.resumes-page {
  padding: 20px;
}
</style>

