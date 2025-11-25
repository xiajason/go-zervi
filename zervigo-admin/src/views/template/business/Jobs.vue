<template>
  <div class="jobs-page">
    <DynamicTable
      table-name="job"
      model-label="职位"
      api-path="/api/v1/job/index"
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
.jobs-page {
  padding: 20px;
}
</style>

